package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	directory := os.Getenv("MIGRATIONS_DIR")
	if directory == "" {
		directory = "/app/migrations"
	}
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set migration dialect: %v", err)
	}
	if err := goose.UpContext(ctx, db, directory); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	log.Println("database migrations applied")
}
