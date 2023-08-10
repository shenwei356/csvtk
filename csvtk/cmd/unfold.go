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
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// unfoldCmd represents the unfold command
var unfoldCmd = &cobra.Command{
	Use:   "unfold",
	Short: "unfold multiple values in cells of a field",
	Long: `unfold multiple values in cells of a field

Example:

    $ echo -ne "id,values,meta\n1,a;b,12\n2,c,23\n3,d;e;f,34\n" \
        | csvtk pretty
    id   values   meta
    1    a;b      12
    2    c        23
    3    d;e;f    34


    $ echo -ne "id,values,meta\n1,a;b,12\n2,c,23\n3,d;e;f,34\n" \
        | csvtk unfold -f values -s ";" \
        | csvtk pretty
    id   values   meta
    1    a        12
    1    b        12
    2    c        23
    3    d        34
    3    e        34
    3    f        34

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}

		separater := getFlagString(cmd, "separater")
		if separater == "" {
			checkError(fmt.Errorf("flag -s (--separater) needed"))
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

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					log.Warningf("csvtk unfold: skipping empty input file: %s", file)
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr: fieldStr,

				DoNotAllowDuplicatedColumnName: true,
			})

			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					checkFirstLine = false
					if len(record.Fields) > 1 {
						checkError(fmt.Errorf("should no choosing more than one field"))
					}

					if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
						checkError(writer.Write(record.All))
						continue
					}
				}

				for _, v := range strings.Split(record.Selected[0], separater) {
					record.All[record.Fields[0]-1] = v
					checkError(writer.Write(record.All))
				}
			}

			readerReport(&config, csvReader, file)
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(unfoldCmd)

	unfoldCmd.Flags().StringP("fields", "f", "", `field to expand, only one field is allowed. type "csvtk unfold -h" for examples`)
	unfoldCmd.Flags().StringP("separater", "s", "; ", "separater for folded values")
}
