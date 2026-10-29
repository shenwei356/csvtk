package cmd

import "testing"

// TestRoundNumericForms preserves exponents and leaves non-numeric fields intact.
func TestRoundNumericForms(t *testing.T) {
	input := writeAccuracyInput(t, "numbers.csv", "x\n1.234\n12\n-2.5\n.1234\n1.234e5\n-1.234E-05\n+2.\n.\ntext\n")
	out, _ := runAccuracyCLI(t, "round", "-n", "2", input)
	want := "x\n1.23\n12.00\n-2.50\n0.12\n1.23e5\n-1.23E-05\n2.00\n.\ntext\n"
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}
