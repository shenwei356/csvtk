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
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

// csv2xlsxCmd represents the seq command
var csv2xlsxCmd = &cobra.Command{
	Use:   "csv2xlsx",
	Short: "convert CSV/TSV files to XLSX file",
	Long: `convert CSV/TSV files to XLSX file

Attention:

  1. Multiple CSV/TSV files are saved as separated sheets in .xlsx file.
  2. All input files should all be CSV or TSV.
  3. First rows are freezed unless given '-H/--no-header-row'.
  
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
		defer checkError(xlsx.Close())

		var sheet, cell, val string
		var col, line int
		var valFloat float64
		var nSheets int
		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			if err != nil {
				if err == xopen.ErrNoContent {
					log.Warningf("csvtk csv2xlsx: skipping empty input file: %s", file)
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
			} else {
				sheet, _ = filepathTrimExtension(filepath.Base(file))
				if nSheets == 1 {
					xlsx.SetSheetName("Sheet1", sheet)
				} else {
					xlsx.NewSheet(sheet)
				}
			}

			if !config.NoHeaderRow {
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
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}
				for col, val = range record.Selected {
					cell = fmt.Sprintf("%s%d", ExcelColumnIndex(col), line)
					if formatNumbers {
						valFloat, err = strconv.ParseFloat(val, 64)
						if err != nil {
							xlsx.SetCellValue(sheet, cell, val)
						} else {
							xlsx.SetCellFloat(sheet, cell, valFloat, -1, 64)
						}
					} else {
						xlsx.SetCellValue(sheet, cell, val)
					}
				}
				line++
			}

			readerReport(&config, csvReader, file)
		}

		xlsx.SetActiveSheet(1)
		checkError(xlsx.SaveAs(outFile))
	},
}

func init() {
	RootCmd.AddCommand(csv2xlsxCmd)

	csv2xlsxCmd.Flags().BoolP("format-numbers", "f", false, `save numbers in number format, instead of text`)

}

func ExcelColumnIndex(col int) string {
	s := make([]byte, 0, 8)

	s = append(s, byte(col%26+65))
	for col >= 26 {
		col = col / 26
		s = append(s, byte(col%26+64))
	}

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return string(s)
}
