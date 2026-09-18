package cmd

import (
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestUnfoldParallelFields(t *testing.T) {
	input := writeAccuracyInput(t, "parallel.csv", "key,en,es,meta\nfoo,one;two,uno;due,x\nbar,a;;c,A;;C,y\n")
	out, _ := runAccuracyCLI(t, "unfold", "-f", "es,en", "-s", ";", input)
	got := parseAccuracyCSV(t, out)
	want := [][]string{
		{"key", "en", "es", "meta"},
		{"foo", "one", "uno", "x"},
		{"foo", "two", "due", "x"},
		{"bar", "a", "A", "y"},
		{"bar", "", "", "y"},
		{"bar", "c", "C", "y"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unfolded rows: got %q, want %q", got, want)
	}

	out, _ = runAccuracyCLI(t, "unfold", "-f", "en", "-s", ";", input)
	got = parseAccuracyCSV(t, out)
	if len(got) != len(want) || got[1][2] != "uno;due" || got[4][2] != "A;;C" {
		t.Fatalf("single field unfold changed: %q", got)
	}

	noHeader := writeAccuracyInput(t, "no-header.csv", "foo,one;two,uno;due\n")
	out, _ = runAccuracyCLI(t, "-H", "unfold", "-f", "2,3", "-s", ";", noHeader)
	got = parseAccuracyCSV(t, out)
	want = [][]string{{"foo", "one", "uno"}, {"foo", "two", "due"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("headerless unfolded rows: got %q, want %q", got, want)
	}
}

func TestUnfoldParallelFieldsLengthMismatch(t *testing.T) {
	input := writeAccuracyInput(t, "mismatch.csv", "key,en,es\nfoo,one;two,uno\n")
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "unfold", "-f", "en,es", "-s", ";", input)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected an error for unequal parallel arrays, got %q", output)
	}
	if !strings.Contains(string(output), "row 1: selected fields have different numbers of values (field 2: 2, field 3: 1)") {
		t.Fatalf("unexpected error: %q", output)
	}
}
