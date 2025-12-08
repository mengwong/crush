package chat

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/crush/internal/tui/styles"
)

const (
	pillHeight = 3 // Height of pills including border
)

func queuePill(queue int, t *styles.Theme) string {
	if queue <= 0 {
		return ""
	}
	triangles := styles.ForegroundGrad("▶▶▶▶▶▶▶▶▶", false, t.RedDark, t.Accent)
	if queue < 10 {
		triangles = triangles[:queue]
	}

	allTriangles := strings.Join(triangles, "")

	return t.S().Base.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.BgOverlay).
		PaddingLeft(1).
		PaddingRight(1).
		Render(fmt.Sprintf("%s %d Queued", allTriangles, queue))
}

// todoPill renders a pill showing todo progress and current task.
// spinnerView is the current spinner frame (only shown if there's an in-progress todo).
func todoPill(todos []session.Todo, spinnerView string, t *styles.Theme) string {
	if len(todos) == 0 {
		return ""
	}

	// Count completed and find current in-progress todo
	completed := 0
	var currentTodo *session.Todo
	for i := range todos {
		switch todos[i].Status {
		case session.TodoStatusCompleted:
			completed++
		case session.TodoStatusInProgress:
			if currentTodo == nil {
				currentTodo = &todos[i]
			}
		}
	}

	total := len(todos)

	// Build the content
	label := t.S().Base.Foreground(t.FgMuted).Render("To-Do")
	progress := t.S().Base.Foreground(t.FgMuted).Render(fmt.Sprintf("%d/%d", completed, total))

	var content string
	if currentTodo != nil {
		// Show spinner and task text when in progress
		taskText := currentTodo.Content
		if currentTodo.ActiveForm != "" {
			taskText = currentTodo.ActiveForm
		}
		// Truncate if too long
		maxTaskLen := 40
		if len(taskText) > maxTaskLen {
			taskText = taskText[:maxTaskLen-1] + "…"
		}
		task := t.S().Base.Foreground(t.FgSubtle).Render(taskText)
		content = fmt.Sprintf("%s %s %s  %s", spinnerView, label, progress, task)
	} else {
		// No spinner or task text when nothing in progress
		icon := t.S().Base.Foreground(t.FgMuted).Render("∴")
		content = fmt.Sprintf("%s %s %s", icon, label, progress)
	}

	return t.S().Base.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.BgOverlay).
		PaddingLeft(1).
		PaddingRight(1).
		Render(content)
}

// addPillSpacing adds spacing between pills for display.
func addPillSpacing(pills []string) []string {
	if len(pills) <= 1 {
		return pills
	}
	result := make([]string, 0, len(pills)*2-1)
	for i, pill := range pills {
		result = append(result, pill)
		if i < len(pills)-1 {
			result = append(result, "  ")
		}
	}
	return result
}
