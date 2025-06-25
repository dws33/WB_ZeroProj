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
)

func main() {
	topic := "orders"
	partition := 0

	ctx := context.Background()

	p, err := kafka.DefaultDialer.LookupPartition(ctx, "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	p.Leader.Host = "localhost" // fixme

	c, err := kafka.DefaultDialer.DialPartition(ctx, "tcp", "localhost:9092", p)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
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

	store := storage.NewStorage(pgxPool)

	order := new(model.Order)
	for {
		m, err := c.ReadMessage(10e6)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(m.Value, order)
		if err != nil {
			log.Fatal(err)
		}
		err = store.CreateOrder(ctx, &*order)
		if err != nil {
			log.Fatal(err)
		}

	}

	//////////////////

	//r := kafka.NewReader(kafka.ReaderConfig{
	//	Brokers:   []string{"localhost:9092", "localhost:9093", "localhost:9094"},
	//	Topic:     "topic-A",
	//	Partition: 0,
	//	MaxBytes:  10e6, // 10MB
	//})
	//r.SetOffset(42)
	//
	//for {
	//	m, err := r.ReadMessage(context.Background())
	//	if err != nil {
	//		break
	//	}
	//	fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	//}
}
