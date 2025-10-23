package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
)

type urlCheck func(*url.URL) bool

func qsCheck(k, v string) bool {
	k = strings.ToLower(k)
	v = strings.ToLower(v)

	// check for utm_referrer are rarely interesting
	if strings.HasPrefix(k, "utm_") {
		return false
	}

	// value checks
	return strings.HasPrefix(v, "http") ||
		strings.Contains(v, "{") ||
		strings.Contains(v, "[") ||
		strings.Contains(v, "/") ||
		strings.Contains(v, "\\") ||
		strings.Contains(v, "<") ||
		strings.Contains(v, "(") ||
		strings.Contains(v, "eyj") ||
		// key checks
		strings.Contains(k, "redirect") ||
		strings.Contains(k, "debug") ||
		strings.Contains(k, "password") ||
		strings.Contains(k, "passwd") ||
		strings.Contains(k, "file") ||
		strings.Contains(k, "fn") ||
		strings.Contains(k, "template") ||
		strings.Contains(k, "include") ||
		strings.Contains(k, "require") ||
		strings.Contains(k, "url") ||
		strings.Contains(k, "uri") ||
		strings.Contains(k, "src") ||
		strings.Contains(k, "href") ||
		strings.Contains(k, "func") ||
		strings.Contains(k, "callback")
}

func isBoringStaticFile(u *url.URL) bool {
	extns := []string {
		".js",
		".html",
		".htm",
		".svg",
		".eot",
		".ttf",
		".woff",
		".woff2",
		".png",
		".jpg",
		".jpeg",
		".gif",
		".ico",
	}

	p := strings.ToLower(u.EscapedPath())
	for _, e := range extns {
		if strings.HasSuffix(p, e) {
			return true
		}
	}
	return false
}

func main() {
	flag.Parse()

	checks := []urlCheck {
		// Query string
		func(u *url.URL) bool {
			interesting := 0
			for k, vv := range u.Query() {
				for _, v := range vv {
					if qsCheck(k, v) {
						interesting++
					}
				}
			}
			return interesting > 0
		},

		// extensions
		func(u *url.URL) bool {
			extns := []string {
				".php",
				".phtml",
				".asp",
				".aspx",
				".asmx",
				".ashx",
				".cgi",
				".pl",
				".json",
				".xml",
				".rb",
				".py",
				".sh",
				".yaml",
				".yml",
				".toml",
				".ini",
				".md",
				".mkd",
				".do",
				".jsp",
				".jspa",
			}

			p := strings.ToLower(u.EscapedPath())
			for _, e := range extns {
				if strings.HasSuffix(p, e) {
					return true
				}
			}
			return false
		},

		// paths
		func(u *url.URL) bool {
			p := strings.ToLower(u.EscapedPath())
			return strings.Contains(p, "ajax") ||
				strings.Contains(p, "jsonp") ||
				strings.Contains(p, "admin") ||
				strings.Contains(p, "include") ||
				strings.Contains(p, "src") ||
				strings.Contains(p, "redirect") ||
				strings.Contains(p, "proxy") ||
				strings.Contains(p, "test") ||
				strings.Contains(p, "tmp") ||
				strings.Contains(p, "temp")
		},

		// non standard port
		func(u *url.URL) bool {
			return (u.Port() != "80" && u.Port() != "443" && u.Port() != "")
		},
	}

	seen := make(map[string]bool)

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		u, err := url.Parse(sc.Text())
		if err != nil {
			continue
		}

		if isBoringStaticFile(u) {
			continue
		}

		pp := make([]string, 0)
		for p, _ := range u.Query() {
			pp = append(pp, p)
		}
		sort.Strings(pp)

		key := fmt.Sprintf("%s%s?%s", u.Hostname(), u.EscapedPath(), strings.Join(pp, "&"))

		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = true
		interesting := 0

		for _, check := range checks {
			if check(u) {
				interesting++
			}
		}

		if interesting > 0 {
			fmt.Println(sc.Text())
		}
	}
}
