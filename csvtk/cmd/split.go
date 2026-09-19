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
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// splitCmd represents the split command
var splitCmd = &cobra.Command{
	GroupID: "set",

	Use:   "split",
	Short: "split CSV/TSV by column values, rows per chunk, or number of chunks",
	Long: `split CSV/TSV by column values, rows per chunk, or number of chunks

Notes:

  1. flag -o/--out-file can specify the output directory for split files.
  2. flag -s/--prefix-as-subdir can create subdirectories with prefixes of
     keys of length X, to avoid writing too many files in the output directory.
  3. Special characters in key values are percent-encoded in output file names.
     Long encoded names use a hash.
  4. flag -n/--nlines splits the input into chunks of up to N records instead of
     splitting by key values. The header row, when present, is written to every chunk.
  5. flag -c/--nchunks splits the input into N chunks in a round-robin manner.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		nlines := getFlagInt(cmd, "nlines")
		if nlines < 0 || (cmd.Flags().Lookup("nlines").Changed && nlines == 0) {
			checkError(fmt.Errorf("value of flag --nlines should be greater than 0"))
		}
		nchunks := getFlagInt(cmd, "nchunks")
		if nchunks < 0 || (cmd.Flags().Lookup("nchunks").Changed && nchunks == 0) {
			checkError(fmt.Errorf("value of flag --nchunks should be greater than 0"))
		}

		fieldStr := getFlagString(cmd, "fields")
		if nlines == 0 && nchunks == 0 && fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		ignoreCase := getFlagBool(cmd, "ignore-case")
		bufRowsSize := getFlagNonNegativeInt(cmd, "buf-rows")
		bufGroupsSize := getFlagNonNegativeInt(cmd, "buf-groups")
		gzipped := getFlagBool(cmd, "out-gzip")
		outPrefix := getFlagString(cmd, "out-prefix")
		subdirLen := getFlagNonNegativeInt(cmd, "prefix-as-subdir")
		force := getFlagBool(cmd, "force")

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)
		checkError(err)

		if nlines > 0 || nchunks > 0 {
			csvReader.Read(ReadOption{FieldStr: "1-"})
		} else {
			csvReader.Read(ReadOption{
				FieldStr:    fieldStr,
				FuzzyFields: fuzzyFields,

				DoNotAllowDuplicatedColumnName: true,
			})
		}

		var outFilePrefix, outFileSuffix string
		if isStdin(file) {
			if config.OutTabs || config.Tabs {
				outFilePrefix, outFileSuffix = "stdin", ".tsv"
			} else {
				outFilePrefix, outFileSuffix = "stdin", ".csv"
			}
		} else {
			outFilePrefix, outFileSuffix = filepathTrimExtension(filepath.Base(file))
		}
		if gzipped &&
			!strings.HasSuffix(strings.ToLower(outFileSuffix), ".gz") {
			outFileSuffix = outFileSuffix + ".gz"
		}

		outdir := "./"
		if config.OutFile != "-" { // outdir
			outdir = config.OutFile
			makeOutDir(outdir, force, "-o/--outfile", true)
		}

		if outPrefix != "" || cmd.Flags().Lookup("out-prefix").Changed {
			outFilePrefix = outPrefix
		} else {
			outFilePrefix += "-"
		}

		chunkFile := func(chunk int) string {
			return filepath.Join(outdir, fmt.Sprintf("%s%d%s", outFilePrefix, chunk, outFileSuffix))
		}
		if nlines > 0 {
			splitByNLines(config, csvReader, nlines, chunkFile)
			readerReport(&config, csvReader, file)
			return
		}
		if nchunks > 0 {
			splitByNChunks(config, csvReader, nchunks, chunkFile)
			readerReport(&config, csvReader, file)
			return
		}

		groupFilenames := make(map[string]string)
		usedFilenames := make(map[string]struct{})
		maxKeyLength := 255 - len(outFilePrefix) - len(outFileSuffix)
		outfile := func(key string) string {
			name := groupFilenames[key]
			if subdirLen == 0 {
				return filepath.Join(outdir, outFilePrefix+name+outFileSuffix)
			}
			var subdir string
			if len(name) > subdirLen {
				subdir = name[:subdirLen]
				return filepath.Join(outdir, subdir, outFilePrefix+name+outFileSuffix)
			}
			return filepath.Join(outdir, outFilePrefix+name+outFileSuffix)
		}

		var key string
		var headerRow []string
		// moreThanOneWrite := make(map[string]bool)
		rowsBuf := make(map[string][][]string, bufGroupsSize)
		var ok bool

		checkFirstLine := true
		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if checkFirstLine {
				checkFirstLine = false

				if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
					headerRow = record.All
					continue
				}
			}

			key = encodeFields(record.Selected, ignoreCase)
			if _, exists := groupFilenames[key]; !exists {
				base := encodeFilenameFields(record.Selected, ignoreCase)
				if len(base) > maxKeyLength {
					if maxKeyLength < 2 {
						checkError(fmt.Errorf("output file prefix is too long for encoded key values"))
					}
					base = fmt.Sprintf("~%x", sha256.Sum256([]byte(key)))
					if len(base) > maxKeyLength {
						base = base[:maxKeyLength]
					}
				}
				name := base
				for n := 2; ; n++ {
					if _, used := usedFilenames[strings.ToLower(name)]; !used {
						break
					}
					suffix := fmt.Sprintf("~%d", n)
					if len(suffix) >= maxKeyLength {
						checkError(fmt.Errorf("too many output file names for the available key length"))
					}
					if len(base)+len(suffix) > maxKeyLength {
						name = base[:maxKeyLength-len(suffix)] + suffix
					} else {
						name = base + suffix
					}
				}
				groupFilenames[key] = name
				usedFilenames[strings.ToLower(name)] = struct{}{}
			}

			row := make([]string, len(record.All))
			copy(row, record.All)

			if _, ok = rowsBuf[key]; ok {
				rowsBuf[key] = append(rowsBuf[key], row)
				if len(rowsBuf[key]) == bufRowsSize {
					appendRows(config,
						csvReader,
						headerRow,
						outfile(key),
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
								csvReader,
								headerRow,
								outfile(key),
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
					csvReader,
					headerRow,
					outfile(key),
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
	splitCmd.Flags().StringP("out-prefix", "p", "", `output file prefix, the default value is the input file's base name. use -p "" to disable outputting prefix`)
	splitCmd.Flags().IntP("prefix-as-subdir", "s", 0, `create subdirectories with prefixes of keys of length X, to avoid writing too many files in the output directory`)
	splitCmd.Flags().BoolP("force", "", false, `overwrite existing output directory (given by -o).`)
	splitCmd.Flags().IntP("nlines", "n", 0, `split into chunks of up to N records; incompatible with field-grouping options`)
	splitCmd.Flags().IntP("nchunks", "c", 0, `split into N chunks in a round-robin manner; incompatible with field-grouping options`)
	splitCmd.MarkFlagsMutuallyExclusive("nlines", "nchunks")
	for _, flag := range []string{"fields", "fuzzy-fields", "ignore-case", "buf-rows", "buf-groups", "prefix-as-subdir"} {
		splitCmd.MarkFlagsMutuallyExclusive("nlines", flag)
		splitCmd.MarkFlagsMutuallyExclusive("nchunks", flag)
	}
}

func splitByNLines(config Config, csvReader *CSVReader, nlines int, outfile func(int) string) {
	var headerRow []string
	chunk := 1
	rows := 0
	checkFirstLine := true
	var output *splitOutput

	closeChunk := func() {
		output.close()
		output = nil
		chunk++
		rows = 0
	}

	for record := range csvReader.Ch {
		if record.Err != nil {
			checkError(record.Err)
		}

		if checkFirstLine {
			checkFirstLine = false
			if !config.NoHeaderRow || record.IsHeaderRow {
				headerRow = append([]string(nil), record.All...)
				continue
			}
		}

		if output == nil {
			output = newSplitOutput(config, outfile(chunk), headerRow)
		}
		checkError(output.writer.Write(record.All))
		rows++
		if rows == nlines {
			closeChunk()
		}
	}
	if output != nil {
		closeChunk()
	}
}

func splitByNChunks(config Config, csvReader *CSVReader, nchunks int, outfile func(int) string) {
	var headerRow []string
	outputs := make([]*splitOutput, nchunks)
	opened := false
	openOutputs := func() {
		for i := range outputs {
			outputs[i] = newSplitOutput(config, outfile(i+1), headerRow)
		}
		opened = true
	}

	row := 0
	checkFirstLine := true
	for record := range csvReader.Ch {
		if record.Err != nil {
			checkError(record.Err)
		}

		if checkFirstLine {
			checkFirstLine = false
			if !config.NoHeaderRow || record.IsHeaderRow {
				headerRow = append([]string(nil), record.All...)
				openOutputs()
				continue
			}
		}

		if !opened {
			openOutputs()
		}
		checkError(outputs[row%nchunks].writer.Write(record.All))
		row++
	}

	if !opened {
		openOutputs()
	}
	for _, output := range outputs {
		output.close()
	}
}

type splitOutput struct {
	fh     *xopen.Writer
	writer *csvOutputWriter
}

func newSplitOutput(config Config, outfile string, headerRow []string) *splitOutput {
	fh, err := xopen.Wopen(outfile)
	checkError(err)
	writer := newCSVOutputWriter(fh, csvOutputOption{QuoteAll: config.QuoteAll})
	if config.OutTabs || config.Tabs {
		if config.OutDelimiter == ',' {
			writer.Comma = '\t'
		} else {
			writer.Comma = config.OutDelimiter
		}
	} else {
		writer.Comma = config.OutDelimiter
	}
	if headerRow != nil {
		checkError(writer.Write(headerRow))
	}
	return &splitOutput{fh: fh, writer: writer}
}

func (output *splitOutput) close() {
	output.writer.Flush()
	checkError(output.writer.Error())
	checkError(output.fh.Close())
}

var writtenFiles sync.Map

func appendRows(config Config,
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

	outOpt := csvOutputOption{QuoteAll: config.QuoteAll}


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

	if !written && headerRow != nil {
		checkError(writer.Write(headerRow))
	}

	for _, row := range rows {
		checkError(writer.Write(row))
	}

}
