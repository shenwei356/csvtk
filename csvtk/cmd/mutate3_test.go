package cmd

import (
	"os"
	"runtime"
	"testing"
)

func TestMutate3(t *testing.T) {
	cases := []struct {
		expect   string
		noHeader bool
		opts     mutate3Opts
		tabs     bool
	}{
		// Strings
		{
			opts: mutate3Opts{
				ExprStr: ` $first_name + " " + $last_name `,
				Files:   []string{"../../testdata/names.csv"},
				Name:    "full_name",
			},
			expect: `id,first_name,last_name,username,full_name
11,Rob,Pike,rob,Rob Pike
2,Ken,Thompson,ken,Ken Thompson
4,Robert,Griesemer,gri,Robert Griesemer
1,Robert,Thompson,abc,Robert Thompson
NA,Robert,Abel,123,Robert Abel
`,
		},

		// Constants
		{
			tabs:     true,
			noHeader: true,
			opts: mutate3Opts{
				ExprStr: ` "abc" `,
				Files:   []string{"../../testdata/digitals.tsv"},
			},
			expect: `4	5	6	abc
1	2	3	abc
7	8	0	abc
8	1,000	4	abc
`,
		},

		// Math
		{
			tabs:     true,
			noHeader: true,
			opts: mutate3Opts{
				ExprStr:      ` $1 + $3 `,
				Files:        []string{"../../testdata/digitals.tsv"},
				DecimalWidth: 0,
			},
			expect: `4	5	6	10
1	2	3	4
7	8	0	7
8	1,000	4	12
`,
		},

		// Bool
		{
			tabs:     true,
			noHeader: true,
			opts: mutate3Opts{
				ExprStr: ` $1 > 5 `,
				Files:   []string{"../../testdata/digitals.tsv"},
			},
			expect: `4	5	6	false
1	2	3	false
7	8	0	true
8	1,000	4	true
`,
		},

		// Ternary
		{
			tabs:     true,
			noHeader: true,
			opts: mutate3Opts{
				ExprStr: `$1 > 5 ? "big" : "small"`,
				Files:   []string{"../../testdata/digitals.tsv"},
			},
			expect: `4	5	6	small
1	2	3	small
7	8	0	big
8	1,000	4	big
`,
		},

		// Null coalescence
		{
			opts: mutate3Opts{
				ExprStr: `$one ?? $two`,
				Files:   []string{"../../testdata/null_coalescence.csv"},
				Name:    "three",
			},
			expect: `one,two,three
a1,a2,a1
,b2,b2
a2,,a2
`,
		},

		// Position: --at 1
		{
			opts: mutate3Opts{
				ExprStr:      `$a+$c`,
				Files:        []string{"../../testdata/positions.csv"},
				Name:         "x",
				DecimalWidth: 0,
				At:           1,
			},
			expect: `x,a,b,c
4,1,2,3
`,
		},

		// Position: --at 3
		{
			opts: mutate3Opts{
				ExprStr:      `$a+$c`,
				Files:        []string{"../../testdata/positions.csv"},
				Name:         "x",
				DecimalWidth: 0,
				At:           3,
			},
			expect: `a,b,x,c
1,2,4,3
`,
		},

		// Position: --after a
		{
			opts: mutate3Opts{
				ExprStr:      `$a+$c`,
				Files:        []string{"../../testdata/positions.csv"},
				Name:         "x",
				DecimalWidth: 0,
				After:        "a",
			},
			expect: `a,x,b,c
1,4,2,3
`,
		},

		// Position: --before c
		{
			opts: mutate3Opts{
				ExprStr:      `$a+$c`,
				Files:        []string{"../../testdata/positions.csv"},
				Name:         "x",
				DecimalWidth: 0,
				Before:       "c",
			},
			expect: `a,b,x,c
1,2,4,3
`,
		},

		// Date math
		{
			opts: mutate3Opts{
				ExprStr: `(date(${Out}) - date($In)).Hours() | int()`,
				Files:   []string{"../../testdata/datesub.csv"},
				Name:    "Hours",
			},
			expect: `ID,Name,In,Out,Hours
1,Tom,2023-08-25 11:24:00,2023-08-27 08:33:02,45
2,Sally,2023-08-25 11:28:00,2023-08-26 14:17:35,26
3,Alf,2023-08-26 11:29:00,2023-08-29 20:43:00,81
`,
		},
	}

	for _, c := range cases {
		f, err := os.CreateTemp("", "outfile")
		if err != nil {
			t.Fatalf("failed to open temp file: %s\n", err)
		}
		defer os.Remove(f.Name())

		config := Config{
			CommentChar:  '#',
			Delimiter:    ',',
			NoHeaderRow:  c.noHeader,
			NumCPUs:      runtime.NumCPU(),
			OutDelimiter: ',',
			OutFile:      f.Name(),
			Tabs:         c.tabs,
		}

		doMutate3(config, c.opts)

		output, err := os.ReadFile(f.Name())
		if err != nil {
			t.Fatalf("failed to read temp file %q: %s\n", f.Name(), err)
		}

		if string(output) != c.expect {
			t.Errorf("test failed:\noptions:\n\t%#v\nwant:\n\t%q\ngot:\n\t%q\n", c.opts, c.expect, output)
		}
	}
}
