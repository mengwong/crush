// Package vcs provides version control system detection and information
// extraction. It supports multiple VCS systems like Git and Jujutsu.
package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Type represents the type of version control system.
type Type string

const (
	// TypeGit represents a Git repository.
	TypeGit Type = "git"
	// TypeJujutsu represents a Jujutsu repository.
	TypeJujutsu Type = "jj"
	// TypeNone represents no VCS detected.
	TypeNone Type = ""
)

// Status represents the current state of a VCS repository.
type Status struct {
	HasUncommitted   bool // Uncommitted changes (modified/added/deleted files)
	HasUntracked     bool // Untracked files
	HasConflicts     bool // Merge conflicts
	HasStaged        bool // Staged changes ready to commit
	AheadCount       int  // Commits ahead of remote
	BehindCount      int  // Commits behind remote
	CurrentBranch    string
	IsDetached       bool // Detached HEAD state
	HasUnpushed      bool // Has commits not pushed to remote
	RemoteTrackingOK bool // Remote tracking branch exists and is accessible
}

// Info contains information about a VCS repository.
type Info struct {
	Type     Type
	RepoName string
	RootPath string
	Status   Status
}

// Detector is an interface for detecting VCS repositories.
type Detector interface {
	// Detect checks if a VCS repository exists at or above the given path.
	// Returns Info with Type set to TypeNone if no repository is found.
	Detect(path string) (Info, error)
}

// detector implements Detector by checking for multiple VCS types.
type detector struct {
	detectors []Detector
}

// NewDetector creates a new Detector that checks for multiple VCS types
// in priority order (Git, then Jujutsu).
func NewDetector() Detector {
	return &detector{
		detectors: []Detector{
			&gitDetector{},
			&jujutsuDetector{},
		},
	}
}

// Detect tries each VCS detector in order and returns the first match.
func (d *detector) Detect(path string) (Info, error) {
	for _, det := range d.detectors {
		info, err := det.Detect(path)
		if err != nil {
			return Info{}, err
		}
		if info.Type != TypeNone {
			return info, nil
		}
	}
	return Info{Type: TypeNone}, nil
}

// findVCSRoot walks up the directory tree looking for a VCS marker directory.
func findVCSRoot(startPath, markerDir string) (string, bool) {
	path, err := filepath.Abs(startPath)
	if err != nil {
		return "", false
	}

	for {
		vcsPath := filepath.Join(path, markerDir)
		if info, err := os.Stat(vcsPath); err == nil && info.IsDir() {
			return path, true
		}

		parent := filepath.Dir(path)
		if parent == path {
			// Reached the root directory.
			break
		}
		path = parent
	}

	return "", false
}

// extractRepoName extracts a repository name from the root path.
// For example, /path/to/myrepo becomes "myrepo".
func extractRepoName(rootPath string) string {
	return filepath.Base(rootPath)
}

// gitDetector detects Git repositories.
type gitDetector struct{}

// Detect checks for a .git directory.
func (g *gitDetector) Detect(path string) (Info, error) {
	rootPath, found := findVCSRoot(path, ".git")
	if !found {
		return Info{Type: TypeNone}, nil
	}

	// Check if .git is a file (submodule) or directory.
	gitPath := filepath.Join(rootPath, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return Info{Type: TypeNone}, nil
	}

	// Handle git worktrees and submodules (where .git is a file).
	if !info.IsDir() {
		// Read the .git file to find the actual git directory.
		content, err := os.ReadFile(gitPath)
		if err == nil {
			// The file contains something like "gitdir: /path/to/actual/.git"
			line := strings.TrimSpace(string(content))
			if strings.HasPrefix(line, "gitdir: ") {
				// We still use the current directory as the repo root.
				// The .git file indicates this is a valid git working tree.
			}
		}
	}

	status := getGitStatus(rootPath)

	return Info{
		Type:     TypeGit,
		RepoName: extractRepoName(rootPath),
		RootPath: rootPath,
		Status:   status,
	}, nil
}

// getGitStatus retrieves the current status of a Git repository.
func getGitStatus(repoPath string) Status {
	status := Status{}

	// Get current branch and detached HEAD state.
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil {
		status.CurrentBranch = strings.TrimSpace(string(output))
	} else {
		// Check if we're in detached HEAD.
		cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			status.CurrentBranch = strings.TrimSpace(string(output))
			status.IsDetached = true
		}
	}

	// Check for conflicts.
	cmd = exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
		status.HasConflicts = true
	}

	// Check for staged changes.
	cmd = exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		// Non-zero exit means there are staged changes.
		status.HasStaged = true
	}

	// Check for uncommitted changes.
	cmd = exec.Command("git", "diff", "--quiet")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		// Non-zero exit means there are uncommitted changes.
		status.HasUncommitted = true
	}

	// Check for untracked files.
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
		status.HasUntracked = true
	}

	// Get ahead/behind counts if we have a tracking branch.
	if !status.IsDetached && status.CurrentBranch != "" {
		cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			status.RemoteTrackingOK = true
			parts := strings.Fields(strings.TrimSpace(string(output)))
			if len(parts) == 2 {
				// Format is "ahead behind"
				if ahead := strings.TrimSpace(parts[0]); ahead != "0" {
					status.AheadCount = len(ahead) // Simple approximation
					status.HasUnpushed = true
				}
				if behind := strings.TrimSpace(parts[1]); behind != "0" {
					status.BehindCount = len(behind) // Simple approximation
				}
			}
		}
	}

	return status
}

// jujutsuDetector detects Jujutsu repositories.
type jujutsuDetector struct{}

// Detect checks for a .jj directory.
func (j *jujutsuDetector) Detect(path string) (Info, error) {
	rootPath, found := findVCSRoot(path, ".jj")
	if !found {
		return Info{Type: TypeNone}, nil
	}

	status := getJujutsuStatus(rootPath)

	return Info{
		Type:     TypeJujutsu,
		RepoName: extractRepoName(rootPath),
		RootPath: rootPath,
		Status:   status,
	}, nil
}

// getJujutsuStatus retrieves the current status of a Jujutsu repository.
func getJujutsuStatus(repoPath string) Status {
	status := Status{}

	// Get current change/branch information.
	// Use jj log to get the current change with its branches.
	cmd := exec.Command("jj", "log", "-r", "@", "--no-graph", "-T", "branches")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil {
		branches := strings.TrimSpace(string(output))
		if branches != "" {
			// If multiple branches, take the first one.
			branchList := strings.Fields(branches)
			if len(branchList) > 0 {
				status.CurrentBranch = branchList[0]
			}
		}
	}

	// If no branch name found, try to get the change ID.
	if status.CurrentBranch == "" {
		cmd = exec.Command("jj", "log", "-r", "@", "--no-graph", "-T", "change_id.short()")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			changeID := strings.TrimSpace(string(output))
			if changeID != "" {
				status.CurrentBranch = changeID
			}
		}
	}

	// Check for uncommitted changes.
	cmd = exec.Command("jj", "status")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil {
		statusOutput := string(output)
		// Jujutsu shows "Working copy changes:" when there are uncommitted changes.
		if strings.Contains(statusOutput, "Working copy changes:") {
			status.HasUncommitted = true
		}
		// Check for conflicts.
		if strings.Contains(statusOutput, "conflict") {
			status.HasConflicts = true
		}
	}

	return status
}
