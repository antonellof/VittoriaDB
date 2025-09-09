package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/index"
)

func main() {
	fmt.Println("🧪 VittoriaDB Simple Index Test")
	fmt.Println("=================================")

	// Test flat index
	fmt.Println("\n1. Testing Flat Index...")
	testFlatIndex()

	// Test HNSW index
	fmt.Println("\n2. Testing HNSW Index...")
	testHNSWIndex()

	// Test index factory
	fmt.Println("\n3. Testing Index Factory...")
	testIndexFactory()

	fmt.Println("\n✅ All tests passed!")
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
	vectors := generateTestVectors(2000, dimensions) // Smaller set for HNSW test

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
	if hnswIdx, ok := idx.(index.HNSWIndex); ok {
		node := hnswIdx.GetNode(newVector.ID)
		if node != nil {
			totalConnections := 0
			for _, connections := range node.Connections {
				totalConnections += len(connections)
			}
			fmt.Printf("   ✓ New vector is at layer %d with %d total connections\n",
				node.Layer, totalConnections)
		}
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
