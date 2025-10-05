package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	dbPool, err := initDB(ctx)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Здесь будет запуск inbox worker
	// ...
}

func initDB(ctx context.Context) (*pgxpool.Pool, error) {
	// Строка подключения из переменной окружения или напрямую
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Или соберём вручную
		databaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", "postgres"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "mydb"),
		)
	}

	// Конфигурация pool'а
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Настройки pool'а
	config.MaxConns = 10                      // максимум соединений
	config.MinConns = 2                       // минимум соединений
	config.MaxConnLifetime = time.Hour        // время жизни соединения
	config.MaxConnIdleTime = 30 * time.Minute // время простоя
	config.HealthCheckPeriod = time.Minute    // проверка здоровья

	// Создаём pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
