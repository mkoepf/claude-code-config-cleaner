# CleanClaudeConfig Implementation Plan

## Overview

A CLI utility to clean up Claude Code configuration by:
1. Removing stale project session data when directories no longer exist
2. Removing orphaned data (empty sessions, orphan todos, file-history)
3. Deduplicating local config that mirrors global settings

## Critical Discovery: Path Resolution

The project directory naming scheme (e.g., `-Users-mhk-Code-bench-cscia-cscia-1`) is **ambiguous** when decoded naively. However, each session JSONL file contains a `cwd` field with the **actual filesystem path**:

```json
{"cwd": "/Users/mhk/Code/bench-cscia/cscia-1", ...}
```

**Solution**: Read `cwd` from session files rather than attempting to decode directory names.

## Data Structures

### Claude Code Directory Layout (`~/.claude/`)

```
~/.claude/
├── settings.json          # Global settings (permissions, etc.)
├── projects/              # Session data per project
│   └── {encoded-path}/    # e.g., -Users-mhk-Code-myproject
│       └── *.jsonl        # Session files (JSON Lines format)
├── todos/                 # Todo tracking files
│   └── {session-uuid}-agent-{agent-uuid}.json
├── file-history/          # File version history
│   └── {session-uuid}/    # Per-session directories
├── session-env/           # Session environment (often empty)
│   └── {session-uuid}/
└── ...
```

### Session JSONL Format

Each line is a JSON object:
```json
{
  "sessionId": "uuid",
  "cwd": "/absolute/path/to/project",
  "timestamp": "2025-12-06T15:00:14.163Z",
  "version": "2.0.60",
  ...
}
```

### Settings JSON Format

```json
{
  "permissions": {
    "allow": ["Bash(git add:*)", ...],
    "deny": [],
    "ask": []
  },
  "alwaysThinkingEnabled": false
}
```

## Implementation

### Phase 1: Core Library

#### 1.1 Path Discovery (`internal/claude/paths.go`)

```go
type Paths struct {
    Root        string // ~/.claude
    Projects    string // ~/.claude/projects
    Todos       string // ~/.claude/todos
    FileHistory string // ~/.claude/file-history
    SessionEnv  string // ~/.claude/session-env
    Settings    string // ~/.claude/settings.json
}

func DiscoverPaths() (*Paths, error)
```

**Test cases**:
- Returns correct paths based on home directory
- Handles missing directories gracefully
- Works with `CLAUDE_HOME` override (if applicable)

#### 1.2 Project Scanner (`internal/claude/projects.go`)

```go
type Project struct {
    EncodedName string   // Directory name: -Users-mhk-Code-ccc
    ActualPath  string   // From cwd field: /Users/mhk/Code/ccc
    SessionIDs  []string // UUIDs of sessions in this project
    TotalSize   int64    // Bytes used by session files
    LastUsed    time.Time
}

func ScanProjects(projectsDir string) ([]Project, error)
func (p *Project) Exists() bool  // Check if ActualPath exists on disk
```

**Test cases**:
- Parses cwd from session files correctly
- Handles empty session files (0 bytes)
- Handles malformed JSONL gracefully
- Calculates size correctly
- Determines last used from timestamps

#### 1.3 Session Parser (`internal/claude/sessions.go`)

```go
type SessionInfo struct {
    ID        string
    CWD       string
    Timestamp time.Time
    FilePath  string
    Size      int64
    IsEmpty   bool
}

func ParseSessionFile(path string) (*SessionInfo, error)
func ExtractCWD(projectDir string) (string, error)
```

**Test cases**:
- Extracts cwd from first valid line
- Handles empty files
- Handles files with no cwd field
- Returns appropriate errors for malformed JSON

#### 1.4 Settings Parser (`internal/claude/config.go`)

```go
type Settings struct {
    Permissions Permissions `json:"permissions"`
}

type Permissions struct {
    Allow []string `json:"allow"`
    Deny  []string `json:"deny"`
    Ask   []string `json:"ask"`
}

func LoadSettings(path string) (*Settings, error)
func (s *Settings) Diff(other *Settings) *Settings // Returns entries in s not in other
```

**Test cases**:
- Parses valid settings.json
- Handles missing file
- Handles empty permissions arrays
- Diff correctly identifies unique entries

### Phase 2: Cleanup Operations

#### 2.1 Stale Project Cleaner (`internal/cleaner/stale.go`)

```go
type StaleResult struct {
    Project     Project
    SizeSaved   int64
    FilesRemoved int
}

func FindStaleProjects(projects []Project) []Project
func CleanStaleProject(project Project, dryRun bool) (*StaleResult, error)
```

**Logic**:
1. For each project, check if `ActualPath` exists on disk
2. If not exists, mark as stale
3. When cleaning: remove entire project directory from `~/.claude/projects/`

**Test cases**:
- Identifies projects where path no longer exists
- Preserves projects where path exists
- Dry run doesn't delete anything
- Reports correct size savings

#### 2.2 Orphan Cleaner (`internal/cleaner/orphans.go`)

```go
type OrphanResult struct {
    Type        string // "session", "todo", "file-history", "session-env"
    Path        string
    SizeSaved   int64
}

func FindOrphans(paths *Paths, validSessionIDs []string) ([]OrphanResult, error)
func CleanOrphans(orphans []OrphanResult, dryRun bool) error
```

**Orphan types**:
1. **Empty session files**: 0-byte .jsonl files in projects
2. **Orphan todos**: Files in `~/.claude/todos/` referencing non-existent sessions
3. **Orphan file-history**: Directories in `~/.claude/file-history/` for deleted sessions
4. **Empty session-env**: Empty directories in `~/.claude/session-env/`

**Test cases**:
- Identifies empty session files
- Identifies todos without matching sessions
- Identifies file-history without matching sessions
- Preserves valid orphan-looking data

#### 2.3 Config Deduplicator (`internal/cleaner/dedup.go`)

```go
type DedupResult struct {
    LocalPath       string
    RemovedEntries  []string
    RemainingEntries int
    SuggestDelete   bool // True if local becomes empty
}

func FindLocalConfigs(projectPath string) ([]string, error)
func DeduplicateConfig(globalSettings, localSettings *Settings) *DedupResult
func ApplyDedup(result *DedupResult, dryRun bool) error
```

**Logic**:
1. Load global `~/.claude/settings.json`
2. Find local `.claude/settings.json` or project settings
3. Remove entries from local that exist in global
4. If local becomes empty, suggest deletion

**Test cases**:
- Identifies duplicate allow entries
- Preserves unique local entries
- Handles empty local config
- Suggests deletion when appropriate

### Phase 3: CLI Interface

#### 3.1 Commands (`cmd/ccc/`)

```
ccc clean [--dry-run] [--interactive]
    Clean all: stale projects, orphans, and config duplicates

ccc clean projects [--dry-run] [--force]
    Remove stale project session data

ccc clean orphans [--dry-run]
    Remove orphaned data (empty sessions, orphan todos, etc.)

ccc clean config [--dry-run]
    Deduplicate local configs against global settings

ccc list projects [--stale-only]
    List all projects with their status

ccc list orphans
    List orphaned data without removing
```

#### 3.2 Output Format

```
$ ccc clean projects --dry-run

Stale Projects Found:
  /Users/mhk/Code/ghcrctl (deleted)
    Session data: 14 MB, 39 files
    Would remove: ~/.claude/projects/-Users-mhk-Code-ghcrctl/

  /Users/mhk/ghcrctl (deleted)
    Session data: 48 MB, 89 files
    Would remove: ~/.claude/projects/-Users-mhk-ghcrctl/

Total: 162 MB would be freed

Run without --dry-run to delete.
```

### Phase 4: Safety Features (CRITICAL)

**Principle: Every destructive operation MUST preview first and require explicit confirmation.**

#### 4.1 Mandatory Preview-Confirm Flow

All delete/write operations follow this pattern:

```
1. Analyze and collect all changes
2. Display detailed preview of what will happen
3. Prompt for explicit confirmation (default: No)
4. Only proceed if user confirms with 'y' or 'yes'
```

#### 4.2 Preview Display

```
$ ccc clean projects

=== PREVIEW: Stale Project Cleanup ===

The following projects will be DELETED (directory no longer exists):

  1. /Users/mhk/Code/ghcrctl
     Session data: ~/.claude/projects/-Users-mhk-Code-ghcrctl/
     Size: 14 MB (39 files)

  2. /Users/mhk/ghcrctl
     Session data: ~/.claude/projects/-Users-mhk-ghcrctl/
     Size: 48 MB (89 files)

The following projects will be KEPT (directory exists):

  1. /Users/mhk/Code/ccc
     Session data: ~/.claude/projects/-Users-mhk-Code-ccc/
     Size: 2 MB (4 files)

----------------------------------------
Total to delete: 162 MB (10 projects)
Total to keep:   2 MB (2 projects)
----------------------------------------

Proceed with deletion? [y/N]:
```

#### 4.3 Confirmation Requirements

1. **Default is always No** - pressing Enter without input cancels
2. **Only 'y' or 'yes' (case-insensitive) proceeds**
3. **Any other input cancels with message**: "Aborted. No changes made."
4. **Non-interactive mode**: `--yes` flag to skip confirmation (for scripting), but still prints preview

#### 4.4 Implementation (`internal/ui/confirm.go`)

```go
type Change struct {
    Action      string // "DELETE", "MODIFY", "CREATE"
    Path        string
    Description string
    Size        int64
}

type Preview struct {
    Title   string
    Changes []Change
    Kept    []Change // Items that will NOT be changed (for context)
}

func (p *Preview) Display(w io.Writer)
func ConfirmChanges(preview *Preview, autoYes bool) (bool, error)
```

#### 4.5 Safety Flags

| Flag | Behavior |
|------|----------|
| (default) | Preview + confirmation prompt |
| `--yes` | Preview + auto-confirm (for scripts) |
| `--dry-run` | Preview only, no confirmation prompt, no changes |

#### 4.6 Error Handling

- If preview generation fails, abort before showing anything
- If confirmation read fails (e.g., pipe closed), abort
- If any delete operation fails mid-way, stop and report what was/wasn't deleted
- Never silently skip errors

#### 4.7 Audit Trail

Every run logs to `~/.claude/ccc-audit.log`:
```
2025-12-06T16:00:00Z DELETE ~/.claude/projects/-Users-mhk-ghcrctl/ (48 MB)
2025-12-06T16:00:01Z DELETE ~/.claude/projects/-Users-mhk-Code-ghcrctl/ (14 MB)
```

## File Structure

```
ccc/
├── cmd/
│   └── ccc/
│       ├── main.go
│       ├── clean.go      # Clean subcommand
│       └── list.go       # List subcommand
├── internal/
│   ├── claude/
│   │   ├── paths.go
│   │   ├── paths_test.go
│   │   ├── projects.go
│   │   ├── projects_test.go
│   │   ├── sessions.go
│   │   ├── sessions_test.go
│   │   ├── config.go
│   │   └── config_test.go
│   ├── cleaner/
│   │   ├── stale.go
│   │   ├── stale_test.go
│   │   ├── orphans.go
│   │   ├── orphans_test.go
│   │   ├── dedup.go
│   │   └── dedup_test.go
│   └── ui/
│       ├── preview.go    # Preview display formatting
│       ├── preview_test.go
│       ├── confirm.go    # Confirmation prompts
│       ├── confirm_test.go
│       ├── audit.go      # Audit trail logging
│       └── audit_test.go
├── testdata/
│   ├── projects/         # Sample project directories
│   ├── sessions/         # Sample session files
│   └── settings/         # Sample config files
├── scripts/
│   └── code_quality.sh
├── go.mod
├── go.sum
├── README.md
└── plan.md
```

## Implementation Order (TDD)

1. **`internal/claude/sessions.go`** - Parse session files, extract cwd
2. **`internal/claude/paths.go`** - Discover Claude directories
3. **`internal/claude/projects.go`** - Scan and analyze projects
4. **`internal/ui/preview.go`** - Preview display formatting (safety-critical)
5. **`internal/ui/confirm.go`** - Confirmation prompts (safety-critical)
6. **`internal/ui/audit.go`** - Audit trail logging
7. **`internal/cleaner/stale.go`** - Find and clean stale projects
8. **`internal/claude/config.go`** - Parse settings files
9. **`internal/cleaner/orphans.go`** - Find and clean orphans
10. **`internal/cleaner/dedup.go`** - Config deduplication
11. **`cmd/ccc/`** - CLI interface

## Testing Strategy

### Level 1: Unit Tests

Each package has unit tests using test fixtures in `testdata/`.

**Test fixtures** (`testdata/`):
```
testdata/
├── sessions/
│   ├── valid.jsonl           # Valid session with cwd
│   ├── empty.jsonl           # 0 bytes
│   ├── malformed.jsonl       # Invalid JSON
│   ├── no_cwd.jsonl          # Valid JSON but missing cwd field
│   └── multi_line.jsonl      # Multiple messages
├── settings/
│   ├── global.json           # Full global settings
│   ├── local_duplicate.json  # All entries duplicate global
│   ├── local_partial.json    # Some entries duplicate, some unique
│   └── local_unique.json     # No duplicates
└── projects/
    └── (created dynamically in tests)
```

**Unit test examples**:
```go
// sessions_test.go
func TestExtractCWD_ValidSession(t *testing.T)
func TestExtractCWD_EmptyFile(t *testing.T)
func TestExtractCWD_MalformedJSON(t *testing.T)
func TestExtractCWD_MissingCWDField(t *testing.T)

// confirm_test.go
func TestConfirm_YesInput(t *testing.T)
func TestConfirm_NoInput(t *testing.T)
func TestConfirm_EmptyInput_DefaultsNo(t *testing.T)
func TestConfirm_InvalidInput_Aborts(t *testing.T)

// stale_test.go
func TestFindStaleProjects_AllStale(t *testing.T)
func TestFindStaleProjects_NoneStale(t *testing.T)
func TestFindStaleProjects_Mixed(t *testing.T)
```

### Level 2: Integration Tests

Tests that verify multiple components work together using temporary directories.

```go
// internal/cleaner/integration_test.go

func TestCleanStaleProjects_Integration(t *testing.T) {
    // Setup: create temp ~/.claude structure
    tmpHome := t.TempDir()
    claudeDir := filepath.Join(tmpHome, ".claude")
    projectsDir := filepath.Join(claudeDir, "projects")

    // Create a "stale" project (cwd doesn't exist)
    staleProject := filepath.Join(projectsDir, "-tmp-deleted-project")
    os.MkdirAll(staleProject, 0755)
    writeSession(t, staleProject, "/tmp/deleted-project")

    // Create a "valid" project (cwd exists)
    validPath := t.TempDir() // This exists
    validProject := filepath.Join(projectsDir, encodePath(validPath))
    os.MkdirAll(validProject, 0755)
    writeSession(t, validProject, validPath)

    // Run cleaner
    result, err := cleaner.CleanStaleProjects(projectsDir, false)

    // Verify: stale removed, valid kept
    assert.NoDirExists(t, staleProject)
    assert.DirExists(t, validProject)
}
```

### Level 3: End-to-End Tests (Docker)

Full system tests using Docker to simulate a realistic environment.

#### Docker Test Environment

```dockerfile
# test/Dockerfile
FROM golang:1.23-alpine

# Create test user
RUN adduser -D testuser
USER testuser
WORKDIR /home/testuser

# Pre-populate a realistic ~/.claude structure
COPY --chown=testuser:testuser test/fixtures/claude-home /home/testuser/.claude

# Create some project directories that "exist"
RUN mkdir -p /home/testuser/Code/active-project
RUN mkdir -p /home/testuser/Code/another-active

# These directories are intentionally NOT created (simulating deleted projects):
# - /home/testuser/Code/deleted-project
# - /home/testuser/old-stuff/archived

COPY --chown=testuser:testuser . /app
WORKDIR /app

ENTRYPOINT ["go", "test", "-v", "./test/e2e/..."]
```

#### E2E Test Fixtures

```
test/fixtures/claude-home/
├── settings.json
├── projects/
│   ├── -home-testuser-Code-active-project/      # EXISTS
│   │   └── session1.jsonl                        # cwd: /home/testuser/Code/active-project
│   ├── -home-testuser-Code-another-active/      # EXISTS
│   │   └── session2.jsonl                        # cwd: /home/testuser/Code/another-active
│   ├── -home-testuser-Code-deleted-project/     # STALE (dir doesn't exist)
│   │   ├── session3.jsonl                        # cwd: /home/testuser/Code/deleted-project
│   │   └── session4.jsonl
│   └── -home-testuser-old-stuff-archived/       # STALE (dir doesn't exist)
│       └── session5.jsonl                        # cwd: /home/testuser/old-stuff/archived
├── todos/
│   ├── session1-agent-xxx.json                  # Valid (session1 exists)
│   ├── session3-agent-yyy.json                  # Orphan (session3 in stale project)
│   └── orphan-session-agent-zzz.json            # Orphan (no matching session)
├── file-history/
│   ├── session1/                                # Valid
│   ├── session3/                                # Orphan
│   └── nonexistent-session/                     # Orphan
└── session-env/
    ├── session1/                                # Valid but empty
    └── session99/                               # Orphan
```

#### E2E Test Cases

```go
// test/e2e/clean_test.go

func TestE2E_CleanProjects(t *testing.T) {
    // Run: ccc clean projects --yes
    cmd := exec.Command("./ccc", "clean", "projects", "--yes")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)

    // Verify output shows preview
    assert.Contains(t, string(output), "DELETED")
    assert.Contains(t, string(output), "/home/testuser/Code/deleted-project")
    assert.Contains(t, string(output), "/home/testuser/old-stuff/archived")

    // Verify stale projects removed
    assert.NoDirExists(t, "/home/testuser/.claude/projects/-home-testuser-Code-deleted-project")
    assert.NoDirExists(t, "/home/testuser/.claude/projects/-home-testuser-old-stuff-archived")

    // Verify active projects preserved
    assert.DirExists(t, "/home/testuser/.claude/projects/-home-testuser-Code-active-project")
    assert.DirExists(t, "/home/testuser/.claude/projects/-home-testuser-Code-another-active")

    // Verify audit log written
    auditLog, _ := os.ReadFile("/home/testuser/.claude/ccc-audit.log")
    assert.Contains(t, string(auditLog), "DELETE")
}

func TestE2E_CleanProjects_DryRun(t *testing.T) {
    // Run: ccc clean projects --dry-run
    cmd := exec.Command("./ccc", "clean", "projects", "--dry-run")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)

    // Verify preview shown
    assert.Contains(t, string(output), "PREVIEW")
    assert.Contains(t, string(output), "deleted-project")

    // Verify NOTHING was deleted
    assert.DirExists(t, "/home/testuser/.claude/projects/-home-testuser-Code-deleted-project")
    assert.DirExists(t, "/home/testuser/.claude/projects/-home-testuser-old-stuff-archived")
}

func TestE2E_CleanProjects_Confirmation_No(t *testing.T) {
    // Run with "n" input
    cmd := exec.Command("./ccc", "clean", "projects")
    cmd.Stdin = strings.NewReader("n\n")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)

    // Verify aborted
    assert.Contains(t, string(output), "Aborted")

    // Verify NOTHING was deleted
    assert.DirExists(t, "/home/testuser/.claude/projects/-home-testuser-Code-deleted-project")
}

func TestE2E_CleanOrphans(t *testing.T) {
    // First clean stale projects
    exec.Command("./ccc", "clean", "projects", "--yes").Run()

    // Then clean orphans
    cmd := exec.Command("./ccc", "clean", "orphans", "--yes")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)

    // Verify orphan todos removed
    assert.NoFileExists(t, "/home/testuser/.claude/todos/session3-agent-yyy.json")
    assert.NoFileExists(t, "/home/testuser/.claude/todos/orphan-session-agent-zzz.json")

    // Verify valid todos preserved
    assert.FileExists(t, "/home/testuser/.claude/todos/session1-agent-xxx.json")
}

func TestE2E_FullClean(t *testing.T) {
    // Run full cleanup
    cmd := exec.Command("./ccc", "clean", "--yes")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)

    // Verify summary
    assert.Contains(t, string(output), "Freed:")

    // Calculate remaining size - should only be active projects
    // (implementation detail: verify significant reduction)
}
```

### Level 4: Safety Tests

Specific tests to verify safety guarantees.

```go
// test/safety/safety_test.go

func TestSafety_NeverDeletesExistingProject(t *testing.T) {
    // Create 100 projects with random paths
    // Ensure ALL paths exist on disk
    // Run cleaner
    // Verify ZERO projects deleted
}

func TestSafety_DefaultConfirmationIsNo(t *testing.T) {
    // Send empty input (just Enter)
    cmd := exec.Command("./ccc", "clean", "projects")
    cmd.Stdin = strings.NewReader("\n")
    output, _ := cmd.CombinedOutput()

    assert.Contains(t, string(output), "Aborted")
}

func TestSafety_AuditLogAlwaysWritten(t *testing.T) {
    // Run cleanup
    // Verify audit log exists and contains entries
    // Even on partial failure, log should have entries
}

func TestSafety_PartialFailureStops(t *testing.T) {
    // Make one project directory read-only (can't delete)
    // Run cleaner
    // Verify it stops and reports the error
    // Verify projects after the failure are NOT deleted
}

func TestSafety_PreviewMatchesAction(t *testing.T) {
    // Run with --dry-run, capture preview
    // Run with --yes, capture what was deleted
    // Verify preview exactly matches what was deleted
}
```

### Running Tests

```bash
# Unit tests
go test ./...

# Integration tests (requires temp dirs)
go test ./internal/... -tags=integration

# E2E tests (requires Docker)
make test-e2e
# or
docker build -t ccc-test -f test/Dockerfile .
docker run --rm ccc-test

# Safety tests
go test ./test/safety/... -tags=safety

# All tests with coverage
make test-all
```

### Makefile Targets

```makefile
.PHONY: test test-unit test-integration test-e2e test-safety test-all

test: test-unit

test-unit:
	go test ./internal/... ./cmd/...

test-integration:
	go test ./internal/... -tags=integration

test-e2e:
	docker build -t ccc-test -f test/Dockerfile .
	docker run --rm ccc-test

test-safety:
	go test ./test/safety/... -tags=safety -v

test-all: test-unit test-integration test-e2e test-safety
	@echo "All tests passed"
```

### CI Pipeline

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go test ./...

  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go test ./... -tags=integration

  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker build -t ccc-test -f test/Dockerfile .
      - run: docker run --rm ccc-test

  safety:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go test ./test/safety/... -tags=safety -v
```

## Success Criteria

- [ ] Correctly identifies stale projects by reading cwd from sessions
- [ ] Never deletes projects that still exist on disk
- [ ] Dry run shows accurate preview without modifications
- [ ] Frees expected disk space when run
- [ ] All tests pass
- [ ] Code quality checks pass (scripts/code_quality.sh)
