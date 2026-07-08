package geox

import "testing"

func TestDistanceMeters(t *testing.T) {
	got := DistanceMeters(0, 0, 0, 1)
	if got < 111000 || got > 112000 {
		t.Fatalf("expected about 111km, got %f", got)
	}
}
