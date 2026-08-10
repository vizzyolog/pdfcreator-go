package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Align represents text alignment.
type Align string

const (
	AlignLeft   Align = "L"
	AlignCenter Align = "C"
	AlignRight  Align = "R"
)

// Field describes a single text field mapped to an Excel column.
type Field struct {
	Column     string  `json:"column"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Font       string  `json:"font"`
	FontSize   int     `json:"fontSize"`
	Color      string  `json:"color"`
	Align      Align   `json:"align"`
	MaxWidth   float64 `json:"maxWidth"`
	AutoWrap   bool    `json:"autoWrap"`
	WrapPattern []int  `json:"-"`
}

// Config stores all application settings.
type Config struct {
	TemplatePath string  `json:"templatePath"`
	ExcelPath    string  `json:"excelPath"`
	Fields       []Field `json:"fields"`
}

var (
	cfg  Config
	mu   sync.RWMutex
	path string
)

// SetPath sets the file path used for Load/Save.
func SetPath(p string) {
	path = p
}

// Load reads config from disk or returns a default empty config.
func Load() (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	cfg = Config{Fields: []Field{}}

	if path == "" {
		return &cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save persists the current config to disk.
func Save(c *Config) error {
	mu.Lock()
	defer mu.Unlock()

	cfg = *c

	if path == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Get returns the current config copy.
func Get() Config {
	mu.RLock()
	defer mu.RUnlock()
	return cfg
}
