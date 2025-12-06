package claude

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestProject(t *testing.T, baseDir, encodedName, cwd string) string {
	t.Helper()
	projectDir := filepath.Join(baseDir, encodedName)
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	content := `{"sessionId":"test-session","cwd":"` + cwd + `","timestamp":"2025-12-06T10:00:00Z"}`
	sessionFile := filepath.Join(projectDir, "session.jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(content), 0644))

	return projectDir
}

func TestProject_Exists_True(t *testing.T) {
	existingDir := t.TempDir()

	project := Project{
		EncodedName: "-tmp-existing",
		ActualPath:  existingDir,
	}

	assert.True(t, project.Exists(), "expected Exists() to return true for existing directory")
}

func TestProject_Exists_False(t *testing.T) {
	project := Project{
		EncodedName: "-tmp-nonexistent",
		ActualPath:  "/this/path/does/not/exist/at/all",
	}

	assert.False(t, project.Exists(), "expected Exists() to return false for non-existent directory")
}

func TestScanProjects_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	projects, err := ScanProjects(tmpDir)
	require.NoError(t, err)

	assert.Empty(t, projects)
}

func TestScanProjects_SingleProject(t *testing.T) {
	tmpDir := t.TempDir()
	existingPath := t.TempDir()
	createTestProject(t, tmpDir, "-Users-test-myproject", existingPath)

	projects, err := ScanProjects(tmpDir)
	require.NoError(t, err)

	require.Len(t, projects, 1)
	assert.Equal(t, "-Users-test-myproject", projects[0].EncodedName)
	assert.Equal(t, existingPath, projects[0].ActualPath)
}

func TestScanProjects_MultipleProjects(t *testing.T) {
	tmpDir := t.TempDir()
	path1 := t.TempDir()
	path2 := t.TempDir()
	createTestProject(t, tmpDir, "-Users-test-project1", path1)
	createTestProject(t, tmpDir, "-Users-test-project2", path2)

	projects, err := ScanProjects(tmpDir)
	require.NoError(t, err)

	assert.Len(t, projects, 2)
}

func TestScanProjects_CalculatesSize(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-Users-test-sized")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	content := `{"sessionId":"test","cwd":"/tmp/test","timestamp":"2025-12-06T10:00:00Z"}`
	sessionFile := filepath.Join(projectDir, "session.jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(content), 0644))

	projects, err := ScanProjects(tmpDir)
	require.NoError(t, err)

	require.Len(t, projects, 1)
	assert.Equal(t, int64(len(content)), projects[0].TotalSize)
}

func TestScanProjects_CountsFiles(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-Users-test-multifile")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	content := `{"sessionId":"test","cwd":"/tmp/test","timestamp":"2025-12-06T10:00:00Z"}`

	for i := 0; i < 3; i++ {
		sessionFile := filepath.Join(projectDir, filepath.Base(t.Name())+string(rune('a'+i))+".jsonl")
		require.NoError(t, os.WriteFile(sessionFile, []byte(content), 0644))
	}

	projects, err := ScanProjects(tmpDir)
	require.NoError(t, err)

	require.Len(t, projects, 1)
	assert.Equal(t, 3, projects[0].FileCount)
}

func TestScanProjects_SkipsNonDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "not-a-directory.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("test"), 0644))

	existingPath := t.TempDir()
	createTestProject(t, tmpDir, "-Users-test-valid", existingPath)

	projects, err := ScanProjects(tmpDir)
	require.NoError(t, err)

	assert.Len(t, projects, 1)
}

func TestScanProjects_HandlesEmptySessionFiles(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "-Users-test-empty")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	sessionFile := filepath.Join(projectDir, "empty.jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte{}, 0644))

	projects, err := ScanProjects(tmpDir)
	require.NoError(t, err)

	require.Len(t, projects, 1)
	assert.Empty(t, projects[0].ActualPath, "expected empty actual path for project with only empty session files")
}
