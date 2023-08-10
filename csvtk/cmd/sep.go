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

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					log.Warningf("csvtk sep: skipping empty input file: %s", file)
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr:    fieldStr,
				FuzzyFields: false,

				DoNotAllowDuplicatedColumnName: true,
			})

			checkNewNumCols := true
			var nNewCols int

			var items []string
			var fieldsMap map[int]interface{}
			var record2 []string // for output
			var line int
			var f int
			var handleHeaderRow bool
			var ok bool

			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					checkFirstLine = false

					fieldsMap = make(map[int]interface{}, len(record.Selected))
					for _, f = range record.Fields {
						fieldsMap[f-1] = struct{}{}
					}
					if !config.NoHeaderRow || record.IsHeaderRow {
						handleHeaderRow = true
					}
				}

				line = record.Line

				if remove {
					record2 = make([]string, 0, len(record.All))
					for f = range record.All {
						if _, ok := fieldsMap[f]; !ok {
							record2 = append(record2, record.All[f])
						}
					}
				} else {
					record2 = record.All
				}

				for f = range record.All {
					if _, ok = fieldsMap[f]; ok {
						if handleHeaderRow {
							record2 = append(record2, names...)
							handleHeaderRow = false
						} else {
							if useRegexp {
								items = sepRe.Split(record.All[f], -1)
							} else {
								items = strings.Split(record.All[f], sep)
							}

							if numCols > 0 { // preset number of new created columns
								if len(items) <= numCols {
									nNewCols = numCols
								} else if drop || merge {
									nNewCols = numCols
								} else {
									checkError(fmt.Errorf("[line %d] number of new columns (%d) > -N (--num-cols) (%d), please increase -N (--num-cols), or switch on --drop or --merge", line, len(items), numCols))
								}
							} else {
								if !config.NoHeaderRow { // decide the number of newly created columns according to the given colnames names
									if checkNewNumCols {
										if len(items) <= len(names) {
											nNewCols = len(names)
										} else if drop || merge {
											nNewCols = len(names)
										} else {
											checkError(fmt.Errorf("[line %d] number of new columns (%d) > number of new column names (%d), please reset -n (--names) ", line, len(items), len(names)))

										}
										checkNewNumCols = false
									}
								} else {
									if nNewCols == 0 { // first line
										nNewCols = len(items)
									}
								}
							}

							if len(items) <= nNewCols { // fill
								record2 = append(record2, items...)
								for i := 0; i < nNewCols-len(items); i++ {
									record2 = append(record2, na)
								}
							} else if drop { // drop
								record2 = append(record2, items[0:nNewCols]...)
							} else if merge {
								if useRegexp {
									items = sepRe.Split(record.All[f], nNewCols)
								} else {
									items = strings.SplitN(record.All[f], sep, nNewCols)
								}
								record2 = append(record2, items...)
							} else {
								if numCols > 0 {
									checkError(fmt.Errorf("[line %d] number of new columns (%d) > -N (--num-cols) (%d),  please increase -N (--num-cols) or drop extra data using --drop, or append remaining data to the last column using --merge", line, len(items), numCols))
								} else {
									checkError(fmt.Errorf("[line %d] number of new columns (%d) exceeds that of first row (%d), please increase -N (--num-cols) or drop extra data using --drop, or append remaining data to the last column using --merge", line, len(items), nNewCols))
								}
							}
						}
						break
					}
				}
				checkError(writer.Write(record2))
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
