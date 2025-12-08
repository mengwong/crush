package chat

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/crush/internal/tui/styles"
)

const pillHeightWithBorder = 3

func queuePill(queue int, focused, pillsPanelFocused bool, t *styles.Theme) string {
	if queue <= 0 {
		return ""
	}
	triangles := styles.ForegroundGrad("▶▶▶▶▶▶▶▶▶", false, t.RedDark, t.Accent)
	if queue < 10 {
		triangles = triangles[:queue]
	}

	content := fmt.Sprintf("%s %d Queued", strings.Join(triangles, ""), queue)

	style := t.S().Base.PaddingLeft(1).PaddingRight(1)
	if !pillsPanelFocused || focused {
		style = style.BorderStyle(lipgloss.RoundedBorder()).BorderForeground(t.BgOverlay)
	} else {
		style = style.BorderStyle(lipgloss.HiddenBorder())
	}
	return style.Render(content)
}

func todoPill(todos []session.Todo, spinnerView string, focused, pillsPanelFocused bool, t *styles.Theme) string {
	if len(todos) == 0 {
		return ""
	}

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
	allDone := completed == total

	var prefix string
	if allDone {
		prefix = t.S().Base.Foreground(t.Green).Render("✓") + " "
	}

	label := "To-Do"
	progress := fmt.Sprintf("%d/%d", completed, total)

	var content string
	if pillsPanelFocused {
		content = fmt.Sprintf("%s%s %s", prefix, label, progress)
	} else if currentTodo != nil {
		taskText := currentTodo.Content
		if currentTodo.ActiveForm != "" {
			taskText = currentTodo.ActiveForm
		}
		if len(taskText) > 40 {
			taskText = taskText[:39] + "…"
		}
		task := t.S().Base.Foreground(t.FgSubtle).Render(taskText)
		content = fmt.Sprintf("%s %s %s  %s", spinnerView, label, progress, task)
	} else {
		content = fmt.Sprintf("%s%s %s", prefix, label, progress)
	}

	style := t.S().Base.PaddingLeft(1).PaddingRight(1)
	if !pillsPanelFocused || focused {
		style = style.BorderStyle(lipgloss.RoundedBorder()).BorderForeground(t.BgOverlay)
	} else {
		style = style.BorderStyle(lipgloss.HiddenBorder())
	}
	return style.Render(content)
}

// todoList renders the expanded todo list. Sorted: completed, in-progress, pending.
func todoList(todos []session.Todo, spinnerView string, t *styles.Theme) string {
	if len(todos) == 0 {
		return ""
	}

	sorted := make([]session.Todo, len(todos))
	copy(sorted, todos)
	sortTodos(sorted)

	var lines []string
	for _, todo := range sorted {
		var prefix string
		var textStyle lipgloss.Style

		switch todo.Status {
		case session.TodoStatusCompleted:
			prefix = t.S().Base.Foreground(t.FgMuted).Render("  ✓") + " "
			textStyle = t.S().Base.Foreground(t.FgMuted)
		case session.TodoStatusInProgress:
			prefix = "  " + spinnerView + " "
			textStyle = t.S().Base.Foreground(t.GreenDark)
		default:
			prefix = t.S().Base.Foreground(t.FgMuted).Render("  •") + " "
			textStyle = t.S().Base.Foreground(t.FgBase)
		}

		text := todo.Content
		if todo.Status == session.TodoStatusInProgress && todo.ActiveForm != "" {
			text = todo.ActiveForm
		}

		lines = append(lines, prefix+textStyle.Render(text))
	}

	return strings.Join(lines, "\n")
}

func sortTodos(todos []session.Todo) {
	statusOrder := func(s session.TodoStatus) int {
		switch s {
		case session.TodoStatusCompleted:
			return 0
		case session.TodoStatusInProgress:
			return 1
		default:
			return 2
		}
	}
	for i := 0; i < len(todos)-1; i++ {
		for j := i + 1; j < len(todos); j++ {
			if statusOrder(todos[i].Status) > statusOrder(todos[j].Status) {
				todos[i], todos[j] = todos[j], todos[i]
			}
		}
	}
}

func queueList(queueItems []string, t *styles.Theme) string {
	if len(queueItems) == 0 {
		return ""
	}

	var lines []string
	for _, item := range queueItems {
		text := item
		if len(text) > 60 {
			text = text[:59] + "…"
		}
		prefix := t.S().Base.Foreground(t.FgMuted).Render("  •") + " "
		lines = append(lines, prefix+t.S().Base.Foreground(t.FgMuted).Render(text))
	}

	return strings.Join(lines, "\n")
}
