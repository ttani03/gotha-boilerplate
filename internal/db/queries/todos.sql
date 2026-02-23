-- name: CreateTodo :one
INSERT INTO todos (user_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: ListTodosByUserID :many
SELECT * FROM todos
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetTodoByID :one
SELECT * FROM todos
WHERE id = $1 AND user_id = $2;

-- name: UpdateTodoTitle :one
UPDATE todos
SET title = $1, updated_at = NOW()
WHERE id = $2 AND user_id = $3
RETURNING *;

-- name: ToggleTodoCompleted :one
UPDATE todos
SET completed = NOT completed, updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteTodo :exec
DELETE FROM todos
WHERE id = $1 AND user_id = $2;
