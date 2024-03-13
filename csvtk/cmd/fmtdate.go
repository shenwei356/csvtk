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
	"time"

	"github.com/araddon/dateparse"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
	"gitlab.com/metakeule/fmtdate"
)

// fmtdateCmd represents the replace command
var fmtdateCmd = &cobra.Command{
	GroupID: "edit",

	Use:   "fmtdate",
	Short: "format date of selected fields",
	Long: `format date of selected fields

Date parsing is supported by: https://github.com/araddon/dateparse
Date formating is supported by: https://github.com/metakeule/fmtdate

Time zones: 
    format: Asia/Shanghai
    whole list: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

Output format is in MS Excel (TM) syntax.
Placeholders:

    M    - month (1)
    MM   - month (01)
    MMM  - month (Jan)
    MMMM - month (January)
    D    - day (2)
    DD   - day (02)
    DDD  - day (Mon)
    DDDD - day (Monday)
    YY   - year (06)
    YYYY - year (2006)
    hh   - hours (15)
    mm   - minutes (04)
    ss   - seconds (05)

    AM/PM hours: 'h' followed by optional 'mm' and 'ss' followed by 'pm', e.g.

    hpm        - hours (03PM)
    h:mmpm     - hours:minutes (03:04PM)
    h:mm:sspm  - hours:minutes:seconds (03:04:05PM)

    Time zones: a time format followed by 'ZZZZ', 'ZZZ' or 'ZZ', e.g.

    hh:mm:ss ZZZZ (16:05:06 +0100)
    hh:mm:ss ZZZ  (16:05:06 CET)
    hh:mm:ss ZZ   (16:05:06 +01:00)

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		timezone := getFlagString(cmd, "time-zone")
		outfmt := getFlagString(cmd, "format")
		keepUnparsed := getFlagBool(cmd, "keep-unparsed")

		if timezone != "" {
			loc, err := time.LoadLocation(timezone)
			if err != nil {
				checkError(fmt.Errorf("setting time zone: %s", err))
			}
			time.Local = loc
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
					if config.Verbose {
						log.Warningf("csvtk fmtdate: skipping empty input file: %s", file)
					}
					continue
				}
				checkError(err)
			}

			csvReader.Read(ReadOption{
				FieldStr:    fieldStr,
				FuzzyFields: fuzzyFields,
			})

			checkFirstLine := true
			var f int
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

				for _, f = range record.Fields {
					t, err := dateparse.ParseLocal(record.All[f-1])
					if err != nil {
						if !keepUnparsed {
							record.All[f-1] = ""
						}
					} else {
						record.All[f-1] = fmtdate.Format(outfmt, t)
					}
				}
				checkError(writer.Write(record.All))
			}

			readerReport(&config, csvReader, file)
		}
	},
}

func init() {
	RootCmd.AddCommand(fmtdateCmd)
	fmtdateCmd.Flags().StringP("fields", "f", "1", `select only these fields. e.g -f 1,2 or -f columnA,columnB`)
	fmtdateCmd.Flags().BoolP("fuzzy-fields", "F", false, `using fuzzy fields, e.g., -F -f "*name" or -F -f "id123*"`)
	fmtdateCmd.Flags().StringP("format", "", "YYYY-MM-DD hh:mm:ss", `output date format in MS Excel (TM) syntax, type "csvtk fmtdate -h" for details`)
	fmtdateCmd.Flags().BoolP("keep-unparsed", "k", false, "keep the key as value when no value found for the key")
	fmtdateCmd.Flags().StringP("time-zone", "z", "", `timezone aka "Asia/Shanghai" or "America/Los_Angeles" formatted time-zone, type "csvtk fmtdate -h" for details`)
}
