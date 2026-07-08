package colorx

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// HexToRGBA converts a #RRGGBB or RRGGBB string into color.RGBA.
func HexToRGBA(value string) (color.RGBA, error) {
	value = strings.TrimPrefix(value, "#")
	if len(value) != 6 {
		return color.RGBA{}, fmt.Errorf("hex color must have 6 characters")
	}

	parsed, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		return color.RGBA{}, err
	}

	return color.RGBA{
		R: uint8(parsed >> 16 & 0xFF),
		G: uint8(parsed >> 8 & 0xFF),
		B: uint8(parsed & 0xFF),
		A: 255,
	}, nil
}

// RGBAToHex converts color.RGBA into a #RRGGBB string.
func RGBAToHex(value color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", value.R, value.G, value.B)
}
