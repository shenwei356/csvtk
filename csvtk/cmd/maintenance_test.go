package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestMixedExpressionReferences binds both spellings regardless of their order or column type.
func TestMixedExpressionReferences(t *testing.T) {
	for _, field := range []string{"x", "1", "123abc"} {
		input := writeAccuracyInput(t, "input.csv", field+"\n1\nabc\n-2\n")
		for _, reverse := range []bool{false, true} {
			left, right := "$"+field, "${"+field+"}"
			if reverse {
				left, right = right, left
			}
			for _, command := range []string{"filter2", "mutate2", "mutate3"} {
				t.Run(fmt.Sprintf("%s/%s/reverse=%t", command, field, reverse), func(t *testing.T) {
					args := []string{command}
					if command == "filter2" {
						args = append(args, "-f", left+" == "+right)
					} else {
						args = append(args, "-n", "result", "-e", left+" + "+right)
					}
					out, _ := runAccuracyCLI(t, append(args, input)...)
					want := [][]string{{field}, {"1"}, {"abc"}, {"-2"}}
					if command != "filter2" {
						want = [][]string{{field, "result"}, {"1", "2.00"}, {"abc", "abcabc"}, {"-2", "-4.00"}}
					}
					if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
						t.Fatalf("got %q, want %q", got, want)
					}
				})
			}
		}
	}
	input := writeAccuracyInput(t, "prefixes.csv", "x,xy\n1,10\n2,20\n")
	for _, command := range []string{"mutate2", "mutate3"} {
		out, _ := runAccuracyCLI(t, command, "-n", "result", "-e", "$x + ${x} + $xy + ${xy}", input)
		if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, [][]string{{"x", "xy", "result"}, {"1", "10", "22.00"}, {"2", "20", "44.00"}}) {
			t.Fatalf("overlapping names: %q", got)
		}
		out, _ = runAccuracyCLI(t, command, "-n", "result", "-e", "'$x ${x}'", input)
		if got := parseAccuracyCSV(t, out); got[1][2] != "1.0 1.0" || got[2][2] != "2.0 2.0" {
			t.Fatalf("quoted references: %q", got)
		}
	}
}

// TestSummaryRandomSeed preserves the historical seeded stream across process invocations.
func TestSummaryRandomSeed(t *testing.T) {
	input := writeAccuracyInput(t, "groups.csv", "g,x\nB,b0\nB,b1\nA,a0\nA,a1\n")
	for _, tc := range []struct{ seed, want string }{
		{"11", "g,x:rand\nA,a0\nB,b1\n"},
		{"12", "g,x:rand\nA,a1\nB,b1\n"},
	} {
		for repeat := 0; repeat < 3; repeat++ {
			out, _ := runAccuracyCLI(t, "summary", "-g", "g", "-f", "x:rand", "-S", tc.seed, input)
			if out != tc.want {
				t.Fatalf("seed %s: got %q, want %q", tc.seed, out, tc.want)
			}
		}
	}
}

// TestCatInvalidFlags rejects invalid values before opening or truncating output.
func TestCatInvalidFlags(t *testing.T) {
	input := writeAccuracyInput(t, "input.txt", "first\nlast\n")
	for _, flags := range [][]string{{"-b", "0"}, {"-b", "-1"}, {"-p", "0"}, {"-p", "-2"}, {"-L", "-s", "2", "-p", "0"}} {
		t.Run(strings.Join(flags, " "), func(t *testing.T) {
			output := writeAccuracyInput(t, "output.txt", "keep\n")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			args := append([]string{"-test.run=^TestCLIAccuracyHelper$", "--", "cat"}, flags...)
			args = append(args, input, "-o", output)
			cmd := exec.CommandContext(ctx, os.Args[0], args...)
			cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
			out, err := cmd.CombinedOutput()
			if err == nil || ctx.Err() != nil || strings.Contains(string(out), "panic:") || !strings.Contains(string(out), "flag --") {
				t.Fatalf("invalid flags were not rejected cleanly: %v: %s", err, out)
			}
			got, err := os.ReadFile(output)
			if err != nil || string(got) != "keep\n" {
				t.Fatalf("invalid flags changed the output: %q; error: %v", got, err)
			}
		})
	}
}

// TestCatProgressCounts includes residual batches, final-only mode, and each file's own size.
func TestCatProgressCounts(t *testing.T) {
	const text = "x,y\n1,a\n2,b\n"
	input := writeAccuracyInput(t, "input.txt", text)
	for _, lines := range []bool{false, true} {
		for _, frequency := range []int{1, 2, 100, -1} {
			args := []string{"cat", "-b", "4", "-p", strconv.Itoa(frequency)}
			count := len(text)
			if lines {
				args = append(args, "-L", "-s", "3")
				count = 3
			}
			out, stderr := runAccuracyCLI(t, append(args, input)...)
			progress := fmt.Sprintf("%d / %d", count, count)
			if out != text || !strings.Contains(stderr, progress) {
				t.Fatalf("%v: output %q; missing %q in %s", args, out, progress, stderr)
			}
		}
	}
	second := writeAccuracyInput(t, "second.txt", "longer second input\n")
	out, stderr := runAccuracyCLI(t, "cat", "-p", "-1", input, second)
	if out != text+"longer second input\n" || !strings.Contains(stderr, "12 / 12") || !strings.Contains(stderr, "20 / 20") {
		t.Fatalf("per-file totals: stdout %q; stderr %s", out, stderr)
	}
}

// TestCSVReaderQuietSkippedRows keeps exact totals and visible warnings without quiet-mode slices.
func TestCSVReaderQuietSkippedRows(t *testing.T) {
	input := writeAccuracyInput(t, "skipped.csv", "a,b\n,\nbad\n1,2\n,\nbad\n")
	for _, verbose := range []bool{false, true} {
		config := Config{Verbose: verbose, Delimiter: ',', IgnoreEmptyRow: true, IgnoreIllegalRow: true}
		reader, err := newCSVReaderByConfig(config, input)
		if err != nil {
			t.Fatal(err)
		}
		reader.Read(ReadOption{FieldStr: "1-"})
		var rows [][]string
		for record := range reader.Ch {
			if record.Err != nil {
				t.Fatal(record.Err)
			}
			rows = append(rows, record.All)
		}
		if reader.SkippedEmptyRows != 2 || reader.SkippedIllegalRows != 2 || !reflect.DeepEqual(rows, [][]string{{"a", "b"}, {"1", "2"}}) {
			t.Fatalf("incorrect skipped-row counts or output: %+v; rows %q", reader, rows)
		}
		if verbose {
			if !reflect.DeepEqual(reader.NumEmptyRows, []int{2, 5}) || !reflect.DeepEqual(reader.NumIllegalRows, []int{3, 6}) {
				t.Fatalf("lost warning line numbers: %+v", reader)
			}
		} else if cap(reader.NumEmptyRows) != 0 || cap(reader.NumIllegalRows) != 0 {
			t.Fatal("quiet reader retained skipped-row numbers")
		}
	}
	loud, warnings := runAccuracyCLI(t, "cut", "-f", "1-", "-E", "-I", input)
	quiet, stderr := runAccuracyCLI(t, "cut", "-f", "1-", "-E", "-I", "--quiet", input)
	if loud != quiet || stderr != "" || !strings.Contains(warnings, "2 empty rows ignored: [2 5]") || !strings.Contains(warnings, "2 illegal rows ignored: [3 6]") {
		t.Fatalf("quiet mode changed output or warnings: loud=%q quiet=%q warnings=%q stderr=%q", loud, quiet, warnings, stderr)
	}
}

// TestMarkdownLongAlignment preserves all alignments at narrow and wide field widths.
func TestMarkdownLongAlignment(t *testing.T) {
	for _, width := range []int{1, 5, 32768} {
		text := strings.Repeat("a", width)
		input := writeAccuracyInput(t, "wide.csv", "x\n"+text+"\n")
		for _, tc := range []struct{ alignment, left, right string }{{"left", ":", "-"}, {"center", ":", ":"}, {"right", "-", ":"}} {
			out, _ := runAccuracyCLI(t, "csv2md", "-a", tc.alignment, input)
			want := "|x\n|" + tc.left + strings.Repeat("-", max(width, 3)-2) + tc.right + "\n|" + text + "\n"
			if out != want {
				t.Fatalf("width %d, %s: markdown differs", width, tc.alignment)
			}
		}
	}
}
