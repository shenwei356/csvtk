// Copyright © 2016-2026 Wei Shen <shenwei356@gmail.com>
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
	"html"
	"io"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/shenwei356/xopen"
	"github.com/spf13/cobra"
)

var htmlTableWidthRegexp = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?|\.[0-9]+)(px|%|em|rem|vw|vh|ch)?$`)

// csv2htmlCmd represents the csv2html command.
var csv2htmlCmd = &cobra.Command{
	GroupID: "format",

	Use:   "csv2html",
	Short: "convert CSV/TSV to a standalone HTML table",
	Long: `convert CSV/TSV to a standalone HTML table

The output contains all CSS needed to display the table. The table uses its
natural content width and is limited by --table-width.

The value of --table-width can be a number in pixels or a CSS length using
px, %, em, rem, vw, vh, or ch, for example: 1200, 90%, or 70rem. Percentage
values greater than 100% are allowed.
`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		caption := getFlagString(cmd, "caption")
		tableClass := getFlagString(cmd, "table-class")
		tableWidth, err := normalizeHTMLTableWidth(getFlagString(cmd, "table-width"))
		checkError(err)
		minWidths := getFlagStringSliceAsInts(cmd, "min-width")
		maxWidths := getFlagStringSliceAsInts(cmd, "max-width")

		outfh, err := xopen.Wopen(config.OutFile)
		checkError(err)
		defer outfh.Close()

		file := files[0]
		csvReader, err := newCSVReaderByConfig(config, file)
		if err != nil {
			if err == xopen.ErrNoContent {
				checkError(validateHTMLColumnWidths(minWidths, maxWidths, 0))
				writeHTMLDocumentStart(outfh, caption, tableClass, tableWidth, minWidths, maxWidths, 0)
				checkWriteString(outfh, "      <tbody>\n")
				writeHTMLDocumentEnd(outfh)
				if config.Verbose {
					log.Warningf("csvtk csv2html: empty input file: %s", file)
				}
				return
			}
			checkError(err)
		}

		csvReader.Read(ReadOption{FieldStr: "1-"})
		first, ok := <-csvReader.Ch
		if ok && first.Err != nil {
			checkError(first.Err)
		}

		ncols := 0
		if ok {
			ncols = len(first.Selected)
		}
		checkError(validateHTMLColumnWidths(minWidths, maxWidths, ncols))
		writeHTMLDocumentStart(outfh, caption, tableClass, tableWidth, minWidths, maxWidths, ncols)

		hasHeader := ok && (!config.NoHeaderRow || first.IsHeaderRow)
		if hasHeader && !config.NoOutHeader {
			checkWriteString(outfh, "      <thead>\n")
			writeHTMLRow(outfh, first.Selected, true)
			checkWriteString(outfh, "      </thead>\n")
		}

		checkWriteString(outfh, "      <tbody>\n")
		if ok && !hasHeader {
			writeHTMLRow(outfh, first.Selected, false)
		}
		for record := range csvReader.Ch {
			if record.Err != nil {
				checkError(record.Err)
			}
			writeHTMLRow(outfh, record.Selected, false)
		}
		writeHTMLDocumentEnd(outfh)

		readerReport(&config, csvReader, file)
	},
}

func normalizeHTMLTableWidth(width string) (string, error) {
	width = strings.TrimSpace(width)
	match := htmlTableWidthRegexp.FindStringSubmatch(width)
	if match == nil {
		return "", fmt.Errorf("invalid table width: %s", width)
	}
	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil || value <= 0 {
		return "", fmt.Errorf("table width should be greater than zero: %s", width)
	}
	if match[2] == "" {
		return width + "px", nil
	}
	return width, nil
}

func validateHTMLColumnWidths(minWidths, maxWidths []int, ncols int) error {
	for _, widths := range [][]int{minWidths, maxWidths} {
		for _, width := range widths {
			if width < 0 {
				return fmt.Errorf("column widths should not be negative: %d", width)
			}
		}
		if len(widths) > 1 && len(widths) != ncols {
			return fmt.Errorf("number of column widths (%d) should be equal to 1 or number of fields (%d)", len(widths), ncols)
		}
	}

	for i := 0; i < ncols; i++ {
		minWidth := htmlColumnWidthAt(minWidths, i)
		maxWidth := htmlColumnWidthAt(maxWidths, i)
		if minWidth > 0 && maxWidth > 0 && minWidth > maxWidth {
			return fmt.Errorf("minimum width (%d) should not exceed maximum width (%d) for column %d", minWidth, maxWidth, i+1)
		}
	}
	return nil
}

func htmlColumnWidthAt(widths []int, i int) int {
	if len(widths) == 0 {
		return 0
	}
	if len(widths) == 1 {
		return widths[0]
	}
	return widths[i]
}

func htmlColumnWidthCSS(minWidths, maxWidths []int, ncols int) string {
	var css strings.Builder
	for i := 0; i < ncols; i++ {
		minWidth := htmlColumnWidthAt(minWidths, i)
		maxWidth := htmlColumnWidthAt(maxWidths, i)
		if minWidth == 0 && maxWidth == 0 {
			continue
		}
		fmt.Fprintf(&css, "    th:nth-child(%d) > .csvtk-cell, td:nth-child(%d) > .csvtk-cell {", i+1, i+1)
		if minWidth > 0 {
			fmt.Fprintf(&css, " min-width: %dch;", minWidth)
		}
		if maxWidth > 0 {
			fmt.Fprintf(&css, " max-width: %dch;", maxWidth)
		}
		css.WriteString(" }\n")
	}
	return css.String()
}

func writeHTMLDocumentStart(out io.Writer, caption, tableClass, tableWidth string, minWidths, maxWidths []int, ncols int) {
	title := caption
	if title == "" {
		title = "csvtk table"
	}
	checkWriteString(out, "<!doctype html>\n<html lang=\"en\">\n<head>\n")
	checkWriteString(out, "  <meta charset=\"utf-8\">\n")
	checkWriteString(out, "  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	checkFprintf(out, "  <title>%s</title>\n", html.EscapeString(title))
	checkWriteString(out, `  <style>
    :root { color-scheme: light dark; --background: #fff; --foreground: #24292f; --border: #d0d7de; --stripe: #f6f8fa; --header: #eef1f4; }
    @media (prefers-color-scheme: dark) {
      :root { --background: #0d1117; --foreground: #e6edf3; --border: #30363d; --stripe: #161b22; --header: #21262d; }
    }
    * { box-sizing: border-box; }
    body { margin: 0; padding: 1rem; color: var(--foreground); background: var(--background); font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
`)
	checkWriteString(out, "    .csvtk-table-wrapper { width: 100%; overflow-x: auto; }\n")
	checkFprintf(out, "    .csvtk-table-wrapper table { width: max-content; max-width: %s; margin-inline: auto; border-collapse: collapse; border-spacing: 0; }\n", tableWidth)
	checkWriteString(out, `
    caption { padding: 0.5rem; font-size: 1.15rem; font-weight: 600; text-align: left; }
    th, td { padding: 0.5rem 0.75rem; border: 1px solid var(--border); text-align: left; vertical-align: top; }
    th { background: var(--header); }
    tbody tr:nth-child(even) { background: var(--stripe); }
    .csvtk-cell { display: block; white-space: pre-wrap; overflow-wrap: anywhere; }
`)
	checkWriteString(out, htmlColumnWidthCSS(minWidths, maxWidths, ncols))
	checkWriteString(out, "  </style>\n</head>\n<body>\n  <div class=\"csvtk-table-wrapper\">\n")
	checkFprintf(out, "    <table class=\"%s\">\n", html.EscapeString(tableClass))
	if caption != "" {
		checkFprintf(out, "      <caption>%s</caption>\n", html.EscapeString(caption))
	}
}

func writeHTMLRow(out io.Writer, fields []string, header bool) {
	checkWriteString(out, "        <tr>\n")
	for _, field := range fields {
		if header {
			checkFprintf(out, "          <th scope=\"col\"><span class=\"csvtk-cell\">%s</span></th>\n", html.EscapeString(field))
		} else {
			checkFprintf(out, "          <td><span class=\"csvtk-cell\">%s</span></td>\n", html.EscapeString(field))
		}
	}
	checkWriteString(out, "        </tr>\n")
}

func writeHTMLDocumentEnd(out io.Writer) {
	checkWriteString(out, "      </tbody>\n    </table>\n  </div>\n</body>\n</html>\n")
}

func checkWriteString(out io.Writer, s string) {
	_, err := io.WriteString(out, s)
	checkError(err)
}

func checkFprintf(out io.Writer, format string, args ...interface{}) {
	_, err := fmt.Fprintf(out, format, args...)
	checkError(err)
}

func init() {
	RootCmd.AddCommand(csv2htmlCmd)
	csv2htmlCmd.Flags().String("caption", "", "table caption")
	csv2htmlCmd.Flags().String("table-class", "csvtk-table", "CSS class of the table")
	csv2htmlCmd.Flags().String("table-width", "100%", "maximum table width (a number in pixels or a CSS length with px, %, em, rem, vw, vh, or ch)")
	csv2htmlCmd.Flags().StringSliceP("min-width", "w", []string{}, "minimum cell width in characters; one value for all columns or comma-separated values for each column (0 for no limit)")
	csv2htmlCmd.Flags().StringSliceP("max-width", "W", []string{}, "maximum cell width in characters; one value for all columns or comma-separated values for each column (0 for no limit)")
}
