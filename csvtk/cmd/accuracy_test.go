package cmd

import (
	"bytes"
	"encoding/csv"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestCLIAccuracyHelper(t *testing.T) {
	if os.Getenv("CSVTK_ACCURACY_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			RootCmd.SetArgs(os.Args[i+1:])
			if err := RootCmd.Execute(); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
	os.Exit(2)
}

func runAccuracyCLI(t *testing.T, args ...string) (string, string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("csvtk %v: %v\n%s", args, err, stderr.String())
	}
	return stdout.String(), stderr.String()
}

func writeAccuracyInput(t *testing.T, name, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func parseAccuracyCSV(t *testing.T, output string) [][]string {
	t.Helper()
	rows, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	if err != nil {
		t.Fatalf("invalid CSV %q: %v", output, err)
	}
	return rows
}

func TestCompositeFieldKeys(t *testing.T) {
	fields := []string{"x_shenwei356_y", "z", "", "K"}
	if got := decodeFields(encodeFields(fields, false)); !reflect.DeepEqual(got, fields) {
		t.Fatalf("round trip: %q", got)
	}
	if encodeFields([]string{"x_shenwei356_y", "z"}, false) == encodeFields([]string{"x", "y_shenwei356_z"}, false) {
		t.Fatal("distinct tuples have the same key")
	}
	if got := decodeFields(encodeFields([]string{"K"}, true)); !reflect.DeepEqual(got, []string{"k"}) {
		t.Fatalf("case folding changed field length incorrectly: %q", got)
	}

	input := writeAccuracyInput(t, "input.csv", "a,b,v\nx_shenwei356_y,z,1\nx,y_shenwei356_z,2\n")
	right := writeAccuracyInput(t, "right.csv", "a,b,w\nx,y_shenwei356_z,R\n")

	out, _ := runAccuracyCLI(t, "freq", "-f", "a,b", input)
	rows := parseAccuracyCSV(t, out)
	if len(rows) != 3 || !reflect.DeepEqual(rows[1], []string{"x_shenwei356_y", "z", "1"}) || !reflect.DeepEqual(rows[2], []string{"x", "y_shenwei356_z", "1"}) {
		t.Fatalf("freq merged groups: %q", rows)
	}
	for _, option := range []string{"-k", "-n"} {
		out, _ = runAccuracyCLI(t, "freq", option, "-f", "a,b", input)
		rows = parseAccuracyCSV(t, out)
		if len(rows) != 3 || !reflect.DeepEqual(rows[1], []string{"x", "y_shenwei356_z", "1"}) || !reflect.DeepEqual(rows[2], []string{"x_shenwei356_y", "z", "1"}) {
			t.Fatalf("freq %s did not sort distinct keys: %q", option, rows)
		}
	}

	out, _ = runAccuracyCLI(t, "join", "-f", "a,b", input, right)
	rows = parseAccuracyCSV(t, out)
	if len(rows) != 2 || !reflect.DeepEqual(rows[1], []string{"x", "y_shenwei356_z", "2", "R"}) {
		t.Fatalf("join matched a different tuple: %q", rows)
	}
	outerRight := writeAccuracyInput(t, "outer-right.csv", "a,b,w\nx,y_shenwei356_z,R\np_shenwei356_q,r,S\n")
	out, _ = runAccuracyCLI(t, "join", "-O", "-f", "a,b", input, outerRight)
	rows = parseAccuracyCSV(t, out)
	foundOuterKey := false
	for _, row := range rows[1:] {
		if row[0] == "p_shenwei356_q" && row[1] == "r" && row[3] == "S" {
			foundOuterKey = true
		}
	}
	if len(rows) != 4 || !foundOuterKey {
		t.Fatalf("outer join reconstructed a key incorrectly: %q", rows)
	}

	out, _ = runAccuracyCLI(t, "summary", "-f", "v:sum", "-g", "a,b", input)
	rows = parseAccuracyCSV(t, out)
	if len(rows) != 3 || len(rows[1]) != 3 || len(rows[2]) != 3 || rows[1][2] == rows[2][2] {
		t.Fatalf("summary merged groups: %q", rows)
	}

	out, _ = runAccuracyCLI(t, "uniq", "-f", "a,b", input)
	if rows = parseAccuracyCSV(t, out); len(rows) != 3 {
		t.Fatalf("uniq merged keys: %q", rows)
	}

	out, _ = runAccuracyCLI(t, "inter", "-f", "a,b", input, right)
	rows = parseAccuracyCSV(t, out)
	if len(rows) != 2 || !reflect.DeepEqual(rows[1], []string{"x", "y_shenwei356_z"}) {
		t.Fatalf("inter matched a different tuple: %q", rows)
	}

	out, _ = runAccuracyCLI(t, "fold", "-f", "a,b", "-v", "v", input)
	if rows = parseAccuracyCSV(t, out); len(rows) != 3 {
		t.Fatalf("fold merged groups: %q", rows)
	}

	spreadInput := writeAccuracyInput(t, "spread.csv", "a,b,k,v\nx_shenwei356_y,z,K,1\nx,y_shenwei356_z,K,2\n")
	out, _ = runAccuracyCLI(t, "spread", "-k", "k", "-v", "v", spreadInput)
	if rows = parseAccuracyCSV(t, out); len(rows) != 3 || len(rows[1]) != 3 || len(rows[2]) != 3 {
		t.Fatalf("spread merged groups: %q", rows)
	}

	out, _ = runAccuracyCLI(t, "replace", "-f", "v", "-g", "a,b", "-p", ".", "-r", "{enr}", input)
	rows = parseAccuracyCSV(t, out)
	if len(rows) != 3 || rows[1][2] != "1" || rows[2][2] != "2" {
		t.Fatalf("replace merged groups: %q", rows)
	}

	caseInput := writeAccuracyInput(t, "case.csv", "a,b\nA,x\na,x\n")
	out, _ = runAccuracyCLI(t, "freq", "-i", "-f", "a,b", caseInput)
	if rows = parseAccuracyCSV(t, out); len(rows) != 2 || rows[1][2] != "2" {
		t.Fatalf("freq -i did not merge case variants: %q", rows)
	}
}

func TestSplitKeysStayDistinctAndWithinOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.csv")
	if err := os.WriteFile(input, []byte("a,b,v\na-b,c,first\na,b-c,second\n../outside,x,third\n"), 0600); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(dir, "out")
	runAccuracyCLI(t, "split", "-f", "a,b", "-p", "", "-o", outDir, input)
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected three distinct files, got %d", len(entries))
	}
	for _, name := range []string{"a%2Db-c.csv", "a-b%2Dc.csv", "%2E%2E%2Foutside-x.csv"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Errorf("missing output %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "outside-x.csv")); !os.IsNotExist(err) {
		t.Fatalf("split wrote outside its output directory: %v", err)
	}

	defaultDir := filepath.Join(dir, "default")
	runAccuracyCLI(t, "split", "-f", "a,b", "-o", defaultDir, input)
	if _, err := os.Stat(filepath.Join(defaultDir, "input-a%2Db-c.csv")); err != nil {
		t.Fatalf("default prefix should use the input basename: %v", err)
	}
	caseInput := writeAccuracyInput(t, "case.csv", "k\nA\na\n")
	caseDir := filepath.Join(dir, "case")
	runAccuracyCLI(t, "split", "-f", "k", "-p", "", "-o", caseDir, caseInput)
	if _, err := os.Stat(filepath.Join(caseDir, "A.csv")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(caseDir, "a~2.csv")); err != nil {
		t.Fatal(err)
	}
	longInput := writeAccuracyInput(t, "long.csv", "k\n"+strings.Repeat("/", 100)+"\n")
	longDir := filepath.Join(dir, "long")
	runAccuracyCLI(t, "split", "-p", "", "-o", longDir, longInput)
	longNames, err := os.ReadDir(longDir)
	if err != nil || len(longNames) != 1 || !strings.HasPrefix(longNames[0].Name(), "~") {
		t.Fatalf("long encoded key should use a bounded file name: %v, %v", longNames, err)
	}
}

func TestSplitXLSXDistinctSheetNames(t *testing.T) {
	used := map[string]struct{}{"k": {}}
	if got := uniqueSheetName("K", used); got != "K-2" {
		t.Fatalf("Excel sheet names must use Unicode case folding: %q", got)
	}
	if got := uniqueSheetName("a/b", used); got != "a_b" {
		t.Fatalf("invalid sheet character was not replaced: %q", got)
	}
	if got := uniqueSheetName("a:b", used); got != "a_b-2" {
		t.Fatalf("sanitized names collided: %q", got)
	}
	if got := uniqueSheetName(strings.Repeat("a", 30)+"'tail", used); got != strings.Repeat("a", 30) {
		t.Fatalf("truncated sheet name ends with a quote: %q", got)
	}

	dir := t.TempDir()
	input := filepath.Join(dir, "input.xlsx")
	output := filepath.Join(dir, "output.xlsx")
	f := excelize.NewFile()
	for i, row := range [][]interface{}{{"a", "b", "v"}, {"a-b", "c", "first"}, {"a", "b-c", "second"}} {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.SetSheetRow("Sheet1", cell, &row); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.SaveAs(input); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	runAccuracyCLI(t, "splitxlsx", "-f", "a,b", "-o", output, input)
	result, err := excelize.OpenFile(output)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Close()
	names := result.GetSheetList()
	sort.Strings(names)
	if !reflect.DeepEqual(names, []string{"Sheet1", "a-b-c", "a-b-c-2"}) {
		t.Fatalf("colliding sheet names: %q", names)
	}
	for sheet, want := range map[string]string{"a-b-c": "first", "a-b-c-2": "second"} {
		got, err := result.GetCellValue(sheet, "C2")
		if err != nil || got != want {
			t.Errorf("%s: got %q, want %q: %v", sheet, got, want, err)
		}
	}
}

func TestSummaryIndexAndMatrixFilters(t *testing.T) {
	input := writeAccuracyInput(t, "summary.csv", "v\n3\n1\n2\n")
	out, _ := runAccuracyCLI(t, "summary", "-f", "v:argmin", "-f", "v:median", input)
	rows := parseAccuracyCSV(t, out)
	if len(rows) != 2 || !reflect.DeepEqual(rows[1], []string{"2.00", "2.00"}) {
		t.Fatalf("argmin used sorted positions: %q", rows)
	}

	left := writeAccuracyInput(t, "left.csv", "a,b,v\nA,A,bad\nA,B,5\nB,A,6\nB,B,7\n")
	out, _ = runAccuracyCLI(t, "long2matrix", "-m", "2", "--keep-non-numeric", left)
	rows = parseAccuracyCSV(t, out)
	if len(rows) != 3 || rows[1][1] != "bad" {
		t.Fatalf("long2matrix dropped nonnumeric data: %q", rows)
	}

	matrix := writeAccuracyInput(t, "matrix.csv", ",A,B\nA,bad,5\nB,6,7\n")
	out, _ = runAccuracyCLI(t, "matrix2long", "-m", "2", "-N", matrix)
	rows = parseAccuracyCSV(t, out)
	if len(rows) != 5 || !reflect.DeepEqual(rows[1], []string{"A", "A", "bad"}) {
		t.Fatalf("matrix2long dropped nonnumeric data: %q", rows)
	}

	duplicates := writeAccuracyInput(t, "duplicates.csv", "a,b,v\nA,B,first\nA,B,last\n")
	out, stderr := runAccuracyCLI(t, "long2matrix", duplicates)
	rows = parseAccuracyCSV(t, out)
	if len(rows) != 3 || rows[1][2] != "first" || !strings.Contains(stderr, "skip duplicated records") {
		t.Fatalf("duplicate warning disagrees with output: %q, %q", rows, stderr)
	}
}

func TestSummaryColumnNames(t *testing.T) {
	input := writeAccuracyInput(t, "summary-names.csv", "group,a,b\nx,1,2\nx,3,4\n")
	out, _ := runAccuracyCLI(t, "summary", "-g", "1", "-f", "2-3:sum", "-n", "total_a,total_b", input)
	rows := parseAccuracyCSV(t, out)
	want := [][]string{{"group", "total_a", "total_b"}, {"x", "4.00", "6.00"}}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("renamed summary columns: got %q, want %q", rows, want)
	}

	out, _ = runAccuracyCLI(t, "summary", "-g", "1", "-f", "2-3:sum", input)
	rows = parseAccuracyCSV(t, out)
	if !reflect.DeepEqual(rows[0], []string{"group", "a:sum", "b:sum"}) {
		t.Fatalf("default summary columns changed: %q", rows[0])
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIAccuracyHelper$", "--", "summary", "-g", "1", "-f", "2-3:sum", "-n", "only-one", input)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "number of names (1) should be equal to number of summary columns (2)") {
		t.Fatalf("expected a summary-name count error, got %v: %s", err, output)
	}
}

func TestJoinIgnoreNullInCompositeKey(t *testing.T) {
	left := writeAccuracyInput(t, "left.csv", "a,b,l\n,x,L\n")
	right := writeAccuracyInput(t, "right.csv", "a,b,r\n,x,R\n")
	out, _ := runAccuracyCLI(t, "join", "-n", "-f", "a,b", left, right)
	if rows := parseAccuracyCSV(t, out); len(rows) != 1 {
		t.Fatalf("join -n matched an empty key field: %q", rows)
	}
	out, _ = runAccuracyCLI(t, "join", "-L", "-n", "-f", "a,b", "--na", "NA", left, right)
	rows := parseAccuracyCSV(t, out)
	if len(rows) != 2 || !reflect.DeepEqual(rows[1], []string{"", "x", "L", "NA"}) {
		t.Fatalf("left join lost its unmatched empty-key row: %q", rows)
	}
	out, _ = runAccuracyCLI(t, "join", "-O", "-n", "-f", "a,b", "--na", "NA", left, right)
	rows = parseAccuracyCSV(t, out)
	if len(rows) != 3 || !reflect.DeepEqual(rows[1], []string{"", "x", "L", "NA"}) || !reflect.DeepEqual(rows[2], []string{"", "x", "NA", "R"}) {
		t.Fatalf("outer join lost an unmatched empty-key row: %q", rows)
	}
}
