package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// RAG (Retrieval-Augmented Generation) Example in Go
// This example demonstrates how to build a RAG system using VittoriaDB
// It includes document ingestion, embedding generation, and query processing

type RAGClient struct {
	VittoriaClient *VittoriaClient
	CollectionName string
	Dimensions     int
}

type Document struct {
	ID       string
	Title    string
	Content  string
	Category string
	Author   string
	Year     int
}

type DocumentChunk struct {
	ID         string
	DocumentID string
	Content    string
	ChunkIndex int
	Vector     []float32
	Metadata   map[string]interface{}
}

type RAGQuery struct {
	Question string
	K        int
	Filter   map[string]interface{}
}

type RAGResponse struct {
	Answer   string
	Sources  []SearchResult
	Context  []string
	Confidence float32
}

func NewRAGClient(baseURL, collectionName string, dimensions int) *RAGClient {
	return &RAGClient{
		VittoriaClient: NewVittoriaClient(baseURL),
		CollectionName: collectionName,
		Dimensions:     dimensions,
	}
}

func (r *RAGClient) Initialize() error {
	// Clean up existing collection
	r.VittoriaClient.DeleteCollection(r.CollectionName)

	// Create new collection for RAG
	collection := Collection{
		Name:        r.CollectionName,
		Dimensions:  r.Dimensions,
		IndexType:   "hnsw", // HNSW is better for larger datasets
		Metric:      "cosine",
		Description: "RAG knowledge base with document chunks",
	}

	return r.VittoriaClient.CreateCollection(collection)
}

func (r *RAGClient) IngestDocuments(documents []Document) error {
	fmt.Printf("📚 Ingesting %d documents...\n", len(documents))

	var allChunks []Vector
	chunkID := 0

	for _, doc := range documents {
		chunks := r.chunkDocument(doc)
		fmt.Printf("   📄 Processing '%s': %d chunks\n", doc.Title, len(chunks))

		for _, chunk := range chunks {
			// Generate embedding (using simple random vectors for demo)
			// In a real implementation, you would use a proper embedding model
			chunk.Vector = r.generateEmbedding(chunk.Content)
			
			vector := Vector{
				ID:     fmt.Sprintf("chunk_%d", chunkID),
				Vector: chunk.Vector,
				Metadata: map[string]interface{}{
					"document_id":  chunk.DocumentID,
					"chunk_index":  chunk.ChunkIndex,
					"content":      chunk.Content,
					"title":        doc.Title,
					"category":     doc.Category,
					"author":       doc.Author,
					"year":         doc.Year,
					"chunk_length": len(chunk.Content),
				},
			}

			allChunks = append(allChunks, vector)
			chunkID++
		}
	}

	// Batch insert all chunks
	fmt.Printf("   💾 Storing %d chunks in vector database...\n", len(allChunks))
	start := time.Now()
	
	// Insert in batches of 50 to avoid large request sizes
	batchSize := 50
	for i := 0; i < len(allChunks); i += batchSize {
		end := i + batchSize
		if end > len(allChunks) {
			end = len(allChunks)
		}
		
		batch := allChunks[i:end]
		if err := r.VittoriaClient.AddVectorsBatch(r.CollectionName, batch); err != nil {
			return fmt.Errorf("failed to add batch %d-%d: %w", i, end-1, err)
		}
	}

	ingestionTime := time.Since(start)
	fmt.Printf("   ✅ Ingestion completed in %v (%.1f chunks/sec)\n", 
		ingestionTime, float64(len(allChunks))/ingestionTime.Seconds())

	return nil
}

func (r *RAGClient) chunkDocument(doc Document) []DocumentChunk {
	// Simple chunking strategy: split by sentences and group into chunks
	sentences := strings.Split(doc.Content, ". ")
	var chunks []DocumentChunk
	
	chunkSize := 200 // Target chunk size in characters
	currentChunk := ""
	chunkIndex := 0

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		// Add period back if it was removed by split
		if !strings.HasSuffix(sentence, ".") && !strings.HasSuffix(sentence, "!") && !strings.HasSuffix(sentence, "?") {
			sentence += "."
		}

		// Check if adding this sentence would exceed chunk size
		if len(currentChunk)+len(sentence)+1 > chunkSize && currentChunk != "" {
			// Create chunk from current content
			chunks = append(chunks, DocumentChunk{
				DocumentID: doc.ID,
				Content:    strings.TrimSpace(currentChunk),
				ChunkIndex: chunkIndex,
			})
			chunkIndex++
			currentChunk = sentence
		} else {
			if currentChunk != "" {
				currentChunk += " " + sentence
			} else {
				currentChunk = sentence
			}
		}
	}

	// Add final chunk if there's remaining content
	if currentChunk != "" {
		chunks = append(chunks, DocumentChunk{
			DocumentID: doc.ID,
			Content:    strings.TrimSpace(currentChunk),
			ChunkIndex: chunkIndex,
		})
	}

	return chunks
}

func (r *RAGClient) generateEmbedding(text string) []float32 {
	// Simple embedding generation based on text characteristics
	// In a real implementation, you would use a proper embedding model like:
	// - Sentence Transformers
	// - OpenAI embeddings
	// - Local transformer models
	
	vector := make([]float32, r.Dimensions)
	
	// Use text characteristics to generate somewhat meaningful embeddings
	textLower := strings.ToLower(text)
	words := strings.Fields(textLower)
	
	// Seed based on text hash for consistency
	hash := 0
	for _, char := range textLower {
		hash = hash*31 + int(char)
	}
	rand.Seed(int64(hash))
	
	// Generate base random vector
	for i := 0; i < r.Dimensions; i++ {
		vector[i] = rand.Float32()*2 - 1
	}
	
	// Modify vector based on text features
	for i, word := range words {
		if i >= r.Dimensions {
			break
		}
		
		// Add word-specific modifications
		wordHash := 0
		for _, char := range word {
			wordHash = wordHash*7 + int(char)
		}
		
		vector[i%r.Dimensions] += float32(wordHash%100) / 1000.0
	}
	
	// Add category-specific patterns
	categoryWords := map[string]float32{
		"technology": 0.5,
		"science":    0.3,
		"education":  0.2,
		"database":   0.7,
		"vector":     0.8,
		"ai":         0.9,
		"machine":    0.6,
		"learning":   0.6,
	}
	
	for word, weight := range categoryWords {
		if strings.Contains(textLower, word) {
			for i := 0; i < r.Dimensions/4; i++ {
				vector[i] += weight * 0.1
			}
		}
	}
	
	// Normalize vector
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

func (r *RAGClient) Query(query RAGQuery) (*RAGResponse, error) {
	// Generate embedding for the query
	queryVector := r.generateEmbedding(query.Question)

	// Search for relevant chunks
	searchReq := SearchRequest{
		Vector: queryVector,
		K:      query.K,
		Filter: query.Filter,
	}

	searchResp, err := r.VittoriaClient.Search(r.CollectionName, searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	if len(searchResp.Results) == 0 {
		return &RAGResponse{
			Answer:     "I couldn't find any relevant information to answer your question.",
			Sources:    []SearchResult{},
			Context:    []string{},
			Confidence: 0.0,
		}, nil
	}

	// Extract context from search results
	var contexts []string
	var sources []SearchResult
	totalScore := float32(0)

	for _, result := range searchResp.Results {
		if result.Metadata != nil && result.Metadata["content"] != nil {
			content := result.Metadata["content"].(string)
			contexts = append(contexts, content)
			sources = append(sources, result)
			totalScore += result.Score
		}
	}

	// Generate answer based on context
	answer := r.generateAnswer(query.Question, contexts)
	confidence := totalScore / float32(len(searchResp.Results))

	return &RAGResponse{
		Answer:     answer,
		Sources:    sources,
		Context:    contexts,
		Confidence: confidence,
	}, nil
}

func (r *RAGClient) generateAnswer(question string, contexts []string) string {
	// Simple answer generation based on context
	// In a real implementation, you would use:
	// - OpenAI GPT API
	// - Local LLM (Ollama, etc.)
	// - Other language models

	if len(contexts) == 0 {
		return "I don't have enough information to answer your question."
	}

	// Combine contexts
	combinedContext := strings.Join(contexts, " ")
	
	// Simple keyword-based response generation
	questionLower := strings.ToLower(question)
	contextLower := strings.ToLower(combinedContext)

	var answer strings.Builder
	answer.WriteString("Based on the available information: ")

	// Look for key concepts in the question
	if strings.Contains(questionLower, "what is") || strings.Contains(questionLower, "what are") {
		// Find relevant sentences that might contain definitions
		sentences := strings.Split(combinedContext, ".")
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if len(sentence) > 20 && (strings.Contains(strings.ToLower(sentence), "is") || 
				strings.Contains(strings.ToLower(sentence), "are")) {
				answer.WriteString(sentence)
				answer.WriteString(". ")
				break
			}
		}
	} else if strings.Contains(questionLower, "how") {
		// Look for process or method descriptions
		sentences := strings.Split(combinedContext, ".")
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if len(sentence) > 30 && (strings.Contains(strings.ToLower(sentence), "process") ||
				strings.Contains(strings.ToLower(sentence), "method") ||
				strings.Contains(strings.ToLower(sentence), "way") ||
				strings.Contains(strings.ToLower(sentence), "step")) {
				answer.WriteString(sentence)
				answer.WriteString(". ")
				break
			}
		}
	} else {
		// General response - use first relevant sentence
		sentences := strings.Split(combinedContext, ".")
		if len(sentences) > 0 {
			firstSentence := strings.TrimSpace(sentences[0])
			if len(firstSentence) > 10 {
				answer.WriteString(firstSentence)
				answer.WriteString(". ")
			}
		}
	}

	// If no specific answer was generated, provide a general response
	if answer.Len() <= len("Based on the available information: ") {
		answer.WriteString("The information suggests that ")
		// Take first meaningful chunk
		if len(contexts) > 0 && len(contexts[0]) > 50 {
			answer.WriteString(contexts[0][:50])
			answer.WriteString("...")
		}
	}

	return answer.String()
}

func (r *RAGClient) GetStats() error {
	stats, err := r.VittoriaClient.GetStats()
	if err != nil {
		return err
	}

	collections, err := r.VittoriaClient.ListCollections()
	if err != nil {
		return err
	}

	fmt.Printf("📊 RAG System Statistics:\n")
	fmt.Printf("   - Total collections: %d\n", stats.Collections)
	fmt.Printf("   - Total vectors: %d\n", stats.TotalVectors)
	fmt.Printf("   - Memory usage: %.2f MB\n", float64(stats.MemoryUsage)/1024/1024)

	for _, collection := range collections {
		if collection.Name == r.CollectionName {
			fmt.Printf("   - RAG collection '%s': %d dimensions, %s index\n", 
				collection.Name, collection.Dimensions, collection.IndexType)
		}
	}

	return nil
}

func createSampleDocuments() []Document {
	return []Document{
		{
			ID:       "doc_1",
			Title:    "Introduction to Vector Databases",
			Category: "technology",
			Author:   "Alice Johnson",
			Year:     2024,
			Content: `Vector databases are specialized database systems designed to store, index, and query high-dimensional vector data efficiently. They are essential for modern AI applications, particularly those involving machine learning and natural language processing. Vector databases enable semantic search, recommendation systems, and retrieval-augmented generation (RAG) applications. The key advantage of vector databases is their ability to perform similarity searches using various distance metrics like cosine similarity, Euclidean distance, and dot product. Popular vector databases include Pinecone, Weaviate, Chroma, and VittoriaDB. These systems typically use advanced indexing techniques such as HNSW (Hierarchical Navigable Small World) or IVF (Inverted File) to achieve fast approximate nearest neighbor search even with millions of vectors.`,
		},
		{
			ID:       "doc_2",
			Title:    "Understanding Embeddings in AI",
			Category: "education",
			Author:   "Bob Smith",
			Year:     2024,
			Content: `Embeddings are dense vector representations of data that capture semantic meaning in a continuous vector space. In natural language processing, word embeddings like Word2Vec, GloVe, and modern transformer-based embeddings convert words and sentences into numerical vectors. These vectors encode semantic relationships, allowing similar concepts to be positioned close to each other in the vector space. Sentence embeddings, generated by models like Sentence-BERT or OpenAI's text-embedding-ada-002, can represent entire sentences or documents as vectors. The quality of embeddings directly impacts the performance of downstream tasks like search, classification, and clustering. Modern embedding models are trained on large corpora and can capture complex linguistic patterns, making them invaluable for AI applications.`,
		},
		{
			ID:       "doc_3",
			Title:    "RAG Systems Architecture",
			Category: "technology",
			Author:   "Carol Davis",
			Year:     2024,
			Content: `Retrieval-Augmented Generation (RAG) is an AI architecture that combines information retrieval with text generation to produce more accurate and contextually relevant responses. A typical RAG system consists of several components: a knowledge base stored in a vector database, an embedding model to convert queries and documents into vectors, a retrieval system to find relevant documents, and a language model to generate responses based on retrieved context. The process begins with document ingestion, where text is chunked, embedded, and stored in the vector database. During query time, the user's question is embedded and used to retrieve the most relevant document chunks. These chunks provide context to a language model, which generates a response grounded in the retrieved information. RAG systems are particularly effective for question-answering, chatbots, and knowledge management applications.`,
		},
		{
			ID:       "doc_4",
			Title:    "VittoriaDB Features and Capabilities",
			Category: "technology",
			Author:   "David Wilson",
			Year:     2024,
			Content: `VittoriaDB is a high-performance vector database designed for modern AI applications. It supports multiple indexing algorithms including Flat index for exact search and HNSW for approximate nearest neighbor search. The database offers various distance metrics such as cosine similarity, Euclidean distance, and dot product. VittoriaDB provides both HTTP API and native client libraries for different programming languages. Key features include batch operations for efficient data ingestion, metadata filtering for refined searches, and real-time vector operations. The system is designed for scalability and can handle millions of vectors while maintaining fast query performance. VittoriaDB also includes document processing capabilities, supporting formats like PDF, DOCX, HTML, and plain text. The database provides comprehensive monitoring and statistics to help optimize performance and resource usage.`,
		},
		{
			ID:       "doc_5",
			Title:    "Machine Learning in Production",
			Category: "education",
			Author:   "Eve Brown",
			Year:     2023,
			Content: `Deploying machine learning models in production requires careful consideration of various factors including scalability, reliability, and maintainability. Production ML systems must handle real-time inference, batch processing, and model updates. Vector databases play a crucial role in production ML systems, especially for recommendation engines, search systems, and personalization features. Key challenges include model versioning, A/B testing, monitoring model performance, and handling concept drift. Infrastructure considerations include choosing between cloud and on-premises deployment, setting up proper CI/CD pipelines, and ensuring data privacy and security. Monitoring systems must track both technical metrics (latency, throughput) and business metrics (accuracy, user engagement). Successful production ML systems require collaboration between data scientists, ML engineers, and DevOps teams.`,
		},
	}
}

func main() {
	fmt.Println("🤖 VittoriaDB RAG System Example")
	fmt.Println("=================================")

	// Initialize RAG client
	ragClient := NewRAGClient("http://localhost:8080", "rag_knowledge_base", 256)

	// Test connection
	fmt.Println("\n1. Testing connection...")
	stats, err := ragClient.VittoriaClient.GetStats()
	if err != nil {
		log.Fatalf("❌ Failed to connect to VittoriaDB: %v\nMake sure VittoriaDB is running with: ./vittoriadb run", err)
	}
	fmt.Printf("   ✅ Connected! Database has %d collections\n", stats.Collections)

	// Initialize RAG system
	fmt.Println("\n2. Initializing RAG system...")
	if err := ragClient.Initialize(); err != nil {
		log.Fatalf("❌ Failed to initialize RAG system: %v", err)
	}
	fmt.Printf("   ✅ Created collection '%s' for RAG\n", ragClient.CollectionName)

	// Ingest sample documents
	fmt.Println("\n3. Ingesting knowledge base...")
	documents := createSampleDocuments()
	if err := ragClient.IngestDocuments(documents); err != nil {
		log.Fatalf("❌ Failed to ingest documents: %v", err)
	}

	// Show system stats
	fmt.Println("\n4. System statistics...")
	if err := ragClient.GetStats(); err != nil {
		log.Printf("⚠️  Failed to get stats: %v", err)
	}

	// Demonstrate queries
	fmt.Println("\n5. Demonstrating RAG queries...")
	
	queries := []RAGQuery{
		{
			Question: "What is a vector database?",
			K:        3,
		},
		{
			Question: "How do RAG systems work?",
			K:        3,
		},
		{
			Question: "What are the features of VittoriaDB?",
			K:        2,
			Filter: map[string]interface{}{
				"category": "technology",
			},
		},
		{
			Question: "What are embeddings in AI?",
			K:        2,
		},
	}

	for i, query := range queries {
		fmt.Printf("\n   Query %d: %s\n", i+1, query.Question)
		
		start := time.Now()
		response, err := ragClient.Query(query)
		if err != nil {
			log.Printf("   ❌ Query failed: %v", err)
			continue
		}
		queryTime := time.Since(start)

		fmt.Printf("   ⏱️  Query time: %v\n", queryTime)
		fmt.Printf("   🎯 Confidence: %.2f\n", response.Confidence)
		fmt.Printf("   📝 Answer: %s\n", response.Answer)
		fmt.Printf("   📚 Sources (%d):\n", len(response.Sources))
		
		for j, source := range response.Sources {
			title := "Unknown"
			if source.Metadata != nil && source.Metadata["title"] != nil {
				title = source.Metadata["title"].(string)
			}
			fmt.Printf("      %d. %s (score: %.3f) - %s\n", j+1, source.ID, source.Score, title)
		}
	}

	// Interactive mode
	fmt.Println("\n6. Interactive RAG Demo")
	fmt.Println("   Type your questions (or 'quit' to exit):")
	
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n❓ Your question: ")
		if !scanner.Scan() {
			break
		}
		
		question := strings.TrimSpace(scanner.Text())
		if question == "" {
			continue
		}
		if strings.ToLower(question) == "quit" || strings.ToLower(question) == "exit" {
			break
		}

		query := RAGQuery{
			Question: question,
			K:        3,
		}

		start := time.Now()
		response, err := ragClient.Query(query)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}
		queryTime := time.Since(start)

		fmt.Printf("   ⏱️  Response time: %v\n", queryTime)
		fmt.Printf("   🤖 Answer: %s\n", response.Answer)
		
		if len(response.Sources) > 0 {
			fmt.Printf("   📚 Top source: %s (score: %.3f)\n", 
				response.Sources[0].ID, response.Sources[0].Score)
		}
	}

	// Cleanup
	fmt.Println("\n7. Cleaning up...")
	if err := ragClient.VittoriaClient.DeleteCollection(ragClient.CollectionName); err != nil {
		log.Printf("⚠️  Failed to delete collection: %v", err)
	} else {
		fmt.Printf("   ✅ Deleted collection '%s'\n", ragClient.CollectionName)
	}

	fmt.Println("\n🎉 RAG example completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("- Integrate with real embedding models (Sentence Transformers, OpenAI)")
	fmt.Println("- Connect to actual language models (GPT, Claude, local LLMs)")
	fmt.Println("- Try document processing: go run document_processing.go")
	fmt.Println("- Test with larger datasets: go run volume_test.go")
}
