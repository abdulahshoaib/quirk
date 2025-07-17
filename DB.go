package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func InitSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS user_tokens (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) NOT NULL,
		token TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_user_tokens_token ON user_tokens(token);
	CREATE INDEX IF NOT EXISTS idx_user_tokens_email ON user_tokens(email);
	`
	res, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("Error initializing schema: %v", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("DB schema initialized, but couldn't get rows affected: %v", err)
	} else {
		log.Printf("DB schema initialized, rows affected: %d", rowsAffected)
	}
	return nil
}

func connect() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection with database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed ping database: %w", err)
	}

	return db, nil
}
