package cmd

import (
	"github.com/mattn/go-runewidth"
	"strings"
	"testing"
)

// TestCSV2RSTMultilineCells keeps each physical cell line inside aligned borders.
func TestCSV2RSTMultilineCells(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\n\"x\ny\",z\n")
	out, _ := runAccuracyCLI(t, "csv2rst", input)
	want := "+---+---+\n| a | b |\n+===+===+\n| x | z |\n| y |   |\n+---+---+\n"
	if out != want {
		t.Fatalf("got %q; want %q", out, want)
	}
	out, _ = runAccuracyCLI(t, "csv2rst", "-U", input)
	if out != "+---+---+\n| x | z |\n| y |   |\n+---+---+\n" {
		t.Fatal(out)
	}
	input = writeAccuracyInput(t, "unicode.csv", "\"标题\nA\",b\n值,z\n")
	out, _ = runAccuracyCLI(t, "csv2rst", "-p", "中", input)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	width := runewidth.StringWidth(lines[0])
	for _, line := range lines {
		if runewidth.StringWidth(line) != width {
			t.Fatalf("unequal widths: %q", out)
		}
	}
}
