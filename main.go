package main

import (
	"fmt"
	"log"
	"time"

	"30.8.1/storage"
)

func main() {
	// Подключаемся к базе данных
	connStr := "postgres://username:password@localhost:5432/taskdb"
	store, err := storage.New(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Пример создания задачи
	task := storage.Task{
		Title:      "Complete Project",
		Content:    "Finish the project by end of the week",
		AuthorID:   1,
		AssignedID: 2,
		Opened:     time.Now(),
	}

	// Создание задачи
	taskID, err := store.NewTask(task)
	if err != nil {
		log.Fatal("Failed to create task:", err)
	}
	fmt.Println("Created Task ID:", taskID)

	// Пример получения списка задач
	tasks, err := store.Tasks(0, 1)
	if err != nil {
		log.Fatal("Failed to get tasks:", err)
	}
	for _, t := range tasks {
		fmt.Printf("Task ID: %d, Title: %s\n", t.ID, t.Title)
	}

	// Пример закрытия задачи
	err = store.CloseTask(taskID)
	if err != nil {
		log.Fatal("Failed to close task:", err)
	}
	fmt.Println("Task closed successfully")

	// Пример обновления задачи
	task.ID = taskID
	task.Content = "Updated task content"
	err = store.UpdateTask(task)
	if err != nil {
		log.Fatal("Failed to update task:", err)
	}
	fmt.Println("Task updated successfully")

	// Пример удаления задачи
	err = store.DeleteTask(taskID)
	if err != nil {
		log.Fatal("Failed to delete task:", err)
	}
	fmt.Println("Task deleted successfully")
}
