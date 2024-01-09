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
	"encoding/csv"
	"fmt"
	"runtime"
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// interCmd represents the inter command
var interCmd = &cobra.Command{
	GroupID: "set",

	Use:   "inter",
	Short: "intersection of multiple files",
	Long: `intersection of multiple files

Attention:

  1. fields in all files should be the same, 
     if not, extracting to another file using "csvtk cut".

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}

		ignoreCase := getFlagBool(cmd, "ignore-case")

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			if config.OutDelimiter == ',' {
				writer.Comma = '\t'
			} else {
				writer.Comma = config.OutDelimiter
			}
		} else {
			writer.Comma = config.OutDelimiter
		}
		defer func() {
			writer.Flush()
			checkError(writer.Error())
		}()

		keysMaps := make(map[string]bool, 10000)
		valuesMaps := make(map[string][]string) // store selected columns of first file
		var selectedColnames []string
		var hasHeaderLine bool

		var firstFile = true
		var hasInter = true
		var ok bool

		for i, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					log.Warningf("csvtk inter: skipping empty input file: %s", file)
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr:    fieldStr,
				FuzzyFields: fuzzyFields,

				DoNotAllowDuplicatedColumnName: true,
			})

			var key string

			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					checkFirstLine = false

					if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
						if i == 0 { // use colname of the first file
							selectedColnames = record.Selected
						}
						hasHeaderLine = true

						continue
					}
				}

				key = strings.Join(record.Selected, "_shenwei356_")
				if ignoreCase {
					key = strings.ToLower(key)
				}

				if firstFile {
					keysMaps[key] = false

					valuesMaps[key] = record.Selected
					continue
				}

				if _, ok = keysMaps[key]; ok {
					keysMaps[key] = true
				}

			}

			readerReport(&config, csvReader, file)

			if firstFile {
				firstFile = false
				continue
			}

			// remove unseen kmers
			for key = range keysMaps {
				if keysMaps[key] {
					keysMaps[key] = false
				} else {
					delete(keysMaps, key)
				}
			}

			if len(keysMaps) == 0 {
				hasInter = false
				break
			}
		}

		if !hasInter {
			writer.Flush()
			checkError(writer.Error())
			return
		}

		if hasHeaderLine {
			checkError(writer.Write(selectedColnames))
		}
		for key := range keysMaps {
			checkError(writer.Write(valuesMaps[key]))
		}

	},
}

func init() {
	RootCmd.AddCommand(interCmd)
	interCmd.Flags().StringP("fields", "f", "1", `select these fields as the key. e.g -f 1,2 or -f columnA,columnB`)
	interCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	interCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
}
