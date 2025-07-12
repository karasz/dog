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

var showAllReplacer = strings.NewReplacer("\r", "^M", "\n", "^J", "\t", "^I")

func doShowAll(s string) {
	theLineInfo.Content = showAllReplacer.Replace(s)
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
	var builder strings.Builder
	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			theLineInfo.Content = builder.String()
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
			builder.WriteString(url)
			builder.WriteByte('\n')
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

var (
	oogSayDifferent = map[string]string{"I": "OOG", "IM": "OOG", "ME": "OOG",
		"MY": "OOG'S", "MINE": "OOG'S", "ANOTHER": "OTHER", "HAS": "HAVE", "HAD": "HAVE",
		"CANNOT": "NOT", " IS": "", " ARE": "", " AM": "", " A": "", " AN": "",
		" THAT": "", " WHICH": "", " THE": "", " CAN": "", " OUR": "", " ANY": "", " HIS": "", " HERS": ""}
	
	oogReplacer = strings.NewReplacer(
		"'LL", " WILL",
		"WON'T", "WILL NOT",
		"'VE", " HAVE",
		"CAN'T", "CANNOT",
		"N'T", " NOT",
		",", "",
		"'", "",
		"!", "!!!",
		"?", "!!!",
		".", "!!!")
)

func doOOG(s string) {
	// OOG don't like small
	s = strings.ToUpper(s)

	// OOG don't like contractions and punctuation
	s = oogReplacer.Replace(s)

	for k, v := range oogSayDifferent {
		s = replaceWord(s, k, v)
	}

	theLineInfo.Content = s
}

var kralReplacer = strings.NewReplacer(
	"0", "o",
	"1", "l",
	"ate", "8",
	"see", "C",
	"a", "4",
	"e", "3",
	"b", "6",
	"l", "1",
	"o", "0",
	"s", "5",
	"t", "7")

func doKRAD(s string) {
	theLineInfo.Content = kralReplacer.Replace(s)
}

func doHex(s string) {
	theLineInfo.Content = h.Dump([]byte(s))
}

var showEndsReplacer = strings.NewReplacer("\r\n", "$\r\n", "\n", "$\n", "\r", "$\r")

func doShowEnds(s string) {
	theLineInfo.Content = showEndsReplacer.Replace(s)
}

func doRot(s string, r int) {
	if r == 0 {
		theLineInfo.Content = s
		return
	}
	
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
	if len(s) <= cols {
		theLineInfo.Content = s
		return
	}
	
	runes := []rune(s)
	if len(runes) > cols {
		theLineInfo.Content = string(runes[:cols])
	} else {
		theLineInfo.Content = s
	}
}
