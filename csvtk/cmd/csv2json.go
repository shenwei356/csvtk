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
	"fmt"
	"regexp"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// csv2jsonCmd represents the uniq command
var csv2jsonCmd = &cobra.Command{
	Use:   "csv2json",
	Short: "convert CSV to JSON format",
	Long: `convert CSV to JSON format

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}

		runtime.GOMAXPROCS(config.NumCPUs)

		indent := getFlagString(cmd, "indent")
		hasIndent := indent != ""
		var LF, SEP string
		if hasIndent {
			LF = "\n"
			SEP = " "
		}

		fieldStr := getFlagString(cmd, "key")
		var fields []int
		var colnames []string
		var negativeFields, needParseHeaderRow bool
		var fieldsMap map[int]struct{}

		keyed := fieldStr != ""
		var parseHeaderRow bool
		if keyed {
			fields, colnames, negativeFields, needParseHeaderRow = parseFields(cmd, fieldStr, config.NoHeaderRow)

			if len(fields) > 0 {
				if len(fields) > 1 {
					checkError(fmt.Errorf("invalid value of flag -k/--key: only ONE field allowed"))
				}
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
			} else {
				if len(colnames) > 1 {
					checkError(fmt.Errorf("invalid value of flag -k/--key: only ONE field allowed"))
				}
			}
			if negativeFields {
				checkError(fmt.Errorf("invalid value of flag -k/--key: negative field not allowed"))
			}
			parseHeaderRow = needParseHeaderRow // parsing header row
		} else if !config.NoHeaderRow {
			parseHeaderRow = true
		}

		fuzzyFields := false

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)
		checkError(err)
		csvReader.Run()

		var colnames2fileds map[string]int // column name -> field
		var colnamesMap map[string]*regexp.Regexp
		var HeaderRow []string

		var checkFields bool
		if keyed {
			checkFields = true
		}
		var items []string
		var key string

		if keyed {
			outfh.WriteString("{")
		} else {
			outfh.WriteString("[")
		}
		outfh.WriteString(LF)

		keysMaps := make(map[string]struct{}, 1000)
		var i, f int
		var col string
		first := true
		line := 0
		for chunk := range csvReader.Ch {
			checkError(chunk.Err)

			for _, record := range chunk.Data {
				line++
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
					HeaderRow = record
					continue
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
					items = make([]string, len(fields))

					checkFields = false
				}

				for i, f = range fields {
					items[i] = record[f-1]
				}

				if keyed {
					key = items[0]
					if _, ok := keysMaps[key]; ok {
						log.Warningf("ignore record with duplicated key (%s) at line %d", key, line)
						continue
					}
					keysMaps[key] = struct{}{}
				}

				if first {
					first = false
				} else {
					outfh.WriteString("," + LF)
				}

				if !config.NoHeaderRow {
					if keyed {
						outfh.WriteString(indent + `"` + key + `":` + SEP + `{` + LF)
					} else {
						outfh.WriteString(indent + `{` + LF)
					}
					for i, col = range HeaderRow {
						if i < len(record)-1 {
							outfh.WriteString(indent + indent + `"` + unescapeJSONField(col) + `":` + SEP + `"` + unescapeJSONField(record[i]) + `"` + "," + LF)
						} else {
							outfh.WriteString(indent + indent + `"` + unescapeJSONField(col) + `":` + SEP + `"` + unescapeJSONField(record[i]) + `"` + LF)
						}
					}
					outfh.WriteString(indent + "}")
				} else {
					if keyed {
						outfh.WriteString(indent + `"` + key + `":` + SEP + `[` + LF)
					} else {
						outfh.WriteString(indent + `[` + LF)
					}
					for i, col = range record {
						if i < len(record)-1 {
							outfh.WriteString(indent + indent + `"` + unescapeJSONField(col) + `"` + "," + LF)
						} else {
							outfh.WriteString(indent + indent + `"` + unescapeJSONField(col) + `"` + LF)
						}
					}
					outfh.WriteString(indent + "]")
				}
			}
		}

		outfh.WriteString(LF)
		if keyed {
			outfh.WriteString("}\n")
		} else {
			outfh.WriteString("]\n")
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(csv2jsonCmd)
	csv2jsonCmd.Flags().StringP("indent", "i", "  ", `indent. if given blank, output json in one line.`)
	csv2jsonCmd.Flags().StringP("key", "k", "", "output json as an array of objects keyed by a given filed rather than as a list. e.g -k 1 or -k columnA")
}

func unescapeJSONField(s string) string {
	s2 := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '"' {
			s2 = append(s2, rune('\\'))
		}
		s2 = append(s2, r)
	}
	return string(s2)
}
