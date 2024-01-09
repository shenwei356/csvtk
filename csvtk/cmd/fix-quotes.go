// Copyright © 2016-2023 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// fixquotesCmd represents the pretty command
var fixquotesCmd = &cobra.Command{
	GroupID: "edit",

	Use:   "fix-quotes",
	Short: "fix malformed CSV/TSV caused by double-quotes",
	Long: `fix malformed CSV/TSV caused by double-quotes

This command fixes fields not appropriately enclosed by double-quotes
to meet the RFC4180 specification (https://rfc-editor.org/rfc/rfc4180.html).

When and how to:
  1. Values containing bare double quotes. e.g.,
       a,abc" xyz,d
     Error information: bare " in non-quoted-field.
     Fix: adding the flag -l/--lazy-quotes.
     Using this command:
       a,abc" xyz,d   ->   a,"abc"" xyz",d
  2. Values with double quotes in the begining but not in the end. e.g.,
       a,"abc" xyz,d
     Error information: extraneous or missing " in quoted-field.
     Using this command:
       a,"abc" xyz,d  ->   a,"""abc"" xyz",d

Next:
  1. You can process the data without the flag -l/--lazy-quotes.
  2. Use 'csvtk del-quotes' if you want to restore the original format.

Limitation:
  1. Values containing line breaks are not supported.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		if config.Tabs {
			config.Delimiter = '\t'
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		fh, err := xopen.Ropen(files[0])
		checkError(err)
		defer func() {
			checkError(fh.Close())
		}()

		var buf bytes.Buffer

		scanner := bufio.NewScanner(fh)
		var line string
		var i, s int
		var r, p rune
		var firstField, firstChar bool
		var hasLeftQuotes, hasRightQuotes bool
		var nInnerQuotes int // number of inner quotes, might including the right quotes
		d := config.Delimiter
		re := regexp.MustCompile(`"`)
		var field string
		var n, ncols int
		ncols = -1
		var iLine int
		var reQuotedDelimiter = regexp.MustCompile(fmt.Sprintf(`(^|%c)".*%c.*"($|%c)`, d, d, d))
		var hasQuotedDelimiter bool
		var commentChar byte = byte(config.CommentChar)
		for scanner.Scan() {
			iLine++
			line = scanner.Text()
			hasQuotedDelimiter = reQuotedDelimiter.MatchString(line)

			if len(line) == 0 || line[0] == commentChar {
				outfh.WriteString(line)
				outfh.WriteByte('\n')

				continue
			}

			n = 0
			firstField = true

			firstChar = true
			nInnerQuotes, hasLeftQuotes, hasRightQuotes = 0, false, false
			buf.Reset()

			s = 0

			for i, r = range line {
				if r == d {
					if p == '"' {
						hasRightQuotes = true
						nInnerQuotes--
					}

					// might be a comma within a field
					if hasLeftQuotes && !hasRightQuotes && hasQuotedDelimiter {
						continue
					}

					if firstField {
						field = line[s:i]
					} else {
						field = line[s+1 : i]
					}

					if nInnerQuotes > 0 ||
						(hasLeftQuotes && !hasRightQuotes) ||
						(!hasLeftQuotes && hasRightQuotes) {
						field = re.ReplaceAllString(field, `""`)
						field = `"` + field + `"`
					}

					if !firstField {
						buf.WriteRune(d)
					}
					buf.WriteString(field)

					s = i

					firstField = false
					n++

					firstChar = true
					nInnerQuotes, hasLeftQuotes, hasRightQuotes = 0, false, false

					continue
				}

				if firstChar {
					if r == '"' {
						hasLeftQuotes = true
					}
					firstChar = false
				} else if r == '"' {
					nInnerQuotes++
				}
				p = r

			}

			i = len(line)
			// the last record

			if p == '"' {
				hasRightQuotes = true
				nInnerQuotes--
			}

			if firstField {
				field = line[s:i]
			} else {
				field = line[s+1 : i]
			}

			if nInnerQuotes > 0 ||
				(hasLeftQuotes && !hasRightQuotes) ||
				(!hasLeftQuotes && hasRightQuotes) {
				field = re.ReplaceAllString(field, `""`)
				field = `"` + field + `"`
			}

			if !firstField {
				buf.WriteRune(d)
			}
			buf.WriteString(field)

			// the last record

			n++

			buf.WriteByte('\n')

			outfh.Write(buf.Bytes())

			// check ncols
			if ncols < 0 {
				ncols = n
			} else if n != ncols {
				checkError(fmt.Errorf("failed to fix (unequal number of fields: %d (line %d) != %d (line %d), does exist quoted delimiter?): %s",
					n, iLine, ncols, iLine-1, line))

			}

		}
		if err := scanner.Err(); err != nil {
			checkError(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(fixquotesCmd)
}
