package util

import (
	"fmt"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/tui/styles"
	"github.com/charmbracelet/crush/internal/vcs"
)

// VCSInfo returns a styled string representing the current VCS status and
// branch/change name. Returns empty string if no VCS is detected.
func VCSInfo() string {
	workingDir := config.Get().WorkingDir()
	detector := vcs.NewDetector()
	info, err := detector.Detect(workingDir)
	if err != nil || info.Type == vcs.TypeNone {
		return ""
	}

	t := styles.CurrentTheme()

	// Determine the status icon and render it with the appropriate color based on Git status (priority order).
	var styledIcon string

	if info.Type == vcs.TypeGit {
		status := info.Status
		switch {
		case status.HasConflicts:
			styledIcon = t.S().Base.Foreground(t.Error).Render(styles.GitConflictIcon)
		case status.IsDetached:
			styledIcon = t.S().Base.Foreground(t.Warning).Render(styles.GitDetachedIcon)
		case status.HasStaged:
			styledIcon = t.S().Base.Foreground(t.Warning).Render(styles.GitStagedIcon)
		case status.HasUncommitted:
			styledIcon = t.S().Base.Foreground(t.Warning).Render(styles.GitDirtyIcon)
		case status.HasUntracked:
			styledIcon = t.S().Base.Foreground(t.FgSubtle).Render(styles.GitUntrackedIcon)
		case status.AheadCount > 0 && status.BehindCount > 0:
			styledIcon = t.S().Base.Foreground(t.Warning).Render(styles.GitDivergentIcon)
		case status.HasUnpushed || status.AheadCount > 0:
			styledIcon = t.S().Base.Foreground(t.Info).Render(styles.GitUnpushedIcon)
		case status.BehindCount > 0:
			styledIcon = t.S().Base.Foreground(t.Info).Render(styles.GitBehindIcon)
		default:
			// Clean repository - everything committed and pushed.
			styledIcon = t.S().Base.Foreground(t.Success).Render(styles.GitCleanIcon)
		}
	} else if info.Type == vcs.TypeJujutsu {
		status := info.Status
		// Apply similar status logic for Jujutsu.
		switch {
		case status.HasConflicts:
			styledIcon = t.S().Base.Foreground(t.Error).Render(styles.GitConflictIcon)
		case status.HasUncommitted:
			styledIcon = t.S().Base.Foreground(t.Warning).Render(styles.GitDirtyIcon)
		default:
			// Clean or unknown state - use jj icon.
			styledIcon = t.S().Base.Foreground(t.Success).Render("jj")
		}
	} else {
		styledIcon = t.S().Base.Foreground(t.FgMuted).Render(string(info.Type))
	}

	// Display branch/change name for Git and Jujutsu, repo name for other VCS.
	var displayName string
	if (info.Type == vcs.TypeGit || info.Type == vcs.TypeJujutsu) && info.Status.CurrentBranch != "" {
		displayName = info.Status.CurrentBranch
	} else {
		displayName = info.RepoName
	}

	styledName := t.S().Muted.Render(displayName)

	return fmt.Sprintf("%s %s", styledIcon, styledName)
}
