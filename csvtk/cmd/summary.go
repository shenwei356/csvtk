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
	"math"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

var separater string

// summaryCmd represents the stat2 command
var summaryCmd = &cobra.Command{
	GroupID: "info",

	Use:   "summary",
	Short: "summary statistics of selected numeric or text fields (groupby group fields)",
	Long: `summary statistics of selected numeric or text fields (groupby group fields)

Attention:

  1. Do not mix use field (column) numbers and names.
  2. Field range is supported, e.g., "-f 2-5:sum".

Available operations:
 
  # numeric/statistical operations
  # provided by github.com/gonum/stat and github.com/gonum/floats
  countn (count numeric values), min, max, sum, argmin, argmax,
  mean, stdev, variance, median, q1, q2, q3,
  entropy (Shannon entropy), 
  prod (product of the elements)

  # textual/numeric operations
  count, first, last, rand, unique/uniq, collapse, countunique

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		ignore := getFlagBool(cmd, "ignore-non-numbers")
		decimalWidth := getFlagNonNegativeInt(cmd, "decimal-width")
		decimalFormat := fmt.Sprintf("%%.%df", decimalWidth)
		decimalFormatScientificE := fmt.Sprintf("%%.%dE", decimalWidth)
		decimalFormatScientifice := fmt.Sprintf("%%.%de", decimalWidth)
		groupsStr := getFlagString(cmd, "groups")
		separater = getFlagString(cmd, "separater")
		if separater == "" {
			checkError(fmt.Errorf("flag -s (--separater) needed"))
		}
		seed := getFlagInt64(cmd, "rand-seed")
		rand.Seed(seed)

		ops := getFlagStringSlice(cmd, "fields")
		if len(ops) == 0 {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}

		stats := make(map[string][]string)         //  colname -> [stats]
		statsList := make([][]string, 0, len(ops)) // [ [stats] ]
		statsI := make(map[int][]string)           //  field -> [stats]

		var fieldsStrsG []string
		var fieldsStrsGMap map[string]struct{}
		// var numFieldsG int
		if groupsStr != "" {
			fieldsStrsG = strings.Split(groupsStr, ",")
			fieldsStrsGMap = make(map[string]struct{}, len(fieldsStrsG))
			for _, k := range fieldsStrsG {
				fieldsStrsGMap[k] = struct{}{}
			}
			// numFieldsG = len(fieldsStrsG)
		}

		fieldsStrsD := []string{}
		var numFieldsD int
		for _, key := range ops {
			items := strings.Split(key, ":")
			if _, ok := fieldsStrsGMap[items[0]]; ok {
				checkError(fmt.Errorf(`duplicated field in group field and data field: %s`, items[0]))
			}
			if len(items) == 1 {
				fieldsStrsD = append(fieldsStrsD, items[0])
				if _, ok := stats[items[0]]; !ok {
					stats[items[0]] = make([]string, 0, 1)
				}
				stats[items[0]] = append(stats[items[0]], "count")
				statsList = append(statsList, []string{items[0], "count"})
			} else if len(items) == 2 {
				if items[0] == "" {
					checkError(fmt.Errorf(`invalid field: %s`, key))
				}

				_fields := make([]string, 0, 8)

				if reIntegerRange2.MatchString(items[0]) {
					tmp := strings.Split(items[0], "-")
					start, _ := strconv.Atoi(tmp[0])
					end, _ := strconv.Atoi(tmp[1])
					if start > end {
						checkError(fmt.Errorf(`invalid field range: %s, start should be <= end`, items[0]))
					}

					for i := start; i <= end; i++ {
						_fields = append(_fields, strconv.Itoa(i))
					}
				} else if reIntegerRangeOnlyStart.MatchString(items[0]) {
					checkError(fmt.Errorf(`invalid field range: %s, end field is needed, e.g.i 1-2`, items[0]))
				} else {
					_fields = append(_fields, items[0])
				}

				_, ok1 := allStats[items[1]]  // for numbers
				_, ok2 := allStats2[items[1]] // for strings
				if !(ok1 || ok2) {
					checkError(fmt.Errorf(`invalid operation: %s. run "csvtk summary --help" for help`, items[1]))
				}

				for _, f := range _fields {
					fieldsStrsD = append(fieldsStrsD, f)

					if _, ok := stats[f]; !ok {
						stats[f] = make([]string, 0, 1)
					}
					stats[f] = append(stats[f], items[1])
					statsList = append(statsList, []string{f, items[1]})
				}
			} else {
				checkError(fmt.Errorf(`invalid value of flag --fields: %s`, key))
			}
			numFieldsD = len(fieldsStrsD)
		}

		fieldsStrsDMap := make(map[string]struct{}, len(fieldsStrsD))
		for _, k := range fieldsStrsD {
			fieldsStrsDMap[k] = struct{}{}
		}

		var tmp []string
		if len(fieldsStrsG) > 0 {
			tmp = append(fieldsStrsD, fieldsStrsG...)
		} else {
			tmp = fieldsStrsD
		}

		fieldsStr := strings.Join(tmp, ",")

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
		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				if config.Verbose {
					log.Warningf("csvtk summary: skipping empty input file: %s", file)
				}

				writer.Flush()
				checkError(writer.Error())
				readerReport(&config, csvReader, file)
				return
			}
			checkError(err)
		}

		csvReader.Read(ReadOption{
			FieldStr: fieldsStr,

			DoNotAllowDuplicatedColumnName: true,
		})

		var HeaderRow []string

		// group -> field -> data
		data := make(map[string]map[int][]float64) // for numbers
		data2 := make(map[string]map[int][]string) // for strings
		scientifc := make(map[string]map[int]byte) // for numbers

		fieldsG := []int{}
		fieldsD := []int{}
		fieldsDUniq := []int{}
		var f int
		var v float64
		var e error
		var ok bool
		var group string
		var needParseDigits bool

		var hasHeaderLine bool
		checkFirstLine := true

		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if checkFirstLine {
				checkFirstLine = false

				fieldsD = append(fieldsD, record.Fields[:numFieldsD]...) //  copy
				fieldsG = append(fieldsG, record.Fields[numFieldsD:]...) //  copy

				fieldsDUniq = make([]int, len(fieldsD))
				copy(fieldsDUniq, fieldsD)
				fieldsDUniq = UniqInts(fieldsDUniq)

				for i, f := range fieldsD {
					if _, ok = statsI[f]; !ok {
						statsI[f] = []string{statsList[i][1]}
					} else {
						statsI[f] = append(statsI[f], statsList[i][1])
					}
				}

				if !config.NoHeaderRow || record.IsHeaderRow {
					HeaderRow = record.All
					hasHeaderLine = true
					continue
				}
			}

			group = strings.Join(record.Selected[numFieldsD:], "_shenwei356_")
			if _, ok = data[group]; !ok {
				data[group] = make(map[int][]float64, 1024)
				scientifc[group] = make(map[int]byte)
			}
			if _, ok = data2[group]; !ok {
				data2[group] = make(map[int][]string, 1024)
			}

			for _, f = range fieldsDUniq {
				if _, ok = data2[group][f]; !ok {
					data2[group][f] = []string{}
				}
				data2[group][f] = append(data2[group][f], record.All[f-1])

				needParseDigits = false
				for _, op := range statsI[f] {
					if _, ok = allStats[op]; ok {
						needParseDigits = true
						break
					}
				}

				if !needParseDigits {
					continue
				}
				if !reDigitals.MatchString(record.All[f-1]) {
					if ignore {
						continue
					}
					checkError(fmt.Errorf("column %d has non-numeric data: %s, you can use flag -i/--ignore-non-numbers to skip these data", f, record.All[f-1]))
				}
				if strings.Contains(record.All[f-1], "E") {
					scientifc[group][f] = 'E'
				} else if strings.Contains(record.All[f-1], "e") {
					scientifc[group][f] = 'e'
				}

				v, e = strconv.ParseFloat(removeComma(record.All[f-1]), 64)
				checkError(e)
				if _, ok = data[group][f]; !ok {
					data[group][f] = []float64{}
				}
				data[group][f] = append(data[group][f], v)
			}

		}

		readerReport(&config, csvReader, file)

		colsOut := len(fieldsG) + len(fieldsD)
		if hasHeaderLine {
			record := make([]string, 0, colsOut)
			if len(fieldsG) > 0 {
				for _, i := range fieldsG {
					record = append(record, HeaderRow[i-1])
				}
			}

			for i, ss := range statsList {
				record = append(record, HeaderRow[fieldsD[i]-1]+":"+ss[1])
			}

			writer.Write(record)
		}

		groups := make([]string, 0, len(data)+len(data2))
		for group := range data {
			groups = append(groups, group)
		}
		sort.Strings(groups)

		var fu func([]float64) float64
		var fu2 func([]string) string
		for _, group := range groups {
			record := make([]string, 0, colsOut)
			if len(fieldsG) > 0 {
				record = append(record, strings.Split(group, "_shenwei356_")...)
			}

			for i, ss := range statsList {
				s := ss[1]
				f := fieldsD[i]

				sorted := false
				if _, ok = allStats[s]; !ok {
					fu2 = allStats2[s]
					record = append(record, fu2(data2[group][f]))
				} else {
					needSort := false
					for _, s := range statsI[f] {
						if s == "q1" || s == "q2" || s == "q3" || s == "median" {
							needSort = true
							break
						}
					}
					if needSort && !sorted {
						sort.Float64s(data[group][f])
						sorted = true
					}

					fu = allStats[s]
					if s == "countn" {
						record = append(record, fmt.Sprintf("%.0f", fu(data[group][f])))
					} else if scientifc[group][f] == 'E' {
						record = append(record, fmt.Sprintf(decimalFormatScientificE, fu(data[group][f])))
					} else if scientifc[group][f] == 'e' {
						record = append(record, fmt.Sprintf(decimalFormatScientifice, fu(data[group][f])))
					} else {
						record = append(record, fmt.Sprintf(decimalFormat, fu(data[group][f])))
					}
				}
			}
			writer.Write(record)
		}

	},
}

var allStats map[string]func([]float64) float64
var allStats2 map[string]func([]string) string
var allStatsList []string

func init() {
	allStats = make(map[string]func([]float64) float64)
	allStats["sum"] = func(s []float64) float64 {
		if len(s) == 0 {
			return math.NaN()
		}
		return floats.Sum(s)
	}
	allStats["max"] = func(s []float64) float64 {
		if len(s) == 0 {
			return math.NaN()
		}
		return floats.Max(s)
	}
	allStats["min"] = func(s []float64) float64 {
		if len(s) == 0 {
			return math.NaN()
		}
		return floats.Min(s)
	}
	allStats["argmax"] = func(s []float64) float64 {
		if len(s) == 0 {
			return math.NaN()
		}
		return float64(floats.MaxIdx(s) + 1)
	}
	allStats["argmin"] = func(s []float64) float64 {
		if len(s) == 0 {
			return math.NaN()
		}
		return float64(floats.MinIdx(s) + 1)
	}
	allStats["prod"] = floats.Prod
	allStats["countn"] = func(s []float64) float64 { return float64(len(s)) }
	allStats["mean"] = func(s []float64) float64 { return stat.Mean(s, nil) }
	allStats["stdev"] = func(s []float64) float64 { return stat.StdDev(s, nil) }
	allStats["entropy"] = func(s []float64) float64 { return stat.Entropy(s) }
	allStats["variance"] = func(s []float64) float64 { return stat.Variance(s, nil) }
	allStats["median"] = func(s []float64) float64 {
		if len(s) == 0 {
			return math.NaN()
		}
		return median(s)
	}
	allStats["q1"] = func(s []float64) float64 {
		if len(s) == 0 {
			return math.NaN()
		}
		return percentileValue(s, 0.25)
	}
	allStats["q2"] = func(s []float64) float64 {
		if len(s) == 0 {
			return math.NaN()
		}
		return median(s)
	}
	allStats["q3"] = func(s []float64) float64 {
		if len(s) == 0 {
			return math.NaN()
		}
		return percentileValue(s, 0.75)
	}

	allStats2 = make(map[string]func([]string) string)
	allStats2["count"] = func(s []string) string { return fmt.Sprintf("%d", len(s)) }
	allStats2["first"] = func(s []string) string { return s[0] }
	allStats2["last"] = func(s []string) string { return s[len(s)-1] }
	allStats2["rand"] = func(s []string) string { return s[rand.Intn(len(s))] }
	allStats2["uniq"] = func(s []string) string {
		m := make(map[string]struct{}, len(s))
		for _, v := range s {
			m[v] = struct{}{}
		}
		vs := make([]string, len(m))
		i := 0
		for v := range m {
			vs[i] = v
			i++
		}
		return strings.Join(vs, separater)
	}
	allStats2["unique"] = allStats2["uniq"]
	allStats2["countunique"] = func(s []string) string {
		m := make(map[string]struct{}, len(s))
		for _, v := range s {
			m[v] = struct{}{}
		}
		return fmt.Sprintf("%d", len(m))
	}
	allStats2["countuniq"] = allStats2["countunique"]
	allStats2["collapse"] = func(s []string) string { return strings.Join(s, separater) }

	// ---------------

	allStatsList = make([]string, 0, len(allStats)+len(allStats2))
	for k := range allStats {
		allStatsList = append(allStatsList, k)
	}
	for k := range allStats2 {
		allStatsList = append(allStatsList, k)
	}
	sort.Strings(allStatsList)

	RootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().StringP("groups", "g", "", `group via fields. e.g -g 1,2 or -g columnA,columnB`)
	summaryCmd.Flags().StringSliceP("fields", "f", []string{}, fmt.Sprintf(`operations on these fields. e.g "-f 1:count,1:sum", "-f 2-5:sum", or "-f colA:mean". available operations: %s`, strings.Join(allStatsList, ", ")))
	summaryCmd.Flags().BoolP("ignore-non-numbers", "i", false, `ignore non-numeric values like "NA" or "N/A"`)
	summaryCmd.Flags().IntP("decimal-width", "w", 2, "limit floats to N decimal points")
	summaryCmd.Flags().StringP("separater", "s", "; ", "separater for collapsed data")
	summaryCmd.Flags().Int64P("rand-seed", "S", 11, `rand seed for operation "rand"`)
}

func median(sorted []float64) float64 {
	l := len(sorted)
	if l == 0 {
		return 0
	}
	if l%2 == 0 {
		return (sorted[l/2-1] + sorted[l/2]) / 2
	}
	return sorted[l/2]
}

/*
This implementation follows R's summary () and quantile (type=7) functions.

	See discussion here:
	http://tolstoy.newcastle.edu.au/R/e17/help/att-1067/Quartiles_in_R.pdf
*/
func percentileValue(sorted []float64, percentile float64) float64 {
	l := len(sorted)
	if l == 0 || percentile < 0 || percentile > 1 {
		return 0
	}
	if l == 1 {
		return sorted[0]
	}

	h := float64(l-1) * percentile
	fh := math.Floor(h)
	return sorted[int(fh)] + (h-fh)*(sorted[int(fh)+1]-sorted[int(fh)])
}
