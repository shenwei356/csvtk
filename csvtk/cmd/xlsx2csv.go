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

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// xlsx2csvCmd represents the seq command
var xlsx2csvCmd = &cobra.Command{
	Use:   "xlsx2csv",
	Short: "convert XLSX to CSV format",
	Long: `convert XLSX to CSV format

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		if files[0] == "-" {
			checkError(fmt.Errorf("stdin not supported for xlsx2csv"))
		}

		runtime.GOMAXPROCS(config.NumCPUs)

		listSheets := getFlagBool(cmd, "list-sheets")
		sheetName := getFlagString(cmd, "sheet-name")
		sheetIndex := getFlagPositiveInt(cmd, "sheet-index")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		xlsx, err := excelize.OpenFile(files[0])
		checkError(err)

		sheets := xlsx.GetSheetMap()

		if listSheets {
			if len(sheets) > 0 {
				fmt.Println("index\tsheet")
				is := make([]int, len(sheets))
				i := 0
				for index := range sheets {
					is[i] = index
					i++
				}
				sort.Ints(is)
				for _, index := range is {
					fmt.Printf("%d\t%s\n", index, sheets[index])
				}
			}
			return
		}

		if sheetName == "" {
			sheetName = sheets[sheetIndex]
		} else {
			var existed bool
			for _, sheet := range sheets {
				if sheet == sheetName {
					existed = true
				}
			}
			if !existed {
				checkError(fmt.Errorf(`Sheet '%s' does not exist in file: %s`, sheetName, files[0]))
			}
		}

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}

		rows, err := xlsx.GetRows(sheetName)
		checkError(err)

		var notBlank bool
		var data string
		var numEmptyRows int
		for _, row := range rows {
			if config.IgnoreEmptyRow {
				notBlank = false
				for _, data = range row {
					if data != "" {
						notBlank = true
						break
					}
				}
				if !notBlank {
					numEmptyRows++
					continue
				}
			}
			checkError(writer.Write(row))
		}

		writer.Flush()
		checkError(writer.Error())

		if config.IgnoreEmptyRow {
			log.Warningf("file '%s': %d empty rows ignored", files[0], numEmptyRows)
		}
	},
}

func init() {
	RootCmd.AddCommand(xlsx2csvCmd)

	xlsx2csvCmd.Flags().StringP("sheet-name", "n", "", "sheet to retrieve")
	xlsx2csvCmd.Flags().BoolP("list-sheets", "a", false, "list all sheets")
	xlsx2csvCmd.Flags().IntP("sheet-index", "i", 1, "Nth sheet to retrieve")
}
