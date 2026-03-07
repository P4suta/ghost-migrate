package fsjournal

import (
	"os"
	"path/filepath"
	"testing"

	"ghost-migrate/internal/domain"
)

func TestStore_SaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	j := domain.NewJournal()
	j.AddEntry(domain.FileOpCopy, "src/a.jpg", "dst/a.jpg", 1000)
	j.AddEntry(domain.FileOpMove, "src/b.jpg", "dst/b.jpg", 2000)
	j.MarkEntryCompleted(0)

	if err := store.Save(j); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err != nil {
		t.Fatal("manifest.json not created")
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil {
		t.Fatal("Load() returned nil")
	}

	if loaded.Version != 1 {
		t.Errorf("Version = %d", loaded.Version)
	}
	if len(loaded.Entries) != 2 {
		t.Fatalf("Entries = %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Status != domain.EntryStatusCompleted {
		t.Errorf("entry 0 status = %q", loaded.Entries[0].Status)
	}
	if loaded.Entries[1].Status != domain.EntryStatusPending {
		t.Errorf("entry 1 status = %q", loaded.Entries[1].Status)
	}
}

func TestStore_LoadMissing(t *testing.T) {
	store := NewStore(t.TempDir())

	j, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if j != nil {
		t.Error("Load() should return nil for missing file")
	}
}

func TestStore_Remove(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	j := domain.NewJournal()
	if err := store.Save(j); err != nil {
		t.Fatal(err)
	}

	if err := store.Remove(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "manifest.json")); !os.IsNotExist(err) {
		t.Error("manifest.json should be removed")
	}

	// Remove on already-removed should not error
	if err := store.Remove(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	j := domain.NewJournal()
	if err := store.Save(j); err != nil {
		t.Fatal(err)
	}

	// Verify no .tmp file remains
	if _, err := os.Stat(filepath.Join(dir, "manifest.json.tmp")); !os.IsNotExist(err) {
		t.Error("temp file should not remain after save")
	}
}
