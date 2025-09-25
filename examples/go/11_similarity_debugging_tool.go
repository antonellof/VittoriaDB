package main

import (
	"fmt"
	"math"
	"strings"
)

func main() {
	fmt.Println("🔍 VittoriaDB Similarity Debugging Tool")
	fmt.Println("======================================")
	fmt.Println("Analyzing vector generation and similarity calculations")
	fmt.Println()

	// Test 1: Basic Similarity Validation
	fmt.Println("📊 Test 1: Basic Similarity Validation")
	fmt.Println("-------------------------------------")
	testBasicSimilarity()
	fmt.Println()

	// Test 2: Vector Generation Quality
	fmt.Println("🔧 Test 2: Vector Generation Quality")
	fmt.Println("-----------------------------------")
	testVectorGeneration()
	fmt.Println()

	// Test 3: Real-World Text Similarity
	fmt.Println("🌍 Test 3: Real-World Text Similarity")
	fmt.Println("------------------------------------")
	testRealWorldSimilarity()
	fmt.Println()

	// Test 4: Score Distribution Analysis
	fmt.Println("📈 Test 4: Score Distribution Analysis")
	fmt.Println("-------------------------------------")
	testScoreDistribution()
	fmt.Println()

	fmt.Println("✅ Similarity debugging completed!")
}

func testBasicSimilarity() {
	fmt.Println("Testing with known vectors:")

	// Test with orthogonal vectors
	vec1 := []float32{1.0, 0.0, 0.0}
	vec2 := []float32{0.0, 1.0, 0.0}
	vec3 := []float32{1.0, 0.0, 0.0} // Same as vec1

	sim12 := cosineSimilarity(vec1, vec2)
	sim13 := cosineSimilarity(vec1, vec3)

	fmt.Printf("   Orthogonal vectors [1,0,0] vs [0,1,0]: %.6f (should be 0.0) ", sim12)
	if sim12 < 0.001 {
		fmt.Println("✅ CORRECT")
	} else {
		fmt.Println("❌ INCORRECT")
	}

	fmt.Printf("   Identical vectors [1,0,0] vs [1,0,0]: %.6f (should be 1.0) ", sim13)
	if sim13 > 0.999 {
		fmt.Println("✅ CORRECT")
	} else {
		fmt.Println("❌ INCORRECT")
	}

	// Test with normalized vectors
	vec4 := []float32{0.6, 0.8, 0.0} // Normalized
	vec5 := []float32{0.8, 0.6, 0.0} // Different but similar direction
	sim45 := cosineSimilarity(vec4, vec5)

	fmt.Printf("   Similar direction vectors: %.6f (should be > 0.5) ", sim45)
	if sim45 > 0.5 {
		fmt.Println("✅ CORRECT")
	} else {
		fmt.Println("❌ INCORRECT")
	}
}

func testVectorGeneration() {
	fmt.Println("Testing vector generation with different texts:")

	testTexts := []string{
		"machine learning algorithms",
		"machine learning algorithms", // Identical
		"deep learning neural networks",
		"cooking recipes and food",
		"space exploration astronomy",
	}

	vectors := make([][]float32, len(testTexts))
	for i, text := range testTexts {
		vectors[i] = generateEnhancedVector(text, 10) // Small dimension for debugging
		fmt.Printf("   Text %d: '%s'\n", i+1, text)
		fmt.Printf("   Vector: %v\n", vectors[i])
		fmt.Printf("   Norm: %.6f\n\n", vectorNorm(vectors[i]))
	}

	// Test similarity between generated vectors
	fmt.Println("Similarity matrix:")
	fmt.Print("     ")
	for i := range testTexts {
		fmt.Printf("  T%d   ", i+1)
	}
	fmt.Println()

	for i := range vectors {
		fmt.Printf("T%d  ", i+1)
		for j := range vectors {
			sim := cosineSimilarity(vectors[i], vectors[j])
			fmt.Printf(" %.3f", sim)
		}
		fmt.Println()
	}

	// Analyze results
	fmt.Println("\nAnalysis:")
	identical := cosineSimilarity(vectors[0], vectors[1])
	fmt.Printf("   Identical texts (T1 vs T2): %.4f ", identical)
	if identical > 0.99 {
		fmt.Println("✅ HIGH (correct)")
	} else {
		fmt.Println("❌ LOW (should be ~1.0)")
	}

	related := cosineSimilarity(vectors[0], vectors[2])
	fmt.Printf("   Related texts (T1 vs T3): %.4f ", related)
	if related > 0.3 && related < 0.8 {
		fmt.Println("✅ MODERATE (correct)")
	} else {
		fmt.Printf("⚠️  UNEXPECTED (should be 0.3-0.8)")
	}
	fmt.Println()

	unrelated := cosineSimilarity(vectors[0], vectors[3])
	fmt.Printf("   Unrelated texts (T1 vs T4): %.4f ", unrelated)
	if unrelated < 0.3 {
		fmt.Println("✅ LOW (correct)")
	} else {
		fmt.Println("⚠️  HIGH (should be <0.3)")
	}
}

func testRealWorldSimilarity() {
	fmt.Println("Testing with realistic text pairs:")

	testPairs := []struct {
		text1    string
		text2    string
		expected string
		minScore float32
		maxScore float32
	}{
		{
			"VittoriaDB is a vector database for AI applications",
			"Vector databases store embeddings for machine learning",
			"HIGH", 0.4, 1.0,
		},
		{
			"Installation requires downloading the binary file",
			"Setup instructions for installing the software",
			"HIGH", 0.4, 1.0,
		},
		{
			"Performance optimization and indexing strategies",
			"Speed improvements and efficient algorithms",
			"MODERATE", 0.2, 0.6,
		},
		{
			"Vector database performance optimization",
			"Cooking recipes and kitchen techniques",
			"LOW", 0.0, 0.3,
		},
		{
			"API endpoints and REST interface documentation",
			"Space exploration and astronomical research",
			"LOW", 0.0, 0.3,
		},
	}

	for i, pair := range testPairs {
		vec1 := generateEnhancedVector(pair.text1, 384)
		vec2 := generateEnhancedVector(pair.text2, 384)
		similarity := cosineSimilarity(vec1, vec2)

		fmt.Printf("%d. Expected: %s similarity\n", i+1, pair.expected)
		fmt.Printf("   Text 1: '%s'\n", pair.text1)
		fmt.Printf("   Text 2: '%s'\n", pair.text2)
		fmt.Printf("   Similarity: %.4f ", similarity)

		if similarity >= pair.minScore && similarity <= pair.maxScore {
			fmt.Printf("✅ CORRECT (%s)\n", pair.expected)
		} else {
			fmt.Printf("❌ INCORRECT (expected %s: %.2f-%.2f)\n",
				pair.expected, pair.minScore, pair.maxScore)
		}
		fmt.Println()
	}
}

func testScoreDistribution() {
	fmt.Println("Analyzing score distribution across different text types:")

	// Generate a variety of texts
	texts := []string{
		// Technical documentation
		"VittoriaDB vector database performance optimization",
		"API endpoints REST interface documentation",
		"Installation setup configuration instructions",

		// Machine learning
		"machine learning algorithms neural networks",
		"deep learning artificial intelligence models",
		"embedding vectorization semantic search",

		// Completely different
		"cooking recipes kitchen techniques food",
		"space exploration astronomy planets stars",
		"music instruments guitar piano composition",
	}

	fmt.Println("Score distribution matrix:")

	// Calculate all pairwise similarities
	similarities := make([][]float32, len(texts))
	for i := range similarities {
		similarities[i] = make([]float32, len(texts))
	}

	for i := 0; i < len(texts); i++ {
		vec1 := generateEnhancedVector(texts[i], 384)
		for j := 0; j < len(texts); j++ {
			vec2 := generateEnhancedVector(texts[j], 384)
			similarities[i][j] = cosineSimilarity(vec1, vec2)
		}
	}

	// Display matrix
	fmt.Print("     ")
	for i := range texts {
		fmt.Printf("  %2d  ", i+1)
	}
	fmt.Println()

	for i := range similarities {
		fmt.Printf("%2d  ", i+1)
		for j := range similarities[i] {
			fmt.Printf(" %.3f", similarities[i][j])
		}
		fmt.Println()
	}

	// Analyze distribution
	var allScores []float32
	for i := 0; i < len(similarities); i++ {
		for j := i + 1; j < len(similarities[i]); j++ { // Only upper triangle
			allScores = append(allScores, similarities[i][j])
		}
	}

	// Calculate statistics
	var sum, min, max float32
	min = 1.0
	max = 0.0

	for _, score := range allScores {
		sum += score
		if score < min {
			min = score
		}
		if score > max {
			max = score
		}
	}

	avg := sum / float32(len(allScores))

	fmt.Printf("\nStatistics (%d pairwise comparisons):\n", len(allScores))
	fmt.Printf("   Average: %.4f\n", avg)
	fmt.Printf("   Minimum: %.4f\n", min)
	fmt.Printf("   Maximum: %.4f\n", max)
	fmt.Printf("   Range:   %.4f\n", max-min)

	// Count score ranges
	high := 0   // > 0.5
	medium := 0 // 0.2 - 0.5
	low := 0    // < 0.2

	for _, score := range allScores {
		if score > 0.5 {
			high++
		} else if score > 0.2 {
			medium++
		} else {
			low++
		}
	}

	fmt.Printf("\nScore distribution:\n")
	fmt.Printf("   High (>0.5):   %d (%.1f%%)\n", high, float32(high)/float32(len(allScores))*100)
	fmt.Printf("   Medium (0.2-0.5): %d (%.1f%%)\n", medium, float32(medium)/float32(len(allScores))*100)
	fmt.Printf("   Low (<0.2):    %d (%.1f%%)\n", low, float32(low)/float32(len(allScores))*100)

	fmt.Println("\nText categories:")
	for i, text := range texts {
		category := "Other"
		if i < 3 {
			category = "Technical"
		} else if i < 6 {
			category = "ML/AI"
		} else {
			category = "Unrelated"
		}
		fmt.Printf("   %2d. %s: %s\n", i+1, category, text)
	}
}

// Enhanced vector generation (same as in the main demo)
func generateEnhancedVector(text string, dimensions int) []float32 {
	vector := make([]float32, dimensions)

	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		return vector
	}

	for i := 0; i < dimensions; i++ {
		var value float32

		for j, word := range words {
			// Character-based features
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

			// Other features
			lengthFeature := float64(len(word)) * float64(i+1) * 0.3
			posFeature := float64(j+1) * float64(dimensions-i) * 0.1

			hash1 := djb2Hash(word) % (i*97 + 13)
			hash2 := sdbmHash(word) % (i*73 + 17)
			hashFeature := float64(hash1-hash2) * 0.01

			uniqueChars := make(map[rune]bool)
			for _, char := range word {
				uniqueChars[char] = true
			}
			uniquenessFeature := float64(len(uniqueChars)) * float64(i+1) * 0.5

			dimWeight := 1.0 + float64(i%7)*0.3
			combined := (charFeature + lengthFeature + posFeature + hashFeature + uniquenessFeature) * dimWeight
			interaction := float64(j*i+1) * 0.1
			combined += interaction

			value += float32(combined)
		}

		dimBias := float32((i*i)%17)*0.2 - 1.0
		vector[i] = value + dimBias
	}

	// L2 normalize
	var norm float32
	for _, val := range vector {
		norm += val * val
	}

	if norm > 0 {
		normFactor := 1.0 / float32(math.Sqrt(float64(norm)))
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

func vectorNorm(v []float32) float32 {
	var sum float32
	for _, val := range v {
		sum += val * val
	}
	return float32(math.Sqrt(float64(sum)))
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

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}
