package cmd

import (
	"os"
	"strings"
	"testing"
)

// TestPlotInvalidFields rejects ambiguous selections before rendering.
func TestPlotInvalidFields(t *testing.T) {
	input := writeAccuracyInput(t, "plot.csv", "x,y,z\n1,2,3\n4,5,6\n")
	for _, tc := range []struct {
		command string
		args    []string
		flag    string
	}{
		{"box", []string{"-f", ""}, "data-field"},
		{"box", []string{"-f", "1-2"}, "data-field"},
		{"box", []string{"-g", "2-"}, "group-field"},
		{"line", []string{"-x", "1-2", "-y", "3"}, "data-field-x"},
		{"bar", []string{"-x", "1-2", "-y", "3"}, "data-field-x"},
		{"bar", []string{"-x", "1", "-y", "-2"}, "data-field-y"},
		{"line", []string{"-x", "1", "-y", "-2"}, "data-field-y"},
	} {
		args := append([]string{"plot", tc.command}, tc.args...)
		assertInputError(t, tc.flag, append(args, input)...)
	}
}

// TestPlotLargeConstantRange needs padding that changes the floating-point bounds.
func TestPlotLargeConstantRange(t *testing.T) {
	input := writeAccuracyInput(t, "plot.csv", "x,y\n1,1e100\n2,1e100\n")
	out, _ := runAccuracyCLI(t, "plot", "line", "-x", "1", "-y", "2", "--format", "svg", input)
	if !strings.Contains(out, "<svg") || strings.Contains(out, "NaN") || strings.Contains(out, "+Inf") {
		t.Fatal("invalid SVG for a large constant range")
	}
}

// TestPlotNonFiniteOptions preserves outputs when size or axis options are invalid.
func TestPlotNonFiniteOptions(t *testing.T) {
	input := writeAccuracyInput(t, "plot.csv", "x,y\n1,2\n3,4\n")
	for _, flag := range []string{"width", "axis-width", "x-min", "x-max", "y-min", "y-max"} {
		for _, value := range []string{"NaN", "Inf", "-Inf"} {
			output := writeAccuracyInput(t, "output.svg", "sentinel")
			assertInputError(t, flag, "plot", "line", "-x", "1", "-y", "2", "--"+flag, value, "-o", output, input)
			data, err := os.ReadFile(output)
			if err != nil || string(data) != "sentinel" {
				t.Fatalf("output changed for --%s %s: %q, %v", flag, value, data, err)
			}
		}
	}
}

// TestPlotLogAxes diagnoses invalid domains and renders a positive constant range.
func TestPlotLogAxes(t *testing.T) {
	input := writeAccuracyInput(t, "plot.csv", "x,y\n-1,-2\n1,2\n")
	for _, tc := range []struct {
		command string
		args    []string
	}{
		{"line", []string{"-x", "1", "-y", "2", "--y-scale-ln"}},
		{"line", []string{"-x", "1", "-y", "2", "--y-scale-ln", "--y-min", "0.1"}},
		{"box", []string{"-f", "2", "--y-scale-ln"}},
		{"hist", []string{"-f", "1", "--x-scale-ln"}},
	} {
		output := writeAccuracyInput(t, "plot.svg", "sentinel")
		args := append([]string{"plot", tc.command}, tc.args...)
		assertInputError(t, "logarithmic axis", append(args, "-o", output, input)...)
		data, err := os.ReadFile(output)
		if err != nil || string(data) != "sentinel" {
			t.Fatalf("output changed: %q, %v", data, err)
		}
	}
	constant := writeAccuracyInput(t, "constant.csv", "x,y\n0.1,0.2\n0.1,0.2\n")
	out, _ := runAccuracyCLI(t, "plot", "line", "-x", "1", "-y", "2", "--x-scale-ln", "--y-scale-ln", "--format", "svg", constant)
	if !strings.Contains(out, "<svg") || strings.Contains(out, "NaN") {
		t.Fatalf("invalid SVG: %s", out)
	}
}

// TestPlotHistogramLogCounts renders zero-count bins on a logarithmic count axis.
func TestPlotHistogramLogCounts(t *testing.T) {
	input := writeAccuracyInput(t, "hist.csv", "v\n1\n1\n9\n")
	out, _ := runAccuracyCLI(t, "plot", "hist", "--y-scale-ln", "--bins", "5", "--format", "SVG", input)
	if !strings.Contains(out, "<svg") || strings.Contains(out, "NaN") {
		t.Fatalf("invalid SVG: %s", out)
	}
}
