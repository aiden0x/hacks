package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

func perms(raw string) ([]string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	fixed := []string{
		"null",
		"https://evil.com",
		"http://evil.com",
	}

	patterns := []string{
		"https://%s.evil.com",
		"https://%sevil.com",
		"https://evil.com.%s",
		"https://evil.com%s",
		"https://%s@evil.com",
		"https://%s_.evil.com",
		"https://%s.xyz",
	}

	for i, p := range patterns {
		patterns[i] = fmt.Sprintf(p, u.Hostname())
	}

	return append(fixed, patterns...), nil
}

func testOrigin(u string, c *http.Client, d time.Duration) {
	pp, err := perms(u)
	if err != nil {
		log.Fatal(err)
		return
	}

	methods := []string{
		"GET",
		"OPTION",
	}
	for _, p := range pp {
		for _, m := range methods {
			req, err := http.NewRequest(m, u, nil)
			if err != nil {
				return
			}
			req.Header.Set("Origin", p)
			req.Header.Set("Referer", p)
			req.Header.Set("User-Agent", "Mozilla/5.0")

			res, err := c.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error requesting %s: %s\n",u , err)
				continue
			}
			defer res.Body.Close()

			if res.StatusCode == http.StatusTooManyRequests {
				io.Copy(io.Discard, res.Body)
				res.Body.Close()
				time.Sleep(d)
				res, err = c.Do(req)
				if err != nil {
					continue
				}
			}

			io.Copy(io.Discard, res.Body)
			res.Body.Close()

			acao := res.Header.Get("Access-Control-Allow-Origin")
			acac := res.Header.Get("Access-Control-Allow-Credentials")

			if acao != "" && acao == p {
				fmt.Printf("[%s][%s] %s %s\n", u, m, p, acac)
			}
		}
	}
}

func main() {
	url := flag.String("u", "", "Target URL.")
	file := flag.String("f", "", "File containing target URLs.")
	delay := flag.Duration("d", 0, "Delay between requests (if you got rate limit) (default: false).")
	//header := flag.String("H", "", "Custom Header required in the request (authorization headers).")
	//cookie := flag.String("C", "", "Cookie requrired in the request.")
	flag.Parse()

	if *url == "" && *file == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Configure Client
	client := &http.Client{
		Timeout: time.Second * 20,
		Transport: &http.Transport{
			MaxIdleConns:    30,
			IdleConnTimeout: time.Second * 2,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   time.Second * 20,
				KeepAlive: time.Second,
			}).DialContext,
		},
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	urls := make(chan string, 100)

	// workers
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range urls {
				testOrigin(u, client, *delay)
			}
		}()
	}

	if *url != "" {
		urls <- strings.TrimSpace(*url)
	}

	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			l := strings.TrimSpace(sc.Text())
			if l == "" {
				continue
			}
			urls <- l
		}
		if err := sc.Err(); err != nil {
			log.Fatal(err)
		}
	}
	close(urls)

	// wait for workers to finish
	wg.Wait()
}
