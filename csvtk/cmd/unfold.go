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
	GroupID: "transform",

	Use:   "unfold",
	Short: "unfold multiple values in cells of one or more fields",
	Long: `unfold multiple values in cells of one or more fields

When multiple fields are selected, their values are unfolded in parallel.
Each selected field must have the same number of values in every row.

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

    $ echo -ne "key,en,es\nfoo,one;two,uno;due\n" \
        | csvtk unfold -f en,es -s ";" \
        | csvtk pretty
    key   en    es
    foo   one   uno
    foo   two   due

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
		defer func() {
			writer.Flush()
			checkError(writer.Error())
		}()

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk unfold: skipping empty input file: %s", file)
					}
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
					if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
						if config.NoOutHeader {
							continue
						}
						checkError(writer.Write(record.All))
						continue
					}
				}

				values := make([][]string, len(record.Selected))
				for i, selected := range record.Selected {
					values[i] = strings.Split(selected, separater)
					if i > 0 && len(values[i]) != len(values[0]) {
						checkError(fmt.Errorf("row %d: selected fields have different numbers of values (field %d: %d, field %d: %d)", record.Row, record.Fields[0], len(values[0]), record.Fields[i], len(values[i])))
					}
				}

				for j := range values[0] {
					for i, field := range record.Fields {
						record.All[field-1] = values[i][j]
					}
					checkError(writer.Write(record.All))
				}
			}

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	RootCmd.AddCommand(unfoldCmd)

	unfoldCmd.Flags().StringP("fields", "f", "", `fields to expand in parallel. type "csvtk unfold -h" for examples`)
	unfoldCmd.Flags().StringP("separater", "s", "; ", "separater for folded values")
}
