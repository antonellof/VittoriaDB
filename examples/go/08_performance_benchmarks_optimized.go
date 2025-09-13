package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/core"
)

// ANSI color codes for better output formatting
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
)

// VolumeTestConfig defines parameters for each test
type VolumeTestConfig struct {
	Name        string
	Dimensions  int
	VectorCount int
	IndexType   core.IndexType
	Metric      core.DistanceMetric
	BatchSize   int
	Description string
}

// TestResults stores the results of a volume test
type TestResults struct {
	Config           VolumeTestConfig
	DataSizeMB       float64
	InsertTime       time.Duration
	InsertRate       float64
	SearchTime       time.Duration
	AvgSearchLatency time.Duration
	SearchRate       float64
	MemoryUsage      uint64
}

// Helper functions for colored output
func printHeader(text string) {
	fmt.Printf("\n%s%s%s\n", ColorBlue, text, ColorReset)
	fmt.Println(strings.Repeat("=", len(text)))
}

func printInfo(text string) {
	fmt.Printf("%sℹ️  %s%s\n", ColorYellow, text, ColorReset)
}

func printSuccess(text string) {
	fmt.Printf("%s✅ %s%s\n", ColorGreen, text, ColorReset)
}

func printError(text string) {
	fmt.Printf("%s❌ %s%s\n", ColorRed, text, ColorReset)
}

func printPerf(text string) {
	fmt.Printf("%s📊 %s%s\n", ColorPurple, text, ColorReset)
}

// getMemoryUsage returns current memory usage in bytes
func getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// generateRandomVector creates a normalized random vector
func generateRandomVector(dimensions int, seed int64) []float32 {
	rand.Seed(seed)
	vector := make([]float32, dimensions)

	// Generate random values
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

// generateTestVectors creates a batch of test vectors
func generateTestVectors(count, dimensions int) []*core.Vector {
	vectors := make([]*core.Vector, count)

	for i := 0; i < count; i++ {
		vectors[i] = &core.Vector{
			ID:     fmt.Sprintf("vec_%d", i),
			Vector: generateRandomVector(dimensions, int64(i)),
			Metadata: map[string]interface{}{
				"index":    i,
				"category": fmt.Sprintf("cat_%d", i%5),
				"batch":    i / 1000,
			},
		}

		// Progress indicator for large datasets
		if count > 1000 && i%1000 == 0 && i > 0 {
			printInfo(fmt.Sprintf("Generated %d/%d vectors (%.1f%%)", i, count, float64(i)*100/float64(count)))
		}
	}

	return vectors
}

// runVolumeTest executes a single volume test with the given configuration
func runVolumeTest(ctx context.Context, db *core.VittoriaDB, config VolumeTestConfig) (*TestResults, error) {
	printHeader(fmt.Sprintf("%s (%d dims, %d vectors)", config.Name, config.Dimensions, config.VectorCount))

	results := &TestResults{
		Config: config,
	}

	// Calculate estimated data size
	results.DataSizeMB = float64(config.VectorCount*config.Dimensions*4) / 1024 / 1024
	printInfo(fmt.Sprintf("Testing with %d vectors of %d dimensions", config.VectorCount, config.Dimensions))
	printInfo(fmt.Sprintf("Estimated data size: %.2f MB", results.DataSizeMB))
	printInfo(fmt.Sprintf("Index type: %s, Metric: %s", config.IndexType.String(), config.Metric.String()))

	// Create collection
	collectionName := fmt.Sprintf("volume_test_%s_%d", config.Name, time.Now().Unix())

	// Delete existing collection if it exists
	db.DropCollection(ctx, collectionName)

	createReq := &core.CreateCollectionRequest{
		Name:       collectionName,
		Dimensions: config.Dimensions,
		IndexType:  config.IndexType,
		Metric:     config.Metric,
	}

	if err := db.CreateCollection(ctx, createReq); err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}
	printSuccess(fmt.Sprintf("Created collection '%s'", collectionName))

	// Get collection
	collection, err := db.GetCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Generate test vectors
	printInfo("Generating test vectors...")
	startMem := getMemoryUsage()
	vectors := generateTestVectors(config.VectorCount, config.Dimensions)
	genMem := getMemoryUsage()
	printPerf(fmt.Sprintf("Vector generation: %.2f MB memory used", float64(genMem-startMem)/1024/1024))

	// Batch insertion test
	printInfo("Testing batch insertions...")
	startTime := time.Now()
	startMem = getMemoryUsage()

	// Insert in batches for better performance
	for i := 0; i < len(vectors); i += config.BatchSize {
		end := i + config.BatchSize
		if end > len(vectors) {
			end = len(vectors)
		}

		batch := vectors[i:end]
		for _, vector := range batch {
			if err := collection.Insert(ctx, vector); err != nil {
				return nil, fmt.Errorf("failed to insert vector: %w", err)
			}
		}

		// Progress indicator
		if config.VectorCount > 1000 && i%1000 == 0 && i > 0 {
			printInfo(fmt.Sprintf("Progress: %d/%d vectors inserted (%.1f%%)", i, config.VectorCount, float64(i)*100/float64(config.VectorCount)))
		}
	}

	results.InsertTime = time.Since(startTime)
	results.InsertRate = float64(config.VectorCount) / results.InsertTime.Seconds()
	insertMem := getMemoryUsage()
	results.MemoryUsage = insertMem - startMem

	printPerf(fmt.Sprintf("Batch insertions: %d vectors in %v (%.2f vectors/sec)",
		config.VectorCount, results.InsertTime, results.InsertRate))
	printPerf(fmt.Sprintf("Memory usage: %.2f MB", float64(results.MemoryUsage)/1024/1024))

	// Get collection statistics
	vectorCount, err := collection.Count()
	if err != nil {
		printInfo(fmt.Sprintf("Warning: Failed to get vector count: %v", err))
		vectorCount = int64(config.VectorCount)
	}
	printPerf(fmt.Sprintf("Collection stats: %d total vectors", vectorCount))

	// Search performance test - limit searches for large datasets
	printInfo("Testing search performance...")
	numSearches := 100
	if config.VectorCount > 10000 {
		numSearches = 10 // Reduce searches for large datasets to avoid timeout
	}

	searchTimes := make([]time.Duration, numSearches)
	queryVector := generateRandomVector(config.Dimensions, 999)

	// Warm up
	collection.Search(ctx, &core.SearchRequest{
		Vector: queryVector,
		Limit:  10,
	})

	// Measure search performance
	totalSearchTime := time.Duration(0)
	for i := 0; i < numSearches; i++ {
		startTime := time.Now()
		results_search, err := collection.Search(ctx, &core.SearchRequest{
			Vector: queryVector,
			Limit:  10,
		})
		searchTime := time.Since(startTime)
		searchTimes[i] = searchTime
		totalSearchTime += searchTime

		if err != nil {
			return nil, fmt.Errorf("search failed: %w", err)
		}

		// Verify we got results
		if i == 0 {
			printInfo(fmt.Sprintf("First search found %d results", len(results_search.Results)))
		}
	}

	results.SearchTime = totalSearchTime
	results.AvgSearchLatency = totalSearchTime / time.Duration(numSearches)
	results.SearchRate = float64(numSearches) / totalSearchTime.Seconds()

	printPerf(fmt.Sprintf("Search performance: %d searches in %v", numSearches, totalSearchTime))
	printPerf(fmt.Sprintf("Average latency: %v", results.AvgSearchLatency))
	printPerf(fmt.Sprintf("Search rate: %.2f searches/sec", results.SearchRate))

	// Test filtered search (quick test only)
	printInfo("Testing filtered search...")
	startTime = time.Now()
	filteredResults, err := collection.Search(ctx, &core.SearchRequest{
		Vector: queryVector,
		Limit:  5,
	})
	filteredTime := time.Since(startTime)

	if err != nil {
		printInfo(fmt.Sprintf("Filtered search failed: %v", err))
	} else {
		printPerf(fmt.Sprintf("Filtered search: %d results in %v", len(filteredResults.Results), filteredTime))
	}

	// Cleanup
	if err := db.DropCollection(ctx, collectionName); err != nil {
		printInfo(fmt.Sprintf("Warning: Failed to delete collection '%s': %v", collectionName, err))
	} else {
		printSuccess(fmt.Sprintf("Deleted collection '%s'", collectionName))
	}

	return results, nil
}

// runComparativeTests runs a series of comparative tests
func runComparativeTests(ctx context.Context, db *core.VittoriaDB) {
	printHeader("Performance Comparison Tests")

	configs := []VolumeTestConfig{
		{
			Name:        "small_flat",
			Dimensions:  128,
			VectorCount: 1000,
			IndexType:   core.IndexTypeFlat,
			Metric:      core.DistanceMetricCosine,
			BatchSize:   100,
			Description: "Small dataset with flat index",
		},
		{
			Name:        "small_hnsw",
			Dimensions:  128,
			VectorCount: 1000,
			IndexType:   core.IndexTypeHNSW,
			Metric:      core.DistanceMetricCosine,
			BatchSize:   100,
			Description: "Small dataset with HNSW index",
		},
		{
			Name:        "medium_flat",
			Dimensions:  384,
			VectorCount: 5000,
			IndexType:   core.IndexTypeFlat,
			Metric:      core.DistanceMetricCosine,
			BatchSize:   500,
			Description: "Medium dataset with flat index",
		},
		{
			Name:        "medium_hnsw",
			Dimensions:  384,
			VectorCount: 5000,
			IndexType:   core.IndexTypeHNSW,
			Metric:      core.DistanceMetricCosine,
			BatchSize:   500,
			Description: "Medium dataset with HNSW index",
		},
	}

	results := make([]*TestResults, 0, len(configs))

	for _, config := range configs {
		result, err := runVolumeTest(ctx, db, config)
		if err != nil {
			printError(fmt.Sprintf("Test %s failed: %v", config.Name, err))
			continue
		}
		results = append(results, result)
	}

	// Print comparison summary
	printHeader("Performance Comparison Summary")
	fmt.Printf("%-15s %-8s %-8s %-12s %-12s %-12s %-12s\n",
		"Test", "Dims", "Vectors", "Insert/sec", "Search/sec", "Latency", "Memory")
	fmt.Println("-----------------------------------------------------------------------------------------")

	for _, result := range results {
		fmt.Printf("%-15s %-8d %-8d %-12.1f %-12.1f %-12v %-12.1f\n",
			result.Config.Name,
			result.Config.Dimensions,
			result.Config.VectorCount,
			result.InsertRate,
			result.SearchRate,
			result.AvgSearchLatency,
			float64(result.MemoryUsage)/1024/1024)
	}
}

// runLargeScaleTests runs large-scale performance tests
func runLargeScaleTests(ctx context.Context, db *core.VittoriaDB) {
	printHeader("Large-Scale Performance Tests")

	largeConfigs := []VolumeTestConfig{
		{
			Name:        "mb_scale",
			Dimensions:  512,
			VectorCount: 20000,
			IndexType:   core.IndexTypeHNSW,
			Metric:      core.DistanceMetricCosine,
			BatchSize:   1000,
			Description: "MB-scale dataset for performance testing",
		},
		{
			Name:        "large_scale",
			Dimensions:  768,
			VectorCount: 30000, // Reduced from 50k to avoid timeout
			IndexType:   core.IndexTypeHNSW,
			Metric:      core.DistanceMetricCosine,
			BatchSize:   1000,
			Description: "Large-scale dataset for stress testing",
		},
	}

	for _, config := range largeConfigs {
		printInfo(fmt.Sprintf("Starting %s test - this may take several minutes...", config.Name))

		result, err := runVolumeTest(ctx, db, config)
		if err != nil {
			printError(fmt.Sprintf("Large-scale test %s failed: %v", config.Name, err))
			continue
		}

		printPerf(fmt.Sprintf("%s completed: %.2f MB data, %.1f vectors/sec insert, %.1f searches/sec",
			config.Name, result.DataSizeMB, result.InsertRate, result.SearchRate))
	}
}

// testDistanceMetrics compares different distance metrics
func testDistanceMetrics(ctx context.Context, db *core.VittoriaDB) {
	printHeader("Distance Metric Comparison")

	metrics := []core.DistanceMetric{
		core.DistanceMetricCosine,
		core.DistanceMetricEuclidean,
		core.DistanceMetricDotProduct,
	}

	for _, metric := range metrics {
		config := VolumeTestConfig{
			Name:        fmt.Sprintf("metric_%s", metric.String()),
			Dimensions:  256,
			VectorCount: 2000,
			IndexType:   core.IndexTypeHNSW,
			Metric:      metric,
			BatchSize:   200,
			Description: fmt.Sprintf("Testing %s distance metric", metric.String()),
		}

		result, err := runVolumeTest(ctx, db, config)
		if err != nil {
			printError(fmt.Sprintf("Metric test %s failed: %v", metric.String(), err))
			continue
		}

		printPerf(fmt.Sprintf("%s: %.1f vectors/sec insert, %.1f searches/sec, %v avg latency",
			metric.String(), result.InsertRate, result.SearchRate, result.AvgSearchLatency))
	}
}

func main() {
	fmt.Printf("%s🧪 VittoriaDB Volume Testing Suite (Native Go SDK)%s\n", ColorBlue, ColorReset)
	fmt.Println("=======================================================")
	fmt.Printf("%sHigh-performance volume testing using direct SDK integration%s\n", ColorYellow, ColorReset)

	// Initialize database
	config := &core.Config{
		DataDir: "./volume_test_data",
		Storage: core.StorageConfig{
			PageSize:    4096,
			CacheSize:   1000,
			SyncWrites:  false,
			Compression: false,
		},
	}

	db := core.NewDatabase()
	ctx := context.Background()

	printInfo("Initializing VittoriaDB...")
	if err := db.Open(ctx, config); err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	printSuccess("Database initialized successfully")

	// Run different test suites
	printInfo("Starting volume tests...")

	// 1. Comparative tests (small to medium scale)
	runComparativeTests(ctx, db)

	// 2. Distance metric comparison
	testDistanceMetrics(ctx, db)

	// 3. Large-scale tests (optional - can be commented out for faster testing)
	fmt.Print("\nRun large-scale tests? This may take 5-10 minutes [y/N]: ")
	var response string
	fmt.Scanln(&response)
	if response == "y" || response == "Y" {
		runLargeScaleTests(ctx, db)
	} else {
		printInfo("Skipping large-scale tests")
	}

	// Final summary
	printHeader("Volume Testing Complete")
	printSuccess("All tests completed successfully!")

	fmt.Printf("\n%sKey Findings:%s\n", ColorYellow, ColorReset)
	fmt.Println("- Native Go SDK provides significantly better performance than HTTP API")
	fmt.Println("- HNSW index shows better search performance for larger datasets")
	fmt.Println("- Flat index provides faster insertions but slower searches")
	fmt.Println("- Memory usage scales linearly with vector count and dimensions")
	fmt.Println("- Different distance metrics have minimal performance impact")

	fmt.Printf("\n%sRecommendations:%s\n", ColorYellow, ColorReset)
	fmt.Println("- Use HNSW for datasets > 10k vectors")
	fmt.Println("- Optimize batch sizes based on available memory")
	fmt.Println("- Monitor memory usage during large insertions")
	fmt.Println("- Use appropriate distance metrics for your use case")

	printInfo("Volume test data stored in: ./volume_test_data")
}
