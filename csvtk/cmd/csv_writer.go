package cmd

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

// csvOutputOption contains options for CSV output writer.
type csvOutputOption struct {
	QuoteAll bool
}

// csvOutputWriter keeps the standard CSV writer for the default output and
// writes every field quoted when quoteAll is requested.
type csvOutputWriter struct {
	*csv.Writer
	quoted   *bufio.Writer
	quoteAll bool
}

func newCSVOutputWriter(out io.Writer, opt csvOutputOption) *csvOutputWriter {
	w := &csvOutputWriter{Writer: csv.NewWriter(out), quoteAll: opt.QuoteAll}
	if opt.QuoteAll {
		w.quoted = bufio.NewWriter(out)
	}
	return w
}

func (w *csvOutputWriter) Write(record []string) error {
	if !w.quoteAll {
		return w.Writer.Write(record)
	}
	if w.Comma == 0 || w.Comma == '"' || w.Comma == '\r' || w.Comma == '\n' || !utf8.ValidRune(w.Comma) || w.Comma == utf8.RuneError {
		return errors.New("csv: invalid field or comment delimiter")
	}
	for i, field := range record {
		if i > 0 {
			if _, err := w.quoted.WriteRune(w.Comma); err != nil {
				return err
			}
		}
		if err := w.quoted.WriteByte('"'); err != nil {
			return err
		}
		for len(field) > 0 {
			i := strings.IndexAny(field, "\"\r\n")
			if i < 0 {
				i = len(field)
			}
			if _, err := w.quoted.WriteString(field[:i]); err != nil {
				return err
			}
			field = field[i:]
			if len(field) == 0 {
				break
			}
			var err error
			switch field[0] {
			case '"':
				_, err = w.quoted.WriteString(`""`)
			case '\r':
				if !w.UseCRLF {
					err = w.quoted.WriteByte('\r')
				}
			case '\n':
				if w.UseCRLF {
					_, err = w.quoted.WriteString("\r\n")
				} else {
					err = w.quoted.WriteByte('\n')
				}
			}
			if err != nil {
				return err
			}
			field = field[1:]
		}
		if err := w.quoted.WriteByte('"'); err != nil {
			return err
		}
	}
	if w.UseCRLF {
		_, err := w.quoted.WriteString("\r\n")
		return err
	}
	return w.quoted.WriteByte('\n')
}

func (w *csvOutputWriter) Flush() {
	if w.quoteAll {
		w.quoted.Flush()
	} else {
		w.Writer.Flush()
	}
}

func (w *csvOutputWriter) Error() error {
	if w.quoteAll {
		_, err := w.quoted.Write(nil)
		return err
	}
	return w.Writer.Error()
}

func (w *csvOutputWriter) WriteAll(records [][]string) error {
	for _, record := range records {
		if err := w.Write(record); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
