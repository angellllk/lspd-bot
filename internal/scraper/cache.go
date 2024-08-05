package scraper

import (
	"sync"
)

// Cache provides a thread-safe mechanism for storing and retrieving cached data.
type Cache struct {
	data map[string]string
	mu   sync.Mutex
}

// Get retrieves a cached value by its key.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	profileURL, found := c.data[key]
	if !found {
		return "", false
	}
	return profileURL, true
}

// Set stores a key-value pair in the cache.
func (c *Cache) Set(name string, profileURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[name] = profileURL
}

// Invalidate removes a key-value pair from the cache.
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}
