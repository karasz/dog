/*
Copyright © 2021 Nagy Károly Gábriel <k@jpi.io>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "dog [flags] [FILE]\nWith no FILE, or when FILE is -, read standard input.",
	Short: "dog is better than cat",
	Long:  `Dog is a cat replacement with some added thrils like strfry, oog, k-rad talk and others.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var isFlag bool
		cmd.Flags().Visit(func(f *pflag.Flag) {
			isFlag = true
		})

		for _, sf := range args {
			if a, err := processName(sf); err == nil {
				if !isFlag {
					//shortcut, no process just ioCopy
					io.Copy(os.Stdout, a)
				} else {
					processFile(a)
				}
			}
		}

		return nil
	},
}

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

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().BoolVarP(&showAll, "show-all", "A", false, "equivalent to -vET")
	rootCmd.Flags().BoolVarP(&numebrNonBlank, "number-nonblank", "b", false, "precede each non-blank line with its line number")
	rootCmd.Flags().BoolVarP(&noBlanks, "no-blanks", "B", false, "only print lines with non-whitespace characters")
	rootCmd.Flags().BoolVarP(&dos, "dos", "", false, "convert line endings to DOS-style")
	rootCmd.Flags().BoolVarP(&showEnds, "show-ends", "E", false, "display $ at the end of each line")
	rootCmd.Flags().BoolVarP(&hideNonPrinting, "hide-nonprinting", "", false, "hide non-printing characters")
	rootCmd.Flags().BoolVarP(&hex, "hex", "", false, "display the data as a hex dump")
	rootCmd.Flags().BoolVarP(&images, "images", "", false, "list unique, absolute image links from input data")
	rootCmd.Flags().BoolVarP(&krad, "krad", "", false, "convert lines to k-rad")
	rootCmd.Flags().StringVarP(&lines, "lines", "l", "", "list of lines to print, comma delimited, ranges allowed")
	rootCmd.Flags().BoolVarP(&links, "links", "", false, "list unique, absolute URL links from input data")
	rootCmd.Flags().BoolVarP(&lower, "lower", "", false, "convert all upper-case characters to lower")
	rootCmd.Flags().BoolVarP(&mac, "mac", "", false, "convert line endings to Macintosh-style")
	rootCmd.Flags().BoolVarP(&number, "number", "n", false, "precede each line with its line number")
	rootCmd.Flags().BoolVarP(&oog, "oog", "", false, "OOG A STRING!!!")
	rootCmd.Flags().IntVarP(&rot, "rotate", "", 0, "rotate character values (can be negative)")
	rootCmd.Flags().BoolVarP(&squeezeBlank, "squeeze-blank", "s", false, "never more than one single blank link")
	rootCmd.Flags().BoolVarP(&strfry, "strfry", "", false, "stir-fry each line")
	rootCmd.Flags().BoolVarP(&showTabs, "show-tabs", "T", false, "display TAB characters as ^I")
	rootCmd.Flags().BoolVarP(&skipTags, "skip-tags", "", false, "do not process HTML tags from input, and simply output them as-is")
	rootCmd.Flags().BoolVarP(&translate, "translate", "", false, "translate end-of-line characters")
	rootCmd.Flags().BoolVarP(&unix, "unix", "", false, "convert line endings to UNIX-style")
	rootCmd.Flags().BoolVarP(&upper, "upper", "", false, "convert all lower-case characters to upper")
	rootCmd.Flags().BoolVarP(&showNonPrinting, "show-nonprinting", "v", false, "use ^ and M- notation, except for TAB")
	rootCmd.Flags().IntVarP(&cols, "cols", "w", 80, "print first 'cols' characters of each line (default=80)")
}
