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
	"encoding/csv"
	"fmt"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// transposeCmd represents the transpose command
var transposeCmd = &cobra.Command{
	Use:   "transpose",
	Short: "transpose CSV data",
	Long: `transpose CSV data

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		data := [][]string{}

		var numCols0, numCols, numRows uint64
		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			checkError(err)
			csvReader.Run()

			once := true

			printMetaLine := true
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				if printMetaLine && len(csvReader.MetaLine) > 0 {
					if config.OutTabs || config.Tabs {
						outfh.WriteString(fmt.Sprintf("sep=%s\n", "\t"))
					} else {
						outfh.WriteString(fmt.Sprintf("sep=%s\n", string(config.OutDelimiter)))
					}
					printMetaLine = false
				}

				numRows += uint64(len(chunk.Data))
				for _, record := range chunk.Data {
					data = append(data, record)

					if once {
						numCols = uint64(len(record))
						if numCols0 == 0 {
							numCols0 = numCols
						} else if numCols0 != numCols {
							checkError(fmt.Errorf("unmartched number of columns between files"))
						}
						once = false
					}
				}
			}

			readerReport(&config, csvReader, file)
		}

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}
		for j := uint64(0); j < numCols0; j++ {
			rowNew := make([]string, numRows)
			for i, rowOld := range data {
				rowNew[i] = rowOld[j]
			}
			checkError(writer.Write(rowNew))
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(transposeCmd)
}
