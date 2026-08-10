package excel

import (
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// Reader wraps an opened Excel file.
type Reader struct {
	path  string
	file  *excelize.File
	sheet string
}

// Open opens an Excel file and selects the active sheet.
func Open(path string) (*Reader, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}

	sheet := f.GetActiveSheetIndex()
	sheetName := f.GetSheetName(sheet)
	if sheetName == "" {
		sheetName = f.GetSheetName(0)
	}

	return &Reader{
		path:  path,
		file:  f,
		sheet: sheetName,
	}, nil
}

// Close closes the underlying Excel file.
func (r *Reader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// Columns returns the first row values as column names.
func (r *Reader) Columns() ([]string, error) {
	rows, err := r.file.GetRows(r.sheet)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []string{}, nil
	}

	cols := make([]string, len(rows[0]))
	for i, cell := range rows[0] {
		if cell == "" {
			cols[i] = fmt.Sprintf("Column%d", i+1)
		} else {
			cols[i] = cell
		}
	}
	return cols, nil
}

// Preview returns the first data rows (skipping header) up to limit.
func (r *Reader) Preview(limit int) ([][]string, error) {
	rows, err := r.file.GetRows(r.sheet)
	if err != nil {
		return nil, err
	}
	if len(rows) <= 1 {
		return [][]string{}, nil
	}

	data := rows[1:]
	if limit > 0 && len(data) > limit {
		data = data[:limit]
	}

	result := make([][]string, len(data))
	for i, row := range data {
		result[i] = make([]string, len(row))
		for j, cell := range row {
			result[i][j] = cell
		}
	}
	return result, nil
}

// RowCount returns the number of data rows excluding header.
func (r *Reader) RowCount() (int, error) {
	rows, err := r.file.GetRows(r.sheet)
	if err != nil {
		return 0, err
	}
	if len(rows) <= 1 {
		return 0, nil
	}
	return len(rows) - 1, nil
}

// AllRows returns all data rows excluding header.
func (r *Reader) AllRows() ([][]string, error) {
	return r.Preview(0)
}

// ToNumber tries to convert a string to float64.
func ToNumber(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
