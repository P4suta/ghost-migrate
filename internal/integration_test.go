package internal

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ghost-migrate/internal/adapter/fsbundle"
	"ghost-migrate/internal/adapter/fsjournal"
	"ghost-migrate/internal/adapter/ghostjson"
	"ghost-migrate/internal/adapter/htmlconv"
	"ghost-migrate/internal/adapter/mediazip"
	"ghost-migrate/internal/usecase"
)

func TestIntegration_BundleMode(t *testing.T) {
	// Create a test ZIP with media files
	zipPath := createMediaZip(t)

	outputDir := t.TempDir()

	reader := ghostjson.NewReader()
	converter := htmlconv.NewConverter()
	mediaStore, err := mediazip.NewStore(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer mediaStore.Close()

	writer := fsbundle.NewWriter(outputDir)
	journalStore := fsjournal.NewStore(outputDir)

	uc := usecase.NewMigrateUseCase(reader, converter, mediaStore, writer, journalStore, nil)

	plan, err := uc.Execute(usecase.MigrateOptions{
		InputPath:    "../testdata/fixtures/with-media.json",
		OutputDir:    outputDir,
		IncludePages: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify plan stats
	if plan.Stats.TotalPosts != 3 {
		t.Errorf("TotalPosts = %d, want 3", plan.Stats.TotalPosts)
	}

	// Verify bundle directories exist
	assertFileExists(t, filepath.Join(outputDir, "my-travel-post", "index.md"))
	assertFileExists(t, filepath.Join(outputDir, "my-travel-post", "photo.jpg"))
	assertFileExists(t, filepath.Join(outputDir, "my-travel-post", "cover.jpg"))
	assertFileExists(t, filepath.Join(outputDir, "second-post", "index.md"))
	assertFileExists(t, filepath.Join(outputDir, "second-post", "photo.jpg"))
	assertFileExists(t, filepath.Join(outputDir, "draft-post", "index.md"))

	// Verify index.md contains rewritten URLs (no __GHOST_URL__)
	indexContent := readFile(t, filepath.Join(outputDir, "my-travel-post", "index.md"))
	if strings.Contains(indexContent, "__GHOST_URL__") {
		t.Error("index.md still contains __GHOST_URL__")
	}
	if !strings.Contains(indexContent, "photo.jpg") {
		t.Error("index.md should reference local photo.jpg")
	}
	if !strings.Contains(indexContent, "featured_image: cover.jpg") {
		t.Error("index.md should have local featured_image")
	}

	// Verify orphaned media goes to _orphaned/
	assertFileExists(t, filepath.Join(outputDir, "_orphaned", "orphan.jpg"))

	// Verify media content
	photoContent := readFile(t, filepath.Join(outputDir, "my-travel-post", "photo.jpg"))
	if photoContent != "photo data" {
		t.Errorf("photo content = %q", photoContent)
	}

	// Verify journal
	assertFileExists(t, filepath.Join(outputDir, "manifest.json"))

	// Verify shared media is in both bundles
	assertFileExists(t, filepath.Join(outputDir, "second-post", "photo.jpg"))
}

func TestIntegration_FlatMode(t *testing.T) {
	outputDir := t.TempDir()

	reader := ghostjson.NewReader()
	converter := htmlconv.NewConverter()
	writer := fsbundle.NewWriter(outputDir)

	uc := usecase.NewMigrateUseCase(reader, converter, nil, writer, nil, nil)

	plan, err := uc.Execute(usecase.MigrateOptions{
		InputPath:    "../testdata/fixtures/with-media.json",
		OutputDir:    outputDir,
		IncludePages: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if plan.Stats.TotalPosts != 3 {
		t.Errorf("TotalPosts = %d, want 3", plan.Stats.TotalPosts)
	}

	// Flat mode: files should be slug.md, not slug/index.md
	assertFileExists(t, filepath.Join(outputDir, "my-travel-post.md"))
	assertFileExists(t, filepath.Join(outputDir, "second-post.md"))
	assertFileExists(t, filepath.Join(outputDir, "draft-post.md"))

	// Verify draft has draft: true in front matter
	draftContent := readFile(t, filepath.Join(outputDir, "draft-post.md"))
	if !strings.Contains(draftContent, "draft: true") {
		t.Error("draft post should have draft: true")
	}
}

func TestIntegration_StatusFilter(t *testing.T) {
	outputDir := t.TempDir()

	reader := ghostjson.NewReader()
	converter := htmlconv.NewConverter()
	writer := fsbundle.NewWriter(outputDir)

	uc := usecase.NewMigrateUseCase(reader, converter, nil, writer, nil, nil)

	plan, err := uc.Execute(usecase.MigrateOptions{
		InputPath:    "../testdata/fixtures/with-media.json",
		OutputDir:    outputDir,
		IncludePages: true,
		Status:       "published",
	})
	if err != nil {
		t.Fatal(err)
	}

	if plan.Stats.TotalPosts != 2 {
		t.Errorf("TotalPosts = %d, want 2 (published only)", plan.Stats.TotalPosts)
	}

	assertFileExists(t, filepath.Join(outputDir, "my-travel-post.md"))
	assertFileExists(t, filepath.Join(outputDir, "second-post.md"))
	assertFileNotExists(t, filepath.Join(outputDir, "draft-post.md"))
}

// --- Helpers ---

func createMediaZip(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "media.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	files := map[string]string{
		"content/images/2024/09/photo.jpg": "photo data",
		"content/images/2024/09/cover.jpg": "cover data",
		"content/files/2024/doc.pdf":       "pdf data",
		"content/images/orphan.jpg":        "orphan data",
	}
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

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file to NOT exist: %s", path)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}
