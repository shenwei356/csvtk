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

	"github.com/casbin/govaluate"
	"github.com/mattn/go-runewidth"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// filter2Cmd represents the filter command
var filter2Cmd = &cobra.Command{
	GroupID: "set",

	Use:   "filter2",
	Short: "filter rows by awk-like arithmetic/string expressions",
	Long: `filter rows by awk-like arithmetic/string expressions

The arithmetic/string expression is supported by:

  https://github.com/casbin/govaluate

Numeric field conversion preserves 64-bit floating-point precision.
Use -s/--numeric-as-string to keep numeric fields as strings.

String field values retain their literal contents, including quotes,
backslashes, and variable-like text (e.g., $name or ${name}).

Compiled expressions are reused while input types remain unchanged.
Literal-dependent expressions retain per-row parsing for compatibility.

The two variable formats can be mixed for the same column.

Variable formats:
  $1 or ${1}                        The first field/column
  $a or ${a}                        Column "a"
  ${a,b} or ${a b} or ${a (b)}      Column name with special characters,
                                    e.g., commas, spaces, and parentheses

Supported operators and types:

  Modifiers: + - / * & | ^ ** % >> <<
  Comparators: > >= < <= == != =~ !~ in
  Logical ops: || &&
  Numeric constants, as 64-bit floating point (12345.678)
  String constants (single quotes: 'foobar')
  Date constants (single quotes)
  Boolean constants: true false
  Parentheses to control the order of evaluation ( )
  Arrays (anything separated by , within parentheses: (1, 2, 'foo'))
  Prefixes: ! - ~
  Ternary conditional: ? :
  Null coalescence: ??

Custom functions:
  - len(), length of strings, e.g., len($1), len($a), len($1, $2)
  - ulen(), length of Unicode strings/width of Unicode strings rendered
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

		filterStr0 := filterStr
		filterStr, fs := normalizeExpressionReferences(filterStr)

		varSep := "__sep__"
		fieldStr := strings.Join(fs, varSep)

		// ------------------------------------------------------

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
						n += len(s)
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
			if regexp.MustCompile(f + `\(.+\)`).MatchString(filterStr) {
				containCustomFuncs = true
				break
			}
		}

		// -----------------------------------

		hasNullCoalescence := reNullCoalescence.MatchString(filterStr)

		var expression *govaluate.EvaluableExpression
		var err error

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

		showRowNumber := printLineNumber || config.ShowRowNumber

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk filter2: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr:      fieldStr,
				FieldStrSep:   varSep,
				FuzzyFields:   fuzzyFields,
				ShowRowNumber: showRowNumber,

				DoNotAllowDuplicatedColumnName: true,
			})

			var parameters map[string]interface{}

			var flag bool
			var col string
			var fieldTmp int
			var value string
			var result interface{}
			var N int64
			var i int
			var ok bool
			var valueFloat float64
			var colnames2fileds map[string][]int // column name -> []field
			var colnamesMap map[string]*regexp.Regexp

			var selectWithColnames bool

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
							if showRowNumber {
								if i == 0 {
									continue
								}
								i--
							}
							if _, ok = colnames2fileds[col]; !ok {
								colnames2fileds[col] = []int{record.Fields[i]}
							} else {
								colnames2fileds[col] = append(colnames2fileds[col], record.Fields[i])
							}
							colnamesMap[col] = fuzzyField2Regexp(col)
						}

						if config.NoOutHeader {
							continue
						}
						if showRowNumber {
							unshift(&record.All, "row")
						}
						checkError(writer.Write(record.All))
						continue
					}
				}

				N++

				// prepaire parameters
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
					bindings = newExpressionBindings(filterStr, parameters, false)
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
				flag = false

				result, err = expression.Evaluate(boundParameters)
				if err != nil {
					flag = false
					if config.Verbose {
						log.Warningf("row %d: %s", N, err)
					}
					continue
				}
				switch result := result.(type) {
				case bool:
					if result {
						flag = true
					}
				default:
					checkError(fmt.Errorf("filter is not boolean expression: %s", filterStr0))
				}

				if !flag {
					continue
				}

				if showRowNumber {
					unshift(&record.All, strconv.Itoa(record.Row))
				}
				checkError(writer.Write(record.All))
			}

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	RootCmd.AddCommand(filter2Cmd)
	filter2Cmd.Flags().StringP("filter", "f", "", `awk-like filter condition. e.g. '$age>12' or '$1 > $3' or '$name=="abc"' or '$1 % 2 == 0'`)
	filter2Cmd.Flags().BoolP("line-number", "n", false, `print the line number as the first column ("n")`)
	filter2Cmd.Flags().BoolP("numeric-as-string", "s", false, `treat even numeric fields as strings to avoid converting big numbers into scientific notation`)
}

var reFilter2 = regexp.MustCompile(`\$\{([^}]+?)\}|\$([^ +-/*&\|^%><!~=()"']+)`)
