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

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// gatherCmd represents the gather command
var gatherCmd = &cobra.Command{
	Use: "longer",

	Aliases: []string{"gather"},

	Short: "gather columns into key-value pairs, like tidyr::gather/pivot_longer",
	Long: `gather columns into key-value pairs, like tidyr::gather/pivot_longer

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		if config.NoHeaderRow {
			checkError(fmt.Errorf("flag -H/--no-header-row not allowed"))
		}

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

		var i, f int
		var ok bool
		var fieldsMap map[int]interface{}
		var items []string
		var fieldsLeft []int
		var HeaderRow []string
		var nFieldsLeft int

		checkFirstLine := true
		var handleHeaderRow bool
		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if checkFirstLine {
				checkFirstLine = false

				if len(record.Fields) == 0 {
					checkError(fmt.Errorf("no fields matched in file: %s", file))
				}

				fieldsMap = make(map[int]interface{}, len(record.Selected))
				for _, f = range record.Fields {
					fieldsMap[f-1] = struct{}{}
				}

				for f = range record.All {
					if _, ok = fieldsMap[f]; !ok {
						fieldsLeft = append(fieldsLeft, f+1)
					}
				}

				nFieldsLeft = len(fieldsLeft)
				items = make([]string, nFieldsLeft+2)

				if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
					handleHeaderRow = true
					HeaderRow = record.All
				}
			}

			// fill columns that are not key or value column
			for i, f = range fieldsLeft {
				items[i] = record.All[f-1]
			}

			if handleHeaderRow {
				items[nFieldsLeft] = fieldKey
				items[nFieldsLeft+1] = fieldValue
				checkError(writer.Write(items))

				handleHeaderRow = false
			} else {
				for _, f = range record.Fields {
					items[nFieldsLeft] = HeaderRow[f-1]
					items[nFieldsLeft+1] = record.All[f-1]
					checkError(writer.Write(items))
				}
			}
		}

		writer.Flush()
		checkError(writer.Error())

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(gatherCmd)
	gatherCmd.Flags().StringP("fields", "f", "", `fields for gathering. e.g -f 1,2 or -f columnA,columnB, or -f -columnA for unselect columnA`)
	gatherCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	gatherCmd.Flags().StringP("key", "k", "", `name of key column to create in output`)
	gatherCmd.Flags().StringP("value", "v", "", `name of value column to create in output`)
}
