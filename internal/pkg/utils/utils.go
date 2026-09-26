package utils

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
)

func Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func Fatalf(format string, args ...interface{}) {
	Printf(format, args...)
	os.Exit(1)
}

func Println(a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}

func Fatalln(a ...interface{}) {
	Println(a...)
	os.Exit(1)
}

func IntToString(i int) string {
	return strconv.FormatInt(int64(i), 10)
}

func FloatToString(f float64, precision int) string {
	return strconv.FormatFloat(f, 'f', precision, 64)
}

func BoolToString(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

func BoolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func BoolStringToInt(s string) int {
	return BoolToInt(s == "true" || s == "1")
}

func NumberOfDigits(i int) int {
	x, count := 10, 1
	if i < 0 {
		i = -i
		count++
	}
	for ; x <= i; count++ {
		x *= 10
	}
	return count
}

func FormatNumberOfDigits(i int) string {
	return "%0" + IntToString(NumberOfDigits(i)) + "d"
}

func HexToColor(s string) color.Color {
	c := color.RGBA{G: 0, B: 0, A: 255}
	c.R = colorHexToUint8(s[0:1])
	c.G = colorHexToUint8(s[1:2])
	c.B = colorHexToUint8(s[2:3])
	return c
}

func StyleColor(grayScale bool, grayScaleMode int, s string) string {
	if grayScale {
		return HexToGrayHex(grayScaleMode, s)
	}
	return s
}

func HexToGrayHex(mode int, s string) string {
	r := colorHexToUint8(s[0:1])
	g := colorHexToUint8(s[1:2])
	b := colorHexToUint8(s[2:3])
	return strings.Repeat(
		strings.ToUpper(
			strconv.FormatUint(
				uint64(RGBToGray(mode, float32(r), float32(g), float32(b)))%16,
				16,
			),
		),
		3,
	)
}

func colorHexToUint8(s string) uint8 {
	d, _ := strconv.ParseUint(strings.Repeat(s, 2), 16, 64)
	return uint8(d)
}

func RGBToGray(mode int, r, g, b float32) float32 {
	switch mode {
	case 1: // average
		return (r + b + g) / 3
	case 2: // luminance
		return 0.2126*r + 0.7152*g + 0.0722*b
	default:
		return 0.299*r + 0.587*g + 0.114*b
	}
}
