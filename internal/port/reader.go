package port

import (
	"ghost-migrate/internal/domain"
	"io"
)

type ExportReader interface {
	Read(r io.Reader) (domain.RawExport, error)
}
