package cmd

import (
	"reflect"
	"testing"
)

func TestSortByUnicodeLength(t *testing.T) {
	input := writeAccuracyInput(t, "length.csv", "group,text,id\nA,aa,2\nA,中,3\nA,a,1\nB,é,5\nB,中文,6\nB,b,4\n")

	out, _ := runAccuracyCLI(t, "sort", "-k", "group", "-k", "text:l", "-k", "id:n", input)
	want := [][]string{{"group", "text", "id"}, {"A", "a", "1"}, {"A", "中", "3"}, {"A", "aa", "2"}, {"B", "b", "4"}, {"B", "é", "5"}, {"B", "中文", "6"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("ascending Unicode length: got %q, want %q", got, want)
	}

	out, _ = runAccuracyCLI(t, "sort", "-k", "group", "-k", "text:lr", "-k", "id:n", input)
	want = [][]string{{"group", "text", "id"}, {"A", "aa", "2"}, {"A", "a", "1"}, {"A", "中", "3"}, {"B", "中文", "6"}, {"B", "b", "4"}, {"B", "é", "5"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("descending Unicode length: got %q, want %q", got, want)
	}
}

func TestSortByUnicodeLengthFieldRange(t *testing.T) {
	input := writeAccuracyInput(t, "range.csv", "first,second,id\n中,a,3\na,bb,2\na,é,1\n")
	out, _ := runAccuracyCLI(t, "sort", "-k", "1-2:l", "-k", "3:n", input)
	want := [][]string{{"first", "second", "id"}, {"a", "é", "1"}, {"中", "a", "3"}, {"a", "bb", "2"}}
	if got := parseAccuracyCSV(t, out); !reflect.DeepEqual(got, want) {
		t.Fatalf("Unicode length field range: got %q, want %q", got, want)
	}
}
