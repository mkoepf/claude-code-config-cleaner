package claude

import (
	"time"
)

// SessionInfo contains metadata extracted from a session file.
type SessionInfo struct {
	ID        string
	CWD       string
	Timestamp time.Time
	FilePath  string
	Size      int64
	IsEmpty   bool
}

// ParseSessionFile reads a session JSONL file and extracts metadata.
func ParseSessionFile(path string) (*SessionInfo, error) {
	panic("not implemented")
}

// ExtractCWD reads the first valid line from session files in a project directory
// and returns the cwd field.
func ExtractCWD(projectDir string) (string, error) {
	panic("not implemented")
}
