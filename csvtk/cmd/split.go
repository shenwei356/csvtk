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
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/shenwei356/util/pathutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// splitCmd represents the split command
var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "split CSV/TSV into multiple files according to column values",
	Long: `split CSV/TSV into multiple files according to column values

Note:

  1. flag -o/--out-file can specify out directory for splitted files

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
		ignoreCase := getFlagBool(cmd, "ignore-case")
		bufRowsSize := getFlagNonNegativeInt(cmd, "buf-rows")
		bufGroupsSize := getFlagNonNegativeInt(cmd, "buf-groups")
		gzipped := getFlagBool(cmd, "out-gzip")

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)
		checkError(err)
		csvReader.Run()

		var outFilePrefix, outFileSuffix string
		if isStdin(file) {
			if config.OutTabs || config.Tabs {
				outFilePrefix, outFileSuffix = "stdin", ".tsv"
			} else {
				outFilePrefix, outFileSuffix = "stdin", ".csv"
			}
		} else {
			outFilePrefix, outFileSuffix = filepathTrimExtension(file)
		}
		if gzipped &&
			!strings.HasSuffix(strings.ToLower(outFileSuffix), ".gz") {
			outFileSuffix = outFileSuffix + ".gz"
		}

		if config.OutFile != "-" { // outdir
			outdir := config.OutFile
			var existed bool
			existed, err = pathutil.DirExists(outdir)
			checkError(err)
			if !existed {
				checkError(os.MkdirAll(outdir, 0775))
			}
			outFilePrefix = filepath.Join(outdir, filepath.Base(outFilePrefix))
		}

		parseHeaderRow := needParseHeaderRow // parsing header row
		printHeaderRow := needParseHeaderRow
		var colnames2fileds map[string]int // column name -> field
		var colnamesMap map[string]*regexp.Regexp

		checkFields := true
		var items []string
		var key string
		var headerRow []string
		// moreThanOneWrite := make(map[string]bool)
		rowsBuf := make(map[string][][]string, bufGroupsSize)
		var ok bool

		printMetaLine := true
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
						if !fuzzyFields {
							if negativeFields {
								if _, ok = colnames2fileds[col[1:]]; !ok {
									checkError(fmt.Errorf(`column "%s" not existed in file: %s`, col[1:], file))
								}
							} else {
								if _, ok = colnames2fileds[col]; !ok {
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
						checkError(fmt.Errorf("no fields matched in file: %s", file))
					}
					items = make([]string, len(fields))

					checkFields = false
				}

				for i, f := range fields {
					items[i] = record[f-1]
				}

				if printHeaderRow {
					headerRow = record
					printHeaderRow = false
					continue
				}

				key = strings.Join(items, "-")
				if ignoreCase {
					key = strings.ToLower(key)
				}

				row := make([]string, len(record))
				copy(row, record)

				if _, ok = rowsBuf[key]; ok {
					rowsBuf[key] = append(rowsBuf[key], row)
					if len(rowsBuf[key]) == bufRowsSize {
						appendRows(config,
							printMetaLine,
							csvReader,
							headerRow,
							fmt.Sprintf("%s-%s%s", outFilePrefix, key, outFileSuffix),
							rowsBuf[key],
							key,
						)
						rowsBuf[key] = make([][]string, 0, 1)

					}
				} else {
					rowsBuf[key] = make([][]string, 0, 1)
					rowsBuf[key] = append(rowsBuf[key], row)
					if len(rowsBuf) == bufGroupsSize { // empty the buffer
						var wg sync.WaitGroup
						tokens := make(chan int, config.NumCPUs)
						for key, rows := range rowsBuf {
							if len(rows) == 0 {
								continue
							}
							wg.Add(1)
							tokens <- 1
							go func(key string, rows [][]string) {
								appendRows(config,
									printMetaLine,
									csvReader,
									headerRow,
									fmt.Sprintf("%s-%s%s", outFilePrefix, key, outFileSuffix),
									rows,
									key,
								)
								<-tokens
								wg.Done()
							}(key, rows)
						}

						wg.Wait()

						rowsBuf = make(map[string][][]string, bufGroupsSize)
					}
				}
			}
		}

		var wg sync.WaitGroup
		tokens := make(chan int, config.NumCPUs)
		for key, rows := range rowsBuf {
			if len(rows) == 0 {
				continue
			}
			wg.Add(1)
			tokens <- 1
			go func(key string, rows [][]string) {
				appendRows(config,
					printMetaLine,
					csvReader,
					headerRow,
					fmt.Sprintf("%s-%s%s", outFilePrefix, key, outFileSuffix),
					rows,
					key,
				)
				<-tokens
				wg.Done()
			}(key, rows)
		}

		wg.Wait()

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(splitCmd)
	splitCmd.Flags().StringP("fields", "f", "1", `comma separated key fields, column name or index. e.g. -f 1-3 or -f id,id2 or -F -f "group*"`)
	splitCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	splitCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	splitCmd.Flags().BoolP("out-gzip", "G", false, `force output gzipped file`)
	splitCmd.Flags().IntP("buf-rows", "b", 100000, `buffering N rows for every group before writing to file`)
	splitCmd.Flags().IntP("buf-groups", "g", 100, `buffering N groups before writing to file`)

}

var writtenFiles sync.Map

func appendRows(config Config,
	printMetaLine bool,
	csvReader *CSVReader,
	headerRow []string,
	outFile string,
	rows [][]string,
	key string,
) {

	var outfh *xopen.Writer
	var err error

	_, written := writtenFiles.Load(key)
	if written {
		outfh, err = xopen.WopenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		outfh, err = xopen.Wopen(outFile)
		writtenFiles.Store(key, true)
	}
	checkError(err)
	defer outfh.Close()

	writer := csv.NewWriter(outfh)
	if config.OutTabs || config.Tabs {
		writer.Comma = '\t'
	} else {
		writer.Comma = config.OutDelimiter
	}

	if printMetaLine && len(csvReader.MetaLine) > 0 {
		outfh.WriteString(fmt.Sprintf("sep=%s\n", string(writer.Comma)))
	}
	if !written && headerRow != nil {
		checkError(writer.Write(headerRow))
	}

	for _, row := range rows {
		checkError(writer.Write(row))
	}

	writer.Flush()
	checkError(writer.Error())
}
