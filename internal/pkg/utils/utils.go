package utils

import (
	"fmt"
	"os"
)

func Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func Println(a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}
