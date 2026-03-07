package port

import "ghost-migrate/internal/domain"

type JournalStore interface {
	Load() (*domain.Journal, error)
	Save(j *domain.Journal) error
	Remove() error
}
