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
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/botond-sipos/thist"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// watchCmd represents the seq command
var watchCmd = &cobra.Command{
	GroupID: "info",

	Use:   "watch",
	Short: "monitor the specified fields",
	Long: `monitor the specified fields

Notes:
  1. --print-freq must be positive, or -1 to report after EOF.
  2. No histogram is written when there are no finite numeric values.
  3. The image path must differ from input files, the input file list, and -o.
`,

	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		runtime.GOMAXPROCS(config.NumCPUs)

		printField := getFlagString(cmd, "field")
		if printField == "" {
			checkError(fmt.Errorf("flag -f (--field) needed"))
		}
		printPdf := getFlagString(cmd, "image")
		printFreq := getFlagInt(cmd, "print-freq")
		if printFreq != -1 && printFreq <= 0 {
			checkError(fmt.Errorf("--print-freq must be positive or -1"))
		}
		if printPdf != "" {
			// Check both destinations before either can truncate a file.
			checkError(checkCommandOutputFiles(cmd, files, printPdf))
			if !isStdin(config.OutFile) {
				checkError(checkWatchOutputPaths(config.OutFile, printPdf))
			}
		}
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

		if config.Tabs {
			config.OutDelimiter = rune('\t')
		}

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

		var count int
		var p float64

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk watch: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr: printField,

				DoNotAllowDuplicatedColumnName: true,
			})

			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					checkFirstLine = false

					if len(record.Fields) > 1 {
						checkError(fmt.Errorf("only a single field allowed"))
					}

					if !config.NoHeaderRow || record.IsHeaderRow {
						if printPass {
							checkError(writer.Write(record.All))
						}
						continue
					}
				}

				p, err = strconv.ParseFloat(record.Selected[0], 64)
				if err != nil {
					continue
				}
				p = transform(p)
				if math.IsNaN(p) || math.IsInf(p, 0) {
					continue
				}

				count++
				h.Update(p)
				if printPass {
					checkError(writer.Write(record.All))
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
			}
		}

		// Empty inputs and inputs without usable numbers have no histogram.
		if h.DataCount > 0 && (printFreq < 0 || count%printFreq != 0) {
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

// checkWatchOutputPaths rejects aliases between the image and tabular outputs,
// including paths that do not exist yet and aliases through parent directories.
func checkWatchOutputPaths(output, image string) error {
	a, err := resolveWatchOutputPath(output)
	if err != nil {
		return err
	}
	b, err := resolveWatchOutputPath(image)
	if err != nil {
		return err
	}
	if a == b {
		return fmt.Errorf("image and tabular output are the same file; use different output files")
	}
	left, err := os.Stat(a)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	right, err := os.Stat(b)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if left != nil && right != nil && os.SameFile(left, right) {
		return fmt.Errorf("image and tabular output are the same file; use different output files")
	}
	return nil
}

// resolveWatchOutputPath resolves existing links while allowing a new target.
func resolveWatchOutputPath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil || !os.IsNotExist(err) {
		return resolved, err
	}
	// A dangling link may point to another new output path.
	if info, statErr := os.Lstat(path); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(path), target)
		}
		return resolveWatchOutputPath(target)
	}
	parent, err := resolveWatchOutputPath(filepath.Dir(path))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(path)), nil
}

func init() {
	RootCmd.AddCommand(watchCmd)

	watchCmd.Flags().StringP("field", "f", "", "field to watch")
	watchCmd.Flags().IntP("print-freq", "p", -1, "print/report after this many records (-1 to print after EOF)")
	watchCmd.Flags().StringP("image", "O", "", "save histogram to this PDF/image file")
	watchCmd.Flags().IntP("delay", "W", 1, "sleep this many seconds after plotting")
	watchCmd.Flags().IntP("bins", "B", -1, "number of histogram bins")
	watchCmd.Flags().BoolP("dump", "y", false, "print histogram data to stderr instead of plotting")
	watchCmd.Flags().BoolP("log", "L", false, "log10(x+1) transform numeric values")
	watchCmd.Flags().BoolP("reset", "R", false, "reset histogram after every report")
	watchCmd.Flags().BoolP("pass", "x", false, "passthrough mode (forward input to output)")
	watchCmd.Flags().BoolP("quiet", "Q", false, "suppress all plotting to stderr")
}
