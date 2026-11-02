package cmd

import "testing"

// TestAddHeaderInputWidths validates each file against the shared output header.
func TestAddHeaderInputWidths(t *testing.T) {
	first := writeAccuracyInput(t, "first.csv", "a,b\n")
	second := writeAccuracyInput(t, "second.csv", "c\n")
	for _, args := range [][]string{nil, {"-n", "x,y"}} {
		assertInputError(t, "number of fields", append(append([]string{"add-header"}, args...), first, second)...)
	}
	third := writeAccuracyInput(t, "third.csv", "c,d\n")
	out, _ := runAccuracyCLI(t, "add-header", "-n", "x,y", first, third)
	if want := "x,y\na,b\nc,d\n"; out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}
