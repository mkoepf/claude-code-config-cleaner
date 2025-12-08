# CleanClaudeConfig (ccc)

A CLI utility to clean up Claude Code configuration by:

1. **Removing stale project session data** - when project directories no longer exist on disk
2. **Removing orphaned data** - empty sessions, orphan todos, file-history
3. **Deduplicating local config** - removes local settings that mirror global settings

## Features

- **Safe by default** - all destructive operations preview first and require explicit confirmation
- **Dry-run support** - see what would be cleaned without making changes
- **Audit logging** - all deletions are logged to `~/.claude/ccc-audit.log`

## Usage

```bash
ccc clean [--dry-run] [--yes]      # Clean all: stale projects, orphans, config duplicates
ccc clean projects [--dry-run]     # Remove stale project session data
ccc clean orphans [--dry-run]      # Remove orphaned data
ccc clean config [--dry-run]       # Deduplicate local configs against global settings
ccc list projects [--stale-only]   # List all projects with their status
ccc list orphans                   # List orphaned data without removing
```

## Implementation Status

### Phase 1: Core Library (In Progress)

| Component | Status | Description |
|-----------|--------|-------------|
| `internal/claude/sessions.go` | 🔲 Stub | Parse session files, extract cwd |
| `internal/claude/paths.go` | 🔲 Stub | Discover Claude directories |
| `internal/claude/projects.go` | 🔲 Stub | Scan and analyze projects |
| `internal/claude/config.go` | ⬜ Not started | Parse settings files |

### Phase 2: UI Components (In Progress)

| Component | Status | Description |
|-----------|--------|-------------|
| `internal/ui/preview.go` | 🔲 Stub | Preview display formatting |
| `internal/ui/confirm.go` | 🔲 Stub | Confirmation prompts |
| `internal/ui/audit.go` | ⬜ Not started | Audit trail logging |

### Phase 3: Cleanup Operations (Not Started)

| Component | Status | Description |
|-----------|--------|-------------|
| `internal/cleaner/stale.go` | ⬜ Not started | Find and clean stale projects |
| `internal/cleaner/orphans.go` | ⬜ Not started | Find and clean orphans |
| `internal/cleaner/dedup.go` | ⬜ Not started | Config deduplication |

### Phase 4: CLI Interface (Stub)

| Component | Status | Description |
|-----------|--------|-------------|
| `cmd/ccc/main.go` | 🔲 Stub | Basic CLI structure |

**Legend:** ✅ Complete | 🔲 Stub (tests written, not implemented) | ⬜ Not started

## Development

Tests are written before implementation (TDD). Run tests with:

```bash
go test ./...
```

## Claude Code Directory Layout

The tool works with the standard Claude Code directory structure:

```
~/.claude/
├── settings.json          # Global settings
├── projects/              # Session data per project
│   └── {encoded-path}/    # e.g., -Users-mhk-Code-myproject
│       └── *.jsonl        # Session files (JSON Lines format)
├── todos/                 # Todo tracking files
├── file-history/          # File version history
└── session-env/           # Session environment
```
