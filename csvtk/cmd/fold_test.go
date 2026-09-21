package cmd

import (
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestFoldParallelFields(t *testing.T) {
	input := writeAccuracyInput(t, "parallel.csv", "id,en,es,meta\n1,one,uno,x\n1,two,due,y\n2,a,,z\n2,,B,w\n")
	out, _ := runAccuracyCLI(t, "fold", "-f", "id", "-v", "es,en", "-s", ";", input)
	got := parseAccuracyCSV(t, out)
	want := [][]string{
		{"id", "es", "en"},
		{"1", "uno;due", "one;two"},
		{"2", ";B", "a;"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("folded rows: got %q, want %q", got, want)
	}

	folded := writeAccuracyInput(t, "folded.csv", out)
	out, _ = runAccuracyCLI(t, "unfold", "-f", "es,en", "-s", ";", folded)
	got = parseAccuracyCSV(t, out)
	want = [][]string{
		{"id", "es", "en"},
		{"1", "uno", "one"},
		{"1", "due", "two"},
		{"2", "", "a"},
		{"2", "B", ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fold/unfold round trip: got %q, want %q", got, want)
	}

	out, _ = runAccuracyCLI(t, "fold", "-f", "id", "-v", "en", "-s", ";", input)
	got = parseAccuracyCSV(t, out)
	want = [][]string{{"id", "en"}, {"1", "one;two"}, {"2", "a;"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("single field fold changed: got %q, want %q", got, want)
	}
}

func TestFoldParallelFieldsWithRanges(t *testing.T) {
	issueInput := writeAccuracyInput(t, "issue-320.csv", "1,a,A\n1,b,B\n2,x,X\n2,y,Y\n")
	out, _ := runAccuracyCLI(t, "-H", "fold", "-f", "1", "-v", "2,3", "-s", ";", issueInput)
	got := parseAccuracyCSV(t, out)
	want := [][]string{{"1", "a;b", "A;B"}, {"2", "x;y", "X;Y"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("issue #320 fields: got %q, want %q", got, want)
	}

	input := writeAccuracyInput(t, "no-header.csv", "x:y,z,one,uno\nx:y,z,two,due\na,b,three,tres\n")
	out, _ = runAccuracyCLI(t, "-H", "fold", "-f", "1-2", "-v", "3-", "-s", ";", input)
	got = parseAccuracyCSV(t, out)
	want = [][]string{{"x:y", "z", "one;two", "uno;due"}, {"a", "b", "three", "tres"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ranged fields: got %q, want %q", got, want)
	}

	fuzzyInput := writeAccuracyInput(t, "fuzzy.csv", "key1,key2,en,es\na,b,one,uno\na,b,two,due\n")
	out, _ = runAccuracyCLI(t, "fold", "-F", "-f", "key*", "-v", "e*", "-s", ";", fuzzyInput)
	got = parseAccuracyCSV(t, out)
	want = [][]string{{"key1", "key2", "en", "es"}, {"a", "b", "one;two", "uno;due"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fuzzy key fields: got %q, want %q", got, want)
	}
}

func TestFoldRejectsOverlappingFields(t *testing.T) {
	input := writeAccuracyInput(t, "overlap.csv", "key,value,other\na,b,c\n")
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "fold", "-f", "key,value", "-v", "value,other", input)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "field 2 selected more than once") {
		t.Fatalf("expected overlapping fields error, got %q, err: %v", output, err)
	}
}
