package cmd

import (
	"github.com/shenwei356/xopen"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

// TestSplitChunkFileLimit exceeds the process file limit with both plain and gzip chunks.
func TestSplitChunkFileLimit(t *testing.T) {
	for _, compressed := range []bool{false, true} {
		input := writeAccuracyInput(t, "input.csv", "k,v\n1,one\n2,two\n")
		dir := filepath.Join(t.TempDir(), "out")
		args := []string{"-test.run=^TestCSVReaderFileLimitHelper$", "--", "split", "-c", "80", "-o", dir, input}
		suffix := ".csv"
		if compressed {
			args = append(args, "-G")
			suffix += ".gz"
		}
		command := exec.Command(os.Args[0], args...)
		command.Env = append(os.Environ(), "CSVTK_READER_FILE_LIMIT_HELPER=1")
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("split: %v: %s", err, out)
		}
		want := make(map[string]string, 80)
		for i := 1; i <= 80; i++ {
			want["input-"+strconv.Itoa(i)+suffix] = "k,v\n"
		}
		want["input-1"+suffix] += "1,one\n"
		want["input-2"+suffix] += "2,two\n"
		if compressed {
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != len(want) {
				t.Fatalf("outputs: %v, %v", entries, err)
			}
			for name, expected := range want {
				r, err := xopen.Ropen(filepath.Join(dir, name))
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
				if string(data) != expected {
					t.Fatalf("%s: got %q, want %q", name, data, expected)
				}
			}
		} else {
			assertSplitFiles(t, dir, want)
		}
	}
}
