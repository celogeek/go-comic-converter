package utils

import (
	"fmt"
	"path/filepath"
	"testing"
)

func ExampleFloatToString() {
	fmt.Println("test=" + FloatToString(3.14151617, 0) + "=")
	fmt.Println("test=" + FloatToString(3.14151617, 1) + "=")
	fmt.Println("test=" + FloatToString(3.14151617, 2) + "=")
	fmt.Println("test=" + FloatToString(3.14151617, 4) + "=")
	// Output: test=3=
	// test=3.1=
	// test=3.14=
	// test=3.1415=
}

func ExampleIntToString() {
	fmt.Println("test=" + IntToString(159) + "=")
	// Output: test=159=
}

func ExampleNumberOfDigits() {
	fmt.Println(NumberOfDigits(0))
	fmt.Println(NumberOfDigits(4))
	fmt.Println(NumberOfDigits(10))
	fmt.Println(NumberOfDigits(14))
	fmt.Println(NumberOfDigits(256))
	fmt.Println(NumberOfDigits(2004))
	fmt.Println(NumberOfDigits(-5))
	fmt.Println(NumberOfDigits(-10))
	fmt.Println(NumberOfDigits(-12))

	// Output: 1
	// 1
	// 2
	// 2
	// 3
	// 4
	// 2
	// 3
	// 3
}

func ExampleFormatNumberOfDigits() {
	fmt.Println(FormatNumberOfDigits(0))
	fmt.Println(FormatNumberOfDigits(4))
	fmt.Println(FormatNumberOfDigits(10))
	fmt.Println(FormatNumberOfDigits(14))
	fmt.Println(FormatNumberOfDigits(256))
	fmt.Println(FormatNumberOfDigits(2004))
	fmt.Println(FormatNumberOfDigits(-5))
	fmt.Println(FormatNumberOfDigits(-10))
	fmt.Println(FormatNumberOfDigits(-12))
	// Output: %01d
	// %01d
	// %02d
	// %02d
	// %03d
	// %04d
	// %02d
	// %03d
	// %03d
}

func ExamplePartLabel() {
	fmt.Println(PartLabel(1, 1))
	fmt.Println(PartLabel(2, 9))
	fmt.Println(PartLabel(1, 10))
	fmt.Println(PartLabel(10, 12))
	fmt.Println(PartLabel(7, 100))
	// Output: Part 1 of 1
	// Part 2 of 9
	// Part 01 of 10
	// Part 10 of 12
	// Part 007 of 100
}

func TestSplitPath(t *testing.T) {
	sep := string(filepath.Separator)
	for _, c := range []struct{ root, path, dir, name string }{
		{"", "a.jpg", "", "a.jpg"},
		{"", "Chap1/a.jpg", "Chap1" + sep, "a.jpg"},
		{"", "./Chap1//a.jpg", "Chap1" + sep, "a.jpg"},
		{".", "a.jpg", "", "a.jpg"},
		{".", "Chap1/a.jpg", "Chap1" + sep, "a.jpg"},
		{".", "Chap1/sub/a.jpg", "Chap1" + sep + "sub" + sep, "a.jpg"},
		{"in", "in/a.jpg", "", "a.jpg"},
		{"in", "in/Chap1/a.jpg", "Chap1" + sep, "a.jpg"},
		{"/abs/in", "/abs/in/a/b/c.jpg", "a" + sep + "b" + sep, "c.jpg"},
	} {
		dir, name := SplitPath(c.root, c.path)
		if dir != c.dir || name != c.name {
			t.Errorf("SplitPath(%q, %q) = (%q, %q), want (%q, %q)", c.root, c.path, dir, name, c.dir, c.name)
		}
	}
}
