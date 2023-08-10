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
	"strconv"
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// rename2Cmd represents the rename2 command
var rename2Cmd = &cobra.Command{
	Use:   "rename2",
	Short: "rename column names by regular expression",
	Long: `rename column names by regular expression

Special replacement symbols:

  {nr}  ascending number, starting from --start-num
  {kv}  Corresponding value of the key (captured variable $n) by key-value file,
        n can be specified by flag --key-capt-idx (default: 1)


`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		if config.NoHeaderRow {
			checkError(fmt.Errorf("flag --H (--no-header-row) is not allowed for this command"))
		}

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
		startNum := getFlagNonNegativeInt(cmd, "start-num")
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
			log.Infof("read key-value file: %s", kvFile)
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

			log.Infof("%d pairs of key-value loaded", len(kvs))
		}

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}

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
					log.Warningf("csvtk rename2: skipping empty input file: %s", file)
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr:    fieldStr,
				FuzzyFields: fuzzyFields,
			})

			var r string
			var found []string
			var founds [][]string
			var k string
			var nr int
			var ok bool

			checkFirstLine := true
			for record := range csvReader.Ch {
				if record.Err != nil {
					checkError(record.Err)
				}

				if checkFirstLine {
					checkFirstLine = false

					if !config.NoHeaderRow || record.IsHeaderRow {
						for _, f := range record.Fields {
							nr = startNum
							r = replacement

							if replaceWithNR {
								r = reNR.ReplaceAllString(r, strconv.Itoa(nr))
							}

							if replaceWithKV {
								founds = patternRegexp.FindAllStringSubmatch(record.All[f-1], -1)
								if len(founds) > 1 {
									checkError(fmt.Errorf(`pattern "%s" matches multiple targets in "%s", this will cause chaos`, p, record.All[f-1]))
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

							record.All[f-1] = patternRegexp.ReplaceAllString(record.All[f-1], r)

							nr++
						}

						checkError(writer.Write(record.All))
						continue
					}
				}

				checkError(writer.Write(record.All))
			}

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	RootCmd.AddCommand(rename2Cmd)
	rename2Cmd.Flags().StringP("fields", "f", "", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
	rename2Cmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	rename2Cmd.Flags().StringP("pattern", "p", "", "search regular expression")
	rename2Cmd.Flags().StringP("replacement", "r", "",
		"renamement. supporting capture variables. "+
			" e.g. $1 represents the text of the first submatch. "+
			"ATTENTION: use SINGLE quote NOT double quotes in *nix OS or "+
			`use the \ escape character. Ascending number is also supported by "{nr}".`+
			`use ${1} instead of $1 when {kv} given!`)
	rename2Cmd.Flags().BoolP("ignore-case", "i", false, "ignore case")
	rename2Cmd.Flags().StringP("kv-file", "k", "",
		`tab-delimited key-value file for replacing key with value when using "{kv}" in -r (--replacement)`)
	rename2Cmd.Flags().BoolP("keep-key", "K", false, "keep the key as value when no value found for the key")
	rename2Cmd.Flags().IntP("key-capt-idx", "", 1, "capture variable index of key (1-based)")
	rename2Cmd.Flags().StringP("key-miss-repl", "", "", "replacement for key with no corresponding value")
	rename2Cmd.Flags().IntP("start-num", "n", 1, `starting number when using {nr} in replacement`)
	rename2Cmd.Flags().BoolP("kv-file-all-left-columns-as-value", "A", false, "treat all columns except 1th one as value for kv-file with more than 2 columns")
}
