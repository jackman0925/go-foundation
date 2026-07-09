package filex

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Exists 判断路径是否存在。
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsFile 判断路径是否存在且是普通文件。
func IsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

// IsDir 判断路径是否存在且是目录。
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// EnsureDir 确保目录存在。
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// ReadText 读取文本文件内容。
func ReadText(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteText 写入文本文件，并在需要时创建父目录。
func WriteText(path string, content string, perm os.FileMode) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), perm)
}

// CopyFile 复制文件内容和权限，并在需要时创建目标父目录。
func CopyFile(src string, dst string) (err error) {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source must be a regular file")
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// FileSize 返回文件大小。
func FileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if !info.Mode().IsRegular() {
		return 0, fmt.Errorf("path must be a regular file")
	}
	return info.Size(), nil
}

// Ext 返回路径扩展名。
func Ext(path string) string {
	return filepath.Ext(path)
}

// BaseName 返回路径最后一段文件名。
func BaseName(path string) string {
	return filepath.Base(path)
}
