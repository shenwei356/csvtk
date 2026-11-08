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
	"fmt"
	"runtime"
	"sort"
	"strings"
	"unicode/utf16"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

// splitXlsxCmd represents the splitXlsx command
var splitXlsxCmd = &cobra.Command{
	GroupID: "format",

	Use:   "splitxlsx",
	Short: "split an XLSX sheet into multiple sheets according to column values",
	Long: `split an XLSX sheet into multiple sheets according to column values

Strength: Sheet properties are preserved.
Limitation: Complex sheet structures are not fully supported, such as
  1. merged cells
  2. more than one header row

Groups that would have the same sheet name receive distinct numeric suffixes.
Missing cells in selected columns are treated as empty values.
Flag -U/--delete-header suppresses headers in generated sheets; the source sheet is preserved.

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
		defer func() { checkError(xlsx.Close()) }()

		sheets := xlsx.GetSheetMap()

		if listSheets {
			outfh, err := xopen.Wopen(config.OutFile)
			checkError(err)
			defer closeOutput(outfh)
			if len(sheets) > 0 {
				_, err = fmt.Fprintln(outfh, "index\tsheet")
				checkError(err)
				config.flushOutput(outfh)
				is := make([]int, len(sheets))
				i := 0
				for index := range sheets {
					is[i] = index
					i++
				}
				sort.Ints(is)
				for _, index := range is {
					_, err = fmt.Fprintf(outfh, "%d\t%s\n", index, sheets[index])
					checkError(err)
					config.flushOutput(outfh)
				}
			}
			return
		}

		if sheetName == "" {
			var ok bool
			if sheetName, ok = sheets[sheetIndex]; !ok {
				checkError(fmt.Errorf(`%dth sheet does not exist in file: %s`, sheetIndex, files[0]))
			}
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

		selection := newFieldSelection(fieldStr)
		if config.NoHeaderRow && len(selection.names) > 0 && config.Verbose {
			log.Warningf("colnames detected, flag -H (--no-header-row) ignored")
		}
		hasHeader := !config.NoHeaderRow || len(selection.names) > 0

		var key string
		var ok bool

		keysMap := make(map[string]struct{}, 10)
		keysList := make([]string, 0, 10)
		keySheetNames := make(map[string]string, 10)
		usedSheetNames := make(map[string]struct{}, len(sheets))
		for _, name := range sheets {
			usedSheetNames[name] = struct{}{}
		}
		Keys2RowIndex := make(map[string]map[int]struct{}, 10)
		rows, err := xlsx.GetRows(sheetName)
		checkError(err)
		// GetRows omits trailing blank cells, so the longest row defines the
		// available column range even when the first row is short or empty.
		nFields := 0
		for _, record := range rows {
			if len(record) > nFields {
				nFields = len(record)
			}
		}
		// Resolve against the worksheet width, including cells omitted from its header.
		header := make([]string, nFields)
		if hasHeader && len(rows) > 0 {
			copy(header, rows[0])
		}
		var fields []int
		if len(rows) > 0 {
			fields = selection.resolve(header, fuzzyFields)
		}
		// Sheet labels and grouping retain the original physical column order.
		sort.Ints(fields)
		fields = UniqInts(fields)
		if len(rows) > 0 && len(fields) == 0 {
			checkError(fmt.Errorf("no fields matched in sheet: %s", sheetName))
		}
		items := make([]string, len(fields))
		for rowIndex, record := range rows {
			if rowIndex == 0 && hasHeader {
				continue
			}

			for i, f := range fields {
				// Reset omitted cells on every row; items is reused between groups.
				if f > len(record) {
					items[i] = ""
				} else {
					items[i] = record[f-1]
				}
			}

			key = encodeFields(items, ignoreCase)

			if _, ok = keysMap[key]; !ok {
				keysList = append(keysList, key)
				keysMap[key] = struct{}{}
				label := strings.Join(items, "-")
				if ignoreCase {
					label = strings.ToLower(label)
				}
				keySheetNames[key] = uniqueSheetName(label, usedSheetNames)
			}

			if _, ok = Keys2RowIndex[key]; !ok {
				Keys2RowIndex[key] = make(map[int]struct{}, 10)
			}
			Keys2RowIndex[key][rowIndex] = struct{}{}
		}

		from, err := xlsx.GetSheetIndex(sheetName)
		checkError(err)

		for _, key := range keysList {
			name := keySheetNames[key]
			// https://github.com/qax-os/excelize/issues/1617
			to, err := xlsx.NewSheet(name)
			checkError(err)
			checkError(xlsx.CopySheet(from, to))

			for i := len(rows) - 1; i >= 0; i-- {
				if i == 0 && hasHeader && !config.NoOutHeader {
					continue
				}
				if _, ok = Keys2RowIndex[key][i]; !ok {
					checkError(xlsx.RemoveRow(name, i+1))
				}
			}
		}
		if config.Verbose {
			log.Infof("Sheet %s is split into %d sheets", sheetName, len(keysList))
		}

		xlsx.SetActiveSheet(from)
		if config.OutFile == "-" {
			prefx, _ := filepathTrimExtension(files[0])
			config.OutFile = fmt.Sprintf("%s.split.xlsx", prefx)
		}
		checkError(checkCommandOutputFiles(cmd, files, config.OutFile))
		checkError(xlsx.SaveAs(config.OutFile))
	},
}

func uniqueSheetName(label string, used map[string]struct{}) string {
	label = strings.Map(func(r rune) rune {
		if strings.ContainsRune(":\\/?*[]", r) {
			return '_'
		}
		return r
	}, label)
	label = strings.Trim(label, "'")
	if label == "" {
		label = "NA"
	}
	for n := 1; ; n++ {
		suffix := ""
		if n > 1 {
			suffix = fmt.Sprintf("-%d", n)
		}
		name := strings.Trim(truncateSheetName(label, 31-len(suffix)), "'") + suffix
		exists := false
		for previous := range used {
			if strings.EqualFold(previous, name) {
				exists = true
				break
			}
		}
		if !exists {
			used[name] = struct{}{}
			return name
		}
	}
}

func truncateSheetName(name string, limit int) string {
	units := 0
	for i, r := range name {
		n := utf16.RuneLen(r)
		if units+n > limit {
			return name[:i]
		}
		units += n
	}
	return name
}

func init() {
	RootCmd.AddCommand(splitXlsxCmd)
	splitXlsxCmd.Flags().StringP("fields", "f", "1", `comma-separated key fields, column name or index. e.g. -f 1-3 or -f id,id2 or -F -f "group*"`)
	splitXlsxCmd.Flags().BoolP("fuzzy-fields", "F", false, `use fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	splitXlsxCmd.Flags().BoolP("ignore-case", "i", false, `ignore case (cell value)`)
	splitXlsxCmd.Flags().StringP("sheet-name", "n", "", "sheet to retrieve")
	splitXlsxCmd.Flags().BoolP("list-sheets", "a", false, "list all sheets")
	splitXlsxCmd.Flags().IntP("sheet-index", "N", 1, "Nth sheet to retrieve")
}
