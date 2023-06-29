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
	"bufio"
	"fmt"
	"runtime"
	"strings"

	"github.com/shenwei356/util/stringutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// space2tabCmd represents the space2tab command
var space2tabCmd = &cobra.Command{
	Use:   "space2tab",
	Short: "convert space delimited format to TSV",
	Long: `convert space delimited format to TSV

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		bufferSizeS := getFlagString(cmd, "buffer-size")
		if bufferSizeS == "" {
			checkError(fmt.Errorf("value of buffer size. supported unit: K, M, G"))
		}
		bufferSize, err := ParseByteSize(bufferSizeS)
		if err != nil {
			checkError(fmt.Errorf("invalid value of buffer size. supported unit: K, M, G"))
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		buf := make([]byte, bufferSize)

		var scanner *bufio.Scanner
		var line string
		for _, file := range files {
			fh, err := xopen.Ropen(file)
			checkError(err)

			scanner = bufio.NewScanner(fh)
			scanner.Buffer(buf, int(bufferSize))
			for scanner.Scan() {
				line = strings.TrimRight(scanner.Text(), "\r\n")
				if len(strings.TrimSpace(line)) == 0 || rune(line[0]) == config.CommentChar {
					continue
				}
				outfh.WriteString(strings.Join(stringutil.Split(line, "\t "), "\t") + "\n")
			}
			checkError(scanner.Err())
		}
	},
}

func init() {
	RootCmd.AddCommand(space2tabCmd)

	space2tabCmd.Flags().StringP("buffer-size", "b", "1G", `size of buffer, supported unit: K, M, G. You need increase the value when "bufio.Scanner: token too long" error reported`)
}
