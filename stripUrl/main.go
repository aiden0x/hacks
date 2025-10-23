package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func process(domain string, strip bool) {
	domain = strings.TrimSpace(domain)

	if strip {
		domain = strings.TrimPrefix(domain, "http://")
		domain = strings.TrimPrefix(domain, "https://")
		fmt.Println(domain)
	} else {
		if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
			domain = "https://" + domain
		}
		fmt.Println(domain)
	}
}

func main() {
	strip := flag.Bool("d", false, "Remove (https:// | http://) from domains")
	flag.Parse()

	args := flag.Args()

	if len(args) > 0 {
		// Single domain from ClI
		process(args[0], *strip)
	} else {
		// Multiple domains from stdin
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line != "" {
				process(line, *strip)
			}
		}
	}
}
