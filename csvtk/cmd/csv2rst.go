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
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// csv2rstCmd represents the pretty command
var csv2rstCmd = &cobra.Command{
	GroupID: "format",

	Use:   "csv2rst",
	Short: "convert CSV to reStructuredText format",
	Long: `convert CSV to a readable aligned table

Attention:

  1. Row spans are not supported.
  2. Line breaks within cells are preserved inside the table borders.

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
		defer closeOutput(outfh)

		file := files[0]
		var colnames []string
		headerRow, data, csvReader, err := readCSV(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				if config.Verbose {
					log.Warningf("csvtk csv2rst: skipping empty input file: %s", file)
				}
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
				l = rstCellWidth(r)

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
				l = rstCellWidth(r)

				if l > maxLens[i] {
					maxLens[i] = l
				}
			}
		}

		// output
		wPadding := runewidth.StringWidth(padding)

		// top border
		outfh.WriteString(cross)
		for i, l = range maxLens {
			outfh.WriteString(strings.Repeat(borderX, l+wPadding<<1))
			outfh.WriteString(cross)
			if i == ncolsP1 {
				outfh.WriteString("\n")
				config.flushOutput(outfh)
			}
		}

		// Write wrapped cell lines inside the same logical row.
		writeRow := func(row []string) {
			lines := make([][]string, len(row))
			height := 1
			for i, cell := range row {
				lines[i] = strings.Split(cell, "\n")
				if len(lines[i]) > height {
					height = len(lines[i])
				}
			}
			for line := 0; line < height; line++ {
				outfh.WriteString(borderY)
				for i, cellLines := range lines {
					cell := ""
					if line < len(cellLines) {
						cell = cellLines[line]
					}
					outfh.WriteString(padding)
					outfh.WriteString(cell)
					outfh.WriteString(padding)
					outfh.WriteString(strings.Repeat(" ", maxLens[i]-runewidth.StringWidth(cell)))
					outfh.WriteString(borderY)
				}
				outfh.WriteByte('\n')
				config.flushOutput(outfh)
			}
		}

		// Header separators appear only when the header is written.
		if len(colnames) > 0 && !config.NoOutHeader {
			writeRow(colnames)
			outfh.WriteString(cross)
			for i, l = range maxLens {
				outfh.WriteString(strings.Repeat(header, l+wPadding<<1))
				outfh.WriteString(cross)
				if i == ncolsP1 {
					outfh.WriteString("\n")
					config.flushOutput(outfh)
				}
			}
		}

		// data row
		for _, row := range data {
			writeRow(row)

			outfh.WriteString(cross)
			for i, l = range maxLens {
				outfh.WriteString(strings.Repeat(borderX, l+wPadding<<1))
				outfh.WriteString(cross)
				if i == ncolsP1 {
					outfh.WriteString("\n")
					config.flushOutput(outfh)
				}
			}
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(csv2rstCmd)

	csv2rstCmd.Flags().StringP("cross", "k", "+", "character for intersections")
	csv2rstCmd.Flags().StringP("padding", "p", " ", "padding character")
	csv2rstCmd.Flags().StringP("horizontal-border", "b", "-", "horizontal border character")
	csv2rstCmd.Flags().StringP("vertical-border", "B", "|", "vertical border character")
	csv2rstCmd.Flags().StringP("header", "s", "=", "separator character between the header row and data rows")

}

// rstCellWidth returns the display width of a multiline cell's longest line.
func rstCellWidth(cell string) int {
	width := 0
	for _, line := range strings.Split(cell, "\n") {
		if n := runewidth.StringWidth(line); n > width {
			width = n
		}
	}
	return width
}
