package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWatchImageAliases rejects image aliases before truncating either destination.
func TestWatchImageAliases(t *testing.T) {
	const contents = "v\n1\n2\n"
	for _, alias := range []string{"same", "symlink", "hardlink", "list", "later"} {
		t.Run(alias, func(t *testing.T) {
			input := writeAccuracyInput(t, "input.csv", contents)
			image := input
			args := []string{"watch", "-f", "v", "--quiet", input}
			if alias == "symlink" || alias == "hardlink" {
				image = filepath.Join(t.TempDir(), "image.png")
				link := os.Symlink
				if alias == "hardlink" {
					link = os.Link
				}
				if err := link(input, image); err != nil {
					t.Fatal(err)
				}
			} else if alias == "list" {
				image = writeAccuracyInput(t, "list.txt", input+"\n")
				args = []string{"watch", "-f", "v", "--quiet", "-X", image}
			} else if alias == "later" {
				first := writeAccuracyInput(t, "first.csv", contents)
				args = []string{"watch", "-f", "v", "--quiet", first, input}

			}
			before, err := os.ReadFile(image)
			if err != nil {
				t.Fatal(err)
			}
			output := writeAccuracyInput(t, "output.csv", "sentinel")
			assertInputError(t, "same file", append(args, "-O", image, "-o", output)...)
			after, err := os.ReadFile(image)
			if err != nil || string(after) != string(before) {
				t.Fatalf("protected image target changed: %q, %v", after, err)
			}
			after, err = os.ReadFile(output)
			if err != nil || string(after) != "sentinel" {
				t.Fatalf("output opened too early: %q, %v", after, err)
			}
		})
	}
}

// TestWatchOutputAliases covers existing hard links and new targets through directory links.
func TestWatchOutputAliases(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "v\n1\n2\n")
	for _, alias := range []string{"same new", "same existing", "hardlink", "dangling", "parent"} {
		t.Run(alias, func(t *testing.T) {
			dir := t.TempDir()
			output := filepath.Join(dir, "output.png")
			image := output
			if alias == "same existing" || alias == "hardlink" {
				if err := os.WriteFile(output, []byte("sentinel"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			switch alias {
			case "hardlink":
				image = filepath.Join(dir, "alias.png")
				if err := os.Link(output, image); err != nil {
					t.Fatal(err)
				}
			case "dangling":
				image = filepath.Join(dir, "alias.png")
				if err := os.Symlink(output, image); err != nil {
					t.Fatal(err)
				}
			case "parent":
				link := filepath.Join(dir, "alias")
				if err := os.Symlink(dir, link); err != nil {
					t.Fatal(err)
				}
				output = filepath.Join(dir, "new", "output.png")
				image = filepath.Join(link, "new", "output.png")
			}
			assertInputError(t, "same file", "watch", "--quiet", "-f", "v", "-o", output, "-O", image, input)
			data, err := os.ReadFile(output)
			if alias == "same existing" || alias == "hardlink" {
				if err != nil || string(data) != "sentinel" {
					t.Fatalf("output changed: %q, %v", data, err)
				}
			} else if !os.IsNotExist(err) {
				t.Fatalf("output created: %v", err)
			}
		})
	}
}

// TestWatchEmptyHistogram handles inputs without finite data and rejects zero frequency early.
func TestWatchEmptyHistogram(t *testing.T) {
	for _, contents := range []string{"", "v\n", "v\nNA\nNaN\nInf\n"} {
		input := writeAccuracyInput(t, "input.csv", contents)
		image := filepath.Join(t.TempDir(), "image.png")
		runAccuracyCLI(t, "watch", "--quiet", "-f", "v", "-O", image, input)
		if _, err := os.Stat(image); !os.IsNotExist(err) {
			t.Fatalf("empty image created: %v", err)
		}
	}
	input := writeAccuracyInput(t, "input.csv", "v\n1\n2\n")
	for _, freq := range []string{"0", "-2"} {
		output := writeAccuracyInput(t, "output.csv", "sentinel")
		assertInputError(t, "--print-freq", "watch", "-f", "v", "-p", freq, "-o", output, input)
		data, _ := os.ReadFile(output)
		if string(data) != "sentinel" {
			t.Fatal("invalid frequency truncated output")
		}
	}
	image := filepath.Join(t.TempDir(), "image.png")
	runAccuracyCLI(t, "watch", "--quiet", "-f", "v", "-O", image, input)
	data, err := os.ReadFile(image)
	if err != nil || len(data) < 8 || string(data[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Fatalf("valid histogram not written: %v", err)
	}
}

// TestWatchPassthrough retains non-numeric records and honors header suppression.
func TestWatchPassthrough(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "value\n1\ntext\nNaN\nInf\n-2\n")
	for _, logScale := range []bool{false, true} {
		args := []string{"watch", "-f", "1", "--pass", "-Q", "-U"}
		if logScale {
			args = append(args, "--log")
		}
		out, _ := runAccuracyCLI(t, append(args, input)...)
		if out != "1\ntext\nNaN\nInf\n-2\n" {
			t.Fatalf("lost passthrough data: %q", out)
		}
	}
}

// TestWatchRequiresSelectedField rejects an empty complement instead of dropping the input.
func TestWatchRequiresSelectedField(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "v\n1\n")
	assertInputError(t, "expected 1 selected field", "watch", "-f", "-1", "--pass", "-Q", input)
}
