package domain

type PageBundle struct {
	Slug         string
	IndexContent string
	MediaEntries []BundleMediaEntry
}

type BundleMediaEntry struct {
	SourcePath   string
	DestFilename string
	Size         int64
	IsShared     bool
}
