package cmd

import "fmt"

// fieldSelection keeps one field argument independent of other selections.
// In particular, exclusions apply to this selection alone, not to joined flags.
type fieldSelection struct {
	numbers    []int       // One-based requested columns, negative for exclusions.
	names      []string    // Literal or fuzzy column names, including exclusion prefixes.
	exclude    bool        // Whether the result is the complement of the requested fields.
	openRanges map[int]int // Position in numbers -> first column of an open range.
}

// newFieldSelection parses syntax before any input or output is opened.
func newFieldSelection(spec string) fieldSelection {
	numbers, names, exclude, _, openRanges := parseFields(spec, ",", true, false)
	return fieldSelection{numbers: numbers, names: names, exclude: exclude, openRanges: openRanges}
}

// resolve returns one-based columns in requested order, retaining explicit repeats.
// Exclusions instead return the remaining physical columns in input order.
func (s fieldSelection) resolve(header []string, fuzzy bool) []int {
	var fields []int
	if len(s.names) > 0 {
		for _, name := range s.names {
			if s.exclude {
				name = name[1:]
			}
			matches := 0
			match := func(column string) bool { return column == name }
			if fuzzy {
				match = fuzzyField2Regexp(name).MatchString
			}
			for i, column := range header {
				if match(column) {
					fields = append(fields, i+1)
					matches++
				}
			}
			if !fuzzy && matches == 0 {
				checkError(fmt.Errorf("column %q not found", name))
			}
			if !fuzzy && matches > 1 {
				checkError(fmt.Errorf("the selected colname is duplicated in the input data: %s", name))
			}
		}
	} else {
		for i, field := range s.numbers {
			if s.exclude {
				field = -field
			}
			if _, open := s.openRanges[i]; open {
				for j := field; j <= len(header); j++ {
					fields = append(fields, j)
				}
			} else {
				if !s.exclude && field > len(header) {
					checkError(fmt.Errorf("field (%d) out of range (%d)", field, len(header)))
				}
				fields = append(fields, field)
			}
		}
	}
	if !s.exclude {
		return fields
	}
	excluded := make(map[int]struct{}, len(fields))
	for _, field := range fields {
		excluded[field] = struct{}{}
	}
	fields = fields[:0]
	for i := range header {
		if _, skip := excluded[i+1]; !skip {
			fields = append(fields, i+1)
		}
	}
	return fields
}
