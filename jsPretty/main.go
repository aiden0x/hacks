package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ditashi/jsbeautifier-go/jsbeautifier"
)

func processFile(path, outDir string) {
	if filepath.Ext(path) != ".js" {
		fmt.Fprintf(os.Stderr, "[!] skipping '%s' : not a .js file\n", path)
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if !info.Mode().IsRegular() {
		fmt.Fprintf(os.Stderr, "[!] skipping '%s' : not a regular file\n", path)
		return
	}
	
	beautified := jsbeautifier.BeautifyFile(path, jsbeautifier.DefaultOptions())
	if beautified == nil {
		fmt.Fprintf(os.Stderr, "[!] Failed to beautify %s: invalid javascript of file error\n", path)
		return
	}

	// Write to output dir
	outPath := filepath.Join(outDir, filepath.Base(path))
	if err := os.WriteFile(outPath, []byte(*beautified), info.Mode().Perm()); err != nil {
		log.Fatal(err)
		return
	}
}

func main() {
	outDir := "beautified"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal(err)
		return
	}

	inputChan := make(chan string, 100)
	var wg sync.WaitGroup

	// Handle Input
	sc := bufio.NewScanner(os.Stdin)
	go func() {
		defer close(inputChan)
		// Single File from CLI
		if len(os.Args) > 1 {
			inputChan <- os.Args[1]
		} else {
			// Multiple Files from Stdin
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
				processFile(line, outDir)
			}
		}()
	}

	// Wait for all workers to finish
	wg.Wait()

	if err := sc.Err(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println("Done :)")
}
