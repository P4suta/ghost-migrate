package mediafs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"ghost-migrate/internal/domain"
)

type Store struct {
	basePath string
	index    map[string]domain.MediaFile
}

func NewStore(dirPath string) (*Store, error) {
	index := make(map[string]domain.MediaFile)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}
		contentPath := domain.NormalizeContentPath(filepath.ToSlash(rel))

		index[contentPath] = domain.MediaFile{
			Path: contentPath,
			Size: info.Size(),
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk media directory: %w", err)
	}

	return &Store{basePath: dirPath, index: index}, nil
}

func (s *Store) List() ([]domain.MediaFile, error) {
	files := make([]domain.MediaFile, 0, len(s.index))
	for _, f := range s.index {
		files = append(files, f)
	}
	return files, nil
}

func (s *Store) Has(contentPath string) bool {
	_, ok := s.index[contentPath]
	return ok
}

func (s *Store) Open(contentPath string) (io.ReadCloser, error) {
	if !s.Has(contentPath) {
		return nil, fmt.Errorf("media file not found: %s", contentPath)
	}

	// Find actual file path by walking possible prefixed locations
	var fullPath string
	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(s.basePath, path)
		if domain.NormalizeContentPath(filepath.ToSlash(rel)) == contentPath {
			fullPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if fullPath == "" {
		return nil, fmt.Errorf("media file not found on disk: %s", contentPath)
	}

	return os.Open(fullPath)
}

func (s *Store) Close() error {
	return nil
}
