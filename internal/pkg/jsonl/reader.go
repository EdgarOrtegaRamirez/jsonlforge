// Package jsonl provides core JSONL reading, writing, and manipulation.
package jsonl

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

// Record represents a single JSONL record as a map.
type Record map[string]interface{}

// Reader reads JSONL records from an input source.
type Reader struct {
	scanner *bufio.Scanner
	lineNum int
	closeFn func() error
}

// NewReader creates a new JSONL reader from a file path.
// If path is "-", reads from stdin.
func NewReader(path string) (*Reader, error) {
	var r io.Reader
	var closeFn func() error

	if path == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		r = f
		closeFn = f.Close
	}

	return &Reader{
		scanner: bufio.NewScanner(r),
		closeFn: closeFn,
	}, nil
}

// NewReaderFromReader creates a new JSONL reader from an io.Reader.
// Useful for testing.
func NewReaderFromReader(r io.Reader) *Reader {
	return &Reader{
		scanner: bufio.NewScanner(r),
	}
}

// Next reads the next line and parses it as a JSON object.
// Returns false if no more records or on parse error.
func (r *Reader) Next() (Record, error) {
	if r.scanner.Scan() {
		line := r.scanner.Bytes()
		r.lineNum++
		// Skip empty lines
		if len(line) == 0 {
			return r.Next()
		}
		var rec Record
		if err := json.Unmarshal(line, &rec); err != nil {
			return nil, &ParseError{Line: r.lineNum, Err: err}
		}
		return rec, nil
	}
	if err := r.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

// Close closes the underlying reader.
func (r *Reader) Close() error {
	if r.closeFn != nil {
		return r.closeFn()
	}
	return nil
}

// ParseError represents a JSON parse error at a specific line.
type ParseError struct {
	Line int
	Err  error
}

func (e *ParseError) Error() string {
	return "parse error at line " + itoa(e.Line) + ": " + e.Err.Error()
}

// Writer writes JSONL records to an output source.
type Writer struct {
	writer  *bufio.Writer
	closeFn func() error
	pretty  bool
	indent  string
}

// NewWriter creates a new JSONL writer.
// If path is "-", writes to stdout.
func NewWriter(path string, pretty bool) (*Writer, error) {
	var w io.Writer
	var closeFn func() error

	if path == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		w = f
		closeFn = f.Close
	}

	return &Writer{
		writer:  bufio.NewWriter(w),
		closeFn: closeFn,
		pretty:  pretty,
		indent:  "  ",
	}, nil
}

// Write writes a single record.
func (w *Writer) Write(rec Record) error {
	var bytes []byte
	var err error

	if w.pretty {
		bytes, err = json.MarshalIndent(rec, "", w.indent)
	} else {
		bytes, err = json.Marshal(rec)
	}
	if err != nil {
		return err
	}

	if _, err := w.writer.Write(bytes); err != nil {
		return err
	}
	if _, err := w.writer.WriteRune('\n'); err != nil {
		return err
	}
	return nil
}

// Close flushes and closes the writer.
func (w *Writer) Close() error {
	if err := w.writer.Flush(); err != nil {
		return err
	}
	if w.closeFn != nil {
		return w.closeFn()
	}
	return nil
}

// MustGetField safely retrieves a field value from a record.
// Returns nil if the field doesn't exist.
func MustGetField(rec Record, key string) interface{} {
	v, ok := rec[key]
	if !ok {
		return nil
	}
	return v
}

// HasField checks if a record has a specific field.
func HasField(rec Record, key string) bool {
	_, ok := rec[key]
	return ok
}

// itoa converts int to string (avoiding fmt import in hot path).
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	var b [20]byte
	idx := len(b)
	for i > 0 {
		idx--
		b[idx] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		idx--
		b[idx] = '-'
	}
	return string(b[idx:])
}
