package ghostjson

import (
	"os"
	"testing"
)

func TestReader_Read(t *testing.T) {
	f, err := os.Open("../../../testdata/fixtures/minimal.json")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer f.Close()

	r := NewReader()
	export, err := r.Read(f)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	if export.Version != "5.80.0" {
		t.Errorf("Version = %q, want %q", export.Version, "5.80.0")
	}

	if len(export.Posts) != 3 {
		t.Errorf("Posts count = %d, want 3", len(export.Posts))
	}

	if len(export.Tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(export.Tags))
	}

	if len(export.Authors) != 1 {
		t.Errorf("Authors count = %d, want 1", len(export.Authors))
	}

	if len(export.PostsMeta) != 1 {
		t.Errorf("PostsMeta count = %d, want 1", len(export.PostsMeta))
	}

	p1 := export.Posts[0].Post
	if p1.Title != "Hello World" {
		t.Errorf("Post[0].Title = %q, want %q", p1.Title, "Hello World")
	}
	if p1.Status != "published" {
		t.Errorf("Post[0].Status = %q, want %q", p1.Status, "published")
	}
	if p1.FeatureImage != "https://example.com/img.jpg" {
		t.Errorf("Post[0].FeatureImage = %q", p1.FeatureImage)
	}

	p2 := export.Posts[1].Post
	if !p2.Featured {
		t.Error("Post[1].Featured = false, want true (from int 1)")
	}
	if p1.Featured {
		t.Error("Post[0].Featured = true, want false (from int 0)")
	}

	p3 := export.Posts[2].Post
	if p3.Featured {
		t.Error("Post[2].Featured = true, want false (from bool false)")
	}

	pm := export.PostsMeta[0]
	if pm.PostID != "post-1" {
		t.Errorf("PostsMeta[0].PostID = %q", pm.PostID)
	}
	if pm.MetaDescription != "A test post description" {
		t.Errorf("PostsMeta[0].MetaDescription = %q", pm.MetaDescription)
	}
}

func TestIntBool_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var b intBool
			if err := b.UnmarshalJSON([]byte(tt.input)); err != nil {
				t.Fatalf("UnmarshalJSON(%q) error: %v", tt.input, err)
			}
			if bool(b) != tt.want {
				t.Errorf("UnmarshalJSON(%q) = %v, want %v", tt.input, b, tt.want)
			}
		})
	}
}

func TestParseGhostTime(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2024-03-07T12:00:00.000Z", false},
		{"2024-03-07T12:00:00Z", false},
		{"2024-03-07T12:00:00.000+00:00", false},
		{"2024-03-07 12:00:00", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseGhostTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseGhostTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
