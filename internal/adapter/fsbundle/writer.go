package fsbundle

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"ghost-migrate/internal/domain"
)

type Writer struct {
	outputDir string
}

func NewWriter(outputDir string) *Writer {
	return &Writer{outputDir: outputDir}
}

func (w *Writer) WriteIndex(slug, content string) error {
	dir := filepath.Join(w.outputDir, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create bundle directory: %w", err)
	}
	path := filepath.Join(dir, "index.md")
	return os.WriteFile(path, []byte(content), 0o644)
}

func (w *Writer) WriteMedia(slug, destFilename string, src io.Reader) error {
	dir := filepath.Join(w.outputDir, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create bundle directory: %w", err)
	}
	path := filepath.Join(dir, destFilename)
	return writeFile(path, src)
}

func (w *Writer) WriteOrphan(destFilename string, src io.Reader) error {
	dir := filepath.Join(w.outputDir, "_orphaned")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create orphan directory: %w", err)
	}
	path := filepath.Join(dir, destFilename)
	return writeFile(path, src)
}

func (w *Writer) WriteFlat(article domain.Article) error {
	if err := os.MkdirAll(w.outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	content := article.FrontMatter.Marshal() + "\n" + article.Content
	path := filepath.Join(w.outputDir, article.Filename)
	return os.WriteFile(path, []byte(content), 0o644)
}

func writeFile(path string, src io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, src); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
