package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestSplitXLSXMissingCells checks actual output sheets for trailing empty key
// cells, entirely blank rows, and key buffers reused after nonempty values.
func TestSplitXLSXMissingCells(t *testing.T) {
	byGroup := map[string][]int{"A": {0, 1, 6}, "B": {3}, "NA": {2, 4, 5, 7}}
	byBatch := map[string][]int{"x": {0, 6}, "y": {3}, "z": {5}, "NA": {1, 2, 4, 7}}
	byPair := map[string][]int{"A-x": {0, 6}, "A-": {1}, "B-y": {3}, "-": {2, 4, 7}, "-z": {5}}
	shortFirst := [][]interface{}{{"r3"}, {"r1", "A", "x"}, nil, {"r2", "A"}, {"r4", "B", "y"}, {"r5", "", "z"}, {"r6", "A", "x"}, {"r7"}}
	emptyFirst := append([][]interface{}(nil), shortFirst...)
	emptyFirst[0] = nil
	sparseFirstGroups := map[string][]int{"x": {1, 6}, "y": {4}, "z": {5}, "NA": {0, 2, 3, 7}}
	cases := []struct {
		name, fields    string
		noHeader, fuzzy bool
		data            [][]interface{}
		groups          map[string][]int
	}{
		{"column name", "group", false, false, nil, byGroup},
		{"column number", "2", false, false, nil, byGroup},
		{"last column", "3", false, false, nil, byBatch},
		{"multiple names", "group,batch", false, false, nil, byPair},
		{"column range", "2-3", false, false, nil, byPair},
		{"excluded column", "-1", false, false, nil, byPair},
		{"fuzzy names", "g*,b*", false, true, nil, byPair},
		{"no header", "2", true, false, nil, byGroup},
		{"short first row", "3", true, false, shortFirst, sparseFirstGroups},
		{"empty first row", "3", true, false, emptyFirst, sparseFirstGroups},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := [][]interface{}{{"r1", "A", "x"}, {"r2", "A"}, {"r3"}, {"r4", "B", "y"}, nil, {"r5", "", "z"}, {"r6", "A", "x"}, {"r7"}}
			if tc.data != nil {
				data = tc.data
			}
			offset := 0
			rows := data
			if !tc.noHeader {
				offset = 1
				rows = append([][]interface{}{{"id", "group", "batch"}}, data...)
			}
			dir := t.TempDir()
			input, output := filepath.Join(dir, "input.xlsx"), filepath.Join(dir, "output.xlsx")
			book := excelize.NewFile()
			for i, row := range rows {
				cell, err := excelize.CoordinatesToCellName(1, i+1)
				if err != nil {
					t.Fatal(err)
				}
				if err := book.SetSheetRow("Sheet1", cell, &row); err != nil {
					t.Fatal(err)
				}
				if row == nil {
					if err := book.SetRowHeight("Sheet1", i+1, 24); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := book.SaveAs(input); err != nil {
				t.Fatal(err)
			}
			original, err := book.GetRows("Sheet1")
			if err != nil {
				t.Fatal(err)
			}
			if err := book.Close(); err != nil {
				t.Fatal(err)
			}
			// Ensure the fixture actually exposes Excelize's omitted trailing cells.
			for i, row := range data {
				if len(original[i+offset]) != len(row) {
					t.Fatalf("fixture row %d was not trimmed as expected: %q", i, original[i+offset])
				}
			}
			args := []string{"splitxlsx", "-f", tc.fields, "-o", output}
			if tc.noHeader {
				args = append(args, "-H")
			}
			if tc.fuzzy {
				args = append(args, "-F")
			}
			runAccuracyCLI(t, append(args, input)...)
			result, err := excelize.OpenFile(output)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := result.Close(); err != nil {
					t.Error(err)
				}
			}()
			wantSheets := []string{"Sheet1"}
			for sheet, indices := range tc.groups {
				wantSheets = append(wantSheets, sheet)
				var want [][]string
				if offset > 0 {
					want = append(want, original[0])
				}
				for i, index := range indices {
					want = append(want, original[index+offset])
					if data[index] == nil {
						height, err := result.GetRowHeight(sheet, i+offset+1)
						if err != nil || height != 24 {
							t.Fatalf("blank row lost its height in sheet %q: %g, %v", sheet, height, err)
						}
					}
				}
				got, err := result.GetRows(sheet)
				if err != nil || !reflect.DeepEqual(got, want) {
					t.Fatalf("sheet %q: got %q, want %q: %v", sheet, got, want, err)
				}
			}
			names := result.GetSheetList()
			sort.Strings(names)
			sort.Strings(wantSheets)
			if !reflect.DeepEqual(names, wantSheets) {
				t.Fatalf("got sheets %q, want %q", names, wantSheets)
			}
			if got, err := result.GetRows("Sheet1"); err != nil || !reflect.DeepEqual(got, original) {
				t.Fatalf("source sheet changed: %q, %v", got, err)
			}
		})
	}
}

// TestSplitXLSXMissingCellsFieldRange retains validation for columns beyond all
// data, even when the first row is shorter than the worksheet's actual width.
func TestSplitXLSXMissingCellsFieldRange(t *testing.T) {
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.xlsx"), filepath.Join(dir, "output.xlsx")
	book := excelize.NewFile()
	for cell, value := range map[string]string{"A1": "short", "A2": "wide", "C2": "value"} {
		if err := book.SetCellValue("Sheet1", cell, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := book.SaveAs(input); err != nil {
		t.Fatal(err)
	}
	if err := book.Close(); err != nil {
		t.Fatal(err)
	}
	args := []string{"splitxlsx", "-H", "-f", "4", "-o", output, input}
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err == nil {
		t.Fatal("out-of-range field did not fail")
	} else if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("could not run csvtk: %v", err)
	}
	if !strings.Contains(stderr.String(), "field (4) out of range (3)") {
		t.Fatalf("missing worksheet-width diagnostic: %s", stderr.String())
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("unexpected output after field validation failed: %v", err)
	}
}
