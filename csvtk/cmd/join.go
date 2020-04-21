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
	"runtime"
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:     "join",
	Aliases: []string{"merge"},
	Short:   "join files by selected fields (inner, left and outer join)",
	Long: `join files by selected fields (inner, left and outer join).

Attention:

  1. Multiple keys supported
  2. Default operation is inner join, use --left-join for left join 
     and --outer-join for outer join.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) < 2 {
			checkError(fmt.Errorf("two or more files needed"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)
		allFields := getFlagSemicolonSeparatedStrings(cmd, "fields")
		if len(allFields) == 0 {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		} else if len(allFields) == 1 {
			s := make([]string, len(files))
			for i := range files {
				s[i] = allFields[0]
			}
			allFields = s
		} else if len(allFields) != len(files) {
			checkError(fmt.Errorf("number of fields (%d) should be equal to number of files (%d)", len(allFields), len(files)))
		}

		ignoreCase := getFlagBool(cmd, "ignore-case")

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		leftJoin := getFlagBool(cmd, "left-join")
		keepUnmatched := getFlagBool(cmd, "keep-unmatched")
		outerJoin := getFlagBool(cmd, "outer-join")
		na := getFlagString(cmd, "na")

		if outerJoin && leftJoin {
			checkError(fmt.Errorf("flag -O/--out-join and -L/--left-join are exclusive"))
		}

		if outerJoin {
			keepUnmatched = true
			for _, file := range files {
				if isStdin(file) {
					checkError(fmt.Errorf("stdin not allowed when using -O/--outer-join"))
				}
			}
		}
		if leftJoin {
			keepUnmatched = true
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

		var HeaderRow []string
		var Data [][]string
		var Fields []int
		firstFile := true
		var withHeaderRow bool

		var key string
		var items []string

		var keys map[string]bool
		if outerJoin {
			keys = make(map[string]bool)
			for i, file := range files {
				_, fields, _, _, data, _ := parseCSVfile(cmd, config,
					file, allFields[i], fuzzyFields)

				var f int
				var ok bool
				for _, record := range data {
					items = make([]string, len(fields))
					for i, f = range fields {
						items[i] = record[f-1]
					}
					key = strings.Join(items, "_shenwei356_")
					if ignoreCase {
						key = strings.ToLower(key)
					}
					if _, ok = keys[key]; ok {
						continue
					}
					keys[key] = false
				}
			}
		}

		var f int
		var ok bool
		for i, file := range files {
			_, fields, _, headerRow, data, _ := parseCSVfile(cmd, config,
				file, allFields[i], fuzzyFields)
			if firstFile {
				HeaderRow, Data, Fields = headerRow, data, fields
				firstFile = false
				if len(HeaderRow) > 0 {
					withHeaderRow = true
				}

				if !outerJoin {
					continue
				}

				var nCols int
				items = make([]string, len(fields))
				for _, record := range Data {
					nCols = len(record)
					for i, f = range fields {
						items[i] = record[f-1]
					}
					key = strings.Join(items, "_shenwei356_")
					if ignoreCase {
						key = strings.ToLower(key)
					}
					keys[key] = true
				}

				fieldsMap := make(map[int]struct{}, len(fields))
				for _, f := range fields {
					fieldsMap[f] = struct{}{}
				}
				for key, ok = range keys {
					if !ok {
						record := make([]string, nCols)
						items2 := strings.Split(key, "_shenwei356_")
						j := 0
						for i = range record {
							if _, ok = fieldsMap[i+1]; ok {
								record[i] = items2[j]
								j++
							} else {
								record[i] = na
							}
						}
						Data = append(Data, record)
					}
				}

				continue
			}

			// fieldsMap
			fieldsMap := make(map[int]struct{}, len(fields))
			for _, f := range fields {
				fieldsMap[f] = struct{}{}
			}
			// csv to map
			keysMaps := make(map[string][][]string)
			items = make([]string, len(fields))
			for _, record := range data {
				for i, f = range fields {
					items[i] = record[f-1]
				}
				key = strings.Join(items, "_shenwei356_")
				if ignoreCase {
					key = strings.ToLower(key)
				}
				if _, ok = keysMaps[key]; !ok {
					keysMaps[key] = [][]string{}
				}
				keysMaps[key] = append(keysMaps[key], record)
			}

			Data2 := [][]string{}
			var colname string
			if withHeaderRow {
				newHeaderRow := HeaderRow
				for f, colname = range headerRow {
					if _, ok = fieldsMap[f+1]; !ok {
						newHeaderRow = append(newHeaderRow, colname)
					}
				}
				HeaderRow = newHeaderRow
			}
			items = make([]string, len(Fields))
			var records [][]string
			var record, record2 []string
			for _, record0 := range Data {
				for i, f = range Fields {
					items[i] = record0[f-1]
				}
				key = strings.Join(items, "_shenwei356_")
				if ignoreCase {
					key = strings.ToLower(key)
				}
				if records, ok = keysMaps[key]; ok {
					for _, record2 = range records {
						record = record0
						for f, v := range record2 {
							if _, ok = fieldsMap[f+1]; !ok {
								record = append(record, v)
							}
						}
						Data2 = append(Data2, record)
					}
				} else {
					if keepUnmatched {
						record = record0
						for i = 1; i <= len(data[0])-len(fieldsMap); i++ {
							record = append(record, na)
						}
						Data2 = append(Data2, record)
					}
				}
			}
			Data = Data2
		}

		if withHeaderRow {
			checkError(writer.Write(HeaderRow))
		}
		for _, record := range Data {
			checkError(writer.Write(record))
		}

		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(joinCmd)
	joinCmd.Flags().StringP("fields", "f", "1", "Semicolon separated key fields of all files, "+
		`if given one, we think all the files have the same key columns. `+
		`Fields of different files should be separated by ";", e.g -f "1;2" or -f "A,B;C,D" or -f id`)
	joinCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	joinCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	joinCmd.Flags().BoolP("keep-unmatched", "k", false, `keep unmatched data of the first file (left join)`)
	joinCmd.Flags().BoolP("left-join", "L", false, `left join, equals to -k/--keep-unmatched, exclusive with --outer-join`)
	joinCmd.Flags().BoolP("outer-join", "O", false, `outer join, exclusive with --left-join`)
	joinCmd.Flags().StringP("na", "", "", "content for filling NA data")
}
