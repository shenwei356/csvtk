package cmd

import "testing"

// TestRename2LiteralKV preserves dollars when mapped values replace header names.
func TestRename2LiteralKV(t *testing.T) { testLiteralKV(t, "rename2") }

// TestRename2DeleteHeader suppresses the renamed header while preserving data.
func TestRename2DeleteHeader(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\nx,y\n")
	out, _ := runAccuracyCLI(t, "rename2", "-f", "1", "-p", "a", "-r", "new", "-U", input)
	if out != "x,y\n" {
		t.Fatal(out)
	}
}
