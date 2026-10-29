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
	"strconv"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// roundCmd represents the round command
var roundCmd = &cobra.Command{
	GroupID: "edit",

	Use:   "round",
	Short: "round floats to n decimal places",
	Long: `round floats to n decimal places

Scientific notation is preserved; -n rounds the mantissa.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		decimalWidth := getFlagNonNegativeInt(cmd, "decimal-width")
		decimalFormat := fmt.Sprintf("%%.%df", decimalWidth)

		allFields := getFlagBool(cmd, "all-fields")

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}
		if allFields {
			fieldStr = "1-"
		}

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer closeOutput(outfh)

		outOpt := config.csvOutputOptions()

		writer := newCSVOutputWriter(outfh, outOpt)
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
						log.Warningf("csvtk round: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr:    fieldStr,
				FuzzyFields: fuzzyFields,
			})

			var found []string
			var fvalue float64

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

				for _, f := range record.Fields {
					found = reDigitalsCapt.FindStringSubmatch(record.All[f-1])
					if found != nil {
						// Round the mantissa without changing the exponent's spelling.
						fvalue, err = strconv.ParseFloat(found[1], 64)
						checkError(err)
						record.All[f-1] = fmt.Sprintf(decimalFormat, fvalue) + found[2]
					}
				}
				checkError(writer.Write(record.All))
			}

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	RootCmd.AddCommand(roundCmd)
	roundCmd.Flags().StringP("fields", "f", "1", `select only these fields. e.g., -f 1,2 or -f columnA,columnB`)
	roundCmd.Flags().BoolP("all-fields", "a", false, "all fields, overrides -f/--fields")
	roundCmd.Flags().BoolP("fuzzy-fields", "F", false, `use fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	roundCmd.Flags().IntP("decimal-width", "n", 2, "limit floats to N decimal places")

}
