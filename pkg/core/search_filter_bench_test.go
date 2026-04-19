// Benchmarks for metadata filtering on the in-memory search path.
// VittoriaDB currently applies filters as a post-ANN (post-candidate) step: HNSW
// in pkg/index is not yet wired as the primary collection search; the hot path
// scans vectors and skips non-matching metadata before scoring.

package core

import (
	"context"
	"fmt"
	"testing"
)

func benchCollectionFiltered(nVectors int, pctMatch float64, b *testing.B) {
	ctx := context.Background()
	c, err := NewCollection("b", 32, DistanceMetricCosine, IndexTypeFlat, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	if err := c.Initialize(ctx); err != nil {
		b.Fatal(err)
	}

	thresh := int(float64(nVectors) * pctMatch)
	for i := 0; i < nVectors; i++ {
		meta := map[string]interface{}{"tag": "no"}
		if i < thresh {
			meta["tag"] = "yes"
		}
		v := make([]float32, 32)
		for j := range v {
			v[j] = float32(i%7) * 0.01
		}
		if err := c.Insert(ctx, &Vector{ID: fmt.Sprintf("v%d", i), Vector: v, Metadata: meta}); err != nil {
			b.Fatal(err)
		}
	}

	q := make([]float32, 32)
	q[0] = 1

	f := &Filter{Field: "tag", Operator: FilterOpEq, Value: "yes"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Search(ctx, &SearchRequest{
			Vector: q, Limit: 10, Filter: f,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSearch_FilterHighSelectivity_10k(b *testing.B) {
	benchCollectionFiltered(10000, 0.05, b)
}

func BenchmarkSearch_NoFilter_10k(b *testing.B) {
	ctx := context.Background()
	c, err := NewCollection("b", 32, DistanceMetricCosine, IndexTypeFlat, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	if err := c.Initialize(ctx); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 10000; i++ {
		v := make([]float32, 32)
		v[0] = float32(i) * 0.001
		if err := c.Insert(ctx, &Vector{ID: fmt.Sprintf("v%d", i), Vector: v}); err != nil {
			b.Fatal(err)
		}
	}
	q := make([]float32, 32)
	q[0] = 1
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Search(ctx, &SearchRequest{Vector: q, Limit: 10})
		if err != nil {
			b.Fatal(err)
		}
	}
}
