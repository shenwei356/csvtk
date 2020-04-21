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
	"math/rand"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// sampleCmd represents the seq command
var sampleCmd = &cobra.Command{
	Use:   "sample",
	Short: "sampling by proportion",
	Long: `sampling by proportion

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		proportion := getFlagFloat64(cmd, "proportion")
		printLineNumber := getFlagBool(cmd, "line-number")

		if proportion == 0 {
			checkError(fmt.Errorf("flag -p (--proportion) needed"))
		}
		if proportion <= 0 || proportion > 1 {
			checkError(fmt.Errorf("value of -p (--proportion) (%f) should be in range of (0, 1]", proportion))
		}

		outAll := proportion == 1

		seed := getFlagInt64(cmd, "rand-seed")
		rand.Seed(seed)

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			checkError(err)
			csvReader.Run()

			var N int64
			var recordWithN []string

			isHeaderLine := !config.NoHeaderRow
			printMetaLine := true
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				if printMetaLine && len(csvReader.MetaLine) > 0 {
					outfh.WriteString(fmt.Sprintf("sep=%s\n", string(writer.Comma)))
					printMetaLine = false
				}

				for _, record := range chunk.Data {
					if isHeaderLine {
						if printLineNumber {
							recordWithN = []string{"n"}
							recordWithN = append(recordWithN, record...)
							record = recordWithN
						}

						checkError(writer.Write(record))
						isHeaderLine = false
						continue
					}

					N++

					if outAll || rand.Float64() <= proportion {
						if printLineNumber {
							recordWithN = []string{fmt.Sprintf("%d", N)}
							recordWithN = append(recordWithN, record...)
							record = recordWithN
						}

						checkError(writer.Write(record))
					}
				}
			}

			readerReport(&config, csvReader, file)
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(sampleCmd)

	sampleCmd.Flags().Int64P("rand-seed", "s", 11, "rand seed")
	sampleCmd.Flags().Float64P("proportion", "p", 0, "sample by proportion")
	sampleCmd.Flags().BoolP("line-number", "n", false, `print line number as the first column ("n")`)
}
