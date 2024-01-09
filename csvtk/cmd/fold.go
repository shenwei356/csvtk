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
	"runtime"
	"strings"

	"github.com/shenwei356/util/stringutil"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// collapseCmd represents the colapse command
var collapseCmd = &cobra.Command{
	GroupID: "transform",

	Use:     "fold",
	Aliases: []string{"collapse"},
	Short:   "fold multiple values of a field into cells of groups",
	Long: `fold multiple values of a field into cells of groups

Attention:

    Only grouping fields and value filed are outputted.

Example:

    $ echo -ne "id,value,meta\n1,a,12\n1,b,34\n2,c,56\n2,d,78\n" \
        | csvtk pretty
    id   value   meta
    1    a       12
    1    b       34
    2    c       56
    2    d       78
    
    $ echo -ne "id,value,meta\n1,a,12\n1,b,34\n2,c,56\n2,d,78\n" \
        | csvtk fold -f id -v value -s ";" \
        | csvtk pretty
    id   value
    1    a;b
    2    c;d
    
    $ echo -ne "id,value,meta\n1,a,12\n1,b,34\n2,c,56\n2,d,78\n" \
        | csvtk fold -f id -v value -s ";" \
        | csvtk unfold -f value -s ";" \
        | csvtk pretty
    id   value
    1    a
    1    b
    2    c
    2    d

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

		vfieldStr := getFlagString(cmd, "vfield")
		if vfieldStr == "" {
			checkError(fmt.Errorf("flag -v (--vfield) needed"))
		}

		if fieldStr == vfieldStr {
			checkError(fmt.Errorf("values of -v (--vfield) and -f (--fields) should be different"))
		}

		separater := getFlagString(cmd, "separater")
		if separater == "" {
			checkError(fmt.Errorf("flag -s (--separater) needed"))
		}

		fieldStr = fmt.Sprintf("%s,%s", fieldStr, vfieldStr)

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

		key2data := make(map[string][]string, 10000)
		orders := make(map[string]int, 10000)

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				log.Warningf("csvtk fold: skipping empty input file: %s", file)

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

		var items []string
		var key string
		var N int
		var ok bool

		checkFirstLine := true
		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if checkFirstLine {
				checkFirstLine = false

				if !config.NoHeaderRow || record.IsHeaderRow { // do not replace head line
					checkError(writer.Write(record.Selected))
					continue
				}
			}

			N++

			items = record.Selected

			key = strings.Join(items[0:len(items)-1], "_shenwei356_")
			if _, ok = key2data[key]; !ok {
				key2data[key] = make([]string, 0, 1)
			}
			key2data[key] = append(key2data[key], items[len(items)-1])
			orders[key] = N
		}

		orderedKey := stringutil.SortCountOfString(orders, false)
		for _, o := range orderedKey {
			items = strings.Split(o.Key, "_shenwei356_")
			items = append(items, strings.Join(key2data[o.Key], separater))
			checkError(writer.Write(items))
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(collapseCmd)
	collapseCmd.Flags().StringP("fields", "f", "1", `key fields for grouping. e.g -f 1,2 or -f columnA,columnB`)
	collapseCmd.Flags().StringP("vfield", "v", "", `value field for folding`)
	collapseCmd.Flags().BoolP("ignore-case", "i", false, `ignore case`)
	collapseCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields (only for key fields), e.g., -F -f "*name" or -F -f "id123*"`)
	collapseCmd.Flags().StringP("separater", "s", "; ", "separater for folded values")
}
