package usecase

import (
	"fmt"
	"os"
	"path"

	"ghost-migrate/internal/domain"
	"ghost-migrate/internal/port"
)

type MigrateUseCase struct {
	reader    port.ExportReader
	converter port.ContentConverter
	media     port.MediaStore
	writer    port.BundleWriter
	journal   port.JournalStore
	reporter  ProgressReporter
}

func NewMigrateUseCase(
	reader port.ExportReader,
	converter port.ContentConverter,
	media port.MediaStore,
	writer port.BundleWriter,
	journal port.JournalStore,
	reporter ProgressReporter,
) *MigrateUseCase {
	return &MigrateUseCase{
		reader:    reader,
		converter: converter,
		media:     media,
		writer:    writer,
		journal:   journal,
		reporter:  reporter,
	}
}

func (uc *MigrateUseCase) Execute(opts MigrateOptions) (*domain.MigrationPlan, error) {
	f, err := os.Open(opts.InputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open export file: %w", err)
	}
	defer f.Close()

	rawExport, err := uc.reader.Read(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read export: %w", err)
	}

	posts := domain.ResolveRelationships(rawExport)
	posts = filterPosts(posts, opts)

	if uc.media == nil {
		return uc.executeFlatMode(posts, opts)
	}
	return uc.executeBundleMode(posts, opts)
}

func (uc *MigrateUseCase) executeFlatMode(posts []domain.Post, opts MigrateOptions) (*domain.MigrationPlan, error) {
	for _, p := range posts {
		md, err := uc.converter.Convert(p.HTML, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to convert post %q: %w", p.Slug, err)
		}

		fm := domain.FromPost(p, opts.IncludeInternalTags)
		article := domain.Article{
			Filename:    domain.Filename(p),
			FrontMatter: fm,
			Content:     md,
		}

		if !opts.DryRun {
			if err := uc.writer.WriteFlat(article); err != nil {
				return nil, fmt.Errorf("failed to write %q: %w", article.Filename, err)
			}
		}

		if uc.reporter != nil {
			uc.reporter.OnPostProcessed(p.Slug, 0)
		}
	}

	plan := &domain.MigrationPlan{
		Stats: domain.PlanStats{TotalPosts: len(posts)},
	}
	if uc.reporter != nil {
		uc.reporter.OnComplete(MigrateStats{TotalPosts: len(posts)})
	}
	return plan, nil
}

func (uc *MigrateUseCase) executeBundleMode(posts []domain.Post, opts MigrateOptions) (*domain.MigrationPlan, error) {
	// Phase 1: Analysis
	mediaFiles, err := uc.media.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list media: %w", err)
	}

	mapper := domain.NewMediaMapper(mediaFiles)

	// First pass: register all refs for shared detection
	postRefs := make(map[string][]domain.MediaReference, len(posts))
	for _, p := range posts {
		refs := domain.ExtractMediaRefs(p.HTML)
		if fRef := domain.ExtractFeatureImageRef(p.FeatureImage); fRef != nil {
			refs = appendUniqueRef(refs, *fRef)
		}
		postRefs[p.Slug] = refs
		mapper.RegisterPostRefs(refs)
	}

	// Build plan
	var bundles []domain.PageBundle
	totalMedia := 0
	sharedMedia := 0

	for _, p := range posts {
		refs := postRefs[p.Slug]
		mappings := mapper.BuildMappings(refs)

		// Build URL rewriter for this post
		rewriteMap := make(map[string]string, len(mappings))
		for _, m := range mappings {
			for _, ref := range refs {
				if ref.ContentPath == m.ContentPath {
					rewriteMap[ref.OriginalURL] = m.DestFilename
				}
			}
		}
		rewriter := func(url string) string {
			if dest, ok := rewriteMap[url]; ok {
				return dest
			}
			return url
		}

		md, err := uc.converter.Convert(p.HTML, rewriter)
		if err != nil {
			return nil, fmt.Errorf("failed to convert post %q: %w", p.Slug, err)
		}

		// Handle feature image local path
		featureLocal := ""
		if fRef := domain.ExtractFeatureImageRef(p.FeatureImage); fRef != nil {
			if dest, ok := rewriteMap[fRef.OriginalURL]; ok {
				featureLocal = dest
			}
		}

		fm := domain.FromPostWithMedia(p, opts.IncludeInternalTags, featureLocal)
		indexContent := fm.Marshal() + "\n" + md

		var mediaEntries []domain.BundleMediaEntry
		for _, m := range mappings {
			var size int64
			for _, mf := range mediaFiles {
				if mf.Path == m.ContentPath {
					size = mf.Size
					break
				}
			}
			mediaEntries = append(mediaEntries, domain.BundleMediaEntry{
				SourcePath:   m.ContentPath,
				DestFilename: m.DestFilename,
				Size:         size,
				IsShared:     m.IsShared,
			})
			totalMedia++
			if m.IsShared {
				sharedMedia++
			}
		}

		bundles = append(bundles, domain.PageBundle{
			Slug:         p.Slug,
			IndexContent: indexContent,
			MediaEntries: mediaEntries,
		})
	}

	orphans := mapper.FindOrphans()
	var orphanedMedia []domain.MediaFile
	if !opts.DiscardOrphaned {
		orphanedMedia = orphans
	}

	plan := &domain.MigrationPlan{
		Bundles:       bundles,
		OrphanedMedia: orphanedMedia,
		Stats: domain.PlanStats{
			TotalPosts:    len(posts),
			TotalMedia:    totalMedia,
			SharedMedia:   sharedMedia,
			OrphanedMedia: len(orphans),
		},
	}

	if opts.DryRun {
		if uc.reporter != nil {
			uc.reporter.OnComplete(statsFromPlan(plan))
		}
		return plan, nil
	}

	// Phase 2: Execution
	if err := uc.executePlan(plan); err != nil {
		return plan, err
	}

	if uc.reporter != nil {
		uc.reporter.OnComplete(statsFromPlan(plan))
	}
	return plan, nil
}

func (uc *MigrateUseCase) executePlan(plan *domain.MigrationPlan) error {
	j := domain.NewJournal()

	for _, bundle := range plan.Bundles {
		if err := uc.writer.WriteIndex(bundle.Slug, bundle.IndexContent); err != nil {
			return fmt.Errorf("failed to write index for %q: %w", bundle.Slug, err)
		}

		for _, entry := range bundle.MediaEntries {
			entryID := j.AddEntry(domain.FileOpCopy, entry.SourcePath, bundle.Slug+"/"+entry.DestFilename, entry.Size)

			rc, err := uc.media.Open(entry.SourcePath)
			if err != nil {
				j.MarkEntryFailed(entryID, err.Error())
				continue
			}

			if err := uc.writer.WriteMedia(bundle.Slug, entry.DestFilename, rc); err != nil {
				rc.Close()
				j.MarkEntryFailed(entryID, err.Error())
				continue
			}
			rc.Close()
			j.MarkEntryCompleted(entryID)
		}

		if uc.reporter != nil {
			uc.reporter.OnPostProcessed(bundle.Slug, len(bundle.MediaEntries))
		}
	}

	// Write orphans
	for _, orphan := range plan.OrphanedMedia {
		entryID := j.AddEntry(domain.FileOpCopy, orphan.Path, "_orphaned/"+path.Base(orphan.Path), orphan.Size)

		rc, err := uc.media.Open(orphan.Path)
		if err != nil {
			j.MarkEntryFailed(entryID, err.Error())
			continue
		}

		if err := uc.writer.WriteOrphan(path.Base(orphan.Path), rc); err != nil {
			rc.Close()
			j.MarkEntryFailed(entryID, err.Error())
			continue
		}
		rc.Close()
		j.MarkEntryCompleted(entryID)
	}

	j.Complete()
	if uc.journal != nil {
		if err := uc.journal.Save(j); err != nil {
			return fmt.Errorf("failed to save journal: %w", err)
		}
	}

	return nil
}

func filterPosts(posts []domain.Post, opts MigrateOptions) []domain.Post {
	var filtered []domain.Post
	for _, p := range posts {
		if !opts.IncludePages && p.Type == domain.TypePage {
			continue
		}
		if opts.Status != "" && string(p.Status) != opts.Status {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}

func appendUniqueRef(refs []domain.MediaReference, ref domain.MediaReference) []domain.MediaReference {
	for _, r := range refs {
		if r.ContentPath == ref.ContentPath {
			return refs
		}
	}
	return append(refs, ref)
}

func statsFromPlan(plan *domain.MigrationPlan) MigrateStats {
	return MigrateStats{
		TotalPosts:    plan.Stats.TotalPosts,
		TotalMedia:    plan.Stats.TotalMedia,
		SharedMedia:   plan.Stats.SharedMedia,
		OrphanedMedia: plan.Stats.OrphanedMedia,
	}
}

