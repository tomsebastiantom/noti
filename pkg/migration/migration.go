package migrate

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // or the database you're using
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(dsn string) error {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current working directory: %v", err)
		return err
	}

	// Define the migration folder path relative to the current working directory
	migrationFolder := filepath.Join(cwd, "migrations")
	migrationFolder = filepath.ToSlash(migrationFolder) // Convert to URL-friendly format
	fmt.Println("Applying migrations from folder:", migrationFolder)

	// Check if the migration folder exists
	if _, err := os.Stat(migrationFolder); os.IsNotExist(err) {
		log.Printf("Migration folder does not exist: %v", migrationFolder)
		return err
	}

	// List all SQL files in the migration folder
	err = filepath.Walk(migrationFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			fmt.Println("Found migration file:", path)
		}
		return nil
	})
	if err != nil {
		log.Printf("Failed to list SQL files: %v", err)
		return err
	}

	// Update the DSN to disable SSL
	dsn += "?sslmode=disable"

	// Initialize the migrate instance
	m, err := migrate.New(
		"file://"+migrationFolder,
		dsn,
	)
	if err != nil {
		log.Printf("Failed to initialize migrate instance: %v", err)
		return err
	}

	// Get the current version and dirty state
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Printf("Failed to get current migration version: %v", err)
		return err
	}

	if dirty {
		log.Printf("The migration is in a dirty state at version %d", version)
		return fmt.Errorf("migration is in a dirty state at version %d", version)
	}

	// Apply the migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Printf("Failed to apply migrations: %v", err)
		return err
	}

	if err == migrate.ErrNoChange {
		fmt.Println("No new migrations to apply.")
	} else {
		fmt.Println("Migrations applied successfully.")
	}

	return nil
}
