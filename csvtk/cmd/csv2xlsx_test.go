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
