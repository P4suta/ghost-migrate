package usecase

type MigrateOptions struct {
	InputPath           string
	OutputDir           string
	MediaPath           string
	Status              string
	IncludePages        bool
	IncludeInternalTags bool
	DiscardOrphaned     bool
	DryRun              bool
}

type ProgressReporter interface {
	OnPostProcessed(slug string, mediaCount int)
	OnComplete(stats MigrateStats)
}

type MigrateStats struct {
	TotalPosts    int
	TotalMedia    int
	SharedMedia   int
	OrphanedMedia int
}
