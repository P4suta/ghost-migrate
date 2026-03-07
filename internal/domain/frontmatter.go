package domain

import (
	"fmt"
	"strings"
	"time"
)

type FrontMatter struct {
	Title         string
	Slug          string
	Date          *time.Time
	Lastmod       time.Time
	Draft         bool
	Tags          []string
	Authors       []string
	Description   string
	FeaturedImage string
	CanonicalURL  string
	Visibility    string
}

func FromPost(p Post, includeInternalTags bool) FrontMatter {
	tags := make([]string, 0, len(p.Tags))
	for _, t := range p.Tags {
		if !includeInternalTags && t.IsInternal() {
			continue
		}
		tags = append(tags, t.Name)
	}

	authors := make([]string, 0, len(p.Authors))
	for _, a := range p.Authors {
		authors = append(authors, a.Name)
	}

	description := p.CustomExcerpt
	if description == "" {
		description = p.MetaDescription
	}

	return FrontMatter{
		Title:         p.Title,
		Slug:          p.Slug,
		Date:          p.PublishedAt,
		Lastmod:       p.UpdatedAt,
		Draft:         p.Status != StatusPublished,
		Tags:          tags,
		Authors:       authors,
		Description:   description,
		FeaturedImage: p.FeatureImage,
		CanonicalURL:  p.CanonicalURL,
		Visibility:    p.Visibility,
	}
}

func FromPostWithMedia(p Post, includeInternalTags bool, featureImageLocal string) FrontMatter {
	fm := FromPost(p, includeInternalTags)
	if featureImageLocal != "" {
		fm.FeaturedImage = featureImageLocal
	}
	return fm
}

func (fm FrontMatter) Marshal() string {
	var b strings.Builder
	b.WriteString("---\n")

	writeField(&b, "title", escapeYAML(fm.Title))
	writeField(&b, "slug", fm.Slug)

	if fm.Date != nil {
		writeField(&b, "date", fm.Date.Format(time.RFC3339))
	}
	writeField(&b, "lastmod", fm.Lastmod.Format(time.RFC3339))

	if fm.Draft {
		writeField(&b, "draft", "true")
	}

	if len(fm.Tags) > 0 {
		writeList(&b, "tags", fm.Tags)
	}
	if len(fm.Authors) > 0 {
		writeList(&b, "authors", fm.Authors)
	}

	if fm.Description != "" {
		writeField(&b, "description", escapeYAML(fm.Description))
	}
	if fm.FeaturedImage != "" {
		writeField(&b, "featured_image", fm.FeaturedImage)
	}
	if fm.CanonicalURL != "" {
		writeField(&b, "canonical_url", fm.CanonicalURL)
	}
	if fm.Visibility != "public" {
		writeField(&b, "visibility", fm.Visibility)
	}

	b.WriteString("---\n")
	return b.String()
}

func writeField(b *strings.Builder, key, value string) {
	fmt.Fprintf(b, "%s: %s\n", key, value)
}

func writeList(b *strings.Builder, key string, items []string) {
	fmt.Fprintf(b, "%s:\n", key)
	for _, item := range items {
		fmt.Fprintf(b, "  - %s\n", escapeYAML(item))
	}
}

func escapeYAML(s string) string {
	if s == "" {
		return "\"\""
	}
	needsQuoting := strings.ContainsAny(s, ":\n#\"'{}[]|>&*!%@`,?")
	if needsQuoting {
		escaped := strings.ReplaceAll(s, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		return "\"" + escaped + "\""
	}
	return s
}
