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
	QuoteAll        bool
	ImmediateOutput bool // Flush each complete record, including downstream buffers.
}

// csvOutputWriter keeps the standard CSV writer for the default output and
// writes every field quoted when quoteAll is requested.
type csvOutputWriter struct {
	*csv.Writer
	quoted          *bufio.Writer
	quoteAll        bool
	immediateOutput bool         // Whether Write must make each record visible immediately.
	flushOutput     func() error // Flushes a buffered or compressed destination, when present.
	flushErr        error        // Preserve downstream flush errors for Error and subsequent writes.
}

// newCSVOutputWriter retains downstream flushing when immediate output is enabled.
func newCSVOutputWriter(out io.Writer, opt csvOutputOption) *csvOutputWriter {
	w := &csvOutputWriter{Writer: csv.NewWriter(out), quoteAll: opt.QuoteAll}
	w.immediateOutput = opt.ImmediateOutput
	if opt.ImmediateOutput {
		if flusher, ok := out.(interface{ Flush() error }); ok {
			w.flushOutput = flusher.Flush
		}
	}
	if opt.QuoteAll {
		w.quoted = bufio.NewWriter(out)
	}
	return w
}

// Write emits one CSV record and optionally flushes every output buffer.
func (w *csvOutputWriter) Write(record []string) error {
	if w.flushErr != nil {
		return w.flushErr
	}
	var err error
	if w.quoteAll {
		err = w.writeQuotedRecord(record)
	} else {
		err = w.Writer.Write(record)
	}
	if err != nil {
		return err
	}
	if w.immediateOutput {
		w.Flush()
		return w.Error()
	}
	return nil
}

// writeQuotedRecord formats one record with every field quoted.
func (w *csvOutputWriter) writeQuotedRecord(record []string) error {
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

// Flush releases CSV data and, in immediate mode, the destination's buffers.
func (w *csvOutputWriter) Flush() {
	if w.quoteAll {
		w.quoted.Flush()
	} else {
		w.Writer.Flush()
	}
	// Flushing only the CSV buffer leaves small records in xopen's buffer.
	if w.flushOutput != nil && w.Error() == nil {
		w.flushErr = w.flushOutput()
	}
}

// Error reports formatting, write, and downstream flush failures.
func (w *csvOutputWriter) Error() error {
	if w.flushErr != nil {
		return w.flushErr
	}
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
