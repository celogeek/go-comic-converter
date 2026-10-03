package utils

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

func ShortHexToLongHexColor(s string) string {
	if len(s) != 3 {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(s[0:1] + s[0:1] + s[1:2] + s[1:2] + s[2:3] + s[2:3])
}

func HexToColor(s string) color.RGBA {
	s = ShortHexToLongHexColor(s)
	c := color.RGBA{A: 255}
	if len(s) != 6 {
		return c
	}
	c.R = colorHexToUint8(s[0:2])
	c.G = colorHexToUint8(s[2:4])
	c.B = colorHexToUint8(s[4:6])
	return c
}

func colorHexToUint8(s string) uint8 {
	d, _ := strconv.ParseUint(s, 16, 64)
	return uint8(d)
}

func StyleColor(grayScale bool, grayScaleMode int, s string) color.Color {
	c := HexToColor(s)
	if !grayScale {
		return c
	}

	y := RGBToGray(grayScaleMode, float32(c.R)/255, float32(c.G)/255, float32(c.B)/255)
	return color.Gray{Y: uint8(y*255 + 0.5)}
}

func ColorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("%02X%02X%02X", r>>8, g>>8, b>>8)
}

func RGBToGray(mode int, r, g, b float32) (y float32) {
	switch mode {
	case 1: // average
		y = (r + g + b) / 3
	case 2: // luminance (BT.709)
		y = 0.2126*r + 0.7152*g + 0.0722*b
	default: // BT.601 (standard Go)
		y = 0.299*r + 0.587*g + 0.114*b
	}
	return y
}
