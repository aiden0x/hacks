package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	query := flag.String("q", "", "Search query (dork).")
	flag.Parse()

	if *query == "" {
		flag.Usage()
		os.Exit(2)
	}

	client := &http.Client{
		Timeout: time.Second * 20,
	}

	base := fmt.Sprintf("https://www.google.com/search?client=firefox-b-d&channel=entpr&q=%s&gl=us&hl=en", url.QueryEscape(*query))
	req, err := http.NewRequest("GET", base, nil)
	if err != nil {
		fmt.Printf("[!]HTTP Error: %s", err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[!]HTTP Error: %s\n", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Response Error: [%d]\n", res.StatusCode)
		os.Exit(1)
	}

	out := make([]string, 0)
	seen := make(map[string]struct{})

	body, _:= io.ReadAll(res.Body)
	fmt.Println(string(body))

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Printf("Response Error: [%d]\n", res.StatusCode)
		return
	}

	doc.Find("a.zReHs").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if ok {
			if _, exists := seen[href]; !exists {
				seen[href] = struct{}{}
				out = append(out, href)
			}
		}
	})

	for _, u := range out {
		fmt.Println(u)
	}
}
