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
	"sort"
	"strconv"
	"strings"

	"github.com/shenwei356/util/stringutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// freqCmd represents the freq command
var freqCmd = &cobra.Command{
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

		counter := make(map[string]int, 10000)
		orders := make(map[string]int, 10000)

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				log.Warningf("csvtk freq: skipping empty input file: %s", file)

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
				if !config.NoHeaderRow || record.IsHeaderRow {
					checkError(writer.Write(append(record.Selected, "frequency")))
				}
				checkFirstLine = false
				continue
			}

			N++

			key = strings.Join(record.Selected, "_shenwei356_")
			counter[key]++
			orders[key] = N
		}

		var items []string
		if sortByFreq {
			counts := make([]stringutil.StringCount, len(counter))
			i := 0
			for key, count := range counter {
				counts[i] = stringutil.StringCount{Key: key, Count: count}
				i++
			}
			if reverse {
				sort.Sort(stringutil.ReversedStringCountList{counts})
			} else {
				sort.Sort(stringutil.StringCountList(counts))
			}
			for _, count := range counts {
				items = strings.Split(count.Key, "_shenwei356_")
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

			sort.Strings(keys)
			if reverse {
				stringutil.ReverseStringSliceInplace(keys)
			}

			for _, key := range keys {
				items = strings.Split(key, "_shenwei356_")
				items = append(items, strconv.Itoa(counter[key]))
				checkError(writer.Write(items))
			}
		} else {
			orderedKey := stringutil.SortCountOfString(orders, false)
			for _, o := range orderedKey {
				items = strings.Split(o.Key, "_shenwei356_")
				items = append(items, strconv.Itoa(counter[o.Key]))
				checkError(writer.Write(items))
			}
		}

		writer.Flush()
		checkError(writer.Error())

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
