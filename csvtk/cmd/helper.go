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
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/shenwei356/breader"
	"github.com/shenwei356/util/stringutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

func checkError(err error) {
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

func getFileList(args []string, checkFile bool) []string {
	files := make([]string, 0, 1000)
	if len(args) == 0 {
		files = append(files, "-")
	} else {
		for _, file := range args {
			if isStdin(file) {
				continue
			}
			if !checkFile {
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

func getFileListFromFile(file string, checkFile bool) ([]string, error) {
	var fh *os.File
	var err error
	if file == "-" {
		fh = os.Stdin
	} else {
		fh, err = os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("read file list from '%s': %s", file, err)
		}
	}

	var _file string
	lists := make([]string, 0, 1000)
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		_file = scanner.Text()
		if strings.TrimSpace(_file) == "" {
			continue
		}
		if checkFile && !isStdin(_file) {
			if _, err = os.Stat(_file); os.IsNotExist(err) {
				return lists, fmt.Errorf("check file '%s': %s", _file, err)
			}
		}
		lists = append(lists, _file)
	}
	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("read file list from '%s': %s", file, err)
	}

	return lists, nil
}

func getFileListFromArgsAndFile(cmd *cobra.Command, args []string, checkFileFromArgs bool, flag string, checkFileFromFile bool) []string {
	infileList := getFlagString(cmd, flag)
	files := getFileList(args, checkFileFromArgs)
	if infileList != "" {
		_files, err := getFileListFromFile(infileList, checkFileFromFile)
		checkError(err)
		if len(_files) == 0 {
			if !getFlagBool(cmd, "quiet") {
				log.Warningf("no files found in file list: %s", infileList)
			}
			return files
		}

		if len(files) == 1 && isStdin(files[0]) {
			return _files
		}
		files = append(files, _files...)
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

func getFlagPositiveFloat64(cmd *cobra.Command, flag string) float64 {
	value, err := cmd.Flags().GetFloat64(flag)
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
		checkError(fmt.Errorf("value of flag --%s should be greater than or equal to 0", flag))
	}
	return value
}

func getFlagNonNegativeFloat64(cmd *cobra.Command, flag string) float64 {
	value, err := cmd.Flags().GetFloat64(flag)
	checkError(err)
	if value < 0 {
		checkError(fmt.Errorf("value of flag --%s should be greater than or equal to ", flag))
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

func getFlagSemicolonSeparatedStrings(cmd *cobra.Command, flag string) []string {
	value, err := cmd.Flags().GetString(flag)
	checkError(err)
	return stringutil.Split(value, ";")
}

func getFlagCommaSeparatedInts(cmd *cobra.Command, flag string) []int {
	filedsStrList := getFlagCommaSeparatedStrings(cmd, flag)
	fields := make([]int, len(filedsStrList))
	for i, value := range filedsStrList {
		v, err := strconv.Atoi(value)
		if err != nil {
			checkError(fmt.Errorf("value of flag --%s should be comma separated integers", flag))
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

func unshift(list *[]string, val string) {
	if len(*list) == 0 {
		list = &[]string{val}
		return
	}
	*list = append(*list, "")
	copy((*list)[1:], (*list)[0:len(*list)-1])
	(*list)[0] = val
}

// Config is the struct containing all global flags
type Config struct {
	Verbose bool

	NumCPUs int

	Delimiter    rune
	OutDelimiter rune
	// QuoteChar   rune
	CommentChar rune
	LazyQuotes  bool

	Tabs        bool
	OutTabs     bool
	NoHeaderRow bool
	NoOutHeader bool

	ShowRowNumber bool

	OutFile string

	IgnoreEmptyRow   bool
	IgnoreIllegalRow bool
}

func isTrue(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" || strings.ToLower(s) == "false" {
		return false
	}
	return true
}

func getConfigs(cmd *cobra.Command) Config {
	var val string

	var tabs bool
	if val = os.Getenv("CSVTK_T"); val != "" {
		tabs = isTrue(val)
	} else if os.Args[0] == "tsvtk" {
		tabs = true
	} else {
		tabs = getFlagBool(cmd, "tabs")
	}

	var noHeaderRow bool
	if val = os.Getenv("CSVTK_H"); val != "" {
		noHeaderRow = isTrue(val)
	} else {
		noHeaderRow = getFlagBool(cmd, "no-header-row")
	}

	var verbose bool
	if val = os.Getenv("CSVTK_QUIET"); val != "" {
		verbose = !isTrue(val)
	} else {
		verbose = !getFlagBool(cmd, "quiet")
	}

	threads := getFlagPositiveInt(cmd, "num-cpus")
	if threads >= 1000 {
		checkError(fmt.Errorf("are your seriously? %d threads? It will exhaust your RAM", threads))
	} else if threads < 1 {
		threads = runtime.NumCPU()
	}

	return Config{
		Verbose: verbose,
		NumCPUs: threads,

		Delimiter:    getFlagRune(cmd, "delimiter"),
		OutDelimiter: getFlagRune(cmd, "out-delimiter"),
		// QuoteChar:   getFlagRune(cmd, "quote-char"),
		CommentChar: getFlagRune(cmd, "comment-char"),
		LazyQuotes:  getFlagBool(cmd, "lazy-quotes"),

		Tabs:        tabs,
		OutTabs:     getFlagBool(cmd, "out-tabs"),
		NoHeaderRow: noHeaderRow,
		NoOutHeader: getFlagBool(cmd, "delete-header"),

		ShowRowNumber: getFlagBool(cmd, "show-row-number"),

		OutFile: getFlagString(cmd, "out-file"),

		IgnoreEmptyRow:   getFlagBool(cmd, "ignore-empty-row"),
		IgnoreIllegalRow: getFlagBool(cmd, "ignore-illegal-row"),
	}
}

func newCSVReaderByConfig(config Config, file string) (*CSVReader, error) {
	reader, err := NewCSVReader(file)
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
	reader.IgnoreEmptyRow = config.IgnoreEmptyRow
	reader.IgnoreIllegalRow = config.IgnoreIllegalRow

	reader.NoHeaderRow = config.NoHeaderRow

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
var reDigitals = regexp.MustCompile(`^[\-\+]?[\d\.,]+$|^[\-\+]?[\d\.,]+[eE][\-\+\d]+$`)
var reIntegers = regexp.MustCompile(`^[\-\+\d]+$`)
var reIntegerRange = regexp.MustCompile(`^([\-\d]+?)\-([\-\d]*?)$`)

func getFlagFields(cmd *cobra.Command, flag string) string {
	fieldsStr, err := cmd.Flags().GetString(flag)
	checkError(err)
	if fieldsStr == "" {
		checkError(fmt.Errorf("flag --%s needed", flag))
	}
	if !reFields.MatchString(fieldsStr) {
		checkError(fmt.Errorf("invalid value of flag %s", flag))
	}
	return fieldsStr
}

func nth(i int) string {
	switch i {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return fmt.Sprintf("%dth", i+1)
	}
}

func readCSV(config Config, file string) ([]string, [][]string, *CSVReader, error) {
	csvReader, err := newCSVReaderByConfig(config, file)
	if err != nil {
		return nil, nil, nil, xopen.ErrNoContent
	}

	csvReader.Read(ReadOption{
		FieldStr: "1-",
	})

	var headerRow []string
	data := make([][]string, 0, 1024)

	parseHeaderRow := !config.NoHeaderRow
	for record := range csvReader.Ch {
		if record.Err != nil {
			checkError(record.Err)
		}

		if parseHeaderRow {
			headerRow = record.All
			parseHeaderRow = false
			continue
		}
		data = append(data, record.All)
	}
	return headerRow, data, csvReader, nil
}

func readDataFrame(config Config, file string, ignoreCase bool) ([]string, map[string]string, map[string][]string, error) {
	df := make(map[string][]string)
	var colnames []string
	headerRow, data, csvReader, err := readCSV(config, file)

	if err != nil {
		return nil, nil, nil, err
	}

	// in case that col names are not unique in headerRow
	colname2headerRow := make(map[string]string, len(headerRow))

	var newName string
	if len(headerRow) > 0 {
		// in case that col names are not unique in headerRow
		colnames = make([]string, len(headerRow))
		colnamesCount := make(map[string]int, len(headerRow))
		var colLower string
		for i, col := range headerRow {
			if ignoreCase {
				colLower = strings.ToLower(col)
			}

			if colnamesCount[col] > 0 ||
				(ignoreCase && colnamesCount[colLower] > 0) {

				if config.Verbose {
					log.Warningf(`duplicated colname (%s) in file: %s. this may bring incorrect result`, col, file)
				}

				newName = fmt.Sprintf("%s_%d", col, colnamesCount[col])
				if ignoreCase {
					newName = strings.ToLower(newName)
				}
				colname2headerRow[newName] = col
				colnames[i] = newName
				if ignoreCase {
					colnamesCount[colLower]++
				} else {
					colnamesCount[col]++
				}
			} else {
				if ignoreCase {
					colname2headerRow[colLower] = col
					col = colLower
				} else {
					colname2headerRow[col] = col
				}
				colnames[i] = col
				colnamesCount[col] = 1
			}
		}
	} else {
		if len(data) == 0 {
			return colnames, colname2headerRow, df, nil
		} else if len(data) > 0 {
			colnames = make([]string, len(data[0]))
			for i := 0; i < len(data[0]); i++ {
				newName = fmt.Sprintf("%d", i+1)
				colname2headerRow[newName] = newName
				colnames[i] = newName
			}
		}
	}

	var ok bool
	var j int
	for i, col := range colnames {
		if _, ok = df[col]; !ok {
			df[col] = make([]string, 0, 1000)
		}
		for j = range data {
			df[col] = append(df[col], data[j][i])
		}
	}

	readerReport(&config, csvReader, file)

	return colnames, colname2headerRow, df, nil
}

func parseCSVfile(cmd *cobra.Command, config Config, file string,
	fieldStr string, fuzzyFields bool, returnSelectedData, returnAllData bool) ([]string, []int, [][]string, []string, [][]string, error) {

	csvReader, err := newCSVReaderByConfig(config, file)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	csvReader.Read(ReadOption{
		FieldStr:    fieldStr,
		FuzzyFields: fuzzyFields,

		DoNotAllowDuplicatedColumnName: true,
	})

	var fields []int
	var HeaderRow []string
	var HeaderRowAll []string
	var Data [][]string
	if returnSelectedData {
		Data = make([][]string, 0, 1024)
	}
	var DataAll [][]string

	checkFirstLine := true
	for record := range csvReader.Ch {
		if record.Err != nil {
			checkError(record.Err)
		}

		if checkFirstLine {
			checkFirstLine = false

			fields = record.Fields
			DataAll = make([][]string, 0, 1024)

			if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
				HeaderRowAll, HeaderRow = record.All, record.Selected
				continue
			}
		}

		if returnAllData {
			DataAll = append(DataAll, record.All)
		}
		if returnSelectedData {
			Data = append(Data, record.Selected)
		}
	}

	if config.IgnoreEmptyRow {
		if config.Verbose {
			log.Warningf("file '%s': %d empty rows ignored", file, csvReader.NumEmptyRows)
		}
	}
	if config.IgnoreIllegalRow {
		if config.Verbose {
			log.Warningf("file '%s': %d illegal rows ignored", file, csvReader.NumIllegalRows)
		}
	}

	return HeaderRow, fields, Data, HeaderRowAll, DataAll, nil
}

func removeComma(s string) string {
	if !strings.ContainsRune(s, ',') {
		return s
	}

	return strings.ReplaceAll(s, ",", "")
}

func readKVs(file string, allLeftAsValue bool) (map[string]string, error) {
	type KV [2]string
	fn := func(line string) (interface{}, bool, error) {
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			return nil, false, nil
		}
		items := strings.Split(line, "\t")
		if len(items) < 2 {
			return nil, false, nil
		}

		if allLeftAsValue {
			return KV([2]string{items[0], strings.Join(items[1:], "\t")}), true, nil
		}
		return KV([2]string{items[0], items[1]}), true, nil
	}
	kvs := make(map[string]string)
	reader, err := breader.NewBufferedReader(file, 2, 10, fn)
	if err != nil {
		return kvs, err
	}
	var items KV
	for chunk := range reader.Ch {
		if chunk.Err != nil {
			return kvs, err
		}
		for _, data := range chunk.Data {
			items = data.(KV)
			kvs[items[0]] = items[1]
		}
	}
	return kvs, nil
}

type orderedField struct {
	field int
	order int
}

type orderedFields []orderedField

func (s orderedFields) Len() int           { return len(s) }
func (s orderedFields) Less(i, j int) bool { return s[i].order < s[j].order }
func (s orderedFields) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func isStdin(file string) bool {
	return file == "-"
}

func filepathTrimExtension(file string) (string, string) {
	gz := strings.HasSuffix(file, ".gz") || strings.HasSuffix(file, ".GZ")
	if gz {
		file = file[0 : len(file)-3]
	}
	extension := filepath.Ext(file)
	name := file[0 : len(file)-len(extension)]
	if gz {
		extension += ".gz"
	}
	return name, extension
}

func filepathTrimExtension2(file string, suffixes []string) (string, string, string) {
	if suffixes == nil {
		suffixes = []string{".gz", ".xz", ".zst"}
	}

	var e, e1, e2 string
	f := strings.ToLower(file)
	for _, s := range suffixes {
		e = strings.ToLower(s)
		if strings.HasSuffix(f, e) {
			e2 = e
			file = file[0 : len(file)-len(e)]
			break
		}
	}

	e1 = filepath.Ext(file)
	name := file[0 : len(file)-len(e1)]

	return name, e1, e2
}

// ParseByteSize parses byte size from string
func ParseByteSize(val string) (int64, error) {
	val = strings.Trim(val, " \t\r\n")
	if val == "" {
		return 0, nil
	}
	var u int64
	var noUnit bool
	switch val[len(val)-1] {
	case 'B', 'b':
		u = 1
	case 'K', 'k':
		u = 1 << 10
	case 'M', 'm':
		u = 1 << 20
	case 'G', 'g':
		u = 1 << 30
	case 'T', 't':
		u = 1 << 40
	default:
		noUnit = true
		u = 1
	}
	var size float64
	var err error
	if noUnit {
		size, err = strconv.ParseFloat(val, 10)
		if err != nil {
			return 0, fmt.Errorf("invalid byte size: %s", val)
		}
		if size < 0 {
			size = 0
		}
		return int64(size), nil
	}

	if len(val) == 1 { // no value
		return 0, nil
	}

	size, err = strconv.ParseFloat(strings.Trim(val[0:len(val)-1], " \t\r\n"), 10)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size: %s", val)
	}
	if size < 0 {
		size = 0
	}
	return int64(size * float64(u)), nil
}

func UniqInts(list []int) []int {
	if len(list) == 0 {
		return []int{}
	} else if len(list) == 1 {
		return []int{list[0]}
	}

	sort.Ints(list)

	s := make([]int, 0, len(list))
	p := list[0]
	s = append(s, p)
	for _, v := range list[1:] {
		if v != p {
			s = append(s, v)
		}
		p = v
	}
	return s
}
