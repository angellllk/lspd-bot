package scraper

import (
	"sync"
)

type Cache struct {
	data map[string]string
	mu   sync.Mutex
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	profileURL, found := c.data[key]
	if !found {
		return "", false
	}
	return profileURL, true
}

func (c *Cache) Set(name string, profileURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[name] = profileURL
}

func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}
