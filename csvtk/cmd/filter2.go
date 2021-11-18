// Copyright © 2016-2021 Wei Shen <shenwei356@gmail.com>
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
	"sort"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/mattn/go-runewidth"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// filter2Cmd represents the filter command
var filter2Cmd = &cobra.Command{
	Use:   "filter2",
	Short: "filter rows by awk-like arithmetic/string expressions",
	Long: `filter rows by awk-like arithmetic/string expressions

The arithmetic/string expression is supported by:

  https://github.com/Knetic/govaluate

Supported operators and types:

  Modifiers: + - / * & | ^ ** % >> <<
  Comparators: > >= < <= == != =~ !~
  Logical ops: || &&
  Numeric constants, as 64-bit floating point (12345.678)
  String constants (single quotes: 'foobar')
  Date constants (single quotes)
  Boolean constants: true false
  Parenthesis to control order of evaluation ( )
  Arrays (anything separated by , within parenthesis: (1, 2, 'foo'))
  Prefixes: ! - ~
  Ternary conditional: ? :
  Null coalescence: ??

Custom functions:
  - len(), length of strings, e.g., len($1), len($a), len($1, $2)
  - ulen(), length of unicode strings/width of unicode strings rendered
    to a terminal, e.g., len("沈伟")==6, ulen("沈伟")==4

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		filterStr := getFlagString(cmd, "filter")
		printLineNumber := getFlagBool(cmd, "line-number")
		fuzzyFields := false

		if filterStr == "" {
			checkError(fmt.Errorf("flag -f (--filter) needed"))
		}

		if !reFilter2.MatchString(filterStr) {
			checkError(fmt.Errorf("invalid filter: %s", filterStr))
		}

		digitsAsString := getFlagBool(cmd, "numeric-as-string")

		fs := make([]string, 0)
		for _, f := range reFilter2.FindAllStringSubmatch(filterStr, -1) {
			fs = append(fs, f[1])
		}

		fieldStr := strings.Join(fs, ",")

		// ------------------------------------------------------

		// custom functions
		functions := map[string]govaluate.ExpressionFunction{
			"len": func(args ...interface{}) (interface{}, error) {
				n := 0
				for _, s := range args {
					switch s.(type) {
					case int:
						n += len(fmt.Sprintf("%d", s.(int)))
					case float64:
						n += len(fmt.Sprintf("%f", s.(float64)))
					case string:
						n += len(s.(string))
					}

				}
				return float64(n), nil
			},
			"ulen": func(args ...interface{}) (interface{}, error) {
				n := 0
				for _, s := range args {
					switch s.(type) {
					case int:
						n += runewidth.StringWidth(fmt.Sprintf("%d", s.(int)))
					case float64:
						n += runewidth.StringWidth(fmt.Sprintf("%f", s.(float64)))
					case string:
						n += runewidth.StringWidth(s.(string))
					}

				}
				return float64(n), nil
			},
		}

		emptyParams := make(map[string]interface{})

		containCustomFuncs := false
		for f := range functions {
			if regexp.MustCompile(f + `\(.+\)`).MatchString(filterStr) {
				containCustomFuncs = true
				break
			}
		}

		// -----------------------------------

		hasNullCoalescence := reNullCoalescence.MatchString(filterStr)

		var quote string
		filterStr0 := filterStr
		filterStr = reFiler2VarSymbolStartsWithDigits.ReplaceAllString(filterStr, "shenwei_$1$2")
		filterStr = reFilter2VarField.ReplaceAllString(filterStr, "shenwei$1")
		// filterStr = reFilter2VarSymbol.ReplaceAllString(filterStr, "")

		var filterStr1 string
		var expression *govaluate.EvaluableExpression
		var err error

		usingColname := true

		fields, colnames, negativeFields, needParseHeaderRow, _ := parseFields(cmd, fieldStr, config.NoHeaderRow)
		var fieldsMap map[int]struct{}
		if len(fields) > 0 {
			usingColname = false
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
			checkError(err)
			csvReader.Run()

			parseHeaderRow := needParseHeaderRow // parsing header row
			var colnames2fileds map[string]int   // column name -> field
			var colnamesMap map[string]*regexp.Regexp

			parameters := make(map[string]string, len(colnamesMap))
			parameters2 := make(map[string]interface{}, len(colnamesMap))
			parameters2["shenweiNULL"] = nil

			checkFields := true
			var flag bool
			var col string
			var fieldTmp int
			var value string
			var result interface{}
			var N int64
			var recordWithN []string
			var valueFloat float64

			keys := make([]string, 0, 8)

			printMetaLine := true
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				if printMetaLine && len(csvReader.MetaLine) > 0 {
					outfh.WriteString(fmt.Sprintf("sep=%s\n", string(writer.Comma)))
					printMetaLine = false
				}

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
								if ok {
									fields = append(fields, colnames2fileds[col])
								}
							}
						}

						fieldsMap = make(map[int]struct{}, len(fields))
						for _, f := range fields {
							fieldsMap[f] = struct{}{}
						}

						if printLineNumber {
							recordWithN = []string{"n"}
							recordWithN = append(recordWithN, record...)
							record = recordWithN
						}
						checkError(writer.Write(record))
						parseHeaderRow = false
						continue
					}
					N++

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
						fieldsMap = make(map[int]struct{}, len(fields))
						for _, f := range fields {
							fieldsMap[f] = struct{}{}
						}

						checkFields = false
					}

					// prepaire parameters
					if !usingColname {
						for _, fieldTmp = range fields {
							value = record[fieldTmp-1]
							col = fmt.Sprintf("shenwei%d", fieldTmp)

							quote = `'`

							if reDigitals.MatchString(value) {
								if digitsAsString || containCustomFuncs {
									parameters[col] = quote + value + quote
								} else {
									valueFloat, _ = strconv.ParseFloat(removeComma(value), 64)
									parameters[col] = fmt.Sprintf("%.16f", valueFloat)
								}
							} else {
								if value == "" && hasNullCoalescence {
									parameters[col] = "shenweiNULL"
								} else {
									if strings.Contains(value, `'`) {
										value = strings.ReplaceAll(value, `'`, `\'`)
									}
									if strings.Contains(value, `"`) {
										value = strings.ReplaceAll(value, `"`, `\"`)
									}

									parameters[col] = quote + value + quote
								}
							}
						}
					} else {
						for col = range colnamesMap {
							value = record[colnames2fileds[col]-1]

							if reFiler2ColSymbolStartsWithDigits.MatchString(col) {
								col = fmt.Sprintf("shenwei_%s", col)
							} else {
								col = "$" + col
							}

							quote = `'`

							if reDigitals.MatchString(value) {
								if digitsAsString || containCustomFuncs {
									parameters[col] = quote + value + quote
								} else {
									valueFloat, _ = strconv.ParseFloat(removeComma(value), 64)
									parameters[col] = fmt.Sprintf("%.16f", valueFloat)
								}
							} else {
								if value == "" && hasNullCoalescence {
									parameters[col] = "shenweiNULL"
								} else {
									if strings.Contains(value, `'`) {
										value = strings.ReplaceAll(value, `'`, `\'`)
									}
									if strings.Contains(value, `"`) {
										value = strings.ReplaceAll(value, `"`, `\"`)
									}

									parameters[col] = quote + value + quote
								}
							}
						}
					}

					// sort variable names by length, so we can replace variables in the right order.
					// e.g., for -e '$reads_mapped/$reads', we should firstly replace $reads_mapped then $reads.
					keys = keys[:0]
					for col = range parameters {
						keys = append(keys, col)
					}
					sort.Slice(keys, func(i, j int) bool {
						return len(keys[i]) > len(keys[j])
					})

					// replace variable with column data
					filterStr1 = filterStr
					for _, col = range keys {
						filterStr1 = strings.ReplaceAll(filterStr1, col, parameters[col])
					}

					// evaluate
					if containCustomFuncs {
						expression, err = govaluate.NewEvaluableExpressionWithFunctions(filterStr1, functions)
					} else {
						expression, err = govaluate.NewEvaluableExpression(filterStr1)
					}
					checkError(err)

					// check result
					flag = false

					if hasNullCoalescence {
						result, err = expression.Evaluate(parameters2)
					} else {
						result, err = expression.Evaluate(emptyParams)
					}
					if err != nil {
						flag = false
						log.Warningf("row %d: %s", N, err)
						continue
					}
					switch result.(type) {
					case bool:
						if result.(bool) {
							flag = true
						}
					default:
						checkError(fmt.Errorf("filter is not boolean expression: %s", filterStr0))
					}

					if !flag {
						continue
					}

					if printLineNumber {
						recordWithN = []string{fmt.Sprintf("%d", N)}
						recordWithN = append(recordWithN, record...)
						record = recordWithN
					}
					checkError(writer.Write(record))
				}
			}

			readerReport(&config, csvReader, file)
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(filter2Cmd)
	filter2Cmd.Flags().StringP("filter", "f", "", `awk-like filter condition. e.g. '$age>12' or '$1 > $3' or '$name=="abc"' or '$1 % 2 == 0'`)
	filter2Cmd.Flags().BoolP("line-number", "n", false, `print line number as the first column ("n")`)
	filter2Cmd.Flags().BoolP("numeric-as-string", "s", false, `treat even numeric fields as strings to avoid converting big numbers into scientific notation`)
}

var reFilter2 = regexp.MustCompile(`\$([^ +-/*&\|^%><!~=()"']+)`)
var reFilter2VarField = regexp.MustCompile(`\$(\d+)`)

// var reFilter2VarSymbol = regexp.MustCompile(`\$`)

// special colname starting with digits, e.g., 123abc
var reFiler2VarSymbolStartsWithDigits = regexp.MustCompile(`\$(\d+)([^\d +-/*&\|^%><!~=()"']+)`) // for preprocess expression
var reFiler2ColSymbolStartsWithDigits = regexp.MustCompile(`^(\d+)([^\d +-/*&\|^%><!~=()"']+)`)  // for preparing paramters
