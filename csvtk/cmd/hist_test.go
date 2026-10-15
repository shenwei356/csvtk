package cmd

import (
	"encoding/xml"
	"io"
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
