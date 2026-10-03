package utils

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHexToColor(t *testing.T) {
	for i, c := range []struct {
		hex      string
		expected color.RGBA
	}{
		{"FF0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"00FF00", color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"0000FF", color.RGBA{R: 0, G: 0, B: 255, A: 255}},
		{"FFFFFF", color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"000000", color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{"001122", color.RGBA{R: 0, G: 17, B: 34, A: 255}},
		{"XXXXXX", color.RGBA{R: 0, G: 0, B: 0, A: 255}},    // if the hex string is invalid, it will return black color
		{"12345", color.RGBA{R: 0, G: 0, B: 0, A: 255}},     // if the hex string is invalid, it will return black color
		{"FFF", color.RGBA{R: 255, G: 255, B: 255, A: 255}}, // short hex string
		{"000", color.RGBA{R: 0, G: 0, B: 0, A: 255}},       // short hex string
		{"f00", color.RGBA{R: 255, G: 0, B: 0, A: 255}},     // short hex string lowercase
		{"ff0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}},  // long hex string lowercase
	} {
		t.Run(c.hex, func(t *testing.T) {
			assert.Equal(t, c.expected, HexToColor(c.hex), "test %d", i)
		})
	}
}

func TestRGBToGray(t *testing.T) {
	for i, c := range []struct {
		mode      int
		color     color.Color
		expectedY uint8
	}{
		{0, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 76},      // BT.601 (standard Go)
		{1, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 85},      // average
		{2, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 54},      // luminance (BT.709)
		{0, color.RGBA{R: 0, G: 255, B: 0, A: 255}, 150},     // BT.601 (standard Go)
		{1, color.RGBA{R: 0, G: 255, B: 0, A: 255}, 85},      // average
		{2, color.RGBA{R: 0, G: 255, B: 0, A: 255}, 182},     // luminance (BT.709)
		{0, color.RGBA{R: 0, G: 0, B: 255, A: 255}, 29},      // BT.601 (standard Go)
		{1, color.RGBA{R: 0, G: 0, B: 255, A: 255}, 85},      // average
		{2, color.RGBA{R: 0, G: 0, B: 255, A: 255}, 18},      // luminance (BT.709)
		{0, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 255}, // BT.601 (standard Go)
		{1, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 255}, // average
		{2, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 255}, // luminance (BT.709)
		{0, color.RGBA{R: 0, G: 0, B: 0, A: 255}, 0},         // BT.601 (standard Go)
		{1, color.RGBA{R: 0, G: 0, B: 0, A: 255}, 0},         // average
		{2, color.RGBA{R: 0, G: 0, B: 0, A: 255}, 0},         // luminance (BT.709)
		{0, color.RGBA{R: 128, G: 128, B: 128, A: 255}, 128}, // BT.601 (standard Go)
		{1, color.RGBA{R: 128, G: 128, B: 128, A: 255}, 128}, // average
		{2, color.RGBA{R: 128, G: 128, B: 128, A: 255}, 128}, // luminance (BT.709)
	} {
		t.Run(fmt.Sprintf("test RGBToGray %d", i), func(t *testing.T) {
			r, g, b, _ := c.color.RGBA()
			y := RGBToGray(c.mode, float32(r)/65535, float32(g)/65535, float32(b)/65535)
			assert.Equal(t, c.expectedY, uint8(y*255+0.5), "test %d", i)
		})
	}
}

func TestColorToHex(t *testing.T) {
	for i, c := range []struct {
		color    color.Color
		expected string
	}{
		{color.RGBA{R: 255, G: 0, B: 0, A: 255}, "FF0000"},
		{color.RGBA{R: 0, G: 255, B: 0, A: 255}, "00FF00"},
		{color.RGBA{R: 0, G: 0, B: 255, A: 255}, "0000FF"},
		{color.RGBA{R: 255, G: 255, B: 255, A: 255}, "FFFFFF"},
		{color.RGBA{R: 0, G: 0, B: 0, A: 255}, "000000"},
		{color.RGBA{R: 0, G: 17, B: 34, A: 255}, "001122"},
	} {
		t.Run(fmt.Sprintf("test ColorToHex %d", i), func(t *testing.T) {
			assert.Equal(t, c.expected, ColorToHex(c.color), "test %d", i)
		})
	}
}

func TestStyleColor(t *testing.T) {
	for i, c := range []struct {
		grayScale     bool
		grayScaleMode int
		hex           string
		expectedColor color.Color
		expectedStyle string
	}{
		{false, 0, "FF0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}, "FF0000"},
		{true, 0, "FF0000", color.Gray{Y: 76}, "4C4C4C"},
		{true, 1, "FF0000", color.Gray{Y: 85}, "555555"},
		{true, 2, "FF0000", color.Gray{Y: 54}, "363636"},
		{true, 0, "444444", color.Gray{Y: 68}, "444444"},
		{true, 1, "444444", color.Gray{Y: 68}, "444444"},
		{true, 2, "444444", color.Gray{Y: 68}, "444444"},
		{true, 0, "00FF00", color.Gray{Y: 150}, "969696"},
		{true, 1, "00FF00", color.Gray{Y: 85}, "555555"},
		{true, 2, "00FF00", color.Gray{Y: 182}, "B6B6B6"},
		{true, 0, "0000FF", color.Gray{Y: 29}, "1D1D1D"},
		{true, 1, "0000FF", color.Gray{Y: 85}, "555555"},
		{true, 2, "0000FF", color.Gray{Y: 18}, "121212"},
		{true, 0, "A0A0A0", color.Gray{Y: 160}, "A0A0A0"},
		{true, 1, "A0A0A0", color.Gray{Y: 160}, "A0A0A0"},
		{true, 2, "A0A0A0", color.Gray{Y: 160}, "A0A0A0"},
	} {
		t.Run(fmt.Sprintf("test StyleColor %d", i), func(t *testing.T) {
			assert.Equal(t, c.expectedColor, StyleColor(c.grayScale, c.grayScaleMode, c.hex), "test %d", i)
			assert.Equal(t, c.expectedStyle, ColorToHex(StyleColor(c.grayScale, c.grayScaleMode, c.hex)), "test %d", i)
		})
	}
}

func TestShortHexToLongHexColor(t *testing.T) {
	for i, c := range []struct {
		shortHex string
		expected string
	}{
		{"F00", "FF0000"},
		{"0F0", "00FF00"},
		{"00F", "0000FF"},
		{"FFF", "FFFFFF"},
		{"000", "000000"},
		{"123", "112233"},
		{"abc", "AABBCC"},
		{"a0b1c2", "A0B1C2"},
	} {
		t.Run(fmt.Sprintf("test ShortHexToLongHexColor %d", i), func(t *testing.T) {
			assert.Equal(t, c.expected, ShortHexToLongHexColor(c.shortHex), "test %d", i)
		})
	}
}
