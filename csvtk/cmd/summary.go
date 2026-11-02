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
	"maps"
	"math"
	"math/rand"
	"runtime"
	"slices"
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
	Short: "summary statistics of selected numeric or text fields (grouped by group fields)",
	Long: `summary statistics of selected numeric or text fields (grouped by group fields)

Attention:

  1. Group and data selectors are resolved independently. They must not overlap.
  2. Field ranges are supported, e.g., "-f 2-5:sum".
  3. The default operation is count, e.g., "-f 1-2" is equivalent to "-f 1-2:count".
  4. Values are stored only for the numeric or text operations requested for
     each field. Memory use still grows with the retained values and groups.
  5. The same input and --rand-seed give reproducible results for the rand operation.

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
		// Keep seeded selection local; global rand.Seed is a no-op in recent Go versions.
		random := rand.New(rand.NewSource(seed))
		textStats := maps.Clone(allStats2)
		textStats["rand"] = func(values []string) string { return values[random.Intn(len(values))] }

		ops := getFlagStringSlice(cmd, "fields")
		if len(ops) == 0 {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}
		names := getFlagStringSlice(cmd, "names")

		statsList := make([][]string, 0, len(ops)) // [ [stats] ]
		statsI := make(map[int][]string)           //  field -> [stats]

		var groupSelection fieldSelection
		if groupsStr != "" {
			groupSelection = newFieldSelection(groupsStr)
		}

		fieldsStrsD := []string{}
		for _, key := range ops {
			items := strings.Split(key, ":")
			if len(items) == 1 {
				// Expand default-count ranges just like fields with an explicit operation.
				items = append(items, "count")
			}
			if len(items) == 2 {
				if items[0] == "" {
					checkError(fmt.Errorf(`invalid field: %s`, key))
				}

				_fields := make([]string, 0, 8)

				if reIntegerRange2.MatchString(items[0]) {
					tmp := strings.Split(items[0], "-")
					start, err := strconv.Atoi(tmp[0])
					checkError(err)
					end, err := strconv.Atoi(tmp[1])
					checkError(err)
					if start > end {
						checkError(fmt.Errorf(`invalid field range: %s, start should be <= end`, items[0]))
					}

					for i := start; ; i++ {
						_fields = append(_fields, strconv.Itoa(i))
						// Avoid overflow when the final field is the largest integer.
						if i == end {
							break
						}
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

					statsList = append(statsList, []string{f, items[1]})
				}
			} else {
				checkError(fmt.Errorf(`invalid value of flag --fields: %s`, key))
			}
		}
		// Each operation retains its own selector, including exclusions.
		dataSelections := make([]fieldSelection, len(fieldsStrsD))
		needsHeader := len(groupSelection.names) > 0
		for i, spec := range fieldsStrsD {
			dataSelections[i] = newFieldSelection(spec)
			needsHeader = needsHeader || len(dataSelections[i].names) > 0
		}

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

		if needsHeader {
			csvReader.NoHeaderRow = false
		}
		csvReader.Read(ReadOption{FieldStr: "1-"})

		var HeaderRow []string

		// group -> field -> data
		data := make(map[string]map[int][]float64) // for numbers
		data2 := make(map[string]map[int][]string) // for strings
		scientifc := make(map[string]map[int]byte) // for numbers

		fieldsG := []int{}
		var groupValues []string // Reused values for the resolved grouping columns.
		fieldsD := []int{}
		fieldsDUniq := []int{}
		var numericFields, textFields []int // Store only values needed by the requested operations.
		var f int
		var v float64
		var e error
		var ok bool
		var group string

		var hasHeaderLine bool
		checkFirstLine := true

		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if checkFirstLine {
				checkFirstLine = false

				if groupsStr != "" {
					fieldsG = groupSelection.resolve(record.All, false)
				}
				groupValues = make([]string, len(fieldsG))
				groupFields := make(map[int]struct{}, len(fieldsG))
				for _, f := range fieldsG {
					groupFields[f] = struct{}{}
				}
				var expandedStats [][]string
				for i, selection := range dataSelections {
					for _, f := range selection.resolve(record.All, false) {
						if _, overlap := groupFields[f]; overlap {
							checkError(fmt.Errorf("duplicated field in group field and data field: %d", f))
						}
						fieldsD = append(fieldsD, f)
						expandedStats = append(expandedStats, statsList[i])
					}
				}
				statsList = expandedStats
				if len(fieldsD) == 0 {
					checkError(fmt.Errorf("no data fields matched"))
				}
				checkError(validateSummaryNames(names, len(statsList)))

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
				for _, f := range fieldsDUniq {
					var numeric, text bool
					for _, op := range statsI[f] {
						if _, ok := allStats[op]; ok {
							numeric = true
						} else {
							text = true
						}
					}
					if numeric {
						numericFields = append(numericFields, f)
					}
					if text {
						textFields = append(textFields, f)
					}
				}

				if !config.NoHeaderRow || record.IsHeaderRow {
					HeaderRow = record.All
					hasHeaderLine = true
					continue
				}
			}

			for i, f := range fieldsG {
				groupValues[i] = record.All[f-1]
			}
			group = encodeFields(groupValues, false)
			if _, ok = data[group]; !ok {
				// Keep every group, including groups with no valid numeric values.
				data[group] = nil
				if len(numericFields) > 0 {
					data[group] = make(map[int][]float64, len(numericFields))
					scientifc[group] = make(map[int]byte, len(numericFields))
				}
				if len(textFields) > 0 {
					data2[group] = make(map[int][]string, len(textFields))
				}
			}

			for _, f = range textFields {
				data2[group][f] = append(data2[group][f], record.All[f-1])
			}
			for _, f = range numericFields {
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
		if hasHeaderLine && !config.NoOutHeader {
			record := make([]string, 0, colsOut)
			if len(fieldsG) > 0 {
				for _, i := range fieldsG {
					record = append(record, HeaderRow[i-1])
				}
			}

			for i, ss := range statsList {
				if len(names) > 0 {
					record = append(record, names[i])
				} else {
					record = append(record, HeaderRow[fieldsD[i]-1]+":"+ss[1])
				}
			}

			checkError(writer.Write(record))
		}

		groups := make([]string, 0, len(data)+len(data2))
		groupFields := make(map[string][]string, len(data))
		for group := range data {
			groups = append(groups, group)
			groupFields[group] = decodeFields(group)
		}
		sort.Slice(groups, func(i, j int) bool {
			return slices.Compare(groupFields[groups[i]], groupFields[groups[j]]) < 0
		})

		var fu func([]float64) float64
		var fu2 func([]string) string
		needOriginalOrder := make(map[int]bool)
		for f, operations := range statsI {
			for _, operation := range operations {
				if operation == "argmin" || operation == "argmax" {
					needOriginalOrder[f] = true
				}
			}
		}
		for _, group := range groups {
			sortedFields := make(map[int][]float64)
			record := make([]string, 0, colsOut)
			if len(fieldsG) > 0 {
				record = append(record, groupFields[group]...)
			}

			for i, ss := range statsList {
				s := ss[1]
				f := fieldsD[i]

				if _, ok = allStats[s]; !ok {
					fu2 = textStats[s]
					record = append(record, fu2(data2[group][f]))
				} else {
					values := data[group][f]
					if s == "q1" || s == "q2" || s == "q3" || s == "median" {
						if _, found := sortedFields[f]; !found {
							if needOriginalOrder[f] {
								sortedFields[f] = slices.Clone(values)
							} else {
								sortedFields[f] = values
							}
							sort.Float64s(sortedFields[f])
						}
						values = sortedFields[f]
					}
					fu = allStats[s]
					if s == "countn" {
						record = append(record, fmt.Sprintf("%.0f", fu(values)))
					} else if scientifc[group][f] == 'E' {
						record = append(record, fmt.Sprintf(decimalFormatScientificE, fu(values)))
					} else if scientifc[group][f] == 'e' {
						record = append(record, fmt.Sprintf(decimalFormatScientifice, fu(values)))
					} else {
						record = append(record, fmt.Sprintf(decimalFormat, fu(values)))
					}
				}
			}
			checkError(writer.Write(record))
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
	summaryCmd.Flags().StringP("groups", "g", "", `group via fields. e.g., -g 1,2 or -g columnA,columnB`)
	summaryCmd.Flags().StringSliceP("fields", "f", []string{}, fmt.Sprintf(`operations on these fields. e.g., "-f 1:count,1:sum", "-f 2-5:sum", or "-f colA:mean". available operations: %s`, strings.Join(allStatsList, ", ")))
	summaryCmd.Flags().StringSliceP("names", "n", []string{}, "new names for the summary columns")
	summaryCmd.Flags().BoolP("ignore-non-numbers", "i", false, `ignore non-numeric values like "NA" or "N/A"`)
	summaryCmd.Flags().IntP("decimal-width", "w", 2, "limit floats to N decimal places")
	summaryCmd.Flags().StringP("separater", "s", "; ", "separator for collapsed data")
	summaryCmd.Flags().Int64P("rand-seed", "S", 11, `random seed for operation "rand"`)
}

func validateSummaryNames(names []string, n int) error {
	if len(names) > 0 && len(names) != n {
		return fmt.Errorf("number of names (%d) should be equal to number of summary columns (%d)", len(names), n)
	}
	return nil
}

// median uses the same overflow-safe interpolation as the other quantiles.
func median(sorted []float64) float64 {
	return percentileValue(sorted, 0.5)
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
	left := sorted[int(fh)]
	weight := h - fh
	if weight == 0 {
		return left // Includes the last endpoint, which has no right neighbor.
	}
	right := sorted[int(fh)+1]
	if weight == 0.5 {
		// Add before halving to retain subnormal midpoint rounding. Only
		// halve first when the sum overflows, where underflow is impossible.
		if sum := left + right; !math.IsInf(sum, 0) {
			return sum / 2
		}
		return left/2 + right/2
	}
	if math.Signbit(left) != math.Signbit(right) {
		// Subtracting opposite-sign finite values can overflow.
		return (1-weight)*left + weight*right
	}
	// This form also preserves equal subnormal values without underflow.
	return left + weight*(right-left)
}
