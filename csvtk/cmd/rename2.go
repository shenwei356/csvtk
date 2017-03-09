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

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// rename2Cmd represents the rename2 command
var rename2Cmd = &cobra.Command{
	Use:   "rename2",
	Short: "rename column names by regular expression",
	Long: `rename column names by regular expression

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		if config.NoHeaderRow {
			checkError(fmt.Errorf("flag --H (--no-header-row) is not allowed for this command"))
		}

		pattern := getFlagString(cmd, "pattern")
		replacement := getFlagString(cmd, "replacement")
		ignoreCase := getFlagBool(cmd, "ignore-case")
		if pattern == "" {
			checkError(fmt.Errorf("flags -p (--pattern) needed"))
		}
		p := pattern
		if ignoreCase {
			p = "(?i)" + p
		}
		patternRegexp, err := regexp.Compile(p)
		checkError(err)

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}
		fields, colnames, negativeFields, needParseHeaderRow := parseFields(cmd, fieldStr, config.NoHeaderRow)
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

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
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

			var record2 []string // for output

			printMetaLine := true
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				if printMetaLine && len(csvReader.Reader.MetaLine) > 0 {
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
								if _, ok := colnames2fileds[col]; !ok {
									checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col, file))
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

						record2 = make([]string, len(record))

						for f := range record {
							record2[f] = record[f]
							if _, ok := fieldsMap[f+1]; ok {
								record2[f] = patternRegexp.ReplaceAllString(record2[f], replacement)
							}
						}
						checkError(writer.Write(record2))

						checkFields = false

						continue
					}

					checkError(writer.Write(record))
				}
			}
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(rename2Cmd)
	rename2Cmd.Flags().StringP("fields", "f", "", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
	rename2Cmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fileds, e.g. "*name" or "id123*"`)
	rename2Cmd.Flags().StringP("pattern", "p", "", "search regular expression")
	rename2Cmd.Flags().StringP("replacement", "r", "",
		"renamement. supporting capture variables. "+
			" e.g. $1 represents the text of the first submatch. "+
			"ATTENTION: use SINGLE quote NOT double quotes in *nix OS or "+
			"use the \\ escape character.")
	rename2Cmd.Flags().BoolP("ignore-case", "i", false, "ignore case")
}
