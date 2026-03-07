package domain

import (
	"regexp"
	"strings"
)

var (
	unsafeChars    = regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	multipleDashes = regexp.MustCompile(`-{2,}`)
)

func Filename(p Post) string {
	slug := p.Slug
	if slug == "" {
		slug = sanitizeSlug(p.Title)
	}
	slug = sanitizeSlug(slug)
	if slug == "" {
		slug = p.ID
	}
	return slug + ".md"
}

func sanitizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = unsafeChars.ReplaceAllString(s, "")
	s = multipleDashes.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
