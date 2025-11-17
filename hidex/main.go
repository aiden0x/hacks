package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("are we scan ghosts here!\nwhere is the url bit**\nUsage: findx <url>")
		os.Exit(2)
	} 

	u := os.Args[1]
	res, err := http.Get(u)
	if err != nil {
		fmt.Printf("HTTP didn't want to work today -_- : %s\n", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	hdi := doc.Find("input[type='hidden']\n")
	c := hdi.Length()
	if c == 0 {
		fmt.Println("shit, no juicy hidden inputs found :(")
	}

	fmt.Print("we got some work today :)\n\n")

	hdi.Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		val, _ := s.Attr("value")
		fmt.Printf("-> name=\"%s\" value=\"%s\"\n", name, val)
	})
}
