// Package main implements "dog", which is a replacement for the Unix cat command
// with additional features such as strfry, oog, k-rad talk, and more.
//
// Flags:
// - show-all: equivalent to -vET
// - number-nonblank: precede each non-blank line with its line number
// - no-blanks: only print lines with non-whitespace characters
// - dos: convert line endings to DOS-style
// - show-ends: display $ at the end of each line
// - hide-nonprinting: hide non-printing characters
// - hex: display the data as a hex dump
// - images: list unique, absolute image links from input data
// - krad: convert lines to k-rad
// - lines: list of lines to print, comma delimited, ranges allowed
// - links: list unique, absolute URL links from input data
// - lower: convert all upper-case characters to lower
// - mac: convert line endings to Macintosh-style
// - number: precede each line with its line number
// - oog: OOG A STRING!!!
// - rotate: rotate character values (can be negative)
// - squeeze-blank: never more than one single blank line
// - strfry: stir-fry each line
// - show-tabs: display TAB characters as ^I
// - skip-tags: do not process HTML tags from input, and simply output them as-is
// - translate: translate end-of-line characters
// - unix: convert line endings to UNIX-style
// - upper: convert all lower-case characters to upper
// - show-nonprinting: use ^ and M- notation, except for TAB
// - cols: print first 'cols' characters of each line
//
// The Execute function runs the root command and handles any errors that occur.
package main

func main() {
	Execute()
}
