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

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// mutateCmd represents the mutate command
var mutateCmd = &cobra.Command{
	GroupID: "edit",

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
		if !config.NoHeaderRow && name == "" && !config.NoOutHeader {
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
					if config.Verbose {
						log.Warningf("csvtk mutate: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr: fieldStr,

				DoNotAllowDuplicatedColumnName: true,
			})

			var record2 []string // for output
			var _fields []int
			var fieldsMap map[int]interface{}
			var colnames2fileds map[string][]int // column name -> []field
			var ok bool
			var f int
			var value string
			var handleHeaderRow bool
			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					checkFirstLine = false

					if len(record.Fields) > 1 {
						checkError(fmt.Errorf("only a single field allowed"))
					}

					fieldsMap = make(map[int]interface{}, len(record.Selected))
					for _, f = range record.Fields {
						fieldsMap[f-1] = struct{}{}
					}

					if !config.NoHeaderRow || record.IsHeaderRow {
						handleHeaderRow = true

						colnames2fileds = make(map[string][]int, len(record.All))
						for i, col := range record.All {
							if _, ok := colnames2fileds[col]; !ok {
								colnames2fileds[col] = []int{i + 1}
							} else {
								colnames2fileds[col] = append(colnames2fileds[col], i+1)
							}
						}
					}
				}

				if remove {
					record2 = make([]string, 0, len(record.All))
					for f = range record.All {
						if _, ok = fieldsMap[f]; !ok {
							record2 = append(record2, record.All[f])
						}
					}
				} else {
					record2 = record.All
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
							at = _fields[0]
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
					if !config.NoOutHeader {
						checkError(writer.Write(record2))
					}
					continue
				}

				f = record.Fields[0] - 1

				if patternRegexp.MatchString(record.All[f]) {
					found := patternRegexp.FindAllStringSubmatch(record.All[f], -1)
					value = found[0][1]
				} else {
					if naUnmatched {
						value = ""
					} else {
						value = record.All[f]
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
						at = _fields[0]
					} else {
						checkError(fmt.Errorf(`column "%s" not existed in file: %s`, before, file))
					}
					copy(record2[at:], record2[at-1:len(record2)-1])
					record2[at-1] = value
				} else if at > 0 && at <= len(record2) {
					copy(record2[at:], record2[at-1:len(record2)-1])
					record2[at-1] = value
				}

				checkError(writer.Write(record2))
			}

			readerReport(&config, csvReader, file)
		}
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
