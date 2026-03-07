package mediazip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestStore_ListHasOpen(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"content/images/2024/photo.jpg": "image data",
		"content/media/video.mp4":       "video data",
	})

	store, err := NewStore(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	files, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("List() = %d files, want 2", len(files))
	}

	if !store.Has("content/images/2024/photo.jpg") {
		t.Error("Has() = false for existing file")
	}
	if store.Has("content/images/missing.jpg") {
		t.Error("Has() = true for missing file")
	}

	rc, err := store.Open("content/images/2024/photo.jpg")
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
	zipPath := createTestZip(t, map[string]string{
		"japan-travel_123/content/images/photo.jpg":                 "nested once",
		"japan-travel_123/japan-travel_123/content/images/deep.jpg": "nested twice",
	})

	store, err := NewStore(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if !store.Has("content/images/photo.jpg") {
		t.Error("single-nested not found")
	}
	if !store.Has("content/images/deep.jpg") {
		t.Error("double-nested not found")
	}
}

func TestStore_PathTraversalRejected(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"../../../etc/passwd":     "malicious",
		"content/images/safe.jpg": "safe data",
	})

	store, err := NewStore(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	files, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("List() = %d files, want 1 (path traversal should be rejected)", len(files))
	}
}

func TestStore_OpenMissing(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"content/images/photo.jpg": "data",
	})

	store, err := NewStore(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	_, err = store.Open("content/images/missing.jpg")
	if err == nil {
		t.Error("Open() should error for missing file")
	}
}
