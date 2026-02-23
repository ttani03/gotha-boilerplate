package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ttani03/gotha-boilerplate/internal/db/generated"
	"github.com/ttani03/gotha-boilerplate/internal/middleware"
	"github.com/ttani03/gotha-boilerplate/web/templates/components"
	"github.com/ttani03/gotha-boilerplate/web/templates/pages"
)

// TodoHandler handles todo-related requests.
type TodoHandler struct {
	db      *pgxpool.Pool
	queries *generated.Queries
}

// NewTodoHandler creates a new TodoHandler.
func NewTodoHandler(db *pgxpool.Pool, queries *generated.Queries) *TodoHandler {
	return &TodoHandler{
		db:      db,
		queries: queries,
	}
}

// ListPage renders the todo list page.
func (h *TodoHandler) ListPage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	uid, err := parseUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	todos, err := h.queries.ListTodosByUserID(r.Context(), uid)
	if err != nil {
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}

	render(w, r, pages.TodoList(todos))
}

// Create handles creating a new todo.
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	uid, err := parseUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	todo, err := h.queries.CreateTodo(r.Context(), generated.CreateTodoParams{
		UserID: uid,
		Title:  title,
	})
	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	// Return only the new todo item for HTMX to append
	render(w, r, components.TodoItem(todo))
}

// Update handles updating a todo's title.
func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	uid, err := parseUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	todoID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	todo, err := h.queries.UpdateTodoTitle(r.Context(), generated.UpdateTodoTitleParams{
		Title:  title,
		ID:     todoID,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		return
	}

	render(w, r, components.TodoItem(todo))
}

// Toggle handles toggling a todo's completed status.
func (h *TodoHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	uid, err := parseUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	todoID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	todo, err := h.queries.ToggleTodoCompleted(r.Context(), generated.ToggleTodoCompletedParams{
		ID:     todoID,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Failed to toggle todo", http.StatusInternalServerError)
		return
	}

	render(w, r, components.TodoItem(todo))
}

// Delete handles deleting a todo.
func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	uid, err := parseUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	todoID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	if err := h.queries.DeleteTodo(r.Context(), generated.DeleteTodoParams{
		ID:     todoID,
		UserID: uid,
	}); err != nil {
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	// Return empty response for HTMX to remove the element
	w.WriteHeader(http.StatusOK)
}
