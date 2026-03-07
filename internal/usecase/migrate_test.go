package usecase

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ghost-migrate/internal/domain"
	"ghost-migrate/internal/port"
)

// --- Mock implementations ---

type mockReader struct {
	export domain.RawExport
}

func (m *mockReader) Read(r io.Reader) (domain.RawExport, error) {
	return m.export, nil
}

type mockConverter struct {
	convertFn func(html string, rewriter port.URLRewriter) (string, error)
}

func (m *mockConverter) Convert(html string, rewriter port.URLRewriter) (string, error) {
	if m.convertFn != nil {
		return m.convertFn(html, rewriter)
	}
	return html, nil
}

type mockMediaStore struct {
	files map[string]string // contentPath → content
}

func (m *mockMediaStore) List() ([]domain.MediaFile, error) {
	var files []domain.MediaFile
	for p, content := range m.files {
		files = append(files, domain.MediaFile{Path: p, Size: int64(len(content))})
	}
	return files, nil
}

func (m *mockMediaStore) Has(contentPath string) bool {
	_, ok := m.files[contentPath]
	return ok
}

func (m *mockMediaStore) Open(contentPath string) (io.ReadCloser, error) {
	content, ok := m.files[contentPath]
	if !ok {
		return nil, fmt.Errorf("not found: %s", contentPath)
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

func (m *mockMediaStore) Close() error { return nil }

type writtenFile struct {
	path    string
	content string
}

type mockWriter struct {
	written []writtenFile
}

func (m *mockWriter) WriteIndex(slug, content string) error {
	m.written = append(m.written, writtenFile{slug + "/index.md", content})
	return nil
}

func (m *mockWriter) WriteMedia(slug, destFilename string, src io.Reader) error {
	data, _ := io.ReadAll(src)
	m.written = append(m.written, writtenFile{slug + "/" + destFilename, string(data)})
	return nil
}

func (m *mockWriter) WriteOrphan(destFilename string, src io.Reader) error {
	data, _ := io.ReadAll(src)
	m.written = append(m.written, writtenFile{"_orphaned/" + destFilename, string(data)})
	return nil
}

func (m *mockWriter) WriteFlat(article domain.Article) error {
	content := article.FrontMatter.Marshal() + "\n" + article.Content
	m.written = append(m.written, writtenFile{article.Filename, content})
	return nil
}

type mockJournalStore struct {
	saved *domain.Journal
}

func (m *mockJournalStore) Load() (*domain.Journal, error) { return nil, nil }
func (m *mockJournalStore) Save(j *domain.Journal) error {
	m.saved = j
	return nil
}
func (m *mockJournalStore) Remove() error { return nil }

type mockReporter struct {
	posts []string
	stats *MigrateStats
}

func (m *mockReporter) OnPostProcessed(slug string, mediaCount int) {
	m.posts = append(m.posts, slug)
}

func (m *mockReporter) OnComplete(stats MigrateStats) {
	m.stats = &stats
}

// --- Test helpers ---

func createTestExportFile(t *testing.T, export domain.RawExport) string {
	t.Helper()
	// Create a minimal Ghost JSON structure
	type jsonExport struct {
		DB []struct {
			Meta struct{ Version string } `json:"meta"`
			Data struct {
				Posts []struct {
					ID   string `json:"id"`
					Slug string `json:"slug"`
				} `json:"posts"`
			} `json:"data"`
		} `json:"db"`
	}

	// Just create a dummy JSON file - the mock reader ignores it
	dir := t.TempDir()
	path := filepath.Join(dir, "export.json")
	data, _ := json.Marshal(jsonExport{})
	os.WriteFile(path, data, 0o644)
	return path
}

func makePost(slug, html, featureImage string) domain.Post {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return domain.Post{
		ID:           slug,
		Title:        slug,
		Slug:         slug,
		HTML:         html,
		Status:       domain.StatusPublished,
		Type:         domain.TypePost,
		Visibility:   "public",
		FeatureImage: featureImage,
		CreatedAt:    now,
		UpdatedAt:    now,
		PublishedAt:  &now,
	}
}

// --- Tests ---

func TestFlatMode(t *testing.T) {
	post := makePost("hello", "<p>Hello</p>", "")
	reader := &mockReader{export: domain.RawExport{
		Posts: []domain.RawPost{{Post: post}},
	}}
	writer := &mockWriter{}
	reporter := &mockReporter{}

	exportPath := createTestExportFile(t, reader.export)

	uc := NewMigrateUseCase(reader, &mockConverter{}, nil, writer, nil, reporter)
	plan, err := uc.Execute(MigrateOptions{
		InputPath:    exportPath,
		IncludePages: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if plan.Stats.TotalPosts != 1 {
		t.Errorf("TotalPosts = %d", plan.Stats.TotalPosts)
	}
	if len(writer.written) != 1 {
		t.Fatalf("written = %d files", len(writer.written))
	}
	if writer.written[0].path != "hello.md" {
		t.Errorf("path = %q", writer.written[0].path)
	}
	if len(reporter.posts) != 1 || reporter.posts[0] != "hello" {
		t.Errorf("reported posts = %v", reporter.posts)
	}
}

func TestBundleMode(t *testing.T) {
	post := makePost("my-post",
		`<img src="__GHOST_URL__/content/images/2024/photo.jpg">`,
		"__GHOST_URL__/content/images/2024/cover.jpg",
	)
	reader := &mockReader{export: domain.RawExport{
		Posts: []domain.RawPost{{Post: post}},
	}}

	media := &mockMediaStore{files: map[string]string{
		"content/images/2024/photo.jpg": "photo data",
		"content/images/2024/cover.jpg": "cover data",
		"content/images/orphan.jpg":     "orphan data",
	}}

	converter := &mockConverter{
		convertFn: func(html string, rewriter port.URLRewriter) (string, error) {
			if rewriter != nil {
				// Verify the rewriter works
				rewritten := rewriter("__GHOST_URL__/content/images/2024/photo.jpg")
				return fmt.Sprintf("![img](%s)\n", rewritten), nil
			}
			return html, nil
		},
	}

	writer := &mockWriter{}
	journal := &mockJournalStore{}
	reporter := &mockReporter{}

	exportPath := createTestExportFile(t, reader.export)

	uc := NewMigrateUseCase(reader, converter, media, writer, journal, reporter)
	plan, err := uc.Execute(MigrateOptions{
		InputPath:    exportPath,
		IncludePages: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if plan.Stats.TotalPosts != 1 {
		t.Errorf("TotalPosts = %d", plan.Stats.TotalPosts)
	}
	if plan.Stats.TotalMedia != 2 {
		t.Errorf("TotalMedia = %d", plan.Stats.TotalMedia)
	}
	if plan.Stats.OrphanedMedia != 1 {
		t.Errorf("OrphanedMedia = %d", plan.Stats.OrphanedMedia)
	}

	// Should have written: index.md, photo.jpg, cover.jpg, orphan
	if len(writer.written) < 3 {
		t.Errorf("written = %d files, want >= 3", len(writer.written))
	}

	// Check index has rewritten URL
	var indexContent string
	for _, w := range writer.written {
		if strings.HasSuffix(w.path, "index.md") {
			indexContent = w.content
		}
	}
	if indexContent == "" {
		t.Fatal("no index.md written")
	}
	if strings.Contains(indexContent, "__GHOST_URL__") {
		t.Error("index.md still contains __GHOST_URL__")
	}

	// Check journal was saved
	if journal.saved == nil {
		t.Error("journal not saved")
	}
	if journal.saved.Status != domain.JournalStatusCompleted {
		t.Errorf("journal status = %q", journal.saved.Status)
	}
}

func TestDryRun(t *testing.T) {
	post := makePost("dry", "<p>content</p>", "")
	reader := &mockReader{export: domain.RawExport{
		Posts: []domain.RawPost{{Post: post}},
	}}
	writer := &mockWriter{}

	exportPath := createTestExportFile(t, reader.export)

	uc := NewMigrateUseCase(reader, &mockConverter{}, nil, writer, nil, nil)
	_, err := uc.Execute(MigrateOptions{
		InputPath:    exportPath,
		IncludePages: true,
		DryRun:       true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(writer.written) != 0 {
		t.Errorf("dry-run wrote %d files", len(writer.written))
	}
}

func TestFilterByStatus(t *testing.T) {
	pub := makePost("published", "", "")
	pub.Status = domain.StatusPublished
	draft := makePost("draft", "", "")
	draft.Status = domain.StatusDraft

	reader := &mockReader{export: domain.RawExport{
		Posts: []domain.RawPost{{Post: pub}, {Post: draft}},
	}}
	writer := &mockWriter{}

	exportPath := createTestExportFile(t, reader.export)

	uc := NewMigrateUseCase(reader, &mockConverter{}, nil, writer, nil, nil)
	plan, err := uc.Execute(MigrateOptions{
		InputPath:    exportPath,
		IncludePages: true,
		Status:       "published",
	})
	if err != nil {
		t.Fatal(err)
	}

	if plan.Stats.TotalPosts != 1 {
		t.Errorf("TotalPosts = %d, want 1 (filtered)", plan.Stats.TotalPosts)
	}
}

func TestFilterPages(t *testing.T) {
	post := makePost("post", "", "")
	post.Type = domain.TypePost
	page := makePost("page", "", "")
	page.Type = domain.TypePage

	reader := &mockReader{export: domain.RawExport{
		Posts: []domain.RawPost{{Post: post}, {Post: page}},
	}}
	writer := &mockWriter{}

	exportPath := createTestExportFile(t, reader.export)

	uc := NewMigrateUseCase(reader, &mockConverter{}, nil, writer, nil, nil)
	plan, err := uc.Execute(MigrateOptions{
		InputPath:    exportPath,
		IncludePages: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	if plan.Stats.TotalPosts != 1 {
		t.Errorf("TotalPosts = %d, want 1 (pages excluded)", plan.Stats.TotalPosts)
	}
}

func TestDiscardOrphaned(t *testing.T) {
	post := makePost("post", `<img src="__GHOST_URL__/content/images/used.jpg">`, "")
	reader := &mockReader{export: domain.RawExport{
		Posts: []domain.RawPost{{Post: post}},
	}}

	media := &mockMediaStore{files: map[string]string{
		"content/images/used.jpg":   "used",
		"content/images/orphan.jpg": "orphan",
	}}

	writer := &mockWriter{}
	exportPath := createTestExportFile(t, reader.export)

	uc := NewMigrateUseCase(reader, &mockConverter{}, media, writer, nil, nil)
	plan, err := uc.Execute(MigrateOptions{
		InputPath:       exportPath,
		IncludePages:    true,
		DiscardOrphaned: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(plan.OrphanedMedia) != 0 {
		t.Errorf("OrphanedMedia = %d, want 0 (discarded)", len(plan.OrphanedMedia))
	}

	// No orphan should be written
	for _, w := range writer.written {
		if strings.HasPrefix(w.path, "_orphaned/") {
			t.Errorf("orphan written despite discard: %s", w.path)
		}
	}
}

