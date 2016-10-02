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
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/spf13/cobra"
)

// histCmd represents the hist command
var histCmd = &cobra.Command{
	Use:   "hist",
	Short: "plot histogram",
	Long: `plot histogram

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		var fieldStr string

		dataFieldStr := getFlagString(cmd, "data-field")
		if strings.Index(dataFieldStr, ",") >= 0 {
			checkError(fmt.Errorf("only one field allowed for flag --data-field"))
		}
		if dataFieldStr[0] == '-' {
			checkError(fmt.Errorf("unselect not allowed for flag --data-field"))
		}

		groupFieldStr := getFlagString(cmd, "group-field")
		if len(groupFieldStr) > 0 {
			if strings.Index(groupFieldStr, ",") >= 0 {
				checkError(fmt.Errorf("only one field allowed for flag --group-field"))
			}
			if dataFieldStr[0] == '-' {
				checkError(fmt.Errorf("unselect not allowed for flag --group-field"))
			}

			fieldStr = dataFieldStr + "," + groupFieldStr
		} else {
			fieldStr = dataFieldStr
		}

		files := getFileList(args)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		title := getFlagString(cmd, "title")
		if title == "" {
			title = "Histogram"
		}

		width := getFlagPositiveFloat64(cmd, "width")
		height := getFlagPositiveFloat64(cmd, "height")
		bins := getFlagPositiveInt(cmd, "bins")
		xlab := getFlagString(cmd, "xlab")
		ylab := getFlagString(cmd, "ylab")
		if config.OutFile == "-" {
			config.OutFile = "hist.png"
		}

		file := files[0]
		headerRow, data, fields := parseCSVfile(cmd, config, file, fieldStr, false)

		// =======================================

		v := make(plotter.Values, len(data))
		var f float64
		var err error
		for i, d := range data {
			f, err = strconv.ParseFloat(d[0], 64)
			if err != nil {
				if len(headerRow) > 0 {
					checkError(fmt.Errorf("fail to parse float: %s at column %s", d[0], headerRow[0]))
				} else {
					checkError(fmt.Errorf("fail to parse float: %s at column %d", d[0], fields[0]))
				}
			}
			v[i] = f
		}

		p, err := plot.New()
		if err != nil {
			checkError(err)
		}

		p.Title.Text = title
		p.X.Label.Text = xlab
		p.Y.Label.Text = ylab

		h, err := plotter.NewHist(v, bins)
		if err != nil {
			checkError(err)
		}

		h.Normalize(1)
		p.Add(h)

		// Save the plot to a PNG file.
		if err := p.Save(vg.Length(width*float64(vg.Inch)),
			vg.Length(height*float64(vg.Inch)), config.OutFile); err != nil {
			checkError(err)
		}
	},
}

func init() {
	plotCmd.AddCommand(histCmd)
	histCmd.Flags().StringP("data-field", "f", "1", `field index or column name of data`)
	histCmd.Flags().StringP("group-field", "g", "", `field index or column name of group`)
	histCmd.Flags().IntP("bins", "", 50, `number of bins`)
}
