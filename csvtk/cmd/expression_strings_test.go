package cmd

import (
	"encoding/csv"
	"reflect"
	"strings"
	"testing"
)

// TestExpressionStringValues checks that field contents survive substitution
// literally, including tokens used internally for numeric column references.
func TestExpressionStringValues(t *testing.T) {
	values := []string{
		`$x ${x} shenwei2 ${2} shenwei_2x`,
		`both 'single' and "double" quotes`,
		`C:\path\file`,
		`trailing\`,
		`\'\"`,
		`literal\n\t\x41`,
		"line one\nline two\t沈伟\rend",
		"control\x00\x01\a\b\f\v",
	}
	selectors := []struct {
		name, long, short, expected string
		header                      []string
	}{
		{"names", "$long", "$x", "$expected", []string{"long", "x", "expected"}},
		{"shared prefix", "$xlong", "$x", "$expected", []string{"xlong", "x", "expected"}},
		{"braced names", "${long}", "${x}", "${expected}", []string{"long", "x", "expected"}},
		{"numbers", "$1", "$2", "$3", []string{"long", "x", "expected"}},
		{"braced numbers", "${1}", "${2}", "${3}", []string{"long", "x", "expected"}},
		{"names starting with digits", "$2long", "$2x", "$expected", []string{"2long", "2x", "expected"}},
		{"no header", "$1", "$2", "$3", nil},
	}
	for _, selector := range selectors {
		t.Run(selector.name, func(t *testing.T) {
			var rows [][]string
			if selector.header != nil {
				rows = append(rows, selector.header)
			}
			for _, value := range values {
				rows = append(rows, []string{value, "z", value + "z" + value})
			}
			var inputText strings.Builder
			writer := csv.NewWriter(&inputText)
			if err := writer.WriteAll(rows); err != nil {
				t.Fatal(err)
			}
			input := writeAccuracyInput(t, "input.csv", inputText.String())
			expression := selector.long + " + " + selector.short + " + " + selector.long
			for _, command := range []string{"filter2", "mutate2", "mutate3"} {
				t.Run(command, func(t *testing.T) {
					args := []string{command}
					if selector.header == nil {
						args = append(args, "-H")
					}
					var want [][]string
					if command == "filter2" {
						args = append(args, "-f", expression+" == "+selector.expected)
						want = rows
					} else {
						args = append(args, "-n", "result", "-e", expression)
						for i, row := range rows {
							value := row[2]
							if i == 0 && selector.header != nil {
								value = "result"
							}
							want = append(want, append(append([]string(nil), row...), value))
						}
					}
					out, stderr := runAccuracyCLI(t, append(args, input)...)
					if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
						t.Fatalf("got %q, want %q\nstderr: %s", got, want, stderr)
					}
				})
			}
		})
	}
}

// TestExpressionStringNullCoalescence checks empty-field fallback when values
// contain selected variable references.
func TestExpressionStringNullCoalescence(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "long,x,expected\n,$long,$long\n$x,z,$x\n,${long},${long}\n${x},z,${x}\n")
	for _, command := range []string{"filter2", "mutate2", "mutate3"} {
		t.Run(command, func(t *testing.T) {
			args := []string{command}
			var want [][]string
			if command == "filter2" {
				args = append(args, "-f", "($long ?? $x) == $expected")
				want = [][]string{{"long", "x", "expected"}, {"", "$long", "$long"}, {"$x", "z", "$x"}, {"", "${long}", "${long}"}, {"${x}", "z", "${x}"}}
			} else {
				args = append(args, "-n", "result", "-e", "$long ?? $x")
				want = [][]string{{"long", "x", "expected", "result"}, {"", "$long", "$long", "$long"}, {"$x", "z", "$x", "$x"}, {"", "${long}", "${long}", "${long}"}, {"${x}", "z", "${x}", "${x}"}}
			}
			out, _ := runAccuracyCLI(t, append(args, input)...)
			if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}
