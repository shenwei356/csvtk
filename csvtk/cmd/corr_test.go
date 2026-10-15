package cmd

import (
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestCorrFixtureReference independently computes Pearson's r for the shell fixture.
func TestCorrFixtureReference(t *testing.T) {
	input := filepath.Join("..", "..", "testdata", "corr_data.tsv")
	fh, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	reader := csv.NewReader(fh)
	reader.Comma = '\t'
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	var x, y []float64
	var meanX, meanY float64
	for _, row := range rows[1:] {
		a, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			t.Fatal(err)
		}
		b, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			t.Fatal(err)
		}
		x, y = append(x, a), append(y, b)
		meanX, meanY = meanX+a, meanY+b
	}
	meanX, meanY = meanX/float64(len(x)), meanY/float64(len(y))
	var cross, squareX, squareY float64
	for i := range x {
		a, b := x[i]-meanX, y[i]-meanY
		cross += a * b
		squareX += a * a
		squareY += b * b
	}
	reference := cross / math.Sqrt(squareX*squareY)
	if math.Abs(reference-(-0.002785068221821459)) > 1e-14 {
		t.Fatalf("fixture changed: reference r = %.17g", reference)
	}
	out, _ := runAccuracyCLI(t, "corr", "-t", "-f", "A,B", input)
	reader = csv.NewReader(strings.NewReader(out))
	reader.Comma = '\t'
	rows, err = reader.ReadAll()
	if err != nil || len(rows) != 1 || len(rows[0]) != 3 {
		t.Fatalf("invalid correlation report: %q, error: %v", out, err)
	}
	got, err := strconv.ParseFloat(rows[0][2], 64)
	if err != nil || math.IsNaN(got) || math.Abs(got-reference) > 0.00005 {
		t.Fatalf("reported r = %v, reference = %v, error: %v", got, reference, err)
	}
}

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
