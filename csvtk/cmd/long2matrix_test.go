package cmd

import "testing"

// TestLong2MatrixNaNFiltering treats NaN like other non-numeric data when filtering.
func TestLong2MatrixNaNFiltering(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b,v\nx,x,NaN\nx,y,1\ny,x,text\ny,y,-1\n")
	for _, tc := range []struct {
		args []string
		want string
	}{
		{nil, ",x,y\nx,NaN,1\ny,text,-1\n"},
		{[]string{"-m", "0"}, ",x,y\nx,,1\ny,,\n"},
		{[]string{"-m", "0", "-N"}, ",x,y\nx,NaN,1\ny,text,\n"},
		{[]string{"-m", "0", "-B", "--na", "NA"}, ",x,y\nx,NA,1\ny,NA,NA\n"},
	} {
		out, _ := runAccuracyCLI(t, append(append([]string{"long2matrix"}, tc.args...), input)...)
		if out != tc.want {
			t.Fatalf("%v: got %q; want %q", tc.args, out, tc.want)
		}
	}
	for _, args := range [][]string{{"-m", "NaN"}, {"-M", "NaN"}, {"-m", "2", "-M", "1"}} {
		assertInputError(t, "value bounds", append(append([]string{"long2matrix"}, args...), input)...)
	}
}
