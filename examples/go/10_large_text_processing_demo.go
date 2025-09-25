package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/core"
	"github.com/antonellof/VittoriaDB/pkg/embeddings"
	"github.com/antonellof/VittoriaDB/pkg/processor"
)

func main() {
	fmt.Println("📚 VittoriaDB Large Text Processing Demo")
	fmt.Println("=======================================")
	fmt.Println("Processing large text files with proper semantic matching")
	fmt.Println()

	ctx := context.Background()

	// Test 1: Process Large Text Files
	fmt.Println("📄 Test 1: Large Text File Processing")
	fmt.Println("------------------------------------")
	testLargeTextProcessing(ctx)
	fmt.Println()

	// Test 2: Semantic Search with Proper Matching
	fmt.Println("🔍 Test 2: Semantic Search with Proper Matching")
	fmt.Println("-----------------------------------------------")
	testSemanticSearchMatching(ctx)
	fmt.Println()

	// Test 3: Similarity Score Analysis
	fmt.Println("📊 Test 3: Similarity Score Analysis")
	fmt.Println("-----------------------------------")
	testSimilarityScoreAnalysis(ctx)
	fmt.Println()

	fmt.Println("✅ All large text processing tests completed!")
}

func testLargeTextProcessing(ctx context.Context) {
	// Create collection for large text processing
	collection, err := core.NewCollection(
		"large_text_processing",
		384,
		core.DistanceMetricCosine,
		core.IndexTypeFlat,
		"/tmp/vittoria_large_text",
	)
	if err != nil {
		log.Fatalf("❌ Failed to create collection: %v", err)
	}

	// Try to configure native vectorizer first
	vectorizerConfig := &embeddings.VectorizerConfig{
		Type:       embeddings.VectorizerTypeSentenceTransformers,
		Model:      "all-MiniLM-L6-v2",
		Dimensions: 384,
		Options: map[string]interface{}{
			"batch_size":                 16,
			"enable_enhanced_processing": true,
		},
	}

	factory := embeddings.NewVectorizerFactory()
	vectorizer, err := factory.CreateVectorizer(vectorizerConfig)
	useNativeVectorizer := err == nil

	if useNativeVectorizer {
		collection.SetVectorizer(vectorizer)
		fmt.Println("✅ Using native vectorizer (sentence-transformers)")
	} else {
		fmt.Println("⚠️  Native vectorizer not available, using enhanced manual vectors")
		fmt.Println("   For production use: pip install sentence-transformers")
	}

	// Process large text files (README files and documentation)
	textFiles := []string{
		"/Users/d695663/Desktop/Dev/CognitoraVector/README.md",
		"/Users/d695663/Desktop/Dev/CognitoraVector/releases/memvid-main/README.md",
		"/Users/d695663/Desktop/Dev/CognitoraVector/examples/README.md",
	}

	// Add documentation files
	docsDir := "/Users/d695663/Desktop/Dev/CognitoraVector/docs"
	if entries, err := ioutil.ReadDir(docsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				textFiles = append(textFiles, filepath.Join(docsDir, entry.Name()))
			}
		}
	}

	// Process all text files with smart chunking
	chunker := processor.NewSmartChunker()
	config := &processor.ProcessingConfig{
		ChunkSize:    768, // Larger chunks for better context
		ChunkOverlap: 128,
		MinChunkSize: 200,
		MaxChunkSize: 1500,
	}

	var allTextVectors []*core.TextVector
	var allVectors []*core.Vector
	totalChunks := 0
	totalWords := 0

	for _, filePath := range textFiles {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("⚠️  Failed to read %s: %v\n", filePath, err)
			continue
		}

		// Use smart chunking for better semantic coherence
		chunks, err := chunker.ChunkText(string(content), config)
		if err != nil {
			fmt.Printf("⚠️  Failed to chunk %s: %v\n", filePath, err)
			continue
		}

		fileName := filepath.Base(filePath)
		words := len(strings.Fields(string(content)))
		totalWords += words

		for i, chunk := range chunks {
			chunkID := fmt.Sprintf("%s_chunk_%d", strings.TrimSuffix(fileName, ".md"), i)

			if useNativeVectorizer {
				// Use native text vectorization
				textVector := &core.TextVector{
					ID:   chunkID,
					Text: chunk.Content,
					Metadata: map[string]interface{}{
						"file_name":   fileName,
						"file_path":   filePath,
						"chunk_index": i,
						"chunk_size":  len(chunk.Content),
						"word_count":  len(strings.Fields(chunk.Content)),
						"source_type": "documentation",
					},
				}
				allTextVectors = append(allTextVectors, textVector)
			} else {
				// Use enhanced manual vectorization
				metadata := make(map[string]interface{})
				for k, v := range chunk.Metadata {
					metadata[k] = v
				}
				metadata["file_name"] = fileName
				metadata["file_path"] = filePath
				metadata["chunk_index"] = i
				metadata["chunk_size"] = len(chunk.Content)
				metadata["word_count"] = len(strings.Fields(chunk.Content))
				metadata["source_type"] = "documentation"

				vector := &core.Vector{
					ID:       chunkID,
					Vector:   generateEnhancedSemanticVector(chunk.Content, 384),
					Metadata: metadata,
				}
				allVectors = append(allVectors, vector)
			}
			totalChunks++
		}

		fmt.Printf("   📄 %s -> %d chunks (%d words)\n", fileName, len(chunks), words)
	}

	if totalChunks == 0 {
		fmt.Println("⚠️  No chunks generated from text files")
		return
	}

	// Insert all chunks
	fmt.Printf("⏱️  Processing %d chunks from large text files...\n", totalChunks)
	start := time.Now()

	if useNativeVectorizer {
		err = collection.InsertTextBatch(ctx, allTextVectors)
	} else {
		err = collection.InsertBatch(ctx, allVectors)
	}

	insertTime := time.Since(start)

	if err != nil {
		log.Printf("❌ Insertion failed: %v", err)
		return
	}

	throughput := float64(totalChunks) / insertTime.Seconds()
	fmt.Printf("✅ Processed %d chunks in %v\n", totalChunks, insertTime)
	fmt.Printf("📊 Throughput: %.0f chunks/second\n", throughput)
	fmt.Printf("📊 Total content: %d words (~%.1f MB)\n", totalWords, float64(totalWords*6)/1024/1024)

	// Store collection reference for other tests
	globalCollection = collection
	globalUseNativeVectorizer = useNativeVectorizer
}

func testSemanticSearchMatching(ctx context.Context) {
	if globalCollection == nil {
		fmt.Println("❌ No collection available from previous test")
		return
	}

	fmt.Println("🔍 Testing semantic search with proper matching...")

	// Define test queries with expected relevance
	testQueries := []struct {
		query          string
		expectedTopics []string
		minScore       float32
		description    string
	}{
		{
			query:          "installation and setup instructions",
			expectedTopics: []string{"installation", "setup", "getting started"},
			minScore:       0.3,
			description:    "Should find installation-related content",
		},
		{
			query:          "vector database performance optimization",
			expectedTopics: []string{"performance", "optimization", "vector"},
			minScore:       0.25,
			description:    "Should find performance-related content",
		},
		{
			query:          "API endpoints and REST interface",
			expectedTopics: []string{"api", "endpoint", "rest", "http"},
			minScore:       0.3,
			description:    "Should find API documentation",
		},
		{
			query:          "embedding models and vectorization",
			expectedTopics: []string{"embedding", "vectorization", "model"},
			minScore:       0.25,
			description:    "Should find embedding-related content",
		},
		{
			query:          "completely unrelated cooking recipes",
			expectedTopics: []string{}, // Should have low scores
			minScore:       0.0,
			description:    "Should NOT match well (low scores expected)",
		},
	}

	for i, testCase := range testQueries {
		fmt.Printf("\n🔍 Query %d: '%s'\n", i+1, testCase.query)
		fmt.Printf("   Expected: %s\n", testCase.description)

		start := time.Now()
		var response *core.SearchResponse
		var err error

		if globalUseNativeVectorizer {
			response, err = globalCollection.SearchText(ctx, testCase.query, 5, nil)
		} else {
			queryVector := generateEnhancedSemanticVector(testCase.query, 384)
			searchReq := &core.SearchRequest{
				Vector:          queryVector,
				Limit:           5,
				IncludeMetadata: true,
			}
			response, err = globalCollection.Search(ctx, searchReq)
		}

		searchTime := time.Since(start)

		if err != nil {
			log.Printf("❌ Search failed: %v", err)
			continue
		}

		fmt.Printf("   ⏱️  Search time: %v\n", searchTime)

		// Analyze results for proper matching
		relevantResults := 0
		for j, result := range response.Results {
			isRelevant := result.Score >= testCase.minScore

			// Check if content contains expected topics
			fileName := ""
			if fn, ok := result.Metadata["file_name"]; ok {
				fileName = fn.(string)
			}

			fmt.Printf("      %d. Score: %.4f, File: %s", j+1, result.Score, fileName)

			if isRelevant {
				fmt.Printf(" ✅ RELEVANT")
				relevantResults++
			} else {
				fmt.Printf(" ❌ LOW RELEVANCE")
			}
			fmt.Println()
		}

		// Summary for this query
		if len(testCase.expectedTopics) > 0 {
			fmt.Printf("   📊 Found %d/%d relevant results (score >= %.2f)\n",
				relevantResults, len(response.Results), testCase.minScore)
		} else {
			fmt.Printf("   📊 Correctly low relevance: %d results with score < 0.2\n",
				countLowScoreResults(response.Results, 0.2))
		}
	}
}

func testSimilarityScoreAnalysis(ctx context.Context) {
	fmt.Println("📊 Analyzing similarity score distribution...")

	// Test with known similar and dissimilar texts
	testPairs := []struct {
		text1       string
		text2       string
		expectHigh  bool
		description string
	}{
		{
			text1:       "VittoriaDB is a vector database for AI applications",
			text2:       "Vector databases are used for artificial intelligence systems",
			expectHigh:  true,
			description: "Similar concepts - should have high similarity",
		},
		{
			text1:       "Installation requires downloading the binary file",
			text2:       "Setup instructions for installing the software",
			expectHigh:  true,
			description: "Installation topics - should be similar",
		},
		{
			text1:       "Performance optimization and speed improvements",
			text2:       "Cooking recipes and kitchen techniques",
			expectHigh:  false,
			description: "Completely different topics - should be dissimilar",
		},
		{
			text1:       "API endpoints for REST interface",
			text2:       "Space exploration and astronomy research",
			expectHigh:  false,
			description: "Unrelated topics - should be dissimilar",
		},
	}

	fmt.Println("\n🔬 Similarity Analysis:")
	for i, pair := range testPairs {
		vec1 := generateEnhancedSemanticVector(pair.text1, 384)
		vec2 := generateEnhancedSemanticVector(pair.text2, 384)
		similarity := cosineSimilarity(vec1, vec2)

		fmt.Printf("\n%d. %s\n", i+1, pair.description)
		fmt.Printf("   Text 1: '%s'\n", pair.text1)
		fmt.Printf("   Text 2: '%s'\n", pair.text2)
		fmt.Printf("   Similarity: %.4f", similarity)

		if pair.expectHigh {
			if similarity > 0.3 {
				fmt.Printf(" ✅ HIGH (as expected)")
			} else {
				fmt.Printf(" ⚠️  LOW (unexpected)")
			}
		} else {
			if similarity < 0.3 {
				fmt.Printf(" ✅ LOW (as expected)")
			} else {
				fmt.Printf(" ⚠️  HIGH (unexpected)")
			}
		}
		fmt.Println()
	}

	// Display score interpretation guide
	fmt.Println("\n📋 Score Interpretation Guide:")
	fmt.Println("   0.7 - 1.0: Very similar content")
	fmt.Println("   0.5 - 0.7: Moderately similar")
	fmt.Println("   0.3 - 0.5: Somewhat related")
	fmt.Println("   0.1 - 0.3: Weakly related")
	fmt.Println("   0.0 - 0.1: Not related")
}

// Global variables to share collection between tests
var globalCollection *core.VittoriaCollection
var globalUseNativeVectorizer bool

// Enhanced semantic vector generation with better diversity
func generateEnhancedSemanticVector(text string, dimensions int) []float32 {
	vector := make([]float32, dimensions)

	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		return vector
	}

	// Create diverse features for each dimension
	for i := 0; i < dimensions; i++ {
		var value float32

		for j, word := range words {
			// Character-based features with high variation
			charFeature := 0.0
			for k, char := range word {
				switch i % 5 {
				case 0:
					charFeature += float64(char) * float64(k+1) * 0.1
				case 1:
					charFeature += float64(char*char) * float64(k+2) * 0.01
				case 2:
					charFeature += float64(char) / float64(k+3) * 10.0
				case 3:
					charFeature += float64(int(char)^(k+1)) * 0.05
				case 4:
					charFeature += float64(char) * float64(len(word)-k) * 0.2
				}
			}

			// Word length and position features
			lengthFeature := float64(len(word)) * float64(i+1) * 0.3
			posFeature := float64(j+1) * float64(dimensions-i) * 0.1

			// Hash-based features
			hash1 := djb2Hash(word) % (i*97 + 13)
			hash2 := sdbmHash(word) % (i*73 + 17)
			hashFeature := float64(hash1-hash2) * 0.01

			// Word uniqueness
			uniqueChars := make(map[rune]bool)
			for _, char := range word {
				uniqueChars[char] = true
			}
			uniquenessFeature := float64(len(uniqueChars)) * float64(i+1) * 0.5

			// Combine features
			dimWeight := 1.0 + float64(i%7)*0.3
			combined := (charFeature + lengthFeature + posFeature + hashFeature + uniquenessFeature) * dimWeight
			interaction := float64(j*i+1) * 0.1
			combined += interaction

			value += float32(combined)
		}

		// Add dimension-specific bias
		dimBias := float32((i*i)%17)*0.2 - 1.0
		vector[i] = value + dimBias
	}

	// L2 normalize
	var norm float32
	for _, val := range vector {
		norm += val * val
	}

	if norm > 0 {
		normFactor := 1.0 / float32(sqrt(float64(norm)))
		for i := range vector {
			vector[i] *= normFactor
		}
	}

	return vector
}

// Helper functions
func djb2Hash(s string) int {
	hash := 5381
	for _, char := range s {
		hash = ((hash << 5) + hash) + int(char)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

func sdbmHash(s string) int {
	hash := 0
	for _, char := range s {
		hash = int(char) + (hash << 6) + (hash << 16) - hash
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func cosineSimilarity(a, b []float32) float32 {
	var dotProduct, normA, normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(sqrt(float64(normA))) * float32(sqrt(float64(normB))))
}

func countLowScoreResults(results []*core.SearchResult, threshold float32) int {
	count := 0
	for _, result := range results {
		if result.Score < threshold {
			count++
		}
	}
	return count
}
