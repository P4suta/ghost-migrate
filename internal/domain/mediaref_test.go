package domain

import "testing"

func TestExtractMediaRefs(t *testing.T) {
	tests := []struct {
		name string
		html string
		want []string // expected content paths
	}{
		{
			"ghost URL",
			`<img src="__GHOST_URL__/content/images/2024/09/photo.jpg">`,
			[]string{"content/images/2024/09/photo.jpg"},
		},
		{
			"absolute URL",
			`<img src="https://example.com/content/images/2024/09/photo.jpg">`,
			[]string{"content/images/2024/09/photo.jpg"},
		},
		{
			"media and files",
			`<a href="__GHOST_URL__/content/files/2024/doc.pdf"><video src="__GHOST_URL__/content/media/2024/video.mp4">`,
			[]string{"content/files/2024/doc.pdf", "content/media/2024/video.mp4"},
		},
		{
			"dedup same path",
			`<img src="__GHOST_URL__/content/images/a.jpg"><img src="__GHOST_URL__/content/images/a.jpg">`,
			[]string{"content/images/a.jpg"},
		},
		{
			"no matches",
			`<img src="/other/path.jpg">`,
			nil,
		},
		{
			"multiple images",
			`<img src="__GHOST_URL__/content/images/a.jpg"><img src="https://blog.example.com/content/images/b.png">`,
			[]string{"content/images/a.jpg", "content/images/b.png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refs := ExtractMediaRefs(tt.html)
			if len(refs) != len(tt.want) {
				t.Fatalf("got %d refs, want %d", len(refs), len(tt.want))
			}
			for i, ref := range refs {
				if ref.ContentPath != tt.want[i] {
					t.Errorf("refs[%d].ContentPath = %q, want %q", i, ref.ContentPath, tt.want[i])
				}
			}
		})
	}
}

func TestExtractFeatureImageRef(t *testing.T) {
	ref := ExtractFeatureImageRef("__GHOST_URL__/content/images/2024/cover.jpg")
	if ref == nil {
		t.Fatal("expected non-nil ref")
	}
	if ref.ContentPath != "content/images/2024/cover.jpg" {
		t.Errorf("ContentPath = %q", ref.ContentPath)
	}

	if ExtractFeatureImageRef("") != nil {
		t.Error("empty URL should return nil")
	}
	if ExtractFeatureImageRef("https://external.com/photo.jpg") != nil {
		t.Error("non-content URL should return nil")
	}
}

func TestNormalizeContentPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"content/images/2024/photo.jpg", "content/images/2024/photo.jpg"},
		{"japan-travel_123/content/images/2024/photo.jpg", "content/images/2024/photo.jpg"},
		{"japan-travel_123/japan-travel_123/content/images/2024/photo.jpg", "content/images/2024/photo.jpg"},
		{"random/path/no-content.jpg", "random/path/no-content.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := NormalizeContentPath(tt.input); got != tt.want {
				t.Errorf("NormalizeContentPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
