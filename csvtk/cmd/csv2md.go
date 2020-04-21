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
	"fmt"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"github.com/tatsushid/go-prettytable"
)

// csv2mdCmd represents the csv2md command
var csv2mdCmd = &cobra.Command{
	Use:   "csv2md",
	Short: "convert CSV to markdown format",
	Long: `convert CSV to markdown format

Attention:

  csv2md treats the first row as header line and requires them to be unique

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		separator := "|"
		aligns := getFlagCommaSeparatedStrings(cmd, "alignments")
		minWidth := getFlagNonNegativeInt(cmd, "min-width")
		if minWidth < 3 {
			checkError(fmt.Errorf("value of -w (--min-width) should not be less than 3"))
		}
		if len(aligns) == 0 {
			checkError(fmt.Errorf("flag -a (--alignments) needed"))
		}
		for _, a := range aligns {
			switch a {
			case "c", "center":
			case "l", "left":
			case "r", "right":
			default:
				checkError(fmt.Errorf("invalid alignment: %s", a))
			}
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		file := files[0]
		headerRow, data, csvReader := readCSV(config, file)

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

		if len(aligns) == 1 {
			if len(header) > 1 {
				aligns2 := make([]string, len(header))
				for i := range header {
					aligns2[i] = aligns[0]
				}
				aligns = aligns2
			}
		} else if len(aligns) != len(header) {
			checkError(fmt.Errorf("number of alignment symbols (%d) should be equal to 1 or number of fields (%d)", len(aligns), len(header)))
		}

		widths := make([]int, len(header))
		for i, c := range header {
			if len(c) < minWidth {
				widths[i] = minWidth
			} else {
				widths[i] = len(c)
			}
		}

		for _, data := range datas {
			for j, c := range data {
				if len(c) > widths[j] {
					widths[j] = len(c)
				}
			}
		}

		alignRow := make([]string, len(header))
		var l, r, a string
		for i, w := range widths {
			switch aligns[i] {
			case "c", "center":
				l, r = ":", ":"
			case "l", "left":
				l, r = ":", "-"
			case "r", "right":
				l, r = "-", ":"
			}
			a = l
			for j := 0; j < w-2; j++ {
				a += "-"
			}
			a += r
			alignRow[i] = a
		}

		columns := make([]prettytable.Column, len(header))
		for i, c := range header {
			columns[i] = prettytable.Column{Header: c, AlignRight: false, MinWidth: minWidth}
		}
		tbl, err := prettytable.NewTable(columns...)
		checkError(err)
		tbl.Separator = separator

		record2 := make([]interface{}, len(alignRow))
		for i, c := range alignRow {
			record2[i] = c
		}
		tbl.AddRow(record2...)
		for _, record := range datas {
			// have to do this stupid conversion
			record2 := make([]interface{}, len(record))
			for i, c := range record {
				record2[i] = c
			}
			tbl.AddRow(record2...)
		}
		outfh.Write(tbl.Bytes())

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(csv2mdCmd)
	csv2mdCmd.Flags().StringP("alignments", "a", "l", `comma separated alignments. e.g. -a l,c,c,c or -a c`)
	csv2mdCmd.Flags().IntP("min-width", "w", 3, "min width (at least 3)")
}
