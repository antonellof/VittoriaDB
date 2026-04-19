package core

import (
	"encoding/json"
	"testing"
)

func TestParseFilterJSON_FlatMap(t *testing.T) {
	raw := []byte(`{"category":"tech","year":2024}`)
	f, err := ParseFilterJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if f == nil || len(f.And) != 2 {
		t.Fatalf("expected implicit AND of 2 clauses, got %+v", f)
	}
}

func TestParseFilterJSON_Structured(t *testing.T) {
	raw := []byte(`{"field":"genre","operator":"eq","value":"tech"}`)
	f, err := ParseFilterJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if f == nil || f.Field != "genre" || f.Value != "tech" {
		t.Fatalf("got %+v", f)
	}
}

func TestParseFilterJSON_AndChain(t *testing.T) {
	raw := []byte(`{"and":[{"field":"a","operator":"eq","value":1},{"field":"b","operator":"eq","value":"x"}]}`)
	f, err := ParseFilterJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(f)
	if len(f.And) != 2 {
		t.Fatalf("unexpected %s", string(b))
	}
}
