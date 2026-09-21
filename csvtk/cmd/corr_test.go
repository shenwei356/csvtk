package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCorrOutputStreams(t *testing.T) {
	input := writeAccuracyInput(t, "input.tsv", "A\tB\n1\t2\n3\t4\n")
	const report = "A\tB\t1.0000\n"
	const passed = "A\tB\n1\t2\n3\t4\n"

	stdout, stderr := runAccuracyCLI(t, "-t", "corr", "-f", "A,B", input)
	if stdout != report || stderr != "" {
		t.Fatalf("default output: stdout %q, stderr %q", stdout, stderr)
	}

	stdout, stderr = runAccuracyCLI(t, "-t", "corr", "-f", "A,B", "--pass", input)
	if stdout != passed || stderr != report {
		t.Fatalf("passthrough output: stdout %q, stderr %q", stdout, stderr)
	}

	outfile := filepath.Join(t.TempDir(), "report.tsv")
	stdout, stderr = runAccuracyCLI(t, "-t", "corr", "-f", "A,B", "-o", outfile, input)
	if stdout != "" || stderr != "" {
		t.Fatalf("file output: stdout %q, stderr %q", stdout, stderr)
	}
	data, err := os.ReadFile(outfile)
	if err != nil || string(data) != report {
		t.Fatalf("report file: %q, error %v", data, err)
	}

	outfile = filepath.Join(t.TempDir(), "passed.tsv")
	stdout, stderr = runAccuracyCLI(t, "-t", "corr", "-f", "A,B", "--pass", "-o", outfile, input)
	if stdout != "" || stderr != report {
		t.Fatalf("passthrough file output: stdout %q, stderr %q", stdout, stderr)
	}
	data, err = os.ReadFile(outfile)
	if err != nil || string(data) != passed {
		t.Fatalf("passthrough file: %q, error %v", data, err)
	}
}
