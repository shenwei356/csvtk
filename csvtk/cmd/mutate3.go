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
	"regexp"
	"runtime"
	"sort"
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

Variables formats:
  $1 or ${1}                        The first field/column
  $a or ${a}                        Column "a"
  ${a,b} or ${a b} or ${a (b)}      Column name with special charactors,
                                    e.g., commas, spaces, and parentheses

Supported Operators:

  Arithmetic: + - / * ^ ** %
  Comparison: > >= < <= == !=
  Logical: not ! and && or ||
  String: + contains startsWith endsWith
  Regex: matches
  Range: ..
  Slice: [:]
  Pipe: |
  Ternary conditional: ? :
  Null coalescence: ??

Supported Literals:

  Arrays: [1, 2, 3]
  Boolean: true false
  Float: 0.5 .5
  Integer: 42 0x2A 0o52 0b101010
  Map: {a: 1, b: 2}
  Null: nil
  String: "foo" 'bar'

See Expr language definition link for documentation on built-in functions.

Custom functions:
  - ulen(), length of unicode strings/width of unicode strings rendered
    to a terminal, e.g., len("沈伟")==6, ulen("沈伟")==4
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

	emptyParams := make(map[string]interface{})

	fs := make([]string, 0)
	varType := make(map[string]int)
	for _, f := range reFilter2.FindAllStringSubmatch(opts.ExprStr, -1) {
		if reFilter2b.MatchString(f[0]) {
			varType[f[1]] = 1
			fs = append(fs, f[1])
		} else {
			varType[f[2]] = 0
			fs = append(fs, f[2])
		}
	}

	varSep := "__sep__"
	fieldStr := strings.Join(fs, varSep)

	hasNullCoalescence := reNullCoalescence.MatchString(opts.ExprStr)

	var quote string
	opts.ExprStr = reFiler2VarSymbolStartsWithDigits.ReplaceAllString(opts.ExprStr, "shenwei_$1$2")
	opts.ExprStr = reFilter2VarField.ReplaceAllString(opts.ExprStr, "shenwei$1")

	var exprStr1 string
	var program *vm.Program

	customFuncs := []expr.Option{
		expr.Function(
			"ulen",
			func(args ...interface{}) (interface{}, error) {
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

		var parameters map[string]string
		var parameters2 map[string]interface{}

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
		var fieldsUniq []int
		var selectWithColnames bool
		var record2 []string // for output
		keys := make([]string, 0, 8)

		checkFirstLine := true
		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if checkFirstLine {
				checkFirstLine = false

				selectWithColnames = record.SelectWithColnames

				parameters = make(map[string]string, len(record.All))
				parameters2 = make(map[string]interface{}, len(record.All))
				parameters2["shenweiNULL"] = nil

				fieldsUniq = UniqInts(record.Fields)

				if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
					colnames2fileds = make(map[string][]int, len(record.All))

					colnamesMap = make(map[string]*regexp.Regexp, len(record.All))
					for i, col = range record.All {
						if _, ok = colnames2fileds[col]; !ok {
							colnames2fileds[col] = []int{i + 1}
						} else {
							colnames2fileds[col] = append(colnames2fileds[col], i+1)
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
				for _, fieldTmp = range fieldsUniq {
					value = record.All[fieldTmp-1]
					col = strconv.Itoa(fieldTmp)
					if varType[col] == 1 {
						col = "${" + col + "}"
					} else {
						col = fmt.Sprintf("shenwei%d", fieldTmp)
					}

					quote = `'`

					if reDigitals.MatchString(value) {
						if opts.DigitsAsString {
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
					value = record.All[colnames2fileds[col][0]-1]

					if reFiler2ColSymbolStartsWithDigits.MatchString(col) {
						col = fmt.Sprintf("shenwei_%s", col)
					} else if varType[col] == 1 {
						col = "${" + col + "}"
					} else {
						col = "$" + col
					}

					quote = `'`

					if reDigitals.MatchString(value) {
						if opts.DigitsAsString {
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
			exprStr1 = opts.ExprStr
			for _, col = range keys {
				exprStr1 = strings.ReplaceAll(exprStr1, col, parameters[col])
			}

			// evaluate
			program, err = expr.Compile(exprStr1, customFuncs...)
			checkError(err)

			decimalFormat := fmt.Sprintf("%%.%df", opts.DecimalWidth)

			// check result
			if hasNullCoalescence {
				result, err = expr.Run(program, parameters2)
			} else {
				result, err = expr.Run(program, emptyParams)
			}
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
	mutate3Cmd.Flags().IntP("decimal-width", "w", 2, "limit floats to N decimal points")
	mutate3Cmd.Flags().IntP("at", "", 0, "where the new column should appear, 1 for the 1st column, 0 for the last column")
	mutate3Cmd.Flags().StringP("after", "", "", "insert the new column right after the given column name")
	mutate3Cmd.Flags().StringP("before", "", "", "insert the new column right before the given column name")
}
