// Copyright © 2016-2021 Wei Shen <shenwei356@gmail.com>
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

// mutateCmd represents the mutate command
var mutateCmd = &cobra.Command{
	Use:   "mutate",
	Short: "create new column from selected fields by regular expression",
	Long: `create new column from selected fields by regular expression

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		at := getFlagNonNegativeInt(cmd, "at")
		after := getFlagString(cmd, "after")
		before := getFlagString(cmd, "before")
		if config.NoHeaderRow {
			if after != "" {
				checkError(fmt.Errorf("the flag --after is not allowed with -H/--no-header-row"))
			}
			if before != "" {
				checkError(fmt.Errorf("the flag --before is not allowed with -H/--no-header-row"))
			}
		}
		if after != "" && before != "" {
			checkError(fmt.Errorf("the flag --after and --before are incompatible"))
		}
		if at > 0 && !(after == "" && before == "") {
			checkError(fmt.Errorf("the flag --at is incompatible with --after and --before"))
		}

		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		ignoreCase := getFlagBool(cmd, "ignore-case")
		naUnmatched := getFlagBool(cmd, "na")
		pattern := getFlagString(cmd, "pattern")
		if !regexp.MustCompile(`\(.+\)`).MatchString(pattern) {
			checkError(fmt.Errorf(`value of -p (--pattern) must contains "(" and ")" to capture data which is used to create new column`))
		}

		name := getFlagString(cmd, "name")
		if !config.NoHeaderRow && name == "" {
			checkError(fmt.Errorf("flag -n (--name) needed"))
		}

		p := pattern
		if ignoreCase {
			p = "(?i)" + p
		}
		patternRegexp, err := regexp.Compile(p)
		checkError(err)

		remove := getFlagBool(cmd, "remove")

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}
		fields, colnames, negativeFields, needParseHeaderRow, _ := parseFields(cmd, fieldStr, ",", config.NoHeaderRow)
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

		// fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		fuzzyFields := false

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
					log.Warningf("csvtk mutate: skipping empty input file: %s", file)
					continue
				}
				checkError(err)
			}

			csvReader.Run()

			parseHeaderRow := needParseHeaderRow // parsing header row
			var colnames2fileds map[string][]int // column name -> []field
			var colnamesMap map[string]*regexp.Regexp

			handleHeaderRow := needParseHeaderRow
			checkFields := true

			var record []string
			var record2 []string // for output
			var _fields []int
			var ok bool
			var f int
			var value string

			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				for _, record = range chunk.Data {
					if parseHeaderRow { // parsing header row
						colnames2fileds = make(map[string][]int, len(record))
						for i, col := range record {
							if _, ok := colnames2fileds[col]; !ok {
								colnames2fileds[col] = []int{i + 1}
							} else {
								colnames2fileds[col] = append(colnames2fileds[col], i+1)
							}
						}
						colnamesMap = make(map[string]*regexp.Regexp, len(colnames))
						for _, col := range colnames {
							if !fuzzyFields {
								if negativeFields {
									if _, ok := colnames2fileds[col[1:]]; !ok {
										checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col[1:], file))
									} else if len(colnames2fileds[col]) > 1 {
										checkError(fmt.Errorf("the selected colname is duplicated in the input data: %s", col))
									}
								} else {
									if _, ok := colnames2fileds[col]; !ok {
										checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col, file))
									} else if len(colnames2fileds[col]) > 1 {
										checkError(fmt.Errorf("the selected colname is duplicated in the input data: %s", col))
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
									fields = append(fields, colnames2fileds[col]...)
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

					if handleHeaderRow {
						record2 = append(record2, name)
						if after != "" {
							if _fields, ok = colnames2fileds[after]; ok {
								at = _fields[len(_fields)-1] + 1
							} else {
								checkError(fmt.Errorf(`column "%s" not existed in file: %s`, after, file))
							}
							copy(record2[at:], record2[at-1:len(record2)-1])
							record2[at-1] = name
						} else if before != "" {
							if _fields, ok = colnames2fileds[before]; ok {
								at = _fields[len(_fields)-1]
							} else {
								checkError(fmt.Errorf(`column "%s" not existed in file: %s`, before, file))
							}
							copy(record2[at:], record2[at-1:len(record2)-1])
							record2[at-1] = name
						} else if at > 0 && at <= len(record2) {
							copy(record2[at:], record2[at-1:len(record2)-1])
							record2[at-1] = name
						}

						handleHeaderRow = false
						checkError(writer.Write(record2))
						continue
					}

					for f = range record {
						// record2[f] = record[f]
						_, ok = fieldsMap[f+1]
						if !ok {
							continue
						}

						if patternRegexp.MatchString(record[f]) {
							found := patternRegexp.FindAllStringSubmatch(record[f], -1)
							value = found[0][1]
						} else {
							if naUnmatched {
								value = ""
							} else {
								value = record[f]
							}
						}
						record2 = append(record2, value)
						if after != "" {
							if _fields, ok = colnames2fileds[after]; ok {
								at = _fields[len(_fields)-1] + 1
							} else {
								checkError(fmt.Errorf(`column "%s" not existed in file: %s`, after, file))
							}
							copy(record2[at:], record2[at-1:len(record2)-1])
							record2[at-1] = value
						} else if before != "" {
							if _fields, ok = colnames2fileds[before]; ok {
								at = _fields[len(_fields)-1]
							} else {
								checkError(fmt.Errorf(`column "%s" not existed in file: %s`, before, file))
							}
							copy(record2[at:], record2[at-1:len(record2)-1])
							record2[at-1] = value
						} else if at > 0 && at <= len(record2) {
							copy(record2[at:], record2[at-1:len(record2)-1])
							record2[at-1] = value
						}

						break
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
	RootCmd.AddCommand(mutateCmd)
	mutateCmd.Flags().StringP("fields", "f", "1", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
	mutateCmd.Flags().StringP("pattern", "p", "^(.+)$", `search regular expression with capture bracket. e.g.`)
	mutateCmd.Flags().StringP("name", "n", "", `new column name`)
	mutateCmd.Flags().BoolP("ignore-case", "i", false, "ignore case")
	mutateCmd.Flags().BoolP("na", "", false, "for unmatched data, use blank instead of original data")
	mutateCmd.Flags().BoolP("remove", "R", false, `remove input column`)
	mutateCmd.Flags().IntP("at", "", 0, "where the new column should appear, 1 for the 1st column, 0 for the last column")
	mutateCmd.Flags().StringP("after", "", "", "insert the new column right after the given column name")
	mutateCmd.Flags().StringP("before", "", "", "insert the new column right before the given column name")

}
