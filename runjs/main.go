package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/chromedp/chromedp"
)

func main() {
	js := flag.String("js", "", "Js to run on the target.")
	flag.Parse()

	if *js == "" {
		flag.Usage()
		os.Exit(2)
	}

	urls := make(chan string, 100)
	var wg sync.WaitGroup

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		u := sc.Text()
		if u == "" {
			continue
		}
		urls <- u
	}
	close(urls)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range urls {
				ctx, cancel := chromedp.NewContext(context.Background())

				var res string
				err := chromedp.Run(ctx,
					chromedp.Navigate(u),
					chromedp.Evaluate(*js, &res),
				)
				cancel()

				if err != nil {
					continue
				}

				fmt.Printf("[%s] : [%s]\n", u, res)
			}
		}()
	}

	wg.Wait()
}
