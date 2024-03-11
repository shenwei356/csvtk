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
	"fmt"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// addHeaderCmd represents the addHeader command
var addHeaderCmd = &cobra.Command{
	GroupID: "edit",

	Use:   "add-header",
	Short: "add column names",
	Long: `add column names

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		colnames := getFlagStringSlice(cmd, "names")
		if len(colnames) == 0 {
			if config.Verbose {
				log.Warningf("colnames not given, c1, c2, c3... will be used")
			}
		}

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

		printHeaderRow := true
		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk add-header: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr: "1-",
			})

			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if printHeaderRow {
					if len(colnames) == 0 {
						colnames = make([]string, len(record.All))
						for i := 0; i < len(record.All); i++ {
							colnames[i] = fmt.Sprintf("c%d", i+1)
						}
					} else if len(colnames) != len(record.All) {
						checkError(fmt.Errorf("number of fields (%d) and new colnames (%d) do not match", len(record.All), len(colnames)))
					}
					checkError(writer.Write(colnames))
					printHeaderRow = false
				}

				checkError(writer.Write(record.All))
			}

			readerReport(&config, csvReader, file)
		}

		if printHeaderRow { // did not print rowname
			checkError(writer.Write(colnames))
		}

	},
}

func init() {
	RootCmd.AddCommand(addHeaderCmd)

	addHeaderCmd.Flags().StringSliceP("names", "n", []string{}, `column names to add, in CSV format`)
}
