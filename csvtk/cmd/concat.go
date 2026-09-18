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

If there's only one input file, it will be directly outputted.

With --original-file, append an original_file column containing the basename
of each input file. Standard input is labeled "-".

If multiple input files are provided, the second and subsequent files will be
concatenated to the first one by rows. And only columns matching those of the
first file are kept.

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

		var COLNAMES []string
		var COLNAME2OLDNAME map[string]string
		var DF map[string][]string
		var sourceFiles []string
		var col string
		var ok bool
		var j int
		var anyMatches bool
		var flag int // first non-empty file
		for _, file := range files {
			colnames, colname2OldName, df, err := readDataFrame(config, file, ignoreCase)

			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk concat: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}

			if len(df) == 0 {
				if config.Verbose {
					log.Warningf("no data in file: %s", file)
				}
				continue
			}

			flag++
			if flag == 1 {
				COLNAMES, COLNAME2OLDNAME, DF = colnames, colname2OldName, df
				if originalFile {
					for _, col := range COLNAMES {
						if strings.EqualFold(COLNAME2OLDNAME[col], "original_file") {
							checkError(fmt.Errorf("input already has an original_file column: %s", file))
						}
					}
					for range DF[COLNAMES[0]] {
						sourceFiles = append(sourceFiles, filepath.Base(file))
					}
				}
				continue
			}

			anyMatches = false

			for col = range DF {
				if _, ok = df[col]; ok {
					anyMatches = true
					break
				}
			}

			if !anyMatches && !keepUnmatched {
				continue
			}
			if originalFile {
				for range df[colnames[0]] {
					sourceFiles = append(sourceFiles, filepath.Base(file))
				}
			}

			for col = range DF {
				if _, ok = df[col]; ok {
					DF[col] = append(DF[col], df[col]...)
				} else {
					for j = 0; j < len(df[colnames[0]]); j++ {
						DF[col] = append(DF[col], UnmatchedRepl)
					}
				}
			}
		}

		if len(COLNAMES) == 0 {
			if config.Verbose {
				log.Warningf("csvtk concat: no input data")
			}
			return
		}

		if !config.NoHeaderRow && !config.NoOutHeader {
			colnames := make([]string, len(COLNAMES))
			for i, col := range COLNAMES {
				colnames[i] = COLNAME2OLDNAME[col]
			}
			if printLineNumber {
				unshift(&colnames, "row")
			}
			if originalFile {
				colnames = append(colnames, "original_file")
			}

			checkError(writer.Write(colnames))
		}

		ncols := len(COLNAMES)
		nrows := len(DF[COLNAMES[0]])

		if printLineNumber {
			ncols++
		}
		if originalFile {
			ncols++
		}
		row := make([]string, ncols)

		for i := 0; i < nrows; i++ {
			j = 0
			if printLineNumber {
				row[0] = strconv.Itoa(i + 1)
				j = 1
			}
			for _, col = range COLNAMES {
				row[j] = DF[col][i]
				j++
			}
			if originalFile {
				row[j] = sourceFiles[i]
			}

			checkError(writer.Write(row))
		}

	},
}

func init() {
	RootCmd.AddCommand(concatCmd)

	concatCmd.Flags().BoolP("ignore-case", "i", false, `ignore case (column name)`)
	concatCmd.Flags().BoolP("keep-unmatched", "k", false, `keep blanks even if no any data of a file matches`)
	concatCmd.Flags().StringP("unmatched-repl", "u", "", "replacement for unmatched data")
	concatCmd.Flags().Bool("original-file", false, "append an original_file column with each input file's basename")
}
