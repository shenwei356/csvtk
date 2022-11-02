// Copyright © 2016-2021 Wei Shen <shenwei356@gmail.com>
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

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// unfoldCmd represents the unfold command
var unfoldCmd = &cobra.Command{
	Use:   "unfold",
	Short: "unfold multiple values in cells of a field",
	Long: `unfold multiple values in cells of a field

Example:

    $ echo -ne "id,values,meta\n1,a;b,12\n2,c,23\n3,d;e;f,34\n" \
        | csvtk pretty
    id   values   meta
    1    a;b      12
    2    c        23
    3    d;e;f    34


    $ echo -ne "id,values,meta\n1,a;b,12\n2,c,23\n3,d;e;f,34\n" \
        | csvtk unfold -f values -s ";" \
        | csvtk pretty
    id   values   meta
    1    a        12
    1    b        12
    2    c        23
    3    d        34
    3    e        34
    3    f        34

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

		separater := getFlagString(cmd, "separater")
		if separater == "" {
			checkError(fmt.Errorf("flag -s (--separater) needed"))
		}

		fields, colnames, negativeFields, needParseHeaderRow, _ := parseFields(cmd, fieldStr, ",", config.NoHeaderRow)
		if negativeFields {
			checkError(fmt.Errorf("negative field not allowed"))
		}

		var fieldsMap map[int]struct{}
		if len(fields) > 0 {
			if len(fields) > 1 {
				checkError(fmt.Errorf("should no choosing more than one field"))
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
		}

		fuzzyFields := false

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

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					log.Warningf("csvtk unfold: skipping empty input file: %s", file)
					continue
				}
				checkError(err)
			}

			csvReader.Run()

			parseHeaderRow := needParseHeaderRow // parsing header row
			parseHeaderRow2 := needParseHeaderRow
			var colnames2fileds map[string][]int // column name -> []field
			var colnamesMap map[string]*regexp.Regexp

			checkFields := true

			var record2 []string // for output
			nr := 0

			printMetaLine := true
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				if printMetaLine && len(csvReader.MetaLine) > 0 {
					outfh.WriteString(fmt.Sprintf("sep=%s\n", string(writer.Comma)))
					printMetaLine = false
				}

				for _, record := range chunk.Data {
					if parseHeaderRow { // parsing header row
						colnames2fileds = make(map[string][]int, len(record))
						for i, col := range record {
							if _, ok := colnames2fileds[col]; !ok {
								colnames2fileds[col] = []int{i + 1}
							} else {
								checkError(fmt.Errorf("duplicate colnames not allowed: %s", col))
								colnames2fileds[col] = append(colnames2fileds[col], i+1)
							}
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
									fields = append(fields, colnames2fileds[col]...)
								}
							}

							if len(fields) > 1 {
								checkError(fmt.Errorf("should no choosing more than one field"))
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

						// record2 = make([]string, len(record))

						checkFields = false
					}

					if parseHeaderRow2 { // do not replace head line
						checkError(writer.Write(record))
						parseHeaderRow2 = false
						continue
					}
					nr++

					// copy(record2, record)
					record2 = record

					for f := range fieldsMap {
						for _, v := range strings.Split(record[f-1], separater) {
							record2[f-1] = v
							checkError(writer.Write(record2))
						}
					}
				}
			}

			readerReport(&config, csvReader, file)
		}
		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(unfoldCmd)

	unfoldCmd.Flags().StringP("fields", "f", "", `field to expand, only one field is allowed. type "csvtk unfold -h" for examples`)
	unfoldCmd.Flags().StringP("separater", "s", "; ", "separater for folded values")
}
