package cmd

import "testing"

// TestRenameDeleteHeader honors the shared output-header flag without dropping data.
func TestRenameDeleteHeader(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\nx,y\n")
	out, _ := runAccuracyCLI(t, "rename", "-f", "1", "-n", "new", "-U", input)
	if out != "x,y\n" {
		t.Fatal(out)
	}
}
