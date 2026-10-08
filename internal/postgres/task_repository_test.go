package postgres

import (
	"fmt"
	"go-workbench/internal/task"
	"testing"
)

func TestTaskRepositoryList(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)

	var ownerAID, ownerBID, ownerCID string
	createOwnerSql := `
		INSERT INTO owners (oidc_issuer, oidc_subject)
		VALUES ($1, $2)
		RETURNING id::text
	`
	oidcIssuer := "https://issuer.example.com"
	oidcSubject := "test-subject"
	err := conn.QueryRow(ctx, createOwnerSql, fmt.Sprintf("%s-a", oidcIssuer), fmt.Sprintf("%s-a", oidcSubject)).Scan(&ownerAID)
	if err != nil {
		t.Fatalf("createOwnerSql error = %v", err)
	}
	err = conn.QueryRow(ctx, createOwnerSql, fmt.Sprintf("%s-b", oidcIssuer), fmt.Sprintf("%s-b", oidcSubject)).Scan(&ownerBID)
	if err != nil {
		t.Fatalf("createOwnerSql error = %v", err)
	}
	err = conn.QueryRow(ctx, createOwnerSql, fmt.Sprintf("%s-c", oidcIssuer), fmt.Sprintf("%s-c", oidcSubject)).Scan(&ownerCID)
	if err != nil {
		t.Fatalf("createOwnerSql error = %v", err)
	}

	createTaskSql := `
		INSERT INTO tasks (owner_id, title, status)
		VALUES ($1::uuid, $2, $3)
	`
	taskATitle := "task a"
	taskBTitle := "task b"
	taskStatusBacklog := task.StatusBacklog
	taskStatusDone := task.StatusDone
	var ownerBTaskID string
	_, err = conn.Exec(ctx, createTaskSql, ownerAID, taskATitle, taskStatusBacklog)
	if err != nil {
		t.Fatalf("createTaskSql error = %v", err)
	}
	_, err = conn.Exec(ctx, createTaskSql, ownerAID, taskATitle, taskStatusDone)
	if err != nil {
		t.Fatalf("createTaskSql error = %v", err)
	}
	err = conn.QueryRow(ctx, `
		INSERT INTO tasks (owner_id, title, status)
		VALUES ($1::uuid, $2, $3)
		RETURNING id::text
	`, ownerBID, taskBTitle, taskStatusBacklog).Scan(&ownerBTaskID)
	if err != nil {
		t.Fatalf("createTaskSql error = %v", err)
	}

	repository := NewTaskRepository(conn)
	tasks, err := repository.List(ctx, ownerAID, task.ListTasksFilter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}

	for _, got := range tasks {
		if got.OwnerID != ownerAID {
			t.Fatalf("task.OwnerID = %q, want %q", got.OwnerID, ownerAID)
		}
		if got.ID == ownerBTaskID {
			t.Fatalf("returned Owner B task: %#v", got)
		}
	}

	tasks, err = repository.List(ctx, ownerCID, task.ListTasksFilter{})
	if err != nil {
		t.Fatalf("List(empty owner) error = %v", err)
	}

	if tasks == nil {
		t.Fatal("tasks = nil, want empty slice")
	}
	if len(tasks) != 0 {
		t.Fatalf("len(tasks) = %d, want 0", len(tasks))
	}
}
