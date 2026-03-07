package mediafs

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestStore_ListHasOpen(t *testing.T) {
	dir := t.TempDir()

	// Create content/images/2024/09/photo.jpg
	imgDir := filepath.Join(dir, "content", "images", "2024", "09")
	if err := os.MkdirAll(imgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imgDir, "photo.jpg"), []byte("image data"), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	files, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("List() = %d files, want 1", len(files))
	}
	if files[0].Path != "content/images/2024/09/photo.jpg" {
		t.Errorf("Path = %q", files[0].Path)
	}

	if !store.Has("content/images/2024/09/photo.jpg") {
		t.Error("Has() = false for existing file")
	}
	if store.Has("content/images/missing.jpg") {
		t.Error("Has() = true for missing file")
	}

	rc, err := store.Open("content/images/2024/09/photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "image data" {
		t.Errorf("content = %q", string(data))
	}
}

func TestStore_NestedPrefix(t *testing.T) {
	dir := t.TempDir()

	// Simulate nested ZIP extraction: prefix/content/images/photo.jpg
	imgDir := filepath.Join(dir, "backup-123", "content", "images")
	if err := os.MkdirAll(imgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imgDir, "photo.jpg"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	store, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if !store.Has("content/images/photo.jpg") {
		t.Error("Has() should normalize nested prefix path")
	}

	rc, err := store.Open("content/images/photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	rc.Close()
}
