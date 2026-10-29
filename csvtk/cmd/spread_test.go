package cmd

import "testing"

// TestSpreadIndependentSelections validates each scalar selector separately.
func TestSpreadIndependentSelections(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b,c\nx,k1,1\nx,k2,2\n")
	for _, selectors := range [][2]string{{"-1,-3", "-1,-2"}, {"b", "3"}, {"-a,-c", "-a,-b"}} {
		out, _ := runAccuracyCLI(t, "spread", "-k", selectors[0], "-v", selectors[1], input)
		if out != "a,k1,k2\nx,1,2\n" {
			t.Fatalf("%v: %q", selectors, out)
		}
	}
	wide := writeAccuracyInput(t, "wide.csv", "a,b,c,d\nx,y,z,q\n")
	assertInputError(t, "exactly one key field and one value field", "spread", "-k", "-1", "-v", "-2", wide)
}
