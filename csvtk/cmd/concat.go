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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// concatCmd represents the concat command
var concatCmd = &cobra.Command{
	GroupID: "set",

	Use:   "concat",
	Short: "concatenate CSV/TSV files by rows",
	Long: `concatenate CSV/TSV files by rows

If there's only one input file, it will be output directly.

With --original-file, append an original_file column containing the basename
of each input file. Standard input is labeled "-".

If multiple input files are provided, the second and subsequent files will be
concatenated to the first one by rows. Only columns matching those of the
first file are kept.
Rows are streamed after the column mapping is established. Ambiguous column
names after duplicate-name normalization are reported as errors.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		ignoreCase := getFlagBool(cmd, "ignore-case")
		keepUnmatched := getFlagBool(cmd, "keep-unmatched")
		UnmatchedRepl := getFlagString(cmd, "unmatched-repl")
		printLineNumber := config.ShowRowNumber
		originalFile := getFlagBool(cmd, "original-file")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer closeOutput(outfh)

		outOpt := config.csvOutputOptions()

		writer := newCSVOutputWriter(outfh, outOpt)
		if config.OutTabs || config.Tabs {
			if config.OutDelimiter == ',' {
				writer.Comma = '\t'
			} else {
				writer.Comma = config.OutDelimiter
			}
		} else {
			writer.Comma = config.OutDelimiter
		}
		defer func() {
			writer.Flush()
			checkError(writer.Error())
		}()

		// -----------------------------------------------------------------------------

		if len(files) == 1 {
			file := files[0]
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk concat: skipping empty input file: %s", file)
					}
					return
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr: "1-",
			})

			isHeaderLine := !config.NoHeaderRow
			i := 0
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if isHeaderLine {
					isHeaderLine = false
					if originalFile {
						for _, col := range record.All {
							if strings.EqualFold(col, "original_file") {
								checkError(fmt.Errorf("input already has an original_file column: %s", file))
							}
						}
					}
					if config.NoOutHeader {
						continue
					}
					if printLineNumber {
						unshift(&record.All, "row")
					}
					if originalFile {
						record.All = append(record.All, "original_file")
					}
				} else {
					i++
					if printLineNumber {
						unshift(&record.All, strconv.Itoa(record.Row))
					}
					if originalFile {
						record.All = append(record.All, filepath.Base(file))
					}
				}

				checkError(writer.Write(record.All))
			}

			return
		}

		// -----------------------------------------------------------------------------

		var columns []string // Output column keys, fixed by the first nonempty input.
		var row []string
		var rowNumber int
		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk concat: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}
			csvReader.Read(ReadOption{FieldStr: "1-"})

			var indexes []int // Output column to input position; -1 means unmatched.
			var keepFile bool
			firstRecord := true
			for record := range csvReader.Ch {
				checkError(record.Err)
				if firstRecord {
					firstRecord = false
					var header []string
					if !config.NoHeaderRow {
						header = record.All
					}
					names, originalNames := dataFrameColumns(config, file, header, len(record.All), ignoreCase)
					if columns == nil {
						columns = names
						outHeader := make([]string, 0, len(columns)+2)
						if printLineNumber {
							outHeader = append(outHeader, "row")
						}
						for _, col := range columns {
							name := originalNames[col]
							if originalFile && !config.NoHeaderRow && strings.EqualFold(name, "original_file") {
								checkError(fmt.Errorf("input already has an original_file column: %s", file))
							}
							outHeader = append(outHeader, name)
						}
						if originalFile {
							outHeader = append(outHeader, "original_file")
						}
						row = make([]string, len(outHeader))
						if !config.NoHeaderRow && !config.NoOutHeader {
							checkError(writer.Write(outHeader))
						}
					}

					positions := make(map[string]int, len(names))
					for i, col := range names {
						// A generated duplicate name must not silently shadow a real column.
						if _, exists := positions[col]; exists {
							checkError(fmt.Errorf("ambiguous column name after duplicate-name normalization (%s) in file: %s", col, file))
						}
						positions[col] = i
					}
					indexes = make([]int, len(columns))
					keepFile = keepUnmatched
					for i, col := range columns {
						if index, ok := positions[col]; ok {
							indexes[i] = index
							keepFile = true
						} else {
							indexes[i] = -1
						}
					}
					if !config.NoHeaderRow {
						continue
					}
				}
				// Drain skipped files too, so input errors and reader cleanup are preserved.
				if !keepFile {
					continue
				}
				rowNumber++
				offset := 0
				if printLineNumber {
					row[0] = strconv.Itoa(rowNumber)
					offset = 1
				}
				for i, index := range indexes {
					if index < 0 {
						row[offset+i] = UnmatchedRepl
					} else {
						row[offset+i] = record.All[index]
					}
				}
				if originalFile {
					row[len(row)-1] = filepath.Base(file)
				}
				checkError(writer.Write(row))
			}
			readerReport(&config, csvReader, file)
		}
		if columns == nil && config.Verbose {
			log.Warningf("csvtk concat: no input data")
		}

	},
}

func init() {
	RootCmd.AddCommand(concatCmd)

	concatCmd.Flags().BoolP("ignore-case", "i", false, `ignore case (column name)`)
	concatCmd.Flags().BoolP("keep-unmatched", "k", false, `keep blanks even if no data from a file matches`)
	concatCmd.Flags().StringP("unmatched-repl", "u", "", "replacement for unmatched data")
	concatCmd.Flags().Bool("original-file", false, "append an original_file column with each input file's basename")
}
