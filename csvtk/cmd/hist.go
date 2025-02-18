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
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// histCmd represents the hist command
var histCmd = &cobra.Command{
	Use:   "hist",
	Short: "plot histogram",
	Long: `plot histogram

Notes:

  1. Output file can be set by flag -o/--out-file.
  2. File format is determined by the out file suffix.
     Supported formats: eps, jpg|jpeg, pdf, png, svg, and tif|tiff
  3. If flag -o/--out-file not set (default), image is written to stdout,
     you can display the image by pipping to "display" command of Imagemagic
     or just redirect to file.

`,
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfigs(cmd)
		plotConfig := getPlotConfigs(cmd)

		lineWidth := vg.Points(getFlagPositiveFloat64(cmd, "line-width") * plotConfig.scale)

		files := getFileListFromArgsAndFile(cmd, args, true, "infile-list", true)
		if len(files) > 1 {
			checkError(fmt.Errorf("no more than one file should be given"))
		}
		runtime.GOMAXPROCS(config.NumCPUs)

		skipNA := getFlagBool(cmd, "skip-na")
		naValues := getFlagStringSlice(cmd, "na-values")
		if skipNA && len(naValues) == 0 {
			log.Errorf("the value of --na-values should not be empty when using --skip-na")
		}
		naMap := make(map[string]interface{}, len(naValues))
		for _, na := range naValues {
			naMap[strings.ToLower(na)] = struct{}{}
		}

		file := files[0]
		headerRow, fields, data, _, _, err := parseCSVfile(cmd, config, file, plotConfig.fieldStr, false, true, false)

		if err != nil {
			// if err == xopen.ErrNoContent {
			// 	log.Warningf("csvtk hist: skipping empty input file: %s", file)
			// 	return
			// }
			checkError(err)
		}

		// =======================================

		if plotConfig.groupFieldStr != "" {
			log.Warning("flag -g (--group-field) ignored for command hist")
		}
		if plotConfig.ylab == "" {
			plotConfig.ylab = "Count"
		}

		if plotConfig.xlab == "" && plotConfig.groupFieldStr == "" && len(headerRow) > 0 {
			plotConfig.xlab = headerRow[0]
		}

		bins := getFlagPositiveInt(cmd, "bins")
		colorIndex := getFlagPositiveInt(cmd, "color-index")
		if colorIndex > 7 {
			checkError(fmt.Errorf("unsupported color index"))
		}

		v := make(plotter.Values, 0, len(data))
		var f float64
		var ok bool
		for _, d := range data {
			if skipNA {
				if _, ok = naMap[strings.ToLower(d[0])]; ok {
					continue
				}
			}
			f, err = strconv.ParseFloat(d[0], 64)
			if err != nil {
				if len(headerRow) > 0 {
					checkError(fmt.Errorf("fail to parse data: %s at column: %s. please choose the right column by flag -f (--data-field)", d[0], headerRow[0]))
				} else {
					checkError(fmt.Errorf("fail to parse data: %s at column: %d. please choose the right column by flag -f (--data-field)", d[0], fields[0]))
				}
			}
			v = append(v, f)
		}

		percentiles := getFlagBool(cmd, "percentiles")
		if percentiles {
			sort.Float64s(v)
			plotConfig.xlab = fmt.Sprintf("%s\nP99=%.3f P95=%.3f\nMEAN=%.3f STDDEV=%.3f\n", plotConfig.xlab, getPercentile(0.99, v), getPercentile(0.95, v), getPercentile(0.5, v), stat.StdDev(v, nil))
		}

		p := plot.New()

		h, err := plotter.NewHist(v, bins)
		if err != nil {
			checkError(err)
		}

		// h.Normalize(1)
		h.FillColor = plotutil.Color(colorIndex - 1)
		h.LineStyle.Width = lineWidth
		p.Add(h)

		p.Title.Text = plotConfig.title
		p.Title.TextStyle.Font.Size = plotConfig.titleSize
		p.X.Label.Text = plotConfig.xlab
		p.Y.Label.Text = plotConfig.ylab
		p.X.Label.TextStyle.Font.Size = plotConfig.labelSize
		p.Y.Label.TextStyle.Font.Size = plotConfig.labelSize
		p.X.Width = plotConfig.axisWidth
		p.Y.Width = plotConfig.axisWidth
		p.X.Tick.Width = plotConfig.tickWidth
		p.Y.Tick.Width = plotConfig.tickWidth
		p.X.Tick.Label.Font.Size = plotConfig.tickLabelSize
		p.Y.Tick.Label.Font.Size = plotConfig.tickLabelSize

		if plotConfig.scaleLnX {
			p.X.Scale = plot.LogScale{}
		}
		if plotConfig.scaleLnY {
			p.Y.Scale = plot.LogScale{}
		}

		if plotConfig.xminStr != "" {
			p.X.Min = plotConfig.xmin
		}
		if plotConfig.xmaxStr != "" {
			p.X.Max = plotConfig.xmax
		}
		if plotConfig.yminStr != "" {
			p.Y.Min = plotConfig.ymin
		}
		if plotConfig.ymaxStr != "" {
			p.Y.Max = plotConfig.ymax
		}

		// Save image
		if isStdin(config.OutFile) {
			fh, err := p.WriterTo(plotConfig.width*vg.Inch,
				plotConfig.height*vg.Inch,
				plotConfig.format)
			checkError(err)
			_, err = fh.WriteTo(os.Stdout)
			checkError(err)
		} else {
			checkError(p.Save(plotConfig.width*vg.Inch,
				plotConfig.height*vg.Inch,
				config.OutFile))
		}
	},
}

func getPercentile(percentile float64, vals []float64) (p float64) {
	sort.Float64s(vals)
	p = stat.Quantile(percentile, stat.Empirical, vals, nil)
	return
}

func init() {
	plotCmd.AddCommand(histCmd)
	histCmd.Flags().IntP("bins", "", 50, `number of bins`)
	histCmd.Flags().IntP("color-index", "", 1, `color index, 1-7`)
	histCmd.Flags().BoolP("percentiles", "", false, `calculate percentiles`)

	histCmd.Flags().Float64P("line-width", "", 1, "line width")

}
