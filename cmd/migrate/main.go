package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-workbench/internal/postgres"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	dir := flag.String("dir", "migrations", "directory containing versioned .up.sql files")
	flag.Parse()
	if flag.NArg() != 1 || flag.Arg(0) != "up" {
		return fmt.Errorf("usage: go run ./cmd/migrate -dir migrations up")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	conn, err := postgres.ConnectTestDatabase(ctx, os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	applied, err := postgres.Migrate(ctx, conn, *dir, "public")
	if err != nil {
		return err
	}
	fmt.Printf("migrations applied: %d\n", applied)
	return nil
}
