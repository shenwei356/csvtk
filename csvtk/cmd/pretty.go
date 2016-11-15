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
	"fmt"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"github.com/tatsushid/go-prettytable"
)

// prettyCmd represents the pretty command
var prettyCmd = &cobra.Command{
	Use:   "pretty",
	Short: "convert CSV to readable aligned table",
	Long: `convert CSV to readable aligned table

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		alignRight := getFlagBool(cmd, "align-right")
		separator := getFlagString(cmd, "separator")
		minWidth := getFlagNonNegativeInt(cmd, "min-width")
		maxWidth := getFlagNonNegativeInt(cmd, "max-width")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		file := files[0]
		fieldStr := "*"
		fuzzyFields := true
		headerRow, _, data, _, _ := parseCSVfile(cmd, config,
			file, fieldStr, fuzzyFields)

		var header []string
		var datas [][]string
		if len(headerRow) > 0 {
			header = headerRow
			datas = data
		} else {
			if len(data) == 0 {
				checkError(fmt.Errorf("no data found in file: %s", file))
			} else if len(data) > 0 {
				header = data[0]
				datas = data[1:]
			}
		}
		columns := make([]prettytable.Column, len(header))
		for i, c := range header {
			columns[i] = prettytable.Column{Header: c, AlignRight: alignRight,
				MinWidth: minWidth, MaxWidth: maxWidth}
		}
		tbl, err := prettytable.NewTable(columns...)
		checkError(err)
		tbl.Separator = separator
		for _, record := range datas {
			// have to do this stupid conversion
			record2 := make([]interface{}, len(record))
			for i, r := range record {
				record2[i] = r
			}
			tbl.AddRow(record2...)
		}
		outfh.Write(tbl.Bytes())

	},
}

func init() {
	RootCmd.AddCommand(prettyCmd)
	prettyCmd.Flags().StringP("separator", "s", "   ", "fields/columns separator")
	prettyCmd.Flags().BoolP("align-right", "r", false, "align right")
	prettyCmd.Flags().IntP("min-width", "w", 0, "min width")
	prettyCmd.Flags().IntP("max-width", "W", 0, "max width")
}
