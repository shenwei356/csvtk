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
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/shenwei356/breader"
	"github.com/shenwei356/util/stringutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"github.com/twotwotwo/sorts"
)

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	GroupID: "order",

	Use:   "sort",
	Short: "sort by selected fields",
	Long: `sort by selected fields

Field types:
  - Column name                  : -k name
  - Nth field                    : -k 1
  - field range:
     - All fields                : -k 1-
     - From 3rd to 5th column    : -k 3-5
     - Fields except for the 4th : -k -4
     - Fields except for 3rd-5th : -k -3--5

Sort types:
  - Default: alphabetical order  : -k 1-
  - n      : numeric order       : -k 1:n
  - N      : natural order       : -k 1:N   (e.g., "a9" should be in front of "a10")
  - d      : sort by date        : -k 1:d   (support multiple formats of date and time)
  - u      : custom levels       : -k 1:u -L levels.txt

Combinations:
  - All sort types can be used with "r" for reversing the order, e.g., -k 1:nr
  - Multiple fields can be used, e.g., -k year:n -k name

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
				if config.Verbose {
					log.Warningf("overide user-defined level for field %s", items[0])
				}
			}
			levelsMap[items[0]] = m
		}

		sortTypes := []sortType{}
		fieldsStrs := []string{}
		var i int
		var _key, _type string
		for _, key := range keys {
			i = strings.LastIndexByte(key, ':')
			if i < 0 || i == len(key)-1 {
				_key = key
				fieldsStrs = append(fieldsStrs, _key)
				sortTypes = append(sortTypes, sortType{FieldStr: _key, Number: false, Reverse: false})
			} else if i == 0 {
				checkError(fmt.Errorf(`invalid key: "%s"`, key))
			} else {
				_key = key[:i]
				fieldsStrs = append(fieldsStrs, _key)
				_type = key[i+1:]
				switch _type {
				case "N":
					sortTypes = append(sortTypes, sortType{FieldStr: _key, Natural: true, Reverse: false})
				case "Nr", "rN":
					sortTypes = append(sortTypes, sortType{FieldStr: _key, Natural: true, Reverse: true})
				case "n":
					sortTypes = append(sortTypes, sortType{FieldStr: _key, Number: true, Reverse: false})
				case "r":
					sortTypes = append(sortTypes, sortType{FieldStr: _key, Number: false, Reverse: true})
				case "nr", "rn":
					sortTypes = append(sortTypes, sortType{FieldStr: _key, Number: true, Reverse: true})
				case "d":
					sortTypes = append(sortTypes, sortType{FieldStr: _key, Date: true, Reverse: false})
				case "dr":
					sortTypes = append(sortTypes, sortType{FieldStr: _key, Date: true, Reverse: true})
				case "u":
					if _, ok := levelsMap[_key]; !ok {
						checkError(fmt.Errorf("level file not provided for field: %s", _key))
					}
					sortTypes = append(sortTypes, sortType{FieldStr: _key, Number: false, Reverse: false, UserDefined: true, Levels: levelsMap[_key]})
				case "ur", "ru":
					if _, ok := levelsMap[_key]; !ok {
						checkError(fmt.Errorf("level file not provided for field: %s", _key))
					}
					sortTypes = append(sortTypes, sortType{FieldStr: _key, Number: false, Reverse: true, UserDefined: true, Levels: levelsMap[_key]})
				default:
					// checkError(fmt.Errorf("invalid sort type: %s", _type))
					_key = key
					fieldsStrs[len(fieldsStrs)-1] = _key
					sortTypes = append(sortTypes, sortType{FieldStr: _key, Number: false, Reverse: false})
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

		file := files[0]
		colnames, fields, _, headerRow, data, err := parseCSVfile(cmd, config,
			file, fieldsStr, fuzzyFields, false, true)

		if err != nil {
			if err == xopen.ErrNoContent {
				if config.Verbose {
					log.Warningf("csvtk sort: skipping empty input file: %s", file)
				}
				return
			}
			checkError(err)
		}

		if len(data) == 0 {
			log.Warningf("no data to sort from file: %s", file)
			if len(headerRow) > 0 && !config.NoOutHeader {
				checkError(writer.Write(headerRow))
			}
			return
		}

		// checking keys
		_m := make(map[string]interface{}, len(fields))
		for _, f := range fields {
			_m[strconv.Itoa(f)] = struct{}{}
		}
		for _, f := range colnames {
			_m[f] = struct{}{}
		}

		var list []stringutil.MultiKeyStringSlice // data

		sortTypes2 := make([]stringutil.SortType, 0, len(sortTypes))
		var field int
		ncols := len(data[0])
		_fields := make([]int, 0, ncols)
		var start, end int
		var found [][]string
		var ok bool
		var col string
		for _, t := range sortTypes {
			_fields = _fields[:0]

			if reIntegerRange.MatchString(t.FieldStr) { // field range
				found = reIntegerRange.FindAllStringSubmatch(t.FieldStr, -1)
				start, err = strconv.Atoi(found[0][1])
				if err != nil {
					checkError(fmt.Errorf("fail to parse field range: %s. it should be an integer", found[0][1]))
				}

				if found[0][2] == "" {
					end = ncols
				} else {
					end, err = strconv.Atoi(found[0][2])
					if err != nil {
						checkError(fmt.Errorf("fail to parse field range: %s. it should be an integer", found[0][2]))
					}
				}
				if start == 0 || end == 0 {
					checkError(fmt.Errorf("no 0 allowed in field range: %s", t.FieldStr))
				}

				if start < 0 && end < 0 {
					if start < end {
						for i = 1; i <= ncols; i++ {
							if i < -end || i > -start {
								_fields = append(_fields, i-1)
							}
						}
					} else {
						for i = 1; i <= ncols; i++ {
							if i < -start || i > -end {
								_fields = append(_fields, i-1)
							}
						}
					}
				} else if start > 0 && end > 0 {
					if start >= end {
						checkError(fmt.Errorf("invalid field range: %s. start (%d) should be less than end (%d)", t.FieldStr, start, end))
					}
					for i = start; i <= end; i++ {
						_fields = append(_fields, i-1)
					}
				} else {
					checkError(fmt.Errorf("invalid field range: %s. start (%d) and end (%d) should be both > 0 or < 0", t.FieldStr, start, end))
				}

			} else { // a single field
				if reDigitals.MatchString(t.FieldStr) { // field number, might be negative
					field, _ = strconv.Atoi(t.FieldStr)
					if field > 0 {
						_fields = append(_fields, field-1)
					} else {
						for i = 1; i <= ncols; i++ {
							if i != -field {
								_fields = append(_fields, i-1)
							}
						}
					}
				} else {
					if _, ok = _m[t.FieldStr]; !ok {
						checkError(fmt.Errorf("filed %s not matched in file: %s", t.FieldStr, file))
					}

					if len(headerRow) > 0 {
						if reDigitals.MatchString(t.FieldStr) {
							field, err = strconv.Atoi(t.FieldStr)
							checkError(err)
							field--
						} else {
							for i, col = range headerRow {
								if col == t.FieldStr {
									field = i
									break
								}
							}
						}
					} else {
						field, err = strconv.Atoi(t.FieldStr)
						checkError(err)
						field--
					}

					_fields = append(_fields, field)
				}
			}

			for _, field = range _fields { // multiple values if given a field range such as 1-
				sortTypes2 = append(sortTypes2,
					stringutil.SortType{
						Index:       field,
						IgnoreCase:  ignoreCase,
						Natural:     t.Natural,
						Number:      t.Number,
						Date:        t.Date,
						Reverse:     t.Reverse,
						UserDefined: t.UserDefined,
						Levels:      t.Levels,
					})
				// fmt.Println(sortTypes2[len(sortTypes2)-1])
			}
		}

		list = make([]stringutil.MultiKeyStringSlice, len(data))
		for i, record := range data {
			list[i] = stringutil.MultiKeyStringSlice{SortTypes: &sortTypes2, Value: record}
		}
		sorts.Quicksort(stringutil.MultiKeyStringSliceList(list))

		if len(headerRow) > 0 && !config.NoOutHeader {
			checkError(writer.Write(headerRow))
		}
		for _, s := range list {
			checkError(writer.Write(s.Value))
		}

	},
}

type sortType struct {
	FieldStr    string
	Natural     bool
	Number      bool
	Date        bool
	Reverse     bool
	UserDefined bool
	Levels      map[string]int
}

func init() {
	RootCmd.AddCommand(sortCmd)
	sortCmd.Flags().StringSliceP("keys", "k", []string{"1-"}, `keys (multiple values supported). sort type supported, "N" for natural order, "n" for number, "d" for date/time, "u" for user-defined order and "r" for reverse. e.g., "-k 1", "-k 2-", "-k 3-5:nr", "-k A:r", "-k 1:nr -k 2"`)
	sortCmd.Flags().StringSliceP("levels", "L", []string{}, `user-defined level file (one level per line, multiple values supported). format: <field>:<level-file>.  e.g., "-k name:u -L name:level.txt"`)
	sortCmd.Flags().BoolP("ignore-case", "i", false, "ignore-case")
}
