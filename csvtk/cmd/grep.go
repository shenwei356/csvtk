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
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/shenwei356/breader"
	"github.com/shenwei356/xopen"
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
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}
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
		verbose := getFlagBool(cmd, "verbose")
		noHighlight := getFlagBool(cmd, "no-highlight")
		printLineNumber := getFlagBool(cmd, "line-number")
		deleteMatched := getFlagBool(cmd, "delete-matched")

		patternsMap := make(map[string]*regexp.Regexp)
		var outAll bool
		for _, pattern := range patterns {
			if useRegexp {
				p := pattern
				if !outAll && (p == "." || p == ".*") {
					outAll = true
				}
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
			noHighlight = true
			fn := func(line string) (interface{}, bool, error) {
				line = strings.TrimRight(line, "\r\n")
				if line == "" {
					return line, false, nil
				}
				return line, true, nil
			}
			reader, err := breader.NewBufferedReader(patternFile, config.NumCPUs, config.ChunkSize, fn)
			checkError(err)
			for chunk := range reader.Ch {
				checkError(chunk.Err)
				for _, data := range chunk.Data {
					pattern := data.(string)
					if useRegexp {
						p := pattern
						if !outAll && (p == "." || p == ".*") {
							outAll = true
						}
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
		var writer *csv.Writer
		var outfhStd io.Writer
		var outfhFile *xopen.Writer
		var err error
		isstdin := isStdin(config.OutFile)
		if isstdin {
			outfhStd = colorable.NewColorableStdout()
			writer = csv.NewWriter(outfhStd)
		} else {
			noHighlight = true
			outfhFile, err = xopen.Wopen(config.OutFile)
			checkError(err)
			defer outfhFile.Close()
			writer = csv.NewWriter(outfhFile)
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

			checkFields := true
			var N int64
			var recordWithN []string

			var k string
			var re *regexp.Regexp
			var ok bool
			printMetaLine := true
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				if printMetaLine && len(csvReader.MetaLine) > 0 {
					if isstdin {
						outfhStd.Write([]byte(fmt.Sprintf("sep=%s\n", string(writer.Comma))))
					} else {
						outfhFile.WriteString(fmt.Sprintf("sep=%s\n", string(writer.Comma)))
					}
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
									for _, re = range colnamesMap {
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

						checkFields = false
					}

					if verbose && N%1000000 == 0 {
						log.Infof("processed records: %d", N)
					}

					var target string
					var hitOne, hit bool
					var reHit *regexp.Regexp
					for _, f := range fields {
						target = record[f-1]
						hitOne = false
						if useRegexp {
							if outAll {
								hitOne = true
							} else {
								for k, re = range patternsMap {
									if re.MatchString(target) {
										hitOne = true
										reHit = re
										if deleteMatched && !invert {
											delete(patternsMap, k)
										}
										break
									}
								}
							}
						} else {
							k = target
							if ignoreCase {
								k = strings.ToLower(k)
							}
							if _, ok = patternsMap[k]; ok {
								hitOne = true
								if deleteMatched && !invert {
									delete(patternsMap, k)
								}
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
					if !noHighlight && hitOne {
						var j int
						var buf bytes.Buffer
						record2 := make([]string, len(record)) //with color
						for i, c := range record {
							if len(c) == 0 {
								record2[i] = c
								continue
							}
							if _, ok := fieldsMap[i+1]; (!negativeFields && ok) || (negativeFields && !ok) {
								if useRegexp {
									j = 0
									if outAll {
										record2[i] = redText(c)
									} else {
										buf.Reset()
										for _, f := range reHit.FindAllStringIndex(c, -1) {
											buf.WriteString(c[j:f[0]])
											buf.WriteString(redText(c[f[0]:f[1]]))
											j = f[1]
										}
										buf.WriteString(c[j:len(c)])
										record2[i] = buf.String()
									}
								} else {
									record2[i] = redText(c)
								}
							} else {
								record2[i] = c
							}
						}
						record = record2
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

var redText = color.New(color.FgHiRed).SprintFunc()

func init() {
	RootCmd.AddCommand(grepCmd)
	grepCmd.Flags().StringP("fields", "f", "1", `comma separated key fields, column name or index. e.g. -f 1-3 or -f id,id2 or -F -f "group*"`)
	grepCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	grepCmd.Flags().StringSliceP("pattern", "p", []string{""}, `query pattern (multiple values supported)`)
	grepCmd.Flags().StringP("pattern-file", "P", "", `pattern files (one pattern per line)`)
	grepCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	grepCmd.Flags().BoolP("use-regexp", "r", false, `patterns are regular expression`)
	grepCmd.Flags().BoolP("invert", "v", false, `invert match`)
	grepCmd.Flags().BoolP("no-highlight", "N", false, `no highlight`)
	grepCmd.Flags().BoolP("verbose", "", false, `verbose output`)
	grepCmd.Flags().BoolP("line-number", "n", false, `print line number as the first column ("n")`)
	grepCmd.Flags().BoolP("delete-matched", "", false, "delete a pattern right after being matched, this keeps the firstly matched data and speedups when using regular expressions")
}
