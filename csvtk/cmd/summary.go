// Copyright © 2016-2019 Wei Shen <shenwei356@gmail.com>
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
	"regexp"
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
	Use:   "summary",
	Short: "summary statistics of selected digital fields (groupby group fields)",
	Long: `summary statistics of selected digital fields (groupby group fields)

Attention:

  1. Do not mix use digital fields and column names.

Available operations:
 
  # numeric/statistical operations
  # provided by github.com/gonum/stat and github.com/gonum/floats
  countn (count of digits), min, max, sum,
  mean, stdev, variance, median, q1, q2, q3,
  entropy (Shannon entropy), 
  prod (product of the elements)

  # textual/numeric operations
  count, first, last, rand, unique, collapse, countunique

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		ignore := getFlagBool(cmd, "ignore-non-digits")
		decimalWidth := getFlagNonNegativeInt(cmd, "decimal-width")
		decimalFormat := fmt.Sprintf("%%.%df", decimalWidth)
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

		stats := make(map[string][]string)
		statsI := make(map[int][]string)

		var fieldsStrsG []string
		var fieldsStrsGMap map[string]struct{}
		if groupsStr != "" {
			fieldsStrsG = strings.Split(groupsStr, ",")
			fieldsStrsGMap = make(map[string]struct{}, len(fieldsStrsG))
			for _, k := range fieldsStrsG {
				fieldsStrsGMap[k] = struct{}{}
			}
		}

		fieldsStrsD := []string{}
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
			} else if len(items) == 2 {
				if items[0] == "" {
					checkError(fmt.Errorf(`invalid field: %s`, key))
				}
				fieldsStrsD = append(fieldsStrsD, items[0])

				_, ok1 := allStats[items[1]]
				_, ok2 := allStats2[items[1]]
				if !(ok1 || ok2) {
					checkError(fmt.Errorf(`invalid operation: %s. run "csvtk summary --help" for help`, items[1]))
				}
				if _, ok := stats[items[0]]; !ok {
					stats[items[0]] = make([]string, 0, 1)
				}
				stats[items[0]] = append(stats[items[0]], items[1])
			} else {
				checkError(fmt.Errorf(`invalid value of flag --fields: %s`, key))
			}
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

		fuzzyFields := false
		fields, colnames, negativeFields, needParseHeaderRow := parseFields(cmd, fieldsStr, config.NoHeaderRow)
		if negativeFields {
			checkError(fmt.Errorf(`negative field not supported by this command`))
		}
		var fieldsMap map[int]struct{}
		var fieldsMapG map[int]struct{}
		var fieldsMapD map[int]struct{}
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

			fieldsMapG = make(map[int]struct{}, len(fields))
			if groupsStr != "" {
				for _, k := range fieldsStrsG {
					i, e := strconv.Atoi(k)
					if e != nil {
						checkError(fmt.Errorf("fail to convert group field to integer: %s", k))
					}
					fieldsMapG[i] = struct{}{}
				}
			}
			fieldsMapD = make(map[int]struct{}, len(fields))
			for _, k := range fieldsStrsD {
				i, e := strconv.Atoi(k)
				if e != nil {
					checkError(fmt.Errorf("fail to convert data field to integer: %s", k))
				}
				fieldsMapD[i] = struct{}{}
				statsI[i] = stats[k]
			}
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}

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

		// group -> field -> data
		data := make(map[string]map[int][]float64)
		data2 := make(map[string]map[int][]string)

		fieldsG := []int{}
		fieldsD := []int{}
		var i, f int
		var v float64
		var e error
		var ok bool
		var items []string
		var group string
		var needParseDigits bool
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
								if _, ok = fieldsStrsDMap[col]; ok {
									fieldsD = append(fieldsD, colnames2fileds[col])
									statsI[colnames2fileds[col]] = stats[col]
								}
								if _, ok = fieldsStrsGMap[col]; ok {
									fieldsG = append(fieldsG, colnames2fileds[col])
								}
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
						if _, ok = fieldsMapG[f+1]; ok {
							fieldsG = append(fieldsG, f+1)
						}
						if _, ok = fieldsMapD[f+1]; ok {
							fieldsD = append(fieldsD, f+1)
							if needParseHeaderRow {
								stats[HeaderRow[f]] = statsI[f+1]
							}
						}
					}
					fields = fields2
					if len(fields) == 0 {
						checkError(fmt.Errorf("no fields matched in file: %s", file))
					}
					if len(fieldsMapG) > 0 && len(fieldsG) == 0 {
						checkError(fmt.Errorf("no group fields matched in file: %s", file))
					}
					if len(fieldsD) == 0 {
						checkError(fmt.Errorf("no data fields matched in file: %s", file))
					}

					items = make([]string, len(fieldsG))
					checkFields = false
				}

				if isHeaderRow {
					isHeaderRow = false
					continue
				}

				// fmt.Println(fields, fieldsG, fieldsD)

				for i, f = range fieldsG {
					items[i] = record[f-1]
				}
				group = strings.Join(items, "_shenwei356_")
				if _, ok = data[group]; !ok {
					data[group] = make(map[int][]float64, 1000)
				}
				if _, ok = data2[group]; !ok {
					data2[group] = make(map[int][]string, 1000)
				}

				for _, f = range fieldsD {
					if _, ok = data2[group][f]; !ok {
						data2[group][f] = []string{}
					}
					data2[group][f] = append(data2[group][f], record[f-1])

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
					if !reDigitals.MatchString(record[f-1]) {
						if ignore {
							continue
						}
						checkError(fmt.Errorf("column %d has non-digital data: %s, you can use flag -i/--ignore-non-digits to skip these data", f, record[f-1]))
					}
					v, e = strconv.ParseFloat(removeComma(record[f-1]), 64)
					checkError(e)
					if _, ok = data[group][f]; !ok {
						data[group][f] = []float64{}
					}
					data[group][f] = append(data[group][f], v)
				}
			}
		}

		readerReport(&config, csvReader, file)

		colsOut := len(fieldsG) + len(fieldsD)
		if needParseHeaderRow {
			record := make([]string, 0, colsOut)
			if len(fieldsG) > 0 {
				for _, i := range fieldsG {
					record = append(record, HeaderRow[i-1])
				}
			}

			for _, f := range fieldsD {
				for _, s := range statsI[f] {
					record = append(record, HeaderRow[f-1]+":"+s)
				}
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

			for _, f := range fieldsD {
				sorted := false

				for _, s := range statsI[f] {
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
						} else {
							record = append(record, fmt.Sprintf(decimalFormat, fu(data[group][f])))
						}
					}
				}
			}
			writer.Write(record)
		}

		writer.Flush()
		checkError(writer.Error())
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
	allStats2["countunique"] = func(s []string) string {
		m := make(map[string]struct{}, len(s))
		for _, v := range s {
			m[v] = struct{}{}
		}
		return fmt.Sprintf("%d", len(m))
	}
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
	summaryCmd.Flags().StringP("groups", "g", "", `group via fields. e.g -f 1,2 or -f columnA,columnB`)
	summaryCmd.Flags().StringSliceP("fields", "f", []string{}, fmt.Sprintf(`operations on these fields. e.g -f 1:count,1:sum or -f colA:mean. available operations: %s`, strings.Join(allStatsList, ", ")))
	summaryCmd.Flags().BoolP("ignore-non-digits", "i", false, `ignore non-digital values like "NA" or "N/A"`)
	summaryCmd.Flags().IntP("decimal-width", "n", 2, "limit floats to N decimal points")
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

/* This implementation follows R's summary () and quantile (type=7) functions.
   See discussion here:
   http://tolstoy.newcastle.edu.au/R/e17/help/att-1067/Quartiles_in_R.pdf */
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
