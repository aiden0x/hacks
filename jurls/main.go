package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func isJSorJSON(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.Contains(ct, "javascript") ||
		strings.Contains(ct, "json") ||
		strings.Contains(ct, "ecmascript") ||
		strings.Contains(ct, "jsscript")
}

func resolveURL(base *url.URL, href string) (string, error) {
	parsed, err := url.Parse(href)
	if err != nil {
		return "", nil
	}
	resolved := base.ResolveReference(parsed)
	return resolved.String(), nil
}

func fetchURL(target string, client *http.Client, results chan<- string) {
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		fmt.Printf("[!] Request Error: %v\n", err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[!] Fetch Error: %v\n", err)
		return
	}
	defer res.Body.Close()

	ct := res.Header.Get("Content-Type")
	base := res.Request.URL

	if isJSorJSON(ct) {
		results <- base.String()
		return
	}

	if !strings.Contains(strings.ToLower(ct), "html") {
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Printf("[!] HTML Parse Error: %v\n", err)
		return
	}

	// Extract <script src="...">
	doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists && src != "" {
			if resolved, err := resolveURL(base, src); err == nil {
				results <- resolved
			}
		}
	})

	// Extract <link> tag with JSON or JavaScript
	doc.Find("link[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		linkType, _ := s.Attr("link")
		rel, _ := s.Attr("rel")
		as, _ := s.Attr("as")

		if !exists || href == "" {
			return
		}
		// Check for JSON links
		if strings.Contains(strings.ToLower(linkType), "json") {
			if resolved, err := resolveURL(base, href); err == nil {
				results <- resolved
			}
			return
		}
		// Check for perloaded scripts
		if strings.Contains(strings.ToLower(rel), "perloaded") && strings.Contains(strings.ToLower(as), "script") {
			if resolved, err := resolveURL(base, href); err == nil {
				results <- resolved
			}
			return
		}
		// Check for module perload
		if strings.Contains(strings.ToLower(rel), "moduleperload") {
			if resolved, err := resolveURL(base, href); err == nil {
				results <- resolved
			}
		}
		// Extract JSON-LD and empedded JSON
		doc.Find("script[type]").Each(func(i int, s *goquery.Selection) {
			scriptType, exists := s.Attr("type")
			if exists && strings.Contains(strings.ToLower(scriptType), "json") {
				// These are inline, but we can note the page contains them
				// Optionally: extract and analyze inline JSON-LD data
				jsonContent := s.Text()
				fmt.Fprintf(os.Stdout, "[+] [INLINE JSON on %s]\n%s\n", base.String(), jsonContent)
			}
		})
	})
}

func main() {
	client := &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			TLSHandshakeTimeout: time.Second * 30,
			ResponseHeaderTimeout: time.Second * 60,
			ForceAttemptHTTP2: false,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects.")
			}
			return nil
		},
	}

	results := make(chan string, 100)
	inputChan := make(chan string, 100)
	var wg sync.WaitGroup

	// Handle Input
	go func() {
		defer close(inputChan)
		// Single URL from CLI
		if len(os.Args) > 1 {
			inputChan <- strings.TrimSpace(os.Args[1])
		} else {
			// Multiple URLs from stdin
			sc := bufio.NewScanner(os.Stdin)
			for sc.Scan() {
				line := strings.TrimSpace(sc.Text())
				if line != "" {
					inputChan <- line
				}
			}
			if err := sc.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "[!] Scanner Error: %v\n", err)
			}
		}
	}()

	// Worker Pool
	for range make([]struct{}, 10){
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range inputChan {
				fetchURL(target, client, results)
			}
		}()
	}

	// Close results when all workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	seen := make(map[string]bool)
	for url := range results {
		if !seen[url] {
			seen[url] = true
			fmt.Println(url)
		}
	}
}
