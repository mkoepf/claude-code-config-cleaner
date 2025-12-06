package ui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreview_TotalSize(t *testing.T) {
	preview := &Preview{
		Title: "Test",
		Changes: []Change{
			{Action: ActionDelete, Path: "/path1", Size: 1000},
			{Action: ActionDelete, Path: "/path2", Size: 2000},
			{Action: ActionDelete, Path: "/path3", Size: 500},
		},
	}

	total := preview.TotalSize()

	assert.Equal(t, int64(3500), total)
}

func TestPreview_TotalSize_Empty(t *testing.T) {
	preview := &Preview{
		Title:   "Test",
		Changes: []Change{},
	}

	total := preview.TotalSize()

	assert.Equal(t, int64(0), total)
}

func TestPreview_Display_ShowsTitle(t *testing.T) {
	preview := &Preview{
		Title: "Test Preview Title",
		Changes: []Change{
			{Action: ActionDelete, Path: "/test/path", Description: "Test item", Size: 1000},
		},
	}

	var buf bytes.Buffer
	err := preview.Display(&buf)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Test Preview Title")
}

func TestPreview_Display_ShowsChanges(t *testing.T) {
	preview := &Preview{
		Title: "Test",
		Changes: []Change{
			{Action: ActionDelete, Path: "/path/to/delete", Description: "Stale project", Size: 1000},
		},
	}

	var buf bytes.Buffer
	err := preview.Display(&buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "/path/to/delete")
	assert.Contains(t, output, "DELETE")
}

func TestPreview_Display_ShowsKeptItems(t *testing.T) {
	preview := &Preview{
		Title: "Test",
		Changes: []Change{
			{Action: ActionDelete, Path: "/path/to/delete", Size: 1000},
		},
		Kept: []Change{
			{Path: "/path/to/keep", Description: "Active project", Size: 500},
		},
	}

	var buf bytes.Buffer
	err := preview.Display(&buf)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "/path/to/keep")
}

func TestPreview_Display_ShowsTotalSize(t *testing.T) {
	preview := &Preview{
		Title: "Test",
		Changes: []Change{
			{Action: ActionDelete, Path: "/path1", Size: 1024 * 1024},     // 1 MB
			{Action: ActionDelete, Path: "/path2", Size: 2 * 1024 * 1024}, // 2 MB
		},
	}

	var buf bytes.Buffer
	err := preview.Display(&buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "3")
	assert.Contains(t, output, "MB")
}

func TestFormatSize_Bytes(t *testing.T) {
	result := FormatSize(500)

	assert.Equal(t, "500 B", result)
}

func TestFormatSize_Kilobytes(t *testing.T) {
	result := FormatSize(2048)

	assert.Equal(t, "2.0 KB", result)
}

func TestFormatSize_Megabytes(t *testing.T) {
	result := FormatSize(15 * 1024 * 1024)

	assert.Equal(t, "15.0 MB", result)
}

func TestFormatSize_Gigabytes(t *testing.T) {
	result := FormatSize(2 * 1024 * 1024 * 1024)

	assert.Equal(t, "2.0 GB", result)
}

func TestFormatSize_Zero(t *testing.T) {
	result := FormatSize(0)

	assert.Equal(t, "0 B", result)
}
