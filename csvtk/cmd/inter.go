// Copyright © 2016 Wei Shen <shenwei356@gmail.com>
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

	"github.com/brentp/xopen"
	"github.com/spf13/cobra"
)

// interCmd represents the seq command
var interCmd = &cobra.Command{
	Use:   "inter",
	Short: "intersection of multiple files",
	Long: `intersection of multiple files

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)
		runtime.GOMAXPROCS(config.NumCPUs)

		fields, colnames, needParseHeaderRow := parseFields(cmd, "fields", "no-header-row")
		ignoreCase := getFlagBool(cmd, "ignore-case")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}

		keysMaps := make(map[string]map[string]struct{})
		valuesMaps := make(map[string][]string) // store selected columns of first file
		parseSelectedColnames := true
		saveDataOfFirstFile := true
		var selectedColnames []string

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			checkError(err)
			csvReader.Run()

			parseHeaderRow := needParseHeaderRow // parsing header row
			var HeaderRow []string
			var colnames2fileds map[string]int // column name -> field

			checkFields := true
			var items []string
			var key string

			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				for _, record := range chunk.Data {
					if parseHeaderRow { // parsing header row
						colnames2fileds = make(map[string]int, len(record))
						for i, col := range record {
							colnames2fileds[col] = i + 1
						}

						if len(fields) == 0 { // user gives the colnames
							fields = []int{}
							for _, col := range colnames {
								if v, ok := colnames2fileds[col]; ok {
									fields = append(fields, v)
								} else {
									log.Warningf("ignore unknown column name: %s", col)
								}
							}
						}

						HeaderRow = record
						parseHeaderRow = false
						continue
					}

					if checkFields {
						fields2 := []int{}
						for _, f := range fields {
							if f > len(record) {
								log.Warningf("ignore unmatched field: %d", f)
								continue
							}
							fields2 = append(fields2, f)
						}
						fields = fields2
						if len(fields) == 0 {
							checkError(fmt.Errorf("no fields matched"))
						}
						items = make([]string, len(fields))

						if parseSelectedColnames && needParseHeaderRow {
							selectedColnames = make([]string, len(fields))
							for i, f := range fields {
								selectedColnames[i] = HeaderRow[f-1]
							}
							parseSelectedColnames = false
						}

						checkFields = false
					}

					for i, f := range fields {
						items[i] = record[f-1]
					}

					key = strings.Join(items, "_shenwei356_")
					if ignoreCase {
						key = strings.ToLower(key)
					}
					if _, ok := keysMaps[key]; !ok {
						keysMaps[key] = make(map[string]struct{})
					}
					keysMaps[key][file] = struct{}{}

					if saveDataOfFirstFile {
						for i, f := range fields {
							items[i] = record[f-1]
						}
						itemsCopy := make([]string, len(items))
						for i, item := range items {
							itemsCopy[i] = item
						}
						valuesMaps[key] = itemsCopy
					}
				}

				saveDataOfFirstFile = false
			}
		}

		if needParseHeaderRow {
			checkError(writer.Write(selectedColnames))
		}
		n := len(files)
		for key, count := range keysMaps {
			if len(count) < n {
				continue
			}
			checkError(writer.Write(valuesMaps[key]))
		}

		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(interCmd)
	interCmd.Flags().StringP("fields", "f", "1", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
	interCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
}
