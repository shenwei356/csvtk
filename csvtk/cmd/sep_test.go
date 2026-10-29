package cmd

import "testing"

// TestSepDeleteHeader resolves named fields before suppressing the header.
func TestSepDeleteHeader(t *testing.T) {
	for _, data := range []string{"a,b\n1;2,z\n3,w\n", "a,b\n"} {
		input := writeAccuracyInput(t, "input.csv", data)
		out, _ := runAccuracyCLI(t, "sep", "-U", "-R", "-f", "a", "-s", ";", "-n", "x,y", input)
		want := "z,1,2\nw,3,\n"
		if data == "a,b\n" {
			want = ""
		}
		if out != want {
			t.Fatalf("got %q, want %q", out, want)
		}
	}
	input := writeAccuracyInput(t, "no-header.csv", "1;2,z\n")
	out, _ := runAccuracyCLI(t, "sep", "-H", "-U", "-R", "-s", ";", "-N", "2", input)
	if out != "z,1,2\n" {
		t.Fatalf("headerless: %q", out)
	}
}
