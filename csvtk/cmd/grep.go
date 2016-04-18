// Copyright © 2016 Wei Shen <shenwei356@gmail.com>
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
	"strings"

	"github.com/brentp/xopen"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/shenwei356/breader"
	"github.com/spf13/cobra"
)

// grepCmd represents the grep command
var grepCmd = &cobra.Command{
	Use:   "grep",
	Short: "grep data by selected fields with patterns/regular expressions",
	Long: `grep data by selected fields with patterns/regular expressions

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		fieldStr := getFlagString(cmd, "fields")
		fields, colnames, negativeFields, needParseHeaderRow := parseFields(cmd, fieldStr, config.NoHeaderRow)
		// if !(len(fields) == 1 || len(colnames) == 1) {
		// 	checkError(fmt.Errorf("single fields needed"))
		// }
		// if negativeFields {
		// 	checkError(fmt.Errorf("unselect not allowed"))
		// }
		var fieldsMap map[int]struct{}
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
		}

		patterns := getFlagStringSlice(cmd, "pattern")
		patternFile := getFlagString(cmd, "pattern-file")
		if len(patterns) == 0 && patternFile == "" {
			checkError(fmt.Errorf("one of flags -p (--pattern) or -P (--pattern-file) should be given"))
		}

		ignoreCase := getFlagBool(cmd, "ignore-case")
		useRegexp := getFlagBool(cmd, "use-regexp")
		invert := getFlagBool(cmd, "invert")

		patternsMap := make(map[string]*regexp.Regexp)
		for _, pattern := range patterns {
			if useRegexp {
				p := pattern
				if ignoreCase {
					p = "(?i)" + p
				}
				re, err := regexp.Compile(p)
				checkError(err)
				patternsMap[pattern] = re
			} else {
				if ignoreCase {
					patternsMap[strings.ToLower(pattern)] = nil
				} else {
					patternsMap[pattern] = nil
				}
			}
		}

		if patternFile != "" {
			reader, err := breader.NewDefaultBufferedReader(patternFile)
			checkError(err)
			for chunk := range reader.Ch {
				checkError(chunk.Err)
				for _, data := range chunk.Data {
					pattern := data.(string)
					if useRegexp {
						p := pattern
						if ignoreCase {
							p = "(?i)" + p
						}
						re, err := regexp.Compile(p)
						checkError(err)
						patternsMap[pattern] = re
					} else {
						if ignoreCase {
							patternsMap[strings.ToLower(pattern)] = nil
						} else {
							patternsMap[pattern] = nil
						}
					}
				}
			}
		}

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		// fuzzyFields := false
		noHighlight := getFlagBool(cmd, "no-highlight")
		var writer *csv.Writer
		if config.OutFile == "-" {
			outfh := colorable.NewColorableStdout()
			writer = csv.NewWriter(outfh)
		} else {
			noHighlight = true
			outfh, err := xopen.Wopen(config.OutFile)
			checkError(err)
			defer outfh.Close()
			writer = csv.NewWriter(outfh)
		}

		if config.OutTabs || config.Tabs {
			writer.Comma = '\t'
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
			var HeaderRow []string

			checkFields := true
			var target string
			var hitOne, hit bool

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
							if negativeFields {
								colnamesMap[col[1:]] = fuzzyField2Regexp(col)
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
									fields = append(fields, colnames2fileds[col])
								}
							}
						}

						fieldsMap = make(map[int]struct{}, len(fields))
						for _, f := range fields {
							fieldsMap[f] = struct{}{}
						}

						parseHeaderRow = false
						HeaderRow = record
						checkError(writer.Write(HeaderRow))
						continue
					}
					if checkFields {
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

						checkFields = false
					}

					hit = false
					for _, f := range fields {
						target = record[f-1]
						hitOne = false

						if useRegexp {
							for _, re := range patternsMap {
								if re.MatchString(target) {
									hitOne = true
									break
								}
							}
						} else {
							k := target
							if ignoreCase {
								k = strings.ToLower(k)
							}
							if _, ok := patternsMap[k]; ok {
								hitOne = true
							}
						}

						if hitOne {
							hit = true
							break
						}
					}

					if invert {
						if hit {
							continue
						}
					} else {
						if !hit {
							continue
						}
					}

					if !noHighlight {
						record2 := make([]string, len(record)) //with color
						for i, c := range record {
							if _, ok := fieldsMap[i+1]; (!negativeFields && ok) || (negativeFields && !ok) {
								if useRegexp {
									v := ""
									for _, re := range patternsMap {
										if re.MatchString(record[i]) {
											v = re.ReplaceAllString(c, redText(re.FindAllString(c, 1)[0]))
										} else {
											v = c
										}
									}
									record2[i] = v
								} else {
									record2[i] = redText(c)
								}
							} else {
								record2[i] = c
							}
						}
						record = record2
					}
					checkError(writer.Write(record))
				}
			}
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

var redText = color.New(color.FgHiRed).SprintFunc()

func init() {
	RootCmd.AddCommand(grepCmd)
	grepCmd.Flags().StringP("fields", "f", "1", `comma separated key fields, column name or index. e.g. -f 1-3 or -f id,id2 or -F -f "group*"`)
	grepCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g. *name or id123*`)
	grepCmd.Flags().StringSliceP("pattern", "p", []string{""}, `query pattern (multiple values supported)`)
	grepCmd.Flags().StringP("pattern-file", "P", "", `pattern files (could also be CSV format)`)
	grepCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	grepCmd.Flags().BoolP("use-regexp", "r", false, `patterns are regular expression`)
	grepCmd.Flags().BoolP("invert", "v", false, `invert match`)
	grepCmd.Flags().BoolP("no-highlight", "n", false, `no highlight`)
}
