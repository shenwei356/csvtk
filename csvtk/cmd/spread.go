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
	"strings"

	"github.com/shenwei356/util/stringutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// spreadCmd represents the gather command
var spreadCmd = &cobra.Command{
	Use: "spread",

	Aliases: []string{"wider"},

	Short: "spread a key-value pair across multiple columns, like tidyr::spread/pivot_wider",
	Long: `spread a key-value pair across multiple columns, like tidyr::spread/pivot_wider

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		fieldKey := getFlagString(cmd, "key")
		fieldValue := getFlagString(cmd, "value")
		if !config.NoHeaderRow {
			if fieldKey == "" {
				checkError(fmt.Errorf("flag -k/--key needed"))
			}
			if fieldValue == "" {
				checkError(fmt.Errorf("flag -v/--value needed"))
			}
		}
		na := getFlagString(cmd, "na")
		separater := getFlagString(cmd, "separater")

		fieldStr := fieldKey + "," + fieldValue
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
		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				log.Warningf("csvtk gather: skipping empty input file: %s", file)

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

		var fieldsMap map[int]interface{}
		var f int
		var left, key, val string
		var vals []string
		var ok bool
		var items []string
		data := make(map[string]map[string][]string) // other column -> key -> []value
		keysMap := make(map[string]interface{}, 128)
		keysOrder := make(map[string]int, 128)
		var nKey int
		groupOrder := make(map[string]int, 128)

		checkFirstLine := true
		var handleHeaderRow bool
		var HeaderRow []string
		var nLeft int // number of coulmns except the key and value columns

		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if len(record.All) < 2 {
				checkError(fmt.Errorf("input data should have at least two columns"))
			}

			if len(record.Fields) != 2 {
				checkError(fmt.Errorf("only exactly one key field and one value field are allowed"))
			}
			if record.Fields[0] == record.Fields[1] {
				checkError(fmt.Errorf("key field and value field should be different"))
			}

			fieldsMap = make(map[int]interface{}, len(record.Selected))
			for _, f = range record.Fields {
				fieldsMap[f-1] = struct{}{}
			}

			items = make([]string, 0, len(record.All)-2)

			if checkFirstLine {
				checkFirstLine = false

				nLeft = len(record.All) - 2

				if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
					handleHeaderRow = true
				}
			}

			items = items[:0]
			for f, val = range record.All {
				if _, ok = fieldsMap[f]; !ok {
					items = append(items, val)
				}
			}

			if handleHeaderRow {
				handleHeaderRow = false
				HeaderRow = make([]string, len(items))
				copy(HeaderRow, items)

				continue
			}

			left = strings.Join(items, "_shenwei356_")

			if _, ok = groupOrder[left]; !ok {
				groupOrder[left] = record.Row
			}

			key, val = record.Selected[0], record.Selected[1]

			if _, ok = data[left]; !ok {
				data[left] = make(map[string][]string, 8)
			}

			if _, ok = data[left][key]; !ok {
				data[left][key] = []string{val}
			} else {
				// log.Warningf("duplicated record: %s (%s) for %s at line %d", key, val, strings.Join(items, ","), record.Line)
				data[left][key] = append(data[left][key], val)
			}

			if _, ok = keysMap[key]; !ok {
				keysMap[key] = struct{}{}

				nKey++
				keysOrder[key] = nKey
			}
		}

		keys := make([]string, 0, len(keysMap))
		for _, o := range stringutil.SortCountOfString(keysOrder, false) {
			keys = append(keys, o.Key)
		}

		if HeaderRow == nil {
			HeaderRow = make([]string, nLeft)
		}
		checkError(writer.Write(append(HeaderRow, keys...)))

		var m map[string][]string

		for _, o := range stringutil.SortCountOfString(groupOrder, false) {
			items = strings.Split(o.Key, "_shenwei356_")
			m = data[o.Key]

			for _, key = range keys {
				if vals, ok = m[key]; ok {
					items = append(items, strings.Join(vals, separater))
				} else {
					items = append(items, na)
				}
			}

			checkError(writer.Write(items))
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(spreadCmd)

	spreadCmd.Flags().StringP("key", "k", "", `field of the key. e.g -k 1 or -k columnA`)
	spreadCmd.Flags().StringP("value", "v", "", `field of the value. e.g -v 1 or -v columnA`)
	spreadCmd.Flags().StringP("na", "", "", "content for filling NA data")
	spreadCmd.Flags().StringP("separater", "s", "; ", "separater for values that share the same key")
}
