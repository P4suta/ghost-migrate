package mediazip

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"

	"ghost-migrate/internal/domain"
)

type Store struct {
	reader *zip.ReadCloser
	index  map[string]*zip.File
}

func NewStore(zipPath string) (*Store, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}

	index := make(map[string]*zip.File, len(r.File))
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		// Reject path traversal
		if strings.Contains(f.Name, "..") {
			continue
		}
		contentPath := domain.NormalizeContentPath(f.Name)
		if strings.HasPrefix(contentPath, "content/") {
			index[contentPath] = f
		}
	}

	return &Store{reader: r, index: index}, nil
}

func (s *Store) List() ([]domain.MediaFile, error) {
	files := make([]domain.MediaFile, 0, len(s.index))
	for path, f := range s.index {
		files = append(files, domain.MediaFile{
			Path: path,
			Size: int64(f.UncompressedSize64),
		})
	}
	return files, nil
}

func (s *Store) Has(contentPath string) bool {
	_, ok := s.index[contentPath]
	return ok
}

func (s *Store) Open(contentPath string) (io.ReadCloser, error) {
	f, ok := s.index[contentPath]
	if !ok {
		return nil, fmt.Errorf("media file not found in ZIP: %s", contentPath)
	}
	return f.Open()
}

func (s *Store) Close() error {
	return s.reader.Close()
}
