package cmd

import (
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestConcatOriginalFile(t *testing.T) {
	first := writeAccuracyInput(t, "first.csv", "id,value\n1,a\n2,b\n")
	second := writeAccuracyInput(t, "second.csv", "value,id\nc,3\n")

	cases := []struct {
		name string
		args []string
		want [][]string
	}{
		{
			name: "single file",
			args: []string{"concat", "--original-file", first},
			want: [][]string{{"id", "value", "original_file"}, {"1", "a", "first.csv"}, {"2", "b", "first.csv"}},
		},
		{
			name: "multiple files with reordered columns",
			args: []string{"concat", "--original-file", first, second},
			want: [][]string{{"id", "value", "original_file"}, {"1", "a", "first.csv"}, {"2", "b", "first.csv"}, {"3", "c", "second.csv"}},
		},
		{
			name: "row number",
			args: []string{"-Z", "concat", "--original-file", first, second},
			want: [][]string{{"row", "id", "value", "original_file"}, {"1", "1", "a", "first.csv"}, {"2", "2", "b", "first.csv"}, {"3", "3", "c", "second.csv"}},
		},
		{
			name: "suppressed header",
			args: []string{"-U", "concat", "--original-file", first, second},
			want: [][]string{{"1", "a", "first.csv"}, {"2", "b", "first.csv"}, {"3", "c", "second.csv"}},
		},
		{
			name: "default output",
			args: []string{"concat", first, second},
			want: [][]string{{"id", "value"}, {"1", "a"}, {"2", "b"}, {"3", "c"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, _ := runAccuracyCLI(t, tc.args...)
			if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestConcatOriginalFileUnmatchedAndHeaderless(t *testing.T) {
	first := writeAccuracyInput(t, "first.csv", "id,value\n1,a\n")
	unmatched := writeAccuracyInput(t, "unmatched.csv", "other\nx\n")
	third := writeAccuracyInput(t, "third.csv", "id,value\n2,b\n")

	out, _ := runAccuracyCLI(t, "concat", "--original-file", first, unmatched, third)
	want := [][]string{{"id", "value", "original_file"}, {"1", "a", "first.csv"}, {"2", "b", "third.csv"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("skipped unmatched file: got %q, want %q", got, want)
	}

	out, _ = runAccuracyCLI(t, "concat", "--original-file", "-k", "-u", "NA", first, unmatched)
	want = [][]string{{"id", "value", "original_file"}, {"1", "a", "first.csv"}, {"NA", "NA", "unmatched.csv"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("kept unmatched file: got %q, want %q", got, want)
	}

	headerless1 := writeAccuracyInput(t, "headerless1.csv", "1,a\n")
	headerless2 := writeAccuracyInput(t, "headerless2.csv", "3,c\n")
	out, _ = runAccuracyCLI(t, "-H", "concat", "--original-file", headerless1, headerless2)
	want = [][]string{{"1", "a", "headerless1.csv"}, {"3", "c", "headerless2.csv"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("headerless input: got %q, want %q", got, want)
	}
}

func TestConcatOriginalFileColumnCollision(t *testing.T) {
	input := writeAccuracyInput(t, "collision.csv", "id,original_file\n1,old\n")
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "concat", "--original-file", input)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "input already has an original_file column") {
		t.Fatalf("expected a column collision error, got %q (error: %v)", output, err)
	}
}
