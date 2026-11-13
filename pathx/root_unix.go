//go:build darwin || linux

package pathx

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

type unixRoot struct{ fd int }

func openRoot(path string) (rootHandle, error) {
	if path == "" {
		return nil, ErrInvalidPath
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	fd, err := unix.Open(abs, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, secureDirectoryError(err)
	}
	return &unixRoot{fd: fd}, nil
}

func (r *unixRoot) close() error { return unix.Close(r.fd) }

func (r *unixRoot) openDir(name string) (*os.File, error) {
	if name == "" || name == "." {
		// 使用 openat 创建独立的目录打开描述，不能复制根目录描述符；
		// dup 会共享目录读取偏移，后续 ReadDir 可能读到空结果。
		fd, err := openDirectoryAt(r.fd, ".")
		if err != nil {
			return nil, secureDirectoryError(err)
		}
		return os.NewFile(uintptr(fd), "."), nil
	}
	fd, err := unix.Dup(r.fd)
	if err != nil {
		return nil, err
	}
	parts, err := rootParts(name)
	if err != nil {
		unix.Close(fd)
		return nil, err
	}
	for _, part := range parts {
		next, err := openDirectoryAt(fd, part)
		if err != nil {
			unix.Close(fd)
			return nil, secureDirectoryError(err)
		}
		if err := unix.Close(fd); err != nil {
			unix.Close(next)
			return nil, err
		}
		fd = next
	}
	return os.NewFile(uintptr(fd), parts[len(parts)-1]), nil
}

func (r *unixRoot) openFile(name string, flag int, perm fs.FileMode) (*os.File, error) {
	parent, base, err := r.parent(name)
	if err != nil {
		return nil, err
	}
	defer unix.Close(parent)
	fd, err := unix.Openat(parent, base, flag|unix.O_NOFOLLOW|unix.O_CLOEXEC, uint32(perm.Perm()))
	if err != nil {
		return nil, securePathError(err)
	}
	return os.NewFile(uintptr(fd), base), nil
}

func (r *unixRoot) mkdirAll(name string, perm fs.FileMode) error {
	parts, err := rootParts(name)
	if err != nil {
		return err
	}
	fd, err := unix.Dup(r.fd)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	for _, part := range parts {
		next, openErr := openDirectoryAt(fd, part)
		if errors.Is(openErr, unix.ENOENT) {
			if err := unix.Mkdirat(fd, part, uint32(perm.Perm())); err != nil && !errors.Is(err, unix.EEXIST) {
				return securePathError(err)
			}
			next, openErr = openDirectoryAt(fd, part)
		}
		if openErr != nil {
			return secureDirectoryError(openErr)
		}
		if err := unix.Close(fd); err != nil {
			unix.Close(next)
			return err
		}
		fd = next
	}
	return nil
}

func (r *unixRoot) remove(name string) error {
	parent, base, err := r.parent(name)
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	var stat unix.Stat_t
	if err := unix.Fstatat(parent, base, &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return securePathError(err)
	}
	flags := 0
	if stat.Mode&unix.S_IFMT == unix.S_IFDIR {
		flags = unix.AT_REMOVEDIR
	}
	return securePathError(unix.Unlinkat(parent, base, flags))
}

func (r *unixRoot) rename(oldName, newName string) error {
	oldParent, oldBase, err := r.parent(oldName)
	if err != nil {
		return err
	}
	defer unix.Close(oldParent)
	newParent, newBase, err := r.parent(newName)
	if err != nil {
		return err
	}
	defer unix.Close(newParent)
	return securePathError(unix.Renameat(oldParent, oldBase, newParent, newBase))
}

func (r *unixRoot) parent(name string) (int, string, error) {
	parts, err := rootParts(name)
	if err != nil {
		return -1, "", err
	}
	fd, err := unix.Dup(r.fd)
	if err != nil {
		return -1, "", err
	}
	for _, part := range parts[:len(parts)-1] {
		next, err := openDirectoryAt(fd, part)
		if err != nil {
			unix.Close(fd)
			return -1, "", secureDirectoryError(err)
		}
		if err := unix.Close(fd); err != nil {
			unix.Close(next)
			return -1, "", err
		}
		fd = next
	}
	return fd, parts[len(parts)-1], nil
}

func openDirectoryAt(parent int, name string) (int, error) {
	return unix.Openat(parent, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
}

func rootParts(name string) ([]string, error) {
	clean, err := SanitizeRelativePath(name)
	if err != nil {
		return nil, err
	}
	return strings.Split(clean, "/"), nil
}

func securePathError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, unix.ELOOP) {
		return fmt.Errorf("%w: symbolic link is not allowed", ErrPathEscape)
	}
	return err
}

func secureDirectoryError(err error) error {
	// Darwin 在 O_NOFOLLOW|O_DIRECTORY 遇到符号链接时返回 ENOTDIR，
	// Linux 通常返回 ELOOP；两者都不能作为受限根目录中的父目录继续使用。
	if errors.Is(err, unix.ELOOP) || errors.Is(err, unix.ENOTDIR) {
		return fmt.Errorf("%w: untrusted directory component", ErrPathEscape)
	}
	return securePathError(err)
}
