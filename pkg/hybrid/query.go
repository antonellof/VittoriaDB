package hybrid

// HybridWeights controls how lexical and dense scores are fused.
// Alpha applies to dense similarity (typically cosine in [0,1]); Beta to BM25
// scores normalized per-query to [0,1]. Alpha+Beta should sum to 1 for a
// simple convex combination (future implementation).
type HybridWeights struct {
	Alpha float64 `json:"alpha"` // Dense channel
	Beta  float64 `json:"beta"`  // Lexical channel (BM25 / FTS)
}

// NormalizedWeights returns alpha,beta scaled to sum to 1 when both positive.
func NormalizedWeights(w HybridWeights) (a, b float64) {
	a, b = w.Alpha, w.Beta
	if a <= 0 && b <= 0 {
		return 0.5, 0.5
	}
	s := a + b
	if s <= 0 {
		return 0.5, 0.5
	}
	return a / s, b / s
}

// HybridQuery describes a future combined retrieval request (API sketch).
type HybridQuery struct {
	QueryText string        `json:"query_text"`
	QueryVec  []float32     `json:"query_vector,omitempty"`
	Weights   HybridWeights `json:"weights,omitempty"`
	Limit     int           `json:"limit,omitempty"`
}
