package cleaner

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mhk/ccc/internal/claude"
	"github.com/mhk/ccc/internal/ui"
)

// DedupResult represents the result of deduplicating a local config.
type DedupResult struct {
	LocalPath      string
	DuplicateAllow []string
	DuplicateDeny  []string
	DuplicateAsk   []string
	SuggestDelete  bool // True if local becomes empty after dedup
}

// HasDuplicates returns true if any duplicate entries were found.
func (r *DedupResult) HasDuplicates() bool {
	return len(r.DuplicateAllow) > 0 ||
		len(r.DuplicateDeny) > 0 ||
		len(r.DuplicateAsk) > 0
}

// TotalDuplicates returns the total number of duplicate entries found.
func (r *DedupResult) TotalDuplicates() int {
	return len(r.DuplicateAllow) + len(r.DuplicateDeny) + len(r.DuplicateAsk)
}

// FindLocalConfigs searches for local .claude/settings.json files under the given path.
func FindLocalConfigs(searchPath string) ([]string, error) {
	var configs []string

	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return configs, nil
	}

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		// Check if this is a .claude/settings.json file
		dir := filepath.Dir(path)
		if filepath.Base(dir) == ".claude" && filepath.Base(path) == "settings.json" {
			configs = append(configs, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return configs, nil
}

// DeduplicateConfig compares local settings against global settings
// and identifies duplicate entries.
func DeduplicateConfig(localPath string, global, local *claude.Settings) *DedupResult {
	result := &DedupResult{
		LocalPath: localPath,
	}

	// Find duplicates in each permission list
	result.DuplicateAllow = findDuplicates(local.Permissions.Allow, global.Permissions.Allow)
	result.DuplicateDeny = findDuplicates(local.Permissions.Deny, global.Permissions.Deny)
	result.DuplicateAsk = findDuplicates(local.Permissions.Ask, global.Permissions.Ask)

	// Check if local would become empty after removing duplicates
	uniqueSettings := local.Diff(global)
	result.SuggestDelete = uniqueSettings.IsEmpty()

	return result
}

// findDuplicates returns entries in local that also exist in global.
func findDuplicates(local, global []string) []string {
	if len(local) == 0 || len(global) == 0 {
		return nil
	}

	globalSet := make(map[string]struct{}, len(global))
	for _, v := range global {
		globalSet[v] = struct{}{}
	}

	var duplicates []string
	for _, v := range local {
		if _, exists := globalSet[v]; exists {
			duplicates = append(duplicates, v)
		}
	}

	return duplicates
}

// ApplyDedup applies the deduplication result to the local config file.
// If dryRun is true, returns without making changes.
func ApplyDedup(result *DedupResult, dryRun bool) error {
	if dryRun {
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(result.LocalPath); os.IsNotExist(err) {
		return nil
	}

	// If suggest delete, remove the file
	if result.SuggestDelete {
		return os.Remove(result.LocalPath)
	}

	// Otherwise, update the file by removing duplicates
	settings, err := claude.LoadSettings(result.LocalPath)
	if err != nil {
		return err
	}

	// Remove duplicates from each list
	settings.Permissions.Allow = removeEntries(settings.Permissions.Allow, result.DuplicateAllow)
	settings.Permissions.Deny = removeEntries(settings.Permissions.Deny, result.DuplicateDeny)
	settings.Permissions.Ask = removeEntries(settings.Permissions.Ask, result.DuplicateAsk)

	// Write updated settings back
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(result.LocalPath, data, 0644)
}

// removeEntries returns a new slice with specified entries removed.
func removeEntries(slice, toRemove []string) []string {
	if len(slice) == 0 {
		return nil
	}

	removeSet := make(map[string]struct{}, len(toRemove))
	for _, v := range toRemove {
		removeSet[v] = struct{}{}
	}

	var result []string
	for _, v := range slice {
		if _, exists := removeSet[v]; !exists {
			result = append(result, v)
		}
	}

	return result
}

// BuildDedupPreview creates a preview of configs to be deduplicated.
func BuildDedupPreview(results []DedupResult) *ui.Preview {
	preview := &ui.Preview{
		Title: "Config Deduplication",
	}

	for _, r := range results {
		var action ui.Action
		var description string

		if r.SuggestDelete {
			action = ui.ActionDelete
			description = "Empty after deduplication, will be deleted"
		} else {
			action = ui.ActionModify
			description = formatDuplicateDescription(r)
		}

		preview.Changes = append(preview.Changes, ui.Change{
			Action:      action,
			Path:        r.LocalPath,
			Description: description,
			Size:        0, // Config files are typically small
		})
	}

	return preview
}

// formatDuplicateDescription creates a description of duplicates found.
func formatDuplicateDescription(r DedupResult) string {
	total := r.TotalDuplicates()
	if total == 1 {
		return "1 duplicate entry to remove"
	}
	return string(rune(total+'0')) + " duplicate entries to remove"
}
