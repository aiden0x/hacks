package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

type Resolver struct {
	Cache map[string]bool
	count int
}

var randStr string

func randString(n int) string {
	chars := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	out := bytes.Buffer{}

	for range make([]struct{}, n){
		out.WriteByte(chars[rand.Intn(len(chars))])
	}

	return out.String()
}

func (r *Resolver) isWildcard(name string) bool {
	if v, ok := r.Cache[name]; ok {
		return v
	}

	check := fmt.Sprintf("%s.%s", randStr, name)
	_, err := net.LookupHost(check)
	r.count++
	r.Cache[name] = err == nil
	return err == nil
}

func (r *Resolver) containsWildcard(name string) bool {
	parts := strings.Split(name, ".")
	for i:=len(parts)-2; i>0; i-- {
		candidate := strings.Join(parts[i:], ".")
		if r.isWildcard(candidate) {
			return true
		}
	}
	return false
}

func init() {
	rand.Seed(time.Now().UnixNano())
	randStr = randString(16)
}

func main() {
	sc := bufio.NewScanner(os.Stdin)
	r := &Resolver{
		Cache: make(map[string]bool),
	}

	for sc.Scan() {
		name := strings.ToLower(sc.Text())
		
		if !r.containsWildcard(name) {
			fmt.Println(name)
		}
	}
}
