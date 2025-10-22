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

	"github.com/shenwei356/util/stringutil"
	"github.com/spf13/cobra"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

// boxCmd represents the box command
var boxCmd = &cobra.Command{
	Use:   "box",
	Short: "plot boxplot",
	Long: `plot boxplot

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

		lineWidth := vg.Points(getFlagPositiveFloat64(cmd, "line-width") * plotConfig.scale)
		pointSize := vg.Length(getFlagPositiveFloat64(cmd, "point-size") * plotConfig.scale)
		colorIndex := getFlagPositiveInt(cmd, "color-index")
		if colorIndex > 7 {
			checkError(fmt.Errorf("unsupported color index"))
		}

		naMap := make(map[string]interface{}, len(naValues))
		for _, na := range naValues {
			naMap[strings.ToLower(na)] = struct{}{}
		}

		file := files[0]
		headerRow, fields, data, _, _, err := parseCSVfile(cmd, config, file, plotConfig.fieldStr, false, true, false)

		if err != nil {
			// if err == xopen.ErrNoContent {
			// 	log.Warningf("csvtk box: skipping empty input file: %s", file)
			// 	return
			// }
			checkError(err)
		}

		// =======================================

		horiz := getFlagBool(cmd, "horiz")
		w := vg.Length(getFlagNonNegativeFloat64(cmd, "box-width"))

		groups := make(map[string]plotter.Values)
		groupOrderMap := make(map[string]int)
		var f float64
		var ok bool
		var order int
		var groupName string
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
			if len(d) > 1 {
				groupName = d[1]
			} else { // no group, only given a field
				if len(headerRow) > 0 {
					groupName = headerRow[0]
				} else {
					groupName = ""
				}
			}
			if _, ok = groups[groupName]; !ok {
				groups[groupName] = make(plotter.Values, 0)
			}
			groups[groupName] = append(groups[groupName], f)

			if _, ok = groupOrderMap[groupName]; !ok {
				groupOrderMap[groupName] = order
				order++
			}
		}

		p := plot.New()

		var groupOrders []stringutil.StringCount
		for g := range groupOrderMap {
			groupOrders = append(groupOrders, stringutil.StringCount{Key: g, Count: groupOrderMap[g]})
		}
		sort.Sort(stringutil.StringCountList(groupOrders))

		if !horiz {
			if w == 0 {
				w = vg.Points(float64(plotConfig.width*vg.Inch) / float64(len(groupOrders)) / 2.5)
			}
		} else {
			if w == 0 {
				w = vg.Points(float64(plotConfig.height*vg.Inch) / float64(len(groupOrders)) / 2.5)
			}
		}

		groupNames := make([]string, len(groupOrders))
		j := colorIndex - 1
		for i, group := range groupOrders {
			groupNames[i] = group.Key
			b, err := plotter.NewBoxPlot(w, float64(i), groups[group.Key])
			checkError(err)
			if horiz {
				b.Horizontal = true
			}
			p.Add(b)

			b.BoxStyle.Color = plotutil.Color(j)
			b.BoxStyle.Width = lineWidth

			b.MedianStyle.Color = plotutil.Color(j)
			b.MedianStyle.Width = lineWidth

			b.WhiskerStyle.Color = plotutil.Color(j)
			b.WhiskerStyle.Width = lineWidth

			b.GlyphStyle.Color = plotutil.Color(j)
			b.GlyphStyle.Radius = pointSize
			j++
		}

		if !horiz {
			p.NominalX(groupNames...)
			// p.HideX()
		} else {
			p.NominalY(groupNames...)
			// p.HideY()
		}

		if plotConfig.ylab == "" {
			if len(headerRow) > 0 {
				plotConfig.ylab = headerRow[0]
			} else {
				plotConfig.ylab = "Values"
			}
		}
		if plotConfig.xlab == "" {
			if len(headerRow) > 1 {
				plotConfig.xlab = headerRow[1]
			} else {
				plotConfig.xlab = "Groups"
			}
		}
		if horiz {
			plotConfig.xlab, plotConfig.ylab = plotConfig.ylab, plotConfig.xlab
		}

		p.Title.Text = plotConfig.title
		p.Title.TextStyle.Font.Size = plotConfig.titleSize
		p.X.Label.Text = plotConfig.xlab
		p.Y.Label.Text = plotConfig.ylab
		p.Legend.TextStyle.Font.Size = plotConfig.labelSize
		if plotConfig.hideXlabs {
			p.X.Width = vg.Length(0)
			p.X.Tick.Width = vg.Length(0.01) // 0 would remove the space between axis and label
			p.X.Tick.Label.Font.Size = vg.Length(0)
		} else {
			p.X.Width = plotConfig.axisWidth
			p.X.Label.TextStyle.Font.Size = plotConfig.labelSize
			p.X.Tick.Width = plotConfig.tickWidth
			p.X.Tick.Label.Font.Size = plotConfig.tickLabelSize
		}
		if plotConfig.hideYlabs {
			p.Y.Tick.Width = vg.Length(0.01) // 0 would remove the space between axis and label
			p.Y.Tick.Label.Font.Size = vg.Length(0)
			p.Y.Width = vg.Length(0)
		} else {
			p.Y.Width = plotConfig.axisWidth
			p.Y.Label.TextStyle.Font.Size = plotConfig.labelSize
			p.Y.Tick.Width = plotConfig.tickWidth
			p.Y.Tick.Label.Font.Size = plotConfig.tickLabelSize
		}

		if plotConfig.scaleLnX {
			p.X.Scale = plot.LogScale{}
			p.X.Tick.Marker = plot.LogTicks{Prec: -1}
		}
		if plotConfig.scaleLnY {
			p.Y.Scale = plot.LogScale{}
			p.Y.Tick.Marker = plot.LogTicks{Prec: -1}
		}

		if plotConfig.xminStr != "" {
			log.Warning("flag --x-min ignored for command box")
		}
		if plotConfig.xmaxStr != "" {
			log.Warning("flag --x-max ignored for command box")
		}
		if plotConfig.yminStr != "" {
			log.Warning("flag --y-min ignored for command box")
		}
		if plotConfig.ymaxStr != "" {
			log.Warning("flag --y-max ignored for command box")
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

func init() {
	plotCmd.AddCommand(boxCmd)

	boxCmd.Flags().Float64P("box-width", "", 0, "box width")
	boxCmd.Flags().BoolP("horiz", "", false, "horize box plot")

	boxCmd.Flags().Float64P("line-width", "", 1.5, "line width")
	boxCmd.Flags().Float64P("point-size", "", 3, "point size")
	boxCmd.Flags().IntP("color-index", "", 1, `color index, 1-7`)
}
