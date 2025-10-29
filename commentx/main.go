package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	isHTML := flag.Bool("html", false, "Reads HTML/XML/SVG/ Templates.")
	isSCRIPT := flag.Bool("script", false, "Reads Javascript/Typescript.")
	flag.Parse()


	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
		return
	}

	content := string(input)

	if !*isHTML && !*isSCRIPT {
		flag.Usage()
		os.Exit(1)
	}

	if *isHTML {
		z := html.NewTokenizer(strings.NewReader(content))
		for {
			tt := z.Next()
			if tt == html.ErrorToken {
				break
			}

			t := z.Token()
			if t.Type == html.CommentToken {
				d := strings.ReplaceAll(t.Data, "\n", "")
				if d == "" {
					continue
				}
				fmt.Println(d)
			}
		}
	}

	if *isSCRIPT {
		// Regex for single-line comments.
		singleLineCommentRegex := regexp.MustCompile(`//.*`)
		singleLineComment := singleLineCommentRegex.FindAllString(content, -1)

		// Regex for multi-line comments.
		multiLineCommentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/`)
		multiLineComment := multiLineCommentRegex.FindAllString(content, -1)

		for _, c := range singleLineComment {
			c = strings.ReplaceAll(c, "\n", "")
			fmt.Println(c)
		}

		for _, c := range multiLineComment {
			c = strings.ReplaceAll(c, "\n", "")
			fmt.Println(c)
		}
	}
}
