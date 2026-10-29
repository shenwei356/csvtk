package cmd

import (
	"reflect"
	"strings"
	"testing"
)

// TestCutMissingExcludedColumns keeps the missing-column flags aligned after exclusion.
func TestCutMissingExcludedColumns(t *testing.T) {
	input := writeAccuracyInput(t, "columns.csv", "a,b,c,d\n1,2,3,4\n")
	for _, noHeader := range []bool{false, true} {
		for _, missing := range []string{"-m", "-b"} {
			for _, fields := range []string{"-1", "-1,-9", "-1,-3-"} {
				args := []string{"cut", missing, "-f", fields, input}
				if noHeader {
					args = append(args, "-H")
				}
				out, _ := runAccuracyCLI(t, args...)
				want := "b,c,d\n2,3,4\n"
				if fields == "-1,-3-" {
					want = "b\n2\n"
				}
				if out != want {
					t.Fatalf("%v: got %q, want %q", args, out, want)
				}
			}
		}
	}
}

// TestCutOpenRanges checks that each range expands at its requested position.
func TestCutOpenRanges(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		columns []int // Zero-based output columns, including requested repeats.
	}{
		{"overlapping", []string{"-f", "1-,3-"}, []int{0, 1, 2, 3, 2, 3}},
		{"reordered", []string{"-f", "3-,1-"}, []int{2, 3, 0, 1, 2, 3}},
		{"closed_between_open", []string{"-f", "2-,1-2,4-"}, []int{1, 2, 3, 0, 1, 3}},
		{"closed_before_open", []string{"-f", "1-2,3-,2-"}, []int{0, 1, 2, 3, 1, 2, 3}},
		{"single_between_open", []string{"-f", "2-,1,4-"}, []int{1, 2, 3, 0, 3}},
		{"identical", []string{"-f", "1-,1-"}, []int{0, 1, 2, 3, 0, 1, 2, 3}},
		{"repeated_flags", []string{"-f", "1-", "-f", "3-"}, []int{0, 1, 2, 3, 2, 3}},
		{"excluded", []string{"-f", "-2-,-4-"}, []int{0}},
		{"excluded_reordered", []string{"-f", "-4-,-2-"}, []int{0}},
	}
	for _, noHeader := range []bool{false, true} {
		mode := "header"
		rows := [][]string{{"a", "b", "c", "d"}, {"1", "2", "3", "4"}, {"5", "6", "7", "8"}}
		if noHeader {
			mode = "no_header"
			rows = rows[1:]
		}
		var contents strings.Builder
		for _, row := range rows {
			contents.WriteString(strings.Join(row, ","))
			contents.WriteByte('\n')
		}
		input := writeAccuracyInput(t, "ranges.csv", contents.String())
		for _, tc := range cases {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				args := append([]string{"cut"}, tc.args...)
				if noHeader {
					args = append(args, "-H")
				}
				out, stderr := runAccuracyCLI(t, append(args, input)...)
				want := make([][]string, len(rows))
				for i, row := range rows {
					for _, column := range tc.columns {
						want[i] = append(want[i], row[column])
					}
				}
				if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
					t.Fatalf("selected columns: got %q, want %q", got, want)
				}
				if stderr != "" {
					t.Fatalf("unexpected stderr: %s", stderr)
				}
			})
		}
	}
}
