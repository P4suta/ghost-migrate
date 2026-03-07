package port

import (
	"ghost-migrate/internal/domain"
	"io"
)

type MediaStore interface {
	List() ([]domain.MediaFile, error)
	Has(contentPath string) bool
	Open(contentPath string) (io.ReadCloser, error)
	Close() error
}
