package migration

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database"
    "github.com/golang-migrate/migrate/v4/database/mysql"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    "github.com/golang-migrate/migrate/v4/database/sqlite"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"
    _ "modernc.org/sqlite"
)

func Migrate(dsn string, dbType string, isMainDB bool) error {
    // Get the current working directory
    cwd, err := os.Getwd()
    if err != nil {
        log.Printf("Failed to get current working directory: %v", err)
        return err
    }

    // Move up one directory level if we're in the cmd folder
    if filepath.Base(cwd) == "cmd" {
        cwd = filepath.Dir(cwd)
    }

    // Define the migration folder path based on whether it's the main DB or tenant DB
    var migrationFolder string
    if isMainDB {
        migrationFolder = filepath.Join(cwd, "migrations", "main")
    } else {
        migrationFolder = filepath.Join(cwd, "migrations", "tenant")
    }
    migrationFolder = filepath.ToSlash(migrationFolder) // Convert to URL-friendly format
    fmt.Printf("Applying %s database migrations from folder: %s\n", 
               map[bool]string{true: "main", false: "tenant"}[isMainDB], 
               migrationFolder)

    // Check if the migration folder exists
    if _, err := os.Stat(migrationFolder); os.IsNotExist(err) {
        log.Printf("Migration folder does not exist: %v", migrationFolder)
        return err
    }

    // Open the database
    db, err := sql.Open(dbType, dsn)
    if err != nil {
        log.Printf("Failed to open database: %v", err)
        return err
    }
    defer db.Close()

    // Create a database-specific driver
    var driver database.Driver
    switch dbType {
    case "postgres":
        driver, err = postgres.WithInstance(db, &postgres.Config{})
    case "mysql":
        driver, err = mysql.WithInstance(db, &mysql.Config{})
    case "sqlite":
        driver, err = sqlite.WithInstance(db, &sqlite.Config{})
    default:
        return fmt.Errorf("unsupported database type: %s", dbType)
    }
    if err != nil {
        log.Printf("Failed to create database driver: %v", err)
        return err
    }

    // Initialize the migrate instance
    m, err := migrate.NewWithDatabaseInstance(
        "file://"+migrationFolder,
        dbType, driver)
    if err != nil {
        log.Printf("Failed to initialize migrate instance: %v", err)
        return err
    }
    defer m.Close()

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

// Currently supported databases:
// - PostgreSQL ("postgres")
// - MySQL ("mysql")
// - SQLite ("sqlite")
