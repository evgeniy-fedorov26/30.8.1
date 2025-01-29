/*
    Схема БД для информационной системы
    отслеживания выполнения задач.
*/

DROP TABLE IF EXISTS tasks_labels, tasks, labels, users;

-- пользователи системы
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- метки задач
CREATE TABLE labels (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

-- задачи
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    opened TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- время создания задачи
    closed TIMESTAMP DEFAULT NULL,  -- время выполнения задачи
    author_id INTEGER REFERENCES users(id) DEFAULT NULL, -- автор задачи
    assigned_id INTEGER REFERENCES users(id) DEFAULT NULL, -- ответственный
    title TEXT, -- название задачи
    content TEXT -- описание задачи
);

-- связь многие - ко- многим между задачами и метками
CREATE TABLE tasks_labels (
    task_id INTEGER REFERENCES tasks(id),
    label_id INTEGER REFERENCES labels(id),
    PRIMARY KEY (task_id, label_id)
);

-- наполнение БД начальными данными
INSERT INTO users (id, name) VALUES (1, 'default');
