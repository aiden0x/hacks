package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var commonReflectiveHeaders = []string{
	"Host",
	"User-Agent",
	"Referer",
	"Origin",
	"X-Forwarded-For",
	"X-Forwarded-Host",
	"X-Real-IP",
	"Accept-Language",
}

// this function loops in the header Set-Cookie
// to search for index '=' in the Set-Cookie value
// if it's found it returns the value of every strig
// before the index '=' which is the cookie name
// in the end it will extract all the cookie names.
func getCookiesNames(header http.Header) []string {
	seen := make(map[string]bool)
	var names []string
	for _, cookieStr := range header["Set-Cookie"] {
		if eqIndx := strings.Index(cookieStr, "="); eqIndx > 0 {
			name := strings.TrimSpace(cookieStr[:eqIndx])
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	return names
}

// this function checks if the response body contains
// the value in the query parameters and apply a regex
// to return that value (if it's exists in the response)
// with the 6 chars in the left and 6 chars of the right of it.
func checkAndPrint(u *url.URL, body, label, searchStr string) {
	if !strings.Contains(body, searchStr) {
		return
	}
	re, err := regexp.Compile("(.{0,6}?" + regexp.QuoteMeta(searchStr) + ".{0,6}?)")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[!] Regex compile error for %s: %v\n", searchStr, err)
		return
	}
	matches := re.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		fmt.Printf("[+] [%s] %s Reflected in response body (...%s...)\n", u, label, strings.TrimSpace(m[1]))
	}
}

// this function loops in the URL queries
// then loops in the values of that queries
// params, checking for values that it's length
// less that 4 chars (reduce false positive)
// then check if this values reflects in the response.
func checkOrgQuery(u *url.URL, body string) {
	for k, vv := range u.Query() {
		for _, v := range vv {
			if len(v) < 4 {
				continue
			}

			checkAndPrint(u, body, fmt.Sprintf("%s=%s", k, v), v)
		}
	}
}

// Ceck original query parameters and capture
// any Set-Cookie names for later injecion tests
func fetchBody(u *url.URL, client *http.Client, extraHeaders map[string]string, extraCookies []*http.Cookie) (string, []string, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Add Extra Headers in Headers
	// Scenario to test for injection
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	// Add Extra Cookies Names with
	// random values in Cookies Scenario
	// to test for injection
	for _, c := range extraCookies {
		req.AddCookie(c)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer res.Body.Close()
	bReader := io.LimitReader(res.Body, 10*1024*1024)
	b, err := io.ReadAll(bReader)
	if err != nil {
		return "", nil, err
	}

	cookiesNames := getCookiesNames(res.Header)

	return string(b), cookiesNames, nil
}

// this function generate random strings
// to use in testing scenarios
func genRandString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		fmt.Fprintf(os.Stderr, "[!] Random Generation Error: %v\n", err)
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func cloneURL(u *url.URL) *url.URL {
	cloned, err := url.Parse(u.String())
	if err != nil {
		return u
	}
	return cloned
}

func checkReflection(u *url.URL, body, injType, injRand string) {
	label := fmt.Sprintf("Random %s '%s'", injType, injRand)
	checkAndPrint(u, body, label, injRand)
}

func processURL(line string, client *http.Client) {
	orgURL, err := url.Parse(strings.TrimSpace(line))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[!] Invalid URL: %v\n", err)
		return
	}
	orgBody, extraCookiesNames, err := fetchBody(orgURL, client, nil, nil)
	if err == nil && orgBody != "" {
		checkOrgQuery(orgURL, orgBody)
	}

	// Random in URL path
	pRandStr := genRandString(9)
	url.PathEscape(pRandStr)
	pURL := cloneURL(orgURL)
	if pURL.Path == "" || pURL.Path == "/" {
		pURL.Path = "/"
	} else if !strings.HasSuffix(pURL.Path, "/") {
		pURL.Path += "/"
	}
	pURL.Path += pRandStr
	pBody, _, _ := fetchBody(pURL, client, nil, nil)
	if pBody != "" {
		checkReflection(pURL, pBody, "PATH SEGMENT", pRandStr)
	}

	// Random Query parameter name
	qNameRandStr := genRandString(9)
	qNameURL := cloneURL(orgURL)
	qName := qNameURL.Query()
	qName.Add(qNameRandStr, "testVal")
	qNameURL.RawQuery = qName.Encode()
	qNameBody, _, _ := fetchBody(qNameURL, client, nil, nil)
	if qNameBody != "" {
		checkReflection(qNameURL, qNameBody, "QUERY PARAMETER NAME", qNameRandStr)
	}

	// Random Query Parameter Value
	qValRandStr := genRandString(9)
	qValURL := cloneURL(orgURL)
	qVal := qValURL.Query()
	qVal.Add("testParam", qValRandStr)
	qValURL.RawQuery = qVal.Encode()
	qValBody, _, _ := fetchBody(qValURL, client, nil, nil)
	if qValBody != "" {
		checkReflection(qValURL, qValBody, "QUERY PARAMETER VALUE", qValRandStr)
	}

	// Random Injection in each common/existing Headers
	for _, hk := range commonReflectiveHeaders {
		hRand := genRandString(9)
		extraHeaders := map[string]string{
			hk: hRand,
		}
		hBody, _, _ := fetchBody(orgURL, client, extraHeaders, nil)
		if hBody != "" {
			checkReflection(orgURL, hBody, "HEADER", hRand)
		}
	}

	// Random Injection into each existing cookie names from base response
	for _, cn := range extraCookiesNames {
		cRand := genRandString(9)
		cookies := []*http.Cookie{
			{
				Name:  cn,
				Value: cRand,
			},
		}
		cBody, _, _ := fetchBody(orgURL, client, nil, cookies)
		if cBody != "" {
			checkReflection(orgURL, cBody, "COOKIE", cRand)
		}
	}

}

func main() {
	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
		},
	}

	inputChan := make(chan string, 100)
	var wg sync.WaitGroup

	// Handle Input
	sc := bufio.NewScanner(os.Stdin)
	go func() {
		defer close(inputChan)
		// Single URL form CLI
		if len(os.Args) > 1 {
			inputChan <- strings.TrimSpace(os.Args[1])
		} else {
			// Multiple URLs from stdin
			for sc.Scan() {
				line := strings.TrimSpace(sc.Text())
				if line != "" {
					inputChan <- line
				}
			}
		}
	}()

	// Start Worker Pool
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range inputChan {
				processURL(line, client)
			}
		}()
	}

	// wait for all workers to finish
	wg.Wait()

	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[!] Scanner Error %v\n", err)
		return
	}
}
