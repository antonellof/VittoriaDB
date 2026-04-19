package hybrid

import "testing"

func TestNormalizedWeights(t *testing.T) {
	a, b := NormalizedWeights(HybridWeights{Alpha: 0.7, Beta: 0.3})
	if a < 0.69 || a > 0.71 || b < 0.29 || b > 0.31 {
		t.Fatalf("got %g, %g", a, b)
	}
	a2, b2 := NormalizedWeights(HybridWeights{})
	if a2 != 0.5 || b2 != 0.5 {
		t.Fatalf("defaults %g %g", a2, b2)
	}
}
