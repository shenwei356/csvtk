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
	"bytes"
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

Attention:

  pretty treats the first row as header line and requires them to be unique

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
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
		var colnames []string
		headerRow, data, csvReader := readCSV(config, file)

		if len(headerRow) > 0 {
			colnames = headerRow
		} else {
			if len(data) == 0 {
				checkError(fmt.Errorf("no data found in file: %s", file))
			} else if len(data) > 0 {
				colnames = make([]string, len(data[0]))
				for i := 0; i < len(data[0]); i++ {
					colnames[i] = fmt.Sprintf("%d", i+1)
				}
			}
		}

		columns := make([]prettytable.Column, len(colnames))
		for i, c := range colnames {
			columns[i] = prettytable.Column{Header: c, AlignRight: alignRight,
				MinWidth: minWidth, MaxWidth: maxWidth}
		}
		tbl, err := prettytable.NewTable(columns...)
		checkError(err)
		tbl.Separator = separator
		for _, record := range data {
			// have to do this stupid conversion
			record2 := make([]interface{}, len(record))
			for i, r := range record {
				record2[i] = r
			}
			tbl.AddRow(record2...)
		}

		if config.NoHeaderRow {
			output := tbl.Bytes()
			outfh.Write(output[bytes.IndexByte(output, '\n')+1:])
		} else {
			outfh.Write(tbl.Bytes())
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(prettyCmd)
	prettyCmd.Flags().StringP("separator", "s", "   ", "fields/columns separator")
	prettyCmd.Flags().BoolP("align-right", "r", false, "align right")
	prettyCmd.Flags().IntP("min-width", "w", 0, "min width")
	prettyCmd.Flags().IntP("max-width", "W", 0, "max width")
}
