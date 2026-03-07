package domain

import (
	"fmt"
	"path"
	"strings"
)

type MediaMapping struct {
	ContentPath  string
	DestFilename string
	IsShared     bool
}

type MediaMapper struct {
	refCount  map[string]int
	fileIndex map[string]MediaFile
}

func NewMediaMapper(files []MediaFile) *MediaMapper {
	index := make(map[string]MediaFile, len(files))
	for _, f := range files {
		index[f.Path] = f
	}
	return &MediaMapper{
		refCount:  make(map[string]int),
		fileIndex: index,
	}
}

func (m *MediaMapper) RegisterPostRefs(refs []MediaReference) {
	for _, ref := range refs {
		if _, ok := m.fileIndex[ref.ContentPath]; ok {
			m.refCount[ref.ContentPath]++
		}
	}
}

func (m *MediaMapper) BuildMappings(refs []MediaReference) []MediaMapping {
	used := make(map[string]bool)
	mappings := make([]MediaMapping, 0, len(refs))

	for _, ref := range refs {
		if _, ok := m.fileIndex[ref.ContentPath]; !ok {
			continue
		}
		dest := uniqueFilename(path.Base(ref.ContentPath), used)
		mappings = append(mappings, MediaMapping{
			ContentPath:  ref.ContentPath,
			DestFilename: dest,
			IsShared:     m.refCount[ref.ContentPath] > 1,
		})
	}
	return mappings
}

func (m *MediaMapper) IsShared(contentPath string) bool {
	return m.refCount[contentPath] > 1
}

func (m *MediaMapper) FindOrphans() []MediaFile {
	var orphans []MediaFile
	for p, f := range m.fileIndex {
		if m.refCount[p] == 0 {
			orphans = append(orphans, f)
		}
	}
	return orphans
}

func uniqueFilename(name string, used map[string]bool) string {
	if !used[name] {
		used[name] = true
		return name
	}
	ext := path.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d%s", base, i, ext)
		if !used[candidate] {
			used[candidate] = true
			return candidate
		}
	}
}
