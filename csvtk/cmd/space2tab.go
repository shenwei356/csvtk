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
	"strings"

	"github.com/shenwei356/breader"
	"github.com/shenwei356/util/stringutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// space2tabCmd represents the space2tab command
var space2tabCmd = &cobra.Command{
	Use:   "space2tab",
	Short: "convert space delimited format to CSV",
	Long: `convert space delimited format to CSV

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		type Slice []string
		fn := func(line string) (interface{}, bool, error) {
			line = strings.TrimRight(line, "\r\n")
			// check comment line
			if len(strings.TrimSpace(line)) == 0 || rune(line[0]) == config.CommentChar {
				return "", false, nil
			}
			return Slice(stringutil.Split(line, "\t ")), true, nil
		}
		for _, file := range files {
			reader, err := breader.NewBufferedReader(file, config.NumCPUs, config.ChunkSize, fn)
			checkError(err)

			for chunk := range reader.Ch {
				for _, data := range chunk.Data {
					items := data.(Slice)
					outfh.WriteString(fmt.Sprintf("%s\n", strings.Join(items, "\t")))
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(space2tabCmd)
}
