package cmd

import (
	"encoding/xml"
	"io"
	"os"
	"strings"
	"testing"
)

// TestHistPercentileStatistics checks the statistics actually rendered in the
// SVG label, using skewed samples whose means differ from their medians.
func TestHistPercentileStatistics(t *testing.T) {
	cases := []struct {
		name, input, label, percentiles, statistics string
		args                                        []string
	}{
		{
			name:        "skewed sample",
			input:       "x\n100\n1\n1\n",
			label:       "x",
			percentiles: "P99=100.000 P95=100.000",
			statistics:  "MEAN=34.000 STDDEV=57.158",
			args:        []string{"-f", "x"},
		},
		{
			name:        "even sample without header",
			input:       "0\n8\n0\n0\n",
			label:       "distribution",
			percentiles: "P99=8.000 P95=8.000",
			statistics:  "MEAN=2.000 STDDEV=4.000",
			args:        []string{"-H", "-f", "1", "--xlab", "distribution"},
		},
		{
			name:        "negative values and skipped NA",
			input:       "id,value\nr1,NA\nr2,-10\nr3,\nr4,1\nr5,0\n",
			label:       "selected data",
			percentiles: "P99=1.000 P95=1.000",
			statistics:  "MEAN=-3.000 STDDEV=6.083",
			args:        []string{"-f", "value", "--skip-na", "--xlab", "selected data"},
		},
	}
	for _, tc := range cases {
		input := writeAccuracyInput(t, "input.csv", tc.input)
		for _, showStatistics := range []bool{true, false} {
			name := tc.name + "/plain"
			if showStatistics {
				name = tc.name + "/percentiles"
			}
			t.Run(name, func(t *testing.T) {
				args := append([]string{"plot", "hist", "--bins", "2", "--format", "svg"}, tc.args...)
				if showStatistics {
					args = append(args, "--percentiles")
				}
				out, _ := runAccuracyCLI(t, append(args, input)...)
				decoder := xml.NewDecoder(strings.NewReader(out))
				var labels []string
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
						labels = append(labels, label)
					}
				}
				text := "\n" + strings.Join(labels, "\n") + "\n"
				if !strings.Contains(text, "\n"+tc.label+"\n") {
					t.Fatalf("X label %q missing from rendered labels: %s", tc.label, text)
				}
				if showStatistics {
					for _, want := range []string{tc.percentiles, tc.statistics} {
						if !strings.Contains(text, "\n"+want+"\n") {
							t.Fatalf("expected %q in rendered labels: %s", want, text)
						}
					}
				} else if strings.Contains(text, "P99=") || strings.Contains(text, "MEAN=") || strings.Contains(text, "STDDEV=") {
					t.Fatalf("statistics rendered without --percentiles: %s", text)
				}
			})
		}
	}
}

// TestHistogramUnrepresentableBins rejects finite samples whose bins lose precision.
func TestHistogramUnrepresentableBins(t *testing.T) {
	for _, values := range []string{"-1e308\n0\n1e308", "1e308\n1e308", "5e-324\n1e-323"} {
		input := writeAccuracyInput(t, "hist.csv", "v\n"+values+"\n")
		output := writeAccuracyInput(t, "hist.svg", "sentinel")
		assertInputError(t, "histogram range", "plot", "hist", "-o", output, input)
		data, err := os.ReadFile(output)
		if err != nil || string(data) != "sentinel" {
			t.Fatalf("output changed: %q, %v", data, err)
		}
	}
}

// TestHistogramNoValues reports empty samples instead of computing undefined percentiles.
func TestHistogramNoValues(t *testing.T) {
	for _, content := range []string{"v\n", "v\nNA\n"} {
		input := writeAccuracyInput(t, "hist.csv", content)
		for _, percentile := range []bool{false, true} {
			args := []string{"plot", "hist", "--skip-na", "--format", "svg"}
			if percentile {
				args = append(args, "--percentiles")
			}
			assertInputError(t, "no numeric data", append(args, input)...)
		}
	}
}

// TestHistogramNonFiniteValues rejects values that the histogram binning cannot handle.
func TestHistogramNonFiniteValues(t *testing.T) {
	for _, value := range []string{"NaN", "Inf", "-Inf"} {
		input := writeAccuracyInput(t, "hist.csv", "v\n1\n"+value+"\n")
		assertInputError(t, "non-finite histogram value", "plot", "hist", "--format", "svg", input)
		out, _ := runAccuracyCLI(t, "plot", "hist", "--skip-na", "--na-values", value, "--format", "svg", input)
		if !strings.Contains(out, "<svg") || strings.Contains(out, "NaN") {
			t.Fatal("invalid SVG after skipping NA")
		}
	}
}
