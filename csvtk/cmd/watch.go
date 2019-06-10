// Copyright © 2016 Wei Shen <shenwei356@gmail.com>
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
	"bufio"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bsipos/thist"
	"github.com/spf13/cobra"
)

// watchCmd represents the seq command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "monitor the specified fields",
	Long:  "monitor the specified fields",

	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileList(args)
		runtime.GOMAXPROCS(config.NumCPUs)

		printField := getFlagString(cmd, "field")
		printPdf := getFlagString(cmd, "image")
		printFreq := getFlagInt(cmd, "print-freq")
		printDump := getFlagBool(cmd, "dump")
		printLog := getFlagBool(cmd, "log")
		printQuiet := getFlagBool(cmd, "quiet")
		outFile := config.OutFile
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

		outw := os.Stdout
		if outFile != "-" {
			tw, err := os.Create(outFile)
			checkError(err)
			outw = tw
		}
		outfh := bufio.NewWriter(outw)

		defer outfh.Flush()
		defer outw.Close()

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
				fmt.Fprintf(os.Stderr, "No field specified!")
				os.Exit(1)
			}
			pcol, err := strconv.Atoi(printField)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Illegal field number: %s\n", printField)
				os.Exit(1)
			}
			col = pcol - 1
			if col < 0 {
				fmt.Fprintf(os.Stderr, "Illegal field number: %d!\n", pcol)
				os.Exit(1)
			}
		}
		if len(printField) == 0 {
			fmt.Fprintf(os.Stderr, "No field specified!\n")
			os.Exit(1)
		}

		var count int
		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)
			checkError(err)
			csvReader.Run()

			isHeaderLine := !config.NoHeaderRow
			for chunk := range csvReader.Ch {
				checkError(chunk.Err)

				for _, record := range chunk.Data {
					if isHeaderLine {
						for i, column := range record {
							field2col[column] = i
						}

						isHeaderLine = false
						if printPass {
							outfh.Write([]byte(strings.Join(record, string(config.OutDelimiter)) + "\n"))
						}
						continue
					} // header

					i := col
					if !config.NoHeaderRow {
						var ok bool
						i, ok = field2col[printField]
						if !ok {
							fmt.Fprintf(os.Stderr, "Invalid field specified: %s\n", printField)
							os.Exit(1)
						}
					}
					p, err := strconv.ParseFloat(record[i], 64)
					if err == nil {
						count++
						h.Update(transform(p))
						if printPass {
							outfh.Write([]byte(strings.Join(record, string(config.OutDelimiter)) + "\n"))
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
	watchCmd.Flags().IntP("delay", "w", 1, "sleep this many seconds after plotting")
	watchCmd.Flags().IntP("bins", "B", -1, "number of histogram bins")
	watchCmd.Flags().BoolP("dump", "y", false, "print histogram data to stderr instead of plotting")
	watchCmd.Flags().BoolP("log", "L", false, "log10(x+1) transform numeric values")
	watchCmd.Flags().BoolP("reset", "R", false, "reset histogram after every report")
	watchCmd.Flags().BoolP("pass", "x", false, "passthrough mode (forward input to output)")
	watchCmd.Flags().BoolP("quiet", "Q", false, "supress all plotting to stderr")
}
