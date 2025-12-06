package claude

// Paths contains the standard Claude Code directory paths.
type Paths struct {
	Root        string // ~/.claude
	Projects    string // ~/.claude/projects
	Todos       string // ~/.claude/todos
	FileHistory string // ~/.claude/file-history
	SessionEnv  string // ~/.claude/session-env
	Settings    string // ~/.claude/settings.json
}

// DiscoverPaths returns the Claude Code paths for the current user.
// If claudeHome is empty, it uses the default ~/.claude location.
func DiscoverPaths(claudeHome string) (*Paths, error) {
	panic("not implemented")
}
