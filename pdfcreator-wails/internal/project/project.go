package project

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pdfcreator-wails/internal/config"
)

// Project holds the current session state.
type Project struct {
	Home    string
	DataDir string
	Config  *config.Config
}

// New creates a new Project in the given home directory.
func New(home string) (*Project, error) {
	dataDir := filepath.Join(home, "data")
	outputDir := filepath.Join(home, "output")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	config.SetPath(filepath.Join(dataDir, "config.json"))
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	return &Project{
		Home:    home,
		DataDir: dataDir,
		Config:  cfg,
	}, nil
}

// SeedFonts copies default fonts from sourceDir into data/fonts.
func (p *Project) SeedFonts(sourceDir string) error {
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return nil
	}

	destDir := filepath.Join(p.DataDir, "fonts")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".ttf" && ext != ".otf" {
			continue
		}

		srcPath := filepath.Join(sourceDir, entry.Name())
		dstPath := filepath.Join(destDir, entry.Name())
		if _, err := os.Stat(dstPath); err == nil {
			continue
		}

		src, err := os.Open(srcPath)
		if err != nil {
			continue
		}
		dst, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			continue
		}
		_, _ = io.Copy(dst, src)
		src.Close()
		dst.Close()
	}

	return nil
}

// SetTemplate copies template to data/templates and updates config.
func (p *Project) SetTemplate(sourcePath string) error {
	dir := filepath.Join(p.DataDir, "templates")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(sourcePath))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return fmt.Errorf("only JPG/PNG templates supported")
	}

	dst := filepath.Join(dir, "template"+ext)
	if err := copyFile(sourcePath, dst); err != nil {
		return err
	}

	p.Config.TemplatePath = dst
	return config.Save(p.Config)
}

// SetExcel copies excel to data/uploads and updates config.
func (p *Project) SetExcel(sourcePath string) error {
	dir := filepath.Join(p.DataDir, "uploads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	dst := filepath.Join(dir, filepath.Base(sourcePath))
	if err := copyFile(sourcePath, dst); err != nil {
		return err
	}

	p.Config.ExcelPath = dst
	return config.Save(p.Config)
}

// AddFont copies a font into data/fonts.
func (p *Project) AddFont(sourcePath string) error {
	dir := filepath.Join(p.DataDir, "fonts")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(sourcePath))
	if ext != ".ttf" && ext != ".otf" {
		return fmt.Errorf("only TTF/OTF fonts supported")
	}

	dst := filepath.Join(dir, filepath.Base(sourcePath))
	return copyFile(sourcePath, dst)
}

// GetFonts returns the list of available font files.
func (p *Project) GetFonts() []string {
	dir := filepath.Join(p.DataDir, "fonts")
	files, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}

	fonts := []string{}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext == ".ttf" || ext == ".otf" {
			fonts = append(fonts, f.Name())
		}
	}
	return fonts
}

// SaveFields persists field configuration.
func (p *Project) SaveFields(fields []config.Field) error {
	p.Config.Fields = fields
	return config.Save(p.Config)
}

// LoadFields returns field configuration.
func (p *Project) LoadFields() []config.Field {
	return p.Config.Fields
}

// GetTemplatePath returns current template path.
func (p *Project) GetTemplatePath() string {
	return p.Config.TemplatePath
}

// GetExcelPath returns current excel path.
func (p *Project) GetExcelPath() string {
	return p.Config.ExcelPath
}

// OutputDir returns the output directory path.
func (p *Project) OutputDir() string {
	return filepath.Join(p.Home, "output")
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}
