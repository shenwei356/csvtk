package cmd

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shenwei356/xopen"
	"github.com/xuri/excelize/v2"
)

// TestXLSX2CSVListOutput sends sheet listings to the requested plain or compressed output.
func TestXLSX2CSVListOutput(t *testing.T) {
	input := filepath.Join(t.TempDir(), "input.xlsx")
	book := excelize.NewFile()
	if _, err := book.NewSheet("Second"); err != nil {
		t.Fatal(err)
	}
	if err := book.SaveAs(input); err != nil {
		t.Fatal(err)
	}
	if err := book.Close(); err != nil {
		t.Fatal(err)
	}
	want := "index\tsheet\n1\tSheet1\n2\tSecond\n"
	out, _ := runAccuracyCLI(t, "xlsx2csv", "-a", input)
	if out != want {
		t.Fatalf("stdout: %q", out)
	}
	for _, suffix := range []string{".tsv", ".tsv.gz"} {
		output := filepath.Join(t.TempDir(), "sheets"+suffix)
		if err := os.WriteFile(output, []byte("old contents"), 0600); err != nil {
			t.Fatal(err)
		}
		out, _ = runAccuracyCLI(t, "xlsx2csv", "-a", "--immediate-output", "-o", output, input)
		if out != "" {
			t.Fatalf("unexpected stdout: %q", out)
		}
		r, err := xopen.Ropen(output)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		if string(data) != want {
			t.Fatalf("%s: %q", suffix, data)
		}
	}
	if runtime.GOOS == "linux" {
		assertOutputWriteError(t, []string{"xlsx2csv", "-a", "-o", "/dev/full", input}, false, "")
	}
}
