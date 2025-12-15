package main

import (
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"html"
	"log"
	"net/url"
)

func main() {
	etype := flag.String("type", "", "Encoding Type.")
	encode := flag.Bool("e", false, "Encode.")
	input := flag.String("i", "", "Input.")
	flag.Parse()

	if *encode {
		if *etype == "url" {
			e := url.QueryEscape(*input)
			fmt.Println(e)
		}
		if *etype == "html" {
			e := html.EscapeString(*input)
			fmt.Println(e)
		}
		if *etype == "base64" {
			e := base64.StdEncoding.EncodeToString([]byte(*input))
			fmt.Println(e)
		}
		if *etype == "hex" {
			e := hex.EncodeToString([]byte(*input))
			fmt.Println(e)
		}
		if *etype == "ascii" {
			fmt.Println([]byte(*input))
		}
	} else {
		if *etype == "url" {
			d, err := url.QueryUnescape(*input)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(d)
		}
		if *etype == "html" {
			d := html.UnescapeString(*input)
			fmt.Println(d)
		}
		if *etype == "base64" {
			d, err := base64.StdEncoding.DecodeString(*input)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(d))
		}
		if *etype == "hex" {
			d, err := hex.DecodeString(*input)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(d))
		}
	}

}
