package compressx

import "testing"

func TestGzipRoundTrip(t *testing.T) {
	input := []byte("hello hello hello")

	compressed, err := Gzip(input)
	if err != nil {
		t.Fatalf("Gzip returned error: %v", err)
	}
	output, err := Gunzip(compressed)
	if err != nil {
		t.Fatalf("Gunzip returned error: %v", err)
	}

	if string(output) != string(input) {
		t.Fatalf("expected %q, got %q", input, output)
	}
}

func TestGunzipRejectsInvalidData(t *testing.T) {
	if _, err := Gunzip([]byte("not gzip")); err == nil {
		t.Fatal("expected invalid gzip error")
	}
}

func TestGzipEmptyInputRoundTrip(t *testing.T) {
	compressed, err := Gzip(nil)
	if err != nil {
		t.Fatalf("Gzip returned error: %v", err)
	}
	output, err := Gunzip(compressed)
	if err != nil {
		t.Fatalf("Gunzip returned error: %v", err)
	}
	if len(output) != 0 {
		t.Fatalf("expected empty output, got %q", output)
	}
}
