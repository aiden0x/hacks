package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var extFilter = `xls|xml|xlsx|json|pdf|sql|doc|docx|pptx|txt|git|zip|tar\.gz|tgz|bak|7z|rar|log|cache|secret|db|backup|yml|gz|config|csv|yaml|md|md5|exe|dll|bin|ini|bat|sh|tar|deb|rpm|iso|img|env|apk|msi|dmg|tmp|crt|pem|key|pub|asc`

func buildURL(domain, fsc, ext string, inSubs, aext bool) string {
	base := domain
	if inSubs {
		base = "*." + domain
	}
	url := fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=%s/*&collapse=urlkey&output=text&fl=original,statuscode", base)
	if fsc != "" {
		url += fmt.Sprintf("&filter=statuscode:(%s)", fsc)
	}
	if aext {
		url += fmt.Sprintf("&filter=original:.*.(%s)$", extFilter)
	}
	if ext != "" {
		url += fmt.Sprintf("&filter=original:.*.%s$", ext)
	}
	return url
}

func fetchCDX(url string) error {
	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			MaxIdleConns: 5,
			MaxIdleConnsPerHost: 20,
		},
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	buf := &strings.Builder{}
	if _, err = io.Copy(buf, res.Body); err != nil {
		return err
	}
	fmt.Print(buf.String())
	return nil
}

func main() {
	domain := flag.String("d", "", "Target domain.")
	inSubs := flag.Bool("s", false, "Include subdomains in searching URLs.")
	fsc := flag.String("fsc", "", "Filter Status code you don't want to return it.")
	ext := flag.String("ext", "", "Specific extentions.")
	allext := flag.Bool("allext", false, "Match all extensions.")
	flag.Parse()

	if *domain == "" {
		flag.Usage()
		fmt.Printf("\nsensitive file extensions [%s]\n\n", extFilter)
		os.Exit(2)
	}

	filter := *fsc
	if strings.Contains(filter, ",") {
		filter = strings.ReplaceAll(filter, ",", "|")
	}
	url := buildURL(*domain, filter, *ext, *inSubs, *allext)
	//println(url)
	err := fetchCDX(url)
	if err != nil {
		fmt.Printf("[!] Error Fetching CDX: %s", err)
	}
}
