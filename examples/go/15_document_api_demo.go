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

// DocumentClient provides a client for the VittoriaDB Document API
type DocumentClient struct {
	baseURL string
	client  *http.Client
}

// NewDocumentClient creates a new document API client
func NewDocumentClient(baseURL string) *DocumentClient {
	return &DocumentClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateDatabase creates a document database with schema
func (c *DocumentClient) CreateDatabase(schema map[string]interface{}) error {
	payload := map[string]interface{}{
		"schema":   schema,
		"language": "english",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(c.baseURL+"/create", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create database: %s", string(body))
	}

	return nil
}

// InsertDocument inserts a document
func (c *DocumentClient) InsertDocument(document map[string]interface{}) (string, error) {
	payload := map[string]interface{}{
		"document": document,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Post(c.baseURL+"/documents", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to insert document: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["id"].(string), nil
}

// Search performs a unified search
func (c *DocumentClient) Search(params map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.baseURL+"/search", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetDocument retrieves a document by ID
func (c *DocumentClient) GetDocument(id string) (map[string]interface{}, error) {
	resp, err := c.client.Get(c.baseURL + "/documents/" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if found, ok := result["found"].(bool); !ok || !found {
		return nil, fmt.Errorf("document not found")
	}

	return result["document"].(map[string]interface{}), nil
}

// generateRandomVector creates a random vector for demonstration
func generateRandomVector(dimensions int) []float32 {
	vector := make([]float32, dimensions)
	for i := range vector {
		vector[i] = rand.Float32()
	}
	return vector
}

func main() {
	fmt.Println("🚀 VittoriaDB Unified API Demo (Go)")
	fmt.Println("=" + fmt.Sprintf("%50s", "="))

	// Initialize client
	client := NewDocumentClient("http://localhost:8080")

	// Define schema
	fmt.Println("1. Creating unified database with schema...")
	schema := map[string]interface{}{
		"name":        "string",
		"description": "string",
		"price":       "number",
		"category":    "string",
		"tags":        "string[]",
		"embedding":   "vector[384]",
		"meta": map[string]interface{}{
			"rating":  "number",
			"reviews": "number",
			"brand":   "string",
		},
		"available": "boolean",
	}

	if err := client.CreateDatabase(schema); err != nil {
		log.Printf("   ⚠️ Database creation failed (may already exist): %v", err)
	} else {
		fmt.Println("   ✅ Database created with schema")
	}

	// Insert sample documents
	fmt.Println("\n2. Inserting sample documents...")

	sampleDocs := []map[string]interface{}{
		{
			"name":        "Noise Cancelling Headphones",
			"description": "Premium wireless headphones with active noise cancellation",
			"price":       299.99,
			"category":    "electronics",
			"tags":        []string{"audio", "wireless", "premium", "noise-cancelling"},
			"embedding":   generateRandomVector(384),
			"meta": map[string]interface{}{
				"rating":  4.8,
				"reviews": 1250.0,
				"brand":   "AudioTech",
			},
			"available": true,
		},
		{
			"name":        "Wireless Gaming Mouse",
			"description": "High-precision gaming mouse with RGB lighting",
			"price":       89.99,
			"category":    "electronics",
			"tags":        []string{"gaming", "wireless", "rgb", "precision"},
			"embedding":   generateRandomVector(384),
			"meta": map[string]interface{}{
				"rating":  4.6,
				"reviews": 890.0,
				"brand":   "GameGear",
			},
			"available": true,
		},
		{
			"name":        "Organic Coffee Beans",
			"description": "Single-origin organic coffee beans with rich flavor",
			"price":       24.99,
			"category":    "food",
			"tags":        []string{"organic", "coffee", "single-origin", "premium"},
			"embedding":   generateRandomVector(384),
			"meta": map[string]interface{}{
				"rating":  4.9,
				"reviews": 456.0,
				"brand":   "BrewMaster",
			},
			"available": true,
		},
	}

	var insertedIDs []string
	for _, doc := range sampleDocs {
		id, err := client.InsertDocument(doc)
		if err != nil {
			log.Printf("   ❌ Failed to insert document: %v", err)
			continue
		}
		insertedIDs = append(insertedIDs, id)
		fmt.Printf("   📄 Inserted: %s (ID: %s)\n", doc["name"], id)
	}

	fmt.Printf("\n   ✅ Inserted %d documents\n", len(insertedIDs))

	// Demonstrate full-text search
	fmt.Println("\n3. Full-text search examples...")

	fmt.Println("   🔍 Basic text search for 'wireless':")
	searchParams := map[string]interface{}{
		"term":  "wireless",
		"mode":  "fulltext",
		"limit": 3,
	}

	results, err := client.Search(searchParams)
	if err != nil {
		log.Printf("   ❌ Search failed: %v", err)
	} else {
		hits := results["hits"].([]interface{})
		fmt.Printf("      Found %d results:\n", len(hits))
		for i, hit := range hits {
			hitMap := hit.(map[string]interface{})
			doc := hitMap["document"].(map[string]interface{})
			score := hitMap["score"].(float64)
			fmt.Printf("      %d. %s (score: %.3f)\n", i+1, doc["name"], score)
			fmt.Printf("         Price: $%.2f, Category: %s\n", doc["price"], doc["category"])
		}
	}

	// Advanced text search with filters
	fmt.Println("\n   🎯 Advanced text search with filters:")
	advancedParams := map[string]interface{}{
		"term": "premium",
		"mode": "fulltext",
		"where": map[string]interface{}{
			"category":  "electronics",
			"available": true,
		},
		"boost": map[string]interface{}{
			"name":        2.0,
			"description": 1.0,
		},
		"limit": 3,
	}

	results, err = client.Search(advancedParams)
	if err != nil {
		log.Printf("   ❌ Advanced search failed: %v", err)
	} else {
		hits := results["hits"].([]interface{})
		fmt.Printf("      Found %d electronics results:\n", len(hits))
		for _, hit := range hits {
			hitMap := hit.(map[string]interface{})
			doc := hitMap["document"].(map[string]interface{})
			score := hitMap["score"].(float64)
			fmt.Printf("      • %s - $%.2f (score: %.3f)\n", doc["name"], doc["price"], score)
		}
	}

	// Demonstrate vector search
	fmt.Println("\n4. Vector similarity search...")

	queryVector := generateRandomVector(384)
	vectorParams := map[string]interface{}{
		"mode": "vector",
		"vector": map[string]interface{}{
			"value":    queryVector,
			"property": "embedding",
		},
		"similarity": 0.0, // Lower threshold for demo
		"limit":      3,
	}

	fmt.Println("   🎯 Vector similarity search:")
	results, err = client.Search(vectorParams)
	if err != nil {
		log.Printf("   ❌ Vector search failed: %v", err)
	} else {
		hits := results["hits"].([]interface{})
		fmt.Printf("      Found %d similar items:\n", len(hits))
		for _, hit := range hits {
			hitMap := hit.(map[string]interface{})
			doc := hitMap["document"].(map[string]interface{})
			score := hitMap["score"].(float64)
			meta := doc["meta"].(map[string]interface{})
			fmt.Printf("      • %s (similarity: %.3f)\n", doc["name"], score)
			fmt.Printf("        Category: %s, Rating: %.1f\n", doc["category"], meta["rating"])
		}
	}

	// Demonstrate hybrid search
	fmt.Println("\n5. Hybrid search (text + vector)...")

	hybridParams := map[string]interface{}{
		"term": "high quality audio",
		"mode": "hybrid",
		"vector": map[string]interface{}{
			"value":    queryVector,
			"property": "embedding",
		},
		"hybrid_weights": map[string]interface{}{
			"text":   0.7,
			"vector": 0.3,
		},
		"limit": 3,
	}

	fmt.Println("   🔀 Hybrid search combining text and vector similarity:")
	results, err = client.Search(hybridParams)
	if err != nil {
		log.Printf("   ❌ Hybrid search failed: %v", err)
	} else {
		hits := results["hits"].([]interface{})
		fmt.Printf("      Found %d hybrid results:\n", len(hits))
		for _, hit := range hits {
			hitMap := hit.(map[string]interface{})
			doc := hitMap["document"].(map[string]interface{})
			score := hitMap["score"].(float64)
			description := doc["description"].(string)
			if len(description) > 60 {
				description = description[:60] + "..."
			}
			fmt.Printf("      • %s (combined score: %.3f)\n", doc["name"], score)
			fmt.Printf("        %s\n", description)
		}
	}

	// Demonstrate faceted search
	fmt.Println("\n6. Faceted search and analytics...")

	facetParams := map[string]interface{}{
		"term": "*", // Match all
		"mode": "fulltext",
		"facets": map[string]interface{}{
			"category": map[string]interface{}{
				"type":  "string",
				"limit": 10,
			},
			"meta.brand": map[string]interface{}{
				"type":  "string",
				"limit": 10,
			},
		},
		"limit": 10,
	}

	fmt.Println("   📊 Search with facets for analytics:")
	results, err = client.Search(facetParams)
	if err != nil {
		log.Printf("   ❌ Faceted search failed: %v", err)
	} else {
		count := int(results["count"].(float64))
		fmt.Printf("      Total documents: %d\n", count)

		if facets, ok := results["facets"].(map[string]interface{}); ok {
			if categoryFacet, ok := facets["category"].(map[string]interface{}); ok {
				if values, ok := categoryFacet["values"].(map[string]interface{}); ok {
					fmt.Println("      📈 Category breakdown:")
					for category, count := range values {
						fmt.Printf("         %s: %.0f items\n", category, count.(float64))
					}
				}
			}
		}
	}

	// Demonstrate document retrieval
	fmt.Println("\n7. Document management...")

	if len(insertedIDs) > 0 {
		firstID := insertedIDs[0]
		fmt.Printf("   📖 Retrieving document %s:\n", firstID)

		doc, err := client.GetDocument(firstID)
		if err != nil {
			log.Printf("   ❌ Failed to get document: %v", err)
		} else {
			meta := doc["meta"].(map[string]interface{})
			fmt.Printf("      Found: %s\n", doc["name"])
			fmt.Printf("      Price: $%.2f, Rating: %.1f\n", doc["price"], meta["rating"])
		}
	}

	// Performance demonstration
	fmt.Println("\n8. Performance showcase...")

	fmt.Println("   ⚡ Rapid search performance test:")
	startTime := time.Now()

	searchTypes := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			"Text search",
			map[string]interface{}{"term": "premium", "mode": "fulltext", "limit": 5},
		},
		{
			"Vector search",
			map[string]interface{}{
				"mode": "vector",
				"vector": map[string]interface{}{
					"value":    queryVector,
					"property": "embedding",
				},
				"limit": 5,
			},
		},
		{
			"Hybrid search",
			map[string]interface{}{
				"term": "quality",
				"mode": "hybrid",
				"vector": map[string]interface{}{
					"value":    queryVector,
					"property": "embedding",
				},
				"limit": 5,
			},
		},
	}

	for _, searchType := range searchTypes {
		searchStart := time.Now()
		results, err := client.Search(searchType.params)
		searchTime := time.Since(searchStart)

		if err != nil {
			fmt.Printf("      %s: Error - %v\n", searchType.name, err)
		} else {
			hits := results["hits"].([]interface{})
			fmt.Printf("      %s: %d results in %.1fms\n", searchType.name, len(hits), float64(searchTime.Nanoseconds())/1e6)
		}
	}

	totalTime := time.Since(startTime)
	fmt.Printf("   ⏱️ Total time for 3 searches: %.1fms\n", float64(totalTime.Nanoseconds())/1e6)

	fmt.Println("\n✅ Unified API demo completed successfully!")
	fmt.Println("\n🎯 Key Features Demonstrated:")
	fmt.Println("   • Schema-based document structure")
	fmt.Println("   • Full-text search with advanced filtering")
	fmt.Println("   • Vector similarity search")
	fmt.Println("   • Hybrid search combining text and vectors")
	fmt.Println("   • Faceted search for analytics")
	fmt.Println("   • Document management operations")
	fmt.Println("   • High-performance search capabilities")

	fmt.Println("\n📚 Next Steps:")
	fmt.Println("   • Integrate with real embedding models")
	fmt.Println("   • Build production applications")
	fmt.Println("   • Explore advanced query combinations")
	fmt.Println("   • Scale with larger datasets")
}
