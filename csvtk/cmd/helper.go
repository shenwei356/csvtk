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
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/brentp/xopen"
	"github.com/shenwei356/util/stringutil"
	"github.com/spf13/cobra"
)

func checkError(err error) {
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

func getFileList(args []string) []string {
	files := []string{}
	if len(args) == 0 {
		files = append(files, "-")
	} else {
		for _, file := range files {
			if file == "-" {
				continue
			}
			if _, err := os.Stat(file); os.IsNotExist(err) {
				checkError(err)
			}
		}
		files = args
	}
	return files
}

func getFlagInt(cmd *cobra.Command, flag string) int {
	value, err := cmd.Flags().GetInt(flag)
	checkError(err)
	return value
}

func getFlagPositiveInt(cmd *cobra.Command, flag string) int {
	value, err := cmd.Flags().GetInt(flag)
	checkError(err)
	if value <= 0 {
		checkError(fmt.Errorf("value of flag --%s should be greater than 0", flag))
	}
	return value
}

func getFlagNonNegativeInt(cmd *cobra.Command, flag string) int {
	value, err := cmd.Flags().GetInt(flag)
	checkError(err)
	if value < 0 {
		checkError(fmt.Errorf("value of flag --%s should be greater than 0", flag))
	}
	return value
}

func getFlagBool(cmd *cobra.Command, flag string) bool {
	value, err := cmd.Flags().GetBool(flag)
	checkError(err)
	return value
}

func getFlagString(cmd *cobra.Command, flag string) string {
	value, err := cmd.Flags().GetString(flag)
	checkError(err)
	return value
}

func getFlagCommaSeparatedStrings(cmd *cobra.Command, flag string) []string {
	value, err := cmd.Flags().GetString(flag)
	checkError(err)
	return stringutil.Split(value, ",")
}

func getFlagCommaSeparatedInts(cmd *cobra.Command, flag string) []int {
	filedsStrList := getFlagCommaSeparatedStrings(cmd, flag)
	fields := make([]int, len(filedsStrList))
	for i, value := range filedsStrList {
		v, err := strconv.Atoi(value)
		if err != nil || v < 1 {
			checkError(fmt.Errorf("value of flag --%s should be comma separated positive integers", flag))
		}
		fields[i] = v
	}
	return fields
}

func getFlagRune(cmd *cobra.Command, flag string) rune {
	value, err := cmd.Flags().GetString(flag)
	checkError(err)
	if len(value) > 1 {
		checkError(fmt.Errorf("value of flag --%s should has length of 1", flag))
	}
	var v rune
	for _, r := range value {
		v = r
		break
	}
	return v
}

func getFlagFloat64(cmd *cobra.Command, flag string) float64 {
	value, err := cmd.Flags().GetFloat64(flag)
	checkError(err)
	return value
}

func getFlagInt64(cmd *cobra.Command, flag string) int64 {
	value, err := cmd.Flags().GetInt64(flag)
	checkError(err)
	return value
}

func getFlagStringSlice(cmd *cobra.Command, flag string) []string {
	value, err := cmd.Flags().GetStringSlice(flag)
	checkError(err)
	return value
}

// Config is the struct containing all gloabl flags
type Config struct {
	ChunkSize int
	NumCPUs   int

	Delimiter    rune
	OutDelimiter rune
	// QuoteChar   rune
	CommentChar rune
	LazyQuotes  bool

	Tabs        bool
	OutTabs     bool
	NoHeaderRow bool

	OutFile string
}

func getConfigs(cmd *cobra.Command) Config {
	return Config{
		ChunkSize: getFlagPositiveInt(cmd, "chunk-size"),
		NumCPUs:   getFlagPositiveInt(cmd, "num-cups"),

		Delimiter:    getFlagRune(cmd, "delimiter"),
		OutDelimiter: getFlagRune(cmd, "out-delimiter"),
		// QuoteChar:   getFlagRune(cmd, "quote-char"),
		CommentChar: getFlagRune(cmd, "comment-char"),
		LazyQuotes:  getFlagBool(cmd, "lazy-quotes"),

		Tabs:        getFlagBool(cmd, "tabs"),
		OutTabs:     getFlagBool(cmd, "out-tabs"),
		NoHeaderRow: getFlagBool(cmd, "no-header-row"),

		OutFile: getFlagString(cmd, "out-file"),
	}
}

func newCSVReaderByConfig(config Config, file string) (*CSVReader, error) {
	reader, err := NewCSVReader(file, config.NumCPUs, config.ChunkSize)
	if err != nil {
		return nil, err
	}
	if config.Tabs {
		reader.Reader.Comma = '\t'
	} else {
		reader.Reader.Comma = config.Delimiter
	}
	reader.Reader.Comment = config.CommentChar
	reader.Reader.LazyQuotes = config.LazyQuotes

	return reader, nil
}

// NewCSVWriterChanByConfig returns a chanel which you can send record to write
func NewCSVWriterChanByConfig(config Config) (chan []string, error) {
	outfh, err := xopen.Wopen(config.OutFile)
	if err != nil {
		return nil, err
	}

	ch := make(chan []string, config.NumCPUs)

	writer := csv.NewWriter(outfh)
	if config.OutTabs {
		writer.Comma = '\t'
	} else {
		writer.Comma = config.OutDelimiter
	}
	go func() {
		defer outfh.Close()
		for record := range ch {
			if err := writer.Write(record); err != nil {
				log.Fatal("error writing record to csv:", err)
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			log.Fatal(err)
		}

	}()

	return ch, nil
}

var reFields = regexp.MustCompile(`([^,]+)(,[^,]+)*,?`)
var reDigitals = regexp.MustCompile(`^\d+$`)

func parseFields(cmd *cobra.Command,
	flagOfFields string,
	flagOfNoHeaderRow string) ([]int, []string, bool) {
	fieldsStr, err := cmd.Flags().GetString(flagOfFields)
	checkError(err)
	if fieldsStr == "" {
		checkError(fmt.Errorf("flag --%s needed", flagOfFields))
	}
	if !reFields.MatchString(fieldsStr) {
		checkError(fmt.Errorf("invalid value of flag %s", flagOfFields))
	}

	var fields []int
	var colnames []string
	var parseHeaderRow bool
	firstField := reFields.FindAllStringSubmatch(fieldsStr, -1)[0][1]
	if reDigitals.MatchString(firstField) {
		fields = getFlagCommaSeparatedInts(cmd, flagOfFields)
		if !getFlagBool(cmd, "no-header-row") {
			parseHeaderRow = true
		}
	} else {
		colnames = getFlagCommaSeparatedStrings(cmd, flagOfFields)
		parseHeaderRow = true
	}
	return fields, colnames, parseHeaderRow
}
