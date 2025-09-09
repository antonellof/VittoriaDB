package storage

import (
	"container/list"
	"sync"
)

// LRUCache implements a thread-safe LRU cache for pages
type LRUCache struct {
	capacity int
	cache    map[uint32]*list.Element
	lru      *list.List
	mu       sync.RWMutex
	hits     uint64
	misses   uint64
}

// CacheEntry represents a cache entry
type CacheEntry struct {
	key  uint32
	page *Page
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[uint32]*list.Element),
		lru:      list.New(),
	}
}

// Get retrieves a page from the cache
func (c *LRUCache) Get(pageID uint32) (*Page, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.cache[pageID]; found {
		// Move to front (most recently used)
		c.lru.MoveToFront(elem)
		c.hits++

		entry := elem.Value.(*CacheEntry)
		return c.copyPage(entry.page), true
	}

	c.misses++
	return nil, false
}

// Put adds a page to the cache
func (c *LRUCache) Put(pageID uint32, page *Page) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if page already exists
	if elem, found := c.cache[pageID]; found {
		// Update existing entry
		c.lru.MoveToFront(elem)
		entry := elem.Value.(*CacheEntry)
		entry.page = c.copyPage(page)
		return
	}

	// Add new entry
	entry := &CacheEntry{
		key:  pageID,
		page: c.copyPage(page),
	}

	elem := c.lru.PushFront(entry)
	c.cache[pageID] = elem

	// Evict if over capacity
	if c.lru.Len() > c.capacity {
		c.evictLRU()
	}
}

// Remove removes a page from the cache
func (c *LRUCache) Remove(pageID uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.cache[pageID]; found {
		c.lru.Remove(elem)
		delete(c.cache, pageID)
	}
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[uint32]*list.Element)
	c.lru = list.New()
	c.hits = 0
	c.misses = 0
}

// Stats returns cache statistics
func (c *LRUCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Size:     c.lru.Len(),
		Capacity: c.capacity,
		Hits:     c.hits,
		Misses:   c.misses,
		HitRate:  hitRate,
	}
}

// evictLRU removes the least recently used entry
func (c *LRUCache) evictLRU() {
	elem := c.lru.Back()
	if elem != nil {
		c.lru.Remove(elem)
		entry := elem.Value.(*CacheEntry)
		delete(c.cache, entry.key)
	}
}

// copyPage creates a deep copy of a page
func (c *LRUCache) copyPage(page *Page) *Page {
	data := make([]byte, len(page.Data))
	copy(data, page.Data)

	return &Page{
		ID:       page.ID,
		Type:     page.Type,
		Size:     page.Size,
		Flags:    page.Flags,
		LSN:      page.LSN,
		Checksum: page.Checksum,
		Data:     data,
	}
}
