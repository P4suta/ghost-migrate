package port

import (
	"ghost-migrate/internal/domain"
	"io"
)

type BundleWriter interface {
	WriteIndex(slug, content string) error
	WriteMedia(slug, destFilename string, src io.Reader) error
	WriteOrphan(destFilename string, src io.Reader) error
	WriteFlat(article domain.Article) error
}
