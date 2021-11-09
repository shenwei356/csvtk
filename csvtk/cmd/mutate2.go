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
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/mattn/go-runewidth"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// mutate2Cmd represents the mutate command
var mutate2Cmd = &cobra.Command{
	Use:   "mutate2",
	Short: "create new column from selected fields by awk-like arithmetic/string expressions",
	Long: `create new column from selected fields by awk-like arithmetic/string expressions

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

		name := getFlagString(cmd, "name")
		if !config.NoHeaderRow && name == "" {
			checkError(fmt.Errorf("falg -n (--name) needed"))
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

		exprStr := getFlagString(cmd, "expression")
		if exprStr == "" {
			checkError(fmt.Errorf("flag -e (--expression) needed"))
		}

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
			if regexp.MustCompile(f + `\(.+\)`).MatchString(exprStr) {
				containCustomFuncs = true
				break
			}
		}

		// expressions doe not contains `$`
		if !reFilter2.MatchString(exprStr) {
			// checkError(fmt.Errorf("invalid expression: %s", exprStr))
			var expression *govaluate.EvaluableExpression
			if containCustomFuncs {
				expression, err = govaluate.NewEvaluableExpressionWithFunctions(exprStr, functions)
			} else {
				expression, err = govaluate.NewEvaluableExpression(exprStr)
			}
			checkError(err)
			var result interface{}
			result, err = expression.Evaluate(emptyParams)
			if err != nil {
				checkError(fmt.Errorf("fail to evaluate: %s", exprStr))
			}
			result2 := fmt.Sprintf("%v", result)
			for _, file := range files {
				var csvReader *CSVReader
				csvReader, err = newCSVReaderByConfig(config, file)
				checkError(err)
				csvReader.Run()

				isHeaderLine := !config.NoHeaderRow
				printMetaLine := true
				for chunk := range csvReader.Ch {
					checkError(chunk.Err)

					if printMetaLine && len(csvReader.MetaLine) > 0 {
						outfh.WriteString(fmt.Sprintf("sep=%s\n", string(writer.Comma)))
						printMetaLine = false
					}

					for _, record := range chunk.Data {
						if isHeaderLine {
							checkError(writer.Write(append(record, name)))
							isHeaderLine = false
							continue
						}
						checkError(writer.Write(append(record, result2)))
					}
				}

				readerReport(&config, csvReader, file)
			}

			return
		}

		digits := getFlagNonNegativeInt(cmd, "digits")
		formatDigitals := fmt.Sprintf("%%.%df", digits)

		digitsAsString := getFlagBool(cmd, "numeric-as-string")

		fs := make([]string, 0)
		for _, f := range reFilter2.FindAllStringSubmatch(exprStr, -1) {
			fs = append(fs, f[1])
		}

		fieldStr := strings.Join(fs, ",")

		hasNullCoalescence := reNullCoalescence.MatchString(exprStr)

		var quote string = `'`
		if strings.Contains(exprStr, `"`) {
			quote = `"`
		}
		exprStr = reFiler2VarSymbolStartsWithDigits.ReplaceAllString(exprStr, "shenwei_$1$2")
		exprStr = reFilter2VarField.ReplaceAllString(exprStr, "shenwei$1")
		// exprStr = reFilter2VarSymbol.ReplaceAllString(exprStr, "")

		var exprStr1 string
		var expression *govaluate.EvaluableExpression

		usingColname := true

		fields, colnames, negativeFields, needParseHeaderRow, _ := parseFields(cmd, fieldStr, config.NoHeaderRow)
		if negativeFields {
			checkError(fmt.Errorf("unselect not allowed"))
		}
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

		// fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		fuzzyFields := false

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

			handleHeaderRow := needParseHeaderRow
			checkFields := true
			var col string
			var fieldTmp int
			var value string
			var valueFloat float64
			var result interface{}

			var record2 []string // for output

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

						parseHeaderRow = false
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
						fieldsMap = make(map[int]struct{}, len(fields))
						for _, f := range fields {
							fieldsMap[f] = struct{}{}
						}

						checkFields = false
					}

					record2 = record
					for f := range record {
						record2[f] = record[f]
						if _, ok := fieldsMap[f+1]; ok {
							if handleHeaderRow {
								record2 = append(record2, name)
								handleHeaderRow = false
							} else {
								if !usingColname {
									for _, fieldTmp = range fields {
										value = record[fieldTmp-1]
										col = fmt.Sprintf("shenwei%d", fieldTmp)

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
												parameters[col] = quote + value + quote
											}
										}
									}
								}

								exprStr1 = exprStr
								for col, value = range parameters {
									exprStr1 = strings.ReplaceAll(exprStr1, col, value)
								}
								if containCustomFuncs {
									expression, err = govaluate.NewEvaluableExpressionWithFunctions(exprStr1, functions)
								} else {
									expression, err = govaluate.NewEvaluableExpression(exprStr1)
								}
								checkError(err)

								if hasNullCoalescence {
									result, err = expression.Evaluate(parameters2)
								} else {
									result, err = expression.Evaluate(emptyParams)
								}
								if err != nil {
									checkError(fmt.Errorf("data: %s, err: %s", record, err))
								}
								switch result.(type) {
								case bool:
									record2 = append(record2, fmt.Sprintf("%v", result))
								case float32, float64:
									record2 = append(record2, fmt.Sprintf(formatDigitals, result))
								case int, int32, int64:
									record2 = append(record2, fmt.Sprintf("%d", result))
								default:
									record2 = append(record2, fmt.Sprintf("%s", result))
								}
							}
							break
						}
					}
					checkError(writer.Write(record2))
				}
			}

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	RootCmd.AddCommand(mutate2Cmd)
	mutate2Cmd.Flags().StringP("expression", "e", "", `arithmetic/string expressions. e.g. "'string'", '"abc"', ' $a + "-" + $b ', '$1 + $2', '$a / $b', ' $1 > 100 ? "big" : "small" '`)
	mutate2Cmd.Flags().StringP("name", "n", "", `new column name`)
	mutate2Cmd.Flags().IntP("digits", "L", 2, `number of digits after the decimal dot`)
	mutate2Cmd.Flags().BoolP("numeric-as-string", "s", false, `treat even numeric fields as strings to avoid converting big numbers into scientific notation`)
}

var reNullCoalescence = regexp.MustCompile(`\?\?`)
