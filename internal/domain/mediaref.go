package domain

import (
	"regexp"
	"strings"
)

type MediaReference struct {
	OriginalURL string
	ContentPath string
}

type MediaFile struct {
	Path string
	Size int64
}

var mediaURLPattern = regexp.MustCompile(`(?:__GHOST_URL__|https?://[^/]+)/(content/(?:images|media|files)/[^\s"')\]]+)`)

func ExtractMediaRefs(html string) []MediaReference {
	matches := mediaURLPattern.FindAllStringSubmatch(html, -1)
	seen := make(map[string]bool, len(matches))
	refs := make([]MediaReference, 0, len(matches))
	for _, m := range matches {
		contentPath := m[1]
		if seen[contentPath] {
			continue
		}
		seen[contentPath] = true
		refs = append(refs, MediaReference{
			OriginalURL: m[0],
			ContentPath: contentPath,
		})
	}
	return refs
}

func ExtractFeatureImageRef(url string) *MediaReference {
	if url == "" {
		return nil
	}
	matches := mediaURLPattern.FindStringSubmatch(url)
	if matches == nil {
		return nil
	}
	return &MediaReference{
		OriginalURL: matches[0],
		ContentPath: matches[1],
	}
}

func NormalizeContentPath(zipPath string) string {
	idx := strings.Index(zipPath, "content/")
	if idx < 0 {
		return zipPath
	}
	return zipPath[idx:]
}
