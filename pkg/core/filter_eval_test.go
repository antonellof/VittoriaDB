package core

import "testing"

func TestEvalMetadataFilter_Eq(t *testing.T) {
	meta := map[string]interface{}{"genre": "tech", "year": float64(2024)}
	f := &Filter{Field: "genre", Operator: FilterOpEq, Value: "tech"}
	if !evalMetadataFilter(meta, f) {
		t.Fatal("expected match")
	}
	f2 := &Filter{Field: "genre", Operator: FilterOpEq, Value: "news"}
	if evalMetadataFilter(meta, f2) {
		t.Fatal("expected no match")
	}
}

func TestEvalMetadataFilter_AndOrNot(t *testing.T) {
	meta := map[string]interface{}{"a": float64(1), "b": "x"}

	and := &Filter{
		And: []Filter{
			{Field: "a", Operator: FilterOpEq, Value: float64(1)},
			{Field: "b", Operator: FilterOpEq, Value: "x"},
		},
	}
	if !evalMetadataFilter(meta, and) {
		t.Fatal("AND")
	}

	or := &Filter{
		Or: []Filter{
			{Field: "b", Operator: FilterOpEq, Value: "y"},
			{Field: "b", Operator: FilterOpEq, Value: "x"},
		},
	}
	if !evalMetadataFilter(meta, or) {
		t.Fatal("OR")
	}

	not := &Filter{
		Not: &Filter{Field: "b", Operator: FilterOpEq, Value: "y"},
	}
	if !evalMetadataFilter(meta, not) {
		t.Fatal("NOT")
	}
}

func TestEvalMetadataFilter_Compare(t *testing.T) {
	meta := map[string]interface{}{"score": float64(42)}
	for _, tc := range []struct {
		op     FilterOp
		val    interface{}
		expect bool
	}{
		{FilterOpGt, float64(40), true},
		{FilterOpGte, float64(42), true},
		{FilterOpLt, float64(50), true},
		{FilterOpLte, float64(41), false},
	} {
		f := &Filter{Field: "score", Operator: tc.op, Value: tc.val}
		if got := evalMetadataFilter(meta, f); got != tc.expect {
			t.Fatalf("%s %v: got %v want %v", tc.op, tc.val, got, tc.expect)
		}
	}
}

func TestEvalMetadataFilter_InContainsExists(t *testing.T) {
	meta := map[string]interface{}{
		"tags": []interface{}{"a", "b"},
		"body": "hello world",
	}
	in := &Filter{Field: "tags", Operator: FilterOpIn, Value: []interface{}{"x", "b", "z"}}
	if !evalMetadataFilter(meta, in) {
		t.Fatal("in")
	}
	contains := &Filter{Field: "body", Operator: FilterOpContains, Value: "world"}
	if !evalMetadataFilter(meta, contains) {
		t.Fatal("contains")
	}
	exists := &Filter{Field: "body", Operator: FilterOpExists}
	if !evalMetadataFilter(meta, exists) {
		t.Fatal("exists")
	}
}
