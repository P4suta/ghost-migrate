package fsjournal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ghost-migrate/internal/domain"
)

type Store struct {
	path string
}

func NewStore(outputDir string) *Store {
	return &Store{path: filepath.Join(outputDir, "manifest.json")}
}

func (s *Store) Load() (*domain.Journal, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read journal: %w", err)
	}

	var j domain.Journal
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, fmt.Errorf("failed to parse journal: %w", err)
	}
	return &j, nil
}

func (s *Store) Save(j *domain.Journal) error {
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal journal: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create journal directory: %w", err)
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write journal temp file: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("failed to rename journal file: %w", err)
	}
	return nil
}

func (s *Store) Remove() error {
	err := os.Remove(s.path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove journal: %w", err)
	}
	return nil
}
