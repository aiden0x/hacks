package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const extensionsFilter = `original:.*\.(xls|xml|xlsx|json|pdf|sql|doc|docx|pptx|txt|zip|tar\.gz|tgz|bak|7z|rar|log|cache|secret|db|backup|yml|gz|git|config|csv|yaml|md|md5|exe|dll|bin|ini|bat|sh|tar|deb|rpm|iso|img|apk|msi|env|dmg|tmp|crt|pem|key|pub|asc)$`

func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty URL")
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		return "", errors.New("missing host")
	}
	return u.String(), nil
}

func buildWaybachURL(mode, hostname, fullPath string) (string, error) {
	base := "https://web.archive.org/cdx/search/cdx"
	values := url.Values{}

	switch mode {
	case "wildcard":
		values.Set("url", fmt.Sprintf("*.%s/*", hostname))
	case "domain":
		values.Set("url", fmt.Sprintf("%s/*", hostname))
	case "specific":
		values.Set("url", fmt.Sprintf("%s/*", fullPath))
	case "extensions":
		values.Set("url", fmt.Sprintf("*.%s/*", hostname))
		values.Set("filter", extensionsFilter)
	default:
		return "", fmt.Errorf("Unknown mode: %s", mode)
	}

	values.Set("collapse", "urlkey")
	values.Set("output", "text")
	values.Set("fl", "original")

	return fmt.Sprintf("%s?%s", base, values.Encode()), nil
}

func fetchCDX(urlStr, outPath string) error {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "wayback-cli/1.0")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("Server returns: %v", err)
	}

	buf := &strings.Builder{}
	if _, err = io.Copy(buf, res.Body); err != nil {
		return err
	}
	str := buf.String()
	fmt.Print(str)
	if outPath != "" {
		file, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err := file.WriteString(str); err != nil {
			return err
		}
		fmt.Printf("[+] Results Saved in: %s", outPath)
	}
	return err
}

func main() {
	urlStr := flag.String("u", "", "Target URL or domain")
	mode := flag.String("m", "domain", "Mode: domain | wildcard | specific | extensions")
	outFile := flag.String("o", "", "Save the results in to file")

	flag.Parse()

	if *urlStr == "" {
		fmt.Println("[-] Error: --url is required")
		flag.Usage()
		os.Exit(1)
	}

	normalizeURL, err := normalizeURL(*urlStr)
	if err != nil {
		fmt.Println("[-] Invalid URL")
		os.Exit(1)
	}

	parsed, err := url.Parse(normalizeURL)
	if err != nil {
		fmt.Println("[-] Falied to parse URL")
		os.Exit(1)
	}

	hostname := parsed.Hostname()
	fullPath := parsed.Scheme + "://" + parsed.Host + parsed.Path

	finalURL, err := buildWaybachURL(*mode, hostname, fullPath)
	if err != nil {
		fmt.Println("[-] Error:", err)
		os.Exit(1)
	}

	fmt.Print("[+] Fetching Wayback URLs...\n")
	if err := fetchCDX(finalURL, *outFile); err != nil {
		fmt.Println("[-] Fetching error", err)
		os.Exit(1)
	}

}
