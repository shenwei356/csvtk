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
	"strconv"
	"strings"

	"github.com/shenwei356/stable"
	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

// prettyCmd represents the pretty command
var prettyCmd = &cobra.Command{
	Use:   "pretty",
	Short: "convert CSV to a readable aligned table",
	Long: `convert CSV to a readable aligned table

How to:
  1. First -n/--buf-rows rows are read to check the minimum and maximum widths
     of each columns. You can also set the global thresholds -w/--min-width and
     -W/--max-width.
     1a. Cells longer than the maximum width will be wrapped (default) or
         clipped (--clip).
         Usually, the text is wrapped in space (-x/--wrap-delimiter). But if one
         word is longer than the -W/--max-width, it will be force split.
     1b. Texts are aligned left (default), center (-m/--align-center)
         or right (-r/--align-right).
  2. Remaining rows are read and immediately outputted, one by one, till the end.

Styles:

  Some preset styles are provided (-S/--style).

    default:

        id   size
        --   ----
        1    Huge
        2    Tiny

    plain:

        id   size
        1    Huge
        2    Tiny

    simple:

        -----------
        id   size
        -----------
        1    Huge
        2    Tiny
        -----------


    grid:

        +----+------+
        | id | size |
        +====+======+
        | 1  | Huge |
        +----+------+
        | 2  | Tiny |
        +----+------+

    light:

        ┌----┬------┐
        | id | size |
        ├====┼======┤
        | 1  | Huge |
        ├----┼------┤
        | 2  | Tiny |
        └----┴------┘

    bold:

        ┏━━━━┳━━━━━━┓
        ┃ id ┃ size ┃
        ┣━━━━╋━━━━━━┫
        ┃ 1  ┃ Huge ┃
        ┣━━━━╋━━━━━━┫
        ┃ 2  ┃ Tiny ┃
        ┗━━━━┻━━━━━━┛

    double:

        ╔════╦══════╗
        ║ id ║ size ║
        ╠════╬══════╣
        ║ 1  ║ Huge ║
        ╠════╬══════╣
        ║ 2  ║ Tiny ║
        ╚════╩══════╝

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		alignRights := getFlagStringSlice(cmd, "align-right")
		alignCenters := getFlagStringSlice(cmd, "align-center")
		separator := getFlagString(cmd, "separator")
		minWidth := getFlagNonNegativeInt(cmd, "min-width")
		maxWidth := getFlagNonNegativeInt(cmd, "max-width")
		bufRows := getFlagNonNegativeInt(cmd, "buf-rows")
		style := getFlagString(cmd, "style")
		clip := getFlagBool(cmd, "clip")
		clipMark := getFlagString(cmd, "clip-mark")
		wrapDelimiter := getFlagString(cmd, "wrap-delimiter")

		if len(wrapDelimiter) != 1 {
			checkError(fmt.Errorf("the value of flag -x/--wrap-delimiter should be a single character: %s", wrapDelimiter))
		}

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		file := files[0]

		csvReader, err := newCSVReaderByConfig(config, file)

		if err != nil {
			if err == xopen.ErrNoContent {
				log.Warningf("csvtk pretty: skipping empty input file: %s", file)
				return
			}
			checkError(err)
		}

		csvReader.Read(ReadOption{
			FieldStr:      "1-",
			ShowRowNumber: config.ShowRowNumber,
		})

		styles := map[string]*stable.TableStyle{
			"default": &stable.TableStyle{
				Name:            "simple",
				LineBelowHeader: stable.LineStyle{"", "-", separator, ""},

				HeaderRow: stable.RowStyle{"", separator, ""},
				DataRow:   stable.RowStyle{"", separator, ""},
				Padding:   "",
			},
			"plain":  stable.StylePlain,
			"simple": stable.StyleSimple,
			"grid":   stable.StyleGrid,
			"light":  stable.StyleLight,
			"bold":   stable.StyleBold,
			"double": stable.StyleDouble,
		}

		if style == "" {
			style = "default"
		}

		tbl := stable.New()

		tbl.WrapDelimiter(rune(wrapDelimiter[0]))

		if _style, ok := styles[strings.ToLower(style)]; ok {
			tbl.Style(_style)
		} else {
			checkError(fmt.Errorf("style not available: %s. available vaules: default, plain, simple, grid, light, bold, double", style))
		}

		if minWidth > 0 {
			tbl.MinWidth(minWidth)
		}
		if maxWidth > 0 {
			tbl.MaxWidth(maxWidth)
		}

		if clip {
			tbl.ClipCell(clipMark)
		}

		tbl.Writer(outfh, uint(bufRows))

		checkFirstLine := true
		var header []stable.Column
		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}

			if checkFirstLine {
				checkFirstLine = false

				ncols := len(record.All)
				if config.ShowRowNumber {
					ncols++
				}
				header = make([]stable.Column, ncols)

				colnames2fileds := make(map[string][]int, ncols)

				var i int
				var col string
				var ok bool
				if !config.NoHeaderRow || record.IsHeaderRow {
					for i, col = range record.All {
						if config.ShowRowNumber {
							i++
						}
						if _, ok = colnames2fileds[col]; !ok {
							colnames2fileds[col] = []int{i}
						} else {
							colnames2fileds[col] = append(colnames2fileds[col], i)
						}
					}

					for i, col = range record.Selected {
						header[i].Header = col
					}
				}

				for i = range record.All {
					col = strconv.Itoa(i + 1)
					if _, ok = colnames2fileds[col]; !ok {
						colnames2fileds[col] = []int{i}
					} else {
						colnames2fileds[col] = append(colnames2fileds[col], i)
					}
				}

				for _, col = range alignCenters {
					for _, i = range colnames2fileds[col] {
						if config.ShowRowNumber {
							i++
						}
						header[i].Align = stable.AlignCenter
					}
				}
				for _, col = range alignRights {
					for _, i = range colnames2fileds[col] {
						if config.ShowRowNumber {
							i++
						}
						header[i].Align = stable.AlignRight
					}
				}

				tbl.HeaderWithFormat(header)
				continue
			}

			tbl.AddRowStringSlice(record.Selected)
		}
		tbl.Flush()

		readerReport(&config, csvReader, file)
	},
}

func init() {
	RootCmd.AddCommand(prettyCmd)
	prettyCmd.Flags().StringP("separator", "s", "   ", "fields/columns separator")
	prettyCmd.Flags().StringSliceP("align-right", "r", []string{}, "align right for selected columns (field index or column name)")
	prettyCmd.Flags().StringSliceP("align-center", "m", []string{}, "align center/middle for selected columns (field index or column name)")
	prettyCmd.Flags().IntP("min-width", "w", 0, "min width")
	prettyCmd.Flags().IntP("max-width", "W", 0, "max width")

	prettyCmd.Flags().StringP("wrap-delimiter", "x", " ", "delimiter for wrapping cells")
	prettyCmd.Flags().IntP("buf-rows", "n", 128, "the number of rows to determine the min and max widths")
	prettyCmd.Flags().StringP("style", "S", "", "output syle. available vaules: default, plain, simple, grid, light, bold, double. check https://github.com/shenwei356/stable")
	prettyCmd.Flags().BoolP("clip", "", false, "clip longer cell instead of wrapping")
	prettyCmd.Flags().StringP("clip-mark", "", "...", "clip mark")
}
