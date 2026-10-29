package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestCSV2XLSXUniqueSheets keeps each input in a distinct sheet, including name normalization collisions.
func TestCSV2XLSXUniqueSheets(t *testing.T) {
	for _, names := range [][]string{
		{"sample.csv", "sample.csv", "SAMPLE.csv"},
		{"a[b].csv", "a_b_.csv", strings.Repeat("x", 40) + ".csv"},
	} {
		t.Run(names[0], func(t *testing.T) {
			args := []string{"csv2xlsx"}
			want := make([][][]string, len(names))
			for i, name := range names {
				value := strconv.Itoa(i)
				contents := "id,value\n" + value + ",first\n"
				want[i] = [][]string{{"id", "value"}, {value, "first"}}
				if i == 0 {
					contents += "extra,last\n"
					want[i] = append(want[i], []string{"extra", "last"})
				}
				args = append(args, writeAccuracyInput(t, name, contents))
			}
			output := filepath.Join(t.TempDir(), "combined.xlsx")
			runAccuracyCLI(t, append(args, "-o", output)...)
			xlsx, err := excelize.OpenFile(output)
			if err != nil {
				t.Fatal(err)
			}
			defer xlsx.Close()
			sheets := xlsx.GetSheetList()
			if len(sheets) != len(want) {
				t.Fatalf("got sheets %q, want %d", sheets, len(want))
			}
			for i, sheet := range sheets {
				rows, err := xlsx.GetRows(sheet)
				if err != nil || !reflect.DeepEqual(rows, want[i]) {
					t.Fatalf("sheet %q: got %q, want %q; error: %v", sheet, rows, want[i], err)
				}
			}
		})
	}
}

// TestCSV2XLSXDefaultOutputAliases checks the resolved filename before SaveAs can truncate an input.
func TestCSV2XLSXDefaultOutputAliases(t *testing.T) {
	for _, alias := range []string{"hardlink", "symlink"} {
		t.Run(alias, func(t *testing.T) {
			const contents = "id,value\na,1\n"
			input := writeAccuracyInput(t, "input.csv", contents)
			link := os.Link
			if alias == "symlink" {
				link = os.Symlink
			}
			if err := link(input, input+".xlsx"); err != nil {
				t.Fatal(err)
			}
			assertInputError(t, "same file", "csv2xlsx", input)
			got, err := os.ReadFile(input)
			if err != nil || string(got) != contents {
				t.Fatalf("input changed: %q; error: %v", got, err)
			}
		})
	}
}

// TestCSV2XLSXDefaultOutputFileListAlias protects the file list after resolving the default output.
func TestCSV2XLSXDefaultOutputFileListAlias(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "id,value\na,1\n")
	contents := input + "\n"
	list := writeAccuracyInput(t, "inputs.txt", contents)
	if err := os.Link(list, input+".xlsx"); err != nil {
		t.Fatal(err)
	}
	assertInputError(t, "same file", "csv2xlsx", "-X", list)
	got, err := os.ReadFile(list)
	if err != nil || string(got) != contents {
		t.Fatalf("file list changed: %q; error: %v", got, err)
	}
}

// TestCSV2XLSXWideTables checks all cells across two- and three-letter column
// names, including the XLSX column limit, in text and numeric output modes.
func TestCSV2XLSXWideTables(t *testing.T) {
	for _, width := range []int{703, excelize.MaxColumns} {
		for _, numbers := range []bool{false, true} {
			t.Run(strconv.Itoa(width)+"/numbers="+strconv.FormatBool(numbers), func(t *testing.T) {
				header := make([]string, width)
				values := make([]string, width)
				for i := range header {
					header[i] = "c" + strconv.Itoa(i+1)
					values[i] = strconv.Itoa(i + 1)
				}
				input := writeAccuracyInput(t, "wide.csv", strings.Join(header, ",")+"\n"+strings.Join(values, ",")+"\n")
				output := filepath.Join(t.TempDir(), "wide.xlsx")
				args := []string{"csv2xlsx", input, "-o", output}
				if numbers {
					args = append(args, "--format-numbers")
				}
				runAccuracyCLI(t, args...)
				xlsx, err := excelize.OpenFile(output)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := xlsx.Close(); err != nil {
						t.Error(err)
					}
				}()
				rows, err := xlsx.GetRows("Sheet1", excelize.Options{RawCellValue: true})
				if err != nil {
					t.Fatal(err)
				}
				if len(rows) != 2 {
					t.Fatalf("got %d rows, want 2", len(rows))
				}
				for i, want := range [][]string{header, values} {
					if !reflect.DeepEqual(rows[i], want) {
						t.Fatalf("row %d changed during XLSX conversion (%d columns)", i+1, width)
					}
				}
			})
		}
	}
}

// TestCSV2XLSXRejectsTooManyColumns checks that an oversized output fails
// explicitly, including when the extra column comes from row numbering.
func TestCSV2XLSXRejectsTooManyColumns(t *testing.T) {
	for _, rowNumber := range []bool{false, true} {
		for _, numbers := range []bool{false, true} {
			t.Run("row-number="+strconv.FormatBool(rowNumber)+"/numbers="+strconv.FormatBool(numbers), func(t *testing.T) {
				width := excelize.MaxColumns + 1
				if rowNumber {
					width--
				}
				input := writeAccuracyInput(t, "wide.csv", strings.Repeat("1,", width-1)+"1\n")
				output := filepath.Join(t.TempDir(), "wide.xlsx")
				args := []string{"csv2xlsx", "-H", input, "-o", output}
				if rowNumber {
					args = append(args, "--show-row-number")
				}
				if numbers {
					args = append(args, "--format-numbers")
				}
				cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
				cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
				var stderr bytes.Buffer
				cmd.Stderr = &stderr
				if err := cmd.Run(); err == nil {
					t.Fatal("conversion succeeded despite exceeding the XLSX column limit")
				} else if _, ok := err.(*exec.ExitError); !ok {
					t.Fatalf("could not run csvtk: %v", err)
				}
				if !strings.Contains(stderr.String(), excelize.ErrColumnNumber.Error()) {
					t.Fatalf("missing column limit error: %s", stderr.String())
				}
				if _, err := os.Stat(output); !os.IsNotExist(err) {
					t.Fatalf("unexpected output file after failed conversion: %v", err)
				}
			})
		}
	}
}
