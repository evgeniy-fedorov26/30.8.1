package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Хранилище данных
type Storage struct {
	db *pgxpool.Pool
}

// Конструктор, принимает строку подключения к БД
func New(constr string) (*Storage, error) {
	db, err := pgxpool.Connect(context.Background(), constr)
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

// Задача
type Task struct {
	ID         int
	Opened     time.Time
	Closed     *time.Time
	AuthorID   int
	AssignedID int
	Title      string
	Content    string
}

// Tasks возвращает список задач из БД
func (s *Storage) Tasks(taskID, authorID int) ([]Task, error) {
	rows, err := s.db.Query(context.Background(), `
		SELECT 
			id,
			opened,
			closed,
			author_id,
			assigned_id,
			title,
			content
		FROM tasks
		WHERE
			($1 = 0 OR id = $1) AND
			($2 = 0 OR author_id = $2)
		ORDER BY id;
	`, taskID, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err = rows.Scan(
			&t.ID,
			&t.Opened,
			&t.Closed,
			&t.AuthorID,
			&t.AssignedID,
			&t.Title,
			&t.Content,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// NewTask создаёт новую задачу и возвращает её ID
func (s *Storage) NewTask(t Task) (int, error) {
	var id int
	err := s.db.QueryRow(context.Background(), `
		INSERT INTO tasks (opened, author_id, assigned_id, title, content)
		VALUES ($1, $2, $3, $4, $5) RETURNING id;
	`, t.Opened, t.AuthorID, t.AssignedID, t.Title, t.Content).Scan(&id)
	return id, err
}

// CloseTask закрывает задачу, обновляя время завершения
func (s *Storage) CloseTask(taskID int) error {
	_, err := s.db.Exec(context.Background(), `
		UPDATE tasks
		SET closed = NOW()
		WHERE id = $1;
	`, taskID)
	return err
}

// UpdateTask обновляет существующую задачу
func (s *Storage) UpdateTask(t Task) error {
	_, err := s.db.Exec(context.Background(), `
		UPDATE tasks
		SET title = $1, content = $2, assigned_id = $3
		WHERE id = $4;
	`, t.Title, t.Content, t.AssignedID, t.ID)
	return err
}

// DeleteTask удаляет задачу из БД
func (s *Storage) DeleteTask(taskID int) error {
	_, err := s.db.Exec(context.Background(), `
		DELETE FROM tasks
		WHERE id = $1;
	`, taskID)
	return err
}
