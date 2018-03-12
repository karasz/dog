package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func getHTTP(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	return
}

func processURLS(s []byte) []string {
	result := make([]string, 0)
	e := bytes.NewReader(s)
	z := html.NewTokenizer(e)
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			return result
		case tt == html.StartTagToken:
			t := z.Token()
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}
			ok, url := getHref(t)
			if !ok {
				continue
			}
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				result = append(result, url)
			}
		}
	}
	return result
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("missing argument, please give an url")
		os.Exit(1)
	}

	url := os.Args[1]

	b, err := getHTTP(url)
	if err != nil {
		fmt.Println(err)
	}
	rezult := processURLS(b)
	for _, l := range rezult {
		fmt.Println(l)
	}
}
