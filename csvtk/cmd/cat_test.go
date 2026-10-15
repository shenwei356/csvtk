package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestCatLinesPreservesBytes checks file and stdin streams, including partial lines.
func TestCatLinesPreservesBytes(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"blank_line", "\n"},
		{"terminated", "first\nlast\n"},
		{"unterminated", "first\nlast"},
		{"single_line", "last"},
		{"crlf", "first\r\nlast\r"},
		{"long_line", "first\n" + strings.Repeat("last", 4096)},
	} {
		for _, source := range []string{"file", "stdin"} {
			t.Run(tc.name+"/"+source, func(t *testing.T) {
				lines := strings.Count(tc.text, "\n")
				if tc.text != "" && !strings.HasSuffix(tc.text, "\n") {
					lines++
				}
				input := "-"
				if source == "file" {
					input = writeAccuracyInput(t, "input.txt", tc.text)
				}
				args := []string{"-test.run=^TestCLIAccuracyHelper$", "--", "cat", "-L", "-s", strconv.Itoa(lines), "-b", "16", input}
				cmd := exec.Command(os.Args[0], args...)
				cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
				if source == "stdin" {
					cmd.Stdin = strings.NewReader(tc.text)
				}
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				if err := cmd.Run(); err != nil {
					t.Fatalf("cat -L: %v: %s", err, stderr.String())
				}
				if got := stdout.String(); got != tc.text {
					t.Fatalf("output differs: got %d bytes, want %d bytes", len(got), len(tc.text))
				}
			})
		}
	}
}

// TestCatLinesMultipleFiles preserves partial lines at each file boundary.
func TestCatLinesMultipleFiles(t *testing.T) {
	first := writeAccuracyInput(t, "first.txt", "first\nlast")
	empty := writeAccuracyInput(t, "empty.txt", "")
	last := writeAccuracyInput(t, "last.txt", "next\nfinal")
	output := filepath.Join(t.TempDir(), "output.txt")
	stdout, _ := runAccuracyCLI(t, "cat", "-L", "-s", "2", first, empty, last, "-o", output)
	if stdout != "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "first\nlastnext\nfinal" {
		t.Fatalf("concatenated bytes changed: %q", got)
	}
}
