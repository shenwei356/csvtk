package cmd

import (
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestCSVReaderFileLimitHelper applies a descriptor limit inside the child,
// after Go startup, without changing limits in the main test process.
func TestCSVReaderFileLimitHelper(t *testing.T) {
	if os.Getenv("CSVTK_READER_FILE_LIMIT_HELPER") != "1" {
		return
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{Cur: 64, Max: 64}); err != nil {
		t.Fatal(err)
	}
	for i, arg := range os.Args {
		if arg == "--" {
			RootCmd.SetArgs(os.Args[i+1:])
			if err := RootCmd.Execute(); err != nil {
				t.Fatal(err)
			}
			os.Exit(0)
		}
	}
	t.Fatal("missing command arguments")
}

// TestEarlyCSVCommandsFileLimit requires readers to close before processing the
// next input, including gzip inputs whose decoders can prefetch records.
func TestEarlyCSVCommandsFileLimit(t *testing.T) {
	for _, compressed := range []bool{false, true} {
		data := []byte("a,b\n" + strings.Repeat("1,2\n", 1024))
		suffix := ".csv"
		if compressed {
			suffix += ".gz"
			var buffer bytes.Buffer
			writer := gzip.NewWriter(&buffer)
			if _, err := writer.Write(data); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			data = buffer.Bytes()
		}
		dir := t.TempDir()
		files := make([]string, 100)
		for i := range files {
			files[i] = filepath.Join(dir, strconv.Itoa(i)+suffix)
			if err := os.WriteFile(files[i], data, 0600); err != nil {
				t.Fatal(err)
			}
		}
		cases := []struct {
			name string
			args []string
			want string
		}{
			{"head", []string{"head", "-n", "1"}, strings.Repeat("a,b\n1,2\n", len(files))},
			{"ncol", []string{"ncol"}, strings.Repeat("2\n", len(files))},
			{"dim columns", []string{"dim", "--cols", "-n"}, strings.Repeat("2\n", len(files))},
			{"headers", []string{"headers"}, strings.Repeat("a\nb\n", len(files))},
		}
		for _, tc := range cases {
			t.Run(tc.name+suffix, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				args := append([]string{"-test.run=^TestCSVReaderFileLimitHelper$", "--"}, tc.args...)
				cmd := exec.CommandContext(ctx, os.Args[0], append(args, files...)...)
				cmd.Env = append(os.Environ(), "CSVTK_READER_FILE_LIMIT_HELPER=1")
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				if err := cmd.Run(); err != nil {
					t.Fatalf("csvtk %v: %v\n%s", tc.args, err, stderr.String())
				}
				if stdout.String() != tc.want {
					t.Fatalf("unexpected output for 100 files: %q", stdout.String())
				}
				if stderr.Len() != 0 {
					t.Fatalf("unexpected diagnostics: %s", stderr.String())
				}
			})
		}
	}
}

// TestSpace2TabFileLimit requires explicit closure instead of relying on file finalizers.
func TestSpace2TabFileLimit(t *testing.T) {
	for _, compressed := range []bool{false, true} {
		contents := "# comment\na b\n1 2\n"
		input := writeAccuracyInput(t, "spaces.txt", contents)
		if compressed {
			input = writeAuxiliaryGzip(t, contents, "")
		}
		args := []string{"-test.run=^TestCSVReaderFileLimitHelper$", "--", "space2tab", "-b", "1K"}
		for i := 0; i < 100; i++ {
			args = append(args, input)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		cmd := exec.CommandContext(ctx, os.Args[0], args...)
		cmd.Env = append(os.Environ(), "CSVTK_READER_FILE_LIMIT_HELPER=1", "GOGC=off")
		out, err := cmd.CombinedOutput()
		cancel()
		if err != nil || string(out) != strings.Repeat("a\tb\n1\t2\n", 100) {
			t.Fatalf("100 files, compressed=%t: %v: %s", compressed, err, out)
		}
	}
}
