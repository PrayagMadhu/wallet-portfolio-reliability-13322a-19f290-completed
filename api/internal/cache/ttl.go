package cache

import "time"

type entry struct {
	value     float64
	expiresAt time.Time
}

type TTLCache struct {
	ttl   time.Duration
	items map[string]entry
}

func NewTTLCache(ttl time.Duration) *TTLCache {
	return &TTLCache{ttl: ttl, items: map[string]entry{}}
}

func (c *TTLCache) Get(key string) (float64, bool) {
	item, ok := c.items[key]
	if !ok {
		return 0, false
	}
	if time.Now().After(item.expiresAt) {
		delete(c.items, key)
		return 0, false
	}
	return item.value, true
}

func (c *TTLCache) Set(key string, value float64) {
	c.items[key] = entry{value: value, expiresAt: time.Now().Add(c.ttl)}
}
