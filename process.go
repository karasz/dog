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
var nonBlankLineNumber int

func processNames(names []string) (io.ReadCloser, error) {
	readers := make([]io.ReadCloser, 0, len(names))

	for _, name := range names {
		reader, err := createReader(name)
		if err != nil {
			return nil, err
		}
		readers = append(readers, reader)

		if name == "-" {
			return multireadcloser.MultiReadCloser(readers...), nil
		}
	}
	return multireadcloser.MultiReadCloser(readers...), nil
}

func createReader(name string) (io.ReadCloser, error) {
	if name == "-" {
		return os.Stdin, nil
	}

	if isHTTPURL(name) {
		return createHTTPReader(name)
	}

	return os.Open(name)
}

func isHTTPURL(name string) bool {
	return strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://")
}

func createHTTPReader(url string) (io.ReadCloser, error) {
	netClient := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := netClient.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.ContentLength > maxHTTPBodySize {
		_, _ = fmt.Printf("The resource is too large: %d bytes. What would you like to do?\n",
			resp.ContentLength)
		return nil, fmt.Errorf("resource is too large: %d bytes", resp.ContentLength)
	}

	return resp.Body, nil
}
func processFiles(fl io.ReadCloser, flags pflag.FlagSet) error {
	defer fl.Close()

	theflags = flags
	nonBlankLineNumber = 0 // Reset counter for each file processing
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
	processedContent := processFlags(lineInfo.Content)
	theLineInfo.Content = processedContent

	// Suppress empty lines for HTML extraction flags
	if (theflags.Changed("links") || theflags.Changed("images")) && strings.TrimSpace(theLineInfo.Content) == "" {
		return
	}

	// Handle line numbering
	if theflags.Changed("number") {
		_, _ = fmt.Fprintf(os.Stdout, "%d: %s\n", theLineInfo.LineNumber, theLineInfo.Content)
	} else if theflags.Changed("numberNonBlank") {
		// Only increment and show line number for non-blank lines
		if strings.TrimSpace(theLineInfo.Content) != "" {
			nonBlankLineNumber++
			_, _ = fmt.Fprintf(os.Stdout, "%d: %s\n", nonBlankLineNumber, theLineInfo.Content)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", theLineInfo.Content)
		}
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", theLineInfo.Content)
	}
}

// FlagProcessor manages the transformation pipeline for processing multiple flags
type FlagProcessor struct {
	content         string
	originalContent string
}

func (fp *FlagProcessor) applyTransformation(fn func(string) string) {
	fp.content = fn(fp.content)
}

func (fp *FlagProcessor) applyHTMLExtraction(kind string) {
	readLine := strings.NewReader(fp.originalContent)
	fp.content = extractFromHTML(readLine, kind)
}

func processHTMLExtractionFlags(fp *FlagProcessor) bool {
	if theflags.Changed("links") {
		fp.applyHTMLExtraction("a")
		return true
	}
	if theflags.Changed("images") {
		fp.applyHTMLExtraction("img")
		return true
	}
	return false
}

func processTextTransformationFlags(fp *FlagProcessor) {
	if theflags.Changed("oog") {
		fp.applyTransformation(transformOOG)
	}
	if theflags.Changed("krad") {
		fp.applyTransformation(transformKRAD)
	}
	if theflags.Changed("lower") {
		fp.applyTransformation(strings.ToLower)
	}
	if theflags.Changed("upper") {
		fp.applyTransformation(strings.ToUpper)
	}
	if theflags.Changed("strfry") {
		fp.applyTransformation(transformStrFry)
	}
}

func applyRotFlag(fp *FlagProcessor) {
	if rotValue, err := theflags.GetInt("rot"); err == nil {
		fp.applyTransformation(func(s string) string {
			return transformRot(s, rotValue)
		})
	}
}

func applyColsFlag(fp *FlagProcessor) {
	if colsValue, err := theflags.GetInt("cols"); err == nil && colsValue != 0 {
		fp.applyTransformation(func(s string) string {
			return transformCols(s, colsValue)
		})
	}
}

func processIntegerFlags(fp *FlagProcessor) {
	if theflags.Changed("rot") {
		applyRotFlag(fp)
	}
	if theflags.Changed("cols") {
		applyColsFlag(fp)
	}
}

func processFormatFlags(fp *FlagProcessor) {
	if theflags.Changed("hex") {
		fp.applyTransformation(transformHex)
	}
	if theflags.Changed("showAll") {
		fp.applyTransformation(transformShowAll)
	}
	if theflags.Changed("showEnds") {
		fp.applyTransformation(transformShowEnds)
	}
	if theflags.Changed("showTabs") {
		fp.applyTransformation(transformShowTabs)
	}
	if theflags.Changed("showNonPrinting") {
		fp.applyTransformation(transformShowNonPrinting)
	}
}

func processFlags(line string) string {
	fp := &FlagProcessor{
		content:         line,
		originalContent: line,
	}

	if processHTMLExtractionFlags(fp) {
		return fp.content
	}

	processTextTransformationFlags(fp)
	processIntegerFlags(fp)
	processFormatFlags(fp)

	return fp.content
}

func extractFromHTML(r io.Reader, kind string) string {
	var builder strings.Builder
	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			return builder.String()
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
			_, _ = builder.WriteString(url + "\n")
		}
	}
}

var showAllReplacer = strings.NewReplacer("\r", "^M", "\n", "^J", "\t", "^I")

func transformShowAll(s string) string {
	return showAllReplacer.Replace(s)
}

func transformOOG(s string) string {
	// OOG don't like small
	s = strings.ToUpper(s)
	// OOG don't like contractions and punctuation
	s = oogReplacer.Replace(s)
	for k, v := range oogSayDifferent {
		s = replaceWord(s, k, v)
	}
	return s
}

func transformKRAD(s string) string {
	return kralReplacer.Replace(s)
}

func transformHex(s string) string {
	return h.Dump([]byte(s))
}

func transformShowEnds(s string) string {
	return showEndsReplacer.Replace(s)
}

func transformStrFry(s string) string {
	myRunes := []rune(s)
	var j int
	for i := len(myRunes) - 1; i > 0; i-- {
		j = rand.Intn(i)
		myRunes[j], myRunes[i] = myRunes[i], myRunes[j]
	}
	return string(myRunes)
}

func transformRot(s string, r int) string {
	if r == 0 {
		return s
	}
	myRunes := []rune(s)
	for i, c := range myRunes {
		if unicode.IsLetter(c) {
			myRunes[i] = rune(int(c) + r)
		}
	}
	return string(myRunes)
}

func transformCols(s string, cols int) string {
	if cols <= 0 {
		return s
	}

	runes := []rune(s)
	if len(runes) <= cols {
		return s
	}

	var result strings.Builder
	for i := 0; i < len(runes); i += cols {
		end := min(i+cols, len(runes))
		_, _ = result.WriteString(string(runes[i:end]))
		if end < len(runes) {
			_ = result.WriteByte('\n')
		}
	}

	return result.String()
}

func transformShowTabs(s string) string {
	// TODO: Implement show tabs transformation
	return s
}

func transformShowNonPrinting(s string) string {
	// TODO: Implement show non-printing transformation
	return s
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

var showEndsReplacer = strings.NewReplacer("\r\n", "$\r\n", "\n", "$\n", "\r", "$\r")

func replaceWord(s, original, replace string) string {
	pattern := `\b` + original + `\b`
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllLiteralString(s, replace)
}

func doNumberNonBlank(_ string) {
	// The numberNonBlank functionality is implemented in processLine function
	// This function exists for consistency but the actual logic is handled
	// in the main processing pipeline where line numbering occurs
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
