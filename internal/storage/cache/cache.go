package cache

import (
	"context"
	"github.com/dws33/WB_ZeroProj/internal/storage"
	"sync"

	"github.com/dws33/WB_ZeroProj/internal/model"
)

// CachedStorage — потокобезопасный кэш заказов в памяти.
type CachedStorage struct {
	cache   cache
	Storage *storage.Storage
}

type cache struct {
	mu     *sync.RWMutex
	orders map[string]*model.Order
}

func (c *cache) Get(uid string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.orders[uid]
	return order, ok
}

func (c *cache) Add(order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[order.OrderUID] = order
}

func (c *CachedStorage) CreateOrder(ctx context.Context, order *model.Order) error {
	err := c.Storage.CreateOrder(ctx, order)
	if err != nil {
		return err
	}
	c.cache.Add(order)
	return nil
}

func (c *CachedStorage) GetOrder(ctx context.Context, uid string) (*model.Order, error) {
	order, ok := c.cache.Get(uid)
	if ok {
		return order, nil
	}
	var err error
	order, err = c.Storage.GetOrder(ctx, uid)
	if err != nil {
		return nil, err
	}
	c.cache.Add(order)
	return order, nil
}

func New(ctx context.Context, store *storage.Storage) (*CachedStorage, error) {
	cs := &CachedStorage{
		cache: cache{
			mu:     new(sync.RWMutex),
			orders: make(map[string]*model.Order),
		},
		Storage: store,
	}

	orders, err := cs.Storage.GetAllOrders(ctx)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(orders); i++ {
		cs.cache.orders[orders[i].OrderUID] = orders[i]
	}
	return cs, nil
}
