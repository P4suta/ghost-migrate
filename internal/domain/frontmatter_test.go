package domain

import (
	"strings"
	"testing"
	"time"
)

func TestFrontMatter_Marshal(t *testing.T) {
	date := time.Date(2024, 3, 7, 12, 0, 0, 0, time.UTC)
	lastmod := time.Date(2024, 3, 7, 14, 0, 0, 0, time.UTC)

	fm := FrontMatter{
		Title:         "Hello World",
		Slug:          "hello-world",
		Date:          &date,
		Lastmod:       lastmod,
		Draft:         false,
		Tags:          []string{"Go", "Programming"},
		Authors:       []string{"Alice"},
		Description:   "A test post",
		FeaturedImage: "https://example.com/img.jpg",
		Visibility:    "public",
	}

	got := fm.Marshal()

	checks := []string{
		"---\n",
		"title: Hello World\n",
		"slug: hello-world\n",
		"date: 2024-03-07T12:00:00Z\n",
		"lastmod: 2024-03-07T14:00:00Z\n",
		"tags:\n  - Go\n  - Programming\n",
		"authors:\n  - Alice\n",
		"description: A test post\n",
		"featured_image: https://example.com/img.jpg\n",
	}

	for _, c := range checks {
		if !strings.Contains(got, c) {
			t.Errorf("Marshal() missing %q\ngot:\n%s", c, got)
		}
	}

	if strings.Contains(got, "visibility:") {
		t.Errorf("Marshal() should omit visibility: public")
	}

	if strings.Contains(got, "draft:") {
		t.Errorf("Marshal() should omit draft: false")
	}
}

func TestFrontMatter_MarshalDraft(t *testing.T) {
	lastmod := time.Date(2024, 3, 7, 14, 0, 0, 0, time.UTC)
	fm := FrontMatter{
		Title:      "Draft",
		Slug:       "draft",
		Lastmod:    lastmod,
		Draft:      true,
		Visibility: "members",
	}

	got := fm.Marshal()
	if !strings.Contains(got, "draft: true\n") {
		t.Errorf("Marshal() missing draft: true\ngot:\n%s", got)
	}
	if !strings.Contains(got, "visibility: members\n") {
		t.Errorf("Marshal() missing visibility: members\ngot:\n%s", got)
	}
}

func TestEscapeYAML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"", "\"\""},
		{"has: colon", "\"has: colon\""},
		{`has "quotes"`, `"has \"quotes\""`},
		{"has #hash", "\"has #hash\""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := escapeYAML(tt.input); got != tt.want {
				t.Errorf("escapeYAML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFromPost(t *testing.T) {
	date := time.Date(2024, 3, 7, 12, 0, 0, 0, time.UTC)
	p := Post{
		Title:           "Test",
		Slug:            "test",
		Status:          StatusPublished,
		MetaDescription: "meta desc",
		PublishedAt:     &date,
		UpdatedAt:       date,
		Tags: []Tag{
			{Name: "Go"},
			{Name: "#internal"},
		},
		Authors:    []Author{{Name: "Alice"}},
		Visibility: "public",
	}

	fm := FromPost(p, false)
	if len(fm.Tags) != 1 || fm.Tags[0] != "Go" {
		t.Errorf("FromPost(includeInternal=false) tags = %v, want [Go]", fm.Tags)
	}

	fm2 := FromPost(p, true)
	if len(fm2.Tags) != 2 {
		t.Errorf("FromPost(includeInternal=true) tags = %v, want 2 tags", fm2.Tags)
	}

	if fm.Description != "meta desc" {
		t.Errorf("FromPost description = %q, want %q", fm.Description, "meta desc")
	}
}

func TestFromPostWithMedia(t *testing.T) {
	date := time.Date(2024, 3, 7, 12, 0, 0, 0, time.UTC)
	p := Post{
		Title:        "Test",
		Slug:         "test",
		Status:       StatusPublished,
		FeatureImage: "https://example.com/content/images/2024/photo.jpg",
		PublishedAt:  &date,
		UpdatedAt:    date,
		Visibility:   "public",
	}

	fm := FromPostWithMedia(p, false, "photo.jpg")
	if fm.FeaturedImage != "photo.jpg" {
		t.Errorf("FromPostWithMedia() FeaturedImage = %q, want %q", fm.FeaturedImage, "photo.jpg")
	}

	fm2 := FromPostWithMedia(p, false, "")
	if fm2.FeaturedImage != p.FeatureImage {
		t.Errorf("FromPostWithMedia(empty) FeaturedImage = %q, want original URL", fm2.FeaturedImage)
	}
}
