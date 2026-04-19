package index

import (
	"context"
	"fmt"
	"testing"
)

func TestIVFFlatIndex_BuildSearch(t *testing.T) {
	cfg := &IVFConfig{NList: 4, NProbe: 2}
	idx := NewIVFFlatIndex(4, DistanceMetricEuclidean, cfg)

	var vecs []*IndexVector
	for i := 0; i < 40; i++ {
		v := make([]float32, 4)
		v[0] = float32(i)
		v[1] = float32(i % 3)
		vecs = append(vecs, &IndexVector{ID: fmt.Sprintf("id%d", i), Vector: v})
	}
	if err := idx.Build(vecs); err != nil {
		t.Fatal(err)
	}

	q := []float32{10, 1, 0, 0}
	out, err := idx.Search(context.Background(), q, 5, &SearchParams{NProbes: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 5 {
		t.Fatalf("want 5 results got %d", len(out))
	}
}
