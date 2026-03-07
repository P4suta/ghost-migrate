package domain

import "testing"

func TestResolveRelationships(t *testing.T) {
	raw := RawExport{
		Posts: []RawPost{
			{Post: Post{ID: "p1"}},
			{Post: Post{ID: "p2"}},
		},
		Tags: []Tag{
			{ID: "t1", Name: "Go"},
			{ID: "t2", Name: "Rust"},
		},
		Authors: []Author{
			{ID: "a1", Name: "Alice"},
		},
		PostsTags: []PostTag{
			{PostID: "p1", TagID: "t2", SortOrder: 1},
			{PostID: "p1", TagID: "t1", SortOrder: 0},
			{PostID: "p2", TagID: "t1", SortOrder: 0},
		},
		PostsAuthors: []PostAuthor{
			{PostID: "p1", AuthorID: "a1", SortOrder: 0},
			{PostID: "p2", AuthorID: "a1", SortOrder: 0},
		},
		PostsMeta: []PostMeta{
			{PostID: "p1", MetaDescription: "desc for p1"},
		},
	}

	posts := ResolveRelationships(raw)

	if len(posts) != 2 {
		t.Fatalf("got %d posts, want 2", len(posts))
	}

	// p1 should have tags sorted: Go (0), Rust (1)
	p1 := posts[0]
	if len(p1.Tags) != 2 {
		t.Fatalf("p1 tags: got %d, want 2", len(p1.Tags))
	}
	if p1.Tags[0].Name != "Go" {
		t.Errorf("p1.Tags[0] = %q, want Go", p1.Tags[0].Name)
	}
	if p1.Tags[1].Name != "Rust" {
		t.Errorf("p1.Tags[1] = %q, want Rust", p1.Tags[1].Name)
	}

	if len(p1.Authors) != 1 || p1.Authors[0].Name != "Alice" {
		t.Errorf("p1 authors unexpected: %v", p1.Authors)
	}

	if p1.MetaDescription != "desc for p1" {
		t.Errorf("p1.MetaDescription = %q, want %q", p1.MetaDescription, "desc for p1")
	}

	// p2 should have no meta
	p2 := posts[1]
	if p2.MetaDescription != "" {
		t.Errorf("p2.MetaDescription = %q, want empty", p2.MetaDescription)
	}
}
