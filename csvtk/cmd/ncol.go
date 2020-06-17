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
	"fmt"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// ncolCmd represents the ncol command
var ncolCmd = &cobra.Command{
	Use:     "ncol",
	Aliases: []string{"ncols"},
	Short:   "print number of columns",
	Long: `print number of columns

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
			csvReader, err := newCSVReaderByConfig(config, file)
			checkError(err)
			csvReader.Run()

			var numCols uint64

			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				for _, record := range chunk.Data {
					numCols = uint64(len(record))
					break
				}

				if printFileName {
					outfh.WriteString(fmt.Sprintf("%d\t%s\n", numCols, file))
				} else {
					outfh.WriteString(fmt.Sprintf("%d\n", numCols))
				}
				outfh.Flush()

				break
			}

			readerReport(&config, csvReader, file)

		}
	},
}

func init() {
	ncolCmd.Flags().BoolP("file-name", "n", false, "print file names")

	RootCmd.AddCommand(ncolCmd)

}
