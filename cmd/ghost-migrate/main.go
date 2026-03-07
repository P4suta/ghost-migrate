package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"ghost-migrate/internal/adapter/fsbundle"
	"ghost-migrate/internal/adapter/fsjournal"
	"ghost-migrate/internal/adapter/ghostjson"
	"ghost-migrate/internal/adapter/htmlconv"
	"ghost-migrate/internal/adapter/mediafs"
	"ghost-migrate/internal/adapter/mediazip"
	"ghost-migrate/internal/domain"
	"ghost-migrate/internal/port"
	"ghost-migrate/internal/usecase"
)

var version = "dev"

func main() {
	var (
		mediaPath       string
		outputDir       string
		status          string
		includePages    bool
		internalTags    bool
		discardOrphaned bool
		dryRun          bool
		yes             bool
		verbose         bool
		showVersion     bool
	)

	flag.StringVar(&mediaPath, "media", "", "path to Ghost media backup (ZIP file or directory)")
	flag.StringVar(&outputDir, "output", "./output", "output directory")
	flag.StringVar(&status, "status", "", "filter by post status (published/draft/scheduled)")
	flag.BoolVar(&includePages, "include-pages", true, "include pages in output")
	flag.BoolVar(&internalTags, "internal-tags", false, "include internal tags (#-prefixed) in front matter")
	flag.BoolVar(&discardOrphaned, "discard-orphaned", false, "discard unreferenced media instead of saving to _orphaned/")
	flag.BoolVar(&dryRun, "dry-run", false, "show migration plan without writing files")
	flag.BoolVar(&yes, "yes", false, "skip confirmation prompt")
	flag.BoolVar(&verbose, "verbose", false, "enable debug logging")
	flag.BoolVar(&showVersion, "version", false, "show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ghost-migrate [flags] <ghost-export.json>\n\n")
		fmt.Fprintf(os.Stderr, "Migrate Ghost CMS exports to Hugo Page Bundles with media resolution.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("ghost-migrate %s\n", version)
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputPath := flag.Arg(0)

	if verbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	// Manual DI
	reader := ghostjson.NewReader()
	converter := htmlconv.NewConverter()
	writer := fsbundle.NewWriter(outputDir)
	journalStore := fsjournal.NewStore(outputDir)

	var mediaStore port.MediaStore
	if mediaPath != "" {
		var err error
		mediaStore, err = openMediaStore(mediaPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer mediaStore.Close()
	}

	reporter := &consoleReporter{verbose: verbose}

	uc := usecase.NewMigrateUseCase(reader, converter, mediaStore, writer, journalStore, reporter)

	opts := usecase.MigrateOptions{
		InputPath:           inputPath,
		OutputDir:           outputDir,
		MediaPath:           mediaPath,
		Status:              status,
		IncludePages:        includePages,
		IncludeInternalTags: internalTags,
		DiscardOrphaned:     discardOrphaned,
		DryRun:              dryRun,
	}

	// Dry-run: show plan
	if dryRun {
		plan, err := uc.Execute(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		printPlan(plan)
		return
	}

	// Confirmation prompt
	if !yes {
		mode := "flat Markdown"
		if mediaPath != "" {
			mode = "Hugo Page Bundles"
		}
		fmt.Printf("Migrate %s → %s (%s mode)\n", inputPath, outputDir, mode)
		fmt.Print("Continue? [y/N] ")

		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	plan, err := uc.Execute(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printSummary(plan)
}

func openMediaStore(path string) (port.MediaStore, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access media path: %w", err)
	}

	if info.IsDir() {
		slog.Debug("opening directory media store", "path", path)
		return mediafs.NewStore(path)
	}

	slog.Debug("opening ZIP media store", "path", path)
	return mediazip.NewStore(path)
}

type consoleReporter struct {
	verbose bool
}

func (r *consoleReporter) OnPostProcessed(slug string, mediaCount int) {
	if r.verbose {
		if mediaCount > 0 {
			fmt.Printf("  ✓ %s (%d media files)\n", slug, mediaCount)
		} else {
			fmt.Printf("  ✓ %s\n", slug)
		}
	}
}

func (r *consoleReporter) OnComplete(stats usecase.MigrateStats) {
	// Summary printed separately
}

func printPlan(plan *domain.MigrationPlan) {
	if plan == nil {
		return
	}
	fmt.Printf("\n=== Migration Plan (dry-run) ===\n")
	fmt.Printf("Posts:          %d\n", plan.Stats.TotalPosts)
	fmt.Printf("Media files:   %d\n", plan.Stats.TotalMedia)
	fmt.Printf("Shared media:  %d\n", plan.Stats.SharedMedia)
	fmt.Printf("Orphaned media: %d\n", plan.Stats.OrphanedMedia)

	fmt.Printf("\nBundles:\n")
	for _, b := range plan.Bundles {
		fmt.Printf("  %s/ (%d media)\n", b.Slug, len(b.MediaEntries))
	}

	if len(plan.OrphanedMedia) > 0 {
		fmt.Printf("\nOrphaned media → _orphaned/:\n")
		for _, o := range plan.OrphanedMedia {
			fmt.Printf("  %s (%d bytes)\n", o.Path, o.Size)
		}
	}
}

func printSummary(plan *domain.MigrationPlan) {
	if plan == nil {
		return
	}
	fmt.Printf("\nMigration complete.\n")
	fmt.Printf("  Posts:   %d\n", plan.Stats.TotalPosts)
	if plan.Stats.TotalMedia > 0 {
		fmt.Printf("  Media:   %d (%d shared)\n", plan.Stats.TotalMedia, plan.Stats.SharedMedia)
	}
	if plan.Stats.OrphanedMedia > 0 {
		fmt.Printf("  Orphans: %d\n", plan.Stats.OrphanedMedia)
	}
}
