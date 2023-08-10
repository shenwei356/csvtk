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
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/shenwei356/xopen"
)

// Record is a CSV/TSV record
type Record struct {
	Line int // line number, if the original file contains blank lines, the number would be inaccurate.
	Row  int // the row number, header row skipped
	Err  error

	All      []string
	Fields   []int    // selected fields
	Selected []string // selected columns
}

// CSVReader is
type CSVReader struct {
	file string
	fh   *xopen.Reader

	NoHeaderRow   bool
	ShowRowNumber bool

	Reader *csv.Reader

	Ch chan Record

	IgnoreEmptyRow   bool
	IgnoreIllegalRow bool
	NumEmptyRows     []int // rows of emtpy rows
	NumIllegalRows   []int // rows of illegal rows

}

// NewCSVReader is
func NewCSVReader(file string) (*CSVReader, error) {
	fh, err := xopen.Ropen(file)
	if err != nil {
		// if err == xopen.ErrNoContent {
		// 	return nil, fmt.Errorf("empty file: %s", file)
		// }

		return nil, err
	}

	reader := csv.NewReader(fh)

	ch := make(chan Record, 128)

	csvReader := &CSVReader{
		file:           file,
		fh:             fh,
		Reader:         reader,
		Ch:             ch,
		NumEmptyRows:   make([]int, 0, 128),
		NumIllegalRows: make([]int, 0, 128),
	}
	return csvReader, nil
}

type ReadOption struct {
	FieldStr                       string
	FieldStrSep                    string
	FuzzyFields                    bool
	IgnoreFieldCase                bool
	DoNotAllowDuplicatedColumnName bool
	UniqColumn                     bool // deduplicate columns matched by multiple fuzzy column names
	AllowMissingColumn             bool // allow missing column
	BlankMissingColumn             bool
	ShowRowNumber                  bool
}

// Run begins to read
func (csvReader *CSVReader) Read(opt ReadOption) {
	go func() {
		fieldStr := opt.FieldStr
		if fieldStr == "" {
			fieldStr = "1-"
		}
		fieldStrSep := opt.FieldStrSep
		if fieldStrSep == "" {
			fieldStrSep = ","
		}
		fuzzyFields := opt.FuzzyFields
		ignoreFieldCase := opt.IgnoreFieldCase
		uniqColumn := opt.UniqColumn
		allowMissingColumn := opt.AllowMissingColumn
		blankMissingColumn := opt.BlankMissingColumn
		showRowNumber := opt.ShowRowNumber
		doNotAllowDuplicatedColumnName := opt.DoNotAllowDuplicatedColumnName

		defer func() {
			csvReader.fh.Close()
		}()

		fields, colnames, negativeFields, needParseHeaderRow, x2ends := parseFields(fieldStr, ",", csvReader.NoHeaderRow)
		var fieldsMap map[int]struct{}

		if len(fields) > 0 && negativeFields {
			fieldsMap = make(map[int]struct{}, len(fields))
			for _, f := range fields {
				fieldsMap[f*-1] = struct{}{}
			}
		}

		parseHeaderRow := needParseHeaderRow  // parsing header row
		parseHeaderRow2 := needParseHeaderRow // parsing header row
		handleHeaderRow := len(colnames) > 0
		var colnames2fileds map[string][]int // column name -> []field
		var colnamesMap map[string]*regexp.Regexp

		var ignoreFields []bool // only used for allowMissingColumn
		checkFields := true
		var items []string
		var noRecord bool
		var i, f int
		var col string
		var ok bool
		var re *regexp.Regexp

		var notBlank bool
		var data string
		var lineNum, row int
		ignoreIllegalRow := csvReader.IgnoreIllegalRow
		ignoreEmptyRow := csvReader.IgnoreEmptyRow

		var record []string
		var err error

		for {
			record, err = csvReader.Reader.Read()
			if err == io.EOF {
				break
			}

			lineNum++
			if err != nil {
				if ignoreIllegalRow {
					csvReader.NumIllegalRows = append(csvReader.NumIllegalRows, lineNum)
					continue
				}
				csvReader.Ch <- Record{
					Line: lineNum,
					Err:  err,
				}
			}

			if record == nil {
				continue
			}

			if ignoreEmptyRow {
				notBlank = false
				for _, data = range record {
					if data != "" {
						notBlank = true
						break
					}
				}
				if !notBlank {
					csvReader.NumEmptyRows = append(csvReader.NumEmptyRows, lineNum)
					continue
				}
			}

			// ------------------------------------------------------------------

			if parseHeaderRow { // parsing header row
				if len(fields) == 0 { // user gives the colnames
					// colnames
					colnames2fileds = make(map[string][]int, len(record))
					for i, col = range record {
						if _, ok = colnames2fileds[col]; !ok {
							colnames2fileds[col] = []int{i + 1}
						} else {
							colnames2fileds[col] = append(colnames2fileds[col], i+1)
						}
					}

					// colnames from user
					colnamesMap = make(map[string]*regexp.Regexp, len(colnames))
					for _, col = range colnames {
						if ignoreFieldCase {
							col = strings.ToLower(col)
						}
						if !fuzzyFields {
							if negativeFields {
								if _, ok = colnames2fileds[col[1:]]; !ok {
									if !allowMissingColumn {
										checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col[1:], csvReader.file))
									}
								} else if doNotAllowDuplicatedColumnName && len(colnames2fileds[col]) > 1 {
									checkError(fmt.Errorf("the selected colname is duplicated in the input data: %s", col))
								}
							} else {
								if _, ok = colnames2fileds[col]; !ok {
									if !allowMissingColumn {
										checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col, csvReader.file))
									}
								} else if doNotAllowDuplicatedColumnName && len(colnames2fileds[col]) > 1 {
									checkError(fmt.Errorf("the selected colname is duplicated in the input data: %s", col))
								}
							}
						}
						if negativeFields {
							colnamesMap[col[1:]] = fuzzyField2Regexp(col[1:])
						} else {
							colnamesMap[col] = fuzzyField2Regexp(col)
						}
					}

					// matching colnames
					if negativeFields {
						for _, col = range record {
							if ignoreFieldCase {
								col = strings.ToLower(col)
							}

							ok = false
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

							if !ok {
								fields = append(fields, colnames2fileds[col]...)
							}
						}
					} else {
						if fuzzyFields {
							var flags map[int]interface{}
							if uniqColumn {
								flags = make(map[int]interface{}, len(record))
							}
							for _, col = range colnames {
								if ignoreFieldCase {
									col = strings.ToLower(col)
								}

								for _, col2 := range record {
									if ignoreFieldCase {
										col2 = strings.ToLower(col2)
									}

									if colnamesMap[col].MatchString(col2) {
										for _, i = range colnames2fileds[col2] {
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
							}
						} else {
							for _, col = range colnames {
								if ignoreFieldCase {
									col = strings.ToLower(col)
								}
								fields = append(fields, colnames2fileds[col]...)
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
					for i, f = range fields {
						if f > len(record) {
							if !allowMissingColumn {
								checkError(fmt.Errorf(`field (%d) out of range (%d) in file: %s`, f, len(record), csvReader.file))
							} else {
								ignoreFields[i] = true
							}
						}
					}

					if negativeFields {
						for _, f = range fields { // update fieldsMap
							fieldsMap[f*-1] = struct{}{}
						}

						fields2 := make([]int, 0, len(fields))
						var ok bool
						for i = range record {
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

				checkFields = false
				parseHeaderRow = false
			} else {
				row++
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
				for i, f = range fields {
					if f > len(record) {
						if !allowMissingColumn {
							checkError(fmt.Errorf(`field (%d) out of range (%d) in file: %s`, f, len(record), csvReader.file))
						} else {
							ignoreFields[i] = true
						}
					}
				}

				if negativeFields {
					for _, f = range fields { // update fieldsMap
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

				checkFields = false
			}

			items = make([]string, 0, len(fields))

			if showRowNumber {
				if parseHeaderRow2 {
					items = append(items, "row")
					parseHeaderRow2 = false
				} else {
					items = append(items, strconv.Itoa(row))
				}
			}

			if allowMissingColumn {
				for i, f = range fields {
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

				csvReader.Ch <- Record{
					Line:     lineNum,
					Row:      row,
					All:      record, // copied values
					Fields:   fields, // the first variable
					Selected: items,  // copied values
				}

				continue
			}

			if noRecord {
				break
			}

			for _, f = range fields {
				items = append(items, record[f-1])
			}

			csvReader.Ch <- Record{
				Line:     lineNum,
				Row:      row,
				All:      record, // copied values
				Fields:   fields, // the first variable
				Selected: items,  // copied values
			}
		}

		close(csvReader.Ch)
	}()
}

func parseFields(
	fieldsStr string,
	fieldsStrSep string,
	noHeaderRow bool,
) (
	[]int, // fields
	[]string, // colnames
	bool, // negativeFields
	bool, // parseHeaderRow
	map[int]int, // x2ends
) {
	var fields []int
	var colnames []string
	var parseHeaderRow bool
	var negativeFields bool
	var x2ends map[int]int // [2]int{index of x in fields, x}
	firstField := reFields.FindAllStringSubmatch(strings.Split(fieldsStr, fieldsStrSep)[0], -1)[0][1]
	if reIntegers.MatchString(firstField) {
		fields = []int{}
		fieldsStrs := strings.Split(fieldsStr, fieldsStrSep)
		var j int
		for _, s := range fieldsStrs {
			found := reIntegerRange.FindAllStringSubmatch(s, -1)
			if len(found) > 0 { // field range
				start, err := strconv.Atoi(found[0][1])
				if err != nil {
					checkError(fmt.Errorf("fail to parse field range: %s. it should be an integer", found[0][1]))
				}

				if found[0][2] == "" {
					fields = append(fields, start)
					if x2ends == nil {
						x2ends = make(map[int]int, 8)
					}
					x2ends[j] = start
					continue
				}

				end, err := strconv.Atoi(found[0][2])
				if err != nil {
					checkError(fmt.Errorf("fail to parse field range: %s. it should be an integer", found[0][2]))
				}
				if start == 0 || end == 0 {
					checkError(fmt.Errorf("no 0 allowed in field range: %s", s))
				}

				if start < 0 && end < 0 {
					if start < end {
						for i := start; i <= end; i++ {
							fields = append(fields, i)
							j++
						}
					} else {
						for i := end; i <= start; i++ {
							fields = append(fields, i)
							j++
						}
					}
				} else if start > 0 && end > 0 {
					if start >= end {
						checkError(fmt.Errorf("invalid field range: %s. start (%d) should be less than end (%d)", s, start, end))
					}
					for i := start; i <= end; i++ {
						fields = append(fields, i)
						j++
					}
				} else {
					checkError(fmt.Errorf("invalid field range: %s. start (%d) and end (%d) should be both > 0 or < 0", s, start, end))
				}
			} else {
				field, err := strconv.Atoi(s)
				if err != nil {
					checkError(fmt.Errorf("failed to parse %s as a field number, you may mix the use of field numbers and column names", s))
				}
				fields = append(fields, field)
				j++
			}
		}

		for _, f := range fields {
			if f == 0 {
				checkError(fmt.Errorf(`field should not be 0`))
			} else if f < 0 {
				negativeFields = true
			} else {
				if negativeFields {
					checkError(fmt.Errorf(`fields should not be mixed with positive and negative fields`))
				}
			}
		}
		// 2 pass check
		if negativeFields {
			for _, f := range fields {
				if f > 0 {
					checkError(fmt.Errorf(`fields should not be mixed with positive and negative fields`))
				}
			}
		}

		if !noHeaderRow {
			parseHeaderRow = true
		}
	} else {
		colnames = strings.Split(fieldsStr, fieldsStrSep)
		for i, f := range colnames {
			if f == "" {
				checkError(fmt.Errorf(`%s filed should not be empty: %s`, nth(i+1), fieldsStr))
			} else if f[0] == '-' {
				negativeFields = true
			} else {
				if negativeFields {
					checkError(fmt.Errorf(`filed should not fixed with positive and negative fields`))
				}
			}
		}
		// 2 pass check
		if negativeFields {
			for _, f := range colnames {
				if f[0] != '-' {
					checkError(fmt.Errorf(`filed should not fixed with positive and negative fields`))
				}
			}
		}
		if noHeaderRow {
			log.Warningf("colnames detected, flag -H (--no-header-row) ignored")
		}
		parseHeaderRow = true
	}
	return fields, colnames, negativeFields, parseHeaderRow, x2ends
}

func fuzzyField2Regexp(field string) *regexp.Regexp {
	if strings.ContainsAny(field, "*") {
		field = strings.Replace(field, "*", ".*?", -1)
	}

	field = "^" + field + "$"
	re, err := regexp.Compile(field)
	checkError(err)
	return re
}

func readerReport(config *Config, csvReader *CSVReader, file string) {
	if csvReader == nil {
		return
	}
	if config.IgnoreEmptyRow && len(csvReader.NumEmptyRows) > 0 {
		log.Warningf("file '%s': %d empty rows ignored: %d", file, len(csvReader.NumEmptyRows), csvReader.NumEmptyRows)
	}
	if config.IgnoreIllegalRow && len(csvReader.NumIllegalRows) > 0 {
		log.Warningf("file '%s': %d illegal rows ignored: %d", file, len(csvReader.NumIllegalRows), csvReader.NumIllegalRows)
	}
}
