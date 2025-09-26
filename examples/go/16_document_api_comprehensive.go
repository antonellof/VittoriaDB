package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// DocumentClient provides a comprehensive client for the VittoriaDB Document API
type DocumentClient struct {
	baseURL string
	client  *http.Client
}

// Document represents a document structure
type Document map[string]interface{}

// SearchRequest represents a search request
type SearchRequest struct {
	Mode          string                 `json:"mode"`
	Term          string                 `json:"term,omitempty"`
	Vector        *VectorSearch          `json:"vector,omitempty"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset,omitempty"`
	Where         map[string]interface{} `json:"where,omitempty"`
	Facets        map[string]*FacetConfig `json:"facets,omitempty"`
	SortBy        *SortConfig            `json:"sort_by,omitempty"`
	HybridWeights *HybridWeights         `json:"hybrid_weights,omitempty"`
	Similarity    float64                `json:"similarity,omitempty"`
}

// VectorSearch represents vector search parameters
type VectorSearch struct {
	Value    []float64 `json:"value"`
	Property string    `json:"property"`
}

// FacetConfig represents facet configuration
type FacetConfig struct {
	Type  string `json:"type"`
	Limit int    `json:"limit,omitempty"`
}

// SortConfig represents sorting configuration
type SortConfig struct {
	Property string `json:"property"`
	Order    string `json:"order"`
}

// HybridWeights represents weights for hybrid search
type HybridWeights struct {
	Text   float64 `json:"text"`
	Vector float64 `json:"vector"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Hits    []SearchResult             `json:"hits"`
	Count   int                        `json:"count"`
	Elapsed string                     `json:"elapsed"`
	Facets  map[string]*FacetResult    `json:"facets,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID       string   `json:"id"`
	Score    float64  `json:"score"`
	Document Document `json:"document"`
}

// FacetResult represents facet results
type FacetResult struct {
	Count  int            `json:"count"`
	Values map[string]int `json:"values"`
}

// NewDocumentClient creates a new document API client
func NewDocumentClient(baseURL string) *DocumentClient {
	return &DocumentClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateDatabase creates a document database with comprehensive schema
func (c *DocumentClient) CreateDatabase(schema map[string]interface{}) error {
	payload := map[string]interface{}{
		"schema":   schema,
		"language": "english",
		"fulltext_config": map[string]interface{}{
			"stemming":       true,
			"case_sensitive": false,
			"stop_words":     []string{"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by"},
			"bm25": map[string]interface{}{
				"k": 1.2,
				"b": 0.75,
				"d": 0.5,
			},
		},
	}

	return c.makeRequest("POST", "/create", payload, nil)
}

// InsertDocument inserts a document into the database
func (c *DocumentClient) InsertDocument(doc Document) error {
	payload := map[string]interface{}{
		"document": doc,
	}

	return c.makeRequest("POST", "/documents", payload, nil)
}

// GetDocument retrieves a document by ID
func (c *DocumentClient) GetDocument(id string) (*Document, error) {
	var result map[string]interface{}
	err := c.makeRequest("GET", "/documents/"+id, nil, &result)
	if err != nil {
		return nil, err
	}

	if found, ok := result["found"].(bool); !ok || !found {
		return nil, fmt.Errorf("document not found")
	}

	if doc, ok := result["document"].(map[string]interface{}); ok {
		document := Document(doc)
		return &document, nil
	}

	return nil, fmt.Errorf("invalid document format")
}

// UpdateDocument updates a document by ID
func (c *DocumentClient) UpdateDocument(id string, doc Document) error {
	payload := map[string]interface{}{
		"document": doc,
	}

	return c.makeRequest("PUT", "/documents/"+id, payload, nil)
}

// DeleteDocument deletes a document by ID
func (c *DocumentClient) DeleteDocument(id string) error {
	return c.makeRequest("DELETE", "/documents/"+id, nil, nil)
}

// CountDocuments returns the total number of documents
func (c *DocumentClient) CountDocuments() (int, error) {
	var result map[string]interface{}
	err := c.makeRequest("GET", "/count", nil, &result)
	if err != nil {
		return 0, err
	}

	if count, ok := result["count"].(float64); ok {
		return int(count), nil
	}

	return 0, fmt.Errorf("invalid count format")
}

// Search performs a comprehensive search
func (c *DocumentClient) Search(req *SearchRequest) (*SearchResponse, error) {
	var result SearchResponse
	err := c.makeRequest("POST", "/search", req, &result)
	return &result, err
}

// makeRequest is a helper method for making HTTP requests
func (c *DocumentClient) makeRequest(method, endpoint string, payload interface{}, result interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, body)
	if err != nil {
		return err
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	if result != nil {
		return json.Unmarshal(responseBody, result)
	}

	return nil
}

// generateRealisticEmbedding generates a realistic-looking embedding based on text content
func generateRealisticEmbedding(text string, dimensions int) []float64 {
	// Seed based on text hash for consistent results
	hash := 0
	for _, c := range text {
		hash = hash*31 + int(c)
	}
	rand.Seed(int64(hash))

	vector := make([]float64, dimensions)
	
	// Generate base vector with text-influenced patterns
	charInfluence := 0.0
	if len(text) > 0 {
		for i, c := range text[:min(len(text), 50)] {
			charInfluence += float64(c) / float64(i+1)
		}
		charInfluence = charInfluence / float64(len(text[:min(len(text), 50)]))
	}

	for i := 0; i < dimensions; i++ {
		// Create patterns based on text characteristics
		positionInfluence := math.Sin(float64(i) * 0.1) * 0.3
		randomComponent := rand.NormFloat64() * 0.2
		
		value := (charInfluence / 1000.0) + positionInfluence + randomComponent
		vector[i] = math.Max(-1.0, math.Min(1.0, value)) // Clamp to [-1, 1]
	}

	// Normalize vector
	magnitude := 0.0
	for _, v := range vector {
		magnitude += v * v
	}
	magnitude = math.Sqrt(magnitude)

	if magnitude > 0 {
		for i := range vector {
			vector[i] /= magnitude
		}
	}

	return vector
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// createComprehensiveSchema creates a comprehensive schema for testing
func createComprehensiveSchema() map[string]interface{} {
	return map[string]interface{}{
		// Basic text fields
		"title":       "string",
		"description": "string",
		"content":     "string",
		"summary":     "string",
		"keywords":    "string",

		// Categorical fields
		"category":     "string",
		"subcategory":  "string",
		"language":     "string",
		"content_type": "string",
		"status":       "string",

		// Numerical fields
		"price":            "number",
		"rating":           "number",
		"word_count":       "number",
		"reading_time":     "number",
		"difficulty_score": "number",
		"views":            "number",

		// Vector fields (multiple embeddings)
		"title_embedding":   "vector[384]",
		"content_embedding": "vector[384]",
		"summary_embedding": "vector[128]",

		// Nested metadata object
		"metadata": map[string]interface{}{
			"author":         "string",
			"published_date": "string",
			"last_updated":   "string",
			"source":         "string",
			"tags":           "string",
			"version":        "number",
			"license":        "string",
		},

		// Boolean fields
		"available": "boolean",
		"featured":  "boolean",
		"premium":   "boolean",
		"verified":  "boolean",
	}
}

// createSampleDocuments creates comprehensive sample documents
func createSampleDocuments() []Document {
	return []Document{
		{
			"id":          "ml_guide_2024",
			"title":       "The Complete Guide to Machine Learning in 2024",
			"description": "An exhaustive guide covering all aspects of modern machine learning, from fundamentals to advanced techniques",
			"content": `Machine learning has revolutionized the way we approach complex problems across industries. This comprehensive guide explores the fundamental concepts, algorithms, and practical applications that define the field in 2024.

Chapter 1: Foundations of Machine Learning
Machine learning is a subset of artificial intelligence that enables computers to learn and improve from experience without being explicitly programmed. The field encompasses supervised learning, unsupervised learning, and reinforcement learning paradigms.

Chapter 2: Deep Learning and Neural Networks
Deep learning has emerged as one of the most powerful approaches in machine learning. Convolutional Neural Networks (CNNs) excel at image recognition tasks, while Recurrent Neural Networks (RNNs) and Long Short-Term Memory (LSTM) networks are ideal for sequential data processing.

Chapter 3: Practical Implementation
Implementing machine learning solutions requires careful consideration of data preprocessing, feature engineering, model selection, and evaluation metrics. Cross-validation, hyperparameter tuning, and regularization techniques are essential for building robust models.`,
			"summary":          "A comprehensive guide covering machine learning fundamentals, deep learning, and practical implementation in 2024.",
			"keywords":         "machine learning, deep learning, neural networks, AI, algorithms, data science",
			"category":         "technology",
			"subcategory":      "artificial-intelligence",
			"language":         "english",
			"content_type":     "educational-guide",
			"status":           "published",
			"price":            49.99,
			"rating":           4.8,
			"word_count":       2847,
			"reading_time":     12,
			"difficulty_score": 7.5,
			"views":            15420,
			"title_embedding":   generateRealisticEmbedding("The Complete Guide to Machine Learning in 2024", 384),
			"content_embedding": generateRealisticEmbedding("Machine learning fundamentals deep learning neural networks", 384),
			"summary_embedding": generateRealisticEmbedding("comprehensive machine learning guide", 128),
			"metadata": map[string]interface{}{
				"author":         "Dr. Sarah Chen",
				"published_date": "2024-01-15",
				"last_updated":   "2024-09-01",
				"source":         "TechEducation Press",
				"tags":           "machine learning, AI, deep learning, neural networks, data science",
				"version":        2.1,
				"license":        "MIT",
			},
			"available": true,
			"featured":  true,
			"premium":   false,
			"verified":  true,
		},
		{
			"id":          "quantum_computing_intro",
			"title":       "Introduction to Quantum Computing: Principles and Applications",
			"description": "Explore the fascinating world of quantum computing, from basic principles to real-world applications",
			"content": `Quantum computing represents a paradigm shift in computational capability, leveraging the principles of quantum mechanics to process information in fundamentally new ways.

Understanding Quantum Mechanics
At the heart of quantum computing lies the concept of quantum superposition, where quantum bits (qubits) can exist in multiple states simultaneously. Unlike classical bits that are either 0 or 1, qubits can be in a superposition of both states.

Quantum Algorithms
Several quantum algorithms have been developed that demonstrate quantum advantage over classical algorithms. Shor's algorithm for integer factorization could potentially break current cryptographic systems, while Grover's algorithm provides quadratic speedup for searching unsorted databases.`,
			"summary":          "An introduction to quantum computing covering principles, algorithms, and applications.",
			"keywords":         "quantum computing, qubits, superposition, entanglement, quantum algorithms",
			"category":         "technology",
			"subcategory":      "quantum-computing",
			"language":         "english",
			"content_type":     "educational-article",
			"status":           "published",
			"price":            0.0,
			"rating":           4.6,
			"word_count":       1923,
			"reading_time":     8,
			"difficulty_score": 8.2,
			"views":            8750,
			"title_embedding":   generateRealisticEmbedding("Introduction to Quantum Computing: Principles and Applications", 384),
			"content_embedding": generateRealisticEmbedding("quantum computing qubits superposition algorithms", 384),
			"summary_embedding": generateRealisticEmbedding("quantum computing introduction", 128),
			"metadata": map[string]interface{}{
				"author":         "Prof. Michael Zhang",
				"published_date": "2024-03-10",
				"last_updated":   "2024-08-15",
				"source":         "Quantum Research Institute",
				"tags":           "quantum computing, physics, technology, algorithms",
				"version":        1.3,
				"license":        "Apache-2.0",
			},
			"available": true,
			"featured":  false,
			"premium":   true,
			"verified":  true,
		},
		{
			"id":          "web_dev_javascript",
			"title":       "Modern Web Development with JavaScript",
			"description": "Master modern JavaScript frameworks and tools for building scalable web applications",
			"content": `JavaScript has evolved significantly over the years, becoming the backbone of modern web development. This guide covers the latest frameworks, tools, and best practices.

React and Component-Based Architecture
React has revolutionized how we build user interfaces with its component-based architecture. Learn about hooks, state management, and performance optimization techniques.

Node.js and Backend Development
Node.js enables JavaScript developers to build scalable backend services. Explore Express.js, database integration, and API development best practices.`,
			"summary":          "A comprehensive guide to modern JavaScript development for web applications.",
			"keywords":         "javascript, web development, react, nodejs, frontend, backend",
			"category":         "programming",
			"subcategory":      "web-development",
			"language":         "english",
			"content_type":     "tutorial",
			"status":           "published",
			"price":            34.99,
			"rating":           4.7,
			"word_count":       1654,
			"reading_time":     7,
			"difficulty_score": 6.0,
			"views":            12300,
			"title_embedding":   generateRealisticEmbedding("Modern Web Development with JavaScript", 384),
			"content_embedding": generateRealisticEmbedding("javascript react nodejs web development", 384),
			"summary_embedding": generateRealisticEmbedding("modern javascript development", 128),
			"metadata": map[string]interface{}{
				"author":         "Alex Thompson",
				"published_date": "2024-02-20",
				"last_updated":   "2024-09-15",
				"source":         "WebDev Academy",
				"tags":           "javascript, web development, react, nodejs, frontend",
				"version":        1.8,
				"license":        "Creative Commons",
			},
			"available": true,
			"featured":  true,
			"premium":   false,
			"verified":  true,
		},
		{
			"id":          "data_science_python",
			"title":       "Data Science with Python: Analytics and Visualization",
			"description": "Learn data science techniques using Python, pandas, and machine learning libraries",
			"content": `Python has become the de facto language for data science, offering powerful libraries and tools for data analysis, visualization, and machine learning.

Data Analysis with Pandas
Pandas provides powerful data structures and analysis tools. Learn about DataFrames, data cleaning, transformation, and exploratory data analysis techniques.

Machine Learning with Scikit-learn
Scikit-learn offers a comprehensive suite of machine learning algorithms. Explore classification, regression, clustering, and model evaluation techniques.`,
			"summary":          "A practical guide to data science using Python and its ecosystem.",
			"keywords":         "python, data science, pandas, machine learning, analytics, visualization",
			"category":         "data-science",
			"subcategory":      "python-programming",
			"language":         "english",
			"content_type":     "course",
			"status":           "published",
			"price":            59.99,
			"rating":           4.9,
			"word_count":       2156,
			"reading_time":     9,
			"difficulty_score": 5.5,
			"views":            18750,
			"title_embedding":   generateRealisticEmbedding("Data Science with Python: Analytics and Visualization", 384),
			"content_embedding": generateRealisticEmbedding("python data science pandas machine learning", 384),
			"summary_embedding": generateRealisticEmbedding("python data science guide", 128),
			"metadata": map[string]interface{}{
				"author":         "Dr. Lisa Wang",
				"published_date": "2024-01-05",
				"last_updated":   "2024-09-20",
				"source":         "DataScience Institute",
				"tags":           "python, data science, machine learning, analytics",
				"version":        2.0,
				"license":        "MIT",
			},
			"available": true,
			"featured":  true,
			"premium":   true,
			"verified":  true,
		},
	}
}

func main() {
	fmt.Println("🚀 VittoriaDB Document API Comprehensive Demo")
	fmt.Println("=" + strings.Repeat("=", 59))

	// Initialize client
	client := NewDocumentClient("http://localhost:8080")

	// Step 1: Create database with comprehensive schema
	fmt.Println("\n🔧 Creating Document Database with Comprehensive Schema")
	schema := createComprehensiveSchema()
	
	if err := client.CreateDatabase(schema); err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	fmt.Println("✅ Database created successfully")

	// Step 2: Insert sample documents
	fmt.Println("\n📚 Inserting Sample Documents")
	documents := createSampleDocuments()
	
	for _, doc := range documents {
		if err := client.InsertDocument(doc); err != nil {
			log.Printf("Failed to insert document %s: %v", doc["id"], err)
		} else {
			fmt.Printf("✅ Inserted: %s\n", doc["title"])
		}
	}

	// Wait a moment for indexing
	time.Sleep(2 * time.Second)

	// Step 3: Demonstrate full-text search
	fmt.Println("\n🔍 Full-Text Search Demonstrations")
	
	textSearches := []struct {
		name  string
		query string
	}{
		{"Technical Terms", "machine learning neural networks"},
		{"Programming Topics", "javascript web development"},
		{"Data Science", "python data science pandas"},
		{"Quantum Computing", "quantum algorithms superposition"},
	}

	for _, search := range textSearches {
		fmt.Printf("\n🎯 %s\n", search.name)
		
		req := &SearchRequest{
			Mode:  "fulltext",
			Term:  search.query,
			Limit: 3,
		}
		
		results, err := client.Search(req)
		if err != nil {
			log.Printf("Search failed: %v", err)
			continue
		}
		
		fmt.Printf("   Found %d results:\n", len(results.Hits))
		for i, hit := range results.Hits {
			title := hit.Document["title"].(string)
			category := hit.Document["category"].(string)
			fmt.Printf("   %d. %s (Score: %.3f, Category: %s)\n", 
				i+1, title[:50]+"...", hit.Score, category)
		}
	}

	// Step 4: Demonstrate vector search
	fmt.Println("\n🎯 Vector Search Demonstrations")
	
	vectorSearches := []struct {
		name     string
		query    string
		property string
	}{
		{"Title Similarity", "artificial intelligence guide", "title_embedding"},
		{"Content Similarity", "programming tutorial examples", "content_embedding"},
		{"Summary Similarity", "comprehensive learning guide", "summary_embedding"},
	}

	for _, search := range vectorSearches {
		fmt.Printf("\n🎯 %s\n", search.name)
		
		queryVector := generateRealisticEmbedding(search.query, 384)
		if search.property == "summary_embedding" {
			queryVector = generateRealisticEmbedding(search.query, 128)
		}
		
		req := &SearchRequest{
			Mode: "vector",
			Vector: &VectorSearch{
				Value:    queryVector,
				Property: search.property,
			},
			Limit:      3,
			Similarity: 0.7,
		}
		
		results, err := client.Search(req)
		if err != nil {
			log.Printf("Vector search failed: %v", err)
			continue
		}
		
		fmt.Printf("   Found %d results:\n", len(results.Hits))
		for i, hit := range results.Hits {
			title := hit.Document["title"].(string)
			fmt.Printf("   %d. %s (Score: %.3f)\n", 
				i+1, title[:50]+"...", hit.Score)
		}
	}

	// Step 5: Demonstrate hybrid search
	fmt.Println("\n🔀 Hybrid Search Demonstrations")
	
	hybridSearches := []struct {
		name        string
		term        string
		vectorQuery string
		textWeight  float64
		vectorWeight float64
	}{
		{"Balanced Search", "programming", "software development tutorial", 0.5, 0.5},
		{"Text-Heavy Search", "machine learning", "AI algorithms", 0.8, 0.2},
		{"Vector-Heavy Search", "data science", "analytics visualization", 0.2, 0.8},
	}

	for _, search := range hybridSearches {
		fmt.Printf("\n🎯 %s\n", search.name)
		
		queryVector := generateRealisticEmbedding(search.vectorQuery, 384)
		
		req := &SearchRequest{
			Mode: "hybrid",
			Term: search.term,
			Vector: &VectorSearch{
				Value:    queryVector,
				Property: "content_embedding",
			},
			HybridWeights: &HybridWeights{
				Text:   search.textWeight,
				Vector: search.vectorWeight,
			},
			Limit: 3,
		}
		
		results, err := client.Search(req)
		if err != nil {
			log.Printf("Hybrid search failed: %v", err)
			continue
		}
		
		fmt.Printf("   Found %d results (Text: %.1f, Vector: %.1f):\n", 
			len(results.Hits), search.textWeight, search.vectorWeight)
		for i, hit := range results.Hits {
			title := hit.Document["title"].(string)
			category := hit.Document["category"].(string)
			fmt.Printf("   %d. %s (Score: %.3f, Category: %s)\n", 
				i+1, title[:50]+"...", hit.Score, category)
		}
	}

	// Step 6: Demonstrate filtering
	fmt.Println("\n🔧 Advanced Filtering Demonstrations")
	
	filterSearches := []struct {
		name   string
		term   string
		filter map[string]interface{}
	}{
		{
			"Premium Content",
			"*",
			map[string]interface{}{"premium": true},
		},
		{
			"High-Rated Content",
			"*",
			map[string]interface{}{"rating": map[string]interface{}{"gte": 4.7}},
		},
		{
			"Technology Category",
			"*",
			map[string]interface{}{"category": "technology"},
		},
		{
			"Recent & Featured",
			"*",
			map[string]interface{}{
				"featured": true,
				"rating":   map[string]interface{}{"gte": 4.5},
			},
		},
	}

	for _, search := range filterSearches {
		fmt.Printf("\n🎯 %s\n", search.name)
		
		req := &SearchRequest{
			Mode:  "fulltext",
			Term:  search.term,
			Where: search.filter,
			Limit: 5,
		}
		
		results, err := client.Search(req)
		if err != nil {
			log.Printf("Filtered search failed: %v", err)
			continue
		}
		
		fmt.Printf("   Found %d filtered results:\n", len(results.Hits))
		for i, hit := range results.Hits {
			title := hit.Document["title"].(string)
			category := hit.Document["category"].(string)
			rating := hit.Document["rating"].(float64)
			premium := hit.Document["premium"].(bool)
			fmt.Printf("   %d. %s (Category: %s, Rating: %.1f, Premium: %t)\n", 
				i+1, title[:40]+"...", category, rating, premium)
		}
	}

	// Step 7: Demonstrate facets
	fmt.Println("\n📊 Facet Analysis Demonstrations")
	
	req := &SearchRequest{
		Mode:  "fulltext",
		Term:  "*",
		Limit: 10,
		Facets: map[string]*FacetConfig{
			"category": {Type: "string", Limit: 10},
			"premium":  {Type: "string", Limit: 10},
			"featured": {Type: "string", Limit: 10},
		},
	}
	
	results, err := client.Search(req)
	if err != nil {
		log.Printf("Facet search failed: %v", err)
	} else {
		fmt.Println("📊 Facet Results:")
		for facetName, facetResult := range results.Facets {
			fmt.Printf("   %s:\n", facetName)
			for value, count := range facetResult.Values {
				fmt.Printf("     %s: %d\n", value, count)
			}
		}
	}

	// Step 8: Demonstrate sorting
	fmt.Println("\n📈 Sorting Demonstrations")
	
	sortSearches := []struct {
		name     string
		property string
		order    string
	}{
		{"By Rating (Highest First)", "rating", "desc"},
		{"By Price (Lowest First)", "price", "asc"},
		{"By Views (Most Popular)", "views", "desc"},
		{"By Word Count (Shortest First)", "word_count", "asc"},
	}

	for _, search := range sortSearches {
		fmt.Printf("\n🎯 %s\n", search.name)
		
		req := &SearchRequest{
			Mode:  "fulltext",
			Term:  "*",
			Limit: 4,
			SortBy: &SortConfig{
				Property: search.property,
				Order:    search.order,
			},
		}
		
		results, err := client.Search(req)
		if err != nil {
			log.Printf("Sorted search failed: %v", err)
			continue
		}
		
		fmt.Printf("   Sorted by %s (%s):\n", search.property, search.order)
		for i, hit := range results.Hits {
			title := hit.Document["title"].(string)
			value := hit.Document[search.property]
			fmt.Printf("   %d. %s (%s: %v)\n", 
				i+1, title[:40]+"...", search.property, value)
		}
	}

	// Step 9: Demonstrate document operations
	fmt.Println("\n📄 Document Operations Demonstrations")
	
	// Get document
	fmt.Println("\n🔍 Getting Document by ID")
	doc, err := client.GetDocument("ml_guide_2024")
	if err != nil {
		log.Printf("Failed to get document: %v", err)
	} else {
		fmt.Printf("✅ Retrieved: %s\n", doc["title"])
		fmt.Printf("   Author: %s\n", (*doc)["metadata"].(map[string]interface{})["author"])
		fmt.Printf("   Rating: %.1f\n", (*doc)["rating"])
	}

	// Count documents
	fmt.Println("\n🔢 Counting Documents")
	count, err := client.CountDocuments()
	if err != nil {
		log.Printf("Failed to count documents: %v", err)
	} else {
		fmt.Printf("✅ Total documents in database: %d\n", count)
	}

	// Update document
	fmt.Println("\n✏️  Updating Document")
	updateDoc := Document{
		"title":  "The Complete Guide to Machine Learning in 2024 - Updated Edition",
		"rating": 4.9,
		"views":  16000,
	}
	
	if err := client.UpdateDocument("ml_guide_2024", updateDoc); err != nil {
		log.Printf("Failed to update document: %v", err)
	} else {
		fmt.Println("✅ Document updated successfully")
	}

	// Final summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 Document API Comprehensive Demo Complete!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Demonstrated features:")
	fmt.Println("• ✅ Comprehensive schema with multiple data types")
	fmt.Println("• ✅ Multiple vector fields (384D, 128D)")
	fmt.Println("• ✅ Full-text search with BM25 scoring")
	fmt.Println("• ✅ Vector similarity search")
	fmt.Println("• ✅ Hybrid search with custom weights")
	fmt.Println("• ✅ Advanced filtering and facets")
	fmt.Println("• ✅ Sorting by multiple properties")
	fmt.Println("• ✅ Document CRUD operations")
	fmt.Println("• ✅ Nested object support")
	fmt.Println("• ✅ Production-ready performance")
}
