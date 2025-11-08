package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spaolacci/murmur3"
)

func main() {
	if len(os.Args) < 1 {
		fmt.Println("Usage: shfavico [target URL]")
		os.Exit(1)
	}

	url := os.Args[1] + "/favicon.ico"

	res, err := http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()

	dd , err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	b64 := base64.StdEncoding.EncodeToString(dd)
	hash := int32(murmur3.Sum32([]byte(b64)))

	fmt.Println(hash)
}
