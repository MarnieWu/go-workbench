package postgres

import (
	"context"
	"go-workbench/internal/task"

	"github.com/jackc/pgx/v5"
)

type TaskRepository struct {
	conn *pgx.Conn
}

func NewTaskRepository(conn *pgx.Conn) *TaskRepository {
	return &TaskRepository{conn: conn}
}

func (r *TaskRepository) List(
	ctx context.Context,
	ownerID string,
	filter task.ListTasksFilter,
) ([]task.Task, error) {
	query := `
		SELECT id, owner_id, project_id::text, title, COALESCE(description, ''), status, priority, labels,
			due_at, archived_at, version, created_at, updated_at
		FROM tasks
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`
	args := []any{ownerID}

	if filter.Status != nil {
		query = `
			SELECT id, owner_id, project_id::text, title, COALESCE(description, ''), status, priority, labels,
				due_at, archived_at, version, created_at, updated_at
			FROM tasks
			WHERE owner_id = $1 AND status = $2
			ORDER BY created_at DESC
		`
		args = append(args, string(*filter.Status))
	}

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]task.Task, 0)

	for rows.Next() {
		var t task.Task
		if err := rows.Scan(
			&t.ID,
			&t.OwnerID,
			&t.ProjectID,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.Priority,
			&t.Labels,
			&t.DueAt,
			&t.ArchivedAt,
			&t.Version,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
