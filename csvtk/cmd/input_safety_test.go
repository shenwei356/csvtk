package cmd

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// assertInputError requires a real command failure because checkError exits the process.
func assertInputError(t *testing.T, want string, args ...string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err == nil {
		t.Fatalf("csvtk %v succeeded; stdout: %q; stderr: %s", args, stdout.String(), stderr.String())
	} else if _, ok := err.(*exec.ExitError); !ok {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), want) {
		t.Fatalf("csvtk %v: want error containing %q, got %s", args, want, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("csvtk %v wrote partial output before reporting the error: %q", args, stdout.String())
	}
}

// TestInputOutputAliases preserves the original bytes for direct paths and filesystem aliases.
func TestInputOutputAliases(t *testing.T) {
	for _, alias := range []string{"same path", "relative path", "symlink", "hardlink", "file list", "later input"} {
		t.Run(alias, func(t *testing.T) {
			const contents = "x,y\n1,a\n2,b\n"
			input := writeAccuracyInput(t, "input.csv", contents)
			output := input
			args := []string{"cut", "-f", "x", input}
			switch alias {
			case "relative path":
				cwd, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				output, err = filepath.Rel(cwd, input)
				if err != nil {
					t.Fatal(err)
				}
			case "symlink", "hardlink":
				output = filepath.Join(filepath.Dir(input), "alias.csv")
				link := os.Link
				if alias == "symlink" {
					link = os.Symlink
				}
				if err := link(input, output); err != nil {
					t.Fatal(err)
				}
			case "file list":
				list := writeAccuracyInput(t, "inputs.txt", input+"\n")
				args = []string{"cut", "-f", "x", "-X", list}
			case "later input":
				first := writeAccuracyInput(t, "first.csv", contents)
				args = []string{"csv2tab", first, input}
			}
			assertInputError(t, "same file", append(args, "-o", output)...)
			got, err := os.ReadFile(input)
			if err != nil || string(got) != contents {
				t.Fatalf("input changed: %q; error: %v", got, err)
			}
		})
	}
	// cat opens outputs independently of xopen and must receive the same protection.
	input := writeAccuracyInput(t, "cat.csv", "x\n1\n")
	assertInputError(t, "same file", "cat", input, "-o", input)
}

// TestInterEmptyInputs treats every empty set alike, independent of file order or headers.
func TestInterEmptyInputs(t *testing.T) {
	for _, noHeader := range []bool{false, true} {
		data := "x,y\n1,a\n2,b\n"
		if noHeader {
			data = "1,a\n2,b\n"
		}
		input := writeAccuracyInput(t, "input.csv", data)
		for _, empty := range []string{"", "\n\n", "# comment\n", "x,y\n"} {
			if noHeader && empty == "x,y\n" {
				continue // A header is a data record in headerless mode.
			}
			for position := 0; position < 3; position++ {
				t.Run(fmt.Sprintf("headerless=%t/empty=%q/position=%d", noHeader, empty, position), func(t *testing.T) {
					file := writeAccuracyInput(t, "empty.csv", empty)
					args := []string{"inter", "-f", "1"}
					if noHeader {
						args = append(args, "-H")
					}
					files := []string{input, input, input}
					files[position] = file
					out, _ := runAccuracyCLI(t, append(args, files...)...)
					if out != "" {
						t.Fatalf("empty intersection produced %q", out)
					}
				})
			}
		}
	}
	input := writeAccuracyInput(t, "values.csv", "x,y\n1,a\n")
	out, _ := runAccuracyCLI(t, "inter", "-f", "x", input, input)
	if out != "x\n1\n" {
		t.Fatalf("nonempty intersection changed: %q", out)
	}
	header := writeAccuracyInput(t, "header.csv", "x,y\n")
	out, _ = runAccuracyCLI(t, "inter", header)
	if out != "" {
		t.Fatalf("header-only input produced %q", out)
	}
}

// TestReadCSVOpenErrors keeps format conversion errors distinct from empty input.
func TestReadCSVOpenErrors(t *testing.T) {
	badGzip := writeAccuracyInput(t, "broken.gz", "\x1f\x8bnot a gzip header")
	for _, command := range []string{"csv2md", "csv2rst"} {
		t.Run(command, func(t *testing.T) {
			assertInputError(t, "gzip", command, badGzip)
			assertInputError(t, "", command, t.TempDir())
			empty := writeAccuracyInput(t, "empty.csv", "")
			out, _ := runAccuracyCLI(t, command, empty)
			if out != "" {
				t.Fatalf("empty input produced %q", out)
			}
		})
	}
}

// writeAuxiliaryGzip can damage the payload or checksum after valid initial records.
func writeAuxiliaryGzip(t *testing.T, contents, damage string) string {
	t.Helper()
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := io.WriteString(writer, contents); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	data := compressed.Bytes()
	switch damage {
	case "truncated":
		data = data[:len(data)/2]
	case "checksum":
		data[len(data)-8] ^= 1
	}
	file := filepath.Join(t.TempDir(), "auxiliary.gz")
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

// TestAuxiliaryReadErrors rejects partial key maps, match patterns, and sort levels.
func TestAuxiliaryReadErrors(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "key1\nkey999\nkey1\n")
	var kvs, patterns strings.Builder
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&kvs, "key%d\tvalue%d\n", i, i)
		fmt.Fprintf(&patterns, "key%d\n", i)
	}
	for _, damage := range []string{"truncated", "checksum"} {
		kvFile := writeAuxiliaryGzip(t, kvs.String(), damage)
		patternFile := writeAuxiliaryGzip(t, patterns.String(), damage)
		want := "unexpected EOF"
		if damage == "checksum" {
			want = "invalid checksum"
		}
		for _, command := range []string{"replace", "rename2", "grep", "sort"} {
			t.Run(command+"/"+damage, func(t *testing.T) {
				var args []string
				switch command {
				case "replace", "rename2":
					args = []string{command, "-f", "1", "-p", "^(.+)$", "-r", "{kv}", "-k", kvFile, input}
				case "grep":
					args = []string{command, "-f", "1", "-P", patternFile, input}
				case "sort":
					args = []string{command, "-k", "1:u", "-L", "1:" + patternFile, input}
				}
				assertInputError(t, want, args...)
			})
		}
	}
}

// TestReadKVsPreservesLineSemantics covers CRLF, long values, missing tabs, and final EOF.
func TestReadKVsPreservesLineSemantics(t *testing.T) {
	long := strings.Repeat("v", 1<<20)
	contents := "\r\ninvalid\r\na\tfirst\textra\r\na\tlast\textra\r\n\tempty key\r\nb\t\r\nlong\t" + long
	for _, allLeft := range []bool{false, true} {
		want := map[string]string{"a": "last", "": "empty key", "b": "", "long": long}
		if allLeft {
			want["a"] = "last\textra"
		}
		for _, compressed := range []bool{false, true} {
			file := writeAccuracyInput(t, "keys.tsv", contents)
			if compressed {
				file = writeAuxiliaryGzip(t, contents, "")
			}
			got, err := readKVs(file, allLeft)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("allLeft=%t, compressed=%t: incorrect key map; error: %v", allLeft, compressed, err)
			}
		}
	}
}

// TestAuxiliaryFiles preserves successful command results with plain and compressed auxiliary files.
func TestAuxiliaryFiles(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "key1\nkey1\nkey999\n")
	for _, compressed := range []bool{false, true} {
		kvText := "\r\nkey1\tvalue1\r\nkey999\tvalue999"
		patternText := "key999\r\n\r\nKEY1"
		kvFile := writeAccuracyInput(t, "keys.tsv", kvText)
		patternFile := writeAccuracyInput(t, "patterns.txt", patternText)
		if compressed {
			kvFile = writeAuxiliaryGzip(t, kvText, "")
			patternFile = writeAuxiliaryGzip(t, patternText, "")
		}
		cases := []struct {
			args []string
			want string
		}{
			{[]string{"replace", "-f", "1", "-p", "^(.+)$", "-r", "{kv}", "-k", kvFile, input}, "key1\nvalue1\nvalue999\n"},
			{[]string{"rename2", "-f", "1", "-p", "^(.+)$", "-r", "{kv}", "-k", kvFile, input}, "value1\nkey1\nkey999\n"},
			{[]string{"grep", "-f", "1", "-i", "-P", patternFile, input}, "key1\nkey1\nkey999\n"},
			{[]string{"grep", "-f", "1", "-r", "-i", "-P", patternFile, input}, "key1\nkey1\nkey999\n"},
			{[]string{"sort", "-k", "1:u", "-i", "-L", "1:" + patternFile, input}, "key1\nkey999\nkey1\n"},
		}
		for _, tc := range cases {
			t.Run(fmt.Sprintf("%s/compressed=%t/args=%v", tc.args[0], compressed, tc.args[1:]), func(t *testing.T) {
				out, _ := runAccuracyCLI(t, tc.args...)
				if out != tc.want {
					t.Fatalf("got %q, want %q", out, tc.want)
				}
			})
		}
	}
}
