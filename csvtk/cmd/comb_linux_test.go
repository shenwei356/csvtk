package cmd

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestCombFileLimit requires every plain or compressed input to close before the next opens.
func TestCombFileLimit(t *testing.T) {
	for _, compressed := range []bool{false, true} {
		input := writeAccuracyInput(t, "items.csv", "a,b\n")
		if compressed {
			input = writeAuxiliaryGzip(t, "a,b\n", "")
		}
		args := []string{"-test.run=^TestCSVReaderFileLimitHelper$", "--", "comb", "-H"}
		for i := 0; i < 100; i++ {
			args = append(args, input)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		command := exec.CommandContext(ctx, os.Args[0], args...)
		command.Env = append(os.Environ(), "CSVTK_READER_FILE_LIMIT_HELPER=1", "GOGC=off")
		out, err := command.CombinedOutput()
		cancel()
		if err != nil || string(out) != strings.Repeat("a,b\n", 100) {
			t.Fatalf("100 files, compressed=%t: %v: %s", compressed, err, out)
		}
	}
}
