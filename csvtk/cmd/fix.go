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
	"encoding/csv"
	"fmt"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// fixCmd represents the pretty command
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "fix CSV/TSV with different numbers of columns in rows",
	Long: `fix CSV/TSV with different numbers of columns in rows

How to:
  1. First -n/--buf-rows rows are read to check the maximum number of columns.
     The default value 0 means all rows will be read.
  2. Buffered and remaining rows with fewer columns are appended with empty
     cells before output.
  3. An error will be reported if the number of columns of any remaining row
     is larger than the maximum number of columns.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		bufRows := getFlagNonNegativeInt(cmd, "buf-rows")

		var buf [][]string
		var readAll bool
		if bufRows > 0 {
			buf = make([][]string, 0, bufRows)
		} else {
			readAll = true
			buf = make([][]string, 0, 1024)
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			if config.OutDelimiter == ',' { // default value, no other value given
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
				log.Warningf("csvtk pretty: skipping empty input file: %s", file)
				return
			}
			checkError(err)
		}

		// very important.
		// If FieldsPerRecord is negative, no check is made and
		// records may have a variable number of fields.
		csvReader.Reader.FieldsPerRecord = -1

		csvReader.Run()

		var record []string
		var n int // number of loaded rows
		var maxN int
		var checkedMaxNcols bool
		var row []string
		var ncol int
		var empty []string
		for chunk := range csvReader.Ch {
			checkError(chunk.Err)

			for _, record = range chunk.Data {
				n++

				if readAll {
					buf = append(buf, record)
					continue
				}

				buf = append(buf, record)
				if !checkedMaxNcols {
					if n == bufRows {
						maxN = maxNcols(buf)
						log.Infof("the maximum number of columns in first %d rows: %d", bufRows, maxN)
						checkedMaxNcols = true
						empty = make([]string, maxN)

						for _, row = range buf {
							ncol = len(row)
							if ncol < maxN {
								row = append(row, empty[0:maxN-ncol]...)
							}
							writer.Write(row)
						}
					}

					continue
				}

				ncol = len(record)
				if ncol > maxN {
					checkError(fmt.Errorf("line %d: the number of columns is larger than %d, please increase the value of -n/--buf-rows (%d)", n, maxN, bufRows))
				} else if ncol < maxN {
					record = append(record, empty[0:maxN-ncol]...)
				}
				writer.Write(record)
			}
		}

		if readAll || !checkedMaxNcols {
			maxN = maxNcols(buf)
			empty = make([]string, maxN)
			log.Infof("the maximum number of columns in all %d rows: %d", len(buf), maxN)

			for _, row = range buf {
				ncol = len(row)
				if ncol < maxN {
					row = append(row, empty[0:maxN-ncol]...)
				}
				writer.Write(row)
			}
		}

		writer.Flush()
		checkError(writer.Error())

		readerReport(&config, csvReader, file)
	},
}

func maxNcols(buf [][]string) int {
	maxN := 0
	var ncol int
	for _, row := range buf {
		ncol = len(row)
		if ncol > maxN {
			maxN = ncol
		}
	}
	return maxN
}

func init() {
	RootCmd.AddCommand(fixCmd)

	fixCmd.Flags().IntP("buf-rows", "n", 0, "the number of rows to determine the maximum number of columns. 0 for all rows.")
}
