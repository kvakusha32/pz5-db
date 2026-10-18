package main

import (
	"context"
	"database/sql"
	"time"
)

type Task struct {
	ID        int
	Title     string
	Done      bool
	CreatedAt time.Time
}

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo { return &Repo{DB: db} }

func (r *Repo) CreateTask(ctx context.Context, title string) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx,
		`INSERT INTO tasks (title) VALUES ($1) RETURNING id`, title).Scan(&id)
	return id, err
}

func (r *Repo) ListTasks(ctx context.Context) ([]Task, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, title, done, created_at FROM tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *Repo) ListDone(ctx context.Context, done bool) ([]Task, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, title, done, created_at FROM tasks WHERE done = $1 ORDER BY id`, done)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *Repo) FindByID(ctx context.Context, id int) (*Task, error) {
	var task Task
	err := r.DB.QueryRowContext(ctx,
		`SELECT id, title, done, created_at FROM tasks WHERE id = $1`, id).
		Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *Repo) CreateMany(ctx context.Context, titles []string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, title := range titles {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO tasks (title) VALUES ($1)`, title); err != nil {
			return err
		}
	}
	return tx.Commit()
}
