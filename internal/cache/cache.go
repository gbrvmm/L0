package cache

import (
	"sync"

	"github.com/gbrvmm/L0/internal/model"
)

type Cache struct {
	mu sync.RWMutex
	m  map[string]model.Order
}

func New() *Cache {
	return &Cache{m: make(map[string]model.Order)}
}

func (c *Cache) Get(id string) (model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[id]
	return v, ok
}

func (c *Cache) Set(id string, v model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[id] = v
}

func (c *Cache) SetMany(values map[string]model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range values {
		c.m[k] = v
	}
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.m)
}
