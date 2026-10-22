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

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/mattn/go-runewidth"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// mutate3Cmd represents the mutate command
var mutate3Cmd = &cobra.Command{
	GroupID: "edit",

	Use:   "mutate3",
	Short: "create a new column from selected fields with Go-like expressions",
	Long: `create a new column from selected fields with Go-like expressions

The expression language is supported by Expr:

  https://expr-lang.org/docs/language-definition

Numeric field conversion preserves 64-bit floating-point precision.
Use -s/--numeric-as-string to keep numeric fields as strings.
The -w/--decimal-width flag controls rounding of numeric output.

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

Supported Operators:

  Arithmetic: + - / * ^ ** %
  Comparison: > >= < <= == !=
  Logical: not ! and && or ||
  String: + contains startsWith endsWith
          trim trimPrefix trimSuffix upper lower split splitAfter replace
          repeat indexOf lastIndexOf hasPrefix hasSuffix
  Membership: [] . ?. in
  Regex: matches
  Range: ..
  Slice: [:]
  Pipe: |
  Ternary conditional: ? :
  Null coalescence: ??
  Date: now duration date timezone
  Number: max min abs ceil floor round

Supported Literals:

  Arrays: [1, 2, 3]
  Boolean: true false
  Float: 0.5 .5
  Integer: 42 0x2A 0o52 0b101010
  Map: {a: 1, b: 2}
  Null: nil
  String: "foo" 'bar'

See the Expr language definition link for documentation on built-in functions.

Custom functions:
  - ulen(), length of Unicode strings/width of Unicode strings rendered
    to a terminal, e.g., len("沈伟")==2, ulen("沈伟")==4
`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		opts := mutate3Opts{}
		opts.At = getFlagNonNegativeInt(cmd, "at")
		opts.After = getFlagString(cmd, "after")
		opts.Before = getFlagString(cmd, "before")
		if config.NoHeaderRow {
			if opts.After != "" {
				checkError(fmt.Errorf("the flag --after is not allowed with -H/--no-header-row"))
			}
			if opts.Before != "" {
				checkError(fmt.Errorf("the flag --before is not allowed with -H/--no-header-row"))
			}
		}
		if opts.After != "" && opts.Before != "" {
			checkError(fmt.Errorf("the flag --after and --before are incompatible"))
		}
		if opts.At > 0 && !(opts.After == "" && opts.Before == "") {
			checkError(fmt.Errorf("the flag --at is incompatible with --after and --before"))
		}

		opts.Files = getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(opts.Files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		opts.Name = getFlagString(cmd, "name")
		if !config.NoHeaderRow && opts.Name == "" && !config.NoOutHeader {
			checkError(fmt.Errorf("flag -n (--name) needed"))
		}

		opts.ExprStr = getFlagString(cmd, "expression")
		if opts.ExprStr == "" {
			checkError(fmt.Errorf("flag -e (--expression) needed"))
		}

		opts.DecimalWidth = getFlagNonNegativeInt(cmd, "decimal-width")
		opts.DigitsAsString = getFlagBool(cmd, "numeric-as-string")

		doMutate3(config, opts)
	},
}

type mutate3Opts struct {
	After          string
	At             int
	Before         string
	DecimalWidth   int
	DigitsAsString bool
	ExprStr        string
	Files          []string
	Name           string
}

func doMutate3(config Config, opts mutate3Opts) {
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

	normalized, fs := normalizeExpressionReferences(opts.ExprStr)
	opts.ExprStr = normalized

	varSep := "__sep__"
	fieldStr := strings.Join(fs, varSep)

	hasNullCoalescence := reNullCoalescence.MatchString(opts.ExprStr)

	var program *vm.Program

	customFuncs := []expr.Option{
		expr.Function(
			"ulen",
			func(args ...interface{}) (interface{}, error) {
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
			new(func(int) float64),
			new(func(float64) float64),
			new(func(string) float64),
		),
	}

	fuzzyFields := false

	for _, file := range opts.Files {
		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				if config.Verbose {
					log.Warningf("csvtk mutate3: skipping empty input file: %s", file)
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
		decimalFormat := fmt.Sprintf("%%.%df", opts.DecimalWidth)

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

					value = opts.Name
					record2 = record.All
					record2 = append(record2, value)

					if opts.After != "" {
						if _fields, ok = colnames2fileds[opts.After]; ok {
							opts.At = _fields[len(_fields)-1] + 1
						} else {
							checkError(fmt.Errorf(`column "%s" not existed in file: %s`, opts.After, file))
						}
						copy(record2[opts.At:], record2[opts.At-1:len(record2)-1])
						record2[opts.At-1] = value
					} else if opts.Before != "" {
						if _fields, ok = colnames2fileds[opts.Before]; ok {
							opts.At = _fields[0]
						} else {
							checkError(fmt.Errorf(`column "%s" not existed in file: %s`, opts.Before, file))
						}
						copy(record2[opts.At:], record2[opts.At-1:len(record2)-1])
						record2[opts.At-1] = value
					} else if opts.At > 0 && opts.At <= len(record2) {
						copy(record2[opts.At:], record2[opts.At-1:len(record2)-1])
						record2[opts.At-1] = value
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
						if opts.DigitsAsString {
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
						if opts.DigitsAsString {
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
				bindings = newExpressionBindings(opts.ExprStr, parameters, true)
				program = nil
			}
			source, boundParameters, err := bindings.prepare(parameters)
			checkError(err)
			if program == nil || bindings.recompile {
				program, err = expr.Compile(source, append([]expr.Option{expr.Env(boundParameters)}, customFuncs...)...)
				checkError(err)
			}

			// check result
			result, err = expr.Run(program, boundParameters)
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

			if opts.After != "" {
				if _fields, ok = colnames2fileds[opts.After]; ok {
					opts.At = _fields[len(_fields)-1] + 1
				} else {
					checkError(fmt.Errorf(`column "%s" not existed in file: %s`, opts.After, file))
				}
				copy(record2[opts.At:], record2[opts.At-1:len(record2)-1])
				record2[opts.At-1] = value
			} else if opts.Before != "" {
				if _fields, ok = colnames2fileds[opts.Before]; ok {
					opts.At = _fields[0]
				} else {
					checkError(fmt.Errorf(`column "%s" not existed in file: %s`, opts.Before, file))
				}
				copy(record2[opts.At:], record2[opts.At-1:len(record2)-1])
				record2[opts.At-1] = value
			} else if opts.At > 0 && opts.At <= len(record2) {
				copy(record2[opts.At:], record2[opts.At-1:len(record2)-1])
				record2[opts.At-1] = value
			}

			checkError(writer.Write(record2))
		}

		readerReport(&config, csvReader, file)
	}
}

func init() {
	RootCmd.AddCommand(mutate3Cmd)
	mutate3Cmd.Flags().StringP("expression", "e", "", `arithmetic/string expressions. e.g. "'string'", '"abc"', ' $a + "-" + $b ', '$1 + $2', '$a / $b', ' $1 > 100 ? "big" : "small" '`)
	mutate3Cmd.Flags().StringP("name", "n", "", `new column name`)
	mutate3Cmd.Flags().BoolP("numeric-as-string", "s", false, `treat even numeric fields as strings to avoid converting big numbers into scientific notation`)
	mutate3Cmd.Flags().IntP("decimal-width", "w", 2, "limit floats to N decimal places")
	mutate3Cmd.Flags().IntP("at", "", 0, "where the new column should appear, 1 for the 1st column, 0 for the last column")
	mutate3Cmd.Flags().StringP("after", "", "", "insert the new column right after the given column name")
	mutate3Cmd.Flags().StringP("before", "", "", "insert the new column right before the given column name")
}
