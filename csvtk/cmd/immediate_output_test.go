package cmd

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// assertCLIStreamsBeforeEOF requires a prefix before closing stdin, then checks
// the remaining output after more input. A deadline bounds blocked subprocesses.
func assertCLIStreamsBeforeEOF(t *testing.T, args []string, input, prefixText, later, remaining string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	helperArgs := append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], helperArgs...)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
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
	// Write concurrently so the large case cannot block on pipe capacity.
	written := make(chan error, 1)
	go func() {
		_, err := io.WriteString(stdin, input)
		written <- err
	}()
	prefix := make([]byte, len(prefixText))
	if _, err := io.ReadFull(stdout, prefix); err != nil {
		t.Fatalf("no output before stdin EOF: %v", err)
	}
	if string(prefix) != prefixText {
		t.Fatalf("streamed prefix: got %q, want %q", prefix, prefixText)
	}
	if err := <-written; err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(stdin, later); err != nil {
		t.Fatal(err)
	}
	stdin.Close()
	rest, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait()
	waited = true
	if err != nil {
		t.Fatalf("csvtk %v failed: %v; stderr: %s", args, err, stderr.String())
	}
	if string(rest) != remaining {
		t.Fatalf("remaining output: got %q, want %q", rest, remaining)
	}
}

// TestGlobalImmediateOutput covers inherited flags on either side of commands
// and the independent CSV, quoted CSV, JSON, plain-text, and raw output paths.
func TestGlobalImmediateOutput(t *testing.T) {
	for _, tc := range []struct {
		name                            string
		args                            []string
		input, prefix, later, remaining string
	}{
		{"cut", []string{"--immediate-output", "cut", "-f", "2"}, "key,value\nA,first\n", "value\nfirst\n", "B,last\n", "last\n"},
		{"quoted", []string{"cut", "-f", "2", "--immediate-output", "--quote-all", "-T"}, "key,value\nA,first\n", "\"value\"\n\"first\"\n", "B,last\n", "\"last\"\n"},
		{"grep", []string{"--immediate-output", "grep", "-f", "1", "-p", "A"}, "key,value\nA,first\n", "key,value\nA,first\n", "B,skip\nA,last\n", "A,last\n"},
		{"grep header", []string{"grep", "--immediate-output", "-p", "A"}, "key,value\n", "key,value\n", "A,last\n", "A,last\n"},
		{"filter", []string{"filter", "--immediate-output", "-f", "1>0"}, "number,value\n1,first\n", "number,value\n1,first\n", "0,skip\n2,last\n", "2,last\n"},
		{"concat", []string{"concat", "--immediate-output"}, "key,value\nA,first\n", "key,value\nA,first\n", "B,last\n", "B,last\n"},
		{"csv2tab", []string{"csv2tab", "--immediate-output"}, "key,value\nA,first\n", "key\tvalue\nA\tfirst\n", "B,last\n", "B\tlast\n"},
		{"json", []string{"csv2json", "--immediate-output", "--indent", ""}, "key,value\nA,first\n", `[{"key":"A","value":"first"}`, "B,last\n", `,{"key":"B","value":"last"}]` + "\n"},
		{"space2tab", []string{"space2tab", "--immediate-output", "-b", "1K"}, "key value\nA first\n", "key\tvalue\nA\tfirst\n", "B last\n", "B\tlast\n"},
		{"del quotes", []string{"del-quotes", "--immediate-output"}, "key,value\nA,\"first\"\n", "key,value\nA,first\n", "B,\"last\"\n", "B,last\n"},
		{"fix quotes", []string{"fix-quotes", "--immediate-output", "-b", "1K"}, "key,value\nA,first\n", "key,value\nA,first\n", "B,last\n", "B,last\n"},
		{"fix quotes comment", []string{"fix-quotes", "--immediate-output", "-b", "1K"}, "# comment\n", "# comment\n", "key,value\nA,first\n", "key,value\nA,first\n"},
		{"cat bytes", []string{"cat", "--immediate-output", "-s", "20", "-p", "-1"}, "first\n", "first\n", "last", "last"},
		{"cat lines", []string{"cat", "--immediate-output", "-L", "-s", "2", "-p", "-1"}, "first\n", "first\n", "last", "last"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertCLIStreamsBeforeEOF(t, tc.args, tc.input, tc.prefix, tc.later, tc.remaining)
		})
	}
}

// TestGlobalImmediateFormattedOutput checks streaming after format-specific
// setup, while requiring the same bytes as a completed, buffered invocation.
func TestGlobalImmediateFormattedOutput(t *testing.T) {
	const input = "key,value\nA,first\nB,last\n"
	file := writeAccuracyInput(t, "formatted.csv", input)
	for _, args := range [][]string{{"csv2html"}, {"pretty", "-n", "1", "-S", "plain"}} {
		t.Run(args[0], func(t *testing.T) {
			want, _ := runAccuracyCLI(t, append(args, file)...)
			prefix, suffix := want, ""
			if args[0] == "csv2html" {
				prefix, suffix, _ = strings.Cut(want, "      </tbody>")
				suffix = "      </tbody>" + suffix
			}
			assertCLIStreamsBeforeEOF(t, append(args, "--immediate-output"), input, prefix, "", suffix)
		})
	}
}

// TestGlobalImmediateFileOutput catches flushing only the CSV buffer when -o
// adds another buffer, including grep's separate file-output branch.
func TestGlobalImmediateFileOutput(t *testing.T) {
	const input = "key,value\nA,first\n"
	for _, args := range [][]string{{"cut", "-f", "1-"}, {"grep", "-p", "A"}, {"uniq"}, {"split", "-n", "100", "-p", "chunk-"}} {
		t.Run(args[0], func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "output.csv")
			destination := output
			if args[0] == "split" {
				destination = filepath.Dir(output)
				output = filepath.Join(destination, "chunk-1.csv")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			helperArgs := append([]string{"-test.run=^TestCLIAccuracyHelper$", "--", "--immediate-output"}, args...)
			cmd := exec.CommandContext(ctx, os.Args[0], append(helperArgs, "-o", destination)...)
			cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Start(); err != nil {
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
			if _, err := io.WriteString(stdin, input); err != nil {
				t.Fatal(err)
			}
			// Observe the destination while input remains open; closing stdin
			// first would hide a missing downstream flush.
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			for {
				data, err := os.ReadFile(output)
				if err == nil && string(data) == input {
					break
				}
				if err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
				select {
				case <-ctx.Done():
					t.Fatalf("file output not flushed before EOF: %q (%v)", data, err)
				case <-ticker.C:
				}
			}
			stdin.Close()
			err = cmd.Wait()
			waited = true
			if err != nil {
				t.Fatalf("csvtk %v: %v; stderr: %s", args, err, stderr.String())
			}
			data, err := os.ReadFile(output)
			if err != nil || string(data) != input {
				t.Fatalf("completed output: %q (%v)", data, err)
			}
		})
	}
}

// TestGlobalImmediateOutputPreservesResults also covers commands that emit
// records only after reading the entire input or an initial sample.
func TestGlobalImmediateOutputPreservesResults(t *testing.T) {
	input := writeAccuracyInput(t, "groups.csv", "group,value\nA,1\nB,2\nA,3\n")
	for _, args := range [][]string{
		{"sort", "-k", "value:n", input},
		{"freq", "-f", "group", input},
		{"summary", "-g", "group", "-f", "value:sum", input},
		{"fold", "-f", "group", "-v", "value", input},
		{"join", "-f", "group", input, input},
		{"paste", input, input},
		{"comb", input},
		{"fix", "-n", "1", input},
		{"csv2rst", input},
		{"csv2md", input},
		{"headers", input},
		{"dim", "--tabular", input},
	} {
		t.Run(args[0], func(t *testing.T) {
			want, _ := runAccuracyCLI(t, args...)
			got, _ := runAccuracyCLI(t, append([]string{"--immediate-output"}, args...)...)
			if got != want {
				t.Fatalf("output changed: got %q, want %q", got, want)
			}
		})
	}
}

// TestCSVOutputWriterImmediate flushes both layers without changing quoting or
// forcing default buffered output to flush early.
func TestCSVOutputWriterImmediate(t *testing.T) {
	for _, quoted := range []bool{false, true} {
		for _, immediate := range []bool{false, true} {
			var output bytes.Buffer
			downstream := bufio.NewWriter(&output)
			w := newCSVOutputWriter(downstream, csvOutputOption{QuoteAll: quoted, ImmediateOutput: immediate})
			if err := w.Write([]string{"a", "b"}); err != nil {
				t.Fatal(err)
			}
			want := "a,b\n"
			if quoted {
				want = "\"a\",\"b\"\n"
			}
			if immediate && output.String() != want {
				t.Fatalf("not flushed: %q", output.String())
			}
			if !immediate && output.Len() != 0 {
				t.Fatal("default output was flushed early")
			}
			w.Flush()
			if err := w.Error(); err != nil {
				t.Fatal(err)
			}
			if err := downstream.Flush(); err != nil {
				t.Fatal(err)
			}
			if output.String() != want {
				t.Fatalf("got %q, want %q", output.String(), want)
			}
		}
	}
}

// flushFailureWriter accepts bytes but fails when flushing to the destination.
type flushFailureWriter struct {
	bytes.Buffer
	err error // Destination failure returned by Flush.
}

// Flush simulates a destination failure without rejecting buffered writes.
func (w *flushFailureWriter) Flush() error { return w.err }

// TestCSVOutputWriterImmediateFlushError preserves downstream failures even
// when a caller checks Error or writes again after the first failure.
func TestCSVOutputWriterImmediateFlushError(t *testing.T) {
	want := errors.New("destination flush failed")
	for _, quoted := range []bool{false, true} {
		output := &flushFailureWriter{err: want}
		w := newCSVOutputWriter(output, csvOutputOption{QuoteAll: quoted, ImmediateOutput: true})
		if err := w.Write([]string{"first"}); !errors.Is(err, want) {
			t.Fatalf("Write: %v", err)
		}
		n := output.Len()
		w.Flush()
		if err := w.Error(); !errors.Is(err, want) {
			t.Fatalf("Error: %v", err)
		}
		if err := w.Write([]string{"later"}); !errors.Is(err, want) {
			t.Fatalf("later Write: %v", err)
		}
		if output.Len() != n {
			t.Fatal("wrote more data after a flush failure")
		}
	}
}
