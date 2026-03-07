package domain

type MigrationPlan struct {
	Bundles       []PageBundle
	OrphanedMedia []MediaFile
	Stats         PlanStats
}

type PlanStats struct {
	TotalPosts     int
	TotalMedia     int
	SharedMedia    int
	OrphanedMedia  int
	TotalMediaSize int64
}
