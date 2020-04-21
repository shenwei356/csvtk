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
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// sepCmd represents the mutate command
var sepCmd = &cobra.Command{
	Use:     "sep",
	Aliases: []string{"separate"},
	Short:   "separate column into multiple columns",
	Long: `separate column into multiple columns

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		names := getFlagStringSlice(cmd, "names")
		numCols := getFlagInt(cmd, "num-cols")
		if !config.NoHeaderRow {
			if len(names) == 0 {
				checkError(fmt.Errorf("flag -n (--name) needed"))
			} else if numCols > 0 && numCols != len(names) {
				checkError(fmt.Errorf("number of new column names (%d) does not match -N (--num-cols) (%d)", len(names), numCols))
			}
		}

		sep := getFlagString(cmd, "sep")
		if sep == "" {
			checkError(fmt.Errorf("flag -s (--sep) needed"))
		}
		useRegexp := getFlagBool(cmd, "use-regexp")
		ignoreCase := getFlagBool(cmd, "ignore-case")

		var err error
		var sepRe *regexp.Regexp
		if useRegexp {
			sepS := sep
			if ignoreCase {
				sepS = "(?i)" + sepS
			}
			sepRe, err = regexp.Compile(sepS)
			if err != nil {
				checkError(fmt.Errorf("failed to compile regular expression: %s: %s", sep, err))
			}
		}

		remove := getFlagBool(cmd, "remove")
		na := getFlagString(cmd, "na")
		drop := getFlagBool(cmd, "drop")
		merge := getFlagBool(cmd, "merge")

		if drop && merge {
			checkError(fmt.Errorf("flag --drop and --merge could not be used at the same time"))
		}

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}
		fields, colnames, negativeFields, needParseHeaderRow := parseFields(cmd, fieldStr, config.NoHeaderRow)
		if !(len(fields) == 1 || len(colnames) == 1) {
			checkError(fmt.Errorf("only single field allowed"))
		}
		if negativeFields {
			checkError(fmt.Errorf("unselect not allowed"))
		}

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

		fuzzyFields := false

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

			handleHeaderRow := needParseHeaderRow
			checkFields := true

			checkNewNumCols := true
			var nNewCols int

			var items []string
			var record2 []string // for output
			var line int

			printMetaLine := true
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				if printMetaLine && len(csvReader.MetaLine) > 0 {
					outfh.WriteString(fmt.Sprintf("sep=%s\n", string(writer.Comma)))
					printMetaLine = false
				}

				for _, record := range chunk.Data {
					line++

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

						parseHeaderRow = false
					}
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

					if remove {
						record2 = make([]string, 0, len(record))
						for f := range record {
							if _, ok := fieldsMap[f+1]; !ok {
								record2 = append(record2, record[f])
							}
						}
					} else {
						record2 = record
					}
					for f := range record {
						if _, ok := fieldsMap[f+1]; ok {
							if handleHeaderRow {
								record2 = append(record2, names...)
								handleHeaderRow = false
							} else {
								if useRegexp {
									items = sepRe.Split(record[f], -1)
								} else {
									items = strings.Split(record[f], sep)
								}
								if checkNewNumCols {
									if !config.NoHeaderRow {
										if numCols > 0 {
											if len(items) <= numCols {
												nNewCols = numCols
											} else if drop || merge {
												nNewCols = numCols
											} else {
												checkError(fmt.Errorf("[line %d] number of new columns (%d) > -N (--num-cols) (%d), please increase -N (--num-cols)", line, len(items), numCols))
											}
										} else {
											if len(items) <= len(names) {
												nNewCols = len(items)
											} else if drop || merge {
												nNewCols = len(names)
											} else {
												checkError(fmt.Errorf("[line %d] number of new columns (%d) > number of new column names (%d), please reset -n (--names) ", line, len(items), len(names)))
											}
										}
									} else {
										if numCols > 0 {
											if len(items) < numCols {
												nNewCols = numCols
											} else {
												nNewCols = len(items)
											}
										} else {
											nNewCols = len(items)
										}
									}

									checkNewNumCols = false
								}

								if len(items) <= nNewCols { // fill
									record2 = append(record2, items...)
									for i := 0; i < numCols-len(items); i++ {
										record2 = append(record2, na)
									}
								} else if drop { // drop
									record2 = append(record2, items[0:nNewCols]...)
								} else if merge {
									if useRegexp {
										items = sepRe.Split(record[f], nNewCols)
									} else {
										items = strings.SplitN(record[f], sep, nNewCols)
									}
									record2 = append(record2, items...)
								} else {
									if numCols > 0 {
										checkError(fmt.Errorf("[line %d] number of new columns (%d) > -N (--num-cols) (%d),  please increase -N (--num-cols) or drop extra data using --drop", line, len(items), numCols))
									} else {
										checkError(fmt.Errorf("[line %d] number of new columns (%d) exceeds that of first row (%d), please increase -N (--num-cols) or drop extra data using --drop", line, len(items), nNewCols))
									}
								}
							}
							break
						}
					}
					checkError(writer.Write(record2))
				}
			}

			readerReport(&config, csvReader, file)
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(sepCmd)
	sepCmd.Flags().StringP("fields", "f", "1", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
	sepCmd.Flags().StringP("sep", "s", "", `separator`)
	sepCmd.Flags().BoolP("use-regexp", "r", false, `separator is a regular expression`)
	sepCmd.Flags().BoolP("ignore-case", "i", false, "ignore case")
	sepCmd.Flags().StringSliceP("names", "n", []string{}, `new column names`)
	sepCmd.Flags().IntP("num-cols", "N", 0, `preset number of new created columns`)
	sepCmd.Flags().BoolP("remove", "R", false, `remove input column`)
	sepCmd.Flags().StringP("na", "", "", "content for filling NA data")
	sepCmd.Flags().BoolP("drop", "", false, "drop extra data, exclusive with --merge")
	sepCmd.Flags().BoolP("merge", "", false, "only splits at most N times, exclusive with --drop")
}
