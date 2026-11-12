package filex

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
