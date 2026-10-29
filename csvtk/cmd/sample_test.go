package cmd

import "testing"

// TestSampleInvalidProportions rejects NaN rather than silently discarding every row.
func TestSampleInvalidProportions(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a\nx\ny\n")
	for _, p := range []string{"NaN", "+Inf", "-Inf", "-0.1", "1.1"} {
		assertInputError(t, "should be in range", "sample", "-p", p, input)
	}
	out, _ := runAccuracyCLI(t, "sample", "-p", "1", input)
	if out != "a\nx\ny\n" {
		t.Fatal(out)
	}
}
