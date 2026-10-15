package cmd

import (
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// TestSummaryDefaultCountRanges checks range expansion before grouping and naming.
func TestSummaryDefaultCountRanges(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		header []string
		rows   [][]string
	}{
		{"range", []string{"-f", "1-2"}, []string{"a:count", "b:count"}, [][]string{{"4", "4"}}},
		{"grouped", []string{"-g", "3", "-f", "1-2"}, []string{"g", "a:count", "b:count"}, [][]string{{"A", "2", "2"}, {"B", "2", "2"}}},
		{"renamed", []string{"-g", "3", "-f", "1-2", "-n", "count_a,count_b"}, []string{"g", "count_a", "count_b"}, [][]string{{"A", "2", "2"}, {"B", "2", "2"}}},
		{"mixed_operations", []string{"-f", "1-2,2:sum"}, []string{"a:count", "b:count", "b:sum"}, [][]string{{"4", "4", "100.00"}}},
		{"overlapping_ranges", []string{"-f", "1-2", "-f", "2-2"}, []string{"a:count", "b:count", "b:count"}, [][]string{{"4", "4", "4"}}},
		{"single_fields", []string{"-f", "1,2"}, []string{"a:count", "b:count"}, [][]string{{"4", "4"}}},
	}
	for _, noHeader := range []bool{false, true} {
		mode := "header"
		contents := "x,10,A\nx,20,A\ny,30,B\ny,40,B\n"
		if noHeader {
			mode = "no_header"
		} else {
			contents = "a,b,g\n" + contents
		}
		input := writeAccuracyInput(t, "summary.csv", contents)
		for _, tc := range cases {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				args := append([]string{"summary"}, tc.args...)
				want := tc.rows
				if noHeader {
					args = append(args, "-H")
				} else {
					want = append([][]string{tc.header}, want...)
				}
				out, stderr := runAccuracyCLI(t, append(args, input)...)
				if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
					t.Fatalf("summary rows: got %q, want %q", got, want)
				}
				if stderr != "" {
					t.Fatalf("unexpected stderr: %s", stderr)
				}
			})
		}
	}
}

// TestSummaryCountRangeValidation requires the same validation for both spellings.
func TestSummaryCountRangeValidation(t *testing.T) {
	input := writeAccuracyInput(t, "summary.csv", "a,b\nx,1\ny,2\n")
	for _, tc := range []struct {
		field string
		names []string
		error string
	}{
		{"2-1", nil, "start should be <= end"},
		{"1-", nil, "end field is needed"},
		{"1-2", []string{"-n", "only_one"}, "number of names (1) should be equal to number of summary columns (2)"},
	} {
		for _, suffix := range []string{"", ":count"} {
			t.Run(tc.field+suffix, func(t *testing.T) {
				args := []string{"-test.run=^TestCLIAccuracyHelper$", "--", "summary", "-f", tc.field + suffix}
				args = append(args, tc.names...)
				cmd := exec.Command(os.Args[0], append(args, input)...)
				cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
				out, err := cmd.CombinedOutput()
				if err == nil || !strings.Contains(string(out), tc.error) {
					t.Fatalf("expected %q, got %v: %s", tc.error, err, out)
				}
			})
		}
	}
}
