package ui

import (
	"io"
)

// ConfirmResult represents the result of a confirmation prompt.
type ConfirmResult int

const (
	ConfirmYes ConfirmResult = iota
	ConfirmNo
	ConfirmError
)

// Confirmer handles user confirmation prompts.
type Confirmer struct {
	In  io.Reader
	Out io.Writer
}

// Confirm prompts the user for confirmation and returns the result.
// Default is No (pressing Enter without input returns ConfirmNo).
// Only "y" or "yes" (case-insensitive) returns ConfirmYes.
func (c *Confirmer) Confirm(prompt string) ConfirmResult {
	panic("not implemented")
}

// ConfirmChanges displays a preview and prompts for confirmation.
// If autoYes is true, it displays the preview but skips the prompt.
func ConfirmChanges(preview *Preview, in io.Reader, out io.Writer, autoYes bool) (bool, error) {
	panic("not implemented")
}
