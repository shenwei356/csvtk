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
  6. An explicit output directory must not be the current directory, its ancestor,
     or a directory containing an input file, even with --force. Symlink aliases
     are checked. Without -o, split files are written to the current directory.
  7. Generated output files must not overwrite an input file or the input file list.

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

		// Validate before opening the input or modifying the output directory.
		outdir := "./"
		if config.OutFile != "-" {
			outdir = config.OutFile
			protectedFiles := append([]string(nil), files...)
			if list := getFlagString(cmd, "infile-list"); list != "" {
				protectedFiles = append(protectedFiles, list)
			}
			checkError(checkOutputDirectory(outdir, protectedFiles))
		}

		// Capture input identities before the reader closes stdin at EOF. Generated
		// paths must remain safe when a later batch reopens an existing output.
		protectedInputs := make(map[string]os.FileInfo)
		protectedPaths := append([]string(nil), files...)
		if list := getFlagString(cmd, "infile-list"); list != "" {
			protectedPaths = append(protectedPaths, list)
		}
		for _, path := range protectedPaths {
			var info os.FileInfo
			var err error
			if isStdin(path) {
				info, err = os.Stdin.Stat()
			} else {
				info, err = os.Stat(path)
			}
			checkError(err)
			protectedInputs[path] = info
		}

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

		if config.OutFile != "-" { // outdir
			makeOutDir(outdir, force, "-o/--outfile", true)
		}

		if outPrefix != "" || cmd.Flags().Lookup("out-prefix").Changed {
			outFilePrefix = outPrefix
		} else {
			outFilePrefix += "-"
		}

		// Check each final filename too: a prefix or an existing link can alias
		// an input even when the output directory itself is safe.
		checkOutputPath := func(path string) string {
			info, err := os.Stat(path)
			if os.IsNotExist(err) {
				return path
			}
			checkError(err)
			for input, inputInfo := range protectedInputs {
				if info.Mode().IsRegular() && os.SameFile(inputInfo, info) {
					checkError(fmt.Errorf("input %q and output %q are the same file; use a different output file", input, path))
				}
			}
			return path
		}
		chunkFile := func(chunk int) string {
			return checkOutputPath(filepath.Join(outdir, fmt.Sprintf("%s%d%s", outFilePrefix, chunk, outFileSuffix)))
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
				return checkOutputPath(filepath.Join(outdir, outFilePrefix+name+outFileSuffix))
			}
			var subdir string
			if len(name) > subdirLen {
				subdir = name[:subdirLen]
				return checkOutputPath(filepath.Join(outdir, subdir, outFilePrefix+name+outFileSuffix))
			}
			return checkOutputPath(filepath.Join(outdir, outFilePrefix+name+outFileSuffix))
		}

		var key string
		var headerRow []string
		// moreThanOneWrite := make(map[string]bool)
		rowsBuf := make(map[string][][]string, bufGroupsSize)

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

			_, buffered := rowsBuf[key]
			rowsBuf[key] = append(rowsBuf[key], row)
			// Check the first row too, so -b 1 writes every record.
			if bufRowsSize > 0 && len(rowsBuf[key]) >= bufRowsSize {
				appendRows(config, csvReader, headerRow, outfile(key), rowsBuf[key], key)
				rowsBuf[key] = nil
			}
			if !buffered {
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
	splitCmd.Flags().StringP("fields", "f", "1", `comma-separated key fields, column name or index. e.g. -f 1-3 or -f id,id2 or -F -f "group*"`)
	splitCmd.Flags().BoolP("fuzzy-fields", "F", false, `use fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	splitCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	splitCmd.Flags().BoolP("out-gzip", "G", false, `force gzipped output`)
	splitCmd.Flags().IntP("buf-rows", "b", 100000, `buffer N rows for each group before writing to a file`)
	splitCmd.Flags().IntP("buf-groups", "g", 100, `buffer N groups before writing to a file`)
	splitCmd.Flags().StringP("out-prefix", "p", "", `output file prefix, the default value is the input file's base name. use -p "" to omit the prefix`)
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

// splitByNChunks bounds both buffered data and open files while retaining
// round-robin order. Reopened compressed files contain concatenated streams.
func splitByNChunks(config Config, csvReader *CSVReader, nchunks int, outfile func(int) string) {
	const maxBufferedRows = 4096     // Bound slice overhead even for very short rows.
	const maxBufferedBytes = 4 << 20 // Flush wide records without retaining a full batch.
	const maxOpenChunks = 8          // Keep small splits streaming without unbounded file handles.
	var headerRow []string
	var outputs []*splitOutput
	if nchunks <= maxOpenChunks {
		outputs = make([]*splitOutput, nchunks)
	}
	opened := false
	openOutputs := func() {
		// Create every chunk, including those receiving no data, one at a time.
		for i := 0; i < nchunks; i++ {
			output := newSplitOutput(config, outfile(i+1), headerRow)
			if outputs != nil {
				outputs[i] = output
			} else {
				output.close()
			}
		}
		opened = true
	}
	buffer := make(map[int][][]string)
	bufferedRows, bufferedBytes := 0, 0
	flush := func() {
		for chunk, rows := range buffer {
			output := openSplitOutput(config, outfile(chunk+1), nil, true)
			for _, row := range rows {
				checkError(output.writer.Write(row))
			}
			output.close()
			delete(buffer, chunk)
		}
		bufferedRows, bufferedBytes = 0, 0
	}
	chunk := 0
	checkFirstLine := true
	for record := range csvReader.Ch {
		checkError(record.Err)
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
		if outputs != nil {
			checkError(outputs[chunk].writer.Write(record.All))
			chunk = (chunk + 1) % nchunks
			continue
		}
		buffer[chunk] = append(buffer[chunk], record.All)
		bufferedRows++
		for _, field := range record.All {
			bufferedBytes += len(field)
		}
		if config.ImmediateOutput || bufferedRows >= maxBufferedRows || bufferedBytes >= maxBufferedBytes {
			flush()
		}
		chunk = (chunk + 1) % nchunks
	}
	if !opened {
		openOutputs()
	}
	flush()
	for _, output := range outputs {
		output.close()
	}
}

// splitOutput owns the CSV and compression buffers for an open chunk.
type splitOutput struct {
	fh     *xopen.Writer
	writer *csvOutputWriter
}

// newSplitOutput creates a chunk and writes its header once.
func newSplitOutput(config Config, outfile string, headerRow []string) *splitOutput {
	return openSplitOutput(config, outfile, headerRow, false)
}

// openSplitOutput opens a new chunk or appends another batch to an existing one.
func openSplitOutput(config Config, outfile string, headerRow []string, appendMode bool) *splitOutput {
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if appendMode {
		flag = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}
	fh, err := xopen.WopenFile(outfile, flag, 0666)
	checkError(err)
	writer := newCSVOutputWriter(fh, config.csvOutputOptions())
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

// close finalizes both CSV records and the compressed stream.
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

	if !written && headerRow != nil {
		checkError(writer.Write(headerRow))
	}

	for _, row := range rows {
		checkError(writer.Write(row))
	}

}
