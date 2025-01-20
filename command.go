package main

import (
	"fmt"
	"os"

	"github.com/karasz/dog/version"
	"github.com/spf13/pflag"
)

var (
	showAll         bool
	numebrNonBlank  bool
	noBlanks        bool
	dos             bool
	showEnds        bool
	hideNonPrinting bool
	hex             bool
	images          bool
	krad            bool
	lines           string
	links           bool
	lower           bool
	mac             bool
	number          bool
	oog             bool
	rot             int
	squeezeBlank    bool
	strfry          bool
	showTabs        bool
	skipTags        bool
	translate       bool
	unix            bool
	upper           bool
	showNonPrinting bool
	cols            int
)

func init() {
	pflag.BoolVarP(&showAll, "showAll", "A", false, "equivalent to -vET")
	pflag.BoolVarP(&numebrNonBlank, "numberNonBlank", "b", false, "precede each non-blank line with its line number")
	pflag.BoolVarP(&noBlanks, "noBlanks", "B", false, "only print lines with non-whitespace characters")
	pflag.BoolVar(&dos, "dos", false, "convert line endings to DOS-style")
	pflag.BoolVarP(&showEnds, "showEnds", "E", false, "display $ at the end of each line")
	pflag.BoolVar(&hideNonPrinting, "hideNonPrinting", false, "hide non-printing characters")
	pflag.BoolVar(&hex, "hex", false, "display the data as a hex dump")
	pflag.BoolVar(&images, "images", false, "list unique, absolute image links from input data")
	pflag.BoolVar(&krad, "krad", false, "convert lines to k-rad")
	pflag.StringVarP(&lines, "lines", "l", "", "list of lines to print, comma delimited, ranges allowed")
	pflag.BoolVar(&links, "links", false, "list unique, absolute URL links from input data")
	pflag.BoolVar(&lower, "lower", false, "convert all upper-case characters to lower")
	pflag.BoolVar(&mac, "mac", false, "convert line endings to Macintosh-style")
	pflag.BoolVarP(&number, "number", "n", false, "precede each line with its line number")
	pflag.BoolVar(&oog, "oog", false, "OOG A STRING!!!")
	pflag.IntVar(&rot, "rot", 0, "rotate character values (can be negative)")
	pflag.BoolVarP(&squeezeBlank, "squeezeBlank", "s", false, "never more than one single blank line")
	pflag.BoolVar(&strfry, "strfry", false, "stir-fry each line")
	pflag.BoolVarP(&showTabs, "showTabs", "T", false, "display TAB characters as ^I")
	pflag.BoolVar(&skipTags, "skipTags", false, "do not process HTML tags from input, and simply output them as-is")
	pflag.BoolVar(&translate, "translate", false, "translate end-of-line characters")
	pflag.BoolVar(&unix, "unix", false, "convert line endings to UNIX-style")
	pflag.BoolVar(&upper, "upper", false, "convert all lower-case characters to upper")
	pflag.BoolVarP(&showNonPrinting, "showNonPrinting", "v", false, "use ^ and M- notation, except for TAB")
	pflag.IntVarP(&cols, "cols", "w", 80, "print first 'cols' characters of each line")
	_ = pflag.Bool("version", false, "show dog version and exit")
}

// Execute parses command-line flags and arguments, determines if input is from a pipe,
// and processes each argument accordingly. If no arguments are provided, a default
// argument is used. Errors encountered during processing are printed to standard error.
// revive:disable:cognitive-complexity
func Execute() {
	pflag.Parse()
	flags := pflag.CommandLine
	if ok, err := flags.GetBool("version"); ok && err == nil {
		_, _ = fmt.Printf("Version: %s\nBuilt:   %s\n", version.Version, version.BuildDate)
		return
	}
	args := pflag.Args()
	if len(args) == 0 {
		args = append(args, "-")
	}
	if a, err := processNames(args); err == nil {
		if err := processFiles(a, *flags); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	} else {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
