package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestPaste(t *testing.T) {
	left := writeAccuracyInput(t, "left.csv", "field1,field2\n0,a\n1,b\n3,a\n")
	right := writeAccuracyInput(t, "right.csv", "field3,field4\n0,a\n1,b\n3,a\n")

	cases := []struct {
		name string
		args []string
		want [][]string
	}{
		{
			name: "columns",
			args: []string{"paste", left, right},
			want: [][]string{{"field1", "field2", "field3", "field4"}, {"0", "a", "0", "a"}, {"1", "b", "1", "b"}, {"3", "a", "3", "a"}},
		},
		{
			name: "row numbers",
			args: []string{"-Z", "paste", left, right},
			want: [][]string{{"row", "field1", "field2", "field3", "field4"}, {"1", "0", "a", "0", "a"}, {"2", "1", "b", "1", "b"}, {"3", "3", "a", "3", "a"}},
		},
		{
			name: "suppress header",
			args: []string{"-U", "paste", left, right},
			want: [][]string{{"0", "a", "0", "a"}, {"1", "b", "1", "b"}, {"3", "a", "3", "a"}},
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

func TestPasteHeaderlessAndCSVRecords(t *testing.T) {
	left := writeAccuracyInput(t, "left.csv", "1,\"a\nb\"\n2,c\n")
	right := writeAccuracyInput(t, "right.csv", "x\ny\n")
	out, _ := runAccuracyCLI(t, "-H", "paste", left, right)
	want := [][]string{{"1", "a\nb", "x"}, {"2", "c", "y"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPasteTSV(t *testing.T) {
	left := writeAccuracyInput(t, "left.tsv", "a\tb\n1\t2\n")
	right := writeAccuracyInput(t, "right.tsv", "c\n3\n")
	out, _ := runAccuracyCLI(t, "-t", "paste", left, right)
	if want := "a\tb\tc\n1\t2\t3\n"; out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestPasteStdin(t *testing.T) {
	right := writeAccuracyInput(t, "right.csv", "b\n2\n")
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "paste", "-", right)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	cmd.Stdin = strings.NewReader("a\n1\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("paste failed: %v\nstderr: %s", err, stderr.String())
	}
	want := [][]string{{"a", "b"}, {"1", "2"}}
	if got := parseAccuracyCSV(t, stdout.String()); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPastePadsUnequalRows(t *testing.T) {
	long := writeAccuracyInput(t, "long.csv", "a,c\n1,x\n2,y\n")
	short := writeAccuracyInput(t, "short.csv", "b,d\n3,z\n")

	out, _ := runAccuracyCLI(t, "paste", long, short)
	want := [][]string{{"a", "c", "b", "d"}, {"1", "x", "3", "z"}, {"2", "y", "", ""}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("shorter right input: got %q, want %q", got, want)
	}

	out, _ = runAccuracyCLI(t, "paste", short, long)
	want = [][]string{{"b", "d", "a", "c"}, {"3", "z", "1", "x"}, {"", "", "2", "y"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("shorter left input: got %q, want %q", got, want)
	}
}

func TestPastePadsEmptyInput(t *testing.T) {
	empty := writeAccuracyInput(t, "empty.csv", "")
	right := writeAccuracyInput(t, "right.csv", "b,d\n3,z\n")
	out, _ := runAccuracyCLI(t, "paste", empty, right)
	want := [][]string{{"", "b", "d"}, {"", "3", "z"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPasteRejectsMultipleStdin(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "paste", "-", "-")
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "stdin can only be used once") {
		t.Fatalf("expected a duplicate-stdin error, got %q (error: %v)", output, err)
	}
}
