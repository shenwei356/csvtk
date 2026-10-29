package cmd

import "testing"

// TestMatrix2LongNaNFiltering checks filtering, explicit retention, and invalid bounds.
func TestMatrix2LongNaNFiltering(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", ",x,y\nx,NaN,1\ny,text,-1\n")
	for _, tc := range []struct {
		args []string
		want string
	}{
		{nil, "col1,col2,value\nx,x,NaN\nx,y,1\ny,x,text\ny,y,-1\n"},
		{[]string{"-m", "0"}, "col1,col2,value\nx,y,1\n"},
		{[]string{"-m", "0", "-N"}, "col1,col2,value\nx,x,NaN\nx,y,1\ny,x,text\n"},
	} {
		out, _ := runAccuracyCLI(t, append(append([]string{"matrix2long"}, tc.args...), input)...)
		if out != tc.want {
			t.Fatalf("%v: got %q; want %q", tc.args, out, tc.want)
		}
	}
	for _, args := range [][]string{{"-m", "NaN"}, {"-M", "NaN"}, {"-m", "2", "-M", "1"}} {
		assertInputError(t, "value bounds", append(append([]string{"matrix2long"}, args...), input)...)
	}
}
