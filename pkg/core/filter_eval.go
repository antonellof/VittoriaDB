package core

import (
	"reflect"
	"strings"
)

// evalMetadataFilter evaluates a nested filter tree against vector metadata.
// Empty or nil filter matches everything.
func evalMetadataFilter(metadata map[string]interface{}, f *Filter) bool {
	if f == nil {
		return true
	}

	if len(f.And) > 0 {
		for i := range f.And {
			if !evalMetadataFilter(metadata, &f.And[i]) {
				return false
			}
		}
		return true
	}
	if len(f.Or) > 0 {
		for i := range f.Or {
			if evalMetadataFilter(metadata, &f.Or[i]) {
				return true
			}
		}
		return false
	}
	if f.Not != nil {
		return !evalMetadataFilter(metadata, f.Not)
	}

	return evalFilterLeaf(metadata, f)
}

func evalFilterLeaf(meta map[string]interface{}, f *Filter) bool {
	if f == nil {
		return true
	}

	switch f.Operator {
	case FilterOpExists:
		_, ok := meta[f.Field]
		return ok
	case "", FilterOpEq:
		if f.Field == "" {
			return true
		}
		got, ok := meta[f.Field]
		if !ok {
			return false
		}
		return valuesEqual(got, f.Value)
	case FilterOpNe:
		if f.Field == "" {
			return false
		}
		got, ok := meta[f.Field]
		if !ok {
			return true
		}
		return !valuesEqual(got, f.Value)
	case FilterOpGt, FilterOpGte, FilterOpLt, FilterOpLte:
		if f.Field == "" {
			return false
		}
		got, ok := meta[f.Field]
		if !ok {
			return false
		}
		gf, gok := toComparableFloat(got)
		wf, wok := toComparableFloat(f.Value)
		if !gok || !wok {
			return false
		}
		switch f.Operator {
		case FilterOpGt:
			return gf > wf
		case FilterOpGte:
			return gf >= wf
		case FilterOpLt:
			return gf < wf
		case FilterOpLte:
			return gf <= wf
		default:
			return false
		}
	case FilterOpIn:
		if f.Field == "" {
			return false
		}
		got, ok := meta[f.Field]
		if !ok {
			return false
		}
		return valueInAllowedList(got, f.Value)
	case FilterOpNotIn:
		if f.Field == "" {
			return false
		}
		got, ok := meta[f.Field]
		if !ok {
			return true
		}
		return !valueInAllowedList(got, f.Value)
	case FilterOpContains:
		if f.Field == "" {
			return false
		}
		got, ok := meta[f.Field]
		if !ok {
			return false
		}
		gs, ok := got.(string)
		if !ok {
			return false
		}
		want, ok := f.Value.(string)
		if !ok {
			return false
		}
		return strings.Contains(gs, want)
	default:
		if f.Field != "" {
			return false
		}
		return true
	}
}

// valueInAllowedList implements `in`: scalar values must appear in the allowed list;
// slice/array values match if any element appears in the allowed list (tag-style semantics).
func valueInAllowedList(got interface{}, allowed interface{}) bool {
	rv := reflect.ValueOf(allowed)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return valuesEqual(got, allowed)
	}
	gotRV := reflect.ValueOf(got)
	if gotRV.Kind() == reflect.Slice || gotRV.Kind() == reflect.Array {
		for i := 0; i < gotRV.Len(); i++ {
			g := gotRV.Index(i).Interface()
			for j := 0; j < rv.Len(); j++ {
				if valuesEqual(g, rv.Index(j).Interface()) {
					return true
				}
			}
		}
		return false
	}
	for j := 0; j < rv.Len(); j++ {
		if valuesEqual(got, rv.Index(j).Interface()) {
			return true
		}
	}
	return false
}

func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	fa, aok := toComparableFloat(a)
	fb, bok := toComparableFloat(b)
	if aok && bok {
		return fa == fb
	}
	return reflect.DeepEqual(a, b)
}

func toComparableFloat(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint64:
		return float64(x), true
	default:
		return 0, false
	}
}
