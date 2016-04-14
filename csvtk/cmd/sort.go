// Copyright © 2016 Wei Shen <shenwei356@gmail.com>
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

	"github.com/brentp/xopen"
	"github.com/shenwei356/util/stringutil"
	"github.com/spf13/cobra"
)

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	Use:   "sort",
	Short: "sort by selected fields",
	Long: ` sort by selected fields

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)
		runtime.GOMAXPROCS(config.NumCPUs)

		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}

		keys := getFlagStringSlice(cmd, "keys")
		sortTypes := []sortType{}
		fieldsStrs := []string{}
		var items []string
		for _, key := range keys {
			items = strings.Split(key, ":")
			if len(items) == 1 {
				fieldsStrs = append(fieldsStrs, items[0])
				sortTypes = append(sortTypes, sortType{items[0], false, false})
			} else {
				fieldsStrs = append(fieldsStrs, items[0])
				switch items[1] {
				case "n":
					sortTypes = append(sortTypes, sortType{items[0], true, false})
				case "r":
					sortTypes = append(sortTypes, sortType{items[0], false, true})
				case "nr", "rn":
					sortTypes = append(sortTypes, sortType{items[0], true, true})
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
		headerRow, data, _ := parseCSVfile(cmd, config,
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
			sortTypes2[i] = stringutil.SortType{Index: field, Number: t.Number, Reverse: t.Reverse}
		}

		list = make([]stringutil.MultiKeyStringSlice, len(data))
		for i, record := range data {
			list[i] = stringutil.MultiKeyStringSlice{SortTypes: &sortTypes2, Value: record}
		}
		sort.Sort(stringutil.MultiKeyStringSliceList(list))

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
	FieldStr string
	Number   bool
	Reverse  bool
}

func init() {
	RootCmd.AddCommand(sortCmd)
	sortCmd.Flags().StringSliceP("keys", "k", []string{"1"}, `keys. sort type supported, "n" for number and "r" for reverse. e.g. "-k 1" or "-k A:r" or ""-k 1:nr -k 2"`)
}
