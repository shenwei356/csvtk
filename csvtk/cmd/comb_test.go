package cmd

import (
	"strings"
	"testing"
)

// TestCombCSVRecords covers logical records, variable widths, and shared CSV options.
func TestCombCSVRecords(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
		args              []string
	}{
		{"multiline", "\"a\nb\",c\n", "\"a\nb\",c\n", []string{"-H"}},
		{"comments and header", "# comment\nh1,h2\n# comment\na,b\n", "a,b\n", nil},
		{"long field", strings.Repeat("a", 70000) + ",b\n", strings.Repeat("a", 70000) + ",b\n", []string{"-H"}},
		{"variable widths", "a,b\nc,d,e\n", "a,b\nc,d\nc,e\nd,e\n", []string{"-H"}},
		{"spaces", " a,b \n", "\" a\",b \n", []string{"-H"}},
		{"lazy quotes", "a\"b,c\n", "\"a\"\"b\",c\n", []string{"-H", "-l"}},
		{"skip illegal", "a\"b,c\nx,y\n", "x,y\n", []string{"-H", "-I"}},
		{"ignore case", "A,a,B,C\n", "a,b\na,c\nb,c\n", []string{"-H", "-i"}},
		{"natural sort", "x10,x2,x1\n", "x2,x10\nx1,x10\nx1,x2\n", []string{"-H", "-S"}},
		{"all items", "c,a,b\nx,y\n", "a,b,c\nx,y\n", []string{"-H", "-n", "0", "-s"}},
		{"quoted tabs", "\"a\tb\"\tc\n", "\"a\tb\"\tc\n", []string{"-H", "-t"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := writeAccuracyInput(t, "items.csv", tc.input)
			args := append([]string{"comb"}, tc.args...)
			out, _ := runAccuracyCLI(t, append(args, input)...)
			if out != tc.want {
				t.Fatalf("unexpected output: got %q, want %q", out, tc.want)
			}
		})
	}
}
