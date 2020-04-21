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

	"github.com/spf13/cobra"
	"gopkg.in/cheggaaa/pb.v2"
)

// concateCmd represents the concatenate command
var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "stream file to stdout and report progress on stderr",
	Long:  "stream file to stdout and report progress on stderr",

	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		outFile := config.OutFile
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		flagLines := getFlagBool(cmd, "lines")
		flagBuff := getFlagInt(cmd, "buffsize")
		flagFreq := getFlagInt(cmd, "print-freq")
		flagTotal := getFlagInt(cmd, "total")

		of := os.Stdout
		defer of.Close()
		if outFile != "-" {
			var err error
			of, err = os.Create(outFile)
			checkError(err)
		}
		writer := bufio.NewWriterSize(of, flagBuff)
		defer writer.Flush()

		for _, file := range files {
			fmt.Fprintf(os.Stderr, "Streaming file: %s\n", file)
			fh := os.Stdin
			if file != "-" {
				var err error
				fh, err = os.Open(file)
				checkError(err)
			}
			reader := bufio.NewReaderSize(fh, flagBuff)
			var bar *pb.ProgressBar

			if flagLines {
				if flagTotal < 0 {
					fmt.Fprintf(os.Stderr, "Cannot read lines unless the of expected number of lines is specified via -s!\n")
					os.Exit(1)
				}
				bar = pb.StartNew(flagTotal)
				var line []byte
				var err error
				var count int
				for {
					line, err = reader.ReadBytes('\n')
					if err == io.EOF {
						break
					}
					checkError(err)
					count++
					if count%flagFreq == 0 {
						bar.Add(1)
					}
					_, err := writer.Write(line)
					checkError(err)
				}

			} else {
				var err error
				if flagTotal < 0 {
					if file == "-" {
						fmt.Fprintf(os.Stderr, "Cannot read from stdin unless the number of expected bytes is specified via -s!\n")
						os.Exit(1)
					}
					inputStat, err := os.Stat(file)
					checkError(err)
					flagTotal = int(inputStat.Size())

				}
				bar = pb.StartNew(flagTotal)
				byteBuff := make([]byte, flagBuff)
				var count int
				var bytesSince int
				var readSize int
				for {
					readSize, err = reader.Read(byteBuff)
					if err == io.EOF {
						break
					}
					checkError(err)
					count++
					bytesSince += readSize
					if count%flagFreq == 0 {
						bar.Add(bytesSince)
						bytesSince = 0
					}
					_, err = writer.Write(byteBuff[:readSize])
					checkError(err)
				}

			}

			bar.Finish()

			fh.Close()

		}

	},
}

func init() {
	RootCmd.AddCommand(catCmd)
	catCmd.Flags().IntP("print-freq", "p", 1, "print frequency (-1 for print after parsing)")
	catCmd.Flags().IntP("buffsize", "b", 4096*2, "buffer size")
	catCmd.Flags().BoolP("lines", "L", false, "count lines instead of bytes")
	catCmd.Flags().IntP("total", "s", -1, "expected total bytes/lines")
}
