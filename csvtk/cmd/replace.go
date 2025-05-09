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
	"encoding/csv"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// replaceCmd represents the replace command
var replaceCmd = &cobra.Command{
	GroupID: "edit",

	Use:   "replace",
	Short: "replace data of selected fields by regular expression",
	Long: `replace data of selected fields by regular expression

Note that the replacement supports capture variables.
e.g. $1 represents the text of the first submatch.
ATTENTION: use SINGLE quote NOT double quotes in *nix OS.

Examples: Adding space to cell values.

  csvtk replace -p "(.)" -r '$1 '

Or use the \ escape character.

  csvtk replace -p "(.)" -r "\$1 "

more on: http://shenwei356.github.io/csvtk/usage/#replace

Special replacement symbols:

  {nr}    Record number, starting from 1
  {gnr}   Record number within a group,    defined by value(s) of field(s) via -g/--group
  {enr}   Enumeration numbering of groups, defined by value(s) of field(s) via -g/--group
  {rnr}   Running numbering of groups,     defined by value(s) of field(s) via -g/--group

          Examples:
            group   {nr}   {gnr}   {enr}   {rnr}
            A        1       1       1       1
            A        2       2       1       1
            B        3       1       2       2
            B        4       2       2       2
            A        5       3       1       3
            C        6       1       3       4

  {kv}    Corresponding value of the key (captured variable $n) by key-value file,
          n can be specified by flag --key-capt-idx (default: 1)

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		pattern := getFlagString(cmd, "pattern")
		replacement := getFlagString(cmd, "replacement")
		ignoreCase := getFlagBool(cmd, "ignore-case")
		if pattern == "" {
			checkError(fmt.Errorf("flags -p (--pattern) needed"))
		}

		p := pattern
		if ignoreCase {
			p = "(?i)" + p
		}
		patternRegexp, err := regexp.Compile(p)
		checkError(err)

		kvFile := getFlagString(cmd, "kv-file")
		keepKey := getFlagBool(cmd, "keep-key")
		keyCaptIdx := getFlagPositiveInt(cmd, "key-capt-idx")
		keyMissRepl := getFlagString(cmd, "key-miss-repl")
		nrWidth := getFlagPositiveInt(cmd, "nr-width")
		nrFormat := fmt.Sprintf("%%0%dd", nrWidth)
		startNum := getFlagNonNegativeInt(cmd, "start-num")
		incrNum := getFlagPositiveInt(cmd, "incr-num")

		kvFileAllLeftColumnsAsValue := getFlagBool(cmd, "kv-file-all-left-columns-as-value")

		var replaceWithNR bool
		if reNR.MatchString(replacement) {
			replaceWithNR = true
		}

		var replaceWithKV bool
		var kvs map[string]string
		if reKV.MatchString(replacement) {
			replaceWithKV = true
			if !regexp.MustCompile(`\(.+\)`).MatchString(pattern) {
				checkError(fmt.Errorf(`value of -p (--pattern) must contains "(" and ")" to capture data which is used specify the KEY`))
			}
			if kvFile == "" {
				checkError(fmt.Errorf(`since replacement symbol "{kv}"/"{KV}" found in value of flag -r (--replacement), tab-delimited key-value file should be given by flag -k (--kv-file)`))
			}

			if config.Verbose {
				log.Infof("read key-value file: %s", kvFile)
			}
			kvs, err = readKVs(kvFile, kvFileAllLeftColumnsAsValue)
			if err != nil {
				checkError(fmt.Errorf("read key-value file: %s", err))
			}
			if len(kvs) == 0 {
				checkError(fmt.Errorf("no valid data in key-value file: %s", kvFile))
			}

			if ignoreCase {
				kvs2 := make(map[string]string, len(kvs))
				for k, v := range kvs {
					kvs2[strings.ToLower(k)] = v
				}
				kvs = kvs2
			}

			if config.Verbose {
				log.Infof("%d pairs of key-value loaded", len(kvs))
			}
		}

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}
		groups := getFlagStringSlice(cmd, "group")
		setGroup := len(groups) > 0
		groupCols := len(groups)

		if reGNR.MatchString(replacement) || reENR.MatchString(replacement) || reRNR.MatchString(replacement) {
			if !setGroup {
				checkError(fmt.Errorf(`flag -g (--group) needed for using "{gnr}", "{enr}", and "{rnr}"`))
			}
		} else if setGroup {
			checkError(fmt.Errorf(`flag -g (--group) given, but "{gnr}", "{enr}", "{rnr}" not found in -r/--replacement: %s`, replacement))
		}

		replaceWithGNR := reGNR.MatchString(replacement) && setGroup
		replaceWithENR := reENR.MatchString(replacement) && setGroup
		replaceWithRNR := reRNR.MatchString(replacement) && setGroup
		_replaceWithXNR := replaceWithGNR || replaceWithENR || replaceWithRNR

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
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

		for _, file := range files {
			csvReader, err := newCSVReaderByConfig(config, file)

			if err != nil {
				if err == xopen.ErrNoContent {
					if config.Verbose {
						log.Warningf("csvtk replace: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}

			_fieldStr := fieldStr
			if _replaceWithXNR {
				_fieldStr += "," + strings.Join(groups, ",")
			}

			csvReader.Read(ReadOption{
				FieldStr:    _fieldStr,
				FuzzyFields: fuzzyFields,

				DoNotAllowDuplicatedColumnName: true,
			})

			var i int
			var r string
			var ok bool
			var found []string
			var founds [][]string
			var k string
			nr := startNum

			var group string
			groupColData := make([]string, groupCols)
			var mg map[string]int
			if replaceWithGNR {
				mg = make(map[string]int)
			}
			var me map[string]int
			if replaceWithENR {
				me = make(map[string]int)
			}

			var iGroup, gnr, enr, rnr int
			iGroup = startNum - incrNum
			rnr = startNum - incrNum
			groupPre := "_shenwei356__"

			var fields []int

			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					checkFirstLine = false

					if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
						if config.NoOutHeader {
							continue
						}
						checkError(writer.Write(record.All))
						continue
					}
				}

				if _replaceWithXNR {
					fields = record.Fields[:len(record.Fields)-groupCols]
					for i := 0; i < groupCols; i++ {
						groupColData[i] = record.All[record.Fields[len(record.Fields)-groupCols+i]-1]
					}
					group = strings.Join(groupColData, "_shenwei356_")

					if replaceWithGNR {
						if _, ok = mg[group]; !ok {
							mg[group] = startNum - incrNum
						}
						mg[group] += incrNum
						gnr = mg[group]
					}
					if replaceWithENR {
						if _, ok = me[group]; !ok {
							iGroup += incrNum
							me[group] = iGroup
						}
						enr = me[group]
					}
					if replaceWithRNR {
						if group != groupPre {
							rnr += incrNum
						}
						groupPre = group
					}
				} else {
					fields = record.Fields
				}

				for _, i = range fields {
					i--

					r = replacement

					r = reTab.ReplaceAllString(r, "\t")

					if replaceWithNR {
						r = reNR.ReplaceAllString(r, fmt.Sprintf(nrFormat, nr))
					}

					if replaceWithGNR {
						r = reGNR.ReplaceAllString(r, fmt.Sprintf(nrFormat, gnr))
					}

					if replaceWithENR {
						r = reENR.ReplaceAllString(r, fmt.Sprintf(nrFormat, enr))
					}

					if replaceWithRNR {
						r = reRNR.ReplaceAllString(r, fmt.Sprintf(nrFormat, rnr))
					}

					if replaceWithKV {
						founds = patternRegexp.FindAllStringSubmatch(record.All[i], -1)
						if len(founds) > 1 {
							checkError(fmt.Errorf(`pattern "%s" matches multiple targets in "%s", this will cause chaos`, p, record.All[i]))
						}
						if len(founds) > 0 {
							found = founds[0]
							if keyCaptIdx > len(found)-1 {
								checkError(fmt.Errorf("value of flag -I (--key-capt-idx) overflows"))
							}
							k = string(found[keyCaptIdx])
							if ignoreCase {
								k = strings.ToLower(k)
							}
							if _, ok = kvs[k]; ok {
								r = reKV.ReplaceAllString(r, kvs[k])
							} else if keepKey {
								r = reKV.ReplaceAllString(r, found[keyCaptIdx])
							} else {
								r = reKV.ReplaceAllString(r, keyMissRepl)
							}
						}
					}

					record.All[i] = patternRegexp.ReplaceAllString(record.All[i], r)
				}
				checkError(writer.Write(record.All))

				nr++
			}

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	RootCmd.AddCommand(replaceCmd)
	replaceCmd.Flags().StringP("fields", "f", "1", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
	replaceCmd.Flags().StringSliceP("group", "g", []string{}, `select field(s) for group-specific record numbering, including {gnr}, {enr}, {rnr}. Please use the same field type (field number or column name) with the value of -f/--fields`)
	replaceCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	replaceCmd.Flags().StringP("pattern", "p", ".*", "search regular expression")
	replaceCmd.Flags().StringP("replacement", "r", "",
		"replacement. supporting capture variables. "+
			" e.g. $1 represents the text of the first submatch. "+
			"ATTENTION: for *nix OS, use SINGLE quote NOT double quotes or "+
			`use the \ escape character. Record number is also supported by "{nr}".`+
			`use ${1} instead of $1 when {kv} given!`)
	replaceCmd.Flags().BoolP("ignore-case", "i", false, "ignore case")
	replaceCmd.Flags().StringP("kv-file", "k", "",
		`tab-delimited key-value file for replacing key with value when using "{kv}" in -r (--replacement)`)
	replaceCmd.Flags().BoolP("keep-key", "K", false, "keep the key as value when no value found for the key")
	replaceCmd.Flags().IntP("key-capt-idx", "", 1, "capture variable index of key (1-based)")
	replaceCmd.Flags().StringP("key-miss-repl", "", "", "replacement for key with no corresponding value")
	replaceCmd.Flags().IntP("nr-width", "", 1, `minimum width for {nr}, {gnr}, {enr}, {rnr} in flag -r/--replacement. e.g., formating "1" to "001" by --nr-width 3`)
	replaceCmd.Flags().IntP("start-num", "n", 1, `starting number when using {nr}, {gnr}, {enr}, {rnr} in replacement`)
	replaceCmd.Flags().IntP("incr-num", "", 1, `increment number when using  {nr}, {gnr}, {enr}, {rnr} in replacement`)

	replaceCmd.Flags().BoolP("kv-file-all-left-columns-as-value", "A", false, "treat all columns except 1th one as value for kv-file with more than 2 columns")
}

var reNR = regexp.MustCompile(`\{(NR|nr)\}`)
var reGNR = regexp.MustCompile(`\{(GNR|gnr)\}`)
var reENR = regexp.MustCompile(`\{(ENR|enr)\}`)
var reRNR = regexp.MustCompile(`\{(RNR|rnr)\}`)
var reKV = regexp.MustCompile(`\{(KV|kv)\}`)
var reTab = regexp.MustCompile(`\\t`)
