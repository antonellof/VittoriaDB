package core

import (
	"encoding/json"
)

// ParseFilterJSON interprets filter JSON either as a structured core.Filter tree
// or as a flat map of field names to values (implicit AND of equality clauses),
// matching examples in docs that use {"category": "tech"}.
func ParseFilterJSON(raw []byte) (*Filter, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var structured Filter
	if err := json.Unmarshal(raw, &structured); err != nil {
		return nil, err
	}

	if filterIsStructured(&structured) {
		return &structured, nil
	}

	var flat map[string]interface{}
	if err := json.Unmarshal(raw, &flat); err != nil {
		return nil, err
	}
	if len(flat) == 0 {
		return nil, nil
	}

	clauses := make([]Filter, 0, len(flat))
	for k, v := range flat {
		key := k
		clauses = append(clauses, Filter{
			Field:    key,
			Operator: FilterOpEq,
			Value:    v,
		})
	}
	if len(clauses) == 1 {
		f := clauses[0]
		return &f, nil
	}
	return &Filter{And: clauses}, nil
}

func filterIsStructured(f *Filter) bool {
	if f == nil {
		return false
	}
	if len(f.And) > 0 || len(f.Or) > 0 || f.Not != nil {
		return true
	}
	if f.Field != "" {
		return true
	}
	return false
}
