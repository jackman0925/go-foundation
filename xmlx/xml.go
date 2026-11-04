// Package xmlx 提供基于 encoding/xml 的 XML 字符串编解码。
package xmlx

import "encoding/xml"

// MarshalToString 将输入值序列化为紧凑 XML 字符串，不添加 XML 声明。
func MarshalToString(value any) (string, error) {
	content, err := xml.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// UnmarshalFromString 将 XML 字符串反序列化到目标指针，遵循 encoding/xml 的解析规则。
func UnmarshalFromString(value string, target any) error {
	return xml.Unmarshal([]byte(value), target)
}

// Pretty 将输入值序列化为两空格缩进的 XML 字符串，不添加 XML 声明。
// 输入应为结构体等可编码值，而非待格式化的原始 XML 文本。
func Pretty(value any) (string, error) {
	content, err := xml.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}
