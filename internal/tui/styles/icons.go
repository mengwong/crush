package styles

const (
	CheckIcon         string = "✓"
	ErrorIcon         string = "×"
	WarningIcon       string = "⚠"
	InfoIcon          string = "ⓘ"
	HintIcon          string = "∵"
	SpinnerIcon       string = "..."
	ArrowRightIcon    string = "→"
	CenterSpinnerIcon string = "⋯"
	LoadingIcon       string = "⟳"
	ImageIcon         string = "■"
	TextIcon          string = "☰"
	ModelIcon         string = "◇"

	// VCS icons
	GitIcon          string = ""
	GitCleanIcon     string = "✓" // Clean working tree, everything committed and pushed
	GitDirtyIcon     string = "✗" // Uncommitted changes
	GitStagedIcon    string = "●" // Staged changes ready to commit
	GitConflictIcon  string = "✖" // Merge conflicts
	GitUnpushedIcon  string = "↑" // Commits not pushed to remote
	GitBehindIcon    string = "↓" // Behind remote
	GitDivergentIcon string = "↕" // Diverged from remote (both ahead and behind)
	GitUntrackedIcon string = "?" // Untracked files
	GitDetachedIcon  string = "⚠" // Detached HEAD state

	// Tool call icons
	ToolPending string = "●"
	ToolSuccess string = "✓"
	ToolError   string = "×"

	BorderThin  string = "│"
	BorderThick string = "▌"

	// Todo icons
	TodoCompletedIcon string = "✓"
	TodoPendingIcon   string = "•"
)

var SelectionIgnoreIcons = []string{
	// CheckIcon,
	// ErrorIcon,
	// WarningIcon,
	// InfoIcon,
	// HintIcon,
	// SpinnerIcon,
	// LoadingIcon,
	// DocumentIcon,
	// ModelIcon,
	//
	// // Tool call icons
	// ToolPending,
	// ToolSuccess,
	// ToolError,

	BorderThin,
	BorderThick,
}
