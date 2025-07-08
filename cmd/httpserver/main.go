package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dws33/WB_ZeroProj/internal/handler"
	"github.com/dws33/WB_ZeroProj/internal/storage"
	"github.com/dws33/WB_ZeroProj/internal/storage/cache"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	pgxPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pgxPool.Close()

	dbStore, err := storage.New(ctx, pgxPool)
	if err != nil {
		log.Fatal(err)
	}
	cachedStore, err := cache.New(ctx, dbStore)
	if err != nil {
		log.Fatal(err)
	}
	h := handler.New(cachedStore)

	http.HandleFunc("GET /order/{order_uid}", h.GetOrder)

	addr := net.JoinHostPort(
		os.Getenv("SERVER_HOST"),
		os.Getenv("SERVER_PORT"))

	log.Println("HTTP server started on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil && errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %s", err)
	}
}
