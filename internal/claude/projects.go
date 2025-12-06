package claude

import (
	"time"
)

// Project represents a Claude Code project with its session data.
type Project struct {
	EncodedName string    // Directory name: -Users-mhk-Code-ccc
	ActualPath  string    // From cwd field: /Users/mhk/Code/ccc
	SessionIDs  []string  // UUIDs of sessions in this project
	TotalSize   int64     // Bytes used by session files
	LastUsed    time.Time // Most recent session timestamp
	FileCount   int       // Number of session files
}

// Exists checks if the project's actual path exists on disk.
func (p *Project) Exists() bool {
	panic("not implemented")
}

// ScanProjects scans the projects directory and returns information about each project.
func ScanProjects(projectsDir string) ([]Project, error) {
	panic("not implemented")
}
