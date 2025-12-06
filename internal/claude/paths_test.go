package claude

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverPaths_DefaultLocation(t *testing.T) {
	paths, err := DiscoverPaths("")
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expectedRoot := filepath.Join(home, ".claude")
	assert.Equal(t, expectedRoot, paths.Root)
	assert.Equal(t, filepath.Join(expectedRoot, "projects"), paths.Projects)
	assert.Equal(t, filepath.Join(expectedRoot, "settings.json"), paths.Settings)
}

func TestDiscoverPaths_CustomLocation(t *testing.T) {
	customHome := "/custom/claude/home"

	paths, err := DiscoverPaths(customHome)
	require.NoError(t, err)

	assert.Equal(t, customHome, paths.Root)
	assert.Equal(t, filepath.Join(customHome, "projects"), paths.Projects)
}

func TestDiscoverPaths_AllPathsPopulated(t *testing.T) {
	paths, err := DiscoverPaths("/test/home")
	require.NoError(t, err)

	assert.NotEmpty(t, paths.Root, "Root path should not be empty")
	assert.NotEmpty(t, paths.Projects, "Projects path should not be empty")
	assert.NotEmpty(t, paths.Todos, "Todos path should not be empty")
	assert.NotEmpty(t, paths.FileHistory, "FileHistory path should not be empty")
	assert.NotEmpty(t, paths.SessionEnv, "SessionEnv path should not be empty")
	assert.NotEmpty(t, paths.Settings, "Settings path should not be empty")
}
