package domain

import "testing"

func TestJournal_NewJournal(t *testing.T) {
	j := NewJournal()
	if j.Version != 1 {
		t.Errorf("Version = %d, want 1", j.Version)
	}
	if j.Status != JournalStatusInProgress {
		t.Errorf("Status = %q, want %q", j.Status, JournalStatusInProgress)
	}
	if len(j.Entries) != 0 {
		t.Errorf("Entries = %d, want 0", len(j.Entries))
	}
}

func TestJournal_AddEntry(t *testing.T) {
	j := NewJournal()
	id := j.AddEntry(FileOpCopy, "src/a.jpg", "dst/a.jpg", 1000)

	if id != 0 {
		t.Errorf("first ID = %d, want 0", id)
	}
	if len(j.Entries) != 1 {
		t.Fatalf("Entries = %d, want 1", len(j.Entries))
	}
	e := j.Entries[0]
	if e.Operation != FileOpCopy {
		t.Errorf("Operation = %q", e.Operation)
	}
	if e.Status != EntryStatusPending {
		t.Errorf("Status = %q", e.Status)
	}
}

func TestJournal_StateTransitions(t *testing.T) {
	j := NewJournal()
	id0 := j.AddEntry(FileOpMove, "a", "b", 100)
	id1 := j.AddEntry(FileOpCopy, "c", "d", 200)
	id2 := j.AddEntry(FileOpCopy, "e", "f", 300)

	j.MarkEntryCompleted(id0)
	j.MarkEntryFailed(id1, "disk full")

	if j.Entries[id0].Status != EntryStatusCompleted {
		t.Errorf("entry 0 status = %q", j.Entries[id0].Status)
	}
	if j.Entries[id1].Status != EntryStatusFailed {
		t.Errorf("entry 1 status = %q", j.Entries[id1].Status)
	}
	if j.Entries[id1].Error != "disk full" {
		t.Errorf("entry 1 error = %q", j.Entries[id1].Error)
	}

	pending := j.PendingEntries()
	if len(pending) != 1 {
		t.Fatalf("pending = %d, want 1", len(pending))
	}
	if pending[0].ID != id2 {
		t.Errorf("pending[0].ID = %d, want %d", pending[0].ID, id2)
	}
}

func TestJournal_Complete(t *testing.T) {
	j := NewJournal()
	j.Complete()
	if j.Status != JournalStatusCompleted {
		t.Errorf("Status = %q, want %q", j.Status, JournalStatusCompleted)
	}
}

func TestJournal_Fail(t *testing.T) {
	j := NewJournal()
	j.Fail()
	if j.Status != JournalStatusFailed {
		t.Errorf("Status = %q, want %q", j.Status, JournalStatusFailed)
	}
}
