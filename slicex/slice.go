package slicex

// Contains 判断切片是否包含目标值。
func Contains[T comparable](values []T, target T) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// Unique 返回去重后的切片，并保留首次出现顺序。
func Unique[T comparable](values []T) []T {
	seen := make(map[T]struct{}, len(values))
	result := make([]T, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

// Reverse 原地反转切片。
func Reverse[T any](values []T) {
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
}
