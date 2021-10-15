package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func processName(name string) (io.ReadCloser, error) {
	if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
		//it's an URI not a file
		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		resp, err := netClient.Get(name)

		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	}
	return os.Open(name)
}

func processFile(fl io.ReadCloser) error {

	scanner := bufio.NewScanner(fl)
	for scanner.Scan() {
		//we add back the newline
		processLine(fmt.Sprintf("%s\n", scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func processLine(line string) {
	readLine := strings.NewReader(line)

	if showAll {
		doShowAll(line)
	}
	if links {
		doWeb(readLine, "a")
	}
	if images {
		doWeb(readLine, "img")
	}
	if oog {
		doOOG(line)
	}
	if krad {
		doKRAD(line)
	}

	if lower {
		doCase(line, "lower")
	}
	if upper {
		doCase(line, "upper")
	}

}

func doShowAll(s string) {
	s = strings.Replace(s, "\r", "^M", -1)
	s = strings.Replace(s, "\n", "^J", -1)
	s = strings.Replace(s, "\t", "^I", -1)

	fmt.Fprint(os.Stdout, s)
}

func doCase(s, acase string) {
	switch acase {
	case "lower":
		fmt.Fprint(os.Stdout, strings.ToLower(s))
	case "upper":
		fmt.Fprint(os.Stdout, strings.ToUpper(s))
	default:
		fmt.Fprint(os.Stdout, s)
	}

}

func doWeb(r io.Reader, kind string) {
	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := z.Token()
			isElement := t.Data == kind
			if !isElement {
				continue
			}
			ok, url := getVal(t, kind)
			if !ok {
				continue
			}
			fmt.Fprintln(os.Stdout, url)
		}
	}
}

func getVal(t html.Token, s string) (ok bool, href string) {
	var elname string
	if s == "a" {
		elname = "href"
	} else {
		elname = "src"
	}
	for _, a := range t.Attr {
		if a.Key == elname {
			href = a.Val
			ok = true
		}
	}
	return
}

func doOOG(s string) {
	s = strings.ToUpper(s)
	repl := strings.NewReplacer(" IS ", " ", " ARE ", " ",
		" AM ", " ", " A ", " ", " AN ", " ", " THAT ", " ", " WHICH ", " ", " THE ", " ", " CAN ", " ", " OUR ", " ", " ANY ", " ", " HIS ", " ", " HERS ", " ",
		" I ", " OOG ", " IM ", " OOG ", " ME ", " OOG ", " MY ", " OOG'S ", " MINE ", " OOG'S ", " ANOTHER ", " OTHER ", " HAS ", " HAVE ", " HAD ", " HAVE ", " CANNOT ", " NOT ",
		".", "!!!", ";", "!!!", ":", "!!!", "!", "!!!", "?", "!!!", ",", "", "'", "", "`", "")
	fmt.Fprint(os.Stdout, repl.Replace(s))

}

func doKRAD(s string) {
	repl := strings.NewReplacer("0", "o", "1", "l", "a", "4", "ate", "8",
		"e", "3", "b", "6", "l", "1", "o", "0", "s", "5", "see", "C", "t", "7")
	fmt.Fprint(os.Stdout, repl.Replace(s))

}
