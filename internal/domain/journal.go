package domain

import "time"

type JournalStatus string

const (
	JournalStatusInProgress JournalStatus = "in_progress"
	JournalStatusCompleted  JournalStatus = "completed"
	JournalStatusFailed     JournalStatus = "failed"
)

type FileOp string

const (
	FileOpMove FileOp = "move"
	FileOpCopy FileOp = "copy"
)

type EntryStatus string

const (
	EntryStatusPending   EntryStatus = "pending"
	EntryStatusCompleted EntryStatus = "completed"
	EntryStatusFailed    EntryStatus = "failed"
	EntryStatusSkipped   EntryStatus = "skipped"
)

type Journal struct {
	Version   int
	Status    JournalStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	Entries   []JournalEntry
}

type JournalEntry struct {
	ID          int
	Operation   FileOp
	Source      string
	Destination string
	Size        int64
	Status      EntryStatus
	Error       string
}

func NewJournal() *Journal {
	now := time.Now()
	return &Journal{
		Version:   1,
		Status:    JournalStatusInProgress,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (j *Journal) AddEntry(op FileOp, source, destination string, size int64) int {
	id := len(j.Entries)
	j.Entries = append(j.Entries, JournalEntry{
		ID:          id,
		Operation:   op,
		Source:      source,
		Destination: destination,
		Size:        size,
		Status:      EntryStatusPending,
	})
	j.UpdatedAt = time.Now()
	return id
}

func (j *Journal) MarkEntryCompleted(id int) {
	if id >= 0 && id < len(j.Entries) {
		j.Entries[id].Status = EntryStatusCompleted
		j.UpdatedAt = time.Now()
	}
}

func (j *Journal) MarkEntryFailed(id int, err string) {
	if id >= 0 && id < len(j.Entries) {
		j.Entries[id].Status = EntryStatusFailed
		j.Entries[id].Error = err
		j.UpdatedAt = time.Now()
	}
}

func (j *Journal) PendingEntries() []JournalEntry {
	var pending []JournalEntry
	for _, e := range j.Entries {
		if e.Status == EntryStatusPending {
			pending = append(pending, e)
		}
	}
	return pending
}

func (j *Journal) Complete() {
	j.Status = JournalStatusCompleted
	j.UpdatedAt = time.Now()
}

func (j *Journal) Fail() {
	j.Status = JournalStatusFailed
	j.UpdatedAt = time.Now()
}
