package server

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pdfcreator-wails/internal/config"
	"pdfcreator-wails/internal/excel"
	"pdfcreator-wails/internal/pdfgen"
	"pdfcreator-wails/web"
)

// Server holds HTTP handlers and dependencies.
type Server struct {
	dataDir   string
	cfg       *config.Config
	generator *pdfgen.Generator
	mux       *http.ServeMux
}

// New creates a new Server.
func New(dataDir string, cfg *config.Config) *Server {
	s := &Server{
		dataDir:   dataDir,
		cfg:       cfg,
		generator: pdfgen.New(dataDir),
		mux:       http.NewServeMux(),
	}
	s.routes()
	return s
}

// Handler returns the http.Handler.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/upload/template", s.handleUploadTemplate)
	s.mux.HandleFunc("/api/upload/excel", s.handleUploadExcel)
	s.mux.HandleFunc("/api/upload/font", s.handleUploadFont)
	s.mux.HandleFunc("/api/fonts", s.handleFonts)
	s.mux.HandleFunc("/api/template", s.handleTemplate)
	s.mux.HandleFunc("/api/excel/preview", s.handleExcelPreview)
	s.mux.HandleFunc("/api/excel/columns", s.handleExcelColumns)
	s.mux.HandleFunc("/api/fields", s.handleFields)
	s.mux.HandleFunc("/api/preview", s.handlePreview)
	s.mux.HandleFunc("/api/preview/file", s.handlePreviewFile)
	s.mux.HandleFunc("/api/generate", s.handleGenerate)
	s.mux.HandleFunc("/api/download", s.handleDownload)

	fs := http.FileServer(http.FS(web.Files))
	s.mux.Handle("/", fs)
}

func (s *Server) handleUploadTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		http.Error(w, "only JPG/PNG images are supported", http.StatusBadRequest)
		return
	}

	dir := filepath.Join(s.dataDir, "templates")
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	path := filepath.Join(dir, "template"+ext)
	out, err := os.Create(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.cfg.TemplatePath = path
	_ = config.Save(s.cfg)

	s.respondJSON(w, map[string]string{"path": path, "name": header.Filename})
}

func (s *Server) handleUploadExcel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dir := filepath.Join(s.dataDir, "uploads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	path := filepath.Join(dir, header.Filename)
	out, err := os.Create(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.cfg.ExcelPath = path
	_ = config.Save(s.cfg)

	s.respondJSON(w, map[string]string{"path": path, "name": header.Filename})
}

func (s *Server) handleUploadFont(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".ttf" && ext != ".otf" {
		http.Error(w, "only TTF/OTF fonts supported", http.StatusBadRequest)
		return
	}

	dir := filepath.Join(s.dataDir, "fonts")
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	path := filepath.Join(dir, header.Filename)
	out, err := os.Create(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]string{"path": path, "name": header.Filename})
}

func (s *Server) handleFonts(w http.ResponseWriter, r *http.Request) {
	dir := filepath.Join(s.dataDir, "fonts")
	files, err := os.ReadDir(dir)
	if err != nil {
		s.respondJSON(w, []string{})
		return
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
	s.respondJSON(w, fonts)
}

func (s *Server) handleTemplate(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, map[string]string{"path": s.cfg.TemplatePath})
}

func (s *Server) handleExcelPreview(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ExcelPath == "" {
		s.respondJSON(w, map[string]interface{}{"columns": []string{}, "rows": [][]string{}})
		return
	}

	reader, err := excel.Open(s.cfg.ExcelPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	columns, err := reader.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := reader.Preview(10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]interface{}{
		"columns": columns,
		"rows":    rows,
	})
}

func (s *Server) handleExcelColumns(w http.ResponseWriter, r *http.Request) {
	if s.cfg.ExcelPath == "" {
		s.respondJSON(w, []string{})
		return
	}

	reader, err := excel.Open(s.cfg.ExcelPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	columns, err := reader.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.respondJSON(w, columns)
}

func (s *Server) handleFields(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.respondJSON(w, s.cfg.Fields)
		return
	}

	if r.Method == http.MethodPost {
		var fields []config.Field
		if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.cfg.Fields = fields
		if err := config.Save(s.cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.respondJSON(w, s.cfg.Fields)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.cfg.TemplatePath == "" || s.cfg.ExcelPath == "" {
		http.Error(w, "template and excel required", http.StatusBadRequest)
		return
	}

	reader, err := excel.Open(s.cfg.ExcelPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	columns, err := reader.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := reader.Preview(1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(rows) == 0 {
		http.Error(w, "no data rows", http.StatusBadRequest)
		return
	}

	outputPath := filepath.Join(s.dataDir, "output", "preview.pdf")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.generator.GeneratePreview(s.cfg.TemplatePath, s.cfg.Fields, rows[0], columns, outputPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]string{"previewUrl": "/api/preview/file?t=" + fmt.Sprintf("%d", os.Getpid())})
}

func (s *Server) handlePreviewFile(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.dataDir, "output", "preview.pdf")
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=preview.pdf")
	http.ServeFile(w, r, path)
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.cfg.TemplatePath == "" || s.cfg.ExcelPath == "" {
		http.Error(w, "template and excel required", http.StatusBadRequest)
		return
	}

	reader, err := excel.Open(s.cfg.ExcelPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	columns, err := reader.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := reader.AllRows()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	outputDir := filepath.Join(s.dataDir, "output")
	if err := os.RemoveAll(outputDir); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	generated, err := s.generator.GenerateImageTemplate(s.cfg.TemplatePath, s.cfg.Fields, rows, columns, outputDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	zipPath := filepath.Join(s.dataDir, "output", "diplomas.zip")
	if err := pdfgen.CreateZIP(generated, zipPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]interface{}{
		"count":   len(generated),
		"zipUrl":  "/api/download",
		"files":   generated,
	})
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.dataDir, "output", "diplomas.zip")
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=diplomas.zip")
	http.ServeFile(w, r, path)
}

func (s *Server) respondJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func isZipEmpty(path string) bool {
	r, err := zip.OpenReader(path)
	if err != nil {
		return true
	}
	defer r.Close()
	return len(r.File) == 0
}
