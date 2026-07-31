package utils

import (
	"fmt"
	"os"
	"strconv"
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
