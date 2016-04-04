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

	"github.com/brentp/xopen"
	"github.com/spf13/cobra"
)

// cutCmd represents the seq command
var cutCmd = &cobra.Command{
	Use:   "cut",
	Short: "select parts of fields",
	Long: `select parts of fields

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)

		fields, colnames, needParseHeaderRow := parseFields(cmd, "fields", "no-header-row")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			checkError(err)
			csvReader.Run()

			parseHeaderRow := needParseHeaderRow // parsing header row
			var colnames2fileds map[string]int   // column name -> field

			checkFields := true
			var items []string

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

						parseHeaderRow = false
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

						checkFields = false
					}

					for i, f := range fields {
						items[i] = record[f-1]
					}
					checkError(writer.Write(items))
				}
			}
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(cutCmd)
	cutCmd.Flags().StringP("fields", "f", "", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
}
