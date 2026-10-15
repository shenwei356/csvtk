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
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

// csv2xlsxCmd represents the seq command
var csv2xlsxCmd = &cobra.Command{
	GroupID: "format",

	Use:   "csv2xlsx",
	Short: "convert CSV/TSV files to XLSX file",
	Long: `convert CSV/TSV files to XLSX file

Attention:

  1. Multiple CSV/TSV files are saved as separated sheets in .xlsx file.
  2. All input files should all be CSV or TSV.
  3. First rows are freezed unless given '-H/--no-header-row'.
  4. XLSX sheets support up to 16,384 columns and 1,048,576 rows, including the
     output header row and any row-number column. Exceeding these limits causes an error.
  
`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)

		formatNumbers := getFlagBool(cmd, "format-numbers")

		runtime.GOMAXPROCS(config.NumCPUs)

		singleInput := len(files) == 1

		outFile := config.OutFile
		if isStdin(outFile) {
			if singleInput && !isStdin(files[0]) {
				outFile = files[0] + ".xlsx"
			} else {
				outFile = "stdin.xlsx"
			}
		}

		xlsx := excelize.NewFile()
		defer closeOutput(xlsx)

		var sheet, cell, val string
		var col, line int
		var valFloat float64
		var nSheets int
		var idx, firstIdx int
		for i, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk csv2xlsx: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}
			nSheets++

			csvReader.Read(ReadOption{
				FieldStr:      "1-",
				ShowRowNumber: config.ShowRowNumber,
			})

			if singleInput {
				sheet = "Sheet1"

				firstIdx, err = xlsx.GetSheetIndex(sheet)
				checkError(err)
			} else {
				sheet, _ = filepathTrimExtension(filepath.Base(file))
				if nSheets == 1 {
					checkError(xlsx.SetSheetName("Sheet1", sheet))
				} else {
					idx, err = xlsx.NewSheet(sheet)
					checkError(err)

					if i == 0 {
						firstIdx = idx
					}
				}
			}

			if !config.NoHeaderRow && !config.NoOutHeader {
				checkError(xlsx.SetPanes(sheet, &excelize.Panes{
					Freeze:      true,
					Split:       false,
					XSplit:      0,
					YSplit:      1,
					TopLeftCell: "A2",
					ActivePane:  "bottomLeft",
				}))
			}

			line = 1
			handleHeaderRow := !config.NoHeaderRow
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if handleHeaderRow {
					handleHeaderRow = false
					if config.NoOutHeader {
						continue
					}
				}

				for col, val = range record.Selected {
					// Validate XLSX limits and convert the zero-based CSV column index.
					cell, err = excelize.CoordinatesToCellName(col+1, line)
					checkError(err)
					if formatNumbers {
						valFloat, err = strconv.ParseFloat(val, 64)
						if err != nil {
							err = xlsx.SetCellValue(sheet, cell, val)
						} else {
							err = xlsx.SetCellFloat(sheet, cell, valFloat, -1, 64)
						}
					} else {
						err = xlsx.SetCellValue(sheet, cell, val)
					}
					// A failed cell write must not produce a successful, incomplete workbook.
					checkError(err)
				}
				line++
			}

			readerReport(&config, csvReader, file)
		}

		xlsx.SetActiveSheet(firstIdx)
		checkError(xlsx.SaveAs(outFile))
	},
}

func init() {
	RootCmd.AddCommand(csv2xlsxCmd)

	csv2xlsxCmd.Flags().BoolP("format-numbers", "f", false, `save numbers in number format, instead of text`)

}
