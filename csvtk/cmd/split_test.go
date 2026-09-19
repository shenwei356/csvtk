package cmd

import (
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shenwei356/xopen"
)

func TestSplitByNLines(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b,c,d\n1,2,3,4\n2,3,4,5\n3,4,5,6\n4,5,6,7\n5,6,7,8\n")
	outdir := filepath.Join(t.TempDir(), "chunks")
	runAccuracyCLI(t, "split", "-n", "2", "-o", outdir, input)

	want := map[string]string{
		"input-1.csv": "a,b,c,d\n1,2,3,4\n2,3,4,5\n",
		"input-2.csv": "a,b,c,d\n3,4,5,6\n4,5,6,7\n",
		"input-3.csv": "a,b,c,d\n5,6,7,8\n",
	}
	assertSplitFiles(t, outdir, want)
}

func TestSplitByNLinesHeaderOptions(t *testing.T) {
	t.Run("no header", func(t *testing.T) {
		input := writeAccuracyInput(t, "input.csv", "1,one\n2,two\n3,three\n")
		outdir := filepath.Join(t.TempDir(), "chunks")
		runAccuracyCLI(t, "split", "-H", "--nlines", "2", "-p", "part-", "-o", outdir, input)
		assertSplitFiles(t, outdir, map[string]string{
			"part-1.csv": "1,one\n2,two\n",
			"part-2.csv": "3,three\n",
		})
	})

	t.Run("header only", func(t *testing.T) {
		input := writeAccuracyInput(t, "input.csv", "id,value\n")
		outdir := filepath.Join(t.TempDir(), "chunks")
		runAccuracyCLI(t, "split", "--nlines", "2", "-o", outdir, input)
		entries, err := os.ReadDir(outdir)
		if err != nil || len(entries) != 0 {
			t.Fatalf("header-only input should create no chunks: %v, %v", entries, err)
		}
	})
}

func TestSplitByNLinesCountsCSVRecords(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "id,value\n1,\"line one\nline two\"\n2,two\n3,three\n")
	outdir := filepath.Join(t.TempDir(), "chunks")
	runAccuracyCLI(t, "split", "--nlines", "2", "-o", outdir, input)
	assertSplitFiles(t, outdir, map[string]string{
		"input-1.csv": "id,value\n1,\"line one\nline two\"\n2,two\n",
		"input-2.csv": "id,value\n3,three\n",
	})
}

func TestSplitByNLinesFromStdinAndGzip(t *testing.T) {
	outdir := filepath.Join(t.TempDir(), "chunks")
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "split", "--nlines", "2", "-G", "-o", outdir, "-")
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	cmd.Stdin = strings.NewReader("id,value\n1,one\n2,two\n3,three\n")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("split stdin: %v: %s", err, output)
	}

	want := map[string]string{
		"stdin-1.csv.gz": "id,value\n1,one\n2,two\n",
		"stdin-2.csv.gz": "id,value\n3,three\n",
	}
	entries, err := os.ReadDir(outdir)
	if err != nil || len(entries) != len(want) {
		t.Fatalf("got %d output files, want %d: %v", len(entries), len(want), err)
	}
	for name, expected := range want {
		fh, err := os.Open(filepath.Join(outdir, name))
		if err != nil {
			t.Fatal(err)
		}
		reader, err := gzip.NewReader(fh)
		if err != nil {
			fh.Close()
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		fileCloseErr := fh.Close()
		if readErr != nil || closeErr != nil || fileCloseErr != nil {
			t.Fatalf("read %s: %v, %v, %v", name, readErr, closeErr, fileCloseErr)
		}
		if string(data) != expected {
			t.Errorf("%s: got %q, want %q", name, data, expected)
		}
	}
}

func TestSplitByNLinesPreservesCompressedOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.csv.zst")
	fh, err := xopen.Wopen(input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fh.Write([]byte("id,value\n1,one\n2,two\n3,three\n")); err != nil {
		t.Fatal(err)
	}
	if err = fh.Close(); err != nil {
		t.Fatal(err)
	}

	outdir := filepath.Join(dir, "chunks")
	runAccuracyCLI(t, "split", "--nlines", "2", "-o", outdir, input)
	want := map[string]string{
		"input.csv-1.zst": "id,value\n1,one\n2,two\n",
		"input.csv-2.zst": "id,value\n3,three\n",
	}
	entries, err := os.ReadDir(outdir)
	if err != nil || len(entries) != len(want) {
		t.Fatalf("got %d output files, want %d: %v", len(entries), len(want), err)
	}
	for name, expected := range want {
		reader, err := xopen.Ropen(filepath.Join(outdir, name))
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read %s: %v, %v", name, readErr, closeErr)
		}
		if string(data) != expected {
			t.Errorf("%s: got %q, want %q", name, data, expected)
		}
	}
}

func TestSplitByNChunksRoundRobin(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "id,value\n1,one\n2,two\n3,three\n4,four\n5,five\n6,six\n7,seven\n")
	outdir := filepath.Join(t.TempDir(), "chunks")
	runAccuracyCLI(t, "split", "-c", "3", "-o", outdir, input)
	assertSplitFiles(t, outdir, map[string]string{
		"input-1.csv": "id,value\n1,one\n4,four\n7,seven\n",
		"input-2.csv": "id,value\n2,two\n5,five\n",
		"input-3.csv": "id,value\n3,three\n6,six\n",
	})
}

func TestSplitByNChunksCreatesEveryChunk(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "id,value\n1,one\n2,two\n")
	outdir := filepath.Join(t.TempDir(), "chunks")
	runAccuracyCLI(t, "split", "--nchunks", "4", "-o", outdir, input)
	assertSplitFiles(t, outdir, map[string]string{
		"input-1.csv": "id,value\n1,one\n",
		"input-2.csv": "id,value\n2,two\n",
		"input-3.csv": "id,value\n",
		"input-4.csv": "id,value\n",
	})
}

func TestSplitByNChunksWithoutHeader(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "1,one\n2,two\n3,three\n")
	outdir := filepath.Join(t.TempDir(), "chunks")
	runAccuracyCLI(t, "split", "-H", "-c", "2", "-o", outdir, input)
	assertSplitFiles(t, outdir, map[string]string{
		"input-1.csv": "1,one\n3,three\n",
		"input-2.csv": "2,two\n",
	})
}

func TestSplitByRecordCountRejectsInvalidOptions(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "id\n1\n")
	for _, args := range [][]string{
		{"split", "--nlines", "0", input},
		{"split", "--nlines", "-1", input},
		{"split", "--nlines", "2", "-f", "id", input},
		{"split", "--nlines", "2", "-F", input},
		{"split", "--nlines", "2", "-i", input},
		{"split", "--nlines", "2", "-b", "10", input},
		{"split", "--nlines", "2", "-g", "10", input},
		{"split", "--nlines", "2", "-s", "1", input},
		{"split", "--nchunks", "0", input},
		{"split", "--nchunks", "-1", input},
		{"split", "--nchunks", "2", "--nlines", "2", input},
		{"split", "--nchunks", "2", "-f", "id", input},
		{"split", "--nchunks", "2", "-F", input},
		{"split", "--nchunks", "2", "-i", input},
		{"split", "--nchunks", "2", "-b", "10", input},
		{"split", "--nchunks", "2", "-g", "10", input},
		{"split", "--nchunks", "2", "-s", "1", input},
	} {
		cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
		cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
		if err := cmd.Run(); err == nil {
			t.Errorf("csvtk %v unexpectedly succeeded", args)
		}
	}
}

func assertSplitFiles(t *testing.T, dir string, want map[string]string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(want) {
		t.Fatalf("got %d output files, want %d", len(entries), len(want))
	}
	for name, expected := range want {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != expected {
			t.Errorf("%s: got %q, want %q", name, data, expected)
		}
	}
}
