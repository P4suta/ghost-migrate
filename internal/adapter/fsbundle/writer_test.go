package fsbundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ghost-migrate/internal/domain"
)

func TestWriter_WriteIndex(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir)

	if err := w.WriteIndex("my-post", "# Hello\n"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "my-post", "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# Hello\n" {
		t.Errorf("content = %q", string(data))
	}
}

func TestWriter_WriteMedia(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir)

	if err := w.WriteMedia("my-post", "photo.jpg", strings.NewReader("image")); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "my-post", "photo.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "image" {
		t.Errorf("content = %q", string(data))
	}
}

func TestWriter_WriteOrphan(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir)

	if err := w.WriteOrphan("unused.jpg", strings.NewReader("orphan")); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "_orphaned", "unused.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "orphan" {
		t.Errorf("content = %q", string(data))
	}
}

func TestWriter_WriteFlat(t *testing.T) {
	dir := t.TempDir()
	w := NewWriter(dir)

	lastmod := time.Date(2024, 3, 7, 12, 0, 0, 0, time.UTC)
	article := domain.Article{
		Filename: "hello-world.md",
		FrontMatter: domain.FrontMatter{
			Title:      "Hello",
			Slug:       "hello-world",
			Lastmod:    lastmod,
			Visibility: "public",
		},
		Content: "Hello world\n",
	}

	if err := w.WriteFlat(article); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "hello-world.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "title: Hello") {
		t.Errorf("missing frontmatter in output: %q", string(data))
	}
	if !strings.Contains(string(data), "Hello world") {
		t.Errorf("missing content in output: %q", string(data))
	}
}
