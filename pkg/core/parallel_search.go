package core

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

// ParallelSearchConfig holds configuration for parallel search
type ParallelSearchConfig struct {
	Enabled        bool `json:"enabled" yaml:"enabled"`
	MaxWorkers     int  `json:"max_workers" yaml:"max_workers"`
	BatchSize      int  `json:"batch_size" yaml:"batch_size"`
	UseCache       bool `json:"use_cache" yaml:"use_cache"`
	PreloadVectors bool `json:"preload_vectors" yaml:"preload_vectors"`
}

// DefaultParallelSearchConfig returns sensible defaults
func DefaultParallelSearchConfig() *ParallelSearchConfig {
	return &ParallelSearchConfig{
		Enabled:        true,
		MaxWorkers:     runtime.NumCPU(),
		BatchSize:      100,
		UseCache:       true,
		PreloadVectors: false,
	}
}

// ParallelSearchEngine provides enhanced search capabilities
type ParallelSearchEngine struct {
	collection *VittoriaCollection
	cache      *SearchCache
	config     *ParallelSearchConfig
	stats      *ParallelSearchStats
	mu         sync.RWMutex
}

// ParallelSearchStats tracks search performance
type ParallelSearchStats struct {
	TotalSearches      int64         `json:"total_searches"`
	CacheHits          int64         `json:"cache_hits"`
	CacheMisses        int64         `json:"cache_misses"`
	AverageLatency     time.Duration `json:"average_latency"`
	ParallelSearches   int64         `json:"parallel_searches"`
	SequentialSearches int64         `json:"sequential_searches"`
	WorkersUsed        int           `json:"workers_used"`
}

// NewParallelSearchEngine creates a new parallel search engine
func NewParallelSearchEngine(collection *VittoriaCollection, config *ParallelSearchConfig) *ParallelSearchEngine {
	if config == nil {
		config = DefaultParallelSearchConfig()
	}

	var cache *SearchCache
	if config.UseCache {
		cache = NewSearchCache(DefaultSearchCacheConfig())
	}

	return &ParallelSearchEngine{
		collection: collection,
		cache:      cache,
		config:     config,
		stats:      &ParallelSearchStats{},
	}
}

// Search performs enhanced search with caching and parallel processing
func (pse *ParallelSearchEngine) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	startTime := time.Now()

	pse.mu.Lock()
	pse.stats.TotalSearches++
	pse.mu.Unlock()

	// Check cache first if enabled
	if pse.cache != nil {
		if cached, found := pse.cache.Get(req); found {
			pse.mu.Lock()
			pse.stats.CacheHits++
			pse.mu.Unlock()
			return cached, nil
		}
		pse.mu.Lock()
		pse.stats.CacheMisses++
		pse.mu.Unlock()
	}

	// Perform search
	var response *SearchResponse
	var err error

	if pse.config.Enabled && pse.shouldUseParallelSearch(req) {
		response, err = pse.parallelSearch(ctx, req)
		pse.mu.Lock()
		pse.stats.ParallelSearches++
		pse.mu.Unlock()
	} else {
		response, err = pse.collection.legacySearch(ctx, req)
		pse.mu.Lock()
		pse.stats.SequentialSearches++
		pse.mu.Unlock()
	}

	if err != nil {
		return nil, err
	}

	// Cache the result if caching is enabled
	if pse.cache != nil {
		pse.cache.Set(req, response)
	}

	// Update statistics
	latency := time.Since(startTime)
	pse.updateLatencyStats(latency)

	return response, nil
}

// SearchText performs text-based search with enhancements
func (pse *ParallelSearchEngine) SearchText(ctx context.Context, query string, limit int, filter *Filter) (*SearchResponse, error) {
	if pse.collection.vectorizer == nil {
		return nil, fmt.Errorf("no vectorizer configured for collection '%s'", pse.collection.name)
	}

	// Generate embedding from query text
	queryEmbedding, err := pse.collection.vectorizer.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Create search request
	searchReq := &SearchRequest{
		Vector:          queryEmbedding,
		Limit:           limit,
		Offset:          0,
		Filter:          filter,
		IncludeVector:   false,
		IncludeMetadata: true,
		IncludeContent:  true, // Include content for better results
	}

	return pse.Search(ctx, searchReq)
}

// parallelSearch performs search using multiple workers
func (pse *ParallelSearchEngine) parallelSearch(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	pse.collection.mu.RLock()
	defer pse.collection.mu.RUnlock()

	if pse.collection.closed {
		return nil, fmt.Errorf("collection is closed")
	}

	startTime := time.Now()

	// Validate search request
	if err := pse.collection.validateSearchRequest(req); err != nil {
		return nil, err
	}

	// Convert map to slice for parallel processing
	vectors := make([]*Vector, 0, len(pse.collection.vectors))
	for _, vector := range pse.collection.vectors {
		vectors = append(vectors, vector)
	}

	// Determine number of workers and batch size
	numWorkers := pse.config.MaxWorkers
	if numWorkers > len(vectors) {
		numWorkers = len(vectors)
	}
	if numWorkers <= 0 {
		numWorkers = 1
	}

	batchSize := (len(vectors) + numWorkers - 1) / numWorkers
	if batchSize < pse.config.BatchSize {
		batchSize = pse.config.BatchSize
	}

	// Channel for collecting results
	resultsChan := make(chan []*SearchResult, numWorkers)
	var wg sync.WaitGroup

	// Launch workers
	for i := 0; i < numWorkers; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > len(vectors) {
			end = len(vectors)
		}

		if start >= len(vectors) {
			break
		}

		wg.Add(1)
		go func(batch []*Vector) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				results := pse.processBatch(req, batch)
				resultsChan <- results
			}
		}(vectors[start:end])
	}

	// Close results channel when all workers complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results from all workers
	var allResults []*SearchResult
	for results := range resultsChan {
		allResults = append(allResults, results...)
	}

	// Sort by score (descending)
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	// Apply limit and offset
	start := req.Offset
	if start > len(allResults) {
		start = len(allResults)
	}

	end := start + req.Limit
	if end > len(allResults) {
		end = len(allResults)
	}

	finalResults := allResults[start:end]
	tookMS := time.Since(startTime).Milliseconds()

	pse.mu.Lock()
	pse.stats.WorkersUsed = numWorkers
	pse.mu.Unlock()

	return &SearchResponse{
		Results: finalResults,
		Total:   int64(len(allResults)),
		TookMS:  tookMS,
	}, nil
}

// processBatch processes a batch of vectors for similarity search
func (pse *ParallelSearchEngine) processBatch(req *SearchRequest, vectors []*Vector) []*SearchResult {
	var results []*SearchResult

	for _, vector := range vectors {
		// Apply metadata filter if specified
		if req.Filter != nil && !pse.collection.matchesFilter(vector.Metadata, req.Filter) {
			continue
		}

		// Calculate similarity score
		score := pse.collection.calculateSimilarity(req.Vector, vector.Vector)

		result := &SearchResult{
			ID:    vector.ID,
			Score: score,
		}

		// Include vector if requested
		if req.IncludeVector {
			result.Vector = make([]float32, len(vector.Vector))
			copy(result.Vector, vector.Vector)
		}

		// Include metadata if requested
		if req.IncludeMetadata {
			result.Metadata = make(map[string]interface{})
			for k, v := range vector.Metadata {
				result.Metadata[k] = v
			}
		}

		// Include content if requested and content storage is enabled
		if req.IncludeContent && pse.collection.contentStorage != nil && pse.collection.contentStorage.Enabled {
			if content, exists := vector.Metadata[pse.collection.contentStorage.FieldName]; exists {
				if contentStr, ok := content.(string); ok {
					result.Content = contentStr
				}
			}
		}

		results = append(results, result)
	}

	return results
}

// shouldUseParallelSearch determines if parallel search should be used
func (pse *ParallelSearchEngine) shouldUseParallelSearch(req *SearchRequest) bool {
	// Use parallel search for larger datasets or when specifically beneficial
	vectorCount := len(pse.collection.vectors)

	// Use parallel search if we have enough vectors to benefit from parallelization
	minVectorsForParallel := pse.config.MaxWorkers * pse.config.BatchSize

	return vectorCount >= minVectorsForParallel
}

// updateLatencyStats updates average latency statistics
func (pse *ParallelSearchEngine) updateLatencyStats(latency time.Duration) {
	pse.mu.Lock()
	defer pse.mu.Unlock()

	// Simple moving average calculation
	if pse.stats.AverageLatency == 0 {
		pse.stats.AverageLatency = latency
	} else {
		// Weighted average: 90% old, 10% new
		pse.stats.AverageLatency = time.Duration(
			float64(pse.stats.AverageLatency)*0.9 + float64(latency)*0.1,
		)
	}
}

// GetStats returns current search statistics
func (pse *ParallelSearchEngine) GetStats() ParallelSearchStats {
	pse.mu.RLock()
	defer pse.mu.RUnlock()

	stats := *pse.stats

	// Add cache stats if available
	if pse.cache != nil {
		cacheStats := pse.cache.GetStats()
		stats.CacheHits = cacheStats.Hits
		stats.CacheMisses = cacheStats.Misses
	}

	return stats
}

// GetCacheStats returns cache statistics
func (pse *ParallelSearchEngine) GetCacheStats() *SearchCacheStats {
	if pse.cache == nil {
		return nil
	}

	stats := pse.cache.GetStats()
	return &stats
}

// ClearCache clears the search cache
func (pse *ParallelSearchEngine) ClearCache() {
	if pse.cache != nil {
		pse.cache.Clear()
	}
}

// UpdateConfig updates the parallel search configuration
func (pse *ParallelSearchEngine) UpdateConfig(config *ParallelSearchConfig) {
	pse.mu.Lock()
	defer pse.mu.Unlock()

	pse.config = config

	// Update cache if needed
	if config.UseCache && pse.cache == nil {
		pse.cache = NewSearchCache(DefaultSearchCacheConfig())
	} else if !config.UseCache && pse.cache != nil {
		pse.cache.Close()
		pse.cache = nil
	}
}

// Close cleans up resources
func (pse *ParallelSearchEngine) Close() {
	if pse.cache != nil {
		pse.cache.Close()
	}
}
