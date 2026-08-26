// Package lockx 提供仅在单个进程内生效的按键互斥锁。
package lockx

import (
	"context"
	"errors"
	"strings"
	"sync"
)

// ErrEmptyKey 表示锁键为空。
var ErrEmptyKey = errors.New("lock key is required")

// UnlockFunc 释放一次已获得的锁。
//
// 即使被重复调用，也只会实际释放一次。
type UnlockFunc func()

// KeyedLocker 提供按键粒度的进程内互斥锁。
//
// 不同键之间互不阻塞；同一键在同一时刻只允许一个调用方持有。
// 它不具备跨进程、重入、租约续期或公平调度能力。
type KeyedLocker struct {
	mu      sync.Mutex
	entries map[string]*lockEntry
}

type lockEntry struct {
	waiters []chan struct{}
}

// NewKeyedLocker 创建按键互斥锁实例。
func NewKeyedLocker() *KeyedLocker {
	return &KeyedLocker{
		entries: make(map[string]*lockEntry),
	}
}

// Lock 获取 key 对应的锁；若锁已被占用，会等待至获得锁或 ctx 取消。
//
// 返回的 UnlockFunc 必须由调用方执行，通常应紧跟在成功返回后 defer 调用。
func (l *KeyedLocker) Lock(ctx context.Context, key string) (UnlockFunc, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	waiter, acquired := l.acquireOrQueue(key)
	if acquired {
		return l.newUnlockFunc(key), nil
	}

	select {
	case <-waiter:
		return l.newUnlockFunc(key), nil
	case <-ctx.Done():
		if !l.removeWaiter(key, waiter) {
			// 等待通道已被解锁方取走，表示当前调用方已经取得锁。
			return l.newUnlockFunc(key), nil
		}
		return nil, ctx.Err()
	}
}

// TryLock 尝试立即获取 key 对应的锁。
//
// 返回 false 表示该键已被占用，或 key 为空；此时没有锁被获取。
func (l *KeyedLocker) TryLock(key string) (UnlockFunc, bool) {
	if validateKey(key) != nil {
		return nil, false
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, exists := l.entries[key]
	if exists {
		return nil, false
	}
	l.entries[key] = &lockEntry{}
	return l.newUnlockFunc(key), true
}

// Len 返回当前持有或等待中的不同锁键数量。
func (l *KeyedLocker) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

func (l *KeyedLocker) acquireOrQueue(key string) (chan struct{}, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, exists := l.entries[key]
	if !exists {
		l.entries[key] = &lockEntry{}
		return nil, true
	}
	waiter := make(chan struct{})
	entry.waiters = append(entry.waiters, waiter)
	return waiter, false
}

func (l *KeyedLocker) removeWaiter(key string, waiter chan struct{}) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, exists := l.entries[key]
	if !exists {
		return false
	}
	for i, current := range entry.waiters {
		if current != waiter {
			continue
		}
		entry.waiters = append(entry.waiters[:i], entry.waiters[i+1:]...)
		return true
	}
	return false
}

func (l *KeyedLocker) newUnlockFunc(key string) UnlockFunc {
	var once sync.Once
	return func() {
		once.Do(func() {
			l.unlock(key)
		})
	}
}

func (l *KeyedLocker) unlock(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, exists := l.entries[key]
	if !exists {
		return
	}
	if len(entry.waiters) == 0 {
		delete(l.entries, key)
		return
	}

	// 仅唤醒队首等待者，使同一键始终保持单个持有者。
	next := entry.waiters[0]
	entry.waiters = entry.waiters[1:]
	close(next)
}

func validateKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return ErrEmptyKey
	}
	return nil
}
