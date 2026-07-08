package compressx

import "testing"

func TestZlibRoundTrip(t *testing.T) {
	input := []byte("hello zlib")

	compressed, err := Zlib(input)
	if err != nil {
		t.Fatalf("Zlib returned error: %v", err)
	}
	output, err := Unzlib(compressed)
	if err != nil {
		t.Fatalf("Unzlib returned error: %v", err)
	}

	if string(output) != string(input) {
		t.Fatalf("expected %q, got %q", input, output)
	}
}

func TestUnzlibRejectsInvalidData(t *testing.T) {
	if _, err := Unzlib([]byte("not zlib")); err == nil {
		t.Fatal("expected invalid zlib error")
	}
}

func TestZlibEmptyInputRoundTrip(t *testing.T) {
	compressed, err := Zlib(nil)
	if err != nil {
		t.Fatalf("Zlib returned error: %v", err)
	}
	output, err := Unzlib(compressed)
	if err != nil {
		t.Fatalf("Unzlib returned error: %v", err)
	}
	if len(output) != 0 {
		t.Fatalf("expected empty output, got %q", output)
	}
}
