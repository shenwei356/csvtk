// Copyright © 2016-2026 Wei Shen <shenwei356@gmail.com>
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
	"runtime"
	"strconv"
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// matrix2long represents the matrix2long command
var matrix2long = &cobra.Command{
	GroupID: "transform",

	Use:   "matrix2long",
	Short: "convert a matrix to the long format",
	Long: `convert a matrix to the long format

Input: a matrix with the same column and row names. E.g.,

        A           B        
    A   1868883.0   1310746.0
    B   1306936.0   2117177.0

Output: a three-column table. E.g.,

    col1   col2   value    
    A      A      1868883.0
    A      B      1310746.0
    B      A      1306936.0
    B      B      2117177.0

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)

		if config.NoHeaderRow {
			checkError(fmt.Errorf("matrix2long requires a header row"))
		}

		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		colnames := getFlagStringSlice(cmd, "colnames")
		if !config.NoOutHeader && len(colnames) != 3 {
			checkError(fmt.Errorf("three column names are needed for the output"))
		}
		skipSameKeys := getFlagBool(cmd, "skip-same-keys")
		keepSameKeys := getFlagBool(cmd, "keep-same-keys")
		if skipSameKeys && keepSameKeys {
			checkError(fmt.Errorf("flags -s (--skip-same-keys) and -S (--keep-same-keys) cannot be used together"))
		}
		minValue := getFlagFloat64(cmd, "min-value")
		maxValue := getFlagFloat64(cmd, "max-value")
		keepNonNumeric := getFlagBool(cmd, "keep-non-numeric")
		filterByValue := cmd.Flags().Lookup("min-value").Changed || cmd.Flags().Lookup("max-value").Changed

		blanks := getFlagStringSlice(cmd, "blanks")
		skipBlanks := getFlagBool(cmd, "skip-blanks")
		mBlanks := make(map[string]struct{}, 8)
		for _, v := range blanks {
			mBlanks[strings.ToLower(v)] = struct{}{}
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

		if len(data) == 0 {
			log.Warningf("no data to sort from file: %s", file)
			return
		}

		if len(headerRow) < 2 {
			checkError(fmt.Errorf("invalid input, at least 2 columns needed"))
		}

		if len(data)+1 != len(headerRow) {
			checkError(fmt.Errorf("invalid input, the numbers of columns and rows should match"))
		}

		headerRow = headerRow[1:]

		if !config.NoOutHeader {
			checkError(writer.Write(colnames))
		}

		row := make([]string, 3)
		var v string
		var j int
		var v1 float64
		var ok bool
		for i, line := range data {
			if line[0] != headerRow[i] {
				checkError(fmt.Errorf("invalid matrix format"))
			}
			row[0] = line[0]
			for j, v = range line[1:] {
				if skipSameKeys && line[0] == headerRow[j] {
					continue
				}
				if keepSameKeys && line[0] != headerRow[j] {
					continue
				}

				if skipBlanks {
					if _, ok = mBlanks[strings.ToLower(v)]; ok {
						continue
					}
				}

				row[1] = headerRow[j]
				row[2] = v

				if filterByValue {
					v1, err = strconv.ParseFloat(v, 64)
					if err != nil && !keepNonNumeric { // not a number
						continue
					}

					if err == nil && (v1 < minValue || v1 > maxValue) { // out of range
						continue
					}
				}

				checkError(writer.Write(row))
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(matrix2long)

	matrix2long.Flags().StringSliceP("colnames", "n", []string{"col1", "col2", "value"}, `column names of the output (3 values required). e.g -n a,b,v`)
	matrix2long.Flags().BoolP("skip-same-keys", "s", false, `skip records with the same key names`)
	matrix2long.Flags().BoolP("keep-same-keys", "S", false, `keep records with the same key names`)
	matrix2long.Flags().BoolP("skip-blanks", "b", false, `skip records with blank values (defined by --blanks)`)
	matrix2long.Flags().StringSliceP("blanks", "B", []string{"", "na", "n/a", "none", "null", "."}, `blank values, case ignored`)

	matrix2long.Flags().Float64P("min-value", "m", -math.MaxFloat64, "only show records with values >= this value")
	matrix2long.Flags().Float64P("max-value", "M", math.MaxFloat64, "only show records with values <= this value")
	matrix2long.Flags().BoolP("keep-non-numeric", "N", false, "keep non-numeric values when filtering by --min-value or --max-value")
}
