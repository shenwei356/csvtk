package cmd

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shenwei356/xopen"
)

// TestSplitOutputAliases protects inputs when generated filenames collide in the current directory.
func TestSplitOutputAliases(t *testing.T) {
	const contents = "id,value\na,1\nb,2\n"
	for _, mode := range []string{"lines", "chunks", "keys"} {
		for _, alias := range []string{"same path", "hardlink", "symlink"} {
			t.Run(mode+"/"+alias, func(t *testing.T) {
				dir := t.TempDir()
				name := "1.csv"
				flags := []string{"-n", "1"}
				if mode == "chunks" {
					flags = []string{"-c", "2"}
				} else if mode == "keys" {
					name = "a.csv"
					flags = []string{"-f", "id"}
				}
				output := filepath.Join(dir, name)
				input := output
				if alias != "same path" {
					input = filepath.Join(dir, "source.csv")
				}
				if err := os.WriteFile(input, []byte(contents), 0600); err != nil {
					t.Fatal(err)
				}
				if alias != "same path" {
					link := os.Link
					if alias == "symlink" {
						link = os.Symlink
					}
					if err := link(input, output); err != nil {
						t.Fatal(err)
					}
				}
				args := append([]string{"-test.run=^TestCLIAccuracyHelper$", "--", "split", "-p", ""}, flags...)
				command := exec.Command(os.Args[0], append(args, input)...)
				command.Dir = dir
				command.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
				out, err := command.CombinedOutput()
				if err == nil || !strings.Contains(string(out), "same file") {
					t.Fatalf("alias was not rejected: %v: %s", err, out)
				}
				got, err := os.ReadFile(input)
				if err != nil || string(got) != contents {
					t.Fatalf("input changed: %q, %v", got, err)
				}
			})
		}
	}
}

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

// TestSplitRejectsUnsafeDirectories preserves all files before rejecting aliases
// of the working directory and directories containing inputs.
func TestSplitRejectsUnsafeDirectories(t *testing.T) {
	for _, name := range []string{"cwd", "absolute cwd", "parent alias", "missing parent alias", "cwd symlink", "ancestor", "input parent", "input ancestor", "input target", "input link", "input directory link", "output link", "file list"} {
		for _, force := range []bool{false, true} {
			t.Run(name+"/force="+strconv.FormatBool(force), func(t *testing.T) {
				root := t.TempDir()
				cwd := filepath.Join(root, "work")
				data := filepath.Join(root, "data")
				links := filepath.Join(root, "links")
				for _, dir := range []string{cwd, data, links, filepath.Join(cwd, "child")} {
					if err := os.MkdirAll(dir, 0700); err != nil {
						t.Fatal(err)
					}
				}
				input := filepath.Join(data, "input.csv")
				marker := filepath.Join(cwd, "keep.txt")
				list := filepath.Join(data, "files.txt")
				originals := map[string]string{input: "id\na\n", marker: "keep\n", list: input + "\n"}
				for file, content := range originals {
					if err := os.WriteFile(file, []byte(content), 0600); err != nil {
						t.Fatal(err)
					}
				}
				output := "."
				args := []string{"split", "-f", "id"}
				source := input
				want := "current directory"
				switch name {
				case "absolute cwd":
					output = cwd
				case "parent alias":
					output = "child/.."
				case "missing parent alias":
					output = "missing/.."
				case "cwd symlink":
					output = filepath.Join(root, "alias")
					if err := os.Symlink(cwd, output); err != nil {
						t.Fatal(err)
					}
				case "ancestor":
					output = root
				case "input parent", "input ancestor":
					output, want = data, "contains input"
					if name == "input ancestor" {
						source = filepath.Join(data, "nested", "input.csv")
						if err := os.Mkdir(filepath.Dir(source), 0700); err != nil {
							t.Fatal(err)
						}
						originals[source] = "id\na\n"
						if err := os.WriteFile(source, []byte(originals[source]), 0600); err != nil {
							t.Fatal(err)
						}
					}
				case "input target", "input link":
					source = filepath.Join(links, "input.csv")
					if err := os.Symlink(input, source); err != nil {
						t.Fatal(err)
					}
					output, want = data, "contains input"
					if name == "input link" {
						output = links
					}
				case "input directory link":
					link := filepath.Join(links, "data")
					if err := os.Symlink(data, link); err != nil {
						t.Fatal(err)
					}
					source = filepath.Join(link, "input.csv")
					output, want = links, "contains input"
				case "output link":
					output, want = filepath.Join(root, "output-link"), "contains input"
					if err := os.Symlink(data, output); err != nil {
						t.Fatal(err)
					}
				case "file list":
					source = filepath.Join(links, "source.csv")
					originals[source] = "id\na\n"
					originals[list] = source + "\n"
					for _, file := range []string{source, list} {
						if err := os.WriteFile(file, []byte(originals[file]), 0600); err != nil {
							t.Fatal(err)
						}
					}
					args = append(args, "-X", list)
					output, want = data, "contains input"
				}
				if name != "file list" {
					args = append(args, source)
				}
				args = append(args, "-o", output)
				if force {
					args = append(args, "--force")
				}
				cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
				cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
				cmd.Dir = cwd
				out, err := cmd.CombinedOutput()
				if err == nil || !strings.Contains(string(out), want) || strings.Contains(string(out), "removing old output") {
					t.Fatalf("expected rejection before modification, got %v: %s", err, out)
				}
				for file, want := range originals {
					got, err := os.ReadFile(file)
					if err != nil || string(got) != want {
						t.Fatalf("file changed: %s: %q, %v", file, got, err)
					}
				}
				if name == "input target" || name == "input link" {
					if _, err := os.Lstat(source); err != nil {
						t.Fatalf("input link removed: %v", err)
					}
				}
			})
		}
	}
}

// TestSplitSafeDirectories retains default output and forced replacement of a
// separate directory whose name merely shares a prefix with the input directory.
func TestSplitSafeDirectories(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(strconv.FormatBool(explicit), func(t *testing.T) {
			root := t.TempDir()
			cwd := filepath.Join(root, "work")
			if err := os.Mkdir(cwd, 0700); err != nil {
				t.Fatal(err)
			}
			input := filepath.Join(root, "input.csv")
			if err := os.WriteFile(input, []byte("id\na\n"), 0600); err != nil {
				t.Fatal(err)
			}
			output := cwd
			args := []string{"split", "-n", "1", input}
			if explicit {
				output = filepath.Join(root, "input")
				if err := os.Mkdir(output, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(output, "old.txt"), []byte("old"), 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "-o", output, "--force")
			}
			cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
			cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
			cmd.Dir = cwd
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("split: %v: %s", err, out)
			}
			got, err := os.ReadFile(filepath.Join(output, "input-1.csv"))
			if err != nil || string(got) != "id\na\n" {
				t.Fatalf("output: %q, %v", got, err)
			}
			if explicit {
				if _, err := os.Stat(filepath.Join(output, "old.txt")); !os.IsNotExist(err) {
					t.Fatalf("old output retained: %v", err)
				}
			}
		})
	}
}

// TestSplitStreamsBeforeEOF verifies per-group limits and immediate chunk output
// while stdin remains open, then checks appends after the reader reaches EOF.
func TestSplitStreamsBeforeEOF(t *testing.T) {
	manyChunks := make(map[string]string)
	for i := 1; i <= 10; i++ {
		manyChunks["stdin-"+strconv.Itoa(i)+".csv"] = "k,v\n"
	}
	manyChunks["stdin-1.csv"] += "A,one\n"
	manyChunks["stdin-2.csv"] += "A,two\n"
	manyChunks["stdin-3.csv"] += "B,three\n"
	for _, tc := range []struct {
		name  string
		flags []string
		first string
		want  map[string]string
	}{
		{"single row buffer", []string{"-b", "1"}, "stdin-A.csv", map[string]string{"stdin-A.csv": "k,v\nA,one\nA,two\n", "stdin-B.csv": "k,v\nB,three\n"}},
		{"many immediate chunks", []string{"-c", "10", "--immediate-output"}, "stdin-1.csv", manyChunks},
		{"immediate chunks", []string{"-c", "2", "--immediate-output"}, "stdin-1.csv", map[string]string{"stdin-1.csv": "k,v\nA,one\nB,three\n", "stdin-2.csv": "k,v\nA,two\n"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "out")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			args := append([]string{"-test.run=^TestCLIAccuracyHelper$", "--", "split", "-o", dir}, tc.flags...)
			command := exec.CommandContext(ctx, os.Args[0], args...)
			command.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
			var stderr bytes.Buffer
			command.Stderr = &stderr
			stdin, err := command.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			waited := false
			defer func() {
				stdin.Close()
				cancel()
				if !waited {
					command.Wait()
				}
			}()
			if _, err := io.WriteString(stdin, "k,v\nA,one\n"); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, tc.first)
			for {
				data, _ := os.ReadFile(path)
				if string(data) == "k,v\nA,one\n" {
					break
				}
				select {
				case <-ctx.Done():
					t.Fatal("split did not write before EOF")
				case <-time.After(10 * time.Millisecond):
				}
			}
			if _, err := io.WriteString(stdin, "A,two\nB,three\n"); err != nil {
				t.Fatal(err)
			}
			stdin.Close()
			err = command.Wait()
			waited = true
			if err != nil {
				t.Fatalf("split: %v: %s", err, stderr.String())
			}
			assertSplitFiles(t, dir, tc.want)
		})
	}
}

// TestSplitChunkCompressedBatches verifies all concatenated compression streams,
// exact row order and a single header across multiple buffer flushes.
func TestSplitChunkCompressedBatches(t *testing.T) {
	for _, suffix := range []string{".csv", ".csv.gz", ".csv.xz", ".csv.zst", ".csv.bz2", ".csv.lz4"} {
		t.Run(suffix, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "input"+suffix)
			w, err := xopen.Wopen(input)
			if err != nil {
				t.Fatal(err)
			}
			expected := make([]strings.Builder, 10)
			for i := range expected {
				expected[i].WriteString("k,v\n")
			}
			if _, err := io.WriteString(w, "k,v\n"); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 9000; i++ {
				row := strconv.Itoa(i) + ",value\n"
				if _, err := io.WriteString(w, row); err != nil {
					t.Fatal(err)
				}
				expected[i%10].WriteString(row)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(dir, "out")
			runAccuracyCLI(t, "split", "-c", "10", "-o", out, input)
			entries, err := os.ReadDir(out)
			if err != nil || len(entries) != 10 {
				t.Fatalf("outputs: %v, %v", entries, err)
			}
			for _, entry := range entries {
				base := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				if suffix != ".csv" {
					base = strings.TrimSuffix(base, ".csv")
				}
				index := strings.LastIndex(base, "-")
				chunk, err := strconv.Atoi(base[index+1:])
				if err != nil {
					t.Fatal(err)
				}
				i := chunk - 1
				r, err := xopen.Ropen(filepath.Join(out, entry.Name()))
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
				if string(data) != expected[i].String() {
					t.Fatalf("%s row order, values or header changed", entry.Name())
				}
			}
		})
	}
}

// TestSplitDeleteHeader applies header suppression to every splitting mode.
func TestSplitDeleteHeader(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "key,value\nA,1\nB,2\nA,3\n")
	for _, mode := range [][]string{{"-f", "1"}, {"-n", "2"}, {"-c", "2"}} {
		baseline, without := filepath.Join(t.TempDir(), "with"), filepath.Join(t.TempDir(), "without")
		runAccuracyCLI(t, append(append([]string{"split", "-o", baseline}, mode...), input)...)
		runAccuracyCLI(t, append(append([]string{"split", "-U", "-o", without}, mode...), input)...)
		entries, err := os.ReadDir(baseline)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			expected, err := os.ReadFile(filepath.Join(baseline, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			actual, err := os.ReadFile(filepath.Join(without, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			want := strings.TrimPrefix(string(expected), "key,value\n")
			if string(actual) != want {
				t.Fatalf("%v %s: got %q; want %q", mode, entry.Name(), actual, want)
			}
		}
	}
}
