package tools

import (
	"context"
	_ "embed"
	"fmt"

	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/session"
)

//go:embed todos.md
var todosDescription []byte

const TodosToolName = "todos"

type TodosParams struct {
	Todos []TodoItem `json:"todos" description:"The updated todo list"`
}

type TodoItem struct {
	Content    string `json:"content" description:"What needs to be done (imperative form)"`
	Status     string `json:"status" description:"Task status: pending, in_progress, or completed"`
	ActiveForm string `json:"active_form" description:"Present continuous form (e.g., 'Running tests')"`
}

func NewTodosTool(sessions session.Service) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		TodosToolName,
		string(todosDescription),
		func(ctx context.Context, params TodosParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for managing todos")
			}

			// Get current session
			currentSession, err := sessions.Get(ctx, sessionID)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to get session: %w", err)
			}

			// Convert TodoItems to session.Todo
			todos := make([]session.Todo, len(params.Todos))
			for i, item := range params.Todos {
				todos[i] = session.Todo{
					Content:    item.Content,
					Status:     session.TodoStatus(item.Status),
					ActiveForm: item.ActiveForm,
				}
			}

			// Update session with new todos
			currentSession.Todos = todos
			_, err = sessions.Save(ctx, currentSession)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to save todos: %w", err)
			}

			// Build response showing current state
			response := "Todo list updated successfully.\n\n"

			pendingCount := 0
			inProgressCount := 0
			completedCount := 0

			for _, todo := range todos {
				switch todo.Status {
				case session.TodoStatusPending:
					pendingCount++
				case session.TodoStatusInProgress:
					inProgressCount++
				case session.TodoStatusCompleted:
					completedCount++
				}
			}

			response += fmt.Sprintf("Status: %d pending, %d in progress, %d completed\n",
				pendingCount, inProgressCount, completedCount)

			return fantasy.NewTextResponse(response), nil
		})
}
