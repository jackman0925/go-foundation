package lockx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestKeyedLockerLockRejectsEmptyKey(t *testing.T) {
	locker := NewKeyedLocker()

	_, err := locker.Lock(context.Background(), " \t ")
	if !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("Lock() error = %v, want %v", err, ErrEmptyKey)
	}
}

func TestKeyedLockerLockAcceptsNilContext(t *testing.T) {
	locker := NewKeyedLocker()

	unlock, err := locker.Lock(nil, "order:1")
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}
	unlock()
}

func TestKeyedLockerTryLockRejectsEmptyKey(t *testing.T) {
	locker := NewKeyedLocker()

	unlock, ok := locker.TryLock("")
	if ok || unlock != nil {
		t.Fatalf("TryLock() = (%v, %v), want (nil, false)", unlock, ok)
	}
}

func TestKeyedLockerSameKeySerializes(t *testing.T) {
	locker := NewKeyedLocker()
	firstUnlock, err := locker.Lock(context.Background(), "order:1")
	if err != nil {
		t.Fatalf("first Lock() error = %v", err)
	}

	acquired := make(chan UnlockFunc, 1)
	go func() {
		unlock, lockErr := locker.Lock(context.Background(), "order:1")
		if lockErr != nil {
			t.Errorf("second Lock() error = %v", lockErr)
			return
		}
		acquired <- unlock
	}()

	select {
	case <-acquired:
		t.Fatal("second Lock() acquired before the first lock was released")
	case <-time.After(30 * time.Millisecond):
	}

	firstUnlock()
	select {
	case unlock := <-acquired:
		unlock()
	case <-time.After(time.Second):
		t.Fatal("second Lock() did not acquire after the first lock was released")
	}
}

func TestKeyedLockerDifferentKeysDoNotBlock(t *testing.T) {
	locker := NewKeyedLocker()
	firstUnlock, err := locker.Lock(context.Background(), "order:1")
	if err != nil {
		t.Fatalf("first Lock() error = %v", err)
	}
	defer firstUnlock()

	secondUnlock, err := locker.Lock(context.Background(), "order:2")
	if err != nil {
		t.Fatalf("second Lock() error = %v", err)
	}
	secondUnlock()
}

func TestKeyedLockerTryLock(t *testing.T) {
	locker := NewKeyedLocker()
	unlock, ok := locker.TryLock("order:1")
	if !ok {
		t.Fatal("first TryLock() ok = false, want true")
	}

	if blockedUnlock, blocked := locker.TryLock("order:1"); blocked || blockedUnlock != nil {
		t.Fatalf("second TryLock() = (%v, %v), want (nil, false)", blockedUnlock, blocked)
	}

	unlock()
	unlock, ok = locker.TryLock("order:1")
	if !ok {
		t.Fatal("TryLock() after unlock ok = false, want true")
	}
	unlock()
}

func TestKeyedLockerLockContextTimeout(t *testing.T) {
	locker := NewKeyedLocker()
	unlock, err := locker.Lock(context.Background(), "order:1")
	if err != nil {
		t.Fatalf("first Lock() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err = locker.Lock(ctx, "order:1")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waiting Lock() error = %v, want %v", err, context.DeadlineExceeded)
	}

	unlock()
	if got := locker.Len(); got != 0 {
		t.Fatalf("Len() after unlock = %d, want 0", got)
	}
}

func TestKeyedLockerUnlockIsIdempotent(t *testing.T) {
	locker := NewKeyedLocker()
	unlock, err := locker.Lock(context.Background(), "order:1")
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}

	unlock()
	unlock()
	if got := locker.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}

func TestKeyedLockerCancelledWaiterDoesNotBlockNextWaiter(t *testing.T) {
	locker := NewKeyedLocker()
	firstUnlock, err := locker.Lock(context.Background(), "order:1")
	if err != nil {
		t.Fatalf("first Lock() error = %v", err)
	}

	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancelled := make(chan error, 1)
	go func() {
		_, lockErr := locker.Lock(cancelledCtx, "order:1")
		cancelled <- lockErr
	}()

	// 确保第一个等待者已入队，再取消它，验证不会阻塞后续等待者。
	deadline := time.Now().Add(time.Second)
	for waiterCount(locker, "order:1") != 1 {
		if time.Now().After(deadline) {
			t.Fatal("cancelled waiter was not queued")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-cancelled; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Lock() error = %v, want %v", err, context.Canceled)
	}

	nextAcquired := make(chan UnlockFunc, 1)
	go func() {
		unlock, lockErr := locker.Lock(context.Background(), "order:1")
		if lockErr != nil {
			t.Errorf("next Lock() error = %v", lockErr)
			return
		}
		nextAcquired <- unlock
	}()

	firstUnlock()
	select {
	case unlock := <-nextAcquired:
		unlock()
	case <-time.After(time.Second):
		t.Fatal("next waiter did not acquire after cancelled waiter was removed")
	}
}

func waiterCount(locker *KeyedLocker, key string) int {
	locker.mu.Lock()
	defer locker.mu.Unlock()

	entry := locker.entries[key]
	if entry == nil {
		return 0
	}
	return len(entry.waiters)
}

func TestKeyedLockerLenTracksActiveKeys(t *testing.T) {
	locker := NewKeyedLocker()
	firstUnlock, err := locker.Lock(context.Background(), "order:1")
	if err != nil {
		t.Fatalf("first Lock() error = %v", err)
	}
	secondUnlock, err := locker.Lock(context.Background(), "order:2")
	if err != nil {
		t.Fatalf("second Lock() error = %v", err)
	}

	if got := locker.Len(); got != 2 {
		t.Fatalf("Len() = %d, want 2", got)
	}
	firstUnlock()
	if got := locker.Len(); got != 1 {
		t.Fatalf("Len() after first unlock = %d, want 1", got)
	}
	secondUnlock()
	if got := locker.Len(); got != 0 {
		t.Fatalf("Len() after all unlock = %d, want 0", got)
	}
}

func TestKeyedLockerRemoveWaiter(t *testing.T) {
	locker := NewKeyedLocker()
	waiter := make(chan struct{})
	locker.entries["order:1"] = &lockEntry{
		waiters: []chan struct{}{waiter},
	}

	if !locker.removeWaiter("order:1", waiter) {
		t.Fatal("removeWaiter() = false, want true")
	}
	if got := waiterCount(locker, "order:1"); got != 0 {
		t.Fatalf("waiter count = %d, want 0", got)
	}
	if locker.removeWaiter("order:1", make(chan struct{})) {
		t.Fatal("removeWaiter() with unknown waiter = true, want false")
	}
	if locker.removeWaiter("missing", waiter) {
		t.Fatal("removeWaiter() with missing key = true, want false")
	}
}

func TestKeyedLockerUnlockMissingKey(t *testing.T) {
	locker := NewKeyedLocker()

	locker.unlock("missing")
}
