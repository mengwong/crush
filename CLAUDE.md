# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Crush is a terminal-based AI coding assistant that connects users to multiple LLMs through a session-based interface. It's built in Go using Bubble Tea for the TUI and supports LSP integration, MCP servers, and multi-model workflows.

## Build and Development Commands

```bash
# Build
go build .
task build

# Run
go run .
task run

# Test
go test ./...
task test

# Run single test
go test ./internal/llm/prompt -run TestGetContextFromPaths

# Update golden test files (when test output changes intentionally)
go test ./... -update
go test ./internal/tui/components/core -update  # Update specific package

# Lint
task lint
task lint:fix

# Format code
task fmt           # Runs gofumpt -w .
gofumpt -w .       # Direct formatting

# Run with profiling enabled
task dev           # Sets CRUSH_PROFILE=true

# Generate JSON schema
task schema        # Outputs to schema.json
```

## Architecture

### Core Components

**Agent Layer** (`internal/agent/`)
- `coordinator.go`: Orchestrates multiple agents and manages agent lifecycle
- `agent.go`: Core `SessionAgent` implementation handling LLM interactions
- Session-based agent execution with message queuing and context management
- Automatic summarization when context limits are approached
- Tool execution coordination via Fantasy framework

**Session Management** (`internal/session/`)
- Session-based context preservation across conversations
- Multiple concurrent sessions per project
- Session state persistence via SQLite database

**Message System** (`internal/message/`)
- Unified message abstraction for user/assistant/system messages
- Attachment handling (images, files)
- Message storage and retrieval through database layer

**Configuration** (`internal/config/`)
- Multi-level config resolution: `.crush.json`, `crush.json`, `~/.config/crush/crush.json`
- Provider management (OpenAI, Anthropic, custom providers)
- LSP and MCP server configuration
- Mock providers for testing (use `config.UseMockProviders = true` in tests)

**TUI** (`internal/tui/`)
- Built with Bubble Tea v2 (`charm.land/bubbletea/v2`)
- Component-based architecture: chat, editor, header, messages
- Message list component handles real-time updates via pubsub
- Layout management through `layout.Sizeable`, `layout.Focusable` interfaces

**Tools** (`internal/agent/tools/`)
- File operations: view, edit, write, glob, grep
- Shell execution: bash commands with background support
- Web: fetch and search capabilities
- Job management: background job control (kill, output)
- MCP tool integration: stdio, http, and sse transports
- Tool descriptions in markdown files alongside implementations

**Database** (`internal/db/`)
- SQLite backend for messages, sessions, and file tracking
- Schema managed via `sqlc` (see `sqlc.yaml`)
- Migrations embedded in binary

**LSP Integration** (`internal/lsp/`)
- Language Server Protocol client implementation
- Per-language LSP configuration
- Provides code context to agents (definitions, references, diagnostics)

**VCS Integration** (`internal/vcs/`)
- Version control system detection and status reporting
- Supports Git and Jujutsu with pluggable architecture for future VCS systems
- Git status detection: conflicts, staged/uncommitted changes, untracked files, ahead/behind remote
- Jujutsu status detection: conflicts, uncommitted changes, current branch/change ID
- Status-aware icons displayed in sidebar (conflicts, dirty, staged, pushed, etc.)
- Displays current branch/change name in sidebar details pane

**Providers** (via `charm.land/fantasy` and `github.com/charmbracelet/catwalk`)
- Abstracts multiple LLM providers through Fantasy framework
- Catwalk maintains community provider database
- Auto-updates provider lists unless `disable_provider_auto_update` is set

### Key Design Patterns

**Context Propagation**
- Session ID, message ID, model capabilities passed via `context.Context`
- Use constants from `tools` package: `SessionIDContextKey`, `MessageIDContextKey`

**Concurrency**
- Thread-safe collections in `internal/csync/` (maps, slices, versioned maps)
- Message queuing per session to prevent race conditions
- Cancel contexts for interrupting long-running operations

**Event System**
- Pubsub broker (`internal/pubsub/`) for UI updates
- Event recording for metrics (`internal/event/`)
- LSP events separate from agent events

**Testing with Mocks**
```go
func TestWithMockProviders(t *testing.T) {
    originalUseMock := config.UseMockProviders
    config.UseMockProviders = true
    defer func() {
        config.UseMockProviders = originalUseMock
        config.ResetProviders()
    }()
    config.ResetProviders()

    // Test code using providers
}
```

## Code Style

- **Imports**: Use goimports formatting; group stdlib, external, internal packages
- **Formatting**: ALWAYS format with `gofumpt` (stricter than gofmt), fallback to `goimports` then `gofmt`
- **Error handling**: Return errors explicitly, wrap with `fmt.Errorf` and `%w`
- **Context**: Always pass `context.Context` as first parameter for operations
- **Interfaces**: Define in consuming packages, keep small and focused
- **Testing**: Use `testify/require`, `t.Parallel()` for parallel tests, `t.Setenv()` for env vars, `t.Tempdir()` for temp dirs (no cleanup needed)
- **JSON tags**: Use `snake_case` for JSON field names
- **File permissions**: Use octal notation (`0o755`, `0o644`)
- **Comments**:
  - Standalone comments: Start with capital, end with period, wrap at 78 columns
  - Inline comments: No trailing period
- **VCS Status Icons**: Follow oh-my-zsh conventions with priority-based display (conflicts > detached > staged > uncommitted > untracked > ahead/behind > clean)

## Commit Conventions

- ALWAYS use semantic commits: `fix:`, `feat:`, `chore:`, `refactor:`, `docs:`, `sec:`
- Keep commits to one line (excluding attribution) unless additional context is truly necessary
- Attribution is automatically added by Crush itself when committing through the agent

## Important Files

- `Taskfile.yaml`: Task runner configuration (uses https://taskfile.dev)
- `go.mod`: Dependencies (Go 1.25.5, uses greenteagc experiment)
- `.golangci.yml`: Linter configuration
- `schema.json`: Generated JSON schema for `crush.json` configuration
- `crush.json`: Example configuration at repository root
- `CRUSH.md`: Original development guide (this CLAUDE.md supersedes it for AI agents)
- `internal/tui/styles/icons.go`: Icon constants including VCS status icons

## VCS Status Icons

The sidebar displays VCS status with color-coded icons (priority order):

**Git/Jujutsu Status Icons:**
- `✖` (red) - Merge conflicts
- `⚠` (yellow) - Detached HEAD state (Git only)
- `●` (yellow) - Staged changes ready to commit (Git only)
- `✗` (yellow) - Uncommitted changes
- `?` (muted) - Untracked files (Git only)
- `↕` (yellow) - Diverged from remote (both ahead and behind) (Git only)
- `↑` (blue) - Unpushed commits / ahead of remote (Git only)
- `↓` (blue) - Behind remote (Git only)
- `✓` (green) - Clean working tree (Git only)
- `jj` (green) - Clean Jujutsu repository

**Display Format:**
- Git: Shows current branch name (e.g., `✗ main`, `✓ feature-branch`)
- Jujutsu: Shows current branch or change ID (e.g., `✗ my-branch`, `jj abc123`)
- Falls back to repository name if branch info unavailable

## Configuration and Initialization

- Projects can define context in files like `AGENTS.md` (customizable via `initialize_as` option)
- Crush analyzes codebase on init and creates project-specific context
- Logs stored in `./.crush/logs/crush.log` relative to project
- State stored in `~/.local/share/crush/crush.json` (Unix) or `%LOCALAPPDATA%\crush\crush.json` (Windows)

## Testing Philosophy

- Golden file tests: Use `-update` flag to regenerate when output intentionally changes
- Mock providers: Always use for tests involving LLM provider configurations
- Parallel tests: Use `t.Parallel()` when tests are independent
- Temp directories: Use `t.Tempdir()` for isolated test environments

## Environment Variables

Key environment variables:
- `CRUSH_PROFILE`: Enable pprof profiling on `localhost:6060`
- `CRUSH_GLOBAL_CONFIG`: Override user config location
- `CRUSH_GLOBAL_DATA`: Override data directory location
- `CRUSH_DISABLE_METRICS`: Disable telemetry collection
- `CRUSH_DISABLE_PROVIDER_AUTO_UPDATE`: Disable automatic provider updates
- Provider API keys: `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GROQ_API_KEY`, etc.

## Dependencies

- **TUI**: `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, `charm.land/bubbles/v2`
- **LLM Framework**: `charm.land/fantasy` (abstracts providers), `github.com/charmbracelet/catwalk` (provider database)
- **Database**: SQLite via `github.com/mattn/go-sqlite3`, code generation via `sqlc`
- **LSP**: Custom implementation in `internal/lsp/`
- **Testing**: `github.com/stretchr/testify`, `charm.land/x/vcr` (for recording HTTP interactions)
