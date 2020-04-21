// Copyright © 2019 Oxford Nanopore Technologies.
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
	"math"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/bsipos/thist"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// watchCmd represents the seq command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "monitor the specified fields",
	Long:  "monitor the specified fields",

	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		printField := getFlagString(cmd, "field")
		printPdf := getFlagString(cmd, "image")
		printFreq := getFlagInt(cmd, "print-freq")
		printDump := getFlagBool(cmd, "dump")
		printLog := getFlagBool(cmd, "log")
		printQuiet := getFlagBool(cmd, "quiet")
		printDelay := getFlagInt(cmd, "delay")
		printReset := getFlagBool(cmd, "reset")
		if printDelay < 0 {
			printDelay = 0
		}
		printBins := getFlagInt(cmd, "bins")
		printPass := getFlagBool(cmd, "pass")

		if printFreq > 0 {
			config.ChunkSize = printFreq
		}

		if config.Tabs {
			config.OutDelimiter = rune('\t')
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}

		binMode := "termfit"
		if printBins > 0 {
			binMode = "fixed"
		}
		h := thist.NewHist([]float64{}, printField, binMode, printBins, true)

		transform := func(x float64) float64 { return x }
		if printLog {
			transform = func(x float64) float64 {
				return math.Log10(x + 1)
			}
		}

		field2col := make(map[string]int)
		var col int
		if config.NoHeaderRow {
			if len(printField) == 0 {
				checkError(fmt.Errorf("flag -f (--field) needed"))
			}
			pcol, err := strconv.Atoi(printField)
			if err != nil {
				checkError(fmt.Errorf("illegal field number: %s", printField))
			}
			col = pcol - 1
			if col < 0 {
				checkError(fmt.Errorf("illegal field number: %d", pcol))
			}
		}
		if printField == "" {
			checkError(fmt.Errorf("flag -f (--field) needed"))
		}

		var count int
		var i int
		var p float64

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			checkError(err)
			csvReader.Run()

			isHeaderLine := !config.NoHeaderRow
			checkField := true
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				for _, record := range chunk.Data {
					if isHeaderLine {
						for i, column := range record {
							field2col[column] = i
						}

						isHeaderLine = false
						if printPass {
							checkError(writer.Write(record))
						}
						continue
					} // header

					i = col
					if !config.NoHeaderRow {
						var ok bool
						i, ok = field2col[printField]
						if !ok {
							checkError(fmt.Errorf("invalid field specified: %s", printField))
						}
					} else if checkField {
						if i > len(record) {
							checkError(fmt.Errorf(`field (%d) out of range (%d) in file: %s`, i+1, len(record), file))
						}
						checkField = false
					}

					p, err = strconv.ParseFloat(record[i], 64)
					if err == nil {
						count++
						h.Update(transform(p))
						if printPass {
							checkError(writer.Write(record))
						}
					} else {
						continue
					}
					if printFreq > 0 && count%printFreq == 0 {
						if printDump {
							os.Stderr.Write([]byte(h.Dump()))
						} else {
							if !printQuiet {
								os.Stderr.Write([]byte(thist.ClearScreenString()))
								os.Stderr.Write([]byte(h.Draw()))
							}
							if printPdf != "" {
								h.SaveImage(printPdf)
							}
						}
						outfh.Flush()
						if printReset {
							h = thist.NewHist([]float64{}, printField, binMode, printBins, true)
						}
						time.Sleep(time.Duration(printDelay) * time.Second)
					}

				} // record
			} //chunk

		} //file
		if printFreq < 0 || count%printFreq != 0 {
			if printDump {
				os.Stderr.Write([]byte(h.Dump()))
			} else {
				if !printQuiet {
					os.Stderr.Write([]byte(thist.ClearScreenString()))
					os.Stderr.Write([]byte(h.Draw()))
				}
			}
			outfh.Flush()
			if printPdf != "" {
				h.SaveImage(printPdf)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(watchCmd)

	watchCmd.Flags().StringP("field", "f", "", "field to watch")
	watchCmd.Flags().IntP("print-freq", "p", -1, "print/report after this many records (-1 for print after EOF)")
	watchCmd.Flags().StringP("image", "O", "", "save histogram to this PDF/image file")
	watchCmd.Flags().IntP("delay", "W", 1, "sleep this many seconds after plotting")
	watchCmd.Flags().IntP("bins", "B", -1, "number of histogram bins")
	watchCmd.Flags().BoolP("dump", "y", false, "print histogram data to stderr instead of plotting")
	watchCmd.Flags().BoolP("log", "L", false, "log10(x+1) transform numeric values")
	watchCmd.Flags().BoolP("reset", "R", false, "reset histogram after every report")
	watchCmd.Flags().BoolP("pass", "x", false, "passthrough mode (forward input to output)")
	watchCmd.Flags().BoolP("quiet", "Q", false, "supress all plotting to stderr")
}
