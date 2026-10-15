package cmd

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shenwei356/xopen"
)

// csvReaderTestInput tracks input closure and can delay its completion.
type csvReaderTestInput struct {
	io.Reader
	closing chan struct{} // Signals that Close has begun.
	closed  chan struct{} // Signals that Close has completed.
	release chan struct{} // An optional gate for a slow Close.
}

// Close records the lifetime of the input owned by CSVReader.
func (input *csvReaderTestInput) Close() error {
	close(input.closing)
	if input.release != nil {
		<-input.release
	}
	close(input.closed)
	return nil
}

// TestCSVReaderRecordLimit checks normal, missing-column, and unlimited reads,
// and requires the producer to release inputs without waiting for consumption.
func TestCSVReaderRecordLimit(t *testing.T) {
	cases := []struct {
		name     string
		limit    uint64
		missing  bool
		selected [][]string
	}{
		{"header only", 1, false, [][]string{{"a", "b"}}},
		{"header and data", 2, false, [][]string{{"a", "b"}, {"1", "2"}}},
		{"unlimited", 0, false, [][]string{{"a", "b"}, {"1", "2"}, {"3", "4"}}},
		{"limit past EOF", 100, false, [][]string{{"a", "b"}, {"1", "2"}, {"3", "4"}}},
		{"missing column", 2, true, [][]string{{"a", "missing"}, {"1", ""}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := &csvReaderTestInput{Reader: strings.NewReader("a,b\n1,2\n3,4\n"), closing: make(chan struct{}), closed: make(chan struct{})}
			fh, err := xopen.Buf(input)
			if err != nil {
				t.Fatal(err)
			}
			reader := &CSVReader{fh: fh, Reader: csv.NewReader(fh), Ch: make(chan Record, 128)}
			opt := ReadOption{FieldStr: "1-", MaxRecords: tc.limit}
			if tc.missing {
				opt.FieldStr = "a,missing"
				opt.BlankMissingColumn = true
			}
			reader.Read(opt)
			select {
			case <-input.closed:
			case <-time.After(5 * time.Second):
				t.Fatal("reader did not release its input after the record limit or EOF")
			}
			var selected [][]string
			for record := range reader.Ch {
				if record.Err != nil {
					t.Fatal(record.Err)
				}
				selected = append(selected, record.Selected)
			}
			if !reflect.DeepEqual(selected, tc.selected) {
				t.Fatalf("got records %q, want %q", selected, tc.selected)
			}
		})
	}
}

// TestCSVReaderClosesInputBeforeChannel prevents consumers from moving to the
// next file while the previous input is still being closed.
func TestCSVReaderClosesInputBeforeChannel(t *testing.T) {
	gate := make(chan struct{})
	defer func() {
		select {
		case <-gate:
		default:
			close(gate)
		}
	}()
	input := &csvReaderTestInput{Reader: strings.NewReader("a,b\n1,2\n"), closing: make(chan struct{}), closed: make(chan struct{}), release: gate}
	fh, err := xopen.Buf(input)
	if err != nil {
		t.Fatal(err)
	}
	reader := &CSVReader{fh: fh, Reader: csv.NewReader(fh), Ch: make(chan Record, 128)}
	reader.Read(ReadOption{MaxRecords: 1})
	select {
	case <-input.closing:
	case <-time.After(5 * time.Second):
		t.Fatal("input closure did not start")
	}
	if record := <-reader.Ch; !reflect.DeepEqual(record.All, []string{"a", "b"}) {
		t.Fatalf("unexpected first record: %+v", record)
	}
	select {
	case _, ok := <-reader.Ch:
		t.Fatalf("record channel ended before input closure completed (open=%v)", ok)
	default:
	}
	close(gate)
	for record := range reader.Ch {
		t.Fatalf("unexpected extra record: %+v", record)
	}
	select {
	case <-input.closed:
	default:
		t.Fatal("input still open after record channel ended")
	}
}

// TestEarlyCSVCommandsOpenStdin checks that a producer waiting for more input
// cannot prevent a prefix command from exiting after receiving enough records.
func TestEarlyCSVCommandsOpenStdin(t *testing.T) {
	cases := []struct {
		name, input, want string
		args              []string
	}{
		{"head", "alpha,beta\n1,2\n", "alpha,beta\n1,2\n", []string{"head", "-n", "1"}},
		{"ncol", "alpha,beta\n", "2\n", []string{"ncol"}},
		{"dim columns", "alpha,beta\n", "2\n", []string{"dim", "--cols", "-n"}},
		{"headers", "alpha,beta\n", "alpha\nbeta\n", []string{"headers"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			args := append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, tc.args...)
			cmd := exec.CommandContext(ctx, os.Args[0], args...)
			cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			defer stdin.Close()
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			if _, err := io.WriteString(stdin, tc.input); err != nil {
				t.Fatal(err)
			}
			// Keep stdin open through Wait: the command must not wait for EOF.
			if err := cmd.Wait(); err != nil {
				t.Fatalf("command did not finish on an open input pipe: %v\n%s", err, stderr.String())
			}
			if stdout.String() != tc.want || stderr.Len() != 0 {
				t.Fatalf("got stdout %q, stderr %q; want %q", stdout.String(), stderr.String(), tc.want)
			}
		})
	}
}

// TestHeadRecordLimit checks header handling, multiline records, row numbers,
// skipped records, and the largest supported --number value.
func TestHeadRecordLimit(t *testing.T) {
	const text = "a,b\n1,2\n3,4\n5,6\n"
	cases := []struct {
		name, input, want string
		args              []string
	}{
		{"header", text, "a,b\n1,2\n3,4\n", []string{"-n", "2"}},
		{"no header", text, "a,b\n1,2\n", []string{"-H", "-n", "2"}},
		{"delete header", text, "1,2\n", []string{"-U", "-n", "1"}},
		{"row numbers", text, "row,a,b\n1,1,2\n2,3,4\n", []string{"-Z", "-n", "2"}},
		{"multiline record", "a,b\n\"one\ntwo\",1\nnext,2\nlast,3\n", "a,b\n\"one\ntwo\",1\n", []string{"-n", "1"}},
		{"ignored rows", "a,b\n,\nbad\n1,2\n\"unfinished", "a,b\n1,2\n", []string{"-E", "-I", "-n", "1"}},
		{"largest number", text, text, []string{"-n", strconv.Itoa(int(^uint(0) >> 1))}},
		{"header only", "a,b\n", "a,b\n", []string{"-n", "1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := writeAccuracyInput(t, "input.csv", tc.input)
			args := append([]string{"head"}, tc.args...)
			out, stderr := runAccuracyCLI(t, append(args, input)...)
			if out != tc.want {
				t.Fatalf("got %q, want %q", out, tc.want)
			}
			if tc.name == "ignored rows" {
				for _, message := range []string{"1 empty rows ignored", "1 illegal rows ignored"} {
					if !strings.Contains(stderr, message) {
						t.Fatalf("missing diagnostic %q: %s", message, stderr)
					}
				}
			} else if stderr != "" {
				t.Fatalf("unexpected diagnostics: %s", stderr)
			}
		})
	}
}
