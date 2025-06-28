package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dws33/WB_ZeroProj/internal/model"
	"github.com/dws33/WB_ZeroProj/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"log"
	"net"
	"os"
)

func main() {

	ctx := context.Background()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
		os.Getenv("PGDATABASE"),
	)

	pgxPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pgxPool.Close()

	store, err := storage.New(ctx, pgxPool)
	if err != nil {
		log.Fatal(err)
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{net.JoinHostPort(os.Getenv("KAFKA_HOST"), os.Getenv("KAFKA_PORT"))},
		Topic:    os.Getenv("TOPIC_NAME"),
		MaxBytes: 10e6, // 10MB
	})
	defer r.Close()

	order := new(model.Order)
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Println("fail to read message", err)
			continue
		}
		err = json.Unmarshal(m.Value, order)
		if err != nil {
			log.Println("fail to unmarshal order", err)
			continue
		}
		err = order.Validate()
		if err != nil {
			log.Println("invalid order", err)
			continue
		}
		err = store.CreateOrder(ctx, &*order)
		if err != nil {
			log.Println("fail to save order in db", err)
			continue
		}
		log.Println("success handle order")
	}
}
