package utils

import "fmt"

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
