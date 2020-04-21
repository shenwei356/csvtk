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
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// concatCmd represents the concat command
var concatCmd = &cobra.Command{
	Use:   "concat",
	Short: "concatenate CSV/TSV files by rows",
	Long: `concatenate CSV/TSV files by rows

Note that the second and later files are concatenated to the first one,
so only columns match that of the first files kept.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		ignoreCase := getFlagBool(cmd, "ignore-case")
		keepUnmatched := getFlagBool(cmd, "keep-unmatched")
		UnmatchedRepl := getFlagString(cmd, "unmatched-repl")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		var COLNAMES []string
		var COLNAME2OLDNAME map[string]string
		var DF map[string][]string
		var col string
		var ok bool
		var j int
		var anyMatches bool
		for i, file := range files {
			colnames, colname2OldName, df := readDataFrame(config, file, ignoreCase)

			if len(df) == 0 {
				log.Warningf("no data in file: %s", file)
			}

			if i == 0 {
				COLNAMES, COLNAME2OLDNAME, DF = colnames, colname2OldName, df
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

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}

		if !config.NoHeaderRow {
			colnames := make([]string, len(COLNAMES))
			for i, col := range COLNAMES {
				colnames[i] = COLNAME2OLDNAME[col]
			}
			checkError(writer.Write(colnames))
		}

		row := make([]string, len(COLNAMES))
		for i := 0; i < len(DF[COLNAMES[0]]); i++ {
			for j, col = range COLNAMES {
				row[j] = DF[col][i]
			}

			checkError(writer.Write(row))
		}

		writer.Flush()
		checkError(writer.Error())
	},
}

func init() {
	RootCmd.AddCommand(concatCmd)

	concatCmd.Flags().BoolP("ignore-case", "i", false, `ignore case (column name)`)
	concatCmd.Flags().BoolP("keep-unmatched", "k", false, `keep blanks even if no any data of a file matches`)
	concatCmd.Flags().StringP("unmatched-repl", "u", "", "replacement for unmatched data")
}
