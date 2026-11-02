package cmd

import "testing"

// TestSortNumericColumnNames distinguishes literal names from integer indices.
func TestSortNumericColumnNames(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "id,1.0,1e2\na,z,z\nb,a,a\n")
	for _, key := range []string{"1.0", "1e2", "1.0:r"} {
		out, _ := runAccuracyCLI(t, "sort", "-k", key, input)
		want := "id,1.0,1e2\nb,a,a\na,z,z\n"
		if key == "1.0:r" {
			want = "id,1.0,1e2\na,z,z\nb,a,a\n"
		}
		if out != want {
			t.Errorf("key %s: got %q, want %q", key, out, want)
		}
	}
}

// TestSortExcludedOpenRange uses the remaining columns for every sort mode.
func TestSortExcludedOpenRange(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b,c\n2,x,a\n10,a,b\n")
	for _, tc := range []struct{ key, want string }{
		{"-2-", "a,b,c\n10,a,b\n2,x,a\n"},
		{"-2-:n", "a,b,c\n2,x,a\n10,a,b\n"},
		{"-2-:nr", "a,b,c\n10,a,b\n2,x,a\n"},
	} {
		out, _ := runAccuracyCLI(t, "sort", "-k", tc.key, input)
		if out != tc.want {
			t.Errorf("key %s: got %q, want %q", tc.key, out, tc.want)
		}
	}
	input = writeAccuracyInput(t, "input.csv", "a,b,c\nx,y,z\nx,y,a\n")
	out, _ := runAccuracyCLI(t, "sort", "-k", "-4-", input)
	if want := "a,b,c\nx,y,a\nx,y,z\n"; out != want {
		t.Errorf("exclusion past the last column: got %q, want %q", out, want)
	}
}

// TestSortSingleColumnOpenRanges includes the implicit default key and an open
// range starting at the last column, without changing explicit range validation.
func TestSortSingleColumnOpenRanges(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
		args              []string
	}{
		{"default", "v\nb\na\n", "v\na\nb\n", nil},
		{"no header", "b\na\n", "a\nb\n", []string{"-H"}},
		{"last column", "id,v\nx,b\ny,a\n", "id,v\ny,a\nx,b\n", []string{"-k", "2-"}},
		{"numeric reverse", "v\n2\n10\n", "v\n10\n2\n", []string{"-k", "1-:nr"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := writeAccuracyInput(t, "input.csv", tc.input)
			args := append([]string{"sort"}, tc.args...)
			out, _ := runAccuracyCLI(t, append(args, input)...)
			if out != tc.want {
				t.Fatalf("got %q, want %q", out, tc.want)
			}
		})
	}
	input := writeAccuracyInput(t, "input.csv", "v\nb\na\n")
	assertInputError(t, "invalid field range", "sort", "-k", "1-1", input)
}
