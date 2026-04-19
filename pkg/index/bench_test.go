package index

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
)

// randomVectors generates n unit-ish random float32 vectors of dimension d.
func randomVectors(n, d int, seed int64) []*IndexVector {
	r := rand.New(rand.NewSource(seed))
	out := make([]*IndexVector, n)
	for i := 0; i < n; i++ {
		v := make([]float32, d)
		var norm float32
		for j := 0; j < d; j++ {
			x := r.Float32()*2 - 1
			v[j] = x
			norm += x * x
		}
		// normalize so cosine similarity is meaningful
		if norm > 0 {
			inv := 1.0 / float32(sqrt32(norm))
			for j := 0; j < d; j++ {
				v[j] *= inv
			}
		}
		out[i] = &IndexVector{
			ID:     fmt.Sprintf("v-%d", i),
			Vector: v,
		}
	}
	return out
}

func sqrt32(x float32) float32 {
	// Fast approximation isn't important here; use float64 sqrt.
	if x <= 0 {
		return 0
	}
	z := float32(1)
	for i := 0; i < 6; i++ {
		z = 0.5 * (z + x/z)
	}
	return z
}

// BenchmarkHNSWInsert measures single-vector insertion throughput at a fixed
// dimension, using default HNSW parameters (M=16, efConstruction=200).
func BenchmarkHNSWInsert(b *testing.B) {
	const dim = 384
	vectors := randomVectors(b.N, dim, 1)

	idx, err := CreateIndex(IndexTypeHNSW, dim, DistanceMetricCosine, nil)
	if err != nil {
		b.Fatalf("CreateIndex: %v", err)
	}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := idx.Add(ctx, vectors[i]); err != nil {
			b.Fatalf("Add: %v", err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "vectors/sec")
}

// BenchmarkHNSWSearch measures KNN search throughput on a 10k-vector index.
func BenchmarkHNSWSearch(b *testing.B) {
	const (
		dim = 384
		n   = 10000
		k   = 10
	)

	idx, err := CreateIndex(IndexTypeHNSW, dim, DistanceMetricCosine, nil)
	if err != nil {
		b.Fatalf("CreateIndex: %v", err)
	}
	ctx := context.Background()

	build := randomVectors(n, dim, 2)
	for _, v := range build {
		if err := idx.Add(ctx, v); err != nil {
			b.Fatalf("Add during build: %v", err)
		}
	}
	queries := randomVectors(b.N, dim, 3)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := idx.Search(ctx, queries[i].Vector, k, nil); err != nil {
			b.Fatalf("Search: %v", err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "queries/sec")
}

// BenchmarkFlatSearch is the brute-force baseline at the same dataset size.
func BenchmarkFlatSearch(b *testing.B) {
	const (
		dim = 384
		n   = 10000
		k   = 10
	)

	idx, err := CreateIndex(IndexTypeFlat, dim, DistanceMetricCosine, nil)
	if err != nil {
		b.Fatalf("CreateIndex: %v", err)
	}
	ctx := context.Background()

	build := randomVectors(n, dim, 2)
	for _, v := range build {
		if err := idx.Add(ctx, v); err != nil {
			b.Fatalf("Add during build: %v", err)
		}
	}
	queries := randomVectors(b.N, dim, 3)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := idx.Search(ctx, queries[i].Vector, k, nil); err != nil {
			b.Fatalf("Search: %v", err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "queries/sec")
}
