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
	"fmt"
	"runtime"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// uniqCmd represents the uniq command
var uniqCmd = &cobra.Command{
	GroupID: "set",

	Use:   "uniq",
	Short: "deduplicate records by selected fields without sorting",
	Long: `Deduplicate records by selected fields without sorting.

By default, keep the first record for each key; -n keeps the first N records.
-d prints the first record for each key occurring more than once; -u prints
records whose keys occur exactly once. Keys are compared across the entire
input, including non-adjacent records (unlike GNU uniq).

For this command, -d means --repeated; use --delimiter to set the input delimiter.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}

		runtime.GOMAXPROCS(config.NumCPUs)

		fieldStr := getFlagString(cmd, "fields")
		if fieldStr == "" {
			checkError(fmt.Errorf("flag -f (--fields) needed"))
		}

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		ignoreCase := getFlagBool(cmd, "ignore-case")
		keepN := getFlagPositiveInt(cmd, "keep-n")
		repeated := getFlagBool(cmd, "repeated")
		unique := getFlagBool(cmd, "unique")

		outfh, err := xopen.Wopen(config.OutFile)
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

		keysMaps := make(map[string]int, 10000)
		type group struct {
			first []string
			count int
		}
		groups := make(map[string]*group)
		var keyOrder []string

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				if config.Verbose {
					log.Warningf("csvtk uniq: skipping empty input file: %s", file)
				}

				writer.Flush()
				checkError(writer.Error())
				readerReport(&config, csvReader, file)
				return
			}
			checkError(err)
		}

		csvReader.Read(ReadOption{
			FieldStr:    fieldStr,
			FuzzyFields: fuzzyFields,

			DoNotAllowDuplicatedColumnName: true,
		})

		var key string
		var n int
		var ok bool

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

			key = encodeFields(record.Selected, ignoreCase)
			if repeated || unique {
				if g, found := groups[key]; found {
					g.count++
				} else {
					groups[key] = &group{first: append([]string(nil), record.All...), count: 1}
					keyOrder = append(keyOrder, key)
				}
				continue
			}
			if n, ok = keysMaps[key]; ok {
				if n >= keepN {
					continue
				}
				keysMaps[key]++
			} else {
				keysMaps[key] = 1
			}
			checkError(writer.Write(record.All))
		}
		for _, key = range keyOrder {
			g := groups[key]
			if (repeated && g.count > 1) || (unique && g.count == 1) {
				checkError(writer.Write(g.first))
			}
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(uniqCmd)
	uniqCmd.Flags().StringP("fields", "f", "1", `select these fields as keys. e.g -f 1,2 or -f columnA,columnB`)
	uniqCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	uniqCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	uniqCmd.Flags().IntP("keep-n", "n", 1, `keep at most N records for a key`)
	// Shadow the inherited -d shorthand, but share its --delimiter value so
	// either position of the long flag still configures the input delimiter.
	delimiter := *RootCmd.PersistentFlags().Lookup("delimiter")
	delimiter.Shorthand = ""
	uniqCmd.Flags().AddFlag(&delimiter)
	uniqCmd.Flags().BoolP("repeated", "d", false, `only print repeated keys, one first record per key`)
	uniqCmd.Flags().BoolP("unique", "u", false, `only print records whose keys occur exactly once`)
	uniqCmd.MarkFlagsMutuallyExclusive("repeated", "unique", "keep-n")

}
