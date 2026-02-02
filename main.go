package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"order-service/internal/api"
	"order-service/internal/cache"
	"order-service/internal/database"
	"order-service/internal/service"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Подключаемся к PostgreSQL
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"),
	)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Инициализируем БД
	db := database.NewPostgresBase(pool)
	if err := db.InitDB(ctx); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized")

	// Инициализируем кэш
	cache := cache.NewCache()

	// Инициализируем сервис
	orderService := service.NewOrderService(db, cache)

	// Загружаем данные из БД в кэш
	if err := orderService.LoadCacheFromDB(ctx); err != nil {
		log.Printf("Warning: failed to load cache from DB: %v", err)
	}

	handler := api.NewHandler(orderService)

	// Настраиваем роуты
	mux := http.NewServeMux()
	handler.SetupRoutes(mux)

	// Запуск сервера
	server := &http.Server{
		Addr:         ":" + os.Getenv("SERVER_PORT"),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Канал для graceful shutdown
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v", err)
		}
		close(done)
	}()

	log.Printf("Server starting on port %s", os.Getenv("SERVER_PORT"))
	log.Println("Open http://localhost:" + os.Getenv("SERVER_PORT") + " in your browser")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v", os.Getenv("SERVER_PORT"), err)
	}

	<-done
	log.Println("Server stopped")
}
