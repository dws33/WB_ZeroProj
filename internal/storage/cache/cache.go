package cache

import (
	"context"
	"github.com/dws33/WB_ZeroProj/internal/storage"
	"sync"

	"github.com/dws33/WB_ZeroProj/internal/model"
)

// CachedStorage — потокобезопасный кэш заказов в памяти.
type CachedStorage struct {
	mu      sync.RWMutex
	orders  map[string]*model.Order
	Storage *storage.Storage
}

func (c *CachedStorage) CreateOrder(ctx context.Context, order *model.Order) error {
	err := c.Storage.CreateOrder(ctx, order)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.orders[order.OrderUID] = order
	c.mu.Unlock()
	return nil
}

func (c *CachedStorage) GetOrder(ctx context.Context, uid string) (*model.Order, error) {
	c.mu.RLock()
	order, ok := c.orders[uid]
	c.mu.RUnlock()
	if ok {
		return order, nil
	}
	var err error
	order, err = c.Storage.GetOrder(ctx, uid)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.orders[uid] = order
	c.mu.Unlock()
	return order, nil
}

func NewCachedStorage(ctx context.Context, store *storage.Storage) (*CachedStorage, error) {
	cs := &CachedStorage{
		orders:  make(map[string]*model.Order),
		Storage: store,
	}

	orders, err := cs.Storage.GetAllOrders(ctx)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(orders); i++ {
		cs.orders[orders[i].OrderUID] = &orders[i]
	}
	return cs, nil
}
