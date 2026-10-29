package cmd

import (
	"strings"
	"testing"
)

// TestCSV2MDCellBoundaries keeps literal pipes and multiline values within one cell.
func TestCSV2MDCellBoundaries(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\n\"x|y\nz\",q\n")
	out, _ := runAccuracyCLI(t, "csv2md", input)
	want := "|a        |b  |\n|:--------|:--|\n|x\\|y<br>z|q  |\n"
	if out != want {
		t.Fatalf("got %q; want %q", out, want)
	}
	input = writeAccuracyInput(t, "header.csv", "\"a|b\nc\",value\nx\\|y,z\n")
	out, _ = runAccuracyCLI(t, "csv2md", input)
	if !strings.Contains(out, "a\\|b<br>c") || !strings.Contains(out, `x\\\|y`) || strings.Count(out, "\n") != 3 {
		t.Fatalf("escaped table: %q", out)
	}
}
