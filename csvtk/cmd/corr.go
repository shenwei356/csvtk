// Copyright © 2019 Oxford Nanopore Technologies.
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
	"math"
	"os"
	"runtime"
	"strconv"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/stat"
)

// corrCmd represents the corr command
var corrCmd = &cobra.Command{
	GroupID: "info",

	Use:   "corr",
	Short: "calculate Pearson correlation between two columns",
	Long:  "calculate Pearson correlation between two columns",

	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		printField := getFlagString(cmd, "fields")
		printIgnore := getFlagBool(cmd, "ignore_nan")
		printPass := getFlagBool(cmd, "pass")
		printLog := getFlagBool(cmd, "log")

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

		transform := func(x float64) float64 { return x }
		if printLog {
			transform = func(x float64) float64 {
				return math.Log10(x + 1)
			}
		}

		file := files[0]

		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				log.Warningf("csvtk corr: skipping empty input file: %s", file)
				return
			}
			checkError(err)
		}

		csvReader.Read(ReadOption{
			FieldStr: printField,

			DoNotAllowDuplicatedColumnName: true,
		})

		var data [][]float64
		var i, f int
		var val float64
		var fields []int

		var hasHeaderRow bool
		var HeaderRow []string
		checkFirstLine := true
		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if checkFirstLine {
				checkFirstLine = false

				data = make([][]float64, len(record.Fields))
				for i = range record.Fields {
					data[i] = make([]float64, 0, 1024)
				}
				fields = record.Fields

				if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
					hasHeaderRow = true
					HeaderRow = record.All

					if printPass {
						checkError(writer.Write(record.All))
					}
					continue
				}
			}

			for i, f = range record.Fields {
				val, err = strconv.ParseFloat(removeComma(record.All[f-1]), 64)
				if err == nil {
					data[i] = append(data[i], transform(val))
				} else {
					data[i] = append(data[i], math.NaN())
				}
			}

			if printPass {
				checkError(writer.Write(record.All))
			}

		}

		readerReport(&config, csvReader, file)

		for col1, field1 := range fields {
			for col2, field2 := range fields {
				if col1 >= col2 {
					continue
				}

				d1, d2 := data[col1], data[col2]
				if printIgnore {
					d1, d2 = removeNaNs(d1, d2)
				}

				pearsonr := stat.Correlation(d1, d2, nil)

				if hasHeaderRow {
					fmt.Fprintf(os.Stderr, "%s\t%s\t%.4f\n", HeaderRow[field1-1], HeaderRow[field2-1], pearsonr)
				} else {
					fmt.Fprintf(os.Stderr, "%d\t%d\t%.4f\n", field1, field2, pearsonr)
				}

			}
		}
	},
}

// removeNaNs removes entries from a pair of slices if any of the two values is NaN.
func removeNaNs(d1, d2 []float64) ([]float64, []float64) {
	r1 := make([]float64, 0, len(d1))
	r2 := make([]float64, 0, len(d1))

	for i, x1 := range d1 {
		x2 := d2[i]
		if !math.IsNaN(x1) && !math.IsNaN(x2) {
			r1 = append(r1, x1)
			r2 = append(r2, x2)
		}
	}
	return r1, r2
}

func init() {
	RootCmd.AddCommand(corrCmd)

	corrCmd.Flags().StringP("fields", "f", "", "comma separated fields")
	corrCmd.Flags().BoolP("ignore_nan", "i", false, "Ignore non-numeric fields to avoid returning NaN")
	corrCmd.Flags().BoolP("log", "L", false, "Calcute correlations on Log10 transformed data")
	corrCmd.Flags().BoolP("pass", "x", false, "passthrough mode (forward input to output)")
}
