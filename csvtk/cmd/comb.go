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
	"bufio"
	"encoding/csv"
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/shenwei356/natsort"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/stat/combin"
)

// combCmd represents the comb command
var combCmd = &cobra.Command{
	GroupID: "set",

	Use:     "comb",
	Aliases: []string{"combination"},
	Short:   "compute combinations of items at every row",
	Long: `compute combinations of items at every row

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		sortItems := getFlagBool(cmd, "sort")
		sortItemsNatSort := getFlagBool(cmd, "nat-sort")
		number0 := getFlagNonNegativeInt(cmd, "number")
		ignoreCase := getFlagBool(cmd, "ignore-case")

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

		var fh *xopen.Reader
		var text string
		var reader *csv.Reader
		var line int
		var item string
		var _items, items []string
		var m map[string]interface{}
		var combs [][]int
		var comb []int
		var result []string
		var n int
		var ok bool
		var i int
		var number int

		items = make([]string, 1024)
		result = make([]string, 1024)
		m = make(map[string]interface{}, 1024)

		for _, file := range files {
			fh, err = xopen.Ropen(file)
			if err != nil {
				checkError(fmt.Errorf("reading file %s: %s", file, err))
			}

			scanner := bufio.NewScanner(fh)

			line = 0
			n = 0
			for scanner.Scan() {
				line++
				if !config.NoHeaderRow && line == 1 {
					continue
				}
				n++

				text = strings.TrimSpace(scanner.Text())
				if ignoreCase {
					text = strings.ToLower(text)
				}

				reader = csv.NewReader(strings.NewReader(text))
				if config.Tabs {
					reader.Comma = '\t'
				} else {
					reader.Comma = config.Delimiter
				}

				reader.Comment = config.CommentChar
				for {
					_items, err = reader.Read()
					if err != nil {
						checkError(fmt.Errorf("[line %d] failed parsing: %s", line, text))
					}
					break
				}
				items = items[:0]
				for _, item = range _items {
					if item == "" {
						continue
					}
					if _, ok = m[item]; !ok {
						items = append(items, item)
						m[item] = struct{}{}
					}
				}

				if len(items) == 0 {
					continue
				}

				number = number0
				if number == 0 || number > len(items) {
					number = len(items)
				}

				combs = combin.Combinations(len(items), number)
				for _, comb = range combs {
					result = result[:0]
					for _, i = range comb {
						result = append(result, items[i])
					}

					if sortItems {
						sort.Strings(result)
					} else if sortItemsNatSort {
						natsort.Sort(result)
					}
					writer.Write(result)
				}

				for item = range m {
					delete(m, item)
				}
			}
			checkError(scanner.Err())
			if n == 0 {
				log.Warning("no input? or only one row? you may need switch on '-H' for single-line input")
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(combCmd)

	combCmd.Flags().BoolP("sort", "s", false, `sort items in a combination`)
	combCmd.Flags().BoolP("nat-sort", "S", false, `sort items in natural order`)
	combCmd.Flags().IntP("number", "n", 2, `number of items in a combination, 0 for no limit, i.e., return all combinations`)
	combCmd.Flags().BoolP("ignore-case", "i", false, "ignore-case")
}
