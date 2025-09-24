package core

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// SearchCacheConfig holds configuration for search caching
type SearchCacheConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	MaxEntries     int           `json:"max_entries" yaml:"max_entries"`
	TTL            time.Duration `json:"ttl" yaml:"ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// DefaultSearchCacheConfig returns sensible defaults for search caching
func DefaultSearchCacheConfig() *SearchCacheConfig {
	return &SearchCacheConfig{
		Enabled:         true,
		MaxEntries:      1000,
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
}

// CacheEntry represents a cached search result
type CacheEntry struct {
	Key        string          `json:"key"`
	Response   *SearchResponse `json:"response"`
	CreatedAt  time.Time       `json:"created_at"`
	AccessedAt time.Time       `json:"accessed_at"`
	AccessCount int64          `json:"access_count"`
}

// SearchCache provides caching for search results
type SearchCache struct {
	config  *SearchCacheConfig
	entries map[string]*CacheEntry
	mu      sync.RWMutex
	stats   *SearchCacheStats
	stopCh  chan struct{}
}

// SearchCacheStats tracks cache performance
type SearchCacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Entries     int     `json:"entries"`
	HitRate     float64 `json:"hit_rate"`
	Evictions   int64   `json:"evictions"`
	CleanupRuns int64   `json:"cleanup_runs"`
}

// NewSearchCache creates a new search cache
func NewSearchCache(config *SearchCacheConfig) *SearchCache {
	if config == nil {
		config = DefaultSearchCacheConfig()
	}

	cache := &SearchCache{
		config:  config,
		entries: make(map[string]*CacheEntry),
		stats:   &SearchCacheStats{},
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine if enabled
	if config.Enabled && config.CleanupInterval > 0 {
		go cache.cleanupLoop()
	}

	return cache
}

// Get retrieves a cached search result
func (sc *SearchCache) Get(req *SearchRequest) (*SearchResponse, bool) {
	if !sc.config.Enabled {
		return nil, false
	}

	key := sc.generateKey(req)
	
	sc.mu.RLock()
	entry, exists := sc.entries[key]
	sc.mu.RUnlock()

	if !exists {
		sc.incrementMisses()
		return nil, false
	}

	// Check if entry is expired
	if time.Since(entry.CreatedAt) > sc.config.TTL {
		sc.mu.Lock()
		delete(sc.entries, key)
		sc.mu.Unlock()
		sc.incrementMisses()
		return nil, false
	}

	// Update access statistics
	sc.mu.Lock()
	entry.AccessedAt = time.Now()
	entry.AccessCount++
	sc.mu.Unlock()

	sc.incrementHits()
	return entry.Response, true
}

// Set stores a search result in the cache
func (sc *SearchCache) Set(req *SearchRequest, response *SearchResponse) {
	if !sc.config.Enabled {
		return
	}

	key := sc.generateKey(req)
	now := time.Now()

	entry := &CacheEntry{
		Key:         key,
		Response:    sc.copyResponse(response),
		CreatedAt:   now,
		AccessedAt:  now,
		AccessCount: 1,
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Check if we need to evict entries
	if len(sc.entries) >= sc.config.MaxEntries {
		sc.evictLRU()
	}

	sc.entries[key] = entry
}

// Clear removes all cached entries
func (sc *SearchCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.entries = make(map[string]*CacheEntry)
	sc.stats.Evictions += int64(len(sc.entries))
}

// GetStats returns current cache statistics
func (sc *SearchCache) GetStats() SearchCacheStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	stats := *sc.stats
	stats.Entries = len(sc.entries)
	
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
	}

	return stats
}

// Close stops the cache cleanup goroutine
func (sc *SearchCache) Close() {
	if sc.stopCh != nil {
		close(sc.stopCh)
	}
}

// generateKey creates a cache key from a search request
func (sc *SearchCache) generateKey(req *SearchRequest) string {
	// Create a deterministic key from the request
	keyData := struct {
		Vector          []float32              `json:"vector"`
		Limit           int                    `json:"limit"`
		Offset          int                    `json:"offset"`
		Filter          *Filter                `json:"filter"`
		IncludeVector   bool                   `json:"include_vector"`
		IncludeMetadata bool                   `json:"include_metadata"`
		IncludeContent  bool                   `json:"include_content"`
	}{
		Vector:          req.Vector,
		Limit:           req.Limit,
		Offset:          req.Offset,
		Filter:          req.Filter,
		IncludeVector:   req.IncludeVector,
		IncludeMetadata: req.IncludeMetadata,
		IncludeContent:  req.IncludeContent,
	}

	data, _ := json.Marshal(keyData)
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// copyResponse creates a deep copy of a search response
func (sc *SearchCache) copyResponse(response *SearchResponse) *SearchResponse {
	if response == nil {
		return nil
	}

	responseCopy := &SearchResponse{
		Results: make([]*SearchResult, len(response.Results)),
		Total:   response.Total,
		TookMS:  response.TookMS,
	}

	for i, result := range response.Results {
		responseCopy.Results[i] = &SearchResult{
			ID:    result.ID,
			Score: result.Score,
		}

		if result.Vector != nil {
			responseCopy.Results[i].Vector = make([]float32, len(result.Vector))
			copy(responseCopy.Results[i].Vector, result.Vector)
		}

		if result.Metadata != nil {
			responseCopy.Results[i].Metadata = make(map[string]interface{})
			for k, v := range result.Metadata {
				responseCopy.Results[i].Metadata[k] = v
			}
		}

		responseCopy.Results[i].Content = result.Content
	}

	return responseCopy
}

// evictLRU removes the least recently used entry
func (sc *SearchCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range sc.entries {
		if first || entry.AccessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.AccessedAt
			first = false
		}
	}

	if oldestKey != "" {
		delete(sc.entries, oldestKey)
		sc.stats.Evictions++
	}
}

// cleanupLoop periodically removes expired entries
func (sc *SearchCache) cleanupLoop() {
	ticker := time.NewTicker(sc.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sc.cleanup()
		case <-sc.stopCh:
			return
		}
	}
}

// cleanup removes expired entries
func (sc *SearchCache) cleanup() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, entry := range sc.entries {
		if now.Sub(entry.CreatedAt) > sc.config.TTL {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(sc.entries, key)
	}

	sc.stats.CleanupRuns++
	if len(expiredKeys) > 0 {
		sc.stats.Evictions += int64(len(expiredKeys))
	}
}

// incrementHits safely increments the hit counter
func (sc *SearchCache) incrementHits() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.stats.Hits++
}

// incrementMisses safely increments the miss counter
func (sc *SearchCache) incrementMisses() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.stats.Misses++
}
