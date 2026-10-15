package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/shenwei356/xopen"
)

// TestOutputWriteErrors checks failures in the buffers below the CSV writer,
// including errors first reported when a compressed output is finalized.
func TestOutputWriteErrors(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("/dev/full is a Linux device")
	}
	input := writeAccuracyInput(t, "input.csv", "a,b\n1,2\n")
	cases := []struct {
		name string
		args []string
	}{
		{"csv", []string{"cut", "-f", "1", input}},
		{"quoted csv", []string{"cut", "--quote-all", "-f", "1", input}},
		{"grep", []string{"grep", "-p", "1", input}},
		{"json", []string{"csv2json", input}},
		{"markdown", []string{"csv2md", input}},
		{"rst", []string{"csv2rst", input}},
		{"html", []string{"csv2html", input}},
		{"pretty", []string{"pretty", input}},
		{"row count", []string{"nrow", input}},
		{"column count", []string{"ncol", input}},
		{"dimensions", []string{"dim", input}},
		{"headers", []string{"headers", input}},
		{"space conversion", []string{"space2tab", "-b", "1M", input}},
		{"quote repair", []string{"fix-quotes", "-b", "1M", input}},
		{"quote removal", []string{"del-quotes", input}},
		{"cat", []string{"cat", input}},
		{"cat lines", []string{"cat", "-L", "-s", "2", input}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertOutputWriteError(t, append(tc.args, "-o", "/dev/full"), false, "")
		})
	}
	t.Run("stdout", func(t *testing.T) {
		assertOutputWriteError(t, []string{"cut", "-f", "1", input}, true, "")
	})
	for _, suffix := range []string{".gz", ".xz", ".zst", ".bz2", ".lz4"} {
		t.Run(suffix, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "output.csv"+suffix)
			if err := os.Symlink("/dev/full", output); err != nil {
				t.Fatal(err)
			}
			assertOutputWriteError(t, []string{"cut", "-f", "1", input, "-o", output}, false, "")
		})
	}
	t.Run("split groups", func(t *testing.T) {
		dir := t.TempDir()
		// Use the default output directory so split retains the failing target.
		if err := os.Symlink("/dev/full", filepath.Join(dir, "broken-1.csv")); err != nil {
			t.Fatal(err)
		}
		assertOutputWriteError(t, []string{"split", "-f", "a", "-p", "broken-", input}, false, dir)
	})
}

// assertOutputWriteError runs a command in a subprocess because checkError exits
// the process, and requires a write error rather than an argument or input error.
func assertOutputWriteError(t *testing.T, args []string, fullStdout bool, dir string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if fullStdout {
		out, err := os.OpenFile("/dev/full", os.O_WRONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer out.Close()
		cmd.Stdout = out
	}
	if err := cmd.Run(); err == nil {
		t.Fatalf("csvtk %v succeeded despite a failed output write", args)
	} else if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("could not run csvtk: %v", err)
	}
	if !strings.Contains(stderr.String(), "no space left on device") {
		t.Fatalf("csvtk %v did not report the write error: %s", args, stderr.String())
	}
}

// TestOutputCompressionRoundTrip checks that finalizing successful outputs still
// writes complete files in every supported compression format.
func TestOutputCompressionRoundTrip(t *testing.T) {
	const inputText = "a,b\n1,2\n"
	input := writeAccuracyInput(t, "input.csv", inputText)
	for _, suffix := range []string{"", ".gz", ".xz", ".zst", ".bz2", ".lz4"} {
		t.Run("csv"+suffix, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "output.csv"+suffix)
			runAccuracyCLI(t, "cut", "-f", "1-", input, "-o", output)
			reader, err := xopen.Ropen(output)
			if err != nil {
				t.Fatal(err)
			}
			got, err := io.ReadAll(reader)
			closeErr := reader.Close()
			if err != nil || closeErr != nil {
				t.Fatalf("read output: %v; close output: %v", err, closeErr)
			}
			if string(got) != inputText {
				t.Fatalf("output: got %q, want %q", got, inputText)
			}
		})
	}
}
