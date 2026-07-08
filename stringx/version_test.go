package stringx

import "testing"

func TestCompareVersion(t *testing.T) {
	tests := []struct {
		left     string
		right    string
		expected int
	}{
		{"1.0.358_20180820090554", "1.0.358_20180820090553", 1},
		{"1.2.0", "1.2", 0},
		{"1.2.3", "1.10.0", -1},
		{"2.0", "1.9.9", 1},
	}

	for _, tt := range tests {
		t.Run(tt.left+"_"+tt.right, func(t *testing.T) {
			if got := CompareVersion(tt.left, tt.right); got != tt.expected {
				t.Fatalf("CompareVersion(%q, %q) = %d, want %d", tt.left, tt.right, got, tt.expected)
			}
		})
	}
}
