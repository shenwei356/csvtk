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

		fields, colnames, negativeFields, needParseHeaderRow := parseFields(cmd, fieldStr, config.NoHeaderRow)
		var fieldsMap map[int]struct{}
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

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			checkError(err)
			csvReader.Run()

			parseHeaderRow := needParseHeaderRow // parsing header row
			var colnames2fileds map[string]int   // column name -> field
			var colnamesMap map[string]*regexp.Regexp

			checkFields := true
			printMetaLine := true
			var N int64
			var recordWithN []string

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
								}
							}
						}

						fieldsMap = make(map[int]struct{}, len(fields))
						for _, f := range fields {
							fieldsMap[f] = struct{}{}
						}

						if printLineNumber {
							recordWithN = []string{"n"}
							recordWithN = append(recordWithN, record...)
							record = recordWithN
						}
						checkError(writer.Write(record))
						parseHeaderRow = false
						continue
					}
					N++

					if checkFields {
						for field := range fieldsMap {
							if field > len(record) {
								checkError(fmt.Errorf(`field (%d) out of range (%d) in file: %s`, field, len(record), file))
							}
						}
						fields2 := []int{}
						for f := range record {
							_, ok := fieldsMap[f+1]
							if negativeFields {
								if !ok {
									fields2 = append(fields2, f+1)
								}
							} else {
								if ok {
									fields2 = append(fields2, f+1)
								}
							}
						}
						fields = fields2
						if len(fields) == 0 {
							checkError(fmt.Errorf("no fields matched in file: %s", file))
						}
						fieldsMap = make(map[int]struct{}, len(fields))
						for _, f := range fields {
							fieldsMap[f] = struct{}{}
						}

						checkFields = false

					}

					flag := false
					n := 0
					for i, c := range record {
						_, ok := fieldsMap[i+1]
						if (negativeFields && ok) || (!negativeFields && !ok) {
							continue
						}
						if !reDigitals.MatchString(c) {
							flag = false
							break
						}
						v, err := strconv.ParseFloat(removeComma(c), 64)
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
					if (!negativeFields && n == len(fields)) ||
						(negativeFields && n == len(record)-len(fields)) { // all satisfied
						flag = true
					}
					if !flag {
						continue
					}

					if printLineNumber {
						recordWithN = []string{fmt.Sprintf("%d", N)}
						recordWithN = append(recordWithN, record...)
						record = recordWithN
					}
					checkError(writer.Write(record))
				}
			}

			readerReport(&config, csvReader, file)
		}
		writer.Flush()
		checkError(writer.Error())
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
