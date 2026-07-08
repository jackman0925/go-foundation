package jsonx

import "encoding/json"

// MarshalToString marshals value into a compact JSON string.
func MarshalToString(value any) (string, error) {
	content, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// MustToString marshals value into JSON and panics on error.
func MustToString(value any) string {
	text, err := MarshalToString(value)
	if err != nil {
		panic(err)
	}
	return text
}

// UnmarshalFromString unmarshals a JSON string into target.
func UnmarshalFromString(value string, target any) error {
	return json.Unmarshal([]byte(value), target)
}

// Pretty marshals value into indented JSON.
func Pretty(value any) (string, error) {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
}
