package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestExpressionChangingTypes covers reusing programs across numeric, text, and nil rows.
func TestExpressionChangingTypes(t *testing.T) {
	for _, tc := range []struct {
		name, input, expression string
		values                  []string
	}{
		{"addition", "x,y\n1,2\na,b\n4,5\n", "$x + $y", []string{"3.00", "ab", "9.00"}},
		{"coalescence", "x,y\n,1\n2,3\n,4\ntext,backup\n", "$x ?? $y", []string{"1.00", "2.00", "4.00", "text"}},
		{"quoted_reference", "x,y\n1,2\n3,4\n", "'$x'", []string{"1.0", "3.0"}},
	} {
		for _, command := range []string{"mutate2", "mutate3"} {
			t.Run(tc.name+"/"+command, func(t *testing.T) {
				input := writeAccuracyInput(t, "types.csv", tc.input)
				out, _ := runAccuracyCLI(t, command, "-n", "result", "-e", tc.expression, input)
				rows := parseAccuracyCSV(t, out)
				var got []string
				for _, row := range rows[1:] {
					got = append(got, row[len(row)-1])
				}
				if !reflect.DeepEqual(got, tc.values) {
					t.Fatalf("computed values: got %q, want %q", got, tc.values)
				}
			})
		}
	}
}

// TestExpressionDateValues preserves govaluate's date parsing when binding other rows.
func TestExpressionDateValues(t *testing.T) {
	input := writeAccuracyInput(t, "dates.csv", "x\n1\n2021-01-02\n2\ntext\n")
	out, _ := runAccuracyCLI(t, "mutate2", "-n", "result", "-e", "$x", input)
	timestamp := time.Date(2021, 1, 2, 0, 0, 0, 0, time.Local).Unix()
	want := [][]string{{"x", "result"}, {"1", "1.00"}, {"2021-01-02", fmt.Sprintf("%d.00", timestamp)}, {"2", "2.00"}, {"text", "text"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("date values: got %q, want %q", got, want)
	}
}

// TestExpressionBindingValidation preserves type and literal-parser errors.
func TestExpressionBindingValidation(t *testing.T) {
	for _, tc := range []struct {
		command, expression, input, errorText string
	}{
		{"mutate3", "$x == $y", "x,y\n1,2\n3,text\n", "mismatched types float64 and string"},
		{"mutate3", "$x in [1,2]", "x,y\n1,2\ntext,3\n", "cannot use string as type int in array"},
		{"mutate2", "$x + $y", "x,y\na,b\n2021-01-02,2021-01-03\n", "Cannot transition token types"},
		{"filter2", "$x =~ $y", "x,y\nabc,a.*\nabc,[\n", "error parsing regexp"},
		{"mutate2", "-$x", "x,y\n1,2\n-2,3\n", "Invalid token: '--'"},
		{"mutate2", "- $x", "x,y\n1,2\n-2,3\n", "Cannot transition token types"},
	} {
		t.Run(tc.command+"/"+tc.expression, func(t *testing.T) {
			input := writeAccuracyInput(t, "invalid.csv", tc.input)
			args := []string{tc.command}
			if tc.command == "filter2" {
				args = append(args, "-f", tc.expression)
			} else {
				args = append(args, "-n", "result", "-e", tc.expression)
			}
			args = append(args, input)
			cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
			cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), tc.errorText) {
				t.Fatalf("expected %q, got %v: %s", tc.errorText, err, out)
			}
		})
	}
}

// TestExpressionNegativePower keeps the precedence of substituted negative literals.
func TestExpressionNegativePower(t *testing.T) {
	input := writeAccuracyInput(t, "signed.csv", "x\n2\n-2\n3\n")
	out, _ := runAccuracyCLI(t, "mutate3", "-n", "result", "-e", "$x ^ 2", input)
	want := [][]string{{"x", "result"}, {"2", "4.00"}, {"-2", "-4.00"}, {"3", "9.00"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("negative literal precedence: got %q, want %q", got, want)
	}
}

// TestExpressionIdentifierCollision keeps generated bindings distinct from user locals.
func TestExpressionIdentifierCollision(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "x\n1\n2\n")
	out, _ := runAccuracyCLI(t, "mutate3", "-n", "result", "-e", "let csvtkField0 = 10; $x + csvtkField0", input)
	want := [][]string{{"x", "result"}, {"1", "11.00"}, {"2", "12.00"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("identifier collision: got %q, want %q", got, want)
	}
}
