package pagination

import "testing"

func TestTotalPages(t *testing.T) {
	tests := []struct {
		totalRecords int64
		pageSize     int64
		expected     int64
	}{
		{0, 20, 0},
		{40, 20, 2},
		{41, 20, 3},
		{10, 0, 0},
		{10, -1, 0},
	}

	for _, tt := range tests {
		if got := TotalPages(tt.totalRecords, tt.pageSize); got != tt.expected {
			t.Fatalf("TotalPages(%d, %d) = %d, want %d", tt.totalRecords, tt.pageSize, got, tt.expected)
		}
	}
}
