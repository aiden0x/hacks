package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type SecurityTxt struct {
	Contacts      []string
	Encryption    string
	Disclosure    string
	Acknowledment string
	Comments      []string
	Errors        []string
	SourceURL     string
}

func normalizeDomain(input string) string {
	input = strings.TrimSpace(input)

	if input == "" {
		return input
	}

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		u, err := url.Parse(input)
		if err == nil && u.Host != "" {
			return u.Host
		}
	}

	input = strings.TrimSuffix(input, "/")
	return input
}

func fetchSecTxt(targetURL string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseSecTxt(raw string) SecurityTxt {
	var txt SecurityTxt

	sc := bufio.NewScanner(strings.NewReader(raw))
	lines := 0

	for sc.Scan() {
		lines++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			txt.Comments = append(txt.Comments, line)
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			txt.Errors = append(txt.Errors, fmt.Sprintf("Invalid input on line %d: %s", lines, line))
			continue
		}
		
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch key {
		case "contact":
			if validateContact(val) {
				txt.Contacts = append(txt.Contacts, val)
			} else {
				txt.Errors = append(txt.Errors, fmt.Sprintf("Invalid value %s for option 'contact' on line %d", val, lines))
			}
		case "encryption":
			if validateURL(val) {
				txt.Encryption = val
			} else {
				txt.Errors = append(txt.Errors, fmt.Sprintf("Invalid URL %s for option 'contact' on line %d", val, lines))
			}
		case "disclosure":
			if validateDisclosure(val) {
				txt.Disclosure = strings.ToLower(val)
			} else {
				txt.Errors = append(txt.Errors, fmt.Sprintf("Invalid value %s for option 'contact' on line %d", val, lines))
			}
		case "acknowledgement", "acknowledgment":
			if validateURL(val) {
				txt.Acknowledment = val
			} else {
				txt.Errors = append(txt.Errors, fmt.Sprintf("Invalid URL %s for option 'contact' on line %d", val, lines))
			}
		default:
				txt.Errors = append(txt.Errors, fmt.Sprintf("Invalid option %s on line %d", key, lines))
		}
	}

	if len(txt.Contacts) < 1 {
		txt.Errors = append(txt.Errors, "does not contain any conatct filed.")
	}

	return txt
}

func validateContact(v string) bool{
	lower := strings.ToLower(v)
	// mailto
	if strings.HasPrefix(lower, "mailto:") {
		addr := strings.TrimPrefix(lower, "mailto:")
		_, err := mail.ParseAddress(addr)
		return err == nil
	}
	// URL
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return validateURL(v)
	}
	// raw email
	if strings.Contains(v, "@") {
		_, err := mail.ParseAddress(v)
		return err == nil
	}
	// international phone number
	re := regexp.MustCompile(`^\+\d[\d\(\)\s\-]+$`)
	return re.MatchString(v)
}

func validateURL(v string) bool {
	u, err := url.ParseRequestURI(v)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func validateDisclosure(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "full", "partial", "none":
		return true
	default:
	return false
		
	}
}

func printResult(txt SecurityTxt) {
	fmt.Print("\n========== security.txt report ===========\n")
	if txt.SourceURL != "" {
		fmt.Printf("fetched from: %s\n", txt.SourceURL)
	}
	if len(txt.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range txt.Errors {
			fmt.Printf("\t%s\n", e)
		}
	}
	if len(txt.Comments) > 0 {
		fmt.Println("\nComments:")
		for _, c := range txt.Comments{
			fmt.Printf("\t%s\n", c)
		}
	}
	if len(txt.Contacts) > 0 {
		fmt.Println("\nContacts:")
		for _, c := range txt.Contacts{
			fmt.Printf("\t%s\n", c)
		}
	}
	if txt.Encryption != "" {
		fmt.Printf("\nEncryption:\n\t%s\n", txt.Encryption)
	}
	if txt.Disclosure!= "" {
		fmt.Printf("\nDisclosure:\n\t%s\n", txt.Disclosure)
	}
	if txt.Acknowledment!= "" {
		fmt.Printf("\nAcknowledment:\n\t%s\n", txt.Acknowledment)
	}
}

func main() {
	domain := flag.String("d", "", "Target domain (example.com)")
	flag.Parse()

	if *domain == "" {
		fmt.Println("Usage: sectxt -d <domain>")
		os.Exit(1)
	}

	targetHost := normalizeDomain(*domain)

	candidates := []string {
		fmt.Sprintf("https://%s/.well-knows/security.txt", targetHost),
		fmt.Sprintf("http://%s/.well-knows/security.txt", targetHost),
		fmt.Sprintf("https://%s/security.txt", targetHost),
		fmt.Sprintf("http://%s/security.txt", targetHost),
	}

	var src string
	var body string
	var fetchErr error
	for _, u := range candidates {
		body, fetchErr = fetchSecTxt(u)
		if fetchErr == nil {
			src = u
			break
		}
	}

	if fetchErr != nil {
		fmt.Fprintf(os.Stdout, "[!] Failed to fetch security.txt: %v\n", fetchErr)
		os.Exit(1)
	}

	result := parseSecTxt(body)
	result.SourceURL = src
	printResult(result)
}
