// Copyright © 2016-2026 Wei Shen <shenwei356@gmail.com>
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
	"strconv"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// pasteCmd represents the paste command.
var pasteCmd = &cobra.Command{
	GroupID: "set",

	Use:   "paste",
	Short: "paste CSV/TSV files by columns",
	Long: `paste CSV/TSV files by columns

Rows are combined by their position in each file. Shorter inputs are padded
with empty fields to match the longest input. Only one input file may be read
from stdin.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		stdinCount := 0
		for _, file := range files {
			if isStdin(file) {
				stdinCount++
			}
		}
		if stdinCount > 1 {
			checkError(fmt.Errorf("stdin can only be used once"))
		}

		readers := make([]*CSVReader, len(files))
		for i, file := range files {
			reader, err := newCSVReaderByConfig(config, file)
			if err != nil {
				if err == xopen.ErrNoContent {
					continue
				}
				checkError(err)
			}
			reader.Read(ReadOption{FieldStr: "1-"})
			readers[i] = reader
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := newCSVOutputWriter(outfh, csvOutputOption{QuoteAll: config.QuoteAll})
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

		records := make([]Record, len(readers))
		available := make([]bool, len(readers))
		done := make([]bool, len(readers))
		widths := make([]int, len(readers))
		for i := range widths {
			// An empty input has no CSV record from which to infer its width.
			// Like GNU paste, represent it with one empty field.
			widths[i] = 1
		}
		combined := make([]string, 0)
		row := 0
		for {
			availableCount := 0
			for i, reader := range readers {
				if reader == nil || done[i] {
					available[i] = false
					continue
				}
				record, ok := <-reader.Ch
				available[i] = ok
				if !ok {
					done[i] = true
					continue
				}
				if record.Err != nil {
					checkError(fmt.Errorf("read %s: %w", files[i], record.Err))
				}
				records[i] = record
				widths[i] = len(record.All)
				availableCount++
			}

			if availableCount == 0 {
				break
			}
			combined = combined[:0]
			if config.ShowRowNumber {
				if row == 0 && !config.NoHeaderRow {
					combined = append(combined, "row")
				} else if config.NoHeaderRow {
					combined = append(combined, strconv.Itoa(row+1))
				} else {
					combined = append(combined, strconv.Itoa(row))
				}
			}
			for i := range records {
				if available[i] {
					combined = append(combined, records[i].All...)
					continue
				}
				for range widths[i] {
					combined = append(combined, "")
				}
			}

			if !(row == 0 && !config.NoHeaderRow && config.NoOutHeader) {
				checkError(writer.Write(combined))
			}
			row++
		}

		for i, reader := range readers {
			readerReport(&config, reader, files[i])
		}
	},
}

func init() {
	RootCmd.AddCommand(pasteCmd)
}
