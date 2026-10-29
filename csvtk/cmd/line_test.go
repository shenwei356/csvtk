package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLineNominalTicks covers a single category and wider user-defined axis limits.
func TestLineNominalTicks(t *testing.T) {
	for _, data := range []string{"x,y\nAlpha,1\n", "x,y\nAlpha,1\nBeta,2\n"} {
		for _, limits := range [][]string{nil, {"--x-min", "-10", "--x-max", "10"}} {
			input := writeAccuracyInput(t, "input.csv", data)
			output := filepath.Join(t.TempDir(), "line.svg")
			args := append([]string{"plot", "line", "-x", "1", "-y", "2", "--data-field-x-nominal", "-o", output, input}, limits...)
			runAccuracyCLI(t, args...)
			svg, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(svg), "<svg") || strings.Count(string(svg), ">Alpha</text>") != 1 {
				t.Fatalf("invalid category ticks: %s", svg)
			}
			if strings.Contains(data, "Beta") && strings.Count(string(svg), ">Beta</text>") != 1 {
				t.Fatal("missing or duplicated Beta tick")
			}
		}
	}
}
