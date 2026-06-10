package sqlite

import (
	"database/sql"
	"fmt"
	"modernc.org/sqlite"
	"strings"
	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
CREATE TABLE IF NOT EXISTS task(
    id INTEGER PRIMARY KEY,
    taskName TEXT NOT NULL UNIQUE,
    isDone BOOLEAN NOT NULL DEFAULT false);
`)
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateTask(taskName string) (int64, error) {
	const op = "storage.sqlite.CreateTask"

	stmt, err := s.db.Prepare("INSERT INTO task(taskName) VALUES (?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(taskName)
	if err != nil {
		if sqlErr, ok := err.(*sqlite.Error); ok {
			if strings.Contains(sqlErr.Error(), "UNIQUE constraint failed") {
				return 0, fmt.Errorf("%s: %w", op, storage.ErrTaskExists)
			}
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetAllTasks() ([]string, error) {
	const op = "storage.sqlite.GetAllTasks"

	stmt, err := s.db.Prepare("SELECT taskName FROM task")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var resTasks []string
	for rows.Next() {
		var taskName string
		if err := rows.Scan(&taskName); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		resTasks = append(resTasks, taskName)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resTasks, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
	}

	return nil
}
