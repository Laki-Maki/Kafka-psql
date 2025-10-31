package cache

import (
	lru "github.com/hashicorp/golang-lru"
	"order-service/internal/models"
)

type Cache struct {
	store *lru.Cache
}

func New(size int) *Cache {
	c, _ := lru.New(size)
	return &Cache{store: c}
}

func (c *Cache) Set(o models.Order) { c.store.Add(o.OrderUID, o) }

func (c *Cache) Get(id string) (models.Order, bool) {
	v, ok := c.store.Get(id)
	if !ok {
		return models.Order{}, false
	}
	return v.(models.Order), true
}

func (c *Cache) Load(list []models.Order) {
	for _, o := range list {
		c.Set(o)
	}
}

func (c *Cache) Len() int { return c.store.Len() }
