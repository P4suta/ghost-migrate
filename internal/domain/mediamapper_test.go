package domain

import (
	"sort"
	"testing"
)

func TestMediaMapper_SingleRef(t *testing.T) {
	files := []MediaFile{
		{Path: "content/images/2024/photo.jpg", Size: 1000},
	}
	mapper := NewMediaMapper(files)

	refs := []MediaReference{{ContentPath: "content/images/2024/photo.jpg"}}
	mapper.RegisterPostRefs(refs)

	mappings := mapper.BuildMappings(refs)
	if len(mappings) != 1 {
		t.Fatalf("got %d mappings, want 1", len(mappings))
	}
	if mappings[0].DestFilename != "photo.jpg" {
		t.Errorf("DestFilename = %q, want %q", mappings[0].DestFilename, "photo.jpg")
	}
	if mappings[0].IsShared {
		t.Error("single ref should not be shared")
	}
}

func TestMediaMapper_SharedRef(t *testing.T) {
	files := []MediaFile{
		{Path: "content/images/2024/shared.jpg", Size: 2000},
	}
	mapper := NewMediaMapper(files)

	refs := []MediaReference{{ContentPath: "content/images/2024/shared.jpg"}}
	mapper.RegisterPostRefs(refs) // post 1
	mapper.RegisterPostRefs(refs) // post 2

	mappings := mapper.BuildMappings(refs)
	if len(mappings) != 1 {
		t.Fatalf("got %d mappings, want 1", len(mappings))
	}
	if !mappings[0].IsShared {
		t.Error("multi-ref should be shared")
	}
}

func TestMediaMapper_CollisionResolution(t *testing.T) {
	files := []MediaFile{
		{Path: "content/images/2024/01/photo.jpg", Size: 1000},
		{Path: "content/images/2024/02/photo.jpg", Size: 2000},
	}
	mapper := NewMediaMapper(files)

	refs := []MediaReference{
		{ContentPath: "content/images/2024/01/photo.jpg"},
		{ContentPath: "content/images/2024/02/photo.jpg"},
	}
	mapper.RegisterPostRefs(refs)

	mappings := mapper.BuildMappings(refs)
	if len(mappings) != 2 {
		t.Fatalf("got %d mappings, want 2", len(mappings))
	}
	if mappings[0].DestFilename != "photo.jpg" {
		t.Errorf("first = %q, want %q", mappings[0].DestFilename, "photo.jpg")
	}
	if mappings[1].DestFilename != "photo-1.jpg" {
		t.Errorf("second = %q, want %q", mappings[1].DestFilename, "photo-1.jpg")
	}
}

func TestMediaMapper_UnresolvedRef(t *testing.T) {
	mapper := NewMediaMapper(nil)

	refs := []MediaReference{{ContentPath: "content/images/missing.jpg"}}
	mapper.RegisterPostRefs(refs)

	mappings := mapper.BuildMappings(refs)
	if len(mappings) != 0 {
		t.Errorf("unresolved ref should produce no mappings, got %d", len(mappings))
	}
}

func TestMediaMapper_FindOrphans(t *testing.T) {
	files := []MediaFile{
		{Path: "content/images/used.jpg", Size: 100},
		{Path: "content/images/orphan.jpg", Size: 200},
	}
	mapper := NewMediaMapper(files)

	refs := []MediaReference{{ContentPath: "content/images/used.jpg"}}
	mapper.RegisterPostRefs(refs)

	orphans := mapper.FindOrphans()
	if len(orphans) != 1 {
		t.Fatalf("got %d orphans, want 1", len(orphans))
	}
	if orphans[0].Path != "content/images/orphan.jpg" {
		t.Errorf("orphan = %q", orphans[0].Path)
	}
}

func TestUniqueFilename(t *testing.T) {
	used := make(map[string]bool)

	got1 := uniqueFilename("photo.jpg", used)
	if got1 != "photo.jpg" {
		t.Errorf("first = %q", got1)
	}

	got2 := uniqueFilename("photo.jpg", used)
	if got2 != "photo-1.jpg" {
		t.Errorf("second = %q", got2)
	}

	got3 := uniqueFilename("photo.jpg", used)
	if got3 != "photo-2.jpg" {
		t.Errorf("third = %q", got3)
	}

	// Sort the used keys to verify
	var keys []string
	for k := range used {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) != 3 {
		t.Errorf("used map has %d entries, want 3", len(keys))
	}
}
