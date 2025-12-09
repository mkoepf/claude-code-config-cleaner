package claude

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "sessions", name)
}

func TestParseSessionFile_ValidSession(t *testing.T) {
	path := testdataPath(t, "valid.jsonl")

	info, err := ParseSessionFile(path)
	require.NoError(t, err)

	assert.Equal(t, "/Users/testuser/Code/myproject", info.CWD)
	assert.Equal(t, "abc123", info.ID)
	assert.False(t, info.IsEmpty, "expected IsEmpty to be false for valid session")
}

func TestParseSessionFile_EmptyFile(t *testing.T) {
	path := testdataPath(t, "empty.jsonl")

	info, err := ParseSessionFile(path)
	require.NoError(t, err)

	assert.True(t, info.IsEmpty, "expected IsEmpty to be true for empty file")
}

func TestParseSessionFile_MalformedJSON(t *testing.T) {
	path := testdataPath(t, "malformed.jsonl")

	_, err := ParseSessionFile(path)
	assert.Error(t, err, "expected error for malformed JSON")
}

func TestParseSessionFile_MissingCWDField(t *testing.T) {
	path := testdataPath(t, "no_cwd.jsonl")

	_, err := ParseSessionFile(path)
	assert.Error(t, err, "expected error for missing cwd field")
}

func TestParseSessionFile_NonExistentFile(t *testing.T) {
	path := testdataPath(t, "does_not_exist.jsonl")

	_, err := ParseSessionFile(path)
	assert.Error(t, err, "expected error for non-existent file")
}

func TestParseSessionFile_ReturnsFileSize(t *testing.T) {
	path := testdataPath(t, "valid.jsonl")

	info, err := ParseSessionFile(path)
	require.NoError(t, err)

	stat, err := os.Stat(path)
	require.NoError(t, err)

	assert.Equal(t, stat.Size(), info.Size)
}
