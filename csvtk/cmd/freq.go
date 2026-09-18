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
	"slices"
	"sort"
	"strconv"

	"github.com/shenwei356/util/stringutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// freqCmd represents the freq command
var freqCmd = &cobra.Command{
	GroupID: "set",

	Use:   "freq",
	Short: "frequencies of selected fields",
	Long: `frequencies of selected fields

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		sortByFreq := getFlagBool(cmd, "sort-by-freq")
		sortByKey := getFlagBool(cmd, "sort-by-key")
		reverse := getFlagBool(cmd, "reverse")

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
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

		counter := make(map[string]int, 10000)
		orders := make(map[string]int, 10000)

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				if config.Verbose {
					log.Warningf("csvtk freq: skipping empty input file: %s", file)
				}

				writer.Flush()
				checkError(writer.Error())
				readerReport(&config, csvReader, file)
				return
			}
			checkError(err)
		}

		csvReader.Read(ReadOption{
			FieldStr:    fieldStr,
			FuzzyFields: fuzzyFields,

			DoNotAllowDuplicatedColumnName: true,
		})

		var key string
		var N int

		checkFirstLine := true
		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if checkFirstLine {
				checkFirstLine = false
				if !config.NoHeaderRow || record.IsHeaderRow {
					if config.NoOutHeader {
						continue
					}
					checkError(writer.Write(append(record.Selected, "frequency")))
					continue
				}
			}

			N++

			key = encodeFields(record.Selected, ignoreCase)
			counter[key]++
			orders[key] = N
		}

		var items []string
		var keyFields map[string][]string
		if sortByFreq || sortByKey {
			keyFields = make(map[string][]string, len(counter))
			for key := range counter {
				keyFields[key] = decodeFields(key)
			}
		}
		if sortByFreq {
			counts := make([]stringutil.StringCount, len(counter))
			i := 0
			for key, count := range counter {
				counts[i] = stringutil.StringCount{Key: key, Count: count}
				i++
			}
			sort.Slice(counts, func(i, j int) bool {
				if counts[i].Count != counts[j].Count {
					if reverse {
						return counts[i].Count > counts[j].Count
					}
					return counts[i].Count < counts[j].Count
				}
				return slices.Compare(keyFields[counts[i].Key], keyFields[counts[j].Key]) < 0
			})
			for _, count := range counts {
				items = slices.Clone(keyFields[count.Key])
				items = append(items, strconv.Itoa(counter[count.Key]))
				checkError(writer.Write(items))
			}
		} else if sortByKey {
			keys := make([]string, len(counter))
			i := 0
			for key := range counter {
				keys[i] = key
				i++
			}

			sort.Slice(keys, func(i, j int) bool {
				return slices.Compare(keyFields[keys[i]], keyFields[keys[j]]) < 0
			})
			if reverse {
				stringutil.ReverseStringSliceInplace(keys)
			}

			for _, key := range keys {
				items = slices.Clone(keyFields[key])
				items = append(items, strconv.Itoa(counter[key]))
				checkError(writer.Write(items))
			}
		} else {
			orderedKey := stringutil.SortCountOfString(orders, false)
			for _, o := range orderedKey {
				items = decodeFields(o.Key)
				items = append(items, strconv.Itoa(counter[o.Key]))
				checkError(writer.Write(items))
			}
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(freqCmd)
	freqCmd.Flags().StringP("fields", "f", "1", `select these fields as the key. e.g -f 1,2 or -f columnA,columnB`)
	freqCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	freqCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	freqCmd.Flags().BoolP("sort-by-freq", "n", false, `sort by frequency`)
	freqCmd.Flags().BoolP("sort-by-key", "k", false, `sort by key`)
	freqCmd.Flags().BoolP("reverse", "r", false, `reverse order while sorting`)
}
