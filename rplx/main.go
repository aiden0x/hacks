package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strings"
	"time"
)

var Headers = []string{
	"Forwarded",
	"X-Forwarded",
	"X-Forwarded-By",
	"X-Forwarded-For-Original",
	"X-Forwarded-Host",
	"X-Forwarded-For",
	"X-Forwarded-Proto",
	"X-Forwarded-Server",
	"X-Real-IP",
	"X-Original-Url",
	"X-Original-Host",
	"X-Host",
	"X-Rewrite-Url",
}

func parseRawRequest(raw string) (method string, furl *url.URL, headers http.Header, body []byte, err error) {
	r := bufio.NewReader(strings.NewReader(raw))
	tp := textproto.NewReader(r)

	reqLine, err := tp.ReadLine()
	if err != nil {
		return "", nil, nil, nil, err
	}

	parts := strings.SplitN(reqLine, " ", 3)
	if len(parts) < 2 {
		return "", nil, nil, nil, fmt.Errorf("Invalid request: %s", reqLine)
	}

	method = parts[0]
	path := parts[1]

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("Can't reading headers: %s", err)
	}
	headers = http.Header(mimeHeader)

	body, err = io.ReadAll(r)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("Can't reading body: %s", err)
	}

	host := headers.Get("Host")
	if host == "" {
		fmt.Print("wow we now attacking ghosts, Host Header is empty ass whole!\n\n")
		os.Exit(1)
	} else {
		furl = &url.URL{
			Scheme: "https",
			Host:   host,
			Path:   path,
		}
	}

	return method, furl, headers, body, nil
}

func processURL(c *http.Client, rawReq []byte) {
	method, furl, headers, body, err := parseRawRequest(string(rawReq))
	if err != nil {
		return
	}
	ovvHostReq, err := http.NewRequest(method, furl.String(), bytes.NewReader(body))
	if err != nil {
		log.Print(err)
	}
	for k, vv := range headers {
		for _, v := range vv {
			ovvHostReq.Header.Set(k, v)
		}
	}
	ovvHostReq.Host = "asffkasfjksaj234234234.aiden.com"

	res, err := c.Do(ovvHostReq)
	if err != nil {
		return
	}
	fmt.Printf("[%s] Host [%d]\n", furl.String(), res.StatusCode)
	res.Body.Close()

	for _, h := range Headers {
		req, err := http.NewRequest(method, furl.String(), bytes.NewReader(body))
		if err != nil {
			log.Print(err)
		}
		for k, vv := range headers {
			for _, v := range vv {
				req.Header.Set(k, v)
			}
		}
		req.Header.Set(h, "asffkasfjksaj234234234.aiden.com")
		res, err := c.Do(req)
		if err != nil {
			continue
		}

		fmt.Printf("[%s] %s [%d]\n", furl.String(), h, res.StatusCode)
		res.Body.Close()
	}
}

func main() {
	file := flag.String("f", "", "Path to file containing raw HTTP request.")
	flag.Parse()

	if *file == "" {
		flag.Usage()
		os.Exit(2)
	}

	client := &http.Client{
		Timeout: time.Second * 20,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 10,
		},
	}

	rawReq, err := os.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	processURL(client, rawReq)
}
