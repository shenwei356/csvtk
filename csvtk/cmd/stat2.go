// Copyright © 2016 Wei Shen <shenwei356@gmail.com>
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
	"regexp"
	"runtime"
	"sort"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/gonum/floats"
	"github.com/gonum/stat"
	"github.com/shenwei356/util/math"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"github.com/tatsushid/go-prettytable"
)

// stat2Cmd represents the stat2 command
var stat2Cmd = &cobra.Command{
	Use:     "stats2",
	Aliases: []string{"stats2"},
	Short:   "summary of selected digital fields",
	Long: `summary of selected digital fields: num, sum, min, max, mean, stdev

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--field) needed"))
		}

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		fields, colnames, negativeFields, needParseHeaderRow := parseFields(cmd, fieldStr, config.NoHeaderRow)
		var fieldsMap map[int]struct{}
		if len(fields) > 0 {
			fields2 := make([]int, len(fields))
			fieldsMap = make(map[int]struct{}, len(fields))
			for i, f := range fields {
				if negativeFields {
					fieldsMap[f*-1] = struct{}{}
					fields2[i] = f * -1
				} else {
					fieldsMap[f] = struct{}{}
					fields2[i] = f
				}
			}
			fields = fields2
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)
		checkError(err)
		csvReader.Run()

		parseHeaderRow := needParseHeaderRow // parsing header row
		var colnames2fileds map[string]int   // column name -> field
		var colnamesMap map[string]*regexp.Regexp
		var HeaderRow []string
		var isHeaderRow bool

		checkFields := true

		data := make(map[int][]float64)

		for chunk := range csvReader.Ch {
			checkError(chunk.Err)

			for _, record := range chunk.Data {
				if parseHeaderRow { // parsing header row
					colnames2fileds = make(map[string]int, len(record))
					for i, col := range record {
						colnames2fileds[col] = i + 1
					}
					colnamesMap = make(map[string]*regexp.Regexp, len(colnames))
					for _, col := range colnames {
						if !fuzzyFields {
							if negativeFields {
								if _, ok := colnames2fileds[col[1:]]; !ok {
									checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col[1:], file))
								}
							} else {
								if _, ok := colnames2fileds[col]; !ok {
									checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col, file))
								}
							}
						}
						if negativeFields {
							colnamesMap[col[1:]] = fuzzyField2Regexp(col[1:])
						} else {
							colnamesMap[col] = fuzzyField2Regexp(col)
						}
					}

					if len(fields) == 0 { // user gives the colnames
						fields = []int{}
						for _, col := range record {
							var ok bool
							if fuzzyFields {
								for _, re := range colnamesMap {
									if re.MatchString(col) {
										ok = true
										break
									}
								}
							} else {
								_, ok = colnamesMap[col]
							}
							if (negativeFields && !ok) || (!negativeFields && ok) {
								fields = append(fields, colnames2fileds[col])
							}
						}
					}

					fieldsMap = make(map[int]struct{}, len(fields))
					for _, f := range fields {
						fieldsMap[f] = struct{}{}
					}

					HeaderRow = record
					parseHeaderRow = false
					isHeaderRow = true
				}
				if checkFields {
					for field := range fieldsMap {
						if field > len(record) {
							checkError(fmt.Errorf(`field (%d) out of range (%d) in file: %s`, field, len(record), file))
						}
					}
					fields2 := []int{}
					for f := range record {
						_, ok := fieldsMap[f+1]
						if negativeFields {
							if !ok {
								fields2 = append(fields2, f+1)
							}
						} else {
							if ok {
								fields2 = append(fields2, f+1)
							}
						}
					}
					fields = fields2
					if len(fields) == 0 {
						checkError(fmt.Errorf("no fields matched in file: %s", file))
					}

					checkFields = false
				}

				if isHeaderRow {
					isHeaderRow = false
					continue
				}
				for _, f := range fields {
					if !reDigitals.MatchString(record[f-1]) {
						checkError(fmt.Errorf("column %d has non-number data: %s", f, record[f-1]))
					}
					v, e := strconv.ParseFloat(removeComma(record[f-1]), 64)
					checkError(e)
					if _, ok := data[f]; !ok {
						data[f] = []float64{}
					}
					data[f] = append(data[f], v)
				}
			}
		}
		tbl, err := prettytable.NewTable([]prettytable.Column{
			{Header: "field"},
			{Header: "num", AlignRight: true},
			{Header: "sum", AlignRight: true},
			{Header: "min", AlignRight: true},
			{Header: "max", AlignRight: true},
			{Header: "mean", AlignRight: true},
			{Header: "stdev", AlignRight: true}}...)
		checkError(err)
		tbl.Separator = "   "

		fields = []int{}
		for f := range data {
			fields = append(fields, f)
		}
		sort.Ints(fields)

		var fieldS string
		for _, f := range fields {
			if needParseHeaderRow {
				fieldS = HeaderRow[f-1]
			} else {
				fieldS = fmt.Sprintf("%d", f)
			}
			mean, stdev := stat.MeanStdDev(data[f], nil)
			tbl.AddRow(
				fieldS,
				humanize.Comma(int64(len(data[f]))),
				humanize.Commaf(math.Round(floats.Sum(data[f]), 2)),
				humanize.Commaf(math.Round(floats.Min(data[f]), 2)),
				humanize.Commaf(math.Round(floats.Max(data[f]), 2)),
				humanize.Commaf(math.Round(mean, 2)),
				humanize.Commaf(math.Round(stdev, 2)))
		}
		outfh.Write(tbl.Bytes())
	},
}

func init() {
	RootCmd.AddCommand(stat2Cmd)
	stat2Cmd.Flags().StringP("fields", "f", "", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
	stat2Cmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
}
