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
	"slices"
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

// barCmd represents the bar command
var barCmd = &cobra.Command{
	Use:   "bar",
	Short: "bar chart",
	Long: `bar chart

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

		barWidth := vg.Points(getFlagFloat64(cmd, "bar-width") * plotConfig.scale)
		colorIndex := getFlagPositiveInt(cmd, "color-index")
		if colorIndex > 7 {
			checkError(fmt.Errorf("unsupported color index"))
		}

		dataFieldXStr := getFlagString(cmd, "data-field-x")
		if dataFieldXStr == "" {
			checkError(fmt.Errorf("flag -x (--data-field-x) needed"))
		}
		if strings.Contains(dataFieldXStr, ",") {
			checkError(fmt.Errorf("only one field allowed for flag -x (--data-field-x)"))
		}
		if dataFieldXStr[0] == '-' {
			checkError(fmt.Errorf("unselect not allowed for flag -y (--data-field-x)"))
		}

		dataFieldYStr := getFlagString(cmd, "data-field-y")
		if dataFieldYStr == "" {
			checkError(fmt.Errorf("flag -y (--data-field-y) needed"))
		}
		if strings.Contains(dataFieldYStr, ",") {
			checkError(fmt.Errorf("only one field allowed for flag -y (--data-field-y)"))
		}
		if dataFieldXStr[0] == '-' {
			checkError(fmt.Errorf("unselect not allowed for flag -y (--data-field-y)"))
		}

		groupFieldStr := getFlagString(cmd, "group-field")
		if len(groupFieldStr) > 0 {
			if strings.Contains(groupFieldStr, ",") {
				checkError(fmt.Errorf("only one field allowed for flag --group-field"))
			}
			if groupFieldStr[0] == '-' {
				checkError(fmt.Errorf("unselect not allowed for flag --group-field"))
			}
			plotConfig.fieldStr = dataFieldXStr + "," + dataFieldYStr + "," + groupFieldStr
		} else {
			plotConfig.fieldStr = dataFieldXStr + "," + dataFieldYStr
		}

		horizontal := getFlagBool(cmd, "horizontal")
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
			// 	log.Warningf("csvtk box: skipping empty input file: %s", file)
			// 	return
			// }
			checkError(err)
		}

		var xNominalValues []string
		// Collect unique values
		xNominalValues = make([]string, 0, len(data)/4) // Assume a quarter of the data is unique
		for _, d := range data {
			i, found := slices.BinarySearch(xNominalValues, d[0])
			if !found {
				// 0 alloc insert slice trick https://go.dev/wiki/SliceTricks#insert
				xNominalValues = append(xNominalValues, "")
				copy(xNominalValues[i+1:], xNominalValues[i:])
				xNominalValues[i] = d[0]
			}
		}

		// =======================================

		groups := make(map[string]plotter.Values)
		groupOrderMap := make(map[string]int)
		var y float64
		var ok bool
		var order int

		for _, d := range data {
			if skipNA {
				if _, ok = naMap[strings.ToLower(d[0])]; ok {
					continue
				}
			}

			if len(d) > 1 {
				if skipNA {
					if _, ok = naMap[strings.ToLower(d[1])]; ok {
						continue
					}
				}
				y, err = strconv.ParseFloat(d[1], 64)
			} else {
				y, err = strconv.ParseFloat(d[0], 64)
			}
			if err != nil {
				if len(headerRow) > 0 {
					checkError(fmt.Errorf("fail to parse Y value: %s at column: %s. please choose the right column by flag --data-field-y", d[1], headerRow[1]))
				} else {
					checkError(fmt.Errorf("fail to parse Y value: %s at column: %d. please choose the right column by flag --data-field-y", d[1], fields[1]))
				}
			}

			var groupName string
			if len(d) > 2 {
				groupName = d[2]
			}

			if _, ok = groups[groupName]; !ok {
				groups[groupName] = make(plotter.Values, 0)
			}
			groups[groupName] = append(groups[groupName], y)
			if _, ok = groupOrderMap[groupName]; !ok {
				groupOrderMap[groupName] = order
				order++
			}
		}

		if barWidth == 0 {
			fullWidth := plotConfig.width * vg.Inch
			if horizontal {
				fullWidth = plotConfig.height * vg.Inch
			}
			fullWidth -= plotConfig.axisWidth * vg.Inch // leave space for axis

			barWidth = vg.Points(
				(float64(fullWidth) / float64(len(groups)*len(xNominalValues))),
			)
		}

		p := plot.New()

		var groupOrders []stringutil.StringCount
		for g := range groupOrderMap {
			groupOrders = append(groupOrders, stringutil.StringCount{Key: g, Count: groupOrderMap[g]})
		}
		sort.Sort(stringutil.StringCountList(groupOrders))

		addLegend := len(groupOrders) > 1
		for i, gor := range groupOrders {
			v := groups[gor.Key]
			g := gor.Key

			bars, err := plotter.NewBarChart(v, barWidth)
			checkError(err)
			bars.LineStyle.Width = vg.Length(0)
			bars.Color = plotutil.Color(colorIndex - 1 + i)
			bars.Horizontal = horizontal

			// Calculate offset to center the bars
			bars.Offset = barWidth * vg.Length(float64(i)-(float64(len(groupOrders)-1)/2))

			p.Add(bars)
			if addLegend {
				p.Legend.Add(g, bars)
			}
		}

		p.Legend.Top = getFlagBool(cmd, "legend-top")
		p.Legend.Left = getFlagBool(cmd, "legend-left")
		p.Legend.Padding = 1.5

		if horizontal {
			p.X.Width = plotConfig.axisWidth
			if plotConfig.xlab == "" {
				if len(headerRow) > 1 {
					plotConfig.xlab = headerRow[1]
				} else {
					plotConfig.xlab = "Values"
				}
			}
			p.NominalY(xNominalValues...)
		} else {
			p.Y.Width = plotConfig.axisWidth
			if plotConfig.ylab == "" {
				if len(headerRow) > 1 {
					plotConfig.ylab = headerRow[1]
				} else {
					plotConfig.ylab = "Values"
				}
			}
			p.NominalX(xNominalValues...)
		}

		p.Title.Text = plotConfig.title
		p.Title.TextStyle.Font.Size = plotConfig.titleSize
		p.X.Label.Text = plotConfig.xlab
		p.Y.Label.Text = plotConfig.ylab
		p.X.Label.TextStyle.Font.Size = plotConfig.labelSize
		p.Y.Label.TextStyle.Font.Size = plotConfig.labelSize
		p.X.Tick.Width = plotConfig.tickWidth
		p.Y.Tick.Width = plotConfig.tickWidth
		p.X.Tick.Label.Font.Size = plotConfig.tickLabelSize
		p.Y.Tick.Label.Font.Size = plotConfig.tickLabelSize

		// TODO log scale
		// if plotConfig.scaleLnX {
		// 	p.X.Scale = plot.LogScale{}
		// 	p.X.Tick.Marker = plot.LogTicks{Prec: -1}
		// }
		// if plotConfig.scaleLnY {
		// 	p.Y.Scale = plot.LogScale{}
		// 	p.Y.Tick.Marker = plot.LogTicks{Prec: -1}
		// }

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

func init() {
	plotCmd.AddCommand(barCmd)
	barCmd.Flags().StringP("data-field-x", "x", "", `column index or column name of X for command bar`)
	barCmd.Flags().StringP("data-field-y", "y", "", `column index or column name of Y for command bar`)

	barCmd.Flags().BoolP("legend-top", "", false, "locate legend along the top edge of the plot")
	barCmd.Flags().BoolP("legend-left", "", false, "locate legend along the left edge of the plot")

	barCmd.Flags().Float64P("bar-width", "", 0, "bar width (0 for auto)")
	barCmd.Flags().IntP("color-index", "", 1, `color index, 1-7`)
	barCmd.Flags().BoolP("horizontal", "", false, "horizontal bar chart")
}
