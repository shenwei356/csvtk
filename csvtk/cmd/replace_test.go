package cmd

import (
	"fmt"
	"strings"
	"testing"
)

// TestReplaceExpandedGroups checks that group selectors never become replacement fields.
func TestReplaceExpandedGroups(t *testing.T) {
	input := writeAccuracyInput(t, "groups.csv", "key1,key2,value\nx,z,a\ny,z,b\nx,z,c\nx,w,d\nx,z,e\n")
	for _, selection := range []struct {
		fields, groups string
		fuzzy          bool
	}{
		{"3", "1-2", false}, {"value", "key1,key2", false}, {"value", "key*", true},
	} {
		for _, numbering := range []struct {
			symbol string
			values []string
		}{
			{"gnr", []string{"1", "1", "2", "1", "3"}},
			{"enr", []string{"1", "2", "1", "3", "1"}},
			{"rnr", []string{"1", "2", "3", "4", "5"}},
		} {
			t.Run(selection.groups+numbering.symbol, func(t *testing.T) {
				args := []string{"replace", "-f", selection.fields, "-g", selection.groups, "-p", ".+", "-r", "{" + numbering.symbol + "}"}
				if selection.fuzzy {
					args = append(args, "-F")
				}
				out, _ := runAccuracyCLI(t, append(args, input)...)
				rows := parseAccuracyCSV(t, out)
				originals := parseAccuracyCSV(t, "key1,key2,value\nx,z,a\ny,z,b\nx,z,c\nx,w,d\nx,z,e\n")
				for i, row := range rows[1:] {
					if row[0] != originals[i+1][0] || row[1] != originals[i+1][1] || row[2] != numbering.values[i] {
						t.Fatalf("wrong group or modified key: %q", rows)
					}
				}
			})
		}
	}
	// Open group ranges and overlapping replacement/group columns retain their order.
	input = writeAccuracyInput(t, "open.csv", "value,key1,key2\na,x,z\nb,y,z\nc,x,z\n")
	out, _ := runAccuracyCLI(t, "replace", "-f", "1", "-g", "2-", "-p", ".+", "-r", "{gnr}", input)
	if out != "value,key1,key2\n1,x,z\n1,y,z\n2,x,z\n" {
		t.Fatal(out)
	}
	input = writeAccuracyInput(t, "headerless.csv", "x,z,a\ny,z,b\nx,z,c\n")
	out, _ = runAccuracyCLI(t, "replace", "-H", "-f", "3", "-g", "1-2", "-p", ".+", "-r", "{gnr}", input)
	if out != "x,z,1\ny,z,1\nx,z,2\n" {
		t.Fatal(out)
	}
}

// TestReplaceLiteralKV covers literal map values, missing keys and active template captures.
func TestReplaceLiteralKV(t *testing.T) { testLiteralKV(t, "replace") }

// testLiteralKV runs the same literal-value contract against data and header replacement.
func testLiteralKV(t *testing.T, command string) {
	kv := writeAccuracyInput(t, "kv.tsv", "a\tcost$5\nb\t$1\nc\t${name}$$\\path\n")
	for _, tc := range []struct {
		key, want string
		extra     []string
	}{
		{"a", "cost$5", nil}, {"b", "$1", nil}, {"c", "${name}$$\\path", nil},
		{"$missing", "$missing", []string{"-K"}},
		{"missing", "$1${name}$$", []string{"--key-miss-repl", "$1${name}$$"}},
	} {
		for _, template := range []string{"{kv}", "$1:{kv}:{kv}:$$"} {
			t.Run(fmt.Sprintf("%s/%s", tc.key, template), func(t *testing.T) {
				input := writeAccuracyInput(t, "input.csv", "key\n"+tc.key+"\n")
				if command == "rename2" {
					input = writeAccuracyInput(t, "input.csv", tc.key+"\nvalue\n")
				}
				args := []string{command, "-f", "1", "-p", "^(.+)$", "-r", template, "-k", kv}
				args = append(args, tc.extra...)
				args = append(args, input)
				out, _ := runAccuracyCLI(t, args...)
				rows := parseAccuracyCSV(t, out)
				i := 1
				if command == "rename2" {
					i = 0
				}
				want := tc.want
				if strings.HasPrefix(template, "$1:") {
					want = tc.key + ":" + want + ":" + want + ":$"
				}
				if rows[i][0] != want {
					t.Fatalf("got %q; want %q", rows[i][0], want)
				}
			})
		}
	}
}
