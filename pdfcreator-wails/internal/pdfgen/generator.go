package pdfgen

import (
	"archive/zip"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"
	"pdfcreator-wails/internal/config"
)

const (
	pageWidthMM  = 297.0
	pageHeightMM = 210.0
	minFontSize  = 8
)

// FitResult contains the fitted font size and measured text area.
type FitResult struct {
	FontSize int
	Width    float64
	Height   float64
	Lines    int
}

// Generator handles PDF generation.
type Generator struct {
	dataDir string
}

// New creates a new Generator.
func New(dataDir string) *Generator {
	return &Generator{dataDir: dataDir}
}

// GenerateImageTemplate creates one PDF per row using an image background.
// It returns the generated paths and the updated fields with fitted font sizes.
// GenerateResult holds generation output and any warnings.
type GenerateResult struct {
	Paths    []string
	Fields   []config.Field
	Warnings []string
}

func (g *Generator) GenerateImageTemplate(templatePath string, fields []config.Field, rows [][]string, columns []string, outputDir string) (*GenerateResult, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	// Fit font sizes using the longest text across all rows for each field.
	fittedFields := g.fitFieldsForRows(templatePath, fields, rows, columns)

	warnings := g.checkWrapWarnings(fittedFields, rows, columns)

	generated := []string{}
	for rowIdx, row := range rows {
		outputName := fmt.Sprintf("diploma_%03d.pdf", rowIdx+1)
		outputPath := filepath.Join(outputDir, outputName)
		if err := g.generateSingle(templatePath, fittedFields, row, columns, outputPath); err != nil {
			return nil, err
		}
		generated = append(generated, outputPath)
	}

	return &GenerateResult{
		Paths:    generated,
		Fields:   fittedFields,
		Warnings: warnings,
	}, nil
}

// GeneratePreview creates a preview PDF for the given row.
// It returns the fitted fields, warnings and an error if generation fails.
func (g *Generator) GeneratePreview(templatePath string, fields []config.Field, row []string, columns []string, outputPath string) ([]config.Field, []string, error) {
	rows := [][]string{row}
	fittedFields := g.fitFieldsForRows(templatePath, fields, rows, columns)
	warnings := g.checkWrapWarnings(fittedFields, rows, columns)
	if err := g.generateSingle(templatePath, fittedFields, row, columns, outputPath); err != nil {
		return nil, nil, err
	}
	return fittedFields, warnings, nil
}

func (g *Generator) generateSingle(templatePath string, fields []config.Field, row []string, columns []string, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	fontDir := filepath.Join(g.dataDir, "fonts")
	pdf := fpdf.New("L", "mm", "A4", fontDir)
	pdf.AddPage()

	// Add background image scaled to A4 landscape.
	pdf.Image(templatePath, 0, 0, pageWidthMM, pageHeightMM, false, "", 0, "")

	for _, field := range fields {
		value := g.cellValue(row, columns, field.Column)
		if value == "" {
			continue
		}
		if err := g.drawText(pdf, field, value); err != nil {
			return err
		}
	}

	return pdf.OutputFileAndClose(outputPath)
}

// FitFields fits font sizes for the given sample rows and returns updated fields.
func (g *Generator) FitFields(templatePath string, fields []config.Field, rows [][]string, columns []string) []config.Field {
	return g.fitFieldsForRows(templatePath, fields, rows, columns)
}

func (g *Generator) fitFieldsForRows(templatePath string, fields []config.Field, rows [][]string, columns []string) []config.Field {
	fontDir := filepath.Join(g.dataDir, "fonts")
	pdf := fpdf.New("L", "mm", "A4", fontDir)
	pdf.AddPage()

	result := make([]config.Field, len(fields))
	copy(result, fields)

	for i, field := range result {
		// Find the longest text for this field across all rows.
		longest := ""
		for _, row := range rows {
			value := g.cellValue(row, columns, field.Column)
			if len(value) > len(longest) {
				longest = value
			}
		}
		if longest == "" {
			continue
		}

		if field.AutoWrap {
			pattern, size := g.fitAutoWrap(pdf, longest, field)
			result[i].FontSize = size
			result[i].WrapPattern = pattern
		} else {
			fit := g.fitText(pdf, longest, field)
			result[i].FontSize = fit.FontSize
			result[i].WrapPattern = nil
		}
	}

	return result
}

func (g *Generator) fitText(pdf *fpdf.Fpdf, text string, field config.Field) FitResult {
	fontKey := field.Font
	pdf.AddUTF8Font(fontKey, "", field.Font)

	lines := strings.Split(text, "\n")

	for size := field.FontSize; size >= minFontSize; size-- {
		pdf.SetFont(fontKey, "", float64(size))
		lineHeight := float64(size) * 0.5
		maxLineW := 0.0
		fits := true

		for _, line := range lines {
			w := pdf.GetStringWidth(line)
			if w > field.MaxWidth {
				fits = false
				break
			}
			if w > maxLineW {
				maxLineW = w
			}
		}

		if fits {
			return FitResult{
				FontSize: size,
				Width:    maxLineW,
				Height:   lineHeight * float64(len(lines)),
				Lines:    len(lines),
			}
		}
	}

	// Fallback to minimum size, measure anyway.
	pdf.SetFont(fontKey, "", minFontSize)
	lineHeight := float64(minFontSize) * 0.5
	maxLineW := 0.0
	for _, line := range lines {
		w := pdf.GetStringWidth(line)
		if w > maxLineW {
			maxLineW = w
		}
	}
	return FitResult{
		FontSize: minFontSize,
		Width:    maxLineW,
		Height:   lineHeight * float64(len(lines)),
		Lines:    len(lines),
	}
}

func (g *Generator) drawText(pdf *fpdf.Fpdf, field config.Field, text string) error {
	fontPath := filepath.Join(g.dataDir, "fonts", field.Font)
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		return fmt.Errorf("font not found: %s", field.Font)
	}

	fontKey := field.Font
	pdf.AddUTF8Font(fontKey, "", field.Font)
	pdf.SetFont(fontKey, "", float64(field.FontSize))

	r, gt, b, err := parseHexColor(field.Color)
	if err != nil {
		return err
	}
	pdf.SetTextColor(r, gt, b)

	x := field.X
	y := field.Y
	align := string(field.Align)
	if align == "" {
		align = "L"
	}

	if field.AutoWrap && len(field.WrapPattern) > 0 {
		words := splitWords(text)
		lines := applyWrapPattern(words, field.WrapPattern)
		text = strings.Join(lines, "\n")
	}

	lineHeight := float64(field.FontSize) * 0.5
	pdf.SetXY(x, y)
	pdf.MultiCell(field.MaxWidth, lineHeight, text, "", align, false)
	return nil
}

func splitWords(text string) []string {
	return strings.Fields(text)
}

// fitAutoWrap finds a wrap pattern for the longest text and a font size that fits.
func (g *Generator) fitAutoWrap(pdf *fpdf.Fpdf, text string, field config.Field) ([]int, int) {
	words := splitWords(text)
	if len(words) == 0 {
		return nil, field.FontSize
	}

	for size := field.FontSize; size >= minFontSize; size-- {
		pattern := g.buildWrapPattern(pdf, words, field.MaxWidth, field.Font, size)
		if g.canFitPattern(pdf, words, pattern, field.MaxWidth, field.Font, size) {
			return pattern, size
		}
	}

	// Fallback to minimum size.
	pattern := g.buildWrapPattern(pdf, words, field.MaxWidth, field.Font, minFontSize)
	return pattern, minFontSize
}

func (g *Generator) buildWrapPattern(pdf *fpdf.Fpdf, words []string, maxWidth float64, font string, size int) []int {
	pdf.AddUTF8Font(font, "", font)
	pdf.SetFont(font, "", float64(size))
	spaceWidth := pdf.GetStringWidth(" ")

	pattern := []int{}
	currentCount := 0
	currentWidth := 0.0

	for _, word := range words {
		w := pdf.GetStringWidth(word)
		if currentCount == 0 {
			currentCount = 1
			currentWidth = w
		} else {
			newWidth := currentWidth + spaceWidth + w
			if newWidth <= maxWidth {
				currentCount++
				currentWidth = newWidth
			} else {
				pattern = append(pattern, currentCount)
				currentCount = 1
				currentWidth = w
			}
		}
	}
	if currentCount > 0 {
		pattern = append(pattern, currentCount)
	}
	return pattern
}

func (g *Generator) canFitPattern(pdf *fpdf.Fpdf, words []string, pattern []int, maxWidth float64, font string, size int) bool {
	pdf.AddUTF8Font(font, "", font)
	pdf.SetFont(font, "", float64(size))
	lines := applyWrapPattern(words, pattern)
	for _, line := range lines {
		if pdf.GetStringWidth(line) > maxWidth {
			return false
		}
	}
	return true
}

func applyWrapPattern(words []string, pattern []int) []string {
	if len(pattern) == 0 {
		return []string{strings.Join(words, " ")}
	}

	lines := make([]string, 0, len(pattern))
	idx := 0
	for i, count := range pattern {
		if i == len(pattern)-1 {
			lines = append(lines, strings.Join(words[idx:], " "))
			break
		}
		end := idx + count
		if end > len(words) {
			end = len(words)
		}
		lines = append(lines, strings.Join(words[idx:end], " "))
		idx = end
	}
	return lines
}

func (g *Generator) checkWrapWarnings(fields []config.Field, rows [][]string, columns []string) []string {
	var warnings []string
	for rowIdx, row := range rows {
		for _, field := range fields {
			if !field.AutoWrap || len(field.WrapPattern) == 0 {
				continue
			}
			value := g.cellValue(row, columns, field.Column)
			words := splitWords(value)
			if len(words) < len(field.WrapPattern) {
				warnings = append(warnings, fmt.Sprintf("Строка %d, поле %s: слишком мало слов (%d) для шаблона (%d строк)", rowIdx+1, field.Column, len(words), len(field.WrapPattern)))
			}
		}
	}
	return warnings
}

func (g *Generator) cellValue(row []string, columns []string, column string) string {
	for i, col := range columns {
		if col == column && i < len(row) {
			return row[i]
		}
	}
	return ""
}

// CreateZIP packs generated PDFs into a zip file.
func CreateZIP(pdfPaths []string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	defer zw.Close()

	for _, path := range pdfPaths {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.Base(path)
		header.Method = zip.Deflate

		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, f)
		f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// LongestText returns the longest value for a column across rows.
func LongestText(rows [][]string, columns []string, column string) string {
	longest := ""
	for _, row := range rows {
		for i, col := range columns {
			if col == column && i < len(row) {
				value := row[i]
				if len(value) > len(longest) {
					longest = value
				}
			}
		}
	}
	return longest
}

// RoundFloat rounds a float64 to the given precision.
func RoundFloat(val float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Round(val*p) / p
}
