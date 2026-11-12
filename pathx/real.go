package pathx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveRealUnderRoot 解析已存在目标的符号链接，拒绝最终真实路径位于根目录外的目标。
// 根目录本身允许是符号链接，其解析后的真实目录作为边界；返回绝对真实路径。
// 根目录必须存在且是目录，目标必须存在。断链、链接循环及权限错误直接返回。
// 本函数不是安全打开文件的接口：检查期间及返回后目录仍可能被修改，调用方须保证目录树稳定。
func ResolveRealUnderRoot(root, relativePath string) (string, error) {
	// 先应用相同的词法规则，避免真实路径解析意外接受 .. 或绝对路径。
	if _, err := ResolveUnderRoot(root, relativePath); err != nil {
		return "", err
	}
	base, err := ResolveUnderRoot(root, "")
	if err != nil {
		return "", err
	}
	realRoot, err := filepath.EvalSymlinks(base)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(realRoot)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%w: root must be a directory", ErrInvalidPath)
	}
	// 后续解析使用真实根路径，避免再次经过调用方提供的根目录别名。
	candidate, err := ResolveUnderRoot(realRoot, relativePath)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(realRoot, resolved)
	if err != nil {
		// Windows 不同卷无法计算相对路径，同样属于根目录外。
		return "", fmt.Errorf("%w: %v", ErrPathEscape, err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", ErrPathEscape
	}
	return resolved, nil
}
