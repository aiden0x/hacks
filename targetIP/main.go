package main

import (
	"fmt"
	"os"

	"github.com/miekg/dns"
)

func main() {
	target := os.Args[1]

	var msg dns.Msg
	fqdn := dns.Fqdn(target)
	msg.SetQuestion(fqdn, dns.TypeA)

	in, err := dns.Exchange(&msg, "8.8.8.8:53")
	if err != nil {
		panic(err)
	}
	
	if len(in.Answer) < 1 {
		fmt.Println("No Record")
		return
	}

	for _, aa := range in.Answer {
		if a, ok := aa.(*dns.A); ok {
			fmt.Println(a.A)
		}
	}
}
