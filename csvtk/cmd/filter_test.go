package cmd

import "testing"

// TestFilterAnyMissingValues keeps matches regardless of the missing column's position.
func TestFilterAnyMissingValues(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\nNA,1\n1,NA\n.,2\n-1,0\n2,3\n")
	for _, fields := range []string{"a,b", "b,a"} {
		out, _ := runAccuracyCLI(t, "filter", "--any", "-f", fields+">0", input)
		if out != "a,b\nNA,1\n1,NA\n.,2\n2,3\n" {
			t.Fatalf("%s: %q", fields, out)
		}
	}
	out, _ := runAccuracyCLI(t, "filter", "-f", "a,b>0", input)
	if out != "a,b\n2,3\n" {
		t.Fatalf("all fields: %q", out)
	}
}
