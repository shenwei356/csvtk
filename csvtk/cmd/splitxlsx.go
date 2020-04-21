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
	"sort"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/spf13/cobra"
)

// splitXlsxCmd represents the splitXlsx command
var splitXlsxCmd = &cobra.Command{
	Use:   "splitxlsx",
	Short: "split XLSX sheet into multiple sheets according to column values",
	Long: `split XLSX sheet into multiple sheets according to column values

Strengths: Sheet properties are remained unchanged.
Weakness : Complicated sheet structures are not well supported, e.g.,
  1. merged cells
  2. more than one header row

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		if files[0] == "-" {
			checkError(fmt.Errorf("stdin not supported for splitxlsx"))
		}

		runtime.GOMAXPROCS(config.NumCPUs)

		listSheets := getFlagBool(cmd, "list-sheets")
		sheetName := getFlagString(cmd, "sheet-name")
		sheetIndex := getFlagPositiveInt(cmd, "sheet-index")
		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		ignoreCase := getFlagBool(cmd, "ignore-case")

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}

		xlsx, err := excelize.OpenFile(files[0])
		checkError(err)

		sheets := xlsx.GetSheetMap()

		if listSheets {
			if len(sheets) > 0 {
				fmt.Println("index\tsheet")
				is := make([]int, len(sheets))
				i := 0
				for index := range sheets {
					is[i] = index
					i++
				}
				sort.Ints(is)
				for _, index := range is {
					fmt.Printf("%d\t%s\n", index, sheets[index])
				}
			}
			return
		}

		if sheetName == "" {
			sheetName = sheets[sheetIndex]
		} else {
			var existed bool
			for _, sheet := range sheets {
				if sheet == sheetName {
					existed = true
				}
			}
			if !existed {
				checkError(fmt.Errorf(`Sheet '%s' does not exist in file: %s`, sheetName, files[0]))
			}
		}

		sheetName2Index := make(map[string]int, len(sheets))
		for i, sheet := range sheets {
			sheetName2Index[sheet] = i
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

		parseHeaderRow := needParseHeaderRow // parsing header row
		printHeaderRow := needParseHeaderRow
		var colnames2fileds map[string]int // column name -> field
		var colnamesMap map[string]*regexp.Regexp

		checkFields := true
		var items []string
		var key string
		var ok bool

		keysMap := make(map[string]struct{}, 10)
		keysList := make([]string, 0, 10)
		Keys2RowIndex := make(map[string]map[int]struct{}, 10)
		rows, err := xlsx.GetRows(sheetName)
		checkError(err)
		for rowIndex, record := range rows {
			if parseHeaderRow { // parsing header row
				colnames2fileds = make(map[string]int, len(record))
				for i, col := range record {
					colnames2fileds[col] = i + 1
				}
				colnamesMap = make(map[string]*regexp.Regexp, len(colnames))
				for _, col := range colnames {
					if !fuzzyFields {
						if negativeFields {
							if _, ok = colnames2fileds[col[1:]]; !ok {
								checkError(fmt.Errorf(`column "%s" not existed in sheet: %s`, col[1:], sheetName))
							}
						} else {
							if _, ok = colnames2fileds[col]; !ok {
								checkError(fmt.Errorf(`column "%s" not existed in sheet: %s`, col, sheetName))
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
						checkError(fmt.Errorf(`field (%d) out of range (%d) in sheet: %s`, field, len(record), sheetName))
					}
				}
				fields2 := []int{}
				for f := range record {
					_, ok = fieldsMap[f+1]
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
					checkError(fmt.Errorf("no fields matched in sheet: %s", sheetName))
				}
				items = make([]string, len(fields))

				checkFields = false
			}

			for i, f := range fields {
				items[i] = record[f-1]
			}

			if printHeaderRow {
				printHeaderRow = false
				continue
			}

			key = strings.Join(items, "-")
			if ignoreCase {
				key = strings.ToLower(key)
			}

			if key == "" {
				key = "NA"
			}

			if _, ok = keysMap[key]; !ok {
				keysList = append(keysList, key)
				keysMap[key] = struct{}{}
			}

			if _, ok = Keys2RowIndex[key]; !ok {
				Keys2RowIndex[key] = make(map[int]struct{}, 10)
			}
			Keys2RowIndex[key][rowIndex] = struct{}{}
		}

		for _, key := range keysList {
			index := xlsx.NewSheet(key)
			checkError(xlsx.CopySheet(sheetName2Index[sheetName], index))

			for i := len(rows) - 1; i >= 0; i-- {
				if i == 0 && needParseHeaderRow {
					continue
				}
				if _, ok = Keys2RowIndex[key][i]; !ok {
					xlsx.RemoveRow(key, i)
				}
			}
		}

		xlsx.SetActiveSheet(sheetName2Index[sheetName])
		if config.OutFile == "-" {
			prefx, _ := filepathTrimExtension(files[0])
			config.OutFile = fmt.Sprintf("%s.split.xlsx", prefx)
		}
		checkError(xlsx.SaveAs(config.OutFile))
	},
}

func init() {
	RootCmd.AddCommand(splitXlsxCmd)
	splitXlsxCmd.Flags().StringP("fields", "f", "1", `comma separated key fields, column name or index. e.g. -f 1-3 or -f id,id2 or -F -f "group*"`)
	splitXlsxCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	splitXlsxCmd.Flags().BoolP("ignore-case", "i", false, `ignore case (cell value)`)
	splitXlsxCmd.Flags().StringP("sheet-name", "n", "", "sheet to retrieve")
	splitXlsxCmd.Flags().BoolP("list-sheets", "a", false, "list all sheets")
	splitXlsxCmd.Flags().IntP("sheet-index", "N", 1, "Nth sheet to retrieve")
}
