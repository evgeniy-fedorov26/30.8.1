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
	Tags       []string
}

// Tasks возвращает список задач из БД
func (s *Storage) Tasks(taskID, authorID int) ([]Task, error) {
	rows, err := s.db.Query(context.Background(), `
		SELECT 
			t.id,
			t.opened,
			t.closed,
			t.author_id,
			t.assigned_id,
			t.title,
			t.content,
			COALESCE(array_agg(tag.name), '{}') AS tags
		FROM tasks t
		LEFT JOIN task_tags tt ON t.id = tt.task_id
		LEFT JOIN tags tag ON tt.tag_id = tag.id
		WHERE
			($1 = 0 OR t.id = $1) AND
			($2 = 0 OR t.author_id = $2)
		GROUP BY t.id
		ORDER BY t.id;
	`, taskID, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var tags []string
		err = rows.Scan(
			&t.ID,
			&t.Opened,
			&t.Closed,
			&t.AuthorID,
			&t.AssignedID,
			&t.Title,
			&t.Content,
			&tags,
		)
		if err != nil {
			return nil, err
		}
		t.Tags = tags
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
	if err != nil {
		return 0, err
	}

	// Добавление меток, если они указаны
	if len(t.Tags) > 0 {
		err = s.AddTagsToTask(id, t.Tags)
	}
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
	if err != nil {
		return err
	}

	// Обновление меток
	if len(t.Tags) > 0 {
		err = s.UpdateTagsForTask(t.ID, t.Tags)
	}
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

// AddTagsToTask добавляет метки к задаче
func (s *Storage) AddTagsToTask(taskID int, tags []string) error {
	for _, tag := range tags {
		var tagID int
		err := s.db.QueryRow(context.Background(), `
			INSERT INTO tags (name)
			VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id;
		`, tag).Scan(&tagID)
		if err != nil {
			return err
		}

		_, err = s.db.Exec(context.Background(), `
			INSERT INTO task_tags (task_id, tag_id)
			VALUES ($1, $2) ON CONFLICT DO NOTHING;
		`, taskID, tagID)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateTagsForTask обновляет метки для задачи
func (s *Storage) UpdateTagsForTask(taskID int, tags []string) error {
	_, err := s.db.Exec(context.Background(), `
		DELETE FROM task_tags WHERE task_id = $1;
	`, taskID)
	if err != nil {
		return err
	}

	return s.AddTagsToTask(taskID, tags)
}

// TasksByTag возвращает задачи по указанной метке
func (s *Storage) TasksByTag(tag string) ([]Task, error) {
	rows, err := s.db.Query(context.Background(), `
		SELECT 
			t.id,
			t.opened,
			t.closed,
			t.author_id,
			t.assigned_id,
			t.title,
			t.content,
			COALESCE(array_agg(tag.name), '{}') AS tags
		FROM tasks t
		JOIN task_tags tt ON t.id = tt.task_id
		JOIN tags tag ON tt.tag_id = tag.id
		WHERE tag.name = $1
		GROUP BY t.id
		ORDER BY t.id;
	`, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var tags []string
		err = rows.Scan(
			&t.ID,
			&t.Opened,
			&t.Closed,
			&t.AuthorID,
			&t.AssignedID,
			&t.Title,
			&t.Content,
			&tags,
		)
		if err != nil {
			return nil, err
		}
		t.Tags = tags
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
