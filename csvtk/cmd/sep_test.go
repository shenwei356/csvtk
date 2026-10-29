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

// TestSepSingleField prevents removal of columns that are never split.
func TestSepSingleField(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b,c\n1;2,3;4,z\n")
	for _, fields := range []string{"1-2", "a,b", "-1-"} {
		assertInputError(t, "expected 1 selected field", "sep", "-R", "-f", fields, "-s", ";", "-n", "x,y", input)
	}
}

// TestSepNamedFieldOverridesHeaderless keeps the declared output width and names.
func TestSepNamedFieldOverridesHeaderless(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\n1;2,z\n")
	out, _ := runAccuracyCLI(t, "sep", "-H", "-f", "a", "-s", ";", "-n", "x,y,q", input)
	if out != "a,b,x,y,q\n1;2,z,1,2,\n" {
		t.Fatalf("got %q", out)
	}
	assertInputError(t, "--name", "sep", "-H", "-f", "a", "-s", ";", input)
}

// TestSepIgnoreCase applies case folding without interpreting a literal separator as regex.
func TestSepIgnoreCase(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "s\nxA.By\nxA?By\n")
	out, _ := runAccuracyCLI(t, "sep", "-R", "-s", "a.b", "-i", "-n", "x,y", input)
	if out != "x,y\nx,y\nxA?By,\n" {
		t.Fatalf("got %q", out)
	}
}
