package cmd

import "testing"

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
