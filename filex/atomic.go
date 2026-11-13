package filex

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// AtomicWriteOptions 配置同目录临时文件写入；零值不自动创建目录。
type AtomicWriteOptions struct {
	Mode      fs.FileMode
	NoSync    bool
	CreateDir bool
}

// WriteAtomic 写入数据后通过同目录重命名提交；非 Unix 平台原子性取决于系统。
func WriteAtomic(destination string, data []byte, options AtomicWriteOptions) error {
	_, err := CopyAtomic(destination, bytes.NewReader(data), options)
	return err
}

// CopyAtomic 将 source 复制到临时文件再提交，返回已复制字节数，不关闭 source。
// 错误时字节数不表示提交成功；不提供目录同步或符号链接安全保证。
func CopyAtomic(destination string, source io.Reader, options AtomicWriteOptions) (int64, error) {
	return copyAtomic(destination, source, options, atomicOps{
		create: func(dir string) (atomicFile, error) { return os.CreateTemp(dir, ".atomic-*") },
		mkdir:  os.MkdirAll, rename: os.Rename, remove: os.Remove,
	})
}

// AtomicRoot 提供受限目录中的最小原子写入能力。
// filex 只依赖该接口，不直接引用任何其他基础库包。
type AtomicRoot interface {
	OpenFile(name string, flag int, perm fs.FileMode) (*os.File, error)
	MkdirAll(name string, perm fs.FileMode) error
	Remove(name string) error
	Rename(oldName, newName string) error
}

// WriteAtomicAt 通过受限根目录原子写入数据，destination 必须是根目录内的相对路径。
func WriteAtomicAt(root AtomicRoot, destination string, data []byte, options AtomicWriteOptions) error {
	_, err := CopyAtomicAt(root, destination, bytes.NewReader(data), options)
	return err
}

// CopyAtomicAt 在受限根目录内复制并原子替换目标文件，不关闭 source。
// 与 CopyAtomic 不同，该方法不重新解析绝对路径，因此不会跟随符号链接或 reparse point 越出根目录。
func CopyAtomicAt(root AtomicRoot, destination string, source io.Reader, options AtomicWriteOptions) (int64, error) {
	if root == nil || source == nil {
		return 0, fmt.Errorf("root and source are required")
	}
	relative, err := sanitizeAtomicRelativePath(destination)
	if err != nil {
		return 0, err
	}
	mode := options.Mode
	if mode == 0 {
		mode = 0o644
	}
	if mode&^fs.FileMode(0o777) != 0 {
		return 0, fmt.Errorf("mode must contain only permission bits")
	}
	directory := path.Dir(relative)
	if directory == "." {
		directory = ""
	}
	if options.CreateDir && directory != "" {
		if err := root.MkdirAll(directory, 0o755); err != nil {
			return 0, err
		}
	}
	temporary, err := atomicRelativeTempPath(directory)
	if err != nil {
		return 0, err
	}
	file, err := root.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return 0, err
	}
	closed, committed := false, false
	defer func() {
		if !closed {
			_ = file.Close()
		}
		if !committed {
			_ = root.Remove(temporary)
		}
	}()
	n, err := io.Copy(file, source)
	if err != nil {
		return n, err
	}
	if err := file.Chmod(mode); err != nil {
		return n, err
	}
	if !options.NoSync {
		if err := file.Sync(); err != nil {
			return n, err
		}
	}
	closed = true
	if err := file.Close(); err != nil {
		return n, err
	}
	if err := root.Rename(temporary, relative); err != nil {
		return n, err
	}
	committed = true
	return n, nil
}

func sanitizeAtomicRelativePath(destination string) (string, error) {
	if destination == "" || strings.ContainsRune(destination, '\x00') {
		return "", fmt.Errorf("destination is required")
	}
	cleaned := path.Clean(strings.ReplaceAll(destination, "\\", "/"))
	if cleaned == "." || cleaned == "/" || strings.HasPrefix(cleaned, "/") || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("destination must be a relative path")
	}
	return cleaned, nil
}

func atomicRelativeTempPath(directory string) (string, error) {
	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return path.Join(directory, ".atomic-"+hex.EncodeToString(random)), nil
}

// 注入仅在包内使用，避免全局可变钩子造成并发测试污染。
type atomicFile interface {
	io.Writer
	Name() string
	Chmod(fs.FileMode) error
	Sync() error
	Close() error
}
type atomicOps struct {
	create func(string) (atomicFile, error)
	mkdir  func(string, fs.FileMode) error
	rename func(string, string) error
	remove func(string) error
}

func copyAtomic(destination string, source io.Reader, options AtomicWriteOptions, ops atomicOps) (int64, error) {
	if destination == "" || source == nil {
		return 0, fmt.Errorf("destination and source are required")
	}
	mode := options.Mode
	if mode == 0 {
		mode = 0o644
	}
	if mode & ^fs.FileMode(0o777) != 0 {
		return 0, fmt.Errorf("mode must contain only permission bits")
	}
	dir := filepath.Dir(destination)
	if options.CreateDir {
		if err := ops.mkdir(dir, 0o755); err != nil {
			return 0, err
		}
	}
	temp, err := ops.create(dir)
	if err != nil {
		return 0, err
	}
	closed, committed := false, false
	defer func() {
		if !closed {
			_ = temp.Close()
		}
		if !committed {
			_ = ops.remove(temp.Name())
		}
	}()
	n, err := io.Copy(temp, source)
	if err != nil {
		return n, err
	}
	if err := temp.Chmod(mode); err != nil {
		return n, err
	}
	if !options.NoSync {
		if err := temp.Sync(); err != nil {
			return n, err
		}
	}
	closed = true
	if err := temp.Close(); err != nil {
		return n, err
	}
	if err := ops.rename(temp.Name(), destination); err != nil {
		return n, err
	}
	committed = true
	return n, nil
}
