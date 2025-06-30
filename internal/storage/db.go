package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dws33/WB_ZeroProj/internal/model"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, pool *pgxpool.Pool) (*Storage, error) {
	err := pool.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return &Storage{pool: pool}, nil
}

func (s *Storage) CreateOrder(ctx context.Context, order *model.Order) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Вставка оплаты
	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (
			transactions_uid, request_id, currency, provider, amount,
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`,
		order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT,
		order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)
	if err != nil {
		return err
	}

	// Вставка в orders
	_, err = tx.Exec(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature, customer_id,
			delivery_service, shardkey, sm_id, date_created, oof_shard, payment_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID,
		order.DateCreated, order.OofShard, order.Payment.Transaction,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO deliveries (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`,
		order.OrderUID,
		order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region,
		order.Delivery.Email,
	)
	if err != nil {
		return err
	}

	copyCount, err := tx.CopyFrom(ctx,
		pgx.Identifier{"items"},
		[]string{
			"order_uid", "chrt_id", "track_number", "price", "rid", "name",
			"sale", "size", "total_price", "nm_id", "brand", "status",
		},
		pgx.CopyFromSlice(len(order.Items), func(i int) ([]any, error) {
			item := order.Items[i]
			return []any{
				order.OrderUID,
				item.ChrtID,
				item.TrackNumber,
				item.Price,
				item.RID,
				item.Name,
				item.Sale,
				item.Size,
				item.TotalPrice,
				item.NmID,
				item.Brand,
				item.Status,
			}, nil
		}),
	)
	if err != nil {
		return err
	}
	if copyCount != int64(len(order.Items)) {
		return fmt.Errorf("expected to insert %d items, but inserted %d", len(order.Items), copyCount)
	}

	return tx.Commit(ctx)
}

func (s *Storage) GetAllOrders(ctx context.Context) ([]*model.Order, error) {
	const ordersQuery = `
       SELECT
           o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id,
           o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
           d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
           t.transactions_uid, t.request_id, t.currency, t.provider, t.amount, t.payment_dt, t.bank, t.delivery_cost, t.goods_total, t.custom_fee
       FROM orders o
       LEFT JOIN deliveries d ON o.order_uid = d.order_uid
       LEFT JOIN transactions t ON o.payment_id = t.transactions_uid
    `

	rows, err := s.pool.Query(ctx, ordersQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order

	for rows.Next() {
		order := new(model.Order)

		err := rows.Scan(orderToPtrs(order)...)
		if err != nil {
			return nil, err
		}

		order.Items, err = getAllItems(ctx, s.pool, order.OrderUID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

type rowQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func getAllItems(ctx context.Context, q rowQueryer, orderId string) ([]*model.Item, error) {
	const itemsQuery = `
            SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
            FROM items
            WHERE order_uid = $1
        `
	rows, err := q.Query(ctx, itemsQuery, orderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.Item

	for rows.Next() {
		item := new(model.Item)
		err := rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.RID,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func orderToPtrs(order *model.Order) []any {
	order.Delivery = new(model.Delivery)
	order.Payment = new(model.Payment)

	return []any{
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.ShardKey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,

		&order.Delivery.Name,
		&order.Delivery.Phone,
		&order.Delivery.Zip,
		&order.Delivery.City,
		&order.Delivery.Address,
		&order.Delivery.Region,
		&order.Delivery.Email,

		&order.Payment.Transaction,
		&order.Payment.RequestID,
		&order.Payment.Currency,
		&order.Payment.Provider,
		&order.Payment.Amount,
		&order.Payment.PaymentDT,
		&order.Payment.Bank,
		&order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	}
}

func (s *Storage) GetOrder(ctx context.Context, uid string) (*model.Order, error) {
	const orderQuery = `
       SELECT
           o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id,
           o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
           d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
           t.transactions_uid, t.request_id, t.currency, t.provider, t.amount, t.payment_dt, t.bank, t.delivery_cost, t.goods_total, t.custom_fee
       FROM orders o
       LEFT JOIN deliveries d ON o.order_uid = d.order_uid
       LEFT JOIN transactions t ON o.payment_id = t.transactions_uid
       WHERE o.order_uid = $1
   `

	order := new(model.Order)

	err := s.pool.QueryRow(ctx, orderQuery, uid).Scan(orderToPtrs(order)...)
	if err != nil {
		return nil, err
	}

	order.Items, err = getAllItems(ctx, s.pool, order.OrderUID)
	if err != nil {
		return nil, err
	}

	return order, nil
}
