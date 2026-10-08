package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	ownerAID = "00000000-0000-0000-0000-0000000000a1"
	ownerBID = "00000000-0000-0000-0000-0000000000b1"
)

type seedTask struct {
	id       string
	ownerID  string
	title    string
	status   string
	priority string
	labels   []string
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
	conn, err := pgx.Connect(connectCtx, databaseURL)
	connectCancel()
	if err != nil {
		log.Fatal("connect database failed")
	}

	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()

		if err := conn.Close(closeCtx); err != nil {
			log.Print("close database connection failed")
		}
	}()

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer seedCancel()

	if err := seed(seedCtx, conn); err != nil {
		log.Fatal("seed database failed")
	}

	log.Print("seed database completed")
}

func seed(ctx context.Context, conn *pgx.Conn) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	if _, err := tx.Exec(ctx, `
		INSERT INTO owners (id, oidc_issuer, oidc_subject)
		VALUES
			($1, 'local-seed', 'owner-a'),
			($2, 'local-seed', 'owner-b')
		ON CONFLICT (id) DO UPDATE
		SET
			oidc_issuer = EXCLUDED.oidc_issuer,
			oidc_subject = EXCLUDED.oidc_subject
	`, ownerAID, ownerBID); err != nil {
		return err
	}

	tasks := []seedTask{
		{
			id:       "00000000-0000-0000-0000-000000000101",
			ownerID:  ownerAID,
			title:    "Owner A backlog task",
			status:   "backlog",
			priority: "medium",
			labels:   []string{"seed", "owner-a"},
		},
		{
			id:       "00000000-0000-0000-0000-000000000102",
			ownerID:  ownerAID,
			title:    "Owner A done task",
			status:   "done",
			priority: "low",
			labels:   []string{"seed", "owner-a"},
		},
		{
			id:       "00000000-0000-0000-0000-000000000201",
			ownerID:  ownerBID,
			title:    "Owner B backlog task",
			status:   "backlog",
			priority: "high",
			labels:   []string{"seed", "owner-b"},
		},
	}

	for _, task := range tasks {
		if _, err := tx.Exec(ctx, `
			INSERT INTO tasks (id, owner_id, title, status, priority, labels)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE
			SET
				owner_id = EXCLUDED.owner_id,
				title = EXCLUDED.title,
				status = EXCLUDED.status,
				priority = EXCLUDED.priority,
				labels = EXCLUDED.labels,
				updated_at = now()
		`, task.id, task.ownerID, task.title, task.status, task.priority, task.labels); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
