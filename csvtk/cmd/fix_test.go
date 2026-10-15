package cmd

import (
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// TestFixMissingValues checks padding before and after the buffer limit, and at EOF.
func TestFixMissingValues(t *testing.T) {
	input := writeAccuracyInput(t, "unequal.csv", "a,b,c,d\n1\n2,x\n3,y,,\n")
	for _, fill := range []struct {
		name  string
		value string
	}{
		{"default", ""},
		{"NA", "NA"},
		{"quoted", "N,\"A\"\nend"},
	} {
		for _, limit := range []int{0, 1, 2, 4, 8} {
			t.Run(fill.name+"/n="+strconv.Itoa(limit), func(t *testing.T) {
				args := []string{"fix", "-n", strconv.Itoa(limit)}
				if fill.value != "" {
					args = append(args, "--na", fill.value)
				}
				out, _ := runAccuracyCLI(t, append(args, input)...)
				want := [][]string{
					{"a", "b", "c", "d"},
					{"1", fill.value, fill.value, fill.value},
					{"2", "x", fill.value, fill.value},
					{"3", "y", "", ""},
				}
				if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
					t.Fatalf("padded rows: got %q, want %q", got, want)
				}
			})
		}
	}
}

// TestFixShortFirstRow checks that width discovery also pads a short header or first data row.
func TestFixShortFirstRow(t *testing.T) {
	input := writeAccuracyInput(t, "short-first.csv", "a\n1,2,3\n4\n")
	for _, mode := range []string{"header", "no_header"} {
		t.Run(mode, func(t *testing.T) {
			args := []string{"fix", "-n", "2", "--na", "NA"}
			if mode == "no_header" {
				args = append(args, "-H")
			}
			out, _ := runAccuracyCLI(t, append(args, input)...)
			want := [][]string{{"a", "NA", "NA"}, {"1", "2", "3"}, {"4", "NA", "NA"}}
			if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
				t.Fatalf("padded first row: got %q, want %q", got, want)
			}
		})
	}
}

// TestFixRejectsWiderRemainingRows preserves the width limit after buffering.
func TestFixRejectsWiderRemainingRows(t *testing.T) {
	input := writeAccuracyInput(t, "wider.csv", "a,b\n1\n2,x,wide\n")
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "fix", "-n", "2", "--na", "NA", input)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "line 3: the number of columns is larger than 2") {
		t.Fatalf("expected a width error, got %v: %s", err, out)
	}
}
