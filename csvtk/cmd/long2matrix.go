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
	"fmt"
	"math"
	"runtime"
	"strconv"
	"strings"

	"github.com/shenwei356/natsort"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// long2matrix represents the long2matrix command
var long2matrix = &cobra.Command{
	GroupID: "transform",

	Use:   "long2matrix",
	Short: "convert the long format to a matrix",
	Long: `convert the long format to matrix

Input: a three-column table. E.g.,

    col1   col2   value    
    A      A      1868883.0
    A      B      1310746.0
    B      A      1306936.0
    B      B      2117177.0

Output: a matrix with the same column and row names. E.g.,

        A           B        
    A   1868883.0   1310746.0
    B   1306936.0   2117177.0

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)

		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		fieldStrs := getFlagStringSlice(cmd, "fields")
		if len(fieldStrs) == 0 {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}
		fieldStr := strings.Join(fieldStrs, ",")

		minValue := getFlagFloat64(cmd, "min-value")
		maxValue := getFlagFloat64(cmd, "max-value")
		filterByValue := cmd.Flags().Lookup("min-value").Changed || cmd.Flags().Lookup("max-value").Changed
		keepNonNumeric := getFlagBool(cmd, "keep-non-numeric")
		clearBadValues := getFlagBool(cmd, "clear-bad-value")

		na := getFlagString(cmd, "na")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		outOpt := csvOutputOption{QuoteAll: config.QuoteAll}


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

		file := files[0]
		_, _, data, _, _, err := parseCSVfile(cmd, config,
			file, fieldStr, false, true, false)
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

		if len(data[0]) != 3 {
			checkError(fmt.Errorf("three columns required, but got %d", len(data[0])))
		}

		m := make(map[string]map[string]string, 1024)
		var _m map[string]string
		var a, b, v string
		var v1 float64
		var ok bool
		mKeys := make(map[string]struct{}, 128)
		for _, row := range data {
			a, b, v = row[0], row[1], row[2]

			if filterByValue {
				v1, err = strconv.ParseFloat(v, 64)
				if (err != nil && !keepNonNumeric) || (err == nil && (v1 < minValue || v1 > maxValue)) {
					if clearBadValues {
						v = na
					} else {
						continue
					}
				}
			}

			// store key values
			if _m, ok = m[a]; !ok {
				_m = make(map[string]string, 8)
				m[a] = _m
			} else {
				if _, ok = m[a][b]; ok {
					log.Warningf("skip duplicated records: %s", row)
					continue
				}
			}
			m[a][b] = v

			// store keys
			if _, ok = mKeys[a]; !ok {
				mKeys[a] = struct{}{}
			}
			if _, ok = mKeys[b]; !ok {
				mKeys[b] = struct{}{}
			}
		}

		if len(m) == 0 {
			return
		}

		keys := make([]string, 0, len(mKeys))
		for k := range mKeys {
			keys = append(keys, k)
		}
		natsort.Sort(keys)

		// header row
		row := make([]string, 1, len(keys)+1)
		row[0] = ""
		row = append(row, keys...)
		checkError(writer.Write(row))

		var k2 string
		var i int
		for _, k1 := range keys {
			row[0] = k1
			for i, k2 = range keys {
				if _, ok = m[k1][k2]; !ok {
					row[i+1] = na
				} else {
					row[i+1] = m[k1][k2]
				}
			}
			checkError(writer.Write(row))
		}
	},
}

func init() {
	RootCmd.AddCommand(long2matrix)

	long2matrix.Flags().StringSliceP("fields", "f", []string{"1-3"}, `the three fields/column to use. e.g., -f 1,2,3 or -f a,b,v`)

	long2matrix.Flags().Float64P("min-value", "m", -math.MaxFloat64, "only save records with values >= this value")
	long2matrix.Flags().Float64P("max-value", "M", math.MaxFloat64, "only save records with values <= this value")
	long2matrix.Flags().BoolP("keep-non-numeric", "N", false, "keep non-numeric values when filtering by --min-value or --max-value")
	long2matrix.Flags().BoolP("clear-bad-value", "B", false, "keep records failing to pass the filter but clear the value")

	long2matrix.Flags().StringP("na", "", "", "content for filling NA data")
}
