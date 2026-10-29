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
	"fmt"
	"runtime"
	"strings"

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
Valid quoted fields, including escaped double-quotes, are preserved.

When and how to use it:
  1. Values containing bare double quotes, e.g.,
       a,abc" xyz,d
     Error information: bare " in non-quoted-field.
     Fix: add the flag -l/--lazy-quotes.
     Using this command:
       a,abc" xyz,d   ->   a,"abc"" xyz",d
  2. Values with double quotes at the beginning but not at the end, e.g.,
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

		bufferSizeS := getFlagString(cmd, "buffer-size")
		if bufferSizeS == "" {
			checkError(fmt.Errorf("value of buffer size. supported unit: K, M, G"))
		}
		bufferSize, err := ParseByteSize(bufferSizeS)
		if err != nil {
			checkError(fmt.Errorf("invalid value of buffer size. supported unit: K, M, G"))
		}
		_buf := make([]byte, bufferSize)

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer closeOutput(outfh)

		fh, err := xopen.Ropen(files[0])
		checkError(err)
		defer func() {
			checkError(fh.Close())
		}()

		var buf []byte // Reuse storage for repaired lines.
		scanner := bufio.NewScanner(fh)
		scanner.Buffer(_buf, int(bufferSize))
		ncols := -1
		var iLine int
		var commentChar byte = byte(config.CommentChar)
		for scanner.Scan() {
			iLine++
			line := scanner.Text()

			if len(line) == 0 || line[0] == commentChar {
				outfh.WriteString(line)
				outfh.WriteByte('\n')
				config.flushOutput(outfh)

				continue
			}

			var n int
			buf, n = appendFixedCSVLine(buf[:0], line, byte(config.Delimiter))
			buf = append(buf, '\n')
			_, err = outfh.Write(buf)
			checkError(err)
			config.flushOutput(outfh)

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

	fixquotesCmd.Flags().StringP("buffer-size", "b", "1G", `buffer size, supported units: K, M, G. Increase the value if a "bufio.Scanner: token too long" error is reported`)
}

// appendFixedCSVLine preserves valid CSV fields and quotes malformed fields.
// Input is one physical line; multiline fields are intentionally unsupported.
func appendFixedCSVLine(dst []byte, line string, delimiter byte) ([]byte, int) {
	fields := 0
	for start := 0; ; {
		end := -1
		// A quoted field may contain delimiters and doubled quotes. Only an
		// unescaped closing quote followed by a delimiter or EOF ends it.
		if start < len(line) && line[start] == '"' {
			for i := start + 1; i < len(line); i++ {
				if line[i] != '"' {
					continue
				}
				if i+1 < len(line) && line[i+1] == '"' {
					i++
					continue
				}
				if i+1 == len(line) || line[i+1] == delimiter {
					end = i + 1
				}
				break
			}
		}
		validQuoted := end >= 0
		if !validQuoted {
			end = len(line)
			if i := strings.IndexByte(line[start:], delimiter); i >= 0 {
				end = start + i
			}
		}
		field := line[start:end]
		if validQuoted || !strings.ContainsRune(field, '"') {
			dst = append(dst, field...)
		} else {
			// Malformed quoting is literal data; escape every original quote.
			dst = append(dst, '"')
			for i := 0; i < len(field); i++ {
				if field[i] == '"' {
					dst = append(dst, '"')
				}
				dst = append(dst, field[i])
			}
			dst = append(dst, '"')
		}
		fields++
		if end == len(line) {
			return dst, fields
		}
		dst = append(dst, delimiter)
		start = end + 1
	}
}
