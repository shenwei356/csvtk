package cmd

import "testing"

// TestParseByteSize rejects values that cannot safely become buffer lengths.
func TestParseByteSize(t *testing.T) {
	for _, tc := range []struct {
		text string
		want int64
	}{
		{"1K", 1024}, {"1.5M", 1572864}, {" 2 b ", 2}, {"1e3", 1000}, {"0", 0},
	} {
		got, err := ParseByteSize(tc.text)
		if err != nil || got != tc.want {
			t.Fatalf("%q: got %d, error %v", tc.text, got, err)
		}
	}
	for _, value := range []string{"NaN", "Inf", "-Inf", "NaNG", "1e20", "9223372036854775808G", "-1", "-1K", "K"} {
		if got, err := ParseByteSize(value); err == nil {
			t.Fatalf("accepted %q as %d", value, got)
		}
	}
}
