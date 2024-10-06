package migrations

import (
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
	"log"
)

//go:embed *.sql
var embeddedMigrations embed.FS

func ApplyMigrations(db *sql.DB) error {
	goose.SetBaseFS(embeddedMigrations)

	if err := goose.Up(db, "."); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
		return err
	}
	log.Println("Migrations applied successfully!")
	return nil
}
