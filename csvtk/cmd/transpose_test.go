package cmd

import "testing"

// TestTransposeDeleteHeader removes the output header after transposing.
func TestTransposeDeleteHeader(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "id,a,b\nx,1,2\ny,3,4\n")
	for _, tc := range []struct {
		args []string
		want string
	}{
		{nil, "id,x,y\na,1,3\nb,2,4\n"},
		{[]string{"-U"}, "a,1,3\nb,2,4\n"},
		{[]string{"-H", "-U"}, "id,x,y\na,1,3\nb,2,4\n"},
	} {
		out, _ := runAccuracyCLI(t, append(append([]string{"transpose"}, tc.args...), input)...)
		if out != tc.want {
			t.Errorf("args %v: got %q, want %q", tc.args, out, tc.want)
		}
	}
}
