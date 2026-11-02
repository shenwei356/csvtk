package cmd

import (
	"math"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// TestSummaryLargeQuantiles keeps finite order statistics finite.
func TestSummaryLargeQuantiles(t *testing.T) {
	for _, tc := range []struct{ values, want string }{
		{"1e308\n1e308\n", "1.00e+308,1.00e+308,1.00e+308,1.00e+308"},
		{"-1e308\n1e308\n", "0.00e+00,-5.00e+307,0.00e+00,5.00e+307"},
	} {
		input := writeAccuracyInput(t, "input.csv", "x\n"+tc.values)
		out, _ := runAccuracyCLI(t, "summary", "-f", "x:median,x:q1,x:q2,x:q3", input)
		want := "x:median,x:q1,x:q2,x:q3\n" + tc.want + "\n"
		if out != want {
			t.Errorf("got %q, want %q", out, want)
		}
	}
}

// TestPercentileBoundaries checks exact endpoints and tiny nonzero values.
func TestPercentileBoundaries(t *testing.T) {
	for _, tc := range []struct {
		values  []float64
		p, want float64
	}{
		{[]float64{-1e308, 1e308}, 0, -1e308},
		{[]float64{-1e308, 1e308}, 1, 1e308},
		{[]float64{math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64}, .5, math.SmallestNonzeroFloat64},
		{[]float64{math.SmallestNonzeroFloat64, 2 * math.SmallestNonzeroFloat64}, .5, 2 * math.SmallestNonzeroFloat64},
		{[]float64{1, 2, 3, 4}, .25, 1.75},
	} {
		if got := percentileValue(tc.values, tc.p); got != tc.want {
			t.Errorf("percentile(%v, %g) = %g, want %g", tc.values, tc.p, got, tc.want)
		}
	}
}

// TestSummaryOverflowingRange rejects integer conversion errors before expansion.
func TestSummaryOverflowingRange(t *testing.T) {
	input := writeAccuracyInput(t, "summary.csv", "a,b\n1,2\n")
	for _, spec := range []string{"99999999999999999999-99999999999999999999", "1-99999999999999999999"} {
		assertInputError(t, "value out of range", "summary", "-f", spec, input)
	}
}

// TestSummaryOperationStorage covers groups requiring numeric, text, or both value sets.
func TestSummaryOperationStorage(t *testing.T) {
	input := writeAccuracyInput(t, "summary.csv", "g,v,t\nA,3,x\nA,1,y\nB,NA,z\n")
	for _, tc := range []struct {
		fields string
		want   [][]string
	}{
		{"v:sum,v:countn", [][]string{{"g", "v:sum", "v:countn"}, {"A", "4.00", "2"}, {"B", "NaN", "0"}}},
		{"v:count,t:collapse", [][]string{{"g", "v:count", "t:collapse"}, {"A", "2", "x; y"}, {"B", "1", "z"}}},
		{"v:sum,v:count,v:first,t:collapse", [][]string{{"g", "v:sum", "v:count", "v:first", "t:collapse"}, {"A", "4.00", "2", "3", "x; y"}, {"B", "NaN", "1", "NA", "z"}}},
	} {
		t.Run(tc.fields, func(t *testing.T) {
			out, _ := runAccuracyCLI(t, "summary", "-i", "-g", "g", "-f", tc.fields, input)
			if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("summary rows: got %q, want %q", got, tc.want)
			}
		})
	}
}

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

// TestSummaryIndependentFields keeps excluded data columns from becoming group keys.
func TestSummaryIndependentFields(t *testing.T) {
	input := writeAccuracyInput(t, "summary.csv", "g,a,b\nX,1,2\nX,3,4\nY,5,6\n")
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"-f", "-1:count"}, "a:count,b:count\n3,3\n"},
		{[]string{"-f", "-g:count", "-g", "g"}, "g,a:count,b:count\nX,2,2\nY,1,1\n"},
		{[]string{"-f", "2:sum", "-g", "-2,-3"}, "g,a:sum\nX,4.00\nY,5.00\n"},
		{[]string{"-f", "a:sum", "-g", "1"}, "g,a:sum\nX,4.00\nY,5.00\n"},
		{[]string{"-f", "-1:count,2:sum", "-n", "a_count,b_count,total"}, "a_count,b_count,total\n3,3,9.00\n"},
	} {
		out, _ := runAccuracyCLI(t, append(append([]string{"summary"}, tc.args...), input)...)
		if out != tc.want {
			t.Fatalf("args %q: got %q; want %q", tc.args, out, tc.want)
		}
	}
	assertInputError(t, "duplicated field", "summary", "-g", "1-2", "-f", "a:sum", input)
}
