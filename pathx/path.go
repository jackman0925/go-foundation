// Package pathx 提供跨平台词法路径校验，不防御符号链接逃逸。
package pathx

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// ErrPathEscape 表示路径包含目录穿越尝试。
var ErrPathEscape = errors.New("path escapes root")

// ErrInvalidPath 表示路径格式无效或包含跨平台危险名称。
var ErrInvalidPath = errors.New("invalid path")

// NormalizeURLPath 对编码后的 URL 路径严格解码一次，返回以 / 开头的规范路径。
// 不应将结果再次 URL 解码。
func NormalizeURLPath(raw string) (string, error) {
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPath, err)
	}
	decoded = strings.ReplaceAll(decoded, "\\", "/")
	clean, err := cleanParts(strings.TrimLeft(decoded, "/"))
	if err != nil {
		return "", err
	}
	return "/" + clean, nil
}

// SanitizeRelativePath 校验已解码相对文件路径，统一分隔符但不进行 URL 解码。
func SanitizeRelativePath(raw string) (string, error) {
	raw = strings.ReplaceAll(raw, "\\", "/")
	if strings.HasPrefix(raw, "/") {
		return "", ErrInvalidPath
	}
	clean, err := cleanParts(raw)
	if err != nil {
		return "", err
	}
	if clean == "" {
		return "", ErrInvalidPath
	}
	return clean, nil
}

// ResolveUnderRoot 返回 root 内的绝对路径；空相对路径或 . 表示根目录。
// 仅作词法校验，不访问文件系统，也不阻止符号链接或并发替换逃逸。
func ResolveUnderRoot(root, relativePath string) (string, error) {
	if root == "" || strings.ContainsRune(root, 0) || !utf8.ValidString(root) {
		return "", ErrInvalidPath
	}
	base, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	relative := ""
	if relativePath != "" && relativePath != "." {
		relative, err = SanitizeRelativePath(relativePath)
		if err != nil {
			return "", err
		}
	}
	full := filepath.Join(base, filepath.FromSlash(relative))
	check, err := filepath.Rel(base, full)
	if err != nil {
		return "", err
	}
	if check == ".." || strings.HasPrefix(check, ".."+string(filepath.Separator)) || filepath.IsAbs(check) {
		return "", ErrPathEscape
	}
	return full, nil
}

func cleanParts(raw string) (string, error) {
	if !utf8.ValidString(raw) {
		return "", ErrInvalidPath
	}
	parts := strings.Split(raw, "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == ".." {
			return "", ErrPathEscape
		}
		if part == "" || part == "." {
			continue
		}
		if strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") || strings.ContainsAny(part, "<>:\"|?*") {
			return "", ErrInvalidPath
		}
		for _, r := range part {
			if r < 32 || r == 127 {
				return "", ErrInvalidPath
			}
		}
		// Windows 设备名即使附带扩展名仍然具有特殊含义，统一跨平台拒绝。
		stem := strings.ToUpper(strings.TrimRight(strings.SplitN(part, ".", 2)[0], " "))
		switch stem {
		case "CON", "PRN", "AUX", "NUL", "CONIN$", "CONOUT$":
			return "", ErrInvalidPath
		}
		if strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT") {
			suffix := stem[3:]
			if len([]rune(suffix)) == 1 && strings.Contains("123456789¹²³", suffix) {
				return "", ErrInvalidPath
			}
		}
		result = append(result, part)
	}
	return strings.Join(result, "/"), nil
}
