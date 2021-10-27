package cmd

import (
	"bufio"
	h "encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/html"
)

var theString string

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
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		//TODO: we add back the newline but we need to deal with dropCR also?
		processLine(fmt.Sprintf("%s\n", scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func processLine(line string) {
	theString = line

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

	if hex {
		doHex(line)
	}

	if showEnds {
		doShowEnds(line)
	}

	if rot != 0 {
		doRot(line, rot)
	}
	if strfry {
		doStrFry(line)
	}

	fmt.Fprint(os.Stdout, theString)
}

func doShowAll(s string) {
	repl := strings.NewReplacer("\r", "^M", "\n", "^J", "\t", "^I")
	theString = repl.Replace(s)
}

func doCase(s, acase string) {
	switch acase {
	case "lower":
		theString = strings.ToLower(s)
	case "upper":
		theString = strings.ToUpper(s)
	}

}

func doWeb(r io.Reader, kind string) {
	theString = ""
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
			theString = theString + fmt.Sprintf("%s\n", url)
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
	//This implementation of OOG is bogus :-/
	s = strings.ToUpper(s)
	repl := strings.NewReplacer(" IS ", " ", " ARE ", " ",
		" AM ", " ", " A ", " ", " AN ", " ", " THAT ", " ", " WHICH ", " ", " THE ", " ", " CAN ", " ", " OUR ", " ", " ANY ", " ", " HIS ", " ", " HERS ", " ",
		" I ", " OOG ", " IM ", " OOG ", " ME ", " OOG ", " MY ", " OOG'S ", " MINE ", " OOG'S ", " ANOTHER ", " OTHER ", " HAS ", " HAVE ", " HAD ", " HAVE ", " CANNOT ", " NOT ",
		".", "!!!", ";", "!!!", ":", "!!!", "!", "!!!", "?", "!!!", ",", "", "'", "", "`", "")
	theString = repl.Replace(s)

}

func doKRAD(s string) {
	repl := strings.NewReplacer("0", "o", "1", "l", "a", "4", "ate", "8",
		"e", "3", "b", "6", "l", "1", "o", "0", "s", "5", "see", "C", "t", "7")
	theString = repl.Replace(s)

}

func doHex(s string) {
	theString = h.Dump([]byte(s))
}

func doShowEnds(s string) {
	//Windows
	s = strings.Replace(s, "\r\n", "$\r\n", -1)
	//Linux/Unix
	s = strings.Replace(s, "\n", "$\n", -1)
	//Old MacOS
	s = strings.Replace(s, "\r", "$\r", -1)

	theString = s
}

func doRot(s string, r int) {
	myRunes := []rune(s)
	for i, c := range myRunes {
		if unicode.IsLetter(c) {
			myRunes[i] = rune(int(c) + r)
		}
	}
	theString = string(myRunes)
}

func doStrFry(s string) {
	myRunes := []rune(s)
	rand.Seed(time.Now().UnixNano())

	var j int
	for i := len(myRunes) - 1; i > 0; i-- {
		j = rand.Intn(i)
		myRunes[j], myRunes[i] = myRunes[i], myRunes[j]
	}

	theString = string(myRunes)
}
