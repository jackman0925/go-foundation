package stringx

import "testing"

func TestJoinInt64(t *testing.T) {
	got := JoinInt64([]int64{1, 20, 300}, ",")
	if got != "1,20,300" {
		t.Fatalf("unexpected joined text: %q", got)
	}
}
