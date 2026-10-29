package cmd

import "testing"

// TestMutateRemovedColumnPosition checks insertion boundaries after removing a source column.
func TestMutateRemovedColumnPosition(t *testing.T) {
	for _, tc := range []struct{ input, source, option, anchor, want string }{
		{"a,b\nx,y\n", "a", "--after", "b", "b,new\ny,x\n"},
		{"a,b,c\nx,y,z\n", "a", "--before", "b", "new,b,c\nx,y,z\n"},
		{"a,b,c\nx,y,z\n", "a", "--after", "b", "b,new,c\ny,x,z\n"},
		{"a,b,c\nx,y,z\n", "b", "--before", "a", "new,a,c\ny,x,z\n"},
		{"a,b,c\nx,y,z\n", "b", "--after", "c", "a,c,new\nx,z,y\n"},
		{"a,b,c\nx,y,z\n", "b", "--before", "b", "a,new,c\nx,y,z\n"},
		{"a,b,c\nx,y,z\n", "b", "--after", "b", "a,new,c\nx,y,z\n"},
		{"a\nx\n", "a", "--after", "a", "new\nx\n"},
	} {
		input := writeAccuracyInput(t, "input.csv", tc.input)
		for _, deleteHeader := range []bool{false, true} {
			args := []string{"mutate", "-f", tc.source, "-R", tc.option, tc.anchor, "-n", "new"}
			want := tc.want
			if deleteHeader {
				args = append(args, "-U")
				for i, c := range want {
					if c == '\n' {
						want = want[i+1:]
						break
					}
				}
			}
			out, _ := runAccuracyCLI(t, append(args, input)...)
			if out != want {
				t.Fatalf("%v got %q; want %q", args, out, want)
			}
		}
	}
}

// TestMutateCaptureGroups checks compiled regex semantics, including empty captures.
func TestMutateCaptureGroups(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a\nfoo\nbar\n")
	for _, p := range []string{"(?:foo)", `\(foo\)`} {
		assertInputError(t, "capturing group", "mutate", "-p", p, "-n", "new", input)
	}
	for _, tc := range []struct{ pattern, want string }{
		{"()", "a,new\nfoo,\nbar,\n"},
		{"(f)", "a,new\nfoo,f\nbar,bar\n"},
		{"(?P<letter>.)", "a,new\nfoo,f\nbar,b\n"},
	} {
		out, _ := runAccuracyCLI(t, "mutate", "-p", tc.pattern, "-n", "new", input)
		if out != tc.want {
			t.Fatalf("%s got %q", tc.pattern, out)
		}
	}
}

// TestMutateRequiresSelectedField rejects an empty source selection.
func TestMutateRequiresSelectedField(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "v\nx\n")
	for _, args := range [][]string{{"-f", "-1"}, {"-H", "-f", "-1"}} {
		args = append([]string{"mutate", "-p", "(.)", "-n", "new"}, args...)
		assertInputError(t, "expected 1 selected field", append(args, input)...)
	}
}
