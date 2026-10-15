package cmd

import (
	"bytes"
	"encoding/xml"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// TestBarCategoryAlignment compares rendered charts for equivalent tables whose
// groups list categories in different orders, including gaps and skipped rows.
func TestBarCategoryAlignment(t *testing.T) {
	cases := []struct {
		name, ordered, reordered string
		skipNA                   bool
	}{
		{
			name:      "reversed group",
			ordered:   "B,2,g1\nA,1,g1\nB,20,g2\nA,10,g2\n",
			reordered: "B,2,g1\nA,1,g1\nA,10,g2\nB,20,g2\n",
		},
		{
			name:      "interleaved groups",
			ordered:   "B,2,g1\nA,1,g1\nB,20,g2\nA,10,g2\n",
			reordered: "B,2,g1\nA,10,g2\nA,1,g1\nB,20,g2\n",
		},
		{
			name:      "missing middle category",
			ordered:   "B,2,g1\nA,1,g1\nC,3,g1\nB,20,g2\nC,30,g2\n",
			reordered: "B,2,g1\nA,1,g1\nC,3,g1\nC,30,g2\nB,20,g2\n",
		},
		{
			name:      "skipped categories",
			ordered:   "B,2,g1\nA,1,g1\nB,20,g2\nA,10,g2\n",
			reordered: "NA,100,g0\nX,NA,g0\nB,2,g1\nA,1,g1\nA,10,g2\nB,20,g2\n",
			skipNA:    true,
		},
		{
			name:      "category introduced by later group",
			ordered:   "B,2,g1\nA,1,g1\nC,30,g2\nB,20,g2\nA,10,g2\n",
			reordered: "B,2,g1\nA,1,g1\nC,30,g2\nA,10,g2\nB,20,g2\n",
		},
	}
	for _, tc := range cases {
		for _, horizontal := range []bool{false, true} {
			name := tc.name + "/vertical"
			if horizontal {
				name = tc.name + "/horizontal"
			}
			t.Run(name, func(t *testing.T) {
				args := []string{"plot", "bar", "-x", "category", "-y", "value", "-g", "group", "--format", "svg"}
				if horizontal {
					args = append(args, "--horizontal")
				}
				if tc.skipNA {
					args = append(args, "--skip-na")
				}
				ordered := writeAccuracyInput(t, "ordered.csv", "category,value,group\n"+tc.ordered)
				reordered := writeAccuracyInput(t, "reordered.csv", "category,value,group\n"+tc.reordered)
				want, _ := runAccuracyCLI(t, append(args, ordered)...)
				got, _ := runAccuracyCLI(t, append(args, reordered)...)
				if got != want {
					t.Fatal("reordering categories within groups changed the rendered chart")
				}
				// Categories must occur once on the axis, in first-seen order.
				decoder := xml.NewDecoder(strings.NewReader(got))
				var labels []string
				var rectangles [][2]float64 // Bar lengths and category centers, followed by legend swatches.
				for {
					token, err := decoder.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					if start, ok := token.(xml.StartElement); ok && start.Name.Local == "text" {
						var label string
						if err := decoder.DecodeElement(&label, &start); err != nil {
							t.Fatal(err)
						}
						if label == "A" || label == "B" || label == "C" {
							labels = append(labels, label)
						}
					} else if ok && start.Name.Local == "path" {
						var path, style string
						for _, attr := range start.Attr {
							if attr.Name.Local == "d" {
								path = attr.Value
							} else if attr.Name.Local == "style" {
								style = attr.Value
							}
						}
						if strings.HasPrefix(style, "fill:#") && style != "fill:#FFFFFF" {
							rectangles = append(rectangles, barSVGRectangle(t, path, horizontal))
						}
					}
				}
				wantLabels := "B,A"
				if strings.Contains(tc.ordered, "C,") {
					wantLabels = "B,A,C"
				}
				if strings.Join(labels, ",") != wantLabels {
					t.Fatalf("category labels: got %q, want %q", labels, wantLabels)
				}
				wantValues := []float64{2, 1, 20, 10}
				wantPositions := []float64{0, 1, 0, 1}
				secondGroup := 2
				if tc.name == "missing middle category" {
					wantValues = []float64{2, 1, 3, 20, 30}
					wantPositions = []float64{0, 1, 2, 0, 2}
					secondGroup = 3
				} else if tc.name == "category introduced by later group" {
					wantValues = []float64{2, 1, 20, 10, 30}
					wantPositions = []float64{0, 1, 0, 1, 2}
				}
				if len(rectangles) != len(wantValues)+2 {
					t.Fatalf("got %d filled rectangles, want %d bars and 2 legend swatches", len(rectangles), len(wantValues))
				}
				spacing := rectangles[1][1] - rectangles[0][1]
				for i, value := range wantValues {
					if math.Abs(rectangles[i][0]-value*rectangles[1][0]) > 0.05 {
						t.Fatalf("bar %d has the wrong length for value %g", i, value)
					}
					origin := rectangles[0][1]
					if i >= secondGroup {
						origin = rectangles[secondGroup][1]
					}
					if math.Abs(rectangles[i][1]-(origin+wantPositions[i]*spacing)) > 0.05 {
						t.Fatalf("bar %d is plotted at the wrong category position", i)
					}
				}
			})
		}
	}
}

// barSVGRectangle measures an SVG bar in plot coordinates, independently of
// category/value orientation and the chart's scale.
func barSVGRectangle(t *testing.T, path string, horizontal bool) [2]float64 {
	t.Helper()
	coordinates := strings.Fields(strings.NewReplacer("M", " ", "L", " ", "Z", " ", ",", " ").Replace(path))
	if len(coordinates) != 8 {
		t.Fatalf("expected a rectangular bar, got path %q", path)
	}
	xmin, ymin, xmax, ymax := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for i := 0; i < len(coordinates); i += 2 {
		x, err := strconv.ParseFloat(coordinates[i], 64)
		if err != nil {
			t.Fatal(err)
		}
		y, err := strconv.ParseFloat(coordinates[i+1], 64)
		if err != nil {
			t.Fatal(err)
		}
		xmin, xmax = math.Min(xmin, x), math.Max(xmax, x)
		ymin, ymax = math.Min(ymin, y), math.Max(ymax, y)
	}
	if horizontal {
		return [2]float64{xmax - xmin, (ymin + ymax) / 2}
	}
	return [2]float64{ymax - ymin, (xmin + xmax) / 2}
}

// TestBarDuplicateCategories requires an explicit error for ambiguous grouped
// values, including when the first occurrence is zero.
func TestBarDuplicateCategories(t *testing.T) {
	input := writeAccuracyInput(t, "input.csv", "category,value,group\nA,0,g1\nB,1,g1\nA,2,g1\n")
	args := []string{"plot", "bar", "-x", "category", "-y", "value", "-g", "group", "--format", "svg", input}
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestCLIAccuracyHelper$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "CSVTK_ACCURACY_HELPER=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err == nil {
		t.Fatal("duplicate group/category pair did not fail")
	} else if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("could not run csvtk: %v", err)
	}
	if !strings.Contains(stderr.String(), `duplicate category "A" in group "g1"`) {
		t.Fatalf("missing duplicate diagnostic: %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatal("duplicate category error wrote a partial chart")
	}

	// Without grouping, repeated labels still represent separate input rows.
	out, _ := runAccuracyCLI(t, "plot", "bar", "-x", "category", "-y", "value", "--format", "svg", input)
	if strings.Count(out, ">A</text>") != 2 {
		t.Fatal("ungrouped chart did not preserve repeated category labels")
	}
}

// TestBarWithoutHeader checks numeric selectors against the same rendered chart
// as named selectors, with explicit width, negative values, and a missing group.
func TestBarWithoutHeader(t *testing.T) {
	const rows = "B,-2,g1\nA,1,g1\nA,0,g2\nB,-20,g2\nC,3,\n"
	withHeader := writeAccuracyInput(t, "header.csv", "category,value,group\n"+rows)
	withoutHeader := writeAccuracyInput(t, "no-header.csv", rows)
	want, _ := runAccuracyCLI(t, "plot", "bar", "-x", "category", "-y", "value", "-g", "group", "--bar-width", "10", "--format", "svg", withHeader)
	got, _ := runAccuracyCLI(t, "plot", "bar", "-H", "-x", "1", "-y", "2", "-g", "3", "--bar-width", "10", "--ylab", "value", "--format", "svg", withoutHeader)
	if got != want {
		t.Fatal("numeric selectors without a header changed the rendered chart")
	}
}
