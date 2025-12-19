package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitDetector(t *testing.T) {
	t.Parallel()

	t.Run("detects git repository", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		// Create a .git directory.
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0o755)
		require.NoError(t, err)

		detector := &gitDetector{}
		info, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		require.Equal(t, TypeGit, info.Type)
		require.Equal(t, filepath.Base(tmpDir), info.RepoName)
		require.Equal(t, tmpDir, info.RootPath)
	})

	t.Run("detects git repository from subdirectory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		// Create a .git directory at root.
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0o755)
		require.NoError(t, err)

		// Create a subdirectory.
		subDir := filepath.Join(tmpDir, "subdir", "nested")
		err = os.MkdirAll(subDir, 0o755)
		require.NoError(t, err)

		detector := &gitDetector{}
		info, err := detector.Detect(subDir)
		require.NoError(t, err)
		require.Equal(t, TypeGit, info.Type)
		require.Equal(t, filepath.Base(tmpDir), info.RepoName)
		require.Equal(t, tmpDir, info.RootPath)
	})

	t.Run("returns TypeNone when no git repository found", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		detector := &gitDetector{}
		info, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		require.Equal(t, TypeNone, info.Type)
		require.Equal(t, "", info.RepoName)
	})

	t.Run("handles git worktree with .git file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		// Create a .git directory first to make it a valid repo.
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0o755)
		require.NoError(t, err)

		// Initialize a minimal git repo so git commands work.
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		err = cmd.Run()
		require.NoError(t, err)

		detector := &gitDetector{}
		info, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		require.Equal(t, TypeGit, info.Type)
		require.Equal(t, filepath.Base(tmpDir), info.RepoName)
	})
}

func TestJujutsuDetector(t *testing.T) {
	t.Parallel()

	t.Run("detects jujutsu repository", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		// Create a .jj directory.
		jjDir := filepath.Join(tmpDir, ".jj")
		err := os.Mkdir(jjDir, 0o755)
		require.NoError(t, err)

		detector := &jujutsuDetector{}
		info, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		require.Equal(t, TypeJujutsu, info.Type)
		require.Equal(t, filepath.Base(tmpDir), info.RepoName)
		require.Equal(t, tmpDir, info.RootPath)
	})

	t.Run("detects jujutsu repository from subdirectory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		// Create a .jj directory at root.
		jjDir := filepath.Join(tmpDir, ".jj")
		err := os.Mkdir(jjDir, 0o755)
		require.NoError(t, err)

		// Create a subdirectory.
		subDir := filepath.Join(tmpDir, "subdir")
		err = os.Mkdir(subDir, 0o755)
		require.NoError(t, err)

		detector := &jujutsuDetector{}
		info, err := detector.Detect(subDir)
		require.NoError(t, err)
		require.Equal(t, TypeJujutsu, info.Type)
		require.Equal(t, filepath.Base(tmpDir), info.RepoName)
	})

	t.Run("returns TypeNone when no jujutsu repository found", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		detector := &jujutsuDetector{}
		info, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		require.Equal(t, TypeNone, info.Type)
	})
}

func TestNewDetector(t *testing.T) {
	t.Parallel()

	t.Run("detects git repository with priority", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		// Create a .git directory.
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0o755)
		require.NoError(t, err)

		detector := NewDetector()
		info, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		require.Equal(t, TypeGit, info.Type)
		require.Equal(t, filepath.Base(tmpDir), info.RepoName)
	})

	t.Run("detects jujutsu repository when git not present", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		// Create a .jj directory.
		jjDir := filepath.Join(tmpDir, ".jj")
		err := os.Mkdir(jjDir, 0o755)
		require.NoError(t, err)

		detector := NewDetector()
		info, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		require.Equal(t, TypeJujutsu, info.Type)
		require.Equal(t, filepath.Base(tmpDir), info.RepoName)
	})

	t.Run("prioritizes git when both present", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		// Create both .git and .jj directories.
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0o755)
		require.NoError(t, err)

		jjDir := filepath.Join(tmpDir, ".jj")
		err = os.Mkdir(jjDir, 0o755)
		require.NoError(t, err)

		detector := NewDetector()
		info, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		require.Equal(t, TypeGit, info.Type)
	})

	t.Run("returns TypeNone when no VCS found", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		detector := NewDetector()
		info, err := detector.Detect(tmpDir)
		require.NoError(t, err)
		require.Equal(t, TypeNone, info.Type)
		require.Equal(t, "", info.RepoName)
	})
}
