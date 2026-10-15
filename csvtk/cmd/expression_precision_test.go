package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// TestExpressionNumericPrecision checks tiny, subnormal, and large values through
// both column-selection branches and both forms of variable references.
func TestExpressionNumericPrecision(t *testing.T) {
	values := [][]string{
		{"1e-20", "2e-20", "0.50"},
		{"-1e-20", "2e-20", "-0.50"},
		{"5e-324", "5e-324", "1.00"},
		{"-5e-324", "5e-324", "-1.00"},
		{"0.00000000000000012", "0.00000000000000006", "2.00"},
		{"1e20", "2e20", "0.50"},
		{"1.7976931348623157e308", "1.7976931348623157e308", "1.00"},
		{"0", "1", "0.00"},
	}
	var inputText strings.Builder
	for _, row := range values {
		inputText.WriteString(row[0] + "," + row[1] + "\n")
	}
	withHeader := writeAccuracyInput(t, "header.csv", "x,y\n"+inputText.String())
	withoutHeader := writeAccuracyInput(t, "data.csv", inputText.String())
	selectors := []struct {
		name     string
		x, y     string
		noHeader bool
	}{
		{"names", "$x", "$y", false},
		{"numbers", "$1", "$2", false},
		{"braced names", "${x}", "${y}", false},
		{"braced numbers", "${1}", "${2}", false},
		{"no header", "$1", "$2", true},
	}
	for _, selector := range selectors {
		t.Run(selector.name, func(t *testing.T) {
			input := withHeader
			var prefix []string
			if selector.noHeader {
				input = withoutHeader
				prefix = []string{"-H"}
			}
			t.Run("filter2", func(t *testing.T) {
				args := append(append([]string(nil), prefix...), "filter2", "-f", selector.x+" != 0", input)
				out, _ := runAccuracyCLI(t, args...)
				var want [][]string
				if !selector.noHeader {
					want = append(want, []string{"x", "y"})
				}
				for _, row := range values[:len(values)-1] {
					want = append(want, row[:2])
				}
				if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
					t.Fatalf("got %q, want %q", got, want)
				}
			})
			for _, command := range []string{"mutate2", "mutate3"} {
				t.Run(command, func(t *testing.T) {
					args := append(append([]string(nil), prefix...), command, "-n", "ratio", "-e", selector.x+" / "+selector.y, input)
					out, _ := runAccuracyCLI(t, args...)
					var want [][]string
					if !selector.noHeader {
						want = append(want, []string{"x", "y", "ratio"})
					}
					want = append(want, values...)
					if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
						t.Fatalf("got %q, want %q", got, want)
					}
				})
			}
		})
	}
}

// TestExpressionNumericStrings preserves original field text when numeric
// conversion is disabled explicitly or by the legacy string functions.
func TestExpressionNumericStrings(t *testing.T) {
	const text = "x\n1e-20\n"
	input := writeAccuracyInput(t, "input.csv", text)
	for _, command := range []string{"mutate2", "mutate3"} {
		t.Run(command, func(t *testing.T) {
			out, _ := runAccuracyCLI(t, command, "-s", "-n", "copy", "-e", "$x", input)
			want := [][]string{{"x", "copy"}, {"1e-20", "1e-20"}}
			if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
	out, _ := runAccuracyCLI(t, "filter2", "-s", "-f", `$x == "1e-20"`, input)
	if out != text {
		t.Fatalf("numeric string filter: got %q, want %q", out, text)
	}
	out, _ = runAccuracyCLI(t, "filter2", "-f", "len($x) == 5", input)
	if out != text {
		t.Fatalf("string function filter: got %q, want %q", out, text)
	}
	out, _ = runAccuracyCLI(t, "mutate2", "-n", "length", "-e", "len($x)", input)
	want := [][]string{{"x", "length"}, {"1e-20", "5.00"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("string function: got %q, want %q", got, want)
	}
}

// TestMutate3NumericResultFormat ensures whole-valued numeric fields remain
// floats, preserving output rounding and avoiding integer literal overflow.
func TestMutate3NumericResultFormat(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "x\n2\n100000000000000000000\n")
	out, _ := runAccuracyCLI(t, "mutate3", "-n", "result", "-e", "$x + 1", input)
	want := [][]string{{"x", "result"}, {"2", "3.00"}, {"100000000000000000000", "100000000000000000000.00"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestExpressionNumericOverflow requires an explicit conversion error for input
// outside the finite float64 range, before it reaches either expression parser.
func TestExpressionNumericOverflow(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "x\n1e309\n")
	for _, command := range []string{"filter2", "mutate2", "mutate3"} {
		for _, variable := range []string{"$x", "$1"} {
			t.Run(command+"/"+variable, func(t *testing.T) {
				args := []string{command}
				if command == "filter2" {
					args = append(args, "-f", variable+" > 0")
				} else {
					args = append(args, "-n", "result", "-e", variable)
				}
				args = append(args, input)
				cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
				cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
				var stderr bytes.Buffer
				cmd.Stderr = &stderr
				if err := cmd.Run(); err == nil {
					t.Fatal("numeric overflow did not fail")
				} else if _, ok := err.(*exec.ExitError); !ok {
					t.Fatalf("could not run csvtk: %v", err)
				}
				if !strings.Contains(stderr.String(), "value out of range") {
					t.Fatalf("missing numeric overflow error: %s", stderr.String())
				}
			})
		}
	}
}
