# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ghost-migrate is a Go CLI tool for complete migration from Ghost CMS. It parses Ghost JSON exports into Markdown files and resolves media from Ghost Support backup archives (ZIP or directory), outputting Hugo Page Bundle-compatible directory structures. Without `--media`, it falls back to flat Markdown output (ghost-to-md compatible).

## Build & Test Commands

```bash
make build              # Build binary to bin/ghost-migrate
make test               # go test -v -race -count=1 ./...
make lint               # go vet ./... && gofmt -l .
make clean              # rm -rf bin/ output/
make update-golden      # UPDATE_GOLDEN=1 go test ./internal/adapter/htmlconv/...

# Run a single test
go test -v -run TestName ./internal/domain/...

# Run tests for a specific package
go test -v -race ./internal/adapter/ghostjson/...
```

## Architecture

Clean Architecture (Ports & Adapters) with strict dependency direction: outer layers depend on inner layers, never the reverse.

### Layer Structure

```
cmd/ghost-migrate/main.go          # Entry point, manual DI (no DI framework), flag parsing
internal/
  domain/                          # Pure Go, zero external dependencies
    model.go                       # Post, Tag, Author (inherited from ghost-to-md)
    resolver.go                    # RelationshipResolver (posts_tags, posts_authors)
    frontmatter.go                 # YAML front matter generation
    naming.go                      # NamingStrategy for output filenames
    mediaref.go                    # MediaReference, MediaMapping, ExtractMediaRefs
    bundle.go                      # PageBundle, BundleMediaEntry
    journal.go                     # MoveJournal for crash-safe file ops
    plan.go                        # MigrationPlan (output of analysis phase)
    orphan.go                      # OrphanDetector
  port/                            # Interfaces defined by consumers (usecase layer)
    reader.go                      # ExportReader
    converter.go                   # ContentConverter
    mediastore.go                  # MediaStore
    bundlewriter.go                # BundleWriter
    journalstore.go                # JournalStore
  adapter/                         # Concrete implementations of ports
    ghostjson/                     # Ghost JSON parser (reader.go, schema.go)
    htmlconv/                      # HTML-to-Markdown converter with Ghost card support
    mediafs/                       # Directory-based MediaStore
    mediazip/                      # ZIP-based MediaStore (streaming, no full extraction)
    fsbundle/                      # Page Bundle writer
    fsjournal/                     # Journal persistence (manifest.json)
  usecase/
    migrate.go                     # MigrateUseCase orchestrator
    analyzer.go                    # Phase 1: read-only analysis, builds MigrationPlan
    executor.go                    # Phase 2: executes plan with journal tracking
    options.go                     # MigrateOptions
```

### Two-Phase Execution Model

- **Phase 1 (Analysis)**: Read-only. Parses JSON, indexes media backup, extracts media references from HTML, builds mappings, detects orphans, generates MigrationPlan. No filesystem writes.
- **Phase 2 (Execution)**: Executes MigrationPlan. All file operations recorded in a journal (manifest.json) for crash recovery and idempotent re-execution.

### Key Design Patterns

- **Accept Interfaces, Return Structs**: Interfaces defined in `port/` by consumers, adapters return concrete structs
- **Manual DI in main.go**: No framework; constructor injection wired explicitly
- **Table-driven tests** with golden files in `testdata/`
- **`testing/fstest`** for filesystem mocks; `archive/zip` for dynamic test ZIP generation

## Dependencies

- Single runtime dependency: `golang.org/x/net/html` (HTML5 parser for DOM-based HTML-to-Markdown conversion)
- Tests use only standard `testing` package
- CLI uses standard `flag` package (no cobra)
- YAML front matter is hand-generated (no yaml library)

## CLI Usage

```bash
ghost-migrate [flags] <ghost-export.json>

# Key flags:
#   --media <path>          ZIP or directory of Ghost media backup (auto-detected)
#   --output <path>         Output directory (default: ./output)
#   --status <filter>       Filter by post status (published/draft/scheduled)
#   --include-pages         Include pages (default: true)
#   --internal-tags         Include internal tags (#-prefixed) in front matter
#   --discard-orphaned      Discard unreferenced media instead of saving to _orphaned/
#   --dry-run               Show plan without writing files
#   --yes                   Skip confirmation prompt
#   --verbose               Enable debug logging (slog)
```

## Ghost-Specific Concepts

- **Ghost cards**: HTML elements like `kg-image-card`, `kg-gallery-card`, `kg-bookmark-card`, `kg-code-card`, `kg-callout-card`, `kg-toggle-card`, `kg-embed-card`, `kg-button-card` need special Markdown conversion
- **Media URLs**: Two forms — `__GHOST_URL__/content/images/...` and absolute `https://domain/content/images/...`; both map to `content/{images,media,files}/` paths in backups
- **Shared media**: Same file referenced by multiple posts gets COPYed to each bundle; single-reference files can be MOVEd
- **Orphaned media**: Files in backup not referenced by any post go to `_orphaned/` by default

## Git Workflow

### Branch Strategy

- **Main branch**: `main` — always deployable
- **Feature branches**: `feat/`, `fix/`, `refactor/`, `docs/`, `test/`, `chore/` prefixes, kebab-case (e.g., `feat/ghost-json-parser`, `fix/media-url-edge-case`)
- Always PR to `main`, delete branch after merge

### Commit Conventions

[Conventional Commits](https://www.conventionalcommits.org/) with scopes matching architecture layers:

```
type(scope): description
```

- **Types**: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`
- **Scopes**: `domain`, `port`, `adapter`, `usecase`, `cli`, `ghostjson`, `htmlconv`, `mediafs`, `mediazip`, `fsbundle`, `fsjournal`, `claude`
- Language: English
- One logical change per commit
- Co-author trailer: `Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>`

### PR Conventions

- **Merge strategy**: Squash-and-merge only (merge commits and rebase merges disabled)
- **PR title**: Conventional commit format (becomes the squash commit message)
- **PR body**: Summary (bullet points), Architecture notes (if applicable), Test Plan (checklist)
- Branches auto-deleted after merge

## Development Environment

### 1Password SSH Agent (WSL)

SSH authentication and commit signing are handled by 1Password via WSL bridge:

- **SSH**: `core.sshcommand = ssh.exe` — delegates to Windows OpenSSH, which connects to 1Password's SSH agent
- **Signing**: `gpg.format = ssh` with `gpg.ssh.program = op-ssh-sign-wsl.exe` — 1Password signs commits with the SSH key
- **Key**: Ed25519 key managed in 1Password, configured via `user.signingkey`

#### Troubleshooting

- **Auth/signing fails**: Ensure 1Password desktop app is running on the Windows host (the agent lives in the desktop app)
- **Verify SSH**: `ssh -T git@github.com` (should say "Hi P4suta!")
- **Verify signing**: `git log --show-signature` (should show "Good ssh signature")
