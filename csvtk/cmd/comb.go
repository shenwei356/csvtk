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
	Short:   "compute combinations of items in each row",
	Long: `compute combinations of items in each row

Rows may have different numbers of fields. Quoted fields may span multiple lines.
Empty fields and repeated items within each row are ignored.

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
		defer closeOutput(outfh)

		outOpt := config.csvOutputOptions()

		writer := newCSVOutputWriter(outfh, outOpt)
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

		var item string
		var items []string
		var m map[string]interface{}
		var comb []int // Reused indices for the current combination.
		var result []string
		var n int
		var ok bool
		var i int
		var number int

		items = make([]string, 1024)
		result = make([]string, 1024)
		m = make(map[string]interface{}, 1024)

		for _, file := range files {
			reader, err := newCSVReaderByConfig(config, file)
			if err != nil {
				checkError(fmt.Errorf("reading file %s: %s", file, err))
			}
			// Each record is an independent set and may contain a different number
			// of fields. The shared reader handles quoted newlines and closes inputs.
			reader.Reader.FieldsPerRecord = -1
			reader.Read(ReadOption{FieldStr: "1-"})
			n = 0
			for record := range reader.Ch {
				checkError(record.Err)
				if record.IsHeaderRow {
					continue
				}
				n++

				items = items[:0]
				for _, item = range record.All {
					if item == "" {
						continue
					}
					if ignoreCase {
						item = strings.ToLower(item)
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

				// Generate one combination at a time instead of retaining the entire set.
				generator := combin.NewCombinationGenerator(len(items), number)
				if cap(comb) < number {
					comb = make([]int, number)
				} else {
					comb = comb[:number]
				}
				for generator.Next() {
					generator.Combination(comb)
					result = result[:0]
					for _, i = range comb {
						result = append(result, items[i])
					}

					if sortItems {
						sort.Strings(result)
					} else if sortItemsNatSort {
						natsort.Sort(result)
					}
					checkError(writer.Write(result))
				}

				for item = range m {
					delete(m, item)
				}
			}
			readerReport(&config, reader, file)
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
	combCmd.Flags().BoolP("ignore-case", "i", false, "ignore case")
}
