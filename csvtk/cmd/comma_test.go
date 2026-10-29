package cmd

import (
	"reflect"
	"testing"
)

// TestCommaExactNumbers verifies that formatting preserves every significant digit.
func TestCommaExactNumbers(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "n\n9007199254740993\n-123456789012345678901234567890.01234567890123456789\n+1234.5000\n.0123\n12345.67e+999\n1e999999999999999999999\n1e-999999999999999999999\nNA\n")
	out, _ := runAccuracyCLI(t, "comma", input)
	want := [][]string{{"n"}, {"9,007,199,254,740,993"}, {"-123,456,789,012,345,678,901,234,567,890.01234567890123456789"}, {"+1,234.5000"}, {".0123"}, {"12,345.67e+999"}, {"1e999999999999999999999"}, {"1e-999999999999999999999"}, {"NA"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}
