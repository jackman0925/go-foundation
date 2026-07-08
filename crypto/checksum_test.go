package crypto

import "testing"

func TestMapChecksumMD5IsOrderIndependent(t *testing.T) {
	left := map[string]any{"name": "张三", "age": 30, "email": "test@example.com"}
	right := map[string]any{"email": "test@example.com", "age": 30, "name": "张三"}

	leftSum := MapChecksumMD5(left)
	rightSum := MapChecksumMD5(right)

	if leftSum != rightSum {
		t.Fatalf("expected same checksum, got %s and %s", leftSum, rightSum)
	}
	if len(leftSum) != 32 {
		t.Fatalf("expected md5 hex length 32, got %d", len(leftSum))
	}
}

func TestMapChecksumMD5ChangesWithContent(t *testing.T) {
	left := MapChecksumMD5(map[string]any{"name": "张三"})
	right := MapChecksumMD5(map[string]any{"name": "李四"})

	if left == right {
		t.Fatal("expected different content to produce different checksum")
	}
}
