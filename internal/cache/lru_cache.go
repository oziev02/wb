package cache

import (
	"time"

	lru "github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/oziev02/wb/internal/domain"
)

// LRU + TTL. Решает проблему OOM при бесконечном росте ключей.
type OrdersCache struct {
	l *lru.LRU[string, domain.Order]
}

func NewOrdersCache(cap int, ttl time.Duration) *OrdersCache {
	return &OrdersCache{l: lru.NewLRU[string, domain.Order](cap, nil, ttl)}
}

func (c *OrdersCache) Get(id string) (domain.Order, bool) { return c.l.Get(id) }
func (c *OrdersCache) Set(o domain.Order)                 { c.l.Add(o.OrderUID, o) }
func (c *OrdersCache) BulkSet(orders []domain.Order) {
	for _, o := range orders {
		c.Set(o)
	}
}
