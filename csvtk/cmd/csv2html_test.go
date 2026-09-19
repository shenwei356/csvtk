package cmd

import (
	"strings"
	"testing"
)

func TestCSV2HTMLStandaloneAndEscaped(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "name,note\nAlice,<script>alert(1)</script>\nBob,\"a&b\nsecond line\"\n")
	output, _ := runAccuracyCLI(t, "csv2html",
		"--caption", "People <2026>",
		"--table-class", `wide" data-x="bad`,
		"--table-width", "720",
		"-w", "8", "-W", "20",
		input)

	for _, want := range []string{
		"<!doctype html>",
		`<meta name="viewport" content="width=device-width, initial-scale=1">`,
		`<title>People &lt;2026&gt;</title>`,
		`.csvtk-table-wrapper { width: 100%; overflow-x: auto; }`,
		`.csvtk-table-wrapper table { width: max-content; max-width: 720px;`,
		`max-width: 720px`,
		`min-width: 8ch`,
		`max-width: 20ch`,
		`<table class="wide&#34; data-x=&#34;bad">`,
		`<caption>People &lt;2026&gt;</caption>`,
		`<th scope="col"><span class="csvtk-cell">name</span></th>`,
		`&lt;script&gt;alert(1)&lt;/script&gt;`,
		"a&amp;b\nsecond line",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output does not contain %q\n%s", want, output)
		}
	}
	if strings.Contains(output, "<script>") || strings.Contains(output, ` data-x="bad"`) {
		t.Fatalf("user input was emitted as HTML markup:\n%s", output)
	}
	if strings.Count(output, "<thead>") != 1 || strings.Count(output, "<tbody>") != 1 {
		t.Fatalf("invalid table structure:\n%s", output)
	}
}

func TestCSV2HTMLNoHeaderAndColumnWidths(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "a,b\nc,d\n")
	output, _ := runAccuracyCLI(t, "csv2html", "-H", "--table-width", "90%", "-w", "0,5", "-W", "10,20", input)

	if strings.Contains(output, "<thead>") || strings.Contains(output, "<th ") {
		t.Fatalf("header emitted with -H:\n%s", output)
	}
	for _, want := range []string{
		"max-width: 90%",
		`th:nth-child(1) > .csvtk-cell, td:nth-child(1) > .csvtk-cell { max-width: 10ch; }`,
		`th:nth-child(2) > .csvtk-cell, td:nth-child(2) > .csvtk-cell { min-width: 5ch; max-width: 20ch; }`,
		`<td><span class="csvtk-cell">a</span></td>`,
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output does not contain %q\n%s", want, output)
		}
	}
}

func TestCSV2HTMLDeleteHeader(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "name,value\na,1\n")
	output, _ := runAccuracyCLI(t, "csv2html", "-U", input)
	if strings.Contains(output, ">name<") || strings.Contains(output, "<thead>") {
		t.Fatalf("header was not deleted:\n%s", output)
	}
	if !strings.Contains(output, `<td><span class="csvtk-cell">a</span></td>`) {
		t.Fatalf("data row missing:\n%s", output)
	}
}

func TestCSV2HTMLEmptyInput(t *testing.T) {
	input := writeAccuracyInput(t, "empty.csv", "")
	output, _ := runAccuracyCLI(t, "csv2html", input)
	if !strings.Contains(output, "<!doctype html>") ||
		!strings.Contains(output, "width: max-content; max-width: 100%;") ||
		!strings.Contains(output, "<tbody>\n      </tbody>") ||
		!strings.HasSuffix(output, "</html>\n") {
		t.Fatalf("empty input did not produce a complete HTML document:\n%s", output)
	}
}

func TestCSV2HTMLWidthValidation(t *testing.T) {
	valid := map[string]string{
		"1200":  "1200px",
		"90%":   "90%",
		"120%":  "120%",
		"70rem": "70rem",
		".5vw":  ".5vw",
	}
	for input, want := range valid {
		got, err := normalizeHTMLTableWidth(input)
		if err != nil || got != want {
			t.Errorf("normalizeHTMLTableWidth(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"", "0", "-1px", "auto", "100%;color:red", "1ex"} {
		if _, err := normalizeHTMLTableWidth(input); err == nil {
			t.Errorf("normalizeHTMLTableWidth(%q) should fail", input)
		}
	}

	if err := validateHTMLColumnWidths([]int{1, 2}, nil, 3); err == nil {
		t.Fatal("column-count mismatch should fail")
	}
	if err := validateHTMLColumnWidths([]int{-1}, nil, 3); err == nil {
		t.Fatal("negative width should fail")
	}
	if err := validateHTMLColumnWidths([]int{8}, []int{7}, 3); err == nil {
		t.Fatal("minimum width larger than maximum width should fail")
	}
}
