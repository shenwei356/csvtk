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
	"math"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
)

// plotCmd represents the seq command
var plotCmd = &cobra.Command{
	GroupID: "plot",

	Use:   "plot",
	Short: "plot common figures",
	Long: `plot common figures

Notes:

  1. The output file can be set with flag -o/--out-file.
  2. The file format is determined by the output file suffix.
     Supported formats: eps, jpg|jpeg, pdf, png, svg, and tif|tiff
  3. If flag -o/--out-file is not set (the default), the image is written to stdout.
     You can display the image by piping it to the "display" command from ImageMagick
     or redirect it to a file.
  4. Field selectors must identify a single column; ranges are not supported.
     Image dimensions and explicit axis limits must be finite.

`,
}

func init() {
	RootCmd.AddCommand(plotCmd)

	plotCmd.PersistentFlags().StringP("data-field", "f", "1", `single column index or column name of data`)
	plotCmd.PersistentFlags().StringP("group-field", "g", "", `single column index or column name of group`)

	plotCmd.PersistentFlags().StringP("title", "", "", "Figure title")
	plotCmd.PersistentFlags().StringP("xlab", "", "", "x label text")
	plotCmd.PersistentFlags().StringP("ylab", "", "", "y label text")

	plotCmd.PersistentFlags().StringP("x-min", "", "", `minimum value of X axis`)
	plotCmd.PersistentFlags().StringP("x-max", "", "", `maximum value of X axis`)
	plotCmd.PersistentFlags().StringP("y-min", "", "", `minimum value of Y axis`)
	plotCmd.PersistentFlags().StringP("y-max", "", "", `maximum value of Y axis`)

	plotCmd.PersistentFlags().Float64P("width", "", 6, "Figure width")
	plotCmd.PersistentFlags().Float64P("height", "", 4.5, "Figure height")

	plotCmd.PersistentFlags().IntP("title-size", "", 16, "title font size")
	plotCmd.PersistentFlags().IntP("label-size", "", 14, "label font size")
	plotCmd.PersistentFlags().Float64P("axis-width", "", 1.5, "axis width")
	plotCmd.PersistentFlags().Float64P("tick-width", "", 1.5, "axis tick width")
	plotCmd.PersistentFlags().Float64P("tick-label-size", "", 12, "tick label font size")
	plotCmd.PersistentFlags().Float64P("scale", "", 1, "scale the image width/height, tick, axes, line/point and font sizes proportionally")

	plotCmd.PersistentFlags().BoolP("hide-x-labs", "", false, "hide X axis, ticks, and tick labels")
	plotCmd.PersistentFlags().BoolP("hide-y-labs", "", false, "hide Y axis, ticks, and tick labels")

	plotCmd.PersistentFlags().BoolP("x-scale-ln", "", false, "scale the X axis by the natural log")
	plotCmd.PersistentFlags().BoolP("y-scale-ln", "", false, "scale the Y axis by the natural log")

	plotCmd.PersistentFlags().StringP("format", "", "png", `image format for stdout when flag -o/--out-file is not given. Available values: eps, jpg|jpeg, pdf, png, svg, and tif|tiff.`)

	plotCmd.PersistentFlags().StringSliceP("na-values", "", []string{"", "NA", "N/A"}, `NA values, case ignored`)
	plotCmd.PersistentFlags().BoolP("skip-na", "", false, "skip NA values in --na-values")

}

func getPlotConfigs(cmd *cobra.Command) *plotConfigs {
	config := new(plotConfigs)

	config.dataFieldStr = getFlagString(cmd, "data-field")
	checkError(validatePlotField(config.dataFieldStr, "data-field"))

	config.groupFieldStr = getFlagString(cmd, "group-field")
	if len(config.groupFieldStr) > 0 {
		checkError(validatePlotField(config.groupFieldStr, "group-field"))
		config.fieldStr = config.dataFieldStr + "," + config.groupFieldStr
	} else {
		config.fieldStr = config.dataFieldStr
	}

	config.title = getFlagString(cmd, "title")

	scale := getFlagPositiveFloat64(cmd, "scale")
	config.scale = scale

	config.titleSize = vg.Length(int(float64(getFlagPositiveInt(cmd, "title-size")) * scale))
	config.labelSize = vg.Length(int(float64(getFlagPositiveInt(cmd, "label-size")) * scale))
	config.width = vg.Length(getFlagPositiveFloat64(cmd, "width") * scale)
	config.height = vg.Length(getFlagPositiveFloat64(cmd, "height") * scale)
	config.axisWidth = vg.Length(getFlagNonNegativeFloat64(cmd, "axis-width") * scale)
	config.tickWidth = vg.Length(getFlagNonNegativeFloat64(cmd, "tick-width") * scale)
	config.tickLabelSize = vg.Length(int(float64(getFlagNonNegativeFloat64(cmd, "tick-label-size")) * scale))
	config.hideXlabs = getFlagBool(cmd, "hide-x-labs")
	config.hideYlabs = getFlagBool(cmd, "hide-y-labs")

	config.xlab = getFlagString(cmd, "xlab")
	config.ylab = getFlagString(cmd, "ylab")

	config.scaleLnX = getFlagBool(cmd, "x-scale-ln")
	config.scaleLnY = getFlagBool(cmd, "y-scale-ln")

	var err error

	config.xminStr = getFlagString(cmd, "x-min")
	if config.xminStr != "" {
		config.xmin, err = strconv.ParseFloat(config.xminStr, 64)
		if err != nil || math.IsNaN(config.xmin) || math.IsInf(config.xmin, 0) {
			checkError(fmt.Errorf("value of flag --%s should be finite", "x-min"))
		}
	}
	config.xmaxStr = getFlagString(cmd, "x-max")
	if config.xmaxStr != "" {
		config.xmax, err = strconv.ParseFloat(config.xmaxStr, 64)
		if err != nil || math.IsNaN(config.xmax) || math.IsInf(config.xmax, 0) {
			checkError(fmt.Errorf("value of flag --%s should be finite", "x-max"))
		}
	}
	config.yminStr = getFlagString(cmd, "y-min")
	if config.yminStr != "" {
		config.ymin, err = strconv.ParseFloat(config.yminStr, 64)
		if err != nil || math.IsNaN(config.ymin) || math.IsInf(config.ymin, 0) {
			checkError(fmt.Errorf("value of flag --%s should be finite", "y-min"))
		}
	}
	config.ymaxStr = getFlagString(cmd, "y-max")
	if config.ymaxStr != "" {
		config.ymax, err = strconv.ParseFloat(config.ymaxStr, 64)
		if err != nil || math.IsNaN(config.ymax) || math.IsInf(config.ymax, 0) {
			checkError(fmt.Errorf("value of flag --%s should be finite", "y-max"))
		}
	}

	if config.xminStr != "" && config.xmaxStr != "" && config.xmin >= config.xmax {
		checkError(fmt.Errorf("--x-min should be less than --x-max"))
	}
	if config.yminStr != "" && config.ymaxStr != "" && config.ymin >= config.ymax {
		checkError(fmt.Errorf("--y-min should be less than --y-max"))
	}

	config.format = strings.ToLower(getFlagString(cmd, "format"))
	switch config.format {
	case "eps", "jpg", "jpeg", "pdf", "png", "svg", "tif", "tiff":
	default:
		checkError(fmt.Errorf("invalid image format. available format: eps, jpg|jpeg, pdf, png, svg, and tif|tiff"))
	}

	return config
}

type plotConfigs struct {
	dataFieldStr, groupFieldStr, fieldStr string
	title, xlab, ylab                     string
	titleSize, labelSize, tickLabelSize   vg.Length
	width, height, axisWidth, tickWidth   vg.Length
	hideXlabs, hideYlabs                  bool
	scale                                 float64
	scaleLnX, scaleLnY                    bool
	xmin, xmax, ymin, ymax                float64
	xminStr, xmaxStr, yminStr, ymaxStr    string
	format                                string
}

// validatePlotField rejects selections that would shift data and group positions.
func validatePlotField(spec, flag string) error {
	if spec == "" {
		return fmt.Errorf("flag --%s needs a column index or name", flag)
	}
	if strings.Contains(spec, ",") || reIntegerRange.MatchString(spec) {
		return fmt.Errorf("only one field allowed for flag --%s", flag)
	}
	if spec[0] == '-' {
		return fmt.Errorf("unselect not allowed for flag --%s", flag)
	}
	return nil
}

// preparePlotAxes validates ranges before an output file is opened.
// Padding must remain representable and preserve positive logarithmic domains.
func preparePlotAxes(p *plot.Plot) error {
	for _, item := range []struct {
		name string
		axis *plot.Axis
	}{{"X", &p.X}, {"Y", &p.Y}} {
		axis := item.axis
		if _, logarithmic := axis.Scale.(plot.LogScale); !logarithmic {
			if math.IsNaN(axis.Min) || math.IsNaN(axis.Max) || math.IsInf(axis.Min, 0) || math.IsInf(axis.Max, 0) || math.IsInf(axis.Max-axis.Min, 0) {
				return fmt.Errorf("%s axis requires finite bounds and range", item.name)
			}
			if axis.Min == axis.Max {
				// Preserve the library's unit padding unless rounding would erase it.
				padding := 1.0
				if axis.Min-padding == axis.Min || axis.Max+padding == axis.Max {
					padding = math.Abs(axis.Min) * 0.05
				}
				if lower := axis.Min - padding; !math.IsInf(lower, 0) {
					axis.Min = lower
				}
				if upper := axis.Max + padding; !math.IsInf(upper, 0) {
					axis.Max = upper
				}
			}
			continue
		}
		if !(axis.Min > 0) || !(axis.Max > 0) || math.IsInf(axis.Min, 0) || math.IsInf(axis.Max, 0) {
			return fmt.Errorf("%s logarithmic axis requires positive finite bounds and data", item.name)
		}
		if axis.Min == axis.Max {
			if lower := axis.Min / 2; lower > 0 {
				axis.Min = lower
			}
			if upper := axis.Max * 2; !math.IsInf(upper, 0) {
				axis.Max = upper
			}
		}
	}
	return nil
}
