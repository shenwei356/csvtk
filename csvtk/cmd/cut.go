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

// cutCmd represents the cut command
var cutCmd = &cobra.Command{
	Use:   "cut",
	Short: "select and arrange fields",
	Long: `select and arrange fields

Examples:

  1. Single column
     csvtk cut -f 1
     csvtk cut -f colA
  2. Multiple columns (replicates allowed)
     csvtk cut -f 1,3,2,1
     csvtk cut -f colA,colB,colA
  3. Column ranges
     csvtk cut -f 1,3-5       # 1, 3, 4, 5
     csvtk cut -f 3,5-        # 3rd col, and 5th col to the end
     csvtk cut -f 1-          # for all
     csvtk cut -f 2-,1        # move 1th col to the end
  4. Unselect
     csvtk cut -f -1,-3       # discard 1st and 3rd column
     csvtk cut -f -1--3       # discard 1st to 3rd column
     csvtk cut -f -2-         # discard 2nd and all columns on the right.
     csvtu cut -f -colA,-colB # discard colA and colB
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

		uniqColumn := getFlagBool(cmd, "uniq-column")

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		fields, colnames, negativeFields, needParseHeaderRow, x2ends := parseFields(cmd, fieldStr, ",", config.NoHeaderRow)
		var fieldsMap map[int]struct{}

		ignoreCase := getFlagBool(cmd, "ignore-case")

		allowMissingColumn := getFlagBool(cmd, "allow-missing-col")
		blankMissingColumn := getFlagBool(cmd, "blank-missing-col")

		if len(fields) > 0 && negativeFields {
			fieldsMap = make(map[int]struct{}, len(fields))
			for _, f := range fields {
				fieldsMap[f*-1] = struct{}{}
			}
		}
		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			if config.OutDelimiter == ',' { // default value, no other value given
				writer.Comma = '\t'
			} else {
				writer.Comma = config.OutDelimiter
			}
		} else {
			writer.Comma = config.OutDelimiter
		}

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)
		checkError(err)
		csvReader.Run()

		parseHeaderRow := needParseHeaderRow // parsing header row
		handleHeaderRow := len(colnames) > 0
		var colnames2fileds map[string]int // column name -> field
		var colnamesMap map[string]*regexp.Regexp

		checkFields := true
		var items []string
		var noRecord bool

		var ignoreFields []bool

		printMetaLine := true
		for chunk := range csvReader.Ch {
			checkError(chunk.Err)

			if printMetaLine && len(csvReader.MetaLine) > 0 {
				outfh.WriteString(fmt.Sprintf("sep=%s\n", string(writer.Comma)))
				printMetaLine = false
			}

			for _, record := range chunk.Data {
				if parseHeaderRow { // parsing header row
					if len(fields) == 0 { // user gives the colnames
						// colnames
						colnames2fileds = make(map[string]int, len(record))
						for i, col := range record {
							if ignoreCase {
								col = strings.ToLower(col)
							}
							colnames2fileds[col] = i + 1
						}

						// colnames from user
						colnamesMap = make(map[string]*regexp.Regexp, len(colnames))
						i := 0
						for _, col := range colnames {
							if ignoreCase {
								col = strings.ToLower(col)
							}
							if !fuzzyFields {
								if negativeFields {
									if _, ok := colnames2fileds[col[1:]]; !ok {
										if !allowMissingColumn {
											checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col[1:], file))
										}
									}
								} else {
									if _, ok := colnames2fileds[col]; !ok {
										if !allowMissingColumn {
											checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col, file))
										}
									}
								}
							}
							if negativeFields {
								colnamesMap[col[1:]] = fuzzyField2Regexp(col[1:])
							} else {
								colnamesMap[col] = fuzzyField2Regexp(col)
								i++
							}
						}

						// matching colnames
						var ok bool
						if negativeFields {
							for _, col := range record {
								ok = false
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

								if !ok {
									fields = append(fields, colnames2fileds[col])
								}
							}
						} else {
							if fuzzyFields {
								var flags map[int]interface{}
								if uniqColumn {
									flags = make(map[int]interface{}, len(record))
								}
								var i int
								for _, col := range colnames {
									for _, col2 := range record {
										if colnamesMap[col].MatchString(col2) {
											i = colnames2fileds[col2]
											if uniqColumn {
												if _, ok = flags[i]; !ok {
													fields = append(fields, i)
													flags[i] = struct{}{}
												}
											} else {
												fields = append(fields, i)
											}
										}
									}
								}
							} else {
								for _, col := range colnames {
									fields = append(fields, colnames2fileds[col])
								}
							}
						}

					} else {
						if len(x2ends) > 0 { // user use 1-.
							fields1 := make([]int, 0, len(record))

							for i, f := range fields {
								if v, ok := x2ends[i]; ok && v == f {
									if negativeFields {
										for i = -f; i <= len(record); i++ {
											fields1 = append(fields1, i*-1)
										}
									} else {
										for i = f; i <= len(record); i++ {
											fields1 = append(fields1, i)
										}
									}
								} else {
									fields1 = append(fields1, f)
								}
							}
							fields = fields1
						}

						if ignoreFields == nil {
							ignoreFields = make([]bool, len(fields))
						}
						for i, f := range fields {
							if f > len(record) {
								if !allowMissingColumn {
									checkError(fmt.Errorf(`field (%d) out of range (%d) in file: %s`, f, len(record), file))
								} else {
									ignoreFields[i] = true
								}
							}
						}

						if negativeFields {
							for _, f := range fields { // update fieldsMap
								fieldsMap[f*-1] = struct{}{}
							}

							fields2 := make([]int, 0, len(fields))
							var ok bool
							for i := range record {
								if _, ok = fieldsMap[i+1]; !ok {
									fields2 = append(fields2, i+1)
								}
							}
							fields = fields2
						}
					}

					if len(fields) == 0 {
						noRecord = true
						break
					}

					items = make([]string, len(fields))

					checkFields = false
					parseHeaderRow = false
				}

				if checkFields {
					if len(x2ends) > 0 { // user use 1-.
						fields1 := make([]int, 0, len(record))

						for i, f := range fields {
							if v, ok := x2ends[i]; ok && v == f {
								if negativeFields {
									for i = -f; i <= len(record); i++ {
										fields1 = append(fields1, i*-1)
									}
								} else {
									for i = f; i <= len(record); i++ {
										fields1 = append(fields1, i)
									}
								}
							} else {
								fields1 = append(fields1, f)
							}
						}
						fields = fields1
					}

					if ignoreFields == nil {
						ignoreFields = make([]bool, len(fields))
					}
					for i, f := range fields {
						if f > len(record) {
							if !allowMissingColumn {
								checkError(fmt.Errorf(`field (%d) out of range (%d) in file: %s`, f, len(record), file))
							} else {
								ignoreFields[i] = true
							}
						}
					}

					if negativeFields {
						for _, f := range fields { // update fieldsMap
							fieldsMap[f*-1] = struct{}{}
						}

						fields2 := make([]int, 0, len(fields))
						var ok bool
						for i := range record {
							if _, ok = fieldsMap[i+1]; !ok {
								fields2 = append(fields2, i+1)
							}
						}
						fields = fields2
					}

					if len(fields) == 0 {
						noRecord = true
						break
					}

					items = make([]string, len(fields))

					checkFields = false
				}

				if allowMissingColumn {
					items = items[:0]
					for i, f := range fields {
						if needParseHeaderRow { // using column
							if f == 0 || f > len(record) {
								if blankMissingColumn {
									if handleHeaderRow {
										items = append(items, colnames[i])
									} else {
										items = append(items, "")
									}

								}
								continue
							}
						} else {
							if ignoreFields[i] {
								if blankMissingColumn {
									items = append(items, "")
								}
								continue
							}
						}
						items = append(items, record[f-1])
					}

					if handleHeaderRow {
						handleHeaderRow = false
					}
					checkError(writer.Write(items))
					continue
				}

				for i, f := range fields {
					items[i] = record[f-1]
				}
				checkError(writer.Write(items))
			}

			if noRecord {
				break
			}
		}

		writer.Flush()
		checkError(writer.Error())

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(cutCmd)
	cutCmd.Flags().StringP("fields", "f", "", `select only these fields. type "csvtk cut -h" for examples`)
	cutCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	cutCmd.Flags().BoolP("ignore-case", "i", false, `ignore case (column name)`)
	cutCmd.Flags().BoolP("uniq-column", "u", false, `deduplicate columns matched by multiple fuzzy column names`)
	cutCmd.Flags().BoolP("allow-missing-col", "m", false, `allow missing column`)
	cutCmd.Flags().BoolP("blank-missing-col", "b", false, `blank missing column`)
}
