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
	"bufio"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/stat"
)

// corrCmd represents the corr command
var corrCmd = &cobra.Command{
	Use:   "corr",
	Short: "calculate Pearson correlation between two columns",
	Long:  "calculate Pearson correlation between two columns",

	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		printField := getFlagString(cmd, "fields")
		printIgnore := getFlagBool(cmd, "ignore_nan")
		_ = printIgnore
		printPass := getFlagBool(cmd, "pass")
		printLog := getFlagBool(cmd, "log")
		outFile := config.OutFile

		if config.Tabs {
			config.OutDelimiter = rune('\t')
		}

		outw := os.Stdout
		if outFile != "-" {
			tw, err := os.Create(outFile)
			checkError(err)
			outw = tw
		}
		outfh := bufio.NewWriter(outw)

		defer outfh.Flush()
		defer outw.Close()

		transform := func(x float64) float64 { return x }
		if printLog {
			transform = func(x float64) float64 {
				return math.Log10(x + 1)
			}
		}

		field2col := make(map[string]int)
		Data := make(map[int][]float64)

		targetCols := make(map[int]string)

		for x, tok := range strings.Split(printField, ",") {
			tok = strings.TrimSpace(tok)
			var col int
			if config.NoHeaderRow {
				if len(tok) == 0 {
					continue
				}
				pcol, err := strconv.Atoi(tok)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Illegal field number: %s\n", tok)
					os.Exit(1)
				}
				col = pcol - 1
				if col < 0 {
					fmt.Fprintf(os.Stderr, "Illegal field number: %d!\n", pcol)
					os.Exit(1)
				}
				targetCols[col] = tok
			}
			if len(tok) != 0 {
				targetCols[-(x + 1)] = tok
			}

		}

		for _, file := range files[:1] {
			csvReader, err := newCSVReaderByConfig(config, file)
			checkError(err)
			csvReader.Run()

			isHeaderLine := !config.NoHeaderRow
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				for _, record := range chunk.Data {
					if isHeaderLine {
						for i, column := range record {
							field2col[column] = i
							if printField == "" {
								targetCols[i] = column
							}
						}

						isHeaderLine = false
						if printPass {
							outfh.Write([]byte(strings.Join(record, string(config.OutDelimiter)) + "\n"))
						}
						continue
					} else {
						if len(targetCols) == 0 {
							for i, _ := range record {
								targetCols[i] = strconv.Itoa(i + 1)
							}
						}
					}

					for col, field := range targetCols {
						i := col
						if !config.NoHeaderRow && i < 0 {
							var ok bool
							i, ok = field2col[field]
							if !ok {
								fmt.Fprintf(os.Stderr, "Invalid field specified: %s\n", field)
								os.Exit(1)
							}
						}
						if printPass {
							outfh.Write([]byte(strings.Join(record, string(config.OutDelimiter)) + "\n"))
						}
						p, err := strconv.ParseFloat(record[i], 64)
						if err == nil {
							Data[i] = append(Data[i], transform(p))
						} else {
							Data[i] = append(Data[i], math.NaN())
						}
					} // field

				} // record
			} //chunk

		} //file

		// Calculate and print correlations:
		seen := make(map[int]map[int]bool)
		for col1, field1 := range targetCols {
			if col1 < 0 {
				col1 = field2col[field1]
			}
			if seen[col1] == nil {
				seen[col1] = make(map[int]bool)
			}
			for col2, field2 := range targetCols {
				if col2 < 0 {
					col2 = field2col[field2]
				}
				if col1 == col2 {
					continue
				}
				if seen[col1][col2] {
					continue
				}
				d1, d2 := Data[col1], Data[col2]
				if printIgnore {
					d1, d2 = removeNaNs(d1, d2)
				}

				pearsonr := stat.Correlation(d1, d2, nil)
				fmt.Fprintf(os.Stderr, "%s%s%s%s%.4f\n", field1, string(config.OutDelimiter), field2, string(config.OutDelimiter), pearsonr)

				seen[col1][col2] = true
				if seen[col2] == nil {
					seen[col2] = make(map[int]bool)
				}
				seen[col2][col1] = true
			} // col2
		} // col1
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
