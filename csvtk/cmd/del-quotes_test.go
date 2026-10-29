package cmd

import "testing"

// TestDelQuotesDeleteHeader suppresses only a real header, preserving headerless input.
func TestDelQuotesDeleteHeader(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\nx,y\n")
	out, _ := runAccuracyCLI(t, "del-quotes", "-U", input)
	if out != "x,y\n" {
		t.Fatal(out)
	}
	out, _ = runAccuracyCLI(t, "del-quotes", "-H", "-U", input)
	if out != "a,b\nx,y\n" {
		t.Fatal(out)
	}
}
