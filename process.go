package main

import (
	"bufio"
	h "encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/karasz/dog/multireadcloser"
	"github.com/spf13/pflag"
	"golang.org/x/net/html"
)

const maxHTTPBodySize = 10 * 1024 * 1024 // 10 MB

// LineInfo represents information about a specific line in a text file.
// It contains the content of the line and its line number.
type LineInfo struct {
	Content    string
	LineNumber int
}

var theLineInfo LineInfo
var theflags pflag.FlagSet

//revive:disable:cognitive-complexity
func processNames(names []string) (io.ReadCloser, error) {
	// revive:enable:cognitive-complexity
	readers := make([]io.ReadCloser, 0, len(names))

	for _, name := range names {
		if name == "-" {
			// Read from standard input
			readers = append(readers, os.Stdin)
		}
		if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
			// It's an URI, not a file
			var netClient = &http.Client{
				Timeout: time.Second * 10,
			}
			resp, err := netClient.Get(name)

			if err != nil {
				return nil, err
			}

			// Check the Content-Length header
			if resp.ContentLength > maxHTTPBodySize {
				// TODO: handle it, do not just return an error
				_, _ = fmt.Printf("The resource is too large: %d bytes. What would you like to do?\n",
					resp.ContentLength)
				return nil, fmt.Errorf("resource is too large: %d bytes", resp.ContentLength)
			}

			readers = append(readers, resp.Body)
		}
		r, err := os.Open(name)
		if err != nil {
			return nil, err
		}
		readers = append(readers, r)
	}
	return multireadcloser.MultiReadCloser(readers...), nil
}
func processFiles(fl io.ReadCloser, flags pflag.FlagSet) error {
	defer fl.Close()

	theflags = flags
	reader := bufio.NewReader(fl)
	return readLines(reader)
}

func readLines(reader *bufio.Reader) error {
	lineNumber := 1
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if len(line) > 0 {
			processSingleLine(line, lineNumber)
			lineNumber++
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

func processSingleLine(line string, lineNumber int) {
	line = strings.TrimRight(line, "\n")
	processLine(LineInfo{Content: line, LineNumber: lineNumber})
}

func processLine(lineInfo LineInfo) {
	theLineInfo = lineInfo

	readLine := strings.NewReader(lineInfo.Content)

	processFlags(lineInfo.Content, readLine)

	_, _ = fmt.Fprintf(os.Stdout, "%d: %s\n", theLineInfo.LineNumber, theLineInfo.Content)
}

// revive:disable:cognitive-complexity
// revive:disable:cyclomatic
func processFlags(line string, readLine *strings.Reader) {
	// revive:enable:cognitive-complexity
	// revive:enable:cyclomatic

	boolActions := map[string]func(){
		"showAll":         func() { doShowAll(line) },
		"links":           func() { doWeb(readLine, "a") },
		"images":          func() { doWeb(readLine, "img") },
		"oog":             func() { doOOG(line) },
		"krad":            func() { doKRAD(line) },
		"lower":           func() { doCase(line, "lower") },
		"upper":           func() { doCase(line, "upper") },
		"hex":             func() { doHex(line) },
		"showEnds":        func() { doShowEnds(line) },
		"numberNonBlank":  func() { doNumberNonBlank(line) },
		"noBlanks":        func() { doNoBlanks(line) },
		"dos":             func() { doDos(line) },
		"hideNonPrinting": func() { doHideNonPrinting(line) },
		"mac":             func() { doMac(line) },
		"number":          nil,
		"squeezeBlank":    func() { doSqueezeBlank(line) },
		"strfry":          func() { doStrFry(line) },
		"showTabs":        func() { doShowTabs(line) },
		"skipTags":        func() { doSkipTags(line) },
		"translate":       func() { doTranslate(line) },
		"unix":            func() { doUnix(line) },
		"showNonPrinting": func() { doShowNonPrinting(line) },
	}

	intActions := map[string]func(int){
		"rot": func(i int) {
			doRot(line, i)
		},
		"cols": func(i int) {
			if i != 0 {
				doCols(line, i)
			}
		},
	}

	for flag, action := range boolActions {
		//	if flagValue, err := theflags.GetBool(flag); err == nil && flagValue {
		if theflags.Changed(flag) && action != nil {
			action()
		}
		//	}
	}

	for flag, action := range intActions {
		if flagValue, err := theflags.GetInt(flag); err == nil {
			if theflags.Changed(flag) && action != nil {
				action(flagValue)
			}
		}
	}
}

func doShowAll(s string) {
	s = strings.ReplaceAll(s, "\r", "^M")
	s = strings.ReplaceAll(s, "\n", "^J")
	s = strings.ReplaceAll(s, "\t", "^I")
	theLineInfo.Content = s
}

func doCase(s, acase string) {
	switch acase {
	case "lower":
		theLineInfo.Content = strings.ToLower(s)
	case "upper":
		theLineInfo.Content = strings.ToUpper(s)
	}
}

func doWeb(r io.Reader, kind string) {
	theLineInfo.Content = ""
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
			theLineInfo.Content = theLineInfo.Content + fmt.Sprintf("%s\n", url)
		}
	}
}

func getVal(t html.Token, s string) (bool, string) {
	var elname, href string
	var ok bool
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
	return ok, href
}

func doOOG(s string) {
	var oogSayDifferent = map[string]string{"I": "OOG", "IM": "OOG", "ME": "OOG",
		"MY": "OOG'S", "MINE": "OOG'S", "ANOTHER": "OTHER", "HAS": "HAVE", "HAD": "HAVE",
		"CANNOT": "NOT", " IS": "", " ARE": "", " AM": "", " A": "", " AN": "",
		" THAT": "", " WHICH": "", " THE": "", " CAN": "", " OUR": "", " ANY": "", " HIS": "", " HERS": ""}

	// OOG don't like small
	s = strings.ToUpper(s)

	// OOG don't like contractions
	s = strings.ReplaceAll(s, "'LL", " WILL")
	s = strings.ReplaceAll(s, "WON'T", "WILL NOT")
	s = strings.ReplaceAll(s, "'VE", " HAVE")
	s = strings.ReplaceAll(s, "CAN'T", "CANNOT")
	s = strings.ReplaceAll(s, "N'T", " NOT")

	// OOG like special punctuation
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "!", "!!!")
	s = strings.ReplaceAll(s, "?", "!!!")
	s = strings.ReplaceAll(s, ".", "!!!")

	for k, v := range oogSayDifferent {
		s = replaceWord(s, k, v)
	}

	theLineInfo.Content = s
}

func doKRAD(s string) {
	s = strings.ReplaceAll(s, "0", "o")
	s = strings.ReplaceAll(s, "1", "l")
	s = strings.ReplaceAll(s, "a", "4")
	s = strings.ReplaceAll(s, "ate", "8")
	s = strings.ReplaceAll(s, "e", "3")
	s = strings.ReplaceAll(s, "b", "6")
	s = strings.ReplaceAll(s, "l", "1")
	s = strings.ReplaceAll(s, "o", "0")
	s = strings.ReplaceAll(s, "s", "5")
	s = strings.ReplaceAll(s, "see", "C")
	s = strings.ReplaceAll(s, "t", "7")
	theLineInfo.Content = s
}

func doHex(s string) {
	theLineInfo.Content = h.Dump([]byte(s))
}

func doShowEnds(s string) {
	// Windows
	s = strings.ReplaceAll(s, "\r\n", "$\r\n")
	// Linux/Unix
	s = strings.ReplaceAll(s, "\n", "$\n")
	// Old MacOS
	s = strings.ReplaceAll(s, "\r", "$\r")

	theLineInfo.Content = s
}

func doRot(s string, r int) {
	myRunes := []rune(s)
	for i, c := range myRunes {
		if unicode.IsLetter(c) {
			myRunes[i] = rune(int(c) + r)
		}
	}
	theLineInfo.Content = string(myRunes)
}

func doStrFry(s string) {
	myRunes := []rune(s)

	var j int
	for i := len(myRunes) - 1; i > 0; i-- {
		j = rand.Intn(i)
		myRunes[j], myRunes[i] = myRunes[i], myRunes[j]
	}

	theLineInfo.Content = string(myRunes)
}

func replaceWord(s, original, replace string) string {
	pattern := `\b` + original + `\b`
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllLiteralString(s, replace)
}

func doNumberNonBlank(_ string) {
	// TODO: Implement the logic for numberNonBlank flag
}

func doNoBlanks(_ string) {
	// TODO: Implement the logic for noBlanks flag
}

func doDos(_ string) {
	// TODO: Implement the logic for dos flag
}

func doHideNonPrinting(_ string) {
	// TODO: Implement the logic for hideNonPrinting flag
}

func doMac(_ string) {
	// TODO: Implement the logic for mac flag
}

func doNumber(_ string) {
	// TODO: Implement the logic for number flag
}

func doSqueezeBlank(_ string) {
	// TODO: Implement the logic for squeezeBlank flag
}

func doShowTabs(_ string) {
	// TODO: Implement the logic for showTabs flag
}

func doSkipTags(_ string) {
	// TODO: Implement the logic for skipTags flag
}

func doTranslate(_ string) {
	// TODO: Implement the logic for translate flag
}

func doUnix(_ string) {
	// TODO: Implement the logic for unix flag
}

func doShowNonPrinting(_ string) {
	// TODO: Implement the logic for showNonPrinting flag
}

func doCols(s string, cols int) {
	if len(s) > cols {
		theLineInfo.Content = s[:cols]
	} else {
		theLineInfo.Content = s
	}
}
