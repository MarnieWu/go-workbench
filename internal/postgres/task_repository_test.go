package postgres

import (
	"testing"
)

func TestTaskRepositoryList(t *testing.T) {
	ctx, conn, schema := newTestDatabase(t)
	applyBusinessMigration(t, ctx, conn, schema)

	var ownerIdA, ownerIdB string
	createOwnerSql := `
		INSERT INTO owners (oidc_issuer, oidc_subject)
		VALUES ($1, $2)
		RETURNING id::text
	`
	oidcIssuer := "https://issuer.example.com"
	oidcSubject := "test-subject"
	err := conn.QueryRow(ctx, createOwnerSql, oidcIssuer, oidcSubject).Scan(&ownerIdA)
	if err != nil {
		t.Fatal(err)
	}
	err = conn.QueryRow(ctx, createOwnerSql, oidcIssuer, oidcSubject).Scan(&ownerIdB)
	if err != nil {
		t.Fatal(err)
	}

	createTaskSql := `
		INSERT INTO tasks (owner_id, title, status)
		VALUES ($1::uuid, $2, $3)
	`
	taskStatusBacklog := "backlog"
	taskStatusDone := "done"
	_, err = conn.Exec(ctx, createTaskSql, ownerIdA, "task a backlog", taskStatusBacklog)
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(ctx, createTaskSql, ownerIdA, "task a done", taskStatusDone)
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(ctx, createTaskSql, ownerIdB, "task b backlog", taskStatusBacklog)
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Exec(ctx, createTaskSql, ownerIdB, "task b done", taskStatusDone)
	if err != nil {
		t.Fatal(err)
	}

	repository := NewTaskRepository(conn)
}
