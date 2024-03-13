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
	"runtime"
	"strconv"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// headCmd represents the seq command
var headCmd = &cobra.Command{
	GroupID: "set",

	Use:   "head",
	Short: "print first N records",
	Long: `print first N records

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		number := getFlagPositiveInt(cmd, "number")

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
		defer func() {
			writer.Flush()
			checkError(writer.Error())
		}()

		printLineNumber := config.ShowRowNumber

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk head: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr: "1-",
			})

			isHeaderLine := !config.NoHeaderRow
			i := 0
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if isHeaderLine {
					isHeaderLine = false
					if config.NoOutHeader {
						continue
					}
					if printLineNumber {
						unshift(&record.All, "row")
					}
				} else {
					i++
					if printLineNumber {
						unshift(&record.All, strconv.Itoa(record.Row))
					}
				}

				checkError(writer.Write(record.All))

				if number == i {
					break
				}
			}

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	RootCmd.AddCommand(headCmd)

	headCmd.Flags().IntP("number", "n", 10, "print first N records")
}
