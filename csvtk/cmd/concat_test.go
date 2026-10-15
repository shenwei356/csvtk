package cmd

import (
	"context"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestConcatStreamingColumns covers mapping and schema boundaries without row buffering.
func TestConcatStreamingColumns(t *testing.T) {
	empty := writeAccuracyInput(t, "empty.csv", "")
	header := writeAccuracyInput(t, "header.csv", "ID,value\n")
	first := writeAccuracyInput(t, "first.csv", "ID,value\n1,a\n")
	reordered := writeAccuracyInput(t, "reordered.csv", "VALUE,id\nb,2\n")
	partial := writeAccuracyInput(t, "partial.csv", "ID,other\n3,x\n")
	duplicate := writeAccuracyInput(t, "duplicate.csv", "id,id\n1,2\n")
	duplicate2 := writeAccuracyInput(t, "duplicate2.csv", "id,id\n3,4\n")
	for _, tc := range []struct {
		name string
		args []string
		want [][]string
	}{
		{"empty_then_header", []string{empty, header, first}, [][]string{{"ID", "value"}, {"1", "a"}}},
		{"header_only", []string{header, header}, [][]string{{"ID", "value"}}},
		{"case_and_partial", []string{"-i", "-u", "NA", first, reordered, partial}, [][]string{{"ID", "value"}, {"1", "a"}, {"2", "b"}, {"3", "NA"}}},
		{"row_numbers", []string{"-Z", "-i", "-U", first, reordered, partial}, [][]string{{"1", "1", "a"}, {"2", "2", "b"}, {"3", "3", ""}}},
		{"duplicate_columns", []string{duplicate, duplicate2}, [][]string{{"id", "id"}, {"1", "2"}, {"3", "4"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, _ := runAccuracyCLI(t, append([]string{"concat"}, tc.args...)...)
			if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("concatenated rows: got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestConcatEmitsBeforeEOF requires rows to be written while stdin remains open.
func TestConcatEmitsBeforeEOF(t *testing.T) {
	first := writeAccuracyInput(t, "first.csv", "id,value\nfirst,1\n")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "concat", first, "-")
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		cancel()
		stdin.Close()
		if !waited {
			cmd.Wait()
		}
	}()
	written := make(chan error, 1)
	go func() {
		_, err := io.WriteString(stdin, "value,id\n")
		if err == nil {
			_, err = io.WriteString(stdin, strings.Repeat("2,next\n", 20000))
		}
		written <- err
	}()
	const prefix = "id,value\nfirst,1\nnext,2\n"
	observed := make(chan error, 1)
	drained := make(chan error, 1)
	buffer := make([]byte, len(prefix))
	go func() {
		_, err := io.ReadFull(stdout, buffer)
		observed <- err
		_, err = io.Copy(io.Discard, stdout)
		drained <- err
	}()
	if err := <-observed; err != nil {
		t.Fatalf("no output before stdin EOF: %v", err)
	}
	if string(buffer) != prefix {
		t.Fatalf("streamed prefix: got %q, want %q", buffer, prefix)
	}
	if err := <-written; err != nil {
		t.Fatal(err)
	}
	stdin.Close()
	if err := <-drained; err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait()
	waited = true
	if err != nil {
		t.Fatal(err)
	}
}

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

// TestConcatRejectsAmbiguousColumns prevents generated duplicate names from losing data.
func TestConcatRejectsAmbiguousColumns(t *testing.T) {
	input := writeAccuracyInput(t, "ambiguous.csv", "id,id,id_1\n1,2,3\n")
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "concat", input, input)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "ambiguous column name after duplicate-name normalization") {
		t.Fatalf("expected an ambiguous-column error, got %v: %s", err, out)
	}
}
