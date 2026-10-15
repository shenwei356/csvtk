package cmd

import (
	"encoding/csv"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// writeJSONCSVInput retains special characters in CSV headers and field values.
func writeJSONCSVInput(t *testing.T, rows [][]string) string {
	t.Helper()
	var text strings.Builder
	if err := csv.NewWriter(&text).WriteAll(rows); err != nil {
		t.Fatal(err)
	}
	return writeAccuracyInput(t, "input.csv", text.String())
}

// decodeCSV2JSON checks the entire document and retains exact numeric literals,
// including numbers beyond the float64 range and integer precision.
func decodeCSV2JSON(t *testing.T, text string) interface{} {
	t.Helper()
	if !json.Valid([]byte(text)) {
		t.Fatalf("invalid JSON: %q", text)
	}
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	var value interface{}
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

// TestCSV2JSONStrings checks decoded headers, object keys, and field values for
// exact preservation across keyed/unkeyed and object/array output modes.
func TestCSV2JSONStrings(t *testing.T) {
	headers := []string{"id", `back\slash`, `double"quote`, "line\nheader", "tab\theader", "bare\rheader", "control\x00\x01\b\f", "<>&\u2028\u2029沈伟"}
	var data [][]string
	for _, key := range headers {
		row := []string{"key:" + key}
		for _, value := range headers[1:] {
			row = append(row, "value:"+value)
		}
		data = append(data, row)
	}
	for _, noHeader := range []bool{false, true} {
		rows := data
		if !noHeader {
			rows = append([][]string{headers}, data...)
		}
		input := writeJSONCSVInput(t, rows)
		for _, keyed := range []bool{false, true} {
			for _, indent := range []string{"", "  "} {
				name := "header=" + strconv.FormatBool(!noHeader) + "/keyed=" + strconv.FormatBool(keyed) + "/indent=" + strconv.Quote(indent)
				t.Run(name, func(t *testing.T) {
					args := []string{"csv2json", "-i", indent}
					if noHeader {
						args = append(args, "-H")
					}
					if keyed {
						args = append(args, "-k", "1")
					}
					var list []interface{}
					object := make(map[string]interface{})
					for _, row := range data {
						var value interface{}
						if noHeader {
							fields := make([]interface{}, len(row))
							for i, field := range row {
								fields[i] = field
							}
							value = fields
						} else {
							fields := make(map[string]interface{}, len(row))
							for i, field := range row {
								fields[headers[i]] = field
							}
							value = fields
						}
						list = append(list, value)
						object[row[0]] = value
					}
					var want interface{} = list
					if keyed {
						want = object
					}
					out, _ := runAccuracyCLI(t, append(args, input)...)
					if got := decodeCSV2JSON(t, out); !reflect.DeepEqual(got, want) {
						t.Fatalf("decoded fields changed: got %#v, want %#v", got, want)
					}
				})
			}
		}
	}
}

// TestCSV2JSONSkipKey checks separators when omitting the first, middle, last,
// or only field, including empty objects/arrays and compact JSON output.
func TestCSV2JSONSkipKey(t *testing.T) {
	for _, noHeader := range []bool{false, true} {
		for _, indent := range []string{"", "  "} {
			for _, key := range []int{1, 2, 3, 0} {
				name := "header=" + strconv.FormatBool(!noHeader) + "/indent=" + strconv.Quote(indent) + "/key=" + strconv.Itoa(key)
				t.Run(name, func(t *testing.T) {
					headers := []string{"a", "b", "c"}
					data := [][]string{{"r1a", "r1b", "r1c"}, {"r2a", "r2b", "r2c"}}
					if key == 0 {
						headers = headers[:1]
						data = [][]string{{"r1a"}, {"r2a"}}
						key = 1
					}
					rows := data
					if !noHeader {
						rows = append([][]string{headers}, data...)
					}
					input := writeJSONCSVInput(t, rows)
					args := []string{"csv2json", "-k", strconv.Itoa(key), "-K", "-i", indent}
					if noHeader {
						args = append(args, "-H")
					}
					want := make(map[string]interface{})
					for _, row := range data {
						fields := make([]interface{}, 0, len(row)-1)
						object := make(map[string]interface{})
						for i, field := range row {
							if i+1 != key {
								fields = append(fields, field)
								object[headers[i]] = field
							}
						}
						if noHeader {
							want[row[key-1]] = fields
						} else {
							want[row[key-1]] = object
						}
					}
					out, _ := runAccuracyCLI(t, append(args, input)...)
					if got := decodeCSV2JSON(t, out); !reflect.DeepEqual(got, want) {
						t.Fatalf("got %#v, want %#v", got, want)
					}
				})
			}
		}
	}
}

// TestCSV2JSONNumbers checks numeric syntax and exact decimal precision without
// decoding through float64, while retaining strings in unselected columns.
func TestCSV2JSONNumbers(t *testing.T) {
	values := [][2]string{
		{"00", "0"}, {"+00", "0"}, {"-00", "-0"},
		{".5", "0.5"}, {"+.5", "0.5"}, {"-.5", "-0.5"},
		{"1.", "1.0"}, {"001.2300", "1.2300"}, {"-000.50", "-0.50"},
		{"00e+003", "0e+003"}, {"+001.E-0002", "1.0E-0002"},
		{"9007199254740993", "9007199254740993"},
		{"000123456789012345678901234567890.123456789", "123456789012345678901234567890.123456789"},
		{"1e999999999999999999999999999999", "1e999999999999999999999999999999"},
		{"-1e-999999999999999999999999999999", "-1e-999999999999999999999999999999"},
		{" 1", ""}, {"1 ", ""}, {"1,000", ""}, {"NaN", ""}, {"Inf", ""}, {"1e", ""}, {"--1", ""}, {"0x10", ""},
	}
	data := make([][]string, len(values))
	for i, pair := range values {
		data[i] = []string{pair[0], pair[0], "row" + strconv.Itoa(i)}
	}
	for _, noHeader := range []bool{false, true} {
		rows := data
		if !noHeader {
			rows = append([][]string{{"n", "copy", "id"}}, data...)
		}
		input := writeJSONCSVInput(t, rows)
		for _, parseNum := range []string{"", "1", "all"} {
			for _, indent := range []string{"", "  "} {
				name := "header=" + strconv.FormatBool(!noHeader) + "/numeric=" + parseNum + "/indent=" + strconv.Quote(indent)
				t.Run(name, func(t *testing.T) {
					args := []string{"csv2json", "-k", "3", "-K", "-i", indent}
					if noHeader {
						args = append(args, "-H")
					}
					if parseNum != "" {
						args = append(args, "-n", parseNum)
					}
					want := make(map[string]interface{})
					for i, pair := range values {
						var number, copied interface{} = pair[0], pair[0]
						if parseNum != "" && pair[1] != "" {
							number = json.Number(pair[1])
							if parseNum == "all" {
								copied = number
							}
						}
						if noHeader {
							want[data[i][2]] = []interface{}{number, copied}
						} else {
							want[data[i][2]] = map[string]interface{}{"n": number, "copy": copied}
						}
					}
					out, _ := runAccuracyCLI(t, append(args, input)...)
					if got := decodeCSV2JSON(t, out); !reflect.DeepEqual(got, want) {
						t.Fatalf("numeric precision or type changed: got %#v, want %#v", got, want)
					}
				})
			}
		}
	}
}

// TestCSV2JSONValueTypes preserves automatic boolean/null handling and --blanks.
// --skip-key alone must leave all fields in the unkeyed output.
func TestCSV2JSONValueTypes(t *testing.T) {
	values := []string{"TRUE", "False", "", "nA", "N/A", "None", "nULL", ".", "text"}
	input := writeJSONCSVInput(t, [][]string{values})
	for _, blanks := range []bool{false, true} {
		t.Run("blanks="+strconv.FormatBool(blanks), func(t *testing.T) {
			args := []string{"csv2json", "-H", "-K", "-n", "all"}
			if blanks {
				args = append(args, "--blanks")
			}
			want := []interface{}{true, false}
			for _, value := range values[2 : len(values)-1] {
				if blanks {
					want = append(want, value)
				} else {
					want = append(want, nil)
				}
			}
			want = append(want, "text")
			out, _ := runAccuracyCLI(t, append(args, input)...)
			if got := decodeCSV2JSON(t, out); !reflect.DeepEqual(got, []interface{}{want}) {
				t.Fatalf("got %#v, want %#v", got, want)
			}
		})
	}
}

// TestCSV2JSONEmptyInput checks complete empty documents with and without keys.
func TestCSV2JSONEmptyInput(t *testing.T) {
	for _, text := range []string{"", "id\n"} {
		input := writeAccuracyInput(t, "empty.csv", text)
		for _, keyed := range []bool{false, true} {
			t.Run("header="+strconv.FormatBool(text != "")+"/keyed="+strconv.FormatBool(keyed), func(t *testing.T) {
				args := []string{"csv2json"}
				var want interface{} = []interface{}{}
				if keyed {
					args = append(args, "-k", "1", "-K")
					want = map[string]interface{}{}
				}
				out, _ := runAccuracyCLI(t, append(args, input)...)
				if got := decodeCSV2JSON(t, out); !reflect.DeepEqual(got, want) {
					t.Fatalf("got %#v, want %#v", got, want)
				}
			})
		}
	}
}
