package cmd

import (
	"bytes"
	"encoding/csv"
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestFixQuotesPreservesValidCSV exercises escaped quotes, quoted delimiters and empty fields.
func TestFixQuotesPreservesValidCSV(t *testing.T) {
	cells := []string{"", "a", "\"", "a\"", "\"a", "a,b", "a\"b", "a[\"b", "a|b"}
	for _, d := range []rune{',', '\t', '[', '|', '*', '\\', '(', '.'} {
		var input bytes.Buffer
		w := csv.NewWriter(&input)
		w.Comma = d
		for _, a := range cells {
			for _, b := range cells {
				if err := w.Write([]string{a, b}); err != nil {
					t.Fatal(err)
				}
			}
		}
		w.Flush()
		path := writeAccuracyInput(t, "input.csv", input.String())
		out, _ := runAccuracyCLI(t, "fix-quotes", "-d", string(d), "-b", "1M", path)
		if out != input.String() {
			t.Fatalf("delimiter %q changed valid CSV", d)
		}
	}
}

// TestFixQuotesMixedFields preserves valid quoting beside malformed quoted fields.
func TestFixQuotesMixedFields(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "# comment\n\nid,value,bad\n1,\"a\"\"b\",bare\"quote\n2,\"a,b\",\"abc\" xyz\n3,\"\"\"\",tail\"\n")
	out, _ := runAccuracyCLI(t, "fix-quotes", "-b", "1M", input)
	r := csv.NewReader(strings.NewReader(out))
	r.Comment = '#'
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"id", "value", "bad"}, {"1", "a\"b", "bare\"quote"}, {"2", "a,b", "\"abc\" xyz"}, {"3", "\"", "tail\""}}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("got %q; want %q", rows, want)
	}
	fixed := writeAccuracyInput(t, "fixed.csv", out)
	twice, _ := runAccuracyCLI(t, "fix-quotes", "-b", "1M", fixed)
	if twice != out {
		t.Fatal("repair is not idempotent")
	}
	// Keep the documented malformed TSV examples parseable with unchanged values.
	out, _ = runAccuracyCLI(t, "fix-quotes", "-t", "-b", "1M", "../../testdata/malformed.tsv")
	r = csv.NewReader(strings.NewReader(out))
	r.Comma = '\t'
	rows, err = r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 6 || rows[2][1] != "quotes\trequired" || rows[3][1] != "fake\" record" || rows[4][1] != "\"Cellvibrio\" Winogradsky" {
		t.Fatalf("documented examples: %q", rows)
	}
}

// TestFixQuotesDeleteHeader still checks field counts against a suppressed header.
func TestFixQuotesDeleteHeader(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "# comment\n\na,b\nx,y\n")
	out, _ := runAccuracyCLI(t, "fix-quotes", "-U", input)
	if out != "# comment\n\nx,y\n" {
		t.Fatalf("got %q", out)
	}
	out, _ = runAccuracyCLI(t, "fix-quotes", "-H", "-U", input)
	if out != "# comment\n\na,b\nx,y\n" {
		t.Fatalf("got %q", out)
	}
	invalid := writeAccuracyInput(t, "invalid.csv", "a,b\nx,y,z\n")
	assertInputError(t, "unequal number of fields", "fix-quotes", "-U", invalid)
}

// TestFixQuotesBufferSize validates buffer options before truncating output.
func TestFixQuotesBufferSize(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\nx,y\n")
	for _, size := range []string{"0", "-1", "NaN", "Inf", "9223372036854775808G"} {
		output := writeAccuracyInput(t, "output.csv", "sentinel")
		assertInputError(t, "buffer size", "fix-quotes", "-b", size, "-o", output, input)
		got, err := os.ReadFile(output)
		if err != nil || string(got) != "sentinel" {
			t.Fatalf("output changed: %q, %v", got, err)
		}
	}
}
