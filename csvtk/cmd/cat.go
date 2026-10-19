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
	"bufio"
	"fmt"
	"io"
	"os"

	//"runtime"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

// concateCmd represents the concatenate command
var catCmd = &cobra.Command{
	GroupID: "misc",

	Use:   "cat",
	Short: "stream file to stdout and report progress on stderr",
	Long: `stream file to stdout and report progress on stderr

With -L, a final line without a newline is counted and streamed unchanged.
The buffer size must be positive. The print frequency must be positive or -1.
With -p -1, progress is printed only after each file has been read.`,

	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		outFile := config.OutFile
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		flagLines := getFlagBool(cmd, "lines")
		flagBuff := getFlagPositiveInt(cmd, "buffsize")
		flagFreq := getFlagInt(cmd, "print-freq")
		flagTotal := getFlagInt(cmd, "total")
		if flagFreq != -1 && flagFreq <= 0 {
			checkError(fmt.Errorf("value of flag --print-freq should be positive or -1"))
		}
		if flagLines && flagTotal < 0 {
			checkError(fmt.Errorf("flag -s (--total) is required when counting lines"))
		}

		of := os.Stdout
		if outFile != "-" {
			var err error
			of, err = os.Create(outFile)
			checkError(err)
		}
		writer := bufio.NewWriterSize(of, flagBuff)
		defer func() {
			// Flush before closing the selected output so buffered write errors
			// also produce a nonzero exit status.
			checkError(writer.Flush())
			closeOutput(of)
		}()

		for _, file := range files {
			fmt.Fprintf(os.Stderr, "Streaming file: %s\n", file)
			fh := os.Stdin
			if file != "-" {
				var err error
				fh, err = os.Open(file)
				checkError(err)
			}
			reader := bufio.NewReaderSize(fh, flagBuff)
			// Derive the byte total separately for every file; never reuse the first size.
			total := int64(flagTotal)
			if !flagLines && total < 0 {
				if file == "-" {
					checkError(fmt.Errorf("flag -s (--total) is required when reading bytes from stdin"))
				}
				inputStat, err := fh.Stat()
				checkError(err)
				total = inputStat.Size()
			}
			bar := pb.New64(total).SetWriter(os.Stderr)
			if flagFreq == -1 {
				// Static bars have no timer, so final-only mode cannot print early.
				bar.Set(pb.Static, true)
			}
			bar.Start()
			var count int
			var current int64 // Actual lines/bytes read, independent of progress refresh frequency.

			if flagLines {
				var line []byte
				var err error
				for {
					line, err = reader.ReadBytes('\n')
					if err != nil && err != io.EOF {
						checkError(err)
					}
					// ReadBytes can return an unterminated final line together with EOF.
					if len(line) > 0 {
						count++
						current++
						if flagFreq > 0 && count%flagFreq == 0 {
							bar.SetCurrent(current)
						}
						_, writeErr := writer.Write(line)
						checkError(writeErr)
						config.flushOutput(writer)
					}
					if err == io.EOF {
						break
					}
				}

			} else {
				var err error
				byteBuff := make([]byte, flagBuff)
				var readSize int
				for {
					readSize, err = reader.Read(byteBuff)
					if err != nil && err != io.EOF {
						checkError(err)
					}
					if readSize > 0 {
						count++
						current += int64(readSize)
						if flagFreq > 0 && count%flagFreq == 0 {
							bar.SetCurrent(current)
						}
						_, writeErr := writer.Write(byteBuff[:readSize])
						checkError(writeErr)
						config.flushOutput(writer)
					}
					if err == io.EOF {
						break
					}
				}

			}

			// Include any final partial batch even when the input is shorter than -p.
			bar.SetCurrent(current).Finish()
			if flagFreq == -1 {
				bar.Write()
			}
			checkError(fh.Close())

		}

	},
}

func init() {
	RootCmd.AddCommand(catCmd)
	catCmd.Flags().IntP("print-freq", "p", 1, "progress update frequency in lines/read chunks (positive, or -1 for final progress only)")
	catCmd.Flags().IntP("buffsize", "b", 4096*2, "buffer size (positive)")
	catCmd.Flags().BoolP("lines", "L", false, "count lines instead of bytes")
	catCmd.Flags().IntP("total", "s", -1, "expected total bytes/lines")
}
