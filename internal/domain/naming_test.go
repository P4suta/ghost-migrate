package domain

import "testing"

func TestFilename(t *testing.T) {
	tests := []struct {
		name string
		post Post
		want string
	}{
		{
			"normal slug",
			Post{Slug: "hello-world"},
			"hello-world.md",
		},
		{
			"empty slug uses title",
			Post{Title: "My Post Title"},
			"my-post-title.md",
		},
		{
			"empty slug and title uses ID",
			Post{ID: "abc123"},
			"abc123.md",
		},
		{
			"unsafe chars removed",
			Post{Slug: "hello world!@#"},
			"hello-world.md",
		},
		{
			"multiple dashes collapsed",
			Post{Slug: "hello---world"},
			"hello-world.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filename(tt.post); got != tt.want {
				t.Errorf("Filename() = %q, want %q", got, tt.want)
			}
		})
	}
}
