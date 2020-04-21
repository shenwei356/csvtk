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
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/shenwei356/breader"
	"github.com/shenwei356/util/stringutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	Use:   "sort",
	Short: "sort by selected fields",
	Long: `sort by selected fields

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		levels := getFlagStringSlice(cmd, "levels")
		keys := getFlagStringSlice(cmd, "keys")
		ignoreCase := getFlagBool(cmd, "ignore-case")

		levelsMap := make(map[string]map[string]int)
		var items []string
		for _, level := range levels {
			items = strings.Split(level, ":")
			if len(items) != 2 {
				checkError(fmt.Errorf("invalid level information format: %s", level))
			}

			m := make(map[string]int)
			reader, err := breader.NewDefaultBufferedReader(items[1])
			checkError(errors.Wrap(err, "read level file"))
			var i int
			for chunk := range reader.Ch {
				checkError(chunk.Err)
				for _, data := range chunk.Data {
					line := data.(string)
					if line == "" {
						continue
					}
					i++
					if ignoreCase {
						m[strings.ToLower(line)] = i
					} else {
						m[line] = i
					}
				}
			}
			if _, ok := levelsMap[items[0]]; ok {
				log.Warningf("overide user-defined level for field %s", items[0])
			}
			levelsMap[items[0]] = m
		}

		sortTypes := []sortType{}
		fieldsStrs := []string{}
		for _, key := range keys {
			items = strings.Split(key, ":")
			if len(items) == 1 {
				fieldsStrs = append(fieldsStrs, items[0])
				sortTypes = append(sortTypes, sortType{FieldStr: items[0], Number: false, Reverse: false})
			} else {
				if items[0] == "" {
					checkError(fmt.Errorf(`invalid key: "%s"`, key))
				}
				fieldsStrs = append(fieldsStrs, items[0])
				switch items[1] {
				case "N":
					sortTypes = append(sortTypes, sortType{FieldStr: items[0], Natural: true, Reverse: false})
				case "Nr", "rN":
					sortTypes = append(sortTypes, sortType{FieldStr: items[0], Natural: true, Reverse: true})
				case "n":
					sortTypes = append(sortTypes, sortType{FieldStr: items[0], Number: true, Reverse: false})
				case "r":
					sortTypes = append(sortTypes, sortType{FieldStr: items[0], Number: false, Reverse: true})
				case "nr", "rn":
					sortTypes = append(sortTypes, sortType{FieldStr: items[0], Number: true, Reverse: true})
				case "u":
					if _, ok := levelsMap[items[0]]; !ok {
						checkError(fmt.Errorf("level file not provided for field: %s", items[0]))
					}
					sortTypes = append(sortTypes, sortType{FieldStr: items[0], Number: false, Reverse: false, UserDefined: true, Levels: levelsMap[items[0]]})
				case "ur", "ru":
					if _, ok := levelsMap[items[0]]; !ok {
						checkError(fmt.Errorf("level file not provided for field: %s", items[0]))
					}
					sortTypes = append(sortTypes, sortType{FieldStr: items[0], Number: false, Reverse: true, UserDefined: true, Levels: levelsMap[items[0]]})
				default:
					checkError(fmt.Errorf("invalid sort type: %s", items[1]))
				}
			}
		}
		fieldsStr := strings.Join(fieldsStrs, ",")

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

		file := files[0]
		_, _, _, headerRow, data, metaLine := parseCSVfile(cmd, config,
			file, fieldsStr, fuzzyFields)
		if len(data) == 0 {
			checkError(fmt.Errorf("no data to sort"))
		}

		var list []stringutil.MultiKeyStringSlice // data

		sortTypes2 := make([]stringutil.SortType, len(sortTypes))
		var field int
		for i, t := range sortTypes {
			if len(headerRow) > 0 {
				if reDigitals.MatchString(t.FieldStr) {
					field, err = strconv.Atoi(t.FieldStr)
					checkError(err)
					field--
				} else {
					for f, col := range headerRow {
						if col == t.FieldStr {
							field = f
							break
						}
					}
				}
			} else {
				field, err = strconv.Atoi(t.FieldStr)
				checkError(err)
				field--
			}
			sortTypes2[i] = stringutil.SortType{Index: field,
				IgnoreCase:  ignoreCase,
				Natural:     t.Natural,
				Number:      t.Number,
				Reverse:     t.Reverse,
				UserDefined: t.UserDefined,
				Levels:      t.Levels,
			}
		}

		list = make([]stringutil.MultiKeyStringSlice, len(data))
		for i, record := range data {
			list[i] = stringutil.MultiKeyStringSlice{SortTypes: &sortTypes2, Value: record}
		}
		sort.Sort(stringutil.MultiKeyStringSliceList(list))

		if len(metaLine) > 0 {
			outfh.WriteString(fmt.Sprintf("sep=%s\n", string(writer.Comma)))
		}
		if len(headerRow) > 0 {
			checkError(writer.Write(headerRow))
		}
		for _, s := range list {
			checkError(writer.Write(s.Value))
		}

		writer.Flush()
		checkError(writer.Error())
	},
}

type sortType struct {
	FieldStr    string
	Natural     bool
	Number      bool
	Reverse     bool
	UserDefined bool
	Levels      map[string]int
}

func init() {
	RootCmd.AddCommand(sortCmd)
	sortCmd.Flags().StringSliceP("keys", "k", []string{"1"}, `keys (multiple values supported). sort type supported, "N" for natural order, "n" for number, "u" for user-defined order and "r" for reverse. e.g., "-k 1" or "-k A:r" or ""-k 1:nr -k 2"`)
	sortCmd.Flags().StringSliceP("levels", "L", []string{}, `user-defined level file (one level per line, multiple values supported). format: <field>:<level-file>.  e.g., "-k name:u -L name:level.txt"`)
	sortCmd.Flags().BoolP("ignore-case", "i", false, "ignore-case")
}
