package main

import (
	"context"
	"fmt"
	"github.com/dws33/WB_ZeroProj/internal/handler"
	"github.com/dws33/WB_ZeroProj/internal/storage"
	"github.com/dws33/WB_ZeroProj/internal/storage/cache"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()

	// Загрузка переменных окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Using environment variables.")
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "user", "password", "orders_db")

	pgxPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pgxPool.Close()
	err = pgxPool.Ping(ctx)
	if err != nil {
		log.Fatal(err)
	}

	//// Загрузка данных в кэш из базы данных
	//cache, err := storage.LoadCacheFromDB(db)
	//if err != nil {
	//	log.Fatal("failed to initialize cache from db:", err)
	//}

	dbStore := storage.NewStorage(pgxPool)
	cachedStore := cache.NewCachedStorage(dbStore)
	h := handler.New(cachedStore)

	// HTTP маршруты
	mux := http.NewServeMux()
	mux.HandleFunc("GET /order/{order_uid}", h.GetOrder)

	// Запуск HTTP-сервера
	addr := ":8081"
	s := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Println("HTTP server started on", addr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %s", err)
	}
}
