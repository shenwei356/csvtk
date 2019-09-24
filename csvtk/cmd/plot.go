// Copyright © 2016-2019 Wei Shen <shenwei356@gmail.com>
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
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gonum.org/v1/plot/vg"
)

// plotCmd represents the seq command
var plotCmd = &cobra.Command{
	Use:   "plot",
	Short: "plot common figures",
	Long: `plot common figures

Notes:

  1. Output file can be set by flag -o/--out-file.
  2. File format is determined by the out file suffix.
     Supported formats: eps, jpg|jpeg, pdf, png, svg, and tif|tiff
  3. If flag -o/--out-file not set (default), image is written to stdout,
     you can display the image by pipping to "display" command of Imagemagic
     or just redirect to file.

`,
}

func init() {
	RootCmd.AddCommand(plotCmd)

	plotCmd.PersistentFlags().StringP("data-field", "f", "1", `column index or column name of data`)
	plotCmd.PersistentFlags().StringP("group-field", "g", "", `column index or column name of group`)

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

	plotCmd.PersistentFlags().StringP("format", "", "png", `image format for stdout when flag -o/--out-file not given. available values: eps, jpg|jpeg, pdf, png, svg, and tif|tiff.`)

}

func getPlotConfigs(cmd *cobra.Command) *plotConfigs {
	config := new(plotConfigs)

	config.dataFieldStr = getFlagString(cmd, "data-field")
	if strings.Index(config.dataFieldStr, ",") >= 0 {
		checkError(fmt.Errorf("only one field allowed for flag --data-field"))
	}
	if config.dataFieldStr[0] == '-' {
		checkError(fmt.Errorf("unselect not allowed for flag --data-field"))
	}

	config.groupFieldStr = getFlagString(cmd, "group-field")
	if len(config.groupFieldStr) > 0 {
		if strings.Index(config.groupFieldStr, ",") >= 0 {
			checkError(fmt.Errorf("only one field allowed for flag --group-field"))
		}
		if config.groupFieldStr[0] == '-' {
			checkError(fmt.Errorf("unselect not allowed for flag --group-field"))
		}
		config.fieldStr = config.dataFieldStr + "," + config.groupFieldStr
	} else {
		config.fieldStr = config.dataFieldStr
	}

	config.title = getFlagString(cmd, "title")
	config.titleSize = vg.Length(getFlagPositiveInt(cmd, "title-size"))
	config.labelSize = vg.Length(getFlagPositiveInt(cmd, "label-size"))
	config.width = vg.Length(getFlagPositiveFloat64(cmd, "width"))
	config.height = vg.Length(getFlagPositiveFloat64(cmd, "height"))
	config.axisWidth = vg.Length(getFlagPositiveFloat64(cmd, "axis-width"))
	config.tickWidth = vg.Length(getFlagPositiveFloat64(cmd, "tick-width"))
	config.xlab = getFlagString(cmd, "xlab")
	config.ylab = getFlagString(cmd, "ylab")

	var err error

	config.xminStr = getFlagString(cmd, "x-min")
	if config.xminStr != "" {
		config.xmin, err = strconv.ParseFloat(config.xminStr, 64)
		if err != nil {
			checkError(fmt.Errorf("value of flag --%s should be float", "x-min"))
		}
	}
	config.xmaxStr = getFlagString(cmd, "x-max")
	if config.xmaxStr != "" {
		config.xmax, err = strconv.ParseFloat(config.xmaxStr, 64)
		if err != nil {
			checkError(fmt.Errorf("value of flag --%s should be float", "x-max"))
		}
	}
	config.yminStr = getFlagString(cmd, "y-min")
	if config.yminStr != "" {
		config.ymin, err = strconv.ParseFloat(config.yminStr, 64)
		if err != nil {
			checkError(fmt.Errorf("value of flag --%s should be float", "y-min"))
		}
	}
	config.ymaxStr = getFlagString(cmd, "y-max")
	if config.ymaxStr != "" {
		config.ymax, err = strconv.ParseFloat(config.ymaxStr, 64)
		if err != nil {
			checkError(fmt.Errorf("value of flag --%s should be float", "y-max"))
		}
	}

	config.format = getFlagString(cmd, "format")
	switch strings.ToLower(config.format) {
	case "eps", "jpg", "jpeg", "pdf", "png", "svg", "tif", "tiff":
	default:
		checkError(fmt.Errorf("invalid image format. available format: eps, jpg|jpeg, pdf, png, svg, and tif|tiff"))
	}

	return config
}

type plotConfigs struct {
	dataFieldStr, groupFieldStr, fieldStr string
	title, xlab, ylab                     string
	titleSize, labelSize                  vg.Length
	width, height, axisWidth, tickWidth   vg.Length
	xmin, xmax, ymin, ymax                float64
	xminStr, xmaxStr, yminStr, ymaxStr    string
	format                                string
}
