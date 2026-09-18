package cmd

import (
	"bytes"
	"encoding/csv"
	"reflect"
	"strings"
	"testing"
)

func TestQuoteAllSort(t *testing.T) {
	inputText := "\"A A\",\"B\",\"C\",\"D\"\n\"1\",\"  2\",\"3  \",\"4\"\n"
	input := writeAccuracyInput(t, "quoted.csv", inputText)
	out, _ := runAccuracyCLI(t, "sort", "--quote-all", input)
	if out != inputText {
		t.Fatalf("quote-all output: got %q, want %q", out, inputText)
	}
	out, _ = runAccuracyCLI(t, "sort", input)
	if out == inputText || !reflect.DeepEqual(parseAccuracyCSV(t, out), parseAccuracyCSV(t, inputText)) {
		t.Fatalf("default output changed: %q", out)
	}
}

func TestQuoteAllAcrossCSVCommands(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\n1,2\n")
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"head", []string{"head", "--quote-all", "-n", "1", input}, "\"a\",\"b\"\n\"1\",\"2\"\n"},
		{"concat", []string{"concat", "--quote-all", input, input}, "\"a\",\"b\"\n\"1\",\"2\"\n\"1\",\"2\"\n"},
		{"custom delimiter", []string{"-D", ";", "head", "--quote-all", "-n", "1", input}, "\"a\";\"b\"\n\"1\";\"2\"\n"},
		{"tab output", []string{"-T", "head", "--quote-all", "-n", "1", input}, "\"a\"\t\"b\"\n\"1\"\t\"2\"\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, _ := runAccuracyCLI(t, tc.args...)
			if out != tc.want {
				t.Fatalf("got %q, want %q", out, tc.want)
			}
		})
	}
}

func TestCSVOutputWriterQuotedFields(t *testing.T) {
	row := []string{"", "a;b", "say \"hi\"", "line\nbreak", "a\rb", " leading", "trailing "}
	var out bytes.Buffer
	w := newCSVOutputWriter(&out, csvOutputOption{QuoteAll: true})
	w.Comma = ';'
	if err := w.Write(row); err != nil {
		t.Fatal(err)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
	want := "\"\";\"a;b\";\"say \"\"hi\"\"\";\"line\nbreak\";\"a\rb\";\" leading\";\"trailing \"\n"
	if out.String() != want {
		t.Fatalf("got %q, want %q", out.String(), want)
	}
	r := csv.NewReader(strings.NewReader(out.String()))
	r.Comma = ';'
	got, err := r.Read()
	if err != nil || !reflect.DeepEqual(got, row) {
		t.Fatalf("round trip: got %q, error %v", got, err)
	}

	out.Reset()
	w = newCSVOutputWriter(&out, csvOutputOption{QuoteAll: true})
	w.UseCRLF = true
	if err := w.Write([]string{"a\nb", "c"}); err != nil {
		t.Fatal(err)
	}
	w.Flush()
	if out.String() != "\"a\r\nb\",\"c\"\r\n" {
		t.Fatalf("CRLF output: %q", out.String())
	}
}
