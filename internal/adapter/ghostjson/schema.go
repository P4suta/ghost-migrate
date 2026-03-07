package ghostjson

type ghostExport struct {
	DB []dbEntry `json:"db"`
}

type dbEntry struct {
	Meta exportMeta `json:"meta"`
	Data exportData `json:"data"`
}

type exportMeta struct {
	ExportedOn int64  `json:"exported_on"`
	Version    string `json:"version"`
}

type exportData struct {
	Posts        []jsonPost       `json:"posts"`
	PostsMeta    []jsonPostMeta   `json:"posts_meta"`
	Tags         []jsonTag        `json:"tags"`
	Users        []jsonUser       `json:"users"`
	PostsTags    []jsonPostTag    `json:"posts_tags"`
	PostsAuthors []jsonPostAuthor `json:"posts_authors"`
}

type jsonPost struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	HTML          *string `json:"html"`
	Status        string  `json:"status"`
	Type          string  `json:"type"`
	Visibility    string  `json:"visibility"`
	Featured      intBool `json:"featured"`
	FeatureImage  *string `json:"feature_image"`
	CustomExcerpt *string `json:"custom_excerpt"`
	CanonicalURL  *string `json:"canonical_url"`
	PublishedAt   *string `json:"published_at"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type jsonPostMeta struct {
	ID                  string  `json:"id"`
	PostID              string  `json:"post_id"`
	MetaTitle           *string `json:"meta_title"`
	MetaDescription     *string `json:"meta_description"`
	FeatureImageAlt     *string `json:"feature_image_alt"`
	FeatureImageCaption *string `json:"feature_image_caption"`
}

type jsonTag struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	Visibility  string  `json:"visibility"`
}

type jsonUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Email string `json:"email"`
}

type jsonPostTag struct {
	PostID    string `json:"post_id"`
	TagID     string `json:"tag_id"`
	SortOrder int    `json:"sort_order"`
}

type jsonPostAuthor struct {
	PostID    string `json:"post_id"`
	AuthorID  string `json:"author_id"`
	SortOrder int    `json:"sort_order"`
}

type intBool bool

func (b *intBool) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "true", "1":
		*b = true
	default:
		*b = false
	}
	return nil
}
