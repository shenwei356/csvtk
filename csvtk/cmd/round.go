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
	"regexp"
	"runtime"
	"strconv"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// roundCmd represents the round command
var roundCmd = &cobra.Command{
	Use:   "round",
	Short: "round float to n decimal places",
	Long: `round float to n decimal places

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
					log.Warningf("csvtk round: skipping empty input file: %s", file)
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr:    fieldStr,
				FuzzyFields: fuzzyFields,
			})

			var found []string
			var founds [][]string
			var fvalue float64

			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
						checkError(writer.Write(record.All))
					}
					checkFirstLine = false
					continue
				}

				for _, f := range record.Fields {
					founds = reDigitalsCapt.FindAllStringSubmatch(record.All[f-1], -1)
					if len(founds) > 0 {
						found = founds[0]
						if found[2] == "" { // not scientific notation
							fvalue, _ = strconv.ParseFloat(found[1], 64)
							record.All[f-1] = fmt.Sprintf(decimalFormat, fvalue)
						} else if found[1] == "" { // e20
						} else {
							fvalue, _ = strconv.ParseFloat(found[1], 64)
							record.All[f-1] = fmt.Sprintf(decimalFormat, fvalue) + found[2]
						}
					}
				}
				checkError(writer.Write(record.All))
			}

			readerReport(&config, csvReader, file)
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(roundCmd)
	roundCmd.Flags().StringP("fields", "f", "1", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
	roundCmd.Flags().BoolP("all-fields", "a", false, "all fields, overides -f/--fields")
	roundCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	roundCmd.Flags().IntP("decimal-width", "n", 2, "limit floats to N decimal points")

}

var reDigitalsCapt = regexp.MustCompile(`^([\-\+\d\.,]*)([eE][\-\+\d]+)?$`)
