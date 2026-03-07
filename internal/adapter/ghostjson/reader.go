package ghostjson

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"ghost-migrate/internal/domain"
)

type Reader struct{}

func NewReader() *Reader {
	return &Reader{}
}

func (r *Reader) Read(input io.Reader) (domain.RawExport, error) {
	var export ghostExport
	if err := json.NewDecoder(input).Decode(&export); err != nil {
		return domain.RawExport{}, fmt.Errorf("failed to parse Ghost JSON: %w", err)
	}

	if len(export.DB) == 0 {
		return domain.RawExport{}, fmt.Errorf("invalid Ghost export: db array is empty")
	}

	data := export.DB[0].Data
	meta := export.DB[0].Meta

	posts := make([]domain.RawPost, 0, len(data.Posts))
	for _, jp := range data.Posts {
		p, err := convertPost(jp)
		if err != nil {
			return domain.RawExport{}, fmt.Errorf("failed to convert post %q: %w", jp.ID, err)
		}
		posts = append(posts, domain.RawPost{Post: p})
	}

	tags := make([]domain.Tag, len(data.Tags))
	for i, jt := range data.Tags {
		tags[i] = domain.Tag{
			ID:          jt.ID,
			Name:        jt.Name,
			Slug:        jt.Slug,
			Description: deref(jt.Description),
			Visibility:  jt.Visibility,
		}
	}

	authors := make([]domain.Author, len(data.Users))
	for i, ju := range data.Users {
		authors[i] = domain.Author{
			ID:   ju.ID,
			Name: ju.Name,
			Slug: ju.Slug,
		}
	}

	postTags := make([]domain.PostTag, len(data.PostsTags))
	for i, jpt := range data.PostsTags {
		postTags[i] = domain.PostTag{
			PostID:    jpt.PostID,
			TagID:     jpt.TagID,
			SortOrder: jpt.SortOrder,
		}
	}

	postAuthors := make([]domain.PostAuthor, len(data.PostsAuthors))
	for i, jpa := range data.PostsAuthors {
		postAuthors[i] = domain.PostAuthor{
			PostID:    jpa.PostID,
			AuthorID:  jpa.AuthorID,
			SortOrder: jpa.SortOrder,
		}
	}

	postsMeta := make([]domain.PostMeta, len(data.PostsMeta))
	for i, jpm := range data.PostsMeta {
		postsMeta[i] = domain.PostMeta{
			PostID:              jpm.PostID,
			MetaTitle:           deref(jpm.MetaTitle),
			MetaDescription:     deref(jpm.MetaDescription),
			FeatureImageAlt:     deref(jpm.FeatureImageAlt),
			FeatureImageCaption: deref(jpm.FeatureImageCaption),
		}
	}

	return domain.RawExport{
		Posts:        posts,
		Tags:         tags,
		Authors:      authors,
		PostsTags:    postTags,
		PostsAuthors: postAuthors,
		PostsMeta:    postsMeta,
		Version:      meta.Version,
	}, nil
}

func convertPost(jp jsonPost) (domain.Post, error) {
	createdAt, err := parseGhostTime(jp.CreatedAt)
	if err != nil {
		return domain.Post{}, fmt.Errorf("invalid created_at: %w", err)
	}
	updatedAt, err := parseGhostTime(jp.UpdatedAt)
	if err != nil {
		return domain.Post{}, fmt.Errorf("invalid updated_at: %w", err)
	}

	var publishedAt *time.Time
	if jp.PublishedAt != nil {
		t, err := parseGhostTime(*jp.PublishedAt)
		if err != nil {
			return domain.Post{}, fmt.Errorf("invalid published_at: %w", err)
		}
		publishedAt = &t
	}

	return domain.Post{
		ID:            jp.ID,
		Title:         jp.Title,
		Slug:          jp.Slug,
		HTML:          deref(jp.HTML),
		Status:        domain.PostStatus(jp.Status),
		Type:          domain.PostType(jp.Type),
		Visibility:    jp.Visibility,
		Featured:      bool(jp.Featured),
		FeatureImage:  deref(jp.FeatureImage),
		CustomExcerpt: deref(jp.CustomExcerpt),
		CanonicalURL:  deref(jp.CanonicalURL),
		PublishedAt:   publishedAt,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

func parseGhostTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000+00:00",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format: %q", s)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
