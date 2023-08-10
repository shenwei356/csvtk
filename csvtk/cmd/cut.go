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

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// cutCmd represents the cut command
var cutCmd = &cobra.Command{
	Use:   "cut",
	Short: "select and arrange fields",
	Long: `select and arrange fields

Examples:

  1. Single column
     csvtk cut -f 1
     csvtk cut -f colA
  2. Multiple columns (replicates allowed)
     csvtk cut -f 1,3,2,1
     csvtk cut -f colA,colB,colA
  3. Column ranges
     csvtk cut -f 1,3-5       # 1, 3, 4, 5
     csvtk cut -f 3,5-        # 3rd col, and 5th col to the end
     csvtk cut -f 1-          # for all
     csvtk cut -f 2-,1        # move 1th col to the end
  4. Unselect
     csvtk cut -f -1,-3       # discard 1st and 3rd column
     csvtk cut -f -1--3       # discard 1st to 3rd column
     csvtk cut -f -2-         # discard 2nd and all columns on the right.
     csvtu cut -f -colA,-colB # discard colA and colB
	 
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

		uniqColumn := getFlagBool(cmd, "uniq-column")

		fuzzyFields := getFlagBool(cmd, "fuzzy-fields")
		ignoreCase := getFlagBool(cmd, "ignore-case")

		allowMissingColumn := getFlagBool(cmd, "allow-missing-col")
		blankMissingColumn := getFlagBool(cmd, "blank-missing-col")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		writer := csv.NewWriter(outfh)
		if config.OutTabs || config.Tabs {
			if config.OutDelimiter == ',' { // default value, no other value given
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

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				log.Warningf("csvtk cut: skipping empty input file: %s", file)

				writer.Flush()
				checkError(writer.Error())
				readerReport(&config, csvReader, file)
				return
			}
			checkError(err)
		}

		csvReader.Read(ReadOption{
			FieldStr:           fieldStr,
			FuzzyFields:        fuzzyFields,
			IgnoreFieldCase:    ignoreCase,
			UniqColumn:         uniqColumn,
			AllowMissingColumn: allowMissingColumn,
			BlankMissingColumn: blankMissingColumn,
			ShowRowNumber:      config.ShowRowNumber,
		})

		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			writer.Write(record.Selected)
		}

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(cutCmd)
	cutCmd.Flags().StringP("fields", "f", "", `select only these fields. type "csvtk cut -h" for examples`)
	cutCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	cutCmd.Flags().BoolP("ignore-case", "i", false, `ignore case (column name)`)
	cutCmd.Flags().BoolP("uniq-column", "u", false, `deduplicate columns matched by multiple fuzzy column names`)
	cutCmd.Flags().BoolP("allow-missing-col", "m", false, `allow missing column`)
	cutCmd.Flags().BoolP("blank-missing-col", "b", false, `blank missing column, only for using column fields`)
}
