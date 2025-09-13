package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// VittoriaDB Go Client Example
// This example demonstrates basic usage of VittoriaDB through HTTP API
// It shows collection management, vector operations, and search functionality

type VittoriaClient struct {
	BaseURL string
	Client  *http.Client
}

type Collection struct {
	Name        string `json:"name"`
	Dimensions  int    `json:"dimensions"`
	IndexType   int    `json:"index_type"` // 0=flat, 1=hnsw, 2=ivf
	Metric      int    `json:"metric"`     // 0=cosine, 1=euclidean, 2=dot_product, 3=manhattan
	Description string `json:"description,omitempty"`
}

type Vector struct {
	ID       string                 `json:"id"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type SearchRequest struct {
	Vector []float32              `json:"vector"`
	K      int                    `json:"k"`
	Filter map[string]interface{} `json:"filter,omitempty"`
}

type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Time    string         `json:"time"`
}

type CollectionInfo struct {
	Name         string `json:"name"`
	VectorCount  int    `json:"vector_count"`
	Dimensions   int    `json:"dimensions"`
	IndexType    int    `json:"index_type"`
	IndexSize    int    `json:"index_size"`
	LastModified string `json:"last_modified"`
}

type DatabaseStats struct {
	Collections     []CollectionInfo `json:"collections"`
	TotalVectors    int              `json:"total_vectors"`
	TotalSize       int              `json:"total_size"`
	IndexSize       int              `json:"index_size"`
	QueriesTotal    int              `json:"queries_total"`
	QueriesPerSec   float64          `json:"queries_per_sec"`
	AvgQueryLatency float64          `json:"avg_query_latency"`
}

func NewVittoriaClient(baseURL string) *VittoriaClient {
	return &VittoriaClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *VittoriaClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

func (c *VittoriaClient) CreateCollection(collection Collection) error {
	resp, err := c.makeRequest("POST", "/collections", collection)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: %s", string(body))
	}

	return nil
}

type CollectionsResponse struct {
	Collections []Collection `json:"collections"`
	Count       int          `json:"count"`
}

func (c *VittoriaClient) ListCollections() ([]Collection, error) {
	resp, err := c.makeRequest("GET", "/collections", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list collections: status %d", resp.StatusCode)
	}

	var response CollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Collections, nil
}

func (c *VittoriaClient) AddVector(collectionName string, vector Vector) error {
	endpoint := fmt.Sprintf("/collections/%s/vectors", collectionName)
	resp, err := c.makeRequest("POST", endpoint, vector)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add vector: %s", string(body))
	}

	return nil
}

func (c *VittoriaClient) AddVectorsBatch(collectionName string, vectors []Vector) error {
	endpoint := fmt.Sprintf("/collections/%s/vectors/batch", collectionName)
	resp, err := c.makeRequest("POST", endpoint, map[string][]Vector{"vectors": vectors})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add vectors batch: %s", string(body))
	}

	return nil
}

func (c *VittoriaClient) Search(collectionName string, searchReq SearchRequest) (*SearchResponse, error) {
	endpoint := fmt.Sprintf("/collections/%s/search", collectionName)
	resp, err := c.makeRequest("POST", endpoint, searchReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search: %s", string(body))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &searchResp, nil
}

func (c *VittoriaClient) GetStats() (*DatabaseStats, error) {
	resp, err := c.makeRequest("GET", "/stats", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get stats: status %d", resp.StatusCode)
	}

	var stats DatabaseStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	return &stats, nil
}

func (c *VittoriaClient) DeleteCollection(collectionName string) error {
	endpoint := fmt.Sprintf("/collections/%s", collectionName)
	resp, err := c.makeRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete collection: %s", string(body))
	}

	return nil
}

func generateRandomVector(dimensions int) []float32 {
	vector := make([]float32, dimensions)
	for i := 0; i < dimensions; i++ {
		vector[i] = rand.Float32()*2 - 1 // Random values between -1 and 1
	}
	return vector
}

func main() {
	fmt.Println("🚀 VittoriaDB Go Basic Usage Example")
	fmt.Println("====================================")

	// Initialize client
	client := NewVittoriaClient("http://localhost:8080")

	// Test connection
	fmt.Println("\n1. Testing connection...")
	stats, err := client.GetStats()
	if err != nil {
		log.Fatalf("❌ Failed to connect to VittoriaDB: %v\nMake sure VittoriaDB is running with: ./vittoriadb run", err)
	}
	fmt.Printf("   ✅ Connected! Database has %d collections\n", len(stats.Collections))

	// Collection name for this demo
	collectionName := "go_basic_demo"

	// Clean up any existing collection
	fmt.Println("\n2. Setting up collection...")
	client.DeleteCollection(collectionName)

	// Create collection
	collection := Collection{
		Name:        collectionName,
		Dimensions:  128,
		IndexType:   0, // 0=flat, 1=hnsw, 2=ivf
		Metric:      0, // 0=cosine, 1=euclidean, 2=dot_product, 3=manhattan
		Description: "Go basic usage demonstration",
	}

	if err := client.CreateCollection(collection); err != nil {
		log.Fatalf("❌ Failed to create collection: %v", err)
	}
	fmt.Printf("   ✅ Created collection '%s' with %d dimensions\n", collectionName, collection.Dimensions)

	// List collections
	collections, err := client.ListCollections()
	if err != nil {
		log.Fatalf("❌ Failed to list collections: %v", err)
	}
	fmt.Printf("   ✅ Database now has %d collections\n", len(collections))

	// Add individual vectors
	fmt.Println("\n3. Adding individual vectors...")
	rand.Seed(42) // For reproducible results

	sampleVectors := []Vector{
		{
			ID:     "doc_1",
			Vector: generateRandomVector(128),
			Metadata: map[string]interface{}{
				"title":    "Introduction to Vector Databases",
				"category": "technology",
				"author":   "Alice Smith",
				"year":     2024,
			},
		},
		{
			ID:     "doc_2",
			Vector: generateRandomVector(128),
			Metadata: map[string]interface{}{
				"title":    "Machine Learning Fundamentals",
				"category": "education",
				"author":   "Bob Johnson",
				"year":     2023,
			},
		},
		{
			ID:     "doc_3",
			Vector: generateRandomVector(128),
			Metadata: map[string]interface{}{
				"title":    "Advanced AI Techniques",
				"category": "technology",
				"author":   "Carol Davis",
				"year":     2024,
			},
		},
	}

	for _, vector := range sampleVectors {
		if err := client.AddVector(collectionName, vector); err != nil {
			log.Fatalf("❌ Failed to add vector %s: %v", vector.ID, err)
		}
		fmt.Printf("   ✅ Added vector '%s': %s\n", vector.ID, vector.Metadata["title"])
	}

	// Add batch vectors
	fmt.Println("\n4. Adding batch vectors...")
	batchVectors := make([]Vector, 10)
	for i := 0; i < 10; i++ {
		batchVectors[i] = Vector{
			ID:     fmt.Sprintf("batch_%d", i),
			Vector: generateRandomVector(128),
			Metadata: map[string]interface{}{
				"title":     fmt.Sprintf("Batch Document %d", i),
				"category":  []string{"technology", "education", "science"}[i%3],
				"batch_id":  "batch_001",
				"timestamp": time.Now().Unix(),
			},
		}
	}

	start := time.Now()
	if err := client.AddVectorsBatch(collectionName, batchVectors); err != nil {
		log.Fatalf("❌ Failed to add batch vectors: %v", err)
	}
	batchTime := time.Since(start)
	fmt.Printf("   ✅ Added %d vectors in batch in %v\n", len(batchVectors), batchTime)

	// Perform searches
	fmt.Println("\n5. Performing similarity searches...")

	// Basic search
	queryVector := generateRandomVector(128)
	searchReq := SearchRequest{
		Vector: queryVector,
		K:      5,
	}

	searchResp, err := client.Search(collectionName, searchReq)
	if err != nil {
		log.Fatalf("❌ Failed to search: %v", err)
	}

	fmt.Printf("   ✅ Found %d similar vectors (search took %s):\n", len(searchResp.Results), searchResp.Time)
	for i, result := range searchResp.Results {
		title := "Unknown"
		if result.Metadata != nil && result.Metadata["title"] != nil {
			title = result.Metadata["title"].(string)
		}
		fmt.Printf("      %d. %s (score: %.4f) - %s\n", i+1, result.ID, result.Score, title)
	}

	// Filtered search
	fmt.Println("\n6. Performing filtered search...")
	filteredSearchReq := SearchRequest{
		Vector: queryVector,
		K:      3,
		Filter: map[string]interface{}{
			"category": "technology",
		},
	}

	filteredResp, err := client.Search(collectionName, filteredSearchReq)
	if err != nil {
		log.Fatalf("❌ Failed to perform filtered search: %v", err)
	}

	fmt.Printf("   ✅ Found %d technology documents:\n", len(filteredResp.Results))
	for i, result := range filteredResp.Results {
		title := "Unknown"
		if result.Metadata != nil && result.Metadata["title"] != nil {
			title = result.Metadata["title"].(string)
		}
		fmt.Printf("      %d. %s (score: %.4f) - %s\n", i+1, result.ID, result.Score, title)
	}

	// Get final stats
	fmt.Println("\n7. Final database statistics...")
	finalStats, err := client.GetStats()
	if err != nil {
		log.Fatalf("❌ Failed to get final stats: %v", err)
	}

	fmt.Printf("   ✅ Database statistics:\n")
	fmt.Printf("      - Collections: %d\n", len(finalStats.Collections))
	fmt.Printf("      - Total vectors: %d\n", finalStats.TotalVectors)
	fmt.Printf("      - Memory usage: %.2f MB\n", float64(finalStats.TotalSize)/1024/1024)

	// Performance demonstration
	fmt.Println("\n8. Performance demonstration...")
	performanceTest(client, collectionName)

	// Cleanup
	fmt.Println("\n9. Cleaning up...")
	if err := client.DeleteCollection(collectionName); err != nil {
		log.Printf("⚠️  Failed to delete collection: %v", err)
	} else {
		fmt.Printf("   ✅ Deleted collection '%s'\n", collectionName)
	}

	fmt.Println("\n🎉 Basic usage example completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("- Try the RAG example: go run rag_example.go")
	fmt.Println("- Test document processing: go run document_processing.go")
	fmt.Println("- Run performance benchmarks: go run performance_test.go")
}

func performanceTest(client *VittoriaClient, collectionName string) {
	fmt.Println("   Running performance tests...")

	// Test individual vs batch insertion
	testCollection := "perf_test"
	client.DeleteCollection(testCollection)

	collection := Collection{
		Name:       testCollection,
		Dimensions: 64,
		IndexType:  0, // 0=flat
		Metric:     0, // 0=cosine
	}

	if err := client.CreateCollection(collection); err != nil {
		log.Printf("   ⚠️  Failed to create performance test collection: %v", err)
		return
	}

	// Individual insertions
	start := time.Now()
	for i := 0; i < 100; i++ {
		vector := Vector{
			ID:     fmt.Sprintf("individual_%d", i),
			Vector: generateRandomVector(64),
		}
		client.AddVector(testCollection, vector)
	}
	individualTime := time.Since(start)

	// Batch insertion
	batchVectors := make([]Vector, 100)
	for i := 0; i < 100; i++ {
		batchVectors[i] = Vector{
			ID:     fmt.Sprintf("batch_perf_%d", i),
			Vector: generateRandomVector(64),
		}
	}

	start = time.Now()
	client.AddVectorsBatch(testCollection, batchVectors)
	batchTime := time.Since(start)

	fmt.Printf("   ✅ Performance comparison (100 vectors):\n")
	fmt.Printf("      - Individual insertions: %v (%.2f vectors/sec)\n",
		individualTime, 100.0/individualTime.Seconds())
	fmt.Printf("      - Batch insertion: %v (%.2f vectors/sec)\n",
		batchTime, 100.0/batchTime.Seconds())
	fmt.Printf("      - Batch is %.1fx faster\n", individualTime.Seconds()/batchTime.Seconds())

	// Search performance
	queryVector := generateRandomVector(64)
	searchReq := SearchRequest{
		Vector: queryVector,
		K:      10,
	}

	// Warm up
	client.Search(testCollection, searchReq)

	// Measure search time
	searchTimes := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		start = time.Now()
		client.Search(testCollection, searchReq)
		searchTimes[i] = time.Since(start)
	}

	var totalTime time.Duration
	for _, t := range searchTimes {
		totalTime += t
	}
	avgSearchTime := totalTime / time.Duration(len(searchTimes))

	fmt.Printf("   ✅ Search performance (200 vectors, k=10):\n")
	fmt.Printf("      - Average search time: %v\n", avgSearchTime)
	fmt.Printf("      - Searches per second: %.1f\n", 1.0/avgSearchTime.Seconds())

	// Cleanup
	client.DeleteCollection(testCollection)
}
