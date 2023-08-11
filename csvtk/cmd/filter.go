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
	"regexp"
	"runtime"
	"strconv"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// filterCmd represents the filter command
var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "filter rows by values of selected fields with arithmetic expression",
	Long: `filter rows by values of selected fields with arithmetic expression

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		filterStr := getFlagString(cmd, "filter")
		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		any := getFlagBool(cmd, "any")
		printLineNumber := getFlagBool(cmd, "line-number")

		if filterStr == "" {
			checkError(fmt.Errorf("flag -f (--filter) needed"))
		}

		if !reFilter.MatchString(filterStr) {
			checkError(fmt.Errorf("invalid filter: %s", filterStr))
		}
		items := reFilter.FindAllStringSubmatch(filterStr, 1)
		fieldStr, expression := items[0][1], items[0][2]
		switch expression {
		case ">":
		case "<":
		case "=":
		case ">=":
		case "<=":
		case "!=", "<>":
		default:
			checkError(fmt.Errorf("invalid expression: %s", expression))
		}
		threshold, err := strconv.ParseFloat(items[0][3], 64)
		checkError(err)

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

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					log.Warningf("csvtk filter: skipping empty input file: %s", file)
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr:      fieldStr,
				FuzzyFields:   fuzzyFields,
				ShowRowNumber: printLineNumber || config.ShowRowNumber,

				DoNotAllowDuplicatedColumnName: true,
			})

			var N int64
			var flag bool
			var n int
			var v float64
			var val string

			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					checkFirstLine = false

					if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
						if printLineNumber {
							unshift(&record.All, "row")
						}
						checkError(writer.Write(record.All))
						continue
					}
				}

				N++

				flag = false
				n = 0

				for _, val = range record.Selected {
					if !reDigitals.MatchString(val) {
						flag = false
						break
					}

					v, err = strconv.ParseFloat(removeComma(val), 64)
					checkError(err)

					switch expression {
					case ">":
						if v > threshold {
							n++
						}
					case "<":
						if v < threshold {
							n++
						}
					case "=":
						if v == threshold {
							n++
						}
					case ">=":
						if v >= threshold {
							n++
						}
					case "<=":
						if v <= threshold {
							n++
						}
					case "!=", "<>":
						if v != threshold {
							n++
						}
					default:
					}

					if any {
						if n == 1 {
							flag = true
							break
						}
					}
				}

				if n == len(record.Fields) { // all satisfied
					flag = true
				}
				if !flag {
					continue
				}

				if printLineNumber {
					unshift(&record.All, "row")
				}
				checkError(writer.Write(record.All))
			}

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	RootCmd.AddCommand(filterCmd)
	filterCmd.Flags().StringP("filter", "f", "", `filter condition. e.g. -f "age>12" or -f "1,3<=2" or -F -f "c*!=0"`)
	filterCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	filterCmd.Flags().BoolP("any", "", false, `print record if any of the field satisfy the condition`)
	filterCmd.Flags().BoolP("line-number", "n", false, `print line number as the first column ("n")`)
}

var reFilter = regexp.MustCompile(`^(.+?)([!<=>]+)([\-\d\.e,E\+]+)$`)
