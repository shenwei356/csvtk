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

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// nrowCmd represents the nrow command
var nrowCmd = &cobra.Command{
	Use:     "nrow",
	Aliases: []string{"nrows"},
	Short:   "print number of records",
	Long: `print number of records

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		printFileName := getFlagBool(cmd, "file-name")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		for _, file := range files {
			var numRows int

			csvReader, err := newCSVReaderByConfig(config, file)
			if err != nil {
				if err == xopen.ErrNoContent {
					if printFileName {
						outfh.WriteString(fmt.Sprintf("%d\t%s\n", numRows, file))
					} else {
						outfh.WriteString(fmt.Sprintf("%d\n", numRows))
					}
					outfh.Flush()

					continue
				} else {
					checkError(err)
				}
			}

			csvReader.Read(ReadOption{
				FieldStr: "1-",
			})
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				numRows = record.Row
			}

			if printFileName {
				outfh.WriteString(fmt.Sprintf("%d\t%s\n", numRows, file))
			} else {
				outfh.WriteString(fmt.Sprintf("%d\n", numRows))
			}
			outfh.Flush()

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	nrowCmd.Flags().BoolP("file-name", "n", false, "print file names")

	RootCmd.AddCommand(nrowCmd)

}
