// Copyright © 2016-2019 Wei Shen <shenwei356@gmail.com>
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
	"regexp"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// gatherCmd represents the gather command
var gatherCmd = &cobra.Command{
	Use:   "gather",
	Short: "gather columns into key-value pairs",
	Long: `gather columns into key-value pairs

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		if config.NoHeaderRow {
			checkError(fmt.Errorf("flag -H/--no-header-row not allowed"))
		}

		fieldKey := getFlagString(cmd, "key")
		fieldValue := getFlagString(cmd, "value")
		if !config.NoHeaderRow {
			if fieldKey == "" {
				checkError(fmt.Errorf("flag -k/--key needed"))
			}
			if fieldValue == "" {
				checkError(fmt.Errorf("flag -v/--value needed"))
			}
		}

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		fields, colnames, negativeFields, needParseHeaderRow := parseFields(cmd, fieldStr, config.NoHeaderRow)
		var fieldsMap map[int]struct{}
		var fieldsOrder map[int]int      // for set the order of fields
		var colnamesOrder map[string]int // for set the order of fields

		if len(fields) > 0 {
			fields2 := make([]int, len(fields))
			fieldsMap = make(map[int]struct{}, len(fields))
			for i, f := range fields {
				if negativeFields {
					fieldsMap[f*-1] = struct{}{}
					fields2[i] = f * -1
				} else {
					fieldsMap[f] = struct{}{}
					fields2[i] = f
				}
			}
			fields = fields2

			if !negativeFields {
				fieldsOrder = make(map[int]int, len(fields))
				i := 0
				for _, f := range fields {
					fieldsOrder[f] = i
					i++
				}
			}
		} else {
			fieldsOrder = make(map[int]int, len(colnames))
			colnamesOrder = make(map[string]int, len(colnames))
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)
		checkError(err)
		csvReader.Run()

		parseHeaderRow := needParseHeaderRow // parsing header row
		var colnames2fileds map[string]int   // column name -> field
		var colnamesMap map[string]*regexp.Regexp
		var colnames0 []string

		checkFields := true
		var items []string
		var fieldsLeft []int
		var isRowLine bool
		var nFieldsLeft int

		printMetaLine := true
		for chunk := range csvReader.Ch {
			checkError(chunk.Err)

			if printMetaLine && len(csvReader.MetaLine) > 0 {
				outfh.WriteString(fmt.Sprintf("sep=%s\n", string(writer.Comma)))
				printMetaLine = false
			}

			for _, record := range chunk.Data {
				if parseHeaderRow { // parsing header row
					colnames2fileds = make(map[string]int, len(record))
					for i, col := range record {
						colnames2fileds[col] = i + 1
					}
					colnamesMap = make(map[string]*regexp.Regexp, len(colnames))
					i := 0
					for _, col := range colnames {
						if !fuzzyFields {
							if negativeFields {
								if _, ok := colnames2fileds[col[1:]]; !ok {
									checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col[1:], file))
								}
							} else {
								if _, ok := colnames2fileds[col]; !ok {
									checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col, file))
								}
							}
						}
						if negativeFields {
							colnamesMap[col[1:]] = fuzzyField2Regexp(col[1:])
						} else {
							colnamesMap[col] = fuzzyField2Regexp(col)
							colnamesOrder[col] = i
							i++
						}
					}

					if len(fields) == 0 { // user gives the colnames
						fields = []int{}
						for _, col := range record {
							var ok bool
							if fuzzyFields {
								for _, re := range colnamesMap {
									if re.MatchString(col) {
										ok = true
										break
									}
								}
							} else {
								_, ok = colnamesMap[col]
							}
							if ok {
								fields = append(fields, colnames2fileds[col])
								fieldsOrder[colnames2fileds[col]] = colnamesOrder[col]
							}
						}

					}

					fieldsMap = make(map[int]struct{}, len(fields))
					for _, f := range fields {
						fieldsMap[f] = struct{}{}
					}

					colnames0 = record
					isRowLine = true

					parseHeaderRow = false
				}
				if checkFields {
					for field := range fieldsMap {
						if field > len(record) {
							checkError(fmt.Errorf(`field (%d) out of range (%d) in file: %s`, field, len(record), file))
						}
					}
					fields2 := []int{}
					fieldsLeft = []int{}
					for f := range record {
						_, ok := fieldsMap[f+1]
						if negativeFields {
							if !ok {
								fields2 = append(fields2, f+1)
							} else {
								fieldsLeft = append(fieldsLeft, f+1)
							}
						} else {
							if ok {
								fields2 = append(fields2, f+1)
							} else {
								fieldsLeft = append(fieldsLeft, f+1)
							}
						}
					}
					fields = fields2

					if len(fields) == 0 {
						checkError(fmt.Errorf("no fields matched in file: %s", file))
					}

					nFieldsLeft = len(fieldsLeft)
					items = make([]string, nFieldsLeft+2)

					checkFields = false
				}

				for i, f := range fieldsLeft {
					items[i] = record[f-1]
				}

				if isRowLine {
					items[nFieldsLeft] = fieldKey
					items[nFieldsLeft+1] = fieldValue
					checkError(writer.Write(items))

					isRowLine = false
				} else {
					for _, f := range fields {
						items[nFieldsLeft] = colnames0[f-1]
						items[nFieldsLeft+1] = record[f-1]
						checkError(writer.Write(items))
					}
				}

			}
		}

		writer.Flush()
		checkError(writer.Error())

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(gatherCmd)
	gatherCmd.Flags().StringP("fields", "f", "", `fields for gathering. e.g -f 1,2 or -f columnA,columnB, or -f -columnA for unselect columnA`)
	gatherCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	gatherCmd.Flags().StringP("key", "k", "", `name of key column to create in output`)
	gatherCmd.Flags().StringP("value", "v", "", `name of value column to create in output`)
}
