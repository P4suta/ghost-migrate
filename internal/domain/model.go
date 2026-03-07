package domain

import "time"

type Post struct {
	ID                  string
	Title               string
	Slug                string
	HTML                string
	Status              PostStatus
	Type                PostType
	Visibility          string
	Featured            bool
	FeatureImage        string
	FeatureImageAlt     string
	FeatureImageCaption string
	CustomExcerpt       string
	MetaTitle           string
	MetaDescription     string
	CanonicalURL        string
	PublishedAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Tags                []Tag
	Authors             []Author
}

type PostStatus string

const (
	StatusPublished PostStatus = "published"
	StatusDraft     PostStatus = "draft"
	StatusScheduled PostStatus = "scheduled"
)

type PostType string

const (
	TypePost PostType = "post"
	TypePage PostType = "page"
)

type Tag struct {
	ID          string
	Name        string
	Slug        string
	Description string
	Visibility  string
}

func (t Tag) IsInternal() bool {
	return len(t.Name) > 0 && t.Name[0] == '#'
}

type Author struct {
	ID   string
	Name string
	Slug string
}

type Article struct {
	Filename    string
	FrontMatter FrontMatter
	Content     string
}

type RawExport struct {
	Posts        []RawPost
	Tags         []Tag
	Authors      []Author
	PostsTags    []PostTag
	PostsAuthors []PostAuthor
	PostsMeta    []PostMeta
	Version      string
}

type RawPost struct {
	Post
}

type PostTag struct {
	PostID    string
	TagID     string
	SortOrder int
}

type PostAuthor struct {
	PostID    string
	AuthorID  string
	SortOrder int
}

type PostMeta struct {
	PostID              string
	MetaTitle           string
	MetaDescription     string
	FeatureImageAlt     string
	FeatureImageCaption string
}
