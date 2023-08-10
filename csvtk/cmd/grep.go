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
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"strconv"
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

Attentions:

  1. By default, we directly compare the column value with patterns,
     use "-r/--use-regexp" for partly matching.
  2. Multiple patterns can be given by setting '-p/--pattern' more than once,
     or giving comma separated values (CSV formats). 
     Therefore, please use double quotation marks for patterns containing
     comma, e.g., -p '"A{2,}"'

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
		printLineNumber := getFlagBool(cmd, "line-number") || config.ShowRowNumber
		deleteMatched := getFlagBool(cmd, "delete-matched")

		immediateOutput := getFlagBool(cmd, "immediate-output")

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
			noHighlight = true
			fn := func(line string) (interface{}, bool, error) {
				line = strings.TrimRight(line, "\r\n")
				if line == "" {
					return line, false, nil
				}
				return line, true, nil
			}
			reader, err := breader.NewBufferedReader(patternFile, config.NumCPUs, 1000, fn)
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

			if err != nil {
				if err == xopen.ErrNoContent {
					log.Warningf("csvtk grep: skipping empty input file: %s", file)
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr:    fieldStr,
				FuzzyFields: fuzzyFields,
			})

			var k string
			var re *regexp.Regexp
			var ok bool

			var target string
			var hitOne, hit bool
			var reHit *regexp.Regexp
			var i, j int
			var c string
			var buf bytes.Buffer
			var found []int

			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					if !config.NoHeaderRow || record.IsHeaderRow {
						if printLineNumber {
							unshift(&record.All, "row")
						}
						checkError(writer.Write(record.All))
					}
					checkFirstLine = false
					continue
				}

				if verbose && record.Row&8191 == 0 {
					log.Infof("processed records: %d", record.Row)
				}

				hit = false
				for _, target = range record.Selected {
					hitOne = false
					if useRegexp {
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
					for _, i = range record.Fields {
						i--
						c = record.All[i]

						if useRegexp {
							j = 0
							buf.Reset()

							for _, found = range reHit.FindAllStringIndex(c, -1) {
								buf.WriteString(c[j:found[0]])
								buf.WriteString(redText(c[found[0]:found[1]]))
								j = found[1]
							}
							buf.WriteString(c[j:])
							record.All[i] = buf.String()
						} else {
							record.All[i] = redText(c)
						}
					}
				}

				if printLineNumber {
					unshift(&record.All, strconv.Itoa(record.Row))
				}
				checkError(writer.Write(record.All))

				if immediateOutput {
					writer.Flush()
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
	grepCmd.Flags().StringSliceP("pattern", "p", []string{""}, `query pattern (multiple values supported). Attention: use double quotation marks for patterns containing comma, e.g., -p '"A{2,}"'`)
	grepCmd.Flags().StringP("pattern-file", "P", "", `pattern files (one pattern per line)`)
	grepCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	grepCmd.Flags().BoolP("use-regexp", "r", false, `patterns are regular expression`)
	grepCmd.Flags().BoolP("invert", "v", false, `invert match`)
	grepCmd.Flags().BoolP("no-highlight", "N", false, `no highlight`)
	grepCmd.Flags().BoolP("verbose", "", false, `verbose output`)
	grepCmd.Flags().BoolP("line-number", "n", false, `print line number as the first column ("n")`)
	grepCmd.Flags().BoolP("delete-matched", "", false, "delete a pattern right after being matched, this keeps the firstly matched data and speedups when using regular expressions")
	grepCmd.Flags().BoolP("immediate-output", "", false, "print output immediately, do not use write buffer")
}
