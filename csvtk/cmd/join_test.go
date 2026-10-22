package cmd

import (
	"reflect"
	"testing"
)

// TestJoinHeaderOnlyTables checks that zero-row tables retain their columns
// and participate in inner, left, and outer joins at every input position.
func TestJoinHeaderOnlyTables(t *testing.T) {
	left := writeAccuracyInput(t, "left.csv", "k,a\nx,L\n")
	right := writeAccuracyInput(t, "right.csv", "k,b\nx,R\n")
	third := writeAccuracyInput(t, "third.csv", "k,c\nx,T\n")
	emptyLeft := writeAccuracyInput(t, "left.csv", "k,a\n")
	emptyRight := writeAccuracyInput(t, "right.csv", "k,b\n")
	emptyThird := writeAccuracyInput(t, "third.csv", "k,c\n")
	cases := []struct {
		name   string
		files  []string
		header []string
		left   []string
		outer  []string
	}{
		{"first", []string{emptyLeft, right}, []string{"k", "a", "b"}, nil, []string{"x", "NA", "R"}},
		{"last", []string{left, emptyRight}, []string{"k", "a", "b"}, []string{"x", "L", "NA"}, []string{"x", "L", "NA"}},
		{"middle", []string{left, emptyRight, third}, []string{"k", "a", "b", "c"}, []string{"x", "L", "NA", "T"}, []string{"x", "L", "NA", "T"}},
		{"after match", []string{left, right, emptyThird}, []string{"k", "a", "b", "c"}, []string{"x", "L", "R", "NA"}, []string{"x", "L", "R", "NA"}},
		{"all", []string{emptyLeft, emptyRight}, []string{"k", "a", "b"}, nil, nil},
	}
	for _, tc := range cases {
		for _, mode := range []string{"inner", "left", "outer"} {
			t.Run(tc.name+"/"+mode, func(t *testing.T) {
				args := []string{"join", "-f", "k", "--na", "NA"}
				want := [][]string{tc.header}
				switch mode {
				case "left":
					args = append(args, "-L")
					if tc.left != nil {
						want = append(want, tc.left)
					}
				case "outer":
					args = append(args, "-O")
					if tc.outer != nil {
						want = append(want, tc.outer)
					}
				}
				out, _ := runAccuracyCLI(t, append(args, tc.files...)...)
				if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
					t.Fatalf("got %q, want %q", got, want)
				}
			})
		}
	}
}

// TestJoinHeaderOnlyOptions checks header formatting and left-join options when
// the right input contributes columns but has no rows to match.
func TestJoinHeaderOnlyOptions(t *testing.T) {
	left := writeAccuracyInput(t, "left.csv", "k,a\nx,L\n")
	right := writeAccuracyInput(t, "right.csv", "k,b\n")
	cases := []struct {
		name string
		args []string
		want [][]string
	}{
		{"keep unmatched", []string{"-k"}, [][]string{{"k", "a", "b"}, {"x", "L", "NA"}}},
		{"default blank", []string{"-L", "--na", ""}, [][]string{{"k", "a", "b"}, {"x", "L", ""}}},
		{"prefix", []string{"-L", "-p", "-e"}, [][]string{{"k", "left-a", "right-b"}, {"x", "L", "NA"}}},
		{"suffix", []string{"-L", "-s", "L,R"}, [][]string{{"k", "a-L", "b-R"}, {"x", "L", "NA"}}},
		{"delete header", []string{"-L", "-U"}, [][]string{{"x", "L", "NA"}}},
		{"delete empty result header", []string{"-U"}, nil},
		{"fuzzy key", []string{"-L", "-F", "-f", "k*"}, [][]string{{"k", "a", "b"}, {"x", "L", "NA"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"join", "--na", "NA"}, tc.args...)
			out, _ := runAccuracyCLI(t, append(args, left, right)...)
			if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("composite null key", func(t *testing.T) {
		left := writeAccuracyInput(t, "left.csv", "a,b,l\n,x,L\n")
		right := writeAccuracyInput(t, "right.csv", "b,a,r\n")
		out, _ := runAccuracyCLI(t, "join", "-L", "-n", "-f", "a,b;b,a", "--na", "NA", left, right)
		want := [][]string{{"a", "b", "l", "r"}, {"", "x", "L", "NA"}}
		if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

// TestJoinInputsWithoutRecords preserves the existing behavior for inputs that
// contain neither a header nor data, unlike a header-only table.
func TestJoinInputsWithoutRecords(t *testing.T) {
	left := writeAccuracyInput(t, "left.csv", "k,a\nx,L\n")
	right := writeAccuracyInput(t, "right.csv", "k,b\nx,R\n")
	for _, contents := range []string{"", "\n\n"} {
		input := writeAccuracyInput(t, "empty.csv", contents)
		for _, mode := range []string{"inner", "left", "outer"} {
			t.Run(mode+"/"+contents, func(t *testing.T) {
				args := []string{"join"}
				if mode == "left" {
					args = append(args, "-L")
				} else if mode == "outer" {
					args = append(args, "-O")
				}
				out, _ := runAccuracyCLI(t, append(args, left, input, right)...)
				want := [][]string{{"k", "a", "b"}, {"x", "L", "R"}}
				if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
					t.Fatalf("got %q, want %q", got, want)
				}
			})
		}
	}
}

// TestJoinOuterReorderedKeys checks synthetic rows against the selection order
// across multiple tables, physical column orders, and a header-only first table.
func TestJoinOuterReorderedKeys(t *testing.T) {
	for _, headerOnly := range []bool{false, true} {
		for _, fields := range []string{"2,1;1,2;2,1", "k2,k1"} {
			for _, threeFiles := range []bool{false, true} {
				contents := "k1,k2,a\n"
				if !headerOnly {
					contents += "A,B,left\n"
				}
				left := writeAccuracyInput(t, "left.csv", contents)
				right := writeAccuracyInput(t, "right.csv", "k2,k1,b\nY,X,right\n")
				third := writeAccuracyInput(t, "third.csv", "k1,k2,c\nX,Y,third\n")
				files := []string{left, right}
				want := [][]string{{"k1", "k2", "a", "b"}}
				if !headerOnly {
					want = append(want, []string{"A", "B", "left", "NA"})
				}
				want = append(want, []string{"X", "Y", "NA", "right"})
				selection := fields
				if threeFiles {
					files = append(files, third)
					want[0] = append(want[0], "c")
					if !headerOnly {
						want[1] = append(want[1], "NA")
					}
					want[len(want)-1] = append(want[len(want)-1], "third")
				} else if fields != "k2,k1" {
					selection = "2,1;1,2"
				}
				args := append([]string{"join", "-O", "--na", "NA", "-f", selection}, files...)
				out, _ := runAccuracyCLI(t, args...)
				if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
					t.Fatalf("headerOnly=%t fields=%s threeFiles=%t: got %q, want %q", headerOnly, selection, threeFiles, got, want)
				}
			}
		}
	}
}
