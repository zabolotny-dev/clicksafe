package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
)

const utf8BOM = "\uFEFF"

// Config describes reusable CSV mechanics.
//
// It intentionally does not know anything about domain entities,
// bus commands, employees, departments, etc.
type Config struct {
	// RequiredHeaders are validated after header normalization.
	//
	// Example:
	//   []string{"first_name", "last_name", "email"}
	RequiredHeaders []string

	// TrimValues controls whether Row.Get and Row.Prefixed trim cell values.
	//
	// Usually true for import flows.
	TrimValues bool
}

// Reader is a small row-based CSV reader.
//
// It behaves similarly to bufio.Scanner:
//
//	reader, err := csvimport.NewReader(file, cfg)
//	if err != nil { ... }
//
//	for reader.Next() {
//	    row := reader.Row()
//	    ...
//	}
//
//	if errs := reader.Err(); len(errs) > 0 { ... }
type Reader struct {
	r       *csv.Reader
	headers map[string]int

	rowNum int
	row    Row
	errs   []RowError

	cfg Config
}

// RowError describes a CSV parsing error bound to a source row.
type RowError struct {
	Row int    `json:"row"`
	Err string `json:"error"`
}

// Row is one non-empty CSV record plus normalized header lookup.
type Row struct {
	num     int
	record  []string
	headers map[string]int
	cfg     Config
}

// NewReader creates a CSV import reader.
//
// Defaults:
//   - FieldsPerRecord = -1
//   - TrimLeadingSpace = true
//   - first header UTF-8 BOM is stripped
//   - headers are normalized with trim + lowercase
//   - empty header names are ignored
func NewReader(src io.Reader, cfg Config) (*Reader, error) {
	if src == nil {
		return nil, errors.New("csv source is nil")
	}

	r := csv.NewReader(src)
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	headerRecord, err := r.Read()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("csv header is missing")
		}

		return nil, fmt.Errorf("read csv header: %w", err)
	}

	headers := buildHeaderIndex(headerRecord)

	if err := validateRequiredHeaders(headers, cfg.RequiredHeaders); err != nil {
		return nil, err
	}

	return &Reader{
		r:       r,
		headers: headers,
		rowNum:  1, // header is row 1
		cfg:     cfg,
	}, nil
}

// Next advances to the next non-empty CSV row.
//
// Empty rows are skipped.
// If a CSV parsing error happens, Next returns false and Err returns row errors.
func (r *Reader) Next() bool {
	if r == nil || len(r.errs) > 0 {
		return false
	}

	for {
		record, err := r.r.Read()
		if errors.Is(err, io.EOF) {
			return false
		}

		nextRowNum := r.rowNum + 1

		if err != nil {
			r.errs = append(r.errs, newRowError(nextRowNum, err))
			return false
		}

		r.rowNum = nextRowNum

		if isEmptyRecord(record) {
			continue
		}

		r.row = Row{
			num:     r.rowNum,
			record:  record,
			headers: r.headers,
			cfg:     r.cfg,
		}

		return true
	}
}

// Row returns the current row.
//
// It should only be called after Next returns true.
func (r *Reader) Row() Row {
	if r == nil {
		return Row{}
	}

	return r.row
}

// Err returns CSV parsing errors encountered by Reader.
func (r *Reader) Err() []RowError {
	if r == nil {
		return nil
	}

	return r.errs
}

// Headers returns normalized CSV headers.
//
// The returned map is a copy.
func (r *Reader) Headers() map[string]int {
	if r == nil {
		return nil
	}

	out := make(map[string]int, len(r.headers))
	for k, v := range r.headers {
		out[k] = v
	}

	return out
}

// Number returns the original 1-based CSV row number.
//
// Header is row 1, first data row is row 2.
func (row Row) Number() int {
	return row.num
}

// Get returns a cell by header name.
//
// Header lookup is normalized with trim + lowercase.
// Missing headers and missing cells return an empty string.
func (row Row) Get(header string) string {
	idx, ok := row.headers[NormalizeHeader(header)]
	if !ok || idx < 0 || idx >= len(row.record) {
		return ""
	}

	return row.cleanValue(row.record[idx])
}

// Has reports whether a normalized header exists.
func (row Row) Has(header string) bool {
	_, ok := row.headers[NormalizeHeader(header)]
	return ok
}

// Prefixed returns dynamic columns with the given prefix.
//
// Example:
//
// Headers:
//
//	attr.region, attr.role
//
// row.Prefixed("attr.") returns:
//
//	map[string]string{
//	    "region": "EU",
//	    "role": "manager",
//	}
//
// Empty keys and empty values are skipped.
func (row Row) Prefixed(prefix string) map[string]string {
	prefix = NormalizeHeader(prefix)

	out := make(map[string]string)

	for header, idx := range row.headers {
		if !strings.HasPrefix(header, prefix) {
			continue
		}

		if idx < 0 || idx >= len(row.record) {
			continue
		}

		key := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		value := row.cleanValue(row.record[idx])

		if key == "" || value == "" {
			continue
		}

		out[key] = value
	}

	return out
}

// Values returns all known header values for this row.
//
// Unknown extra cells without headers are ignored.
// Empty values are included.
func (row Row) Values() map[string]string {
	out := make(map[string]string, len(row.headers))

	for header, idx := range row.headers {
		if idx < 0 || idx >= len(row.record) {
			out[header] = ""
			continue
		}

		out[header] = row.cleanValue(row.record[idx])
	}

	return out
}

// IsEmpty reports whether the row has no non-whitespace values.
func (row Row) IsEmpty() bool {
	return isEmptyRecord(row.record)
}

// NormalizeHeader is exported so app packages can normalize their own
// header constants consistently with the reader.
func NormalizeHeader(s string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(s, utf8BOM)))
}

func buildHeaderIndex(headerRecord []string) map[string]int {
	headers := make(map[string]int, len(headerRecord))

	for i, h := range headerRecord {
		normalized := NormalizeHeader(h)
		if normalized == "" {
			continue
		}

		// If duplicate headers exist, keep the first one.
		// This makes behavior deterministic and avoids later duplicate columns
		// silently overriding earlier values.
		if _, exists := headers[normalized]; exists {
			continue
		}

		headers[normalized] = i
	}

	return headers
}

func validateRequiredHeaders(headers map[string]int, required []string) error {
	var missing []string

	for _, h := range required {
		normalized := NormalizeHeader(h)
		if normalized == "" {
			continue
		}

		if _, ok := headers[normalized]; !ok {
			missing = append(missing, normalized)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required csv headers: %s", strings.Join(missing, ", "))
	}

	return nil
}

func newRowError(row int, err error) RowError {
	var parseErr *csv.ParseError
	if errors.As(err, &parseErr) && parseErr.Line > 0 {
		row = parseErr.Line
	}

	return RowError{
		Row: row,
		Err: err.Error(),
	}
}

func isEmptyRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}

	return true
}

func (row Row) cleanValue(value string) string {
	if row.cfg.TrimValues {
		return strings.TrimSpace(value)
	}

	return value
}
