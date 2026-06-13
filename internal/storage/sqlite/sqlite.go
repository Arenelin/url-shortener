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

func (s *Storage) CreateTask(taskName string) (storage.TaskStructure, error) {
	const op = "storage.sqlite.CreateTask"

	stmt, err := s.db.Prepare("INSERT INTO task(taskName) VALUES (?)")
	if err != nil {
		return storage.TaskStructure{}, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(taskName)
	if err != nil {
		if sqlErr, ok := err.(*sqlite.Error); ok {
			if strings.Contains(sqlErr.Error(), "UNIQUE constraint failed") {
				return storage.TaskStructure{}, fmt.Errorf("%s: %w", op, storage.ErrTaskExists)
			}
		}
		return storage.TaskStructure{}, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()

	var createdTask storage.TaskStructure
	var name string
	var done bool

	selectQuery := "SELECT taskName, isDone FROM task WHERE id = ?"
	err = s.db.QueryRow(selectQuery, id).Scan(&name, &done)
	if err != nil {
		return storage.TaskStructure{}, fmt.Errorf("%s: select updated task failed: %w", op, err)
	}

	createdTask.TaskName = &name
	createdTask.IsDone = &done
	createdTask.Id = &id

	return createdTask, nil
}

func (s *Storage) UpdateTask(id int64, task storage.Task) (storage.TaskStructure, error) {
	const op = "storage.sqlite.UpdateTask"

	var columns []string
	var args []interface{}

	if task.TaskName != nil {
		columns = append(columns, "taskName = ?")
		args = append(args, *task.TaskName)
	}

	if task.IsDone != nil {
		columns = append(columns, "isDone = ?")
		args = append(args, *task.IsDone)
	}

	if len(columns) > 0 {
		args = append(args, id)

		//language=text
		query := "UPDATE task SET " + strings.Join(columns, ", ") + " WHERE id = ?"

		stmt, err := s.db.Prepare(query)
		if err != nil {
			return storage.TaskStructure{}, fmt.Errorf("%s: %w", op, err)
		}
		defer stmt.Close()

		res, err := stmt.Exec(args...)
		if err != nil {
			if sqlErr, ok := err.(*sqlite.Error); ok {
				if strings.Contains(sqlErr.Error(), "UNIQUE constraint failed") {
					return storage.TaskStructure{}, fmt.Errorf("%s: %w", op, storage.ErrTaskExists)
				}
			}
			return storage.TaskStructure{}, fmt.Errorf("%s: %w", op, err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return storage.TaskStructure{}, fmt.Errorf("%s: %w", op, err)
		}
		if rowsAffected == 0 {
			return storage.TaskStructure{}, fmt.Errorf("%s: task not found", op)
		}
	}

	// 2. Читаем обновленный объект из базы данных
	var updatedTask storage.TaskStructure
	var name string
	var done bool

	selectQuery := "SELECT taskName, isDone FROM task WHERE id = ?"
	err := s.db.QueryRow(selectQuery, id).Scan(&name, &done)
	if err != nil {
		return storage.TaskStructure{}, fmt.Errorf("%s: select updated task failed: %w", op, err)
	}

	updatedTask.TaskName = &name
	updatedTask.IsDone = &done
	updatedTask.Id = &id

	return updatedTask, nil
}

func (s *Storage) GetAllTasks() ([]storage.TaskStructure, error) {
	const op = "storage.sqlite.GetAllTasks"

	stmt, err := s.db.Prepare("SELECT taskName, isDone, id FROM task")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var resTasks []storage.TaskStructure
	for rows.Next() {
		var taskName string
		var isDone bool
		var id int64
		if err := rows.Scan(&taskName, &isDone, &id); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		resTasks = append(resTasks, storage.TaskStructure{Id: &id, IsDone: &isDone, TaskName: &taskName})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resTasks, nil
}

func (s *Storage) DeleteTask(id int64) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM task WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrTaskNotFound)
	}

	return nil
}
