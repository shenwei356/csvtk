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
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// csv2jsonCmd represents the uniq command
var csv2jsonCmd = &cobra.Command{
	GroupID: "format",

	Use:   "csv2json",
	Short: "convert CSV to JSON format",
	Long: `convert CSV to JSON format

Notes:

  1. Column names, keys, and string values are escaped according to JSON syntax.
  2. With -n/--parse-num, numeric text is normalized without rounding:
     00 becomes 0, .5 becomes 0.5, and 1. becomes 1.0.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}

		runtime.GOMAXPROCS(config.NumCPUs)

		blanks := getFlagBool(cmd, "blanks")

		_parseNumCols := getFlagStringSlice(cmd, "parse-num")
		var parseNumAll, parseNum0, parseNum bool
		parseNumCols := make(map[int]interface{})
		var err error
		var n int
		for _, c := range _parseNumCols {
			c = strings.ToLower(c)
			if c == "a" || c == "all" {
				parseNum = true
				parseNumAll = true
				break
			}
			if reIntegers.MatchString(c) {
				n, _ = strconv.Atoi(c)
				if n < 1 {
					checkError(fmt.Errorf("positive column index needed: %s", c))
				}
				parseNumCols[n] = struct{}{}
			} else {
				checkError(fmt.Errorf("positive column index needed: %s", c))
			}
		}
		parseNum0 = true

		indent := getFlagString(cmd, "indent")
		hasIndent := indent != ""
		var LF, SEP string
		if hasIndent {
			LF = "\n"
			SEP = " "
		}

		fieldStr := getFlagString(cmd, "key")
		if strings.Contains(fieldStr, ",") {
			checkError(fmt.Errorf("multiple values not allowed for -k/--key"))
		}
		keyed := fieldStr != ""
		skipKey := keyed && getFlagBool(cmd, "skip-key")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer closeOutput(outfh)

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				if config.Verbose {
					log.Warningf("csvtk csv2json: skipping empty input file: %s", file)
				}
				if keyed {
					outfh.WriteString("{")
				} else {
					outfh.WriteString("[")
				}
				if keyed {
					outfh.WriteString("}\n")
				} else {
					outfh.WriteString("]\n")
				}

				readerReport(&config, csvReader, file)
				return
			}
			checkError(err)
		}

		csvReader.Read(ReadOption{
			FieldStr: fieldStr,

			DoNotAllowDuplicatedColumnName: keyed,
		})

		var key string

		if keyed {
			outfh.WriteString("{")
		} else {
			outfh.WriteString("[")
		}
		outfh.WriteString(LF)

		keysMaps := make(map[string]struct{}, 1024)
		var i int
		var col string
		first := true
		var ok bool
		var HeaderRow []string // JSON-quoted column names, encoded once per input.

		checkFirstLine := true
		var hasHeaderLine bool
		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if checkFirstLine {
				checkFirstLine = false

				if !config.NoHeaderRow || record.IsHeaderRow {
					HeaderRow = make([]string, len(record.All))
					for i, col = range record.All {
						HeaderRow[i] = quoteJSONString(col)
					}
					hasHeaderLine = true

					continue
				}
			}

			if keyed {
				key = record.Selected[0]
				if _, ok = keysMaps[key]; ok {
					if config.Verbose {
						log.Warningf("ignore record with duplicated key (%s) at line %d", key, record.Line)
					}
					continue
				}
				keysMaps[key] = struct{}{}
			}

			if first {
				first = false
			} else {
				outfh.WriteString("," + LF)
			}

			// Omitting the final key column makes the preceding column the last
			// emitted field; separators must follow output positions.
			lastField := len(record.All) - 1
			if skipKey && record.Fields[0] == len(record.All) {
				lastField--
			}

			if hasHeaderLine {
				if keyed {
					outfh.WriteString(indent + quoteJSONString(key) + `:` + SEP + `{` + LF)
				} else {
					outfh.WriteString(indent + `{` + LF)
				}
				for i, col = range HeaderRow {
					if skipKey && i+1 == record.Fields[0] {
						continue
					}

					if parseNumAll {
						parseNum = true
					} else if parseNum0 {
						if _, ok = parseNumCols[i+1]; ok {
							parseNum = true
						}
					}

					if i < lastField {
						outfh.WriteString(indent + indent + col + `:` + SEP + processJSONValue(record.All[i], blanks, parseNum) + "," + LF)
					} else {
						outfh.WriteString(indent + indent + col + `:` + SEP + processJSONValue(record.All[i], blanks, parseNum) + LF)
					}

					parseNum = false
				}
				outfh.WriteString(indent + "}")
			} else {
				if keyed {
					outfh.WriteString(indent + quoteJSONString(key) + `:` + SEP + `[` + LF)
				} else {
					outfh.WriteString(indent + `[` + LF)
				}

				for i, col = range record.All {
					if skipKey && i+1 == record.Fields[0] {
						continue
					}

					if parseNumAll {
						parseNum = true
					} else if parseNum0 {
						if _, ok = parseNumCols[i+1]; ok {
							parseNum = true
						}
					}

					if i < lastField {
						outfh.WriteString(indent + indent + processJSONValue(col, blanks, parseNum) + "," + LF)
					} else {
						outfh.WriteString(indent + indent + processJSONValue(col, blanks, parseNum) + LF)
					}

					parseNum = false
				}
				outfh.WriteString(indent + "]")
			}
		}

		outfh.WriteString(LF)
		if keyed {
			outfh.WriteString("}\n")
		} else {
			outfh.WriteString("]\n")
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(csv2jsonCmd)
	csv2jsonCmd.Flags().StringP("indent", "i", "  ", `indent. if given blank, output json in one line.`)
	csv2jsonCmd.Flags().StringP("key", "k", "", "output json as an array of objects keyed by a given field rather than as a list. e.g -k 1 or -k columnA")
	csv2jsonCmd.Flags().BoolP("skip-key", "K", false, `do not put KEY in the data when using -k KEY`)
	csv2jsonCmd.Flags().BoolP("blanks", "b", false, `do not convert "", "na", "n/a", "none", "null", "." to null`)
	csv2jsonCmd.Flags().StringSliceP("parse-num", "n", []string{}, `parse numeric values for nth column, multiple values are supported and "a"/"all" for all columns`)
}

// quoteJSONString encodes keys and string values with the same JSON escaping.
func quoteJSONString(value string) string {
	encoded, err := json.Marshal(value)
	checkError(err)
	return string(encoded)
}

// processJSONValue preserves boolean/null handling and converts numeric fields
// only when requested for their column.
func processJSONValue(val string, blanks bool, parseNum bool) string {
	switch strings.ToLower(val) {
	case "true":
		return "true"
	case "false":
		return "false"
	case "", "na", "n/a", "none", "null", ".":
		if blanks {
			return quoteJSONString(val)
		}
		return "null"
	}
	if parseNum && reDigitals.MatchString(val) {
		return normalizeJSONNumber(val)
	}

	return quoteJSONString(val)
}

// normalizeJSONNumber converts a value matching reDigitals to JSON number
// syntax without floating-point conversion, preserving decimal precision and
// exponents even when the number is outside the float64 range.
func normalizeJSONNumber(value string) string {
	sign := ""
	if value[0] == '-' {
		sign, value = "-", value[1:]
	} else if value[0] == '+' {
		value = value[1:]
	}
	exponent := ""
	if i := strings.IndexAny(value, "eE"); i >= 0 {
		value, exponent = value[:i], value[i:]
	}
	integer, fraction, hasDecimal := strings.Cut(value, ".")
	integer = strings.TrimLeft(integer, "0")
	if integer == "" {
		integer = "0"
	}
	if hasDecimal {
		// JSON requires digits on both sides of a decimal point.
		if fraction == "" {
			fraction = "0"
		}
		return sign + integer + "." + fraction + exponent
	}
	return sign + integer + exponent
}
