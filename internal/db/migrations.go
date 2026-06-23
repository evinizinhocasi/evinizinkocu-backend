package db

import (
	"context"
	_ "embed"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/000001_init.up.sql
var initSQL string

func RunMigrations(db *Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Running database migrations...")

	// Clean up comments and execute statements
	// Note: Simple split by semicolon is fine for our schema since we don't have triggers/stored routines using semicolons inside statements.
	queries := strings.Split(initSQL, ";")
	for _, query := range queries {
		q := strings.TrimSpace(query)
		if q == "" {
			continue
		}

		_, err := db.Pool.Exec(ctx, q)
		if err != nil {
			log.Printf("Failed executing migration statement: %s\nError: %v\n", q, err)
			return err
		}
	}

	log.Println("Database migrations applied successfully")

	// Seed default catalog if empty
	if err := SeedDefaultCatalog(ctx, db.Pool); err != nil {
		log.Printf("Warning: Failed to seed default catalog: %v\n", err)
	}

	return nil
}

// SeedDefaultCatalog is intentionally a no-op.
// Catalog (exam types, subjects, topics) is now managed dynamically
// through the admin panel. Use the Katalog Yönetimi tab in the
// superadmin dashboard to add/edit/delete exam types, subjects, and topics.
func SeedDefaultCatalog(ctx context.Context, pool *pgxpool.Pool) error {
	log.Println("Catalog seeding skipped: catalog is managed via admin panel.")
	return nil
}
