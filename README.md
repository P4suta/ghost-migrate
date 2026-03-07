# ghost-migrate

A CLI tool for complete migration from Ghost CMS. Parses Ghost JSON exports into Markdown files and resolves media from Ghost Support backup archives (ZIP or directory), outputting Hugo Page Bundle-compatible directory structures.

Without `--media`, it falls back to flat Markdown output (ghost-to-md compatible).

## Installation

```bash
go install ghost-migrate/cmd/ghost-migrate@latest
```

Or build from source:

```bash
make build
# Binary at bin/ghost-migrate
```

## Usage

```bash
# Flat Markdown (ghost-to-md compatible)
ghost-migrate export.json

# Hugo Page Bundles with media from ZIP backup
ghost-migrate --media backup.zip export.json

# Hugo Page Bundles with media from extracted directory
ghost-migrate --media ./backup-dir export.json

# Dry-run to preview the migration plan
ghost-migrate --media backup.zip --dry-run export.json

# Filter by status
ghost-migrate --status published export.json
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--media <path>` | | ZIP or directory of Ghost media backup |
| `--output <path>` | `./output` | Output directory |
| `--status <filter>` | | Filter by post status (published/draft/scheduled) |
| `--include-pages` | `true` | Include pages in output |
| `--internal-tags` | `false` | Include internal tags (#-prefixed) in front matter |
| `--discard-orphaned` | `false` | Discard unreferenced media instead of saving to `_orphaned/` |
| `--dry-run` | `false` | Show migration plan without writing files |
| `--yes` | `false` | Skip confirmation prompt |
| `--verbose` | `false` | Enable debug logging |

## Output Structure

### With `--media` (Page Bundle mode)

```
output/
  my-post/
    index.md          # Markdown with front matter
    photo.jpg         # Referenced media
    cover.jpg         # Feature image
  another-post/
    index.md
  _orphaned/          # Unreferenced media files
    unused.jpg
  manifest.json       # Journal for crash recovery
```

### Without `--media` (Flat mode)

```
output/
  my-post.md
  another-post.md
```

## Architecture

Clean Architecture with strict dependency direction:

```
cmd/ghost-migrate/     CLI entry point, manual DI
internal/
  domain/              Pure Go types and business logic
  port/                Interfaces (ExportReader, ContentConverter, MediaStore, etc.)
  adapter/             Implementations
    ghostjson/         Ghost JSON parser
    htmlconv/          HTML-to-Markdown with Ghost card support
    mediafs/           Directory-based media store
    mediazip/          ZIP-based media store (streaming)
    fsbundle/          Page Bundle file writer
    fsjournal/         Journal persistence
  usecase/             Two-phase migration orchestrator
```

### Two-Phase Execution

1. **Analysis** (read-only): Parse JSON, index media, extract references, detect shared/orphaned files, build migration plan
2. **Execution**: Write bundles, copy media, track operations in journal for crash recovery

## Development

```bash
make build              # Build binary
make test               # Run tests with race detector
make lint               # go vet + gofmt check
make clean              # Remove build artifacts
make update-golden      # Update golden test files
```
