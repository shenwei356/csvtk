package cmd

import (
	"reflect"
	"testing"
)

// TestFieldSelection checks range ordering and complements against physical columns.
func TestFieldSelection(t *testing.T) {
	header := []string{"id", "key1", "key2", "value"}
	for _, tc := range []struct {
		spec  string
		fuzzy bool
		want  []int
	}{
		{"1,3-", false, []int{1, 3, 4}}, {"3-,1-2", false, []int{3, 4, 1, 2}},
		{"-2-", false, []int{1}}, {"-1,-3", false, []int{2, 4}}, {"-9", false, []int{1, 2, 3, 4}},
		{"key2,id", false, []int{3, 1}}, {"key*", true, []int{2, 3}}, {"-key*", true, []int{1, 4}},
		{"-id,-value", false, []int{2, 3}}, {"key*,key2", true, []int{2, 3, 3}},
	} {
		got := newFieldSelection(tc.spec).resolve(header, tc.fuzzy)
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%q: got %v; want %v", tc.spec, got, tc.want)
		}
	}
}
