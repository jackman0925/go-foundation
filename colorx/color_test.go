package colorx

import (
	"image/color"
	"testing"
)

func TestHexToRGBAAndRGBAToHex(t *testing.T) {
	rgba, err := HexToRGBA("#ff7ff0")
	if err != nil {
		t.Fatalf("HexToRGBA returned error: %v", err)
	}
	if rgba != (color.RGBA{R: 255, G: 127, B: 240, A: 255}) {
		t.Fatalf("unexpected rgba: %+v", rgba)
	}
	if got := RGBAToHex(rgba); got != "#ff7ff0" {
		t.Fatalf("unexpected hex: %q", got)
	}
}

func TestHexToRGBARejectsInvalidLength(t *testing.T) {
	if _, err := HexToRGBA("#fff"); err == nil {
		t.Fatal("expected invalid length error")
	}
}
