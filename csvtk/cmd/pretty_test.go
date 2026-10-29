package cmd

import "testing"

// TestPrettyAlignmentWithRowNumbers compares named and numeric column alignment.
func TestPrettyAlignmentWithRowNumbers(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "alpha,beta\nx,long\nlong,y\n")
	for _, flag := range []string{"-r", "-m"} {
		for _, pair := range [][2]string{{"alpha", "1"}, {"beta", "2"}, {"alpha,beta", "1-2"}, {"beta", "-1"}} {
			for _, numbered := range []bool{false, true} {
				args := []string{"pretty", "-S", "plain"}
				if numbered {
					args = append(args, "-Z")
				}
				named, _ := runAccuracyCLI(t, append(append([]string{}, args...), flag, pair[0], input)...)
				numeric, _ := runAccuracyCLI(t, append(append([]string{}, args...), flag, pair[1], input)...)
				if named != numeric {
					t.Fatalf("%v: named=%q numeric=%q", pair, named, numeric)
				}
			}
		}
	}
}

// TestPrettyAlignmentBounds requires a normal diagnostic for invalid indices.
func TestPrettyAlignmentBounds(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\nx,y\n")
	for _, flag := range []string{"-r", "-m"} {
		for _, field := range []string{"1-5", "-3", "-3-", "-4--3"} {
			assertInputError(t, "out of range", "pretty", flag, field, input)
			assertInputError(t, "out of range", "pretty", "-Z", flag, field, input)
		}
	}
}
