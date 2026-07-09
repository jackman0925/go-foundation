package colorx

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// HexToRGBA 将 #RRGGBB 或 RRGGBB 字符串转换为 color.RGBA。
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

// RGBAToHex 将 color.RGBA 转换为 #RRGGBB 字符串。
func RGBAToHex(value color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", value.R, value.G, value.B)
}
