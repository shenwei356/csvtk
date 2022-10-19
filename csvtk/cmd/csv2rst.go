// Copyright © 2016-2021 Wei Shen <shenwei356@gmail.com>
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
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// csv2rstCmd represents the pretty command
var csv2rstCmd = &cobra.Command{
	Use:   "csv2rst",
	Short: "convert CSV to reStructuredText format",
	Long: `convert CSV to readable aligned table

Attention:

  1. row span is not supported.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		cross := getFlagString(cmd, "cross")
		padding := getFlagString(cmd, "padding")
		borderX := getFlagString(cmd, "horizontal-border")
		borderY := getFlagString(cmd, "vertical-border")
		header := getFlagString(cmd, "header")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		file := files[0]
		var colnames []string
		headerRow, data, csvReader, err := readCSV(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				log.Warningf("csvtk csv2rst: skipping empty input file: %s", file)
				return
			}
			checkError(err)
		}

		// compute maximum length of each column
		var maxLens []int
		var i, l int
		var r string
		var record []string
		var ncolsP1 int
		if len(headerRow) > 0 {
			maxLens = make([]int, len(headerRow))
			for i, r = range headerRow {
				// l = len(r)
				l = runewidth.StringWidth(r)

				maxLens[i] = l
			}
			colnames = headerRow
			ncolsP1 = len(colnames) - 1
		} else if len(data) == 0 {
			// checkError(fmt.Errorf("no data found in file: %s", file))
			log.Warningf("no data found in file: %s", file)
			readerReport(&config, csvReader, file)
			return
		} else {
			maxLens = make([]int, len(data[0]))
			ncolsP1 = len(data[0]) - 1
		}
		for _, record = range data {
			for i, r = range record {
				// l = len(r)
				l = runewidth.StringWidth(r)

				if l > maxLens[i] {
					maxLens[i] = l
				}
			}
		}

		// output
		wPadding := len(padding)

		// top border
		outfh.WriteString(cross)
		for i, l = range maxLens {
			outfh.WriteString(strings.Repeat(borderX, l+wPadding<<1))
			outfh.WriteString(cross)
			if i == ncolsP1 {
				outfh.WriteString("\n")
			}
		}

		// header row
		if len(colnames) > 0 {
			outfh.WriteString(borderY)
			for i, r = range colnames {
				// l = len(r)
				l = runewidth.StringWidth(r)

				outfh.WriteString(padding + r + padding + strings.Repeat(" ", maxLens[i]-l))
				outfh.WriteString(borderY)
				if i == ncolsP1 { // not the last column
					outfh.WriteString("\n")
				}
			}

			outfh.WriteString(cross)
			for i, l = range maxLens {
				outfh.WriteString(strings.Repeat(header, l+wPadding<<1))
				outfh.WriteString(cross)
				if i == ncolsP1 {
					outfh.WriteString("\n")
				}
			}
		}

		// data row
		for _, row := range data {
			outfh.WriteString(borderY)
			for i, r = range row {
				// l = len(r)
				l = runewidth.StringWidth(r)

				outfh.WriteString(padding + r + padding + strings.Repeat(" ", maxLens[i]-l))
				outfh.WriteString(borderY)
				if i == ncolsP1 { // not the last column
					outfh.WriteString("\n")
				}
			}

			outfh.WriteString(cross)
			for i, l = range maxLens {
				outfh.WriteString(strings.Repeat(borderX, l+wPadding<<1))
				outfh.WriteString(cross)
				if i == ncolsP1 {
					outfh.WriteString("\n")
				}
			}
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(csv2rstCmd)

	csv2rstCmd.Flags().StringP("cross", "k", "+", "charactor of cross")
	csv2rstCmd.Flags().StringP("padding", "p", " ", "charactor of padding")
	csv2rstCmd.Flags().StringP("horizontal-border", "b", "-", "charactor of horizontal border")
	csv2rstCmd.Flags().StringP("vertical-border", "B", "|", "charactor of vertical border")
	csv2rstCmd.Flags().StringP("header", "s", "=", "charactor of separator between header row and data rowws")

}
