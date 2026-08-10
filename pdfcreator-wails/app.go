package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"pdfcreator-wails/internal/config"
	"pdfcreator-wails/internal/excel"
	"pdfcreator-wails/internal/pdfgen"
	"pdfcreator-wails/internal/project"
)

// App struct
type App struct {
	ctx     context.Context
	project *project.Project
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	projectHome := detectProjectHome()

	proj, err := project.New(projectHome)
	if err != nil {
		return
	}

	// Seed default fonts from ../fonts or ./fonts
	fontDir := filepath.Join(projectHome, "fonts")
	if _, err := os.Stat(fontDir); os.IsNotExist(err) {
		fontDir = filepath.Join(projectHome, "..", "fonts")
	}
	_ = proj.SeedFonts(fontDir)

	a.project = proj
}

// detectProjectHome finds the project root directory.
func detectProjectHome() string {
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		// If running from inside the Wails app bundle or build/bin, go up to project root.
		if strings.Contains(exeDir, "pdfcreator-wails.app") {
			return goUpUntil(exeDir, "pdfcreator-wails")
		}
		if strings.Contains(filepath.Base(exeDir), "pdfcreator-wails") {
			return filepath.Dir(exeDir)
		}
		return exeDir
	}

	wd, err := os.Getwd()
	if err == nil {
		if strings.Contains(filepath.Base(wd), "pdfcreator-wails") {
			return filepath.Dir(wd)
		}
		return wd
	}

	return "."
}

func goUpUntil(dir, target string) string {
	for {
		if filepath.Base(dir) == target {
			return filepath.Dir(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

// SelectTemplate sets the template image.
func (a *App) SelectTemplate(path string) error {
	return a.project.SetTemplate(path)
}

// SelectExcel sets the Excel file.
func (a *App) SelectExcel(path string) error {
	return a.project.SetExcel(path)
}

// SelectFont imports a font file.
func (a *App) SelectFont(path string) error {
	return a.project.AddFont(path)
}

// GetFonts returns available fonts.
func (a *App) GetFonts() []string {
	return a.project.GetFonts()
}

// GetColumns returns Excel column names.
func (a *App) GetColumns() ([]string, error) {
	if a.project.GetExcelPath() == "" {
		return []string{}, nil
	}
	reader, err := excel.Open(a.project.GetExcelPath())
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return reader.Columns()
}

// GetPreviewRows returns first N data rows.
func (a *App) GetPreviewRows(limit int) ([][]string, error) {
	if a.project.GetExcelPath() == "" {
		return [][]string{}, nil
	}
	reader, err := excel.Open(a.project.GetExcelPath())
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return reader.Preview(limit)
}

// SaveFields persists fields.
func (a *App) SaveFields(fields []config.Field) error {
	return a.project.SaveFields(fields)
}

// LoadFields returns saved fields.
func (a *App) LoadFields() []config.Field {
	return a.project.LoadFields()
}

// GeneratePreview creates a preview PDF and returns the fitted fields and path.
func (a *App) GeneratePreview() (map[string]interface{}, error) {
	if a.project.GetTemplatePath() == "" || a.project.GetExcelPath() == "" {
		return nil, fmt.Errorf("template and excel required")
	}

	reader, err := excel.Open(a.project.GetExcelPath())
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	columns, err := reader.Columns()
	if err != nil {
		return nil, err
	}

	rows, err := reader.Preview(1)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no data rows")
	}

	outputPath := filepath.Join(a.project.OutputDir(), "preview.pdf")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, err
	}

	gen := pdfgen.New(a.project.DataDir)
	fittedFields, warnings, err := gen.GeneratePreview(a.project.GetTemplatePath(), a.project.LoadFields(), rows[0], columns, outputPath)
	if err != nil {
		return nil, err
	}

	// Persist fitted sizes back to config.
	_ = a.project.SaveFields(fittedFields)

	return map[string]interface{}{
		"path":     outputPath,
		"fields":   fittedFields,
		"warnings": warnings,
	}, nil
}

// GenerateAll creates all PDFs and returns count, zip path and fitted fields.
func (a *App) GenerateAll(outputDir string) (map[string]interface{}, error) {
	if a.project.GetTemplatePath() == "" || a.project.GetExcelPath() == "" {
		return nil, fmt.Errorf("template and excel required")
	}

	reader, err := excel.Open(a.project.GetExcelPath())
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	columns, err := reader.Columns()
	if err != nil {
		return nil, err
	}

	rows, err := reader.AllRows()
	if err != nil {
		return nil, err
	}

	if outputDir == "" {
		outputDir = a.project.OutputDir()
	}

	if err := os.RemoveAll(outputDir); err != nil {
		return nil, err
	}

	gen := pdfgen.New(a.project.DataDir)
	result, err := gen.GenerateImageTemplate(a.project.GetTemplatePath(), a.project.LoadFields(), rows, columns, outputDir)
	if err != nil {
		return nil, err
	}

	zipPath := filepath.Join(outputDir, "diplomas.zip")
	if err := pdfgen.CreateZIP(result.Paths, zipPath); err != nil {
		return nil, err
	}

	// Persist fitted sizes back to config.
	_ = a.project.SaveFields(result.Fields)

	return map[string]interface{}{
		"count":    len(result.Paths),
		"zipPath":  zipPath,
		"dir":      outputDir,
		"fields":   result.Fields,
		"warnings": result.Warnings,
	}, nil
}

// GetSampleText returns the longest value for a column across all Excel rows.
func (a *App) GetSampleText(column string) (string, error) {
	if a.project.GetExcelPath() == "" {
		return "", fmt.Errorf("no excel selected")
	}
	reader, err := excel.Open(a.project.GetExcelPath())
	if err != nil {
		return "", err
	}
	defer reader.Close()

	columns, err := reader.Columns()
	if err != nil {
		return "", err
	}

	rows, err := reader.AllRows()
	if err != nil {
		return "", err
	}

	return pdfgen.LongestText(rows, columns, column), nil
}

// GetOutputDir returns the default output directory.
func (a *App) GetOutputDir() string {
	return a.project.OutputDir()
}

// GetDataDir returns the data directory.
func (a *App) GetDataDir() string {
	return a.project.DataDir
}

// GetTemplatePath returns current template path.
func (a *App) GetTemplatePath() string {
	return a.project.GetTemplatePath()
}

// GetExcelPath returns current excel path.
func (a *App) GetExcelPath() string {
	return a.project.GetExcelPath()
}

// GetTemplateBase64 returns the template image as a base64 data URL.
func (a *App) GetTemplateBase64() (string, error) {
	path := a.project.GetTemplatePath()
	if path == "" {
		return "", fmt.Errorf("no template selected")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(path))
	mime := "image/jpeg"
	if ext == ".png" {
		mime = "image/png"
	}

	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data)), nil
}

// OpenFileDialog shows a file open dialog.
func (a *App) OpenFileDialog(opts wailsruntime.OpenDialogOptions) (string, error) {
	return wailsruntime.OpenFileDialog(a.ctx, opts)
}

// OpenDirectoryDialog shows a directory open dialog.
func (a *App) OpenDirectoryDialog(opts wailsruntime.OpenDialogOptions) (string, error) {
	return wailsruntime.OpenDirectoryDialog(a.ctx, opts)
}

// OpenFile opens a file with the system default application.
func (a *App) OpenFile(path string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{path}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", "", path}
	default:
		cmd = "xdg-open"
		args = []string{path}
	}
	return exec.Command(cmd, args...).Start()
}
