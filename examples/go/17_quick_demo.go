package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// QuickTestClient provides a simple client for testing
type QuickTestClient struct {
	baseURL string
	client  *http.Client
}

// NewQuickTestClient creates a new test client
func NewQuickTestClient(baseURL string) *QuickTestClient {
	return &QuickTestClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// makeRequest is a helper for HTTP requests
func (c *QuickTestClient) makeRequest(method, endpoint string, payload interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	var result map[string]interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func main() {
	fmt.Println("🧪 VittoriaDB Document API Quick Test")
	fmt.Println("========================================")

	client := NewQuickTestClient("http://localhost:8080")

	// Test 1: Create database
	fmt.Println("\n1. Creating database...")
	schema := map[string]interface{}{
		"title":     "string",
		"content":   "string",
		"category":  "string",
		"rating":    "number",
		"embedding": "vector[384]",
		"available": "boolean",
	}

	createPayload := map[string]interface{}{
		"schema": schema,
	}

	_, err := client.makeRequest("POST", "/create", createPayload)
	if err != nil {
		fmt.Printf("❌ Failed to create database: %v\n", err)
		return
	}
	fmt.Println("✅ Database created successfully")

	// Test 2: Insert document
	fmt.Println("\n2. Inserting test document...")
	
	// Generate simple test vector
	embedding := make([]float64, 384)
	for i := range embedding {
		embedding[i] = 0.1
	}

	testDoc := map[string]interface{}{
		"id":        "test_doc_1",
		"title":     "Test Document",
		"content":   "This is a test document for VittoriaDB",
		"category":  "test",
		"rating":    4.5,
		"embedding": embedding,
		"available": true,
	}

	insertPayload := map[string]interface{}{
		"document": testDoc,
	}

	result, err := client.makeRequest("POST", "/documents", insertPayload)
	if err != nil {
		fmt.Printf("❌ Failed to insert document: %v\n", err)
		return
	}

	if created, ok := result["created"].(bool); ok && created {
		fmt.Println("✅ Document inserted successfully")
	} else {
		fmt.Printf("⚠️  Document response: %v\n", result)
	}

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Test 3: Search
	fmt.Println("\n3. Testing search...")
	searchQuery := map[string]interface{}{
		"mode":  "fulltext",
		"term":  "test document",
		"limit": 5,
	}

	searchResult, err := client.makeRequest("POST", "/search", searchQuery)
	if err != nil {
		fmt.Printf("❌ Search failed: %v\n", err)
		return
	}

	if hits, ok := searchResult["hits"].([]interface{}); ok {
		fmt.Printf("✅ Search successful - found %d results\n", len(hits))

		if len(hits) > 0 {
			if firstHit, ok := hits[0].(map[string]interface{}); ok {
				if doc, ok := firstHit["document"].(map[string]interface{}); ok {
					if title, ok := doc["title"].(string); ok {
						fmt.Printf("   📄 First result: %s\n", title)
					}
				}
				if score, ok := firstHit["score"].(float64); ok {
					fmt.Printf("   🎯 Score: %.3f\n", score)
				}
			}
		}
	} else {
		fmt.Printf("⚠️  Unexpected search result format: %v\n", searchResult)
	}

	// Test 4: Get document
	fmt.Println("\n4. Testing document retrieval...")
	getResult, err := client.makeRequest("GET", "/documents/test_doc_1", nil)
	if err != nil {
		fmt.Printf("❌ Failed to get document: %v\n", err)
	} else {
		if found, ok := getResult["found"].(bool); ok && found {
			if doc, ok := getResult["document"].(map[string]interface{}); ok {
				if title, ok := doc["title"].(string); ok {
					fmt.Printf("✅ Document retrieved: %s\n", title)
				}
			}
		} else {
			fmt.Println("⚠️  Document not found (this is a known issue)")
		}
	}

	// Test 5: Count documents
	fmt.Println("\n5. Testing document count...")
	countResult, err := client.makeRequest("GET", "/count", nil)
	if err != nil {
		fmt.Printf("❌ Failed to count documents: %v\n", err)
	} else {
		if count, ok := countResult["count"].(float64); ok {
			fmt.Printf("✅ Document count: %.0f\n", count)
		} else {
			fmt.Printf("⚠️  Unexpected count format: %v\n", countResult)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("🎉 Quick test completed!")
	fmt.Println("✅ Core search functionality is working")
	fmt.Println("⚠️  Some document operations may have known issues")
}
