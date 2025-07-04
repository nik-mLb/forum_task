package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

func InitDB(db *sql.DB) error {
    sqlBytes, err := os.ReadFile("db.sql")
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %v", err)
	}

	// Выполняем в транзакции для ускорения
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("failed to execute SQL script: %v", err)
	}

	// Оптимизация БД после инициализации
	optimizationQueries := []string{
		"ANALYZE", // Сбор статистики
		"SET synchronous_commit = off", // Уменьшение задержек записи
		"SET work_mem = '16MB'", // Увеличение памяти для операций
	}

	for _, query := range optimizationQueries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			log.Printf("Warning: optimization query failed: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Database schema initialized and optimized successfully")
	return nil
}