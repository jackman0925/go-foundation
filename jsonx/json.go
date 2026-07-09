package jsonx

import "encoding/json"

// MarshalToString 将输入值序列化为紧凑 JSON 字符串。
func MarshalToString(value any) (string, error) {
	content, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// MustToString 将输入值序列化为 JSON，失败时 panic；仅限测试或初始化阶段使用。
func MustToString(value any) string {
	text, err := MarshalToString(value)
	if err != nil {
		panic(err)
	}
	return text
}

// UnmarshalFromString 将 JSON 字符串反序列化到目标对象。
func UnmarshalFromString(value string, target any) error {
	return json.Unmarshal([]byte(value), target)
}

// Pretty 将输入值序列化为带缩进的 JSON 字符串。
func Pretty(value any) (string, error) {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}
