package index

import (
	"math"
)

// CosineDistanceCalculator implements cosine similarity
type CosineDistanceCalculator struct{}

func (c *CosineDistanceCalculator) Calculate(a, b []float32) float32 {
	return 1.0 - cosineSimilarity(a, b)
}

func (c *CosineDistanceCalculator) Name() string {
	return "cosine"
}

func (c *CosineDistanceCalculator) IsSymmetric() bool {
	return true
}

// EuclideanDistanceCalculator implements Euclidean distance
type EuclideanDistanceCalculator struct{}

func (e *EuclideanDistanceCalculator) Calculate(a, b []float32) float32 {
	return euclideanDistance(a, b)
}

func (e *EuclideanDistanceCalculator) Name() string {
	return "euclidean"
}

func (e *EuclideanDistanceCalculator) IsSymmetric() bool {
	return true
}

// DotProductDistanceCalculator implements dot product distance
type DotProductDistanceCalculator struct{}

func (d *DotProductDistanceCalculator) Calculate(a, b []float32) float32 {
	return -dotProduct(a, b) // Negative for max-heap behavior
}

func (d *DotProductDistanceCalculator) Name() string {
	return "dot_product"
}

func (d *DotProductDistanceCalculator) IsSymmetric() bool {
	return true
}

// ManhattanDistanceCalculator implements Manhattan distance
type ManhattanDistanceCalculator struct{}

func (m *ManhattanDistanceCalculator) Calculate(a, b []float32) float32 {
	return manhattanDistance(a, b)
}

func (m *ManhattanDistanceCalculator) Name() string {
	return "manhattan"
}

func (m *ManhattanDistanceCalculator) IsSymmetric() bool {
	return true
}

// NewDistanceCalculator creates a distance calculator for the given metric
func NewDistanceCalculator(metric DistanceMetric) DistanceCalculator {
	switch metric {
	case DistanceMetricCosine:
		return &CosineDistanceCalculator{}
	case DistanceMetricEuclidean:
		return &EuclideanDistanceCalculator{}
	case DistanceMetricDotProduct:
		return &DotProductDistanceCalculator{}
	case DistanceMetricManhattan:
		return &ManhattanDistanceCalculator{}
	default:
		return &CosineDistanceCalculator{} // Default to cosine
	}
}

// Distance calculation functions

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

func euclideanDistance(a, b []float32) float32 {
	var sum float32
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return float32(math.Sqrt(float64(sum)))
}

func dotProduct(a, b []float32) float32 {
	var sum float32
	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
	}
	return sum
}

func manhattanDistance(a, b []float32) float32 {
	var sum float32
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		if diff < 0 {
			diff = -diff
		}
		sum += diff
	}
	return sum
}

// SIMD optimized versions (placeholder for future implementation)
// These would use assembly or CGO for actual SIMD instructions

// dotProductAVX2 would be implemented in assembly for SIMD optimization
func dotProductAVX2(a, b []float32) float32 {
	// Placeholder - would be implemented in assembly
	return dotProduct(a, b)
}

// cosineSimilarityAVX2 would be implemented in assembly for SIMD optimization
func cosineSimilarityAVX2(a, b []float32) float32 {
	// Placeholder - would be implemented in assembly
	return cosineSimilarity(a, b)
}

// CPU feature detection and function selection
var (
	useSIMD bool = false // Would be set based on CPU features
)

func init() {
	// CPU feature detection would go here
	// For now, use standard implementations
}

// OptimizedDotProduct uses SIMD if available
func OptimizedDotProduct(a, b []float32) float32 {
	if useSIMD && len(a) >= 8 { // Minimum vector size for SIMD
		return dotProductAVX2(a, b)
	}
	return dotProduct(a, b)
}

// OptimizedCosineSimilarity uses SIMD if available
func OptimizedCosineSimilarity(a, b []float32) float32 {
	if useSIMD && len(a) >= 8 {
		return cosineSimilarityAVX2(a, b)
	}
	return cosineSimilarity(a, b)
}
