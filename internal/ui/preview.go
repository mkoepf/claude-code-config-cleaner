package ui

import (
	"io"
)

// Action represents the type of change.
type Action string

const (
	ActionDelete Action = "DELETE"
	ActionModify Action = "MODIFY"
	ActionCreate Action = "CREATE"
)

// Change represents a single change to be made.
type Change struct {
	Action      Action
	Path        string
	Description string
	Size        int64
}

// Preview represents a set of changes to be previewed and confirmed.
type Preview struct {
	Title   string
	Changes []Change
	Kept    []Change // Items that will NOT be changed (for context)
}

// TotalSize returns the total size of all changes.
func (p *Preview) TotalSize() int64 {
	panic("not implemented")
}

// Display writes a formatted preview to the given writer.
func (p *Preview) Display(w io.Writer) error {
	panic("not implemented")
}

// FormatSize formats a byte size as a human-readable string (e.g., "14 MB").
func FormatSize(bytes int64) string {
	panic("not implemented")
}
