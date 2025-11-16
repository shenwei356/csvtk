// Copyright © 2016-2025 Wei Shen <shenwei356@gmail.com>
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

// shufCmd represents the sort command
var shufCmd = &cobra.Command{
	GroupID: "order",

	Use:   "shuf",
	Short: "shuffle rows",
	Long: `shuffle rows

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		number := getFlagNonNegativeInt(cmd, "rows")

		seed := getFlagInt64(cmd, "rand-seed")
		_rand := rand.New(rand.NewSource(seed))

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

		file := files[0]
		_, _, _, headerRow, data, err := parseCSVfile(cmd, config,
			file, "1-", false, false, true)
		if err != nil {
			if err == xopen.ErrNoContent {
				if config.Verbose {
					log.Warningf("csvtk sort: skipping empty input file: %s", file)
				}
				return
			}
			checkError(err)
		}

		if len(headerRow) > 0 && !config.NoOutHeader {
			checkError(writer.Write(headerRow))
		}

		if len(data) == 0 {
			log.Warningf("no data to sort from file: %s", file)
			return
		}

		_rand.Shuffle(len(data), func(i, j int) {
			data[i], data[j] = data[j], data[i]
		})

		if number > 0 {
			var i int
			for _, row := range data {
				i++
				checkError(writer.Write(row))

				if i == number {
					break
				}
			}
		} else {
			for _, row := range data {
				checkError(writer.Write(row))
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(shufCmd)

	shufCmd.Flags().Int64P("rand-seed", "s", 11, "rand seed")
	shufCmd.Flags().IntP("rows", "n", 0, "print first N rows, 0 for all")
}
