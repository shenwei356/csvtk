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
	"fmt"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// delQuotesCmd represents the csv2tab command
var delQuotesCmd = &cobra.Command{
	GroupID: "edit",

	Use:   "del-quotes",
	Short: "remove extra double quotes added by 'fix-quotes'",
	Long: `remove extra double quotes added by 'fix-quotes'

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

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		if config.Tabs {
			config.Delimiter = '\t'
		}

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)
		if err != nil {
			if err == xopen.ErrNoContent {
				if config.Verbose {
					log.Warningf("csvtk csv2tab: skipping empty input file: %s", file)
				}
				return
			}
			checkError(err)
		}

		csvReader.Read(ReadOption{
			FieldStr:      "1-",
			ShowRowNumber: config.ShowRowNumber,
		})

		d := string(config.Delimiter)
		var i int
		var v string
		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}
			for i, v = range record.Selected {
				// if fieldNeedsQuotes(v, config.Delimiter) {
				if strings.Contains(v, d) {
					record.Selected[i] = `"` + v + `"`
				}
			}
			outfh.WriteString(strings.Join(record.Selected, d))
			outfh.WriteByte('\n')
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(delQuotesCmd)
}

// copy from https://cs.opensource.google/go/go/+/refs/tags/go1.21.4:src/encoding/csv/writer.go;l=157
func fieldNeedsQuotes(field string, comma rune) bool {
	if field == "" {
		return false
	}

	if field == `\.` {
		return true
	}

	if comma < utf8.RuneSelf {
		for i := 0; i < len(field); i++ {
			c := field[i]
			if c == '\n' || c == '\r' || c == '"' || c == byte(comma) {
				return true
			}
		}
	} else {
		if strings.ContainsRune(field, comma) || strings.ContainsAny(field, "\"\r\n") {
			return true
		}
	}

	r1, _ := utf8.DecodeRuneInString(field)
	return unicode.IsSpace(r1)
}
