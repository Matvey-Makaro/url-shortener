package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Matvey-Makaro/url-shortener/internal/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.NewStorage"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// TODO: Миграция БД

	queryStr := `CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);`
	stmt, err := db.Prepare(queryStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave, alias string) error {
	op := "storage.sqlite.SaveURL"
	queryStr := `INSERT INTO url (url, alias) VALUES (?, ?)`
	stmt, err := s.db.Prepare(queryStr)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(urlToSave, alias)
	if err != nil {

		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: %v", op, storage.ErrURLExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	op := "storage.sqlite.GetURL"
	queryStr := `SELECT url FROM url WHERE alias = ?`
	stmt, err := s.db.Prepare(queryStr)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var url string
	err = stmt.QueryRow(alias).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %v", op, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)

	}
	return url, nil
}

// TODO: Impl...
// func (s *Storage) DeleteURL(alias string) error {

// }
