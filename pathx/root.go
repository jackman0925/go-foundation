package pathx

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

// ErrRootClosed 表示受限根目录已关闭，不能再执行文件操作。
var ErrRootClosed = errors.New("secure root is closed")

// ErrUnsupportedPlatform 表示当前系统没有受支持的受限目录句柄实现。
var ErrUnsupportedPlatform = errors.New("secure root is unsupported on this platform")

// Root 将所有文件操作限制在打开时确定的目录句柄内。
// 不返回可再次解析的绝对路径，从而避免检查和使用之间的目录替换竞态。
type Root struct {
	mu   sync.RWMutex
	impl rootHandle
}

type rootHandle interface {
	close() error
	openDir(name string) (*os.File, error)
	openFile(name string, flag int, perm fs.FileMode) (*os.File, error)
	mkdirAll(name string, perm fs.FileMode) error
	remove(name string) error
	rename(oldName, newName string) error
}

// OpenRoot 打开受限根目录。根目录本身不得是符号链接或 Windows reparse point。
func OpenRoot(path string) (*Root, error) {
	impl, err := openRoot(path)
	if err != nil {
		return nil, err
	}
	return &Root{impl: impl}, nil
}

// Close 关闭根目录句柄。重复关闭安全。
func (r *Root) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.impl == nil {
		return nil
	}
	err := r.impl.close()
	r.impl = nil
	return err
}

// OpenDir 安全打开根目录或其内部目录，用于读取目录项。
// name 为空或为 . 时表示受限根目录本身。
func (r *Root) OpenDir(name string) (*os.File, error) {
	return r.withHandleFile(func(impl rootHandle) (*os.File, error) {
		return impl.openDir(name)
	})
}

// Open 以只读方式安全打开根目录内的现有文件或目录。
func (r *Root) Open(name string) (*os.File, error) {
	return r.OpenFile(name, os.O_RDONLY, 0)
}

// OpenFile 在根目录内安全打开或创建文件。
// 仅支持 os.O_WRONLY、os.O_RDWR、os.O_CREATE、os.O_EXCL 与 os.O_TRUNC。
// 父目录必须已存在；需要创建父目录时先调用 MkdirAll。
func (r *Root) OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	if err := validateOpenFlags(flag); err != nil {
		return nil, err
	}
	if err := validateFileMode(perm); err != nil {
		return nil, err
	}
	return r.withHandleFile(func(impl rootHandle) (*os.File, error) {
		return impl.openFile(name, flag, perm)
	})
}

// MkdirAll 在根目录内创建多级目录，任何符号链接或 reparse point 都会被拒绝。
func (r *Root) MkdirAll(name string, perm fs.FileMode) error {
	if err := validateDirMode(perm); err != nil {
		return err
	}
	return r.withHandle(func(impl rootHandle) error { return impl.mkdirAll(name, perm) })
}

// Remove 删除根目录内的单个文件、空目录或符号链接本身，不递归删除。
func (r *Root) Remove(name string) error {
	return r.withHandle(func(impl rootHandle) error { return impl.remove(name) })
}

// Rename 在同一受限根目录内移动或重命名对象。目标若存在会被平台原子替换。
func (r *Root) Rename(oldName, newName string) error {
	return r.withHandle(func(impl rootHandle) error { return impl.rename(oldName, newName) })
}

func (r *Root) withHandle(fn func(rootHandle) error) error {
	if r == nil {
		return ErrRootClosed
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.impl == nil {
		return ErrRootClosed
	}
	return fn(r.impl)
}

func (r *Root) withHandleFile(fn func(rootHandle) (*os.File, error)) (*os.File, error) {
	if r == nil {
		return nil, ErrRootClosed
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.impl == nil {
		return nil, ErrRootClosed
	}
	return fn(r.impl)
}

func validateOpenFlags(flag int) error {
	const allowed = os.O_WRONLY | os.O_RDWR | os.O_CREATE | os.O_EXCL | os.O_TRUNC
	writeOnly := flag&os.O_WRONLY != 0 || flag&os.O_RDWR != 0
	if flag&^allowed != 0 || (flag&os.O_EXCL != 0 && flag&os.O_CREATE == 0) || (flag&os.O_TRUNC != 0 && !writeOnly) {
		return fmt.Errorf("%w: unsupported open flags", ErrInvalidPath)
	}
	return nil
}

func validateFileMode(mode fs.FileMode) error {
	if mode == 0 {
		return nil
	}
	if mode&^fs.FileMode(0o777) != 0 {
		return fmt.Errorf("%w: mode must contain only permission bits", ErrInvalidPath)
	}
	return nil
}

func validateDirMode(mode fs.FileMode) error {
	if mode == 0 {
		return fmt.Errorf("%w: directory mode is required", ErrInvalidPath)
	}
	return validateFileMode(mode)
}
