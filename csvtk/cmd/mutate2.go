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
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/casbin/govaluate"
	"github.com/mattn/go-runewidth"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// mutate2Cmd represents the mutate command
var mutate2Cmd = &cobra.Command{
	GroupID: "edit",

	Use:   "mutate2",
	Short: "create a new column from selected fields by awk-like arithmetic/string expressions",
	Long: `create a new column from selected fields by awk-like arithmetic/string expressions

The arithmetic/string expression is supported by:

  https://github.com/casbin/govaluate

Numeric field conversion preserves 64-bit floating-point precision.
Use -s/--numeric-as-string to keep numeric fields as strings.
The -w/--decimal-width flag controls rounding of numeric output.

String field values retain their literal contents, including quotes,
backslashes, and variable-like text (e.g., $name or ${name}).

Compiled expressions are reused while input types remain unchanged.
Literal-dependent expressions retain per-row parsing for compatibility.

The two variable formats can be mixed for the same column.

Variables formats:
  $1 or ${1}                        The first field/column
  $a or ${a}                        Column "a"
  ${a,b} or ${a b} or ${a (b)}      Column name with special charactors, 
                                    e.g., commas, spaces, and parentheses

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
  - len(), characters of strings, e.g., len($1), len($a), len($1, $2).
    Unicode is supported, e.g., len("沈伟")==2
  - ulen(), length of unicode strings/width of unicode strings rendered
    to a terminal, e.g., len("沈伟")==2, ulen("沈伟")==4

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		at := getFlagNonNegativeInt(cmd, "at")
		after := getFlagString(cmd, "after")
		before := getFlagString(cmd, "before")
		if config.NoHeaderRow {
			if after != "" {
				checkError(fmt.Errorf("the flag --after is not allowed with -H/--no-header-row"))
			}
			if before != "" {
				checkError(fmt.Errorf("the flag --before is not allowed with -H/--no-header-row"))
			}
		}
		if after != "" && before != "" {
			checkError(fmt.Errorf("the flag --after and --before are incompatible"))
		}
		if at > 0 && !(after == "" && before == "") {
			checkError(fmt.Errorf("the flag --at is incompatible with --after and --before"))
		}

		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		name := getFlagString(cmd, "name")
		if !config.NoHeaderRow && name == "" && !config.NoOutHeader {
			checkError(fmt.Errorf("falg -n (--name) needed"))
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer closeOutput(outfh)

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

		exprStr := getFlagString(cmd, "expression")
		if exprStr == "" {
			checkError(fmt.Errorf("flag -e (--expression) needed"))
		}

		// custom functions
		functions := map[string]govaluate.ExpressionFunction{
			"len": func(args ...interface{}) (interface{}, error) {
				n := 0
				for _, s := range args {
					switch s := s.(type) {
					case int:
						n += len(fmt.Sprintf("%d", s))
					case float64:
						n += len(fmt.Sprintf("%f", s))
					case string:
						n += utf8.RuneCountInString(s)
					}

				}
				return float64(n), nil
			},
			"ulen": func(args ...interface{}) (interface{}, error) {
				n := 0
				for _, s := range args {
					switch s := s.(type) {
					case int:
						n += runewidth.StringWidth(fmt.Sprintf("%d", s))
					case float64:
						n += runewidth.StringWidth(fmt.Sprintf("%f", s))
					case string:
						n += runewidth.StringWidth(s)
					}

				}
				return float64(n), nil
			},
		}

		containCustomFuncs := false
		for f := range functions {
			if regexp.MustCompile(f + `\(.+\)`).MatchString(exprStr) {
				containCustomFuncs = true
				break
			}
		}

		decimalWidth := getFlagNonNegativeInt(cmd, "decimal-width")
		decimalFormat := fmt.Sprintf("%%.%df", decimalWidth)

		digitsAsString := getFlagBool(cmd, "numeric-as-string")

		exprStr, fs := normalizeExpressionReferences(exprStr)

		varSep := "__sep__"
		fieldStr := strings.Join(fs, varSep)

		hasNullCoalescence := reNullCoalescence.MatchString(exprStr)

		var expression *govaluate.EvaluableExpression

		fuzzyFields := false

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk mutate2: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr:    fieldStr,
				FieldStrSep: varSep,
				FuzzyFields: fuzzyFields,

				DoNotAllowDuplicatedColumnName: true,
			})

			var parameters map[string]interface{}

			var col string
			var fieldTmp int
			var _fields []int
			var i int
			var ok bool
			var value string
			var valueFloat float64
			var result interface{}
			var colnames2fileds map[string][]int // column name -> []field
			var colnamesMap map[string]*regexp.Regexp
			var selectWithColnames bool
			var record2 []string // for output
			var bindings *expressionBindings

			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					checkFirstLine = false

					selectWithColnames = record.SelectWithColnames

					parameters = make(map[string]interface{}, len(record.All))

					if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
						colnames2fileds = make(map[string][]int, len(record.Selected))

						colnamesMap = make(map[string]*regexp.Regexp, len(record.Selected))
						for i, col = range record.Selected {
							if _, ok = colnames2fileds[col]; !ok {
								colnames2fileds[col] = []int{record.Fields[i]}
							} else {
								colnames2fileds[col] = append(colnames2fileds[col], record.Fields[i])
							}

							colnamesMap[col] = fuzzyField2Regexp(col)
						}

						value = name
						record2 = record.All
						record2 = append(record2, value)

						if after != "" {
							if _fields, ok = colnames2fileds[after]; ok {
								at = _fields[len(_fields)-1] + 1
							} else {
								checkError(fmt.Errorf(`column "%s" not existed in file: %s`, after, file))
							}
							copy(record2[at:], record2[at-1:len(record2)-1])
							record2[at-1] = value
						} else if before != "" {
							if _fields, ok = colnames2fileds[before]; ok {
								at = _fields[0]
							} else {
								checkError(fmt.Errorf(`column "%s" not existed in file: %s`, before, file))
							}
							copy(record2[at:], record2[at-1:len(record2)-1])
							record2[at-1] = value
						} else if at > 0 && at <= len(record2) {
							copy(record2[at:], record2[at-1:len(record2)-1])
							record2[at-1] = value
						}

						if !config.NoOutHeader {
							checkError(writer.Write(record2))
						}

						continue
					}
				}

				// prepare parameters
				if !selectWithColnames {
					for _, fieldTmp = range record.Fields {
						value = record.All[fieldTmp-1]
						col = "${" + strconv.Itoa(fieldTmp) + "}"

						if reDigitals.MatchString(value) {
							if digitsAsString || containCustomFuncs {
								parameters[col] = value
							} else {
								valueFloat, err = strconv.ParseFloat(removeComma(value), 64)
								checkError(err)
								parameters[col] = valueFloat
							}
						} else {
							if value == "" && hasNullCoalescence {
								parameters[col] = nil
							} else {
								parameters[col] = value
							}
						}
					}
				} else {
					for col = range colnamesMap {
						value = record.All[colnames2fileds[col][0]-1]

						col = "${" + col + "}"

						if reDigitals.MatchString(value) {
							if digitsAsString || containCustomFuncs {
								parameters[col] = value
							} else {
								valueFloat, err = strconv.ParseFloat(removeComma(value), 64)
								checkError(err)
								parameters[col] = valueFloat
							}
						} else {
							if value == "" && hasNullCoalescence {
								parameters[col] = nil
							} else {
								parameters[col] = value
							}
						}
					}
				}

				// Bind column values separately from the expression compiled for this input.
				if bindings == nil {
					bindings = newExpressionBindings(exprStr, parameters, false)
					expression = nil
				}
				source, boundParameters, err := bindings.prepare(parameters)
				checkError(err)
				if expression == nil || bindings.recompile {
					if containCustomFuncs {
						expression, err = govaluate.NewEvaluableExpressionWithFunctions(source, functions)
					} else {
						expression, err = govaluate.NewEvaluableExpression(source)
					}
					checkError(err)
				}

				// check result
				result, err = expression.Evaluate(boundParameters)
				if err != nil {
					checkError(fmt.Errorf("data: %s, err: %s", record.All, err))
				}
				switch result.(type) {
				case bool:
					value = fmt.Sprintf("%v", result)
				case float32, float64:
					value = fmt.Sprintf(decimalFormat, result)
				case int, int32, int64:
					value = fmt.Sprintf("%d", result)
				default:
					value = fmt.Sprintf("%s", result)
				}

				record2 = record.All
				record2 = append(record2, value)

				if after != "" {
					if _fields, ok = colnames2fileds[after]; ok {
						at = _fields[len(_fields)-1] + 1
					} else {
						checkError(fmt.Errorf(`column "%s" not existed in file: %s`, after, file))
					}
					copy(record2[at:], record2[at-1:len(record2)-1])
					record2[at-1] = value
				} else if before != "" {
					if _fields, ok = colnames2fileds[before]; ok {
						at = _fields[0]
					} else {
						checkError(fmt.Errorf(`column "%s" not existed in file: %s`, before, file))
					}
					copy(record2[at:], record2[at-1:len(record2)-1])
					record2[at-1] = value
				} else if at > 0 && at <= len(record2) {
					copy(record2[at:], record2[at-1:len(record2)-1])
					record2[at-1] = value
				}

				checkError(writer.Write(record2))
			}

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	RootCmd.AddCommand(mutate2Cmd)
	mutate2Cmd.Flags().StringP("expression", "e", "", `arithmetic/string expressions. e.g. "'string'", '"abc"', ' $a + "-" + $b ', '$1 + $2', '$a / $b', ' $1 > 100 ? "big" : "small" '`)
	mutate2Cmd.Flags().StringP("name", "n", "", `new column name`)
	mutate2Cmd.Flags().BoolP("numeric-as-string", "s", false, `treat even numeric fields as strings to avoid converting big numbers into scientific notation`)
	mutate2Cmd.Flags().IntP("decimal-width", "w", 2, "limit floats to N decimal points")
	mutate2Cmd.Flags().IntP("at", "", 0, "where the new column should appear, 1 for the 1st column, 0 for the last column")
	mutate2Cmd.Flags().StringP("after", "", "", "insert the new column right after the given column name")
	mutate2Cmd.Flags().StringP("before", "", "", "insert the new column right before the given column name")
}

var reNullCoalescence = regexp.MustCompile(`\?\?`)
