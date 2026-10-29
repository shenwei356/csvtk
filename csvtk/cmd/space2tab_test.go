package cmd

import (
	"os"
	"testing"
)

// TestSpace2TabDeleteHeader skips the first non-comment record of every input.
func TestSpace2TabDeleteHeader(t *testing.T) {
	input := writeAccuracyInput(t, "input.txt", "# comment\n\na b\nx y\n")
	other := writeAccuracyInput(t, "other.txt", "a b\nz w\n")
	out, _ := runAccuracyCLI(t, "space2tab", "-U", input, other)
	if out != "x\ty\nz\tw\n" {
		t.Fatalf("got %q", out)
	}
	out, _ = runAccuracyCLI(t, "space2tab", "-H", "-U", input)
	if out != "a\tb\nx\ty\n" {
		t.Fatalf("got %q", out)
	}
}

// TestSpace2TabBufferSize rejects invalid buffers before creating output.
func TestSpace2TabBufferSize(t *testing.T) {
	input := writeAccuracyInput(t, "input.txt", "a b\nx y\n")
	for _, size := range []string{"0", "-1", "NaN", "Inf", "9223372036854775808G"} {
		output := writeAccuracyInput(t, "output.tsv", "sentinel")
		assertInputError(t, "buffer size", "space2tab", "-b", size, "-o", output, input)
		got, err := os.ReadFile(output)
		if err != nil || string(got) != "sentinel" {
			t.Fatalf("output changed: %q, %v", got, err)
		}
	}
}
