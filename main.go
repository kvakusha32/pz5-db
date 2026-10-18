package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("укажите DATABASE_URL в файле .env")
	}

	db, err := openDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")
	repo := NewRepo(db)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	titles := []string{"Сделать ПЗ №5", "Купить кофе", "Проверить отчёты"}
	var firstID int
	for i, title := range titles {
		id, err := repo.CreateTask(ctx, title)
		if err != nil {
			log.Fatalf("CreateTask: %v", err)
		}
		if i == 0 {
			firstID = id
		}
		log.Printf("Inserted task id=%d (%s)", id, title)
	}

	if err := repo.CreateMany(ctx, []string{"Задача из транзакции 1", "Задача из транзакции 2"}); err != nil {
		log.Fatalf("CreateMany: %v", err)
	}
	fmt.Println("Две задачи добавлены одной транзакцией")

	started := time.Now()
	tasks, err := repo.ListTasks(ctx)
	if err != nil {
		log.Fatalf("ListTasks: %v", err)
	}
	fmt.Printf("Чтение списка заняло %s\n", time.Since(started))
	fmt.Println("=== Все задачи ===")
	printTasks(tasks)

	doneTasks, err := repo.ListDone(ctx, true)
	if err != nil {
		log.Fatalf("ListDone(true): %v", err)
	}
	fmt.Println("=== Выполненные ===")
	printTasks(doneTasks)

	openTasks, err := repo.ListDone(ctx, false)
	if err != nil {
		log.Fatalf("ListDone(false): %v", err)
	}
	fmt.Println("=== Невыполненные ===")
	printTasks(openTasks)

	task, err := repo.FindByID(ctx, firstID)
	if err != nil {
		log.Fatalf("FindByID: %v", err)
	}
	fmt.Printf("Найдена задача по id=%d: %s\n", task.ID, task.Title)

	_, err = repo.FindByID(ctx, 0)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println("Задача с id=0 не найдена — это ожидаемо")
	} else if err != nil {
		log.Fatalf("FindByID(0): %v", err)
	}

	stats := db.Stats()
	fmt.Printf("Пул: открыто=%d, используется=%d, свободно=%d, ожиданий=%d\n",
		stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
}

func printTasks(tasks []Task) {
	if len(tasks) == 0 {
		fmt.Println("(пусто)")
		return
	}
	for _, task := range tasks {
		fmt.Printf("#%d | %s | done=%t | %s\n",
			task.ID, task.Title, task.Done, task.CreatedAt.Format(time.RFC3339))
	}
}
