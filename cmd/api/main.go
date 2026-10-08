package main

import (
	"context"
	"go-workbench/internal/httpapi"
	"go-workbench/internal/postgres"
	"go-workbench/internal/task"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
	conn, err := pgx.Connect(connectCtx, databaseURL)
	connectCancel()
	if err != nil {
		log.Fatal("Failed to connect to database")
	}

	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()

		if err := conn.Close(closeCtx); err != nil {
			log.Print("close database connection failed")
		}
	}()

	repository := postgres.NewTaskRepository(conn)
	service := task.NewService(repository)
	router := httpapi.NewRouter(service, httpapi.LocalOwnerMiddleware(localOwnerId))

	if err := router.Run(":8080"); err != nil {
		log.Fatal("start server failed")
	}
}
