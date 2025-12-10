package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

func MigrationUp(db *sql.DB) error {
	m, err := migrating(db)
	if err != nil {
		return fmt.Errorf("could not initialize migration: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}

func MigrationDown(db *sql.DB) error {
	m, err := migrating(db)
	if err != nil {
		return fmt.Errorf("could not initialize migration: %w", err)
	}
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

func migrating(db *sql.DB) (*migrate.Migrate, error) {

	config := &postgres.Config{}
	driver, err := postgres.WithInstance(db, config)
	if err != nil {
		return nil, fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/database/migrations",
		"postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("could not create migrate instance: %w", err)
	}

	return m, nil
}
