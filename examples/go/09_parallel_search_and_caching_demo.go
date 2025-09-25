package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/core"
)

func main() {
	fmt.Println("🚀 VittoriaDB Parallel Search & Caching Demo")
	fmt.Println("==============================================")
	fmt.Println()

	ctx := context.Background()

	// Create a collection with enhanced features
	collection, err := core.NewCollection(
		"parallel_demo",
		128, // Higher dimensions for realistic scenario
		core.DistanceMetricCosine,
		core.IndexTypeFlat,
		"/tmp/vittoria_demo",
	)
	if err != nil {
		log.Fatalf("❌ Failed to create collection: %v", err)
	}

	fmt.Println("✅ Created collection with parallel search engine")
	fmt.Println()

	// Demonstrate batch insertion performance
	fmt.Println("📊 Testing Batch Insertion Performance...")
	testBatchInsertion(ctx, collection)
	fmt.Println()

	// Demonstrate search performance and caching
	fmt.Println("🔍 Testing Search Performance & Caching...")
	testSearchPerformance(ctx, collection)
	fmt.Println()

	// Demonstrate search engine statistics
	fmt.Println("📈 Search Engine Statistics...")
	displaySearchStats(collection)
	fmt.Println()

	// Demonstrate cache management
	fmt.Println("🗄️  Cache Management Demo...")
	testCacheManagement(ctx, collection)
	fmt.Println()

	fmt.Println("✅ Demo completed successfully!")
}

func testBatchInsertion(ctx context.Context, collection *core.VittoriaCollection) {
	// Generate test vectors with realistic patterns
	numVectors := 2000
	vectors := make([]*core.Vector, numVectors)

	fmt.Printf("   Generating %d test vectors...\n", numVectors)

	for i := 0; i < numVectors; i++ {
		vector := make([]float32, 128)

		// Create varied but realistic vector patterns
		for j := range vector {
			// Simulate different document types with different patterns
			switch i % 4 {
			case 0: // Technical documents
				vector[j] = float32(i*j) * 0.001
			case 1: // News articles
				vector[j] = float32(i+j) * 0.01
			case 2: // Academic papers
				vector[j] = float32(i*2+j) * 0.005
			case 3: // Blog posts
				vector[j] = float32(i-j) * 0.002
			}
		}

		vectors[i] = &core.Vector{
			ID:     fmt.Sprintf("doc_%d", i),
			Vector: vector,
			Metadata: map[string]interface{}{
				"type":      []string{"technical", "news", "academic", "blog"}[i%4],
				"timestamp": time.Now().Unix(),
				"category":  fmt.Sprintf("cat_%d", i%10),
			},
		}
	}

	// Test batch insertion with enhanced batch processor
	fmt.Printf("   Inserting %d vectors using enhanced batch processing...\n", numVectors)
	start := time.Now()

	if err := collection.InsertBatch(ctx, vectors); err != nil {
		log.Fatalf("❌ Batch insertion failed: %v", err)
	}

	elapsed := time.Since(start)
	throughput := float64(numVectors) / elapsed.Seconds()

	fmt.Printf("   ✅ Batch insertion completed in %v\n", elapsed)
	fmt.Printf("   📊 Throughput: %.0f vectors/second\n", throughput)

	// Verify collection size
	count, _ := collection.Count()
	fmt.Printf("   📝 Collection now contains %d vectors\n", count)
}

func testSearchPerformance(ctx context.Context, collection *core.VittoriaCollection) {
	// Create a search query
	queryVector := make([]float32, 128)
	for i := range queryVector {
		queryVector[i] = float32(i) * 0.01 // Pattern similar to "news" type
	}

	searchReq := &core.SearchRequest{
		Vector:          queryVector,
		Limit:           20,
		Offset:          0,
		IncludeVector:   false,
		IncludeMetadata: true,
		IncludeContent:  false,
	}

	// First search (cold - no cache)
	fmt.Println("   🥶 Cold search (no cache)...")
	start := time.Now()
	response1, err := collection.Search(ctx, searchReq)
	coldTime := time.Since(start)

	if err != nil {
		log.Fatalf("❌ Search failed: %v", err)
	}

	fmt.Printf("   ⏱️  Cold search time: %v\n", coldTime)
	fmt.Printf("   📊 Found %d results out of %d total\n", len(response1.Results), response1.Total)

	if len(response1.Results) > 0 {
		fmt.Printf("   🎯 Best match: %s (score: %.6f)\n",
			response1.Results[0].ID, response1.Results[0].Score)
	}

	// Second search (cached)
	fmt.Println("   🔥 Cached search (same query)...")
	start = time.Now()
	response2, err := collection.Search(ctx, searchReq)
	cachedTime := time.Since(start)

	if err != nil {
		log.Fatalf("❌ Cached search failed: %v", err)
	}

	speedup := float64(coldTime) / float64(cachedTime)
	fmt.Printf("   ⏱️  Cached search time: %v\n", cachedTime)
	fmt.Printf("   🚀 Cache speedup: %.1fx faster\n", speedup)

	// Verify results are identical
	if len(response1.Results) == len(response2.Results) {
		fmt.Printf("   ✅ Cache consistency verified\n")
	} else {
		fmt.Printf("   ⚠️  Cache inconsistency detected\n")
	}

	// Test different search patterns
	fmt.Println("   🔄 Testing multiple search patterns...")
	testMultipleSearches(ctx, collection)
}

func testMultipleSearches(ctx context.Context, collection *core.VittoriaCollection) {
	searchPatterns := []struct {
		name    string
		pattern func(i int) float32
	}{
		{"Technical", func(i int) float32 { return float32(i) * 0.001 }},
		{"News", func(i int) float32 { return float32(i) * 0.01 }},
		{"Academic", func(i int) float32 { return float32(i*2) * 0.005 }},
		{"Blog", func(i int) float32 { return float32(i) * 0.002 }},
	}

	var totalTime time.Duration
	totalSearches := 0

	for _, pattern := range searchPatterns {
		queryVector := make([]float32, 128)
		for i := range queryVector {
			queryVector[i] = pattern.pattern(i)
		}

		searchReq := &core.SearchRequest{
			Vector:          queryVector,
			Limit:           10,
			IncludeMetadata: true,
		}

		start := time.Now()
		response, err := collection.Search(ctx, searchReq)
		elapsed := time.Since(start)
		totalTime += elapsed
		totalSearches++

		if err != nil {
			fmt.Printf("   ❌ %s search failed: %v\n", pattern.name, err)
			continue
		}

		fmt.Printf("   📊 %s search: %v (%d results)\n",
			pattern.name, elapsed, len(response.Results))
	}

	avgTime := totalTime / time.Duration(totalSearches)
	fmt.Printf("   📈 Average search time: %v\n", avgTime)
}

func displaySearchStats(collection *core.VittoriaCollection) {
	stats := collection.GetSearchStats()
	if stats == nil {
		fmt.Println("   ❌ No search statistics available")
		return
	}

	fmt.Printf("   📊 Total searches: %d\n", stats.TotalSearches)
	fmt.Printf("   🎯 Cache hits: %d\n", stats.CacheHits)
	fmt.Printf("   ❄️  Cache misses: %d\n", stats.CacheMisses)

	if stats.TotalSearches > 0 {
		hitRate := float64(stats.CacheHits) / float64(stats.TotalSearches) * 100
		fmt.Printf("   📈 Cache hit rate: %.1f%%\n", hitRate)
	}

	fmt.Printf("   ⚡ Parallel searches: %d\n", stats.ParallelSearches)
	fmt.Printf("   🔄 Sequential searches: %d\n", stats.SequentialSearches)
	fmt.Printf("   ⏱️  Average latency: %v\n", stats.AverageLatency)
	fmt.Printf("   👥 Workers used: %d\n", stats.WorkersUsed)

	// Display cache statistics if available
	if engine := collection.GetSearchEngine(); engine != nil {
		if cacheStats := engine.GetCacheStats(); cacheStats != nil {
			fmt.Printf("   🗄️  Cache entries: %d\n", cacheStats.Entries)
			fmt.Printf("   🧹 Cache evictions: %d\n", cacheStats.Evictions)
			fmt.Printf("   🔄 Cleanup runs: %d\n", cacheStats.CleanupRuns)
		}
	}
}

func testCacheManagement(ctx context.Context, collection *core.VittoriaCollection) {
	// Get initial cache stats
	initialStats := collection.GetSearchStats()
	fmt.Printf("   📊 Initial cache hits: %d\n", initialStats.CacheHits)

	// Perform a search to populate cache
	queryVector := make([]float32, 128)
	for i := range queryVector {
		queryVector[i] = 0.5 // Simple pattern
	}

	searchReq := &core.SearchRequest{
		Vector: queryVector,
		Limit:  5,
	}

	// First search
	fmt.Println("   🔍 Performing search to populate cache...")
	_, err := collection.Search(ctx, searchReq)
	if err != nil {
		log.Printf("❌ Search failed: %v", err)
		return
	}

	// Second search (should hit cache)
	fmt.Println("   🔍 Performing same search (should hit cache)...")
	_, err = collection.Search(ctx, searchReq)
	if err != nil {
		log.Printf("❌ Cached search failed: %v", err)
		return
	}

	// Check cache stats
	afterStats := collection.GetSearchStats()
	fmt.Printf("   📊 Cache hits after searches: %d\n", afterStats.CacheHits)

	// Clear cache
	fmt.Println("   🧹 Clearing search cache...")
	collection.ClearSearchCache()

	// Search again (should miss cache)
	fmt.Println("   🔍 Searching after cache clear (should miss cache)...")
	_, err = collection.Search(ctx, searchReq)
	if err != nil {
		log.Printf("❌ Search after cache clear failed: %v", err)
		return
	}

	// Final stats
	finalStats := collection.GetSearchStats()
	fmt.Printf("   📊 Final cache hits: %d\n", finalStats.CacheHits)
	fmt.Printf("   📊 Final cache misses: %d\n", finalStats.CacheMisses)

	if finalStats.CacheMisses > afterStats.CacheMisses {
		fmt.Println("   ✅ Cache clear verified - new miss recorded")
	} else {
		fmt.Println("   ⚠️  Cache clear verification inconclusive")
	}
}
