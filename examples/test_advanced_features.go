package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/index"
	"github.com/antonellof/VittoriaDB/pkg/storage"
)

func main() {
	fmt.Println("🧪 VittoriaDB Advanced Features Test")
	fmt.Println("=========================================")

	// Test storage layer
	fmt.Println("\n1. Testing Storage Layer...")
	testStorageLayer()

	// Test flat index
	fmt.Println("\n2. Testing Flat Index...")
	testFlatIndex()

	// Test HNSW index
	fmt.Println("\n3. Testing HNSW Index...")
	testHNSWIndex()

	// Test index factory
	fmt.Println("\n4. Testing Index Factory...")
	testIndexFactory()

	fmt.Println("\n✅ All advanced features tests passed!")
}

func testStorageLayer() {
	// Create storage engine
	engine := storage.NewFileStorageEngine(100) // 100 page cache

	// Open storage
	if err := engine.Open("./test_storage.db"); err != nil {
		log.Fatalf("Failed to open storage: %v", err)
	}
	defer engine.Close()

	// Test page allocation
	pageID, err := engine.AllocatePage()
	if err != nil {
		log.Fatalf("Failed to allocate page: %v", err)
	}
	fmt.Printf("   ✓ Allocated page: %d\n", pageID)

	// Test page write/read
	testData := []byte("Hello, VittoriaDB Storage!")
	page := &storage.Page{
		ID:   pageID,
		Type: storage.PageTypeVectorLeaf,
		Size: uint16(len(testData)),
		Data: make([]byte, storage.PageSize-32),
	}
	copy(page.Data, testData)

	if err := engine.WritePage(page); err != nil {
		log.Fatalf("Failed to write page: %v", err)
	}
	fmt.Printf("   ✓ Wrote page with %d bytes\n", len(testData))

	// Read page back
	readPage, err := engine.ReadPage(pageID)
	if err != nil {
		log.Fatalf("Failed to read page: %v", err)
	}

	if string(readPage.Data[:len(testData)]) != string(testData) {
		log.Fatalf("Data mismatch: expected %s, got %s", testData, readPage.Data[:len(testData)])
	}
	fmt.Printf("   ✓ Read page data correctly\n")

	// Test transaction
	tx, err := engine.BeginTx()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// Modify page in transaction
	page.Data[0] = 'h' // Change 'H' to 'h'
	if err := tx.WritePage(page); err != nil {
		log.Fatalf("Failed to write page in transaction: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}
	fmt.Printf("   ✓ Transaction committed successfully\n")

	// Get storage stats
	stats := engine.Stats()
	fmt.Printf("   ✓ Storage stats: %d total pages, %.2f%% cache hit rate\n",
		stats.TotalPages, stats.CacheHitRate*100)
}

func testFlatIndex() {
	// Create flat index
	dimensions := 128
	metric := index.DistanceMetricCosine
	idx := index.NewFlatIndex(dimensions, metric, nil)

	// Generate test vectors
	vectors := generateTestVectors(1000, dimensions)

	// Build index
	start := time.Now()
	if err := idx.Build(vectors); err != nil {
		log.Fatalf("Failed to build flat index: %v", err)
	}
	buildTime := time.Since(start)
	fmt.Printf("   ✓ Built flat index with %d vectors in %v\n", len(vectors), buildTime)

	// Test search
	query := generateRandomVector(dimensions)
	start = time.Now()
	results, err := idx.Search(context.Background(), query, 10, nil)
	if err != nil {
		log.Fatalf("Failed to search flat index: %v", err)
	}
	searchTime := time.Since(start)
	fmt.Printf("   ✓ Search found %d results in %v\n", len(results), searchTime)

	// Test add/delete
	newVector := &index.IndexVector{
		ID:     "new_vector",
		Vector: generateRandomVector(dimensions),
	}
	if err := idx.Add(context.Background(), newVector); err != nil {
		log.Fatalf("Failed to add vector: %v", err)
	}
	fmt.Printf("   ✓ Added new vector, index size: %d\n", idx.Size())

	if err := idx.Delete(context.Background(), "new_vector"); err != nil {
		log.Fatalf("Failed to delete vector: %v", err)
	}
	fmt.Printf("   ✓ Deleted vector, index size: %d\n", idx.Size())

	// Get stats
	stats := idx.Stats()
	fmt.Printf("   ✓ Index stats: %d vectors, %d MB memory\n",
		stats.VectorCount, stats.MemoryUsage/1024/1024)
}

func testHNSWIndex() {
	// Create HNSW index
	dimensions := 128
	metric := index.DistanceMetricCosine
	config := index.DefaultHNSWConfig()
	config.M = 16
	config.EfConstruction = 200
	config.EfSearch = 50

	idx := index.NewHNSWIndex(dimensions, metric, config)

	// Generate test vectors
	vectors := generateTestVectors(5000, dimensions) // Smaller set for HNSW test

	// Build index
	start := time.Now()
	if err := idx.Build(vectors); err != nil {
		log.Fatalf("Failed to build HNSW index: %v", err)
	}
	buildTime := time.Since(start)
	fmt.Printf("   ✓ Built HNSW index with %d vectors in %v\n", len(vectors), buildTime)

	// Test search with different ef values
	query := generateRandomVector(dimensions)

	// Search with default ef
	start = time.Now()
	results1, err := idx.Search(context.Background(), query, 10, nil)
	if err != nil {
		log.Fatalf("Failed to search HNSW index: %v", err)
	}
	searchTime1 := time.Since(start)
	fmt.Printf("   ✓ Search (ef=50) found %d results in %v\n", len(results1), searchTime1)

	// Search with higher ef for better recall
	start = time.Now()
	results2, err := idx.Search(context.Background(), query, 10, &index.SearchParams{EF: 100})
	if err != nil {
		log.Fatalf("Failed to search HNSW index with ef=100: %v", err)
	}
	searchTime2 := time.Since(start)
	fmt.Printf("   ✓ Search (ef=100) found %d results in %v\n", len(results2), searchTime2)

	// Test add/delete
	newVector := &index.IndexVector{
		ID:     "hnsw_new_vector",
		Vector: generateRandomVector(dimensions),
	}
	if err := idx.Add(context.Background(), newVector); err != nil {
		log.Fatalf("Failed to add vector to HNSW: %v", err)
	}
	fmt.Printf("   ✓ Added new vector, index size: %d\n", idx.Size())

	// Get node info
	node := idx.GetNode(newVector.ID)
	if node != nil {
		fmt.Printf("   ✓ New vector is at layer %d with %d total connections\n",
			node.Layer, len(node.Connections))
	}

	// Get stats
	stats := idx.Stats()
	fmt.Printf("   ✓ HNSW stats: %d vectors, %d layers, %.1f avg degree, %d MB memory\n",
		stats.VectorCount, stats.MaxLayer, stats.AvgDegree, stats.MemoryUsage/1024/1024)
}

func testIndexFactory() {
	dimensions := 64

	// Test flat index creation
	flatIdx, err := index.CreateIndex(index.IndexTypeFlat, dimensions, index.DistanceMetricCosine, nil)
	if err != nil {
		log.Fatalf("Failed to create flat index: %v", err)
	}
	fmt.Printf("   ✓ Created %s index with %d dimensions\n", flatIdx.Type().String(), flatIdx.Dimensions())

	// Test HNSW index creation with custom config
	hnswConfig := map[string]interface{}{
		"m":               32,
		"ef_construction": 400,
		"ef_search":       100,
	}
	hnswIdx, err := index.CreateIndex(index.IndexTypeHNSW, dimensions, index.DistanceMetricEuclidean, hnswConfig)
	if err != nil {
		log.Fatalf("Failed to create HNSW index: %v", err)
	}
	fmt.Printf("   ✓ Created %s index with %d dimensions\n", hnswIdx.Type().String(), hnswIdx.Dimensions())

	// Test recommended configurations
	smallConfig := index.RecommendedConfig("small", dimensions, 1000)
	fmt.Printf("   ✓ Small dataset config: %v\n", smallConfig["index_type"])

	largeConfig := index.RecommendedConfig("large", dimensions, 1000000)
	fmt.Printf("   ✓ Large dataset config: %v\n", largeConfig["index_type"])

	// Test memory estimation
	memUsage := index.EstimateMemoryUsage(index.IndexTypeHNSW, dimensions, 100000, hnswConfig)
	fmt.Printf("   ✓ Estimated memory for 100k vectors: %d MB\n", memUsage/1024/1024)
}

func generateTestVectors(count, dimensions int) []*index.IndexVector {
	rand.Seed(42) // Fixed seed for reproducible results
	vectors := make([]*index.IndexVector, count)

	for i := 0; i < count; i++ {
		vectors[i] = &index.IndexVector{
			ID:     fmt.Sprintf("vec_%d", i),
			Vector: generateRandomVector(dimensions),
		}
	}

	return vectors
}

func generateRandomVector(dimensions int) []float32 {
	vector := make([]float32, dimensions)
	for i := 0; i < dimensions; i++ {
		vector[i] = rand.Float32()*2 - 1 // Random values between -1 and 1
	}

	// Normalize for cosine similarity
	var norm float32
	for _, v := range vector {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))

	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}

	return vector
}
