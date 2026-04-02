package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/LiviTT/HMCTS/internal/model"
)

type DB struct {
	conn *sql.DB
}

func New(connStr string) (*DB, error) {
	conn, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
		DO $$ BEGIN
			CREATE TYPE task_status AS ENUM ('todo', 'in_progress', 'complete');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$;

		CREATE TABLE IF NOT EXISTS tasks (
			id          UUID PRIMARY KEY,
			title       TEXT NOT NULL,
			description TEXT,
			status      TASK_STATUS NOT NULL,
			due_date    TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT current_timestamp,
			updated_at  TIMESTAMP WITH TIME ZONE
		)
	`)
	return err
}

func (db *DB) CreateTask(input model.CreateTaskRequest) (*model.Task, error) {
	t := &model.Task{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		DueDate:     input.DueDate,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	_, err := db.conn.Exec(`
		INSERT INTO tasks (id, title, description, status, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		t.ID, t.Title, t.Description, string(t.Status),
		t.DueDate, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting task: %w", err)
	}

	return t, nil
}

func (db *DB) GetTaskByID(id string) (*model.Task, error) {
	row := db.conn.QueryRow(`
		SELECT id, title, description, status, due_date, created_at, updated_at
		FROM tasks WHERE id = $1`, id)

	return scanTask(row)
}

func (db *DB) GetAllTasks() ([]*model.Task, error) {
	rows, err := db.conn.Query(`
		SELECT id, title, description, status, due_date, created_at, updated_at
		FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

func (db *DB) UpdateTaskStatus(id string, status model.TaskStatus) (*model.Task, error) {
	now := time.Now().UTC()
	result, err := db.conn.Exec(`
		UPDATE tasks SET status = $1, updated_at = $2 WHERE id = $3`,
		string(status), now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("updating task status: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	return db.GetTaskByID(id)
}

func (db *DB) DeleteTask(id string) error {
	result, err := db.conn.Exec(`DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("task not found: %s", id)
	}
	return nil
}

func scanTask(row *sql.Row) (*model.Task, error) {
	var t model.Task
	var description sql.NullString
	var status string

	err := row.Scan(
		&t.ID, &t.Title, &description, &status,
		&t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scanning task: %w", err)
	}

	if description.Valid {
		t.Description = &description.String
	}
	t.Status = model.TaskStatus(status)
	return &t, nil
}

func scanTaskRow(rows *sql.Rows) (*model.Task, error) {
	var t model.Task
	var description sql.NullString
	var status string

	err := rows.Scan(
		&t.ID, &t.Title, &description, &status,
		&t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning task row: %w", err)
	}

	if description.Valid {
		t.Description = &description.String
	}
	t.Status = model.TaskStatus(status)
	return &t, nil
}
