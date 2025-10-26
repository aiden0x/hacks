package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func txt2bin(line string) string {
	bytes := []byte(line)
	var builder strings.Builder

	for i, b := range bytes {
		binStr := fmt.Sprintf("%08b", b)
		builder.WriteString(binStr)

		if i < len(bytes)-1 {
			builder.WriteString(" ")
		}
	}

	return builder.String()
}

func bin2txt(line string) (string, error) {
	line = strings.ReplaceAll(line, " ", "")

	for _, c := range line {
		if c != '0' && c != '1' {
			return "", errors.New("invalid binary input")
		}
	}
	if len(line)%8 > 0 {
		return "", errors.New("length must be multiple of 8")
	}

	var builder strings.Builder
	for i := 0; i < len(line); i+=8 {
		// Extract 8-bit chunk
		chunk := line[i:i+8]
		// Parse chunk as binary to get byte value
		n, err := strconv.ParseInt(chunk, 2, 8)
		if err != nil {
			log.Fatal(err)
		}
		builder.WriteByte(byte(n))
	}

	return builder.String(), nil
}

func main() {
	encode := flag.Bool("e", false, "Encode a text to binary (default false.)")
	flag.Parse()

	var input []string
	if flag.NArg() > 0 {
		input = append(input, flag.Arg(0))
	} else {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			l := strings.TrimSpace(sc.Text())
			if l != "" {
				input = append(input, l)
			}
		}
		if err := sc.Err(); err != nil {
			log.Fatal(err)
		}
	}

	for _, l := range input {
		var res string
		var err error
		if *encode {
			res = txt2bin(l)
		} else {
			res, err = bin2txt(l)
			if err != nil {
				continue
			}
		}
		fmt.Println(res)
	}
}
