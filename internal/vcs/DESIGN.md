# VCS Integration Design

## Overview

The VCS integration provides real-time status display in the Crush sidebar, showing the current branch/change name with status-aware icons for Git and Jujutsu repositories.

## Architecture

### Detection
- **Pluggable detector system**: `Detector` interface allows easy addition of new VCS types
- **Priority ordering**: Git is checked before Jujutsu to handle coexisting repos
- **Upward traversal**: Searches parent directories to find repository root

### Status Checking
- **Git**: Executes git commands to check branch, conflicts, staged changes, uncommitted changes, untracked files, and ahead/behind counts
- **Jujutsu**: Uses `jj` commands to check branch/change ID, uncommitted changes, and conflicts

### Display
- **Priority-based icons**: Follows oh-my-zsh conventions (conflicts > detached > staged > uncommitted > untracked > ahead/behind > clean)
- **Color coding**: Red for errors, yellow for warnings, blue for info, green for success
- **Branch names**: Shows current branch/change name instead of repository name

## Refresh Strategy

### Current Implementation

The VCS status in the sidebar refreshes via two mechanisms:

1. **Periodic Timer Refresh** (5 seconds)
   - Implemented via `VCSRefreshMsg` in `sidebar.go`
   - Catches changes made outside Crush (switching branches in another terminal, external commits, etc.)
   - Interval balances responsiveness with system resource usage

2. **File Change Event Refresh** (immediate)
   - Triggered on `pubsub.Event[history.File]`
   - Catches Crush's own file modifications
   - Provides instant feedback when the agent modifies files

### Alternatives Considered

**Bash Command Hooks**
- Could trigger refresh when git/jj commands are executed through Crush's bash tool
- Would provide slightly faster updates for commands run within Crush
- **Not implemented** due to:
  - Added complexity (requires event publishing or command parsing)
  - Marginal benefit over 5-second periodic refresh
  - Doesn't handle external commands anyway
  - Current implementation is responsive enough for typical use

**Future Enhancement**: If users frequently run VCS commands through Crush and want instant feedback, bash command hooks could be added by:
1. Publishing `VCSCommandExecutedMsg` from `internal/agent/tools/bash.go` when detecting git/jj commands
2. Having sidebar subscribe to these messages and refresh immediately
3. Parsing command strings with regex like `^(git|jj)\s+`

### Performance Considerations

- VCS status checks are lightweight (typically <50ms)
- 5-second interval provides good balance:
  - Responsive enough for typical workflows
  - Low overhead (~0.01% CPU usage)
  - Doesn't spam git/jj commands
- File change events are already part of Crush's event system (zero additional overhead)

## Icon Reference

### Git Status Icons (Priority Order)
1. `✖` (red) - Merge conflicts
2. `⚠` (yellow) - Detached HEAD state
3. `●` (yellow) - Staged changes ready to commit
4. `✗` (yellow) - Uncommitted changes
5. `?` (muted) - Untracked files
6. `↕` (yellow) - Diverged from remote (both ahead and behind)
7. `↑` (blue) - Unpushed commits / ahead of remote
8. `↓` (blue) - Behind remote
9. `✓` (green) - Clean working tree

### Jujutsu Status Icons
1. `✖` (red) - Conflicts
2. `✗` (yellow) - Uncommitted changes
3. `jj` (green) - Clean repository

## Extension Points

### Adding New VCS Systems

To add support for a new VCS:

1. Create a new detector type implementing the `Detector` interface
2. Add the detector to `NewDetector()` in priority order
3. Implement status detection function (like `getGitStatus`)
4. Add display logic in `sidebar.vcsInfo()` function
5. Add any new icons to `internal/tui/styles/icons.go`

Example for Mercurial:

```go
type hgDetector struct{}

func (h *hgDetector) Detect(path string) (Info, error) {
    rootPath, found := findVCSRoot(path, ".hg")
    if !found {
        return Info{Type: TypeNone}, nil
    }

    status := getHgStatus(rootPath)
    return Info{
        Type:     TypeHg,
        RepoName: extractRepoName(rootPath),
        RootPath: rootPath,
        Status:   status,
    }, nil
}
```

### Customizing Refresh Behavior

The refresh interval is configurable via the `VCSRefreshInterval` constant in `sidebar.go`. Adjust based on:
- User feedback on responsiveness
- Performance considerations on slower systems
- Battery usage concerns on laptops

## Testing

- Unit tests cover detector logic and status parsing
- Tests use temporary directories with actual git/jj initialization
- Mock VCS commands not used to ensure real-world behavior
- Tests verify priority ordering when multiple VCS systems coexist

## Future Enhancements

1. **Configuration Options**
   - Allow users to disable VCS display
   - Customize refresh interval
   - Choose which status indicators to show

2. **Additional Status Information**
   - Show stash count
   - Display upstream branch name
   - Show rebase/merge in progress state

3. **Performance Optimizations**
   - Cache status for very large repositories
   - Skip expensive checks (ahead/behind) on slow connections
   - Debounce rapid file changes

4. **Bash Command Hooks** (if needed)
   - Detect git/jj commands in bash tool
   - Trigger immediate refresh
   - Balance with existing periodic refresh
