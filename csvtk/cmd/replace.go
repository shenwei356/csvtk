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
	"strconv"
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// replaceCmd represents the replace command
var replaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "replace data of selected fields by regular expression",
	Long: `replace data of selected fields by regular expression

Note that the replacement supports capture variables.
e.g. $1 represents the text of the first submatch.
ATTENTION: use SINGLE quote NOT double quotes in *nix OS.

Examples: Adding space to all bases.

    csvtk replace -p "(.)" -r '$1 ' -s

Or use the \ escape character.

    csvtk replace -p "(.)" -r "\$1 " -s

more on: http://shenwei356.github.io/csvtk/usage/#replace

Special repalcement symbols:

	{nr}	Record number, starting from 1
	{kv}	Corresponding value of the key ($1) by key-value file


`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

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

		kvFile := getFlagString(cmd, "kv-file")
		keepKey := getFlagBool(cmd, "keep-key")

		var replaceWithNR bool
		if reNR.MatchString(replacement) {
			replaceWithNR = true
		}

		var replaceWithKV bool
		var kvs map[string]string
		if reKV.MatchString(replacement) {
			replaceWithKV = true
			if !regexp.MustCompile(`\(.+\)`).MatchString(pattern) {
				checkError(fmt.Errorf(`value of -p (--pattern) must contains "(" and ")" to capture data which is used specify the KEY`))
			}
			if kvFile == "" {
				checkError(fmt.Errorf(`since repalcement symbol "{kv}"/"{KV}" found in value of flag -r (--replacement), tab-delimited key-value file should be given by flag -k (--kv-file)`))
			}
			log.Infof("read key-value file: %s", kvFile)
			var err error
			kvs, err = readKVs(kvFile)
			if err != nil {
				checkError(fmt.Errorf("read key-value file: %s", err))
			}
			if len(kvs) == 0 {
				checkError(fmt.Errorf("no valid data in key-value file: %s", kvFile))
			}

			if ignoreCase {
				kvs2 := make(map[string]string, len(kvs))
				for k, v := range kvs {
					kvs2[strings.ToLower(k)] = v
				}
				kvs = kvs2
			}

			log.Infof("%d pairs of key-value loaded", len(kvs))
		}

		fieldStr := getFlagString(cmd, "fields")
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
			parseHeaderRow2 := needParseHeaderRow
			var colnames2fileds map[string]int // column name -> field
			var colnamesMap map[string]*regexp.Regexp

			checkFields := true

			var record2 []string // for output
			var r string
			var found []string
			var k string
			nr := 0
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

						checkFields = false
					}

					if parseHeaderRow2 { // do not replace head line
						checkError(writer.Write(record))
						parseHeaderRow2 = false
						continue
					}
					nr++
					for f := range record {
						record2[f] = record[f]
						if _, ok := fieldsMap[f+1]; ok {

							r = replacement

							if replaceWithNR {
								r = reNR.ReplaceAllString(r, strconv.Itoa(nr))
							}

							if replaceWithKV {
								found = patternRegexp.FindStringSubmatch(record2[f])
								if len(found) > 0 {
									k = string(found[1])
									if ignoreCase {
										k = strings.ToLower(k)
									}
									if _, ok = kvs[k]; ok {
										r = reKV.ReplaceAllString(r, kvs[k])
									} else if keepKey {
										r = reKV.ReplaceAllString(r, found[1])
									} else {
										r = reKV.ReplaceAllString(r, "")
									}
								}
							}

							record2[f] = patternRegexp.ReplaceAllString(record2[f], r)
						}
					}
					checkError(writer.Write(record2))
				}
			}
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(replaceCmd)
	replaceCmd.Flags().StringP("fields", "f", "1", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
	replaceCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fileds, e.g. *name or id123*`)
	replaceCmd.Flags().StringP("pattern", "p", "", "search regular expression")
	replaceCmd.Flags().StringP("replacement", "r", "",
		"replacement. supporting capture variables. "+
			" e.g. $1 represents the text of the first submatch. "+
			"ATTENTION: use SINGLE quote NOT double quotes in *nix OS or "+
			"use the \\ escape character.")
	replaceCmd.Flags().BoolP("ignore-case", "i", false, "ignore case")
	replaceCmd.Flags().StringP("kv-file", "k", "",
		`tab-delimited key-value file for replacing key with value when using "{kv}" in -r (--replacement)`)
	replaceCmd.Flags().BoolP("keep-key", "K", false, "keep the key as value when no value found for the key")
}

var reNR = regexp.MustCompile(`\{(NR|nr)\}`)
var reKV = regexp.MustCompile(`\{(KV|kv)\}`)
