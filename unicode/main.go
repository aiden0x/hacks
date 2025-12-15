package main

import (
	"fmt"
	"net/url"
	"os"
	"unicode/utf8"
)

func fw(r rune) rune {
	// ASCII space
	if r == 0x20 {
		return 0x3000
	}

	// Printable ASCII range
	if r >= 0x21 && r <= 0x7E {
		return r + 0xFEE0
	}

	return r
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fwu <char>")
		return
	}

	char := os.Args[1]
	fww := ""

	for len(char) > 0 {
		r, size := utf8.DecodeLastRuneInString(char)
		fww += string(fw(r))
		char = char[size:]
	}

	out := url.QueryEscape(fww)
	fmt.Println(out)
}
