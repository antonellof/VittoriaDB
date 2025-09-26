package core

import (
	"fmt"
	"strings"
)

// SchemaType represents different data types supported in the schema
type SchemaType string

const (
	SchemaTypeString     SchemaType = "string"
	SchemaTypeNumber     SchemaType = "number"
	SchemaTypeBoolean    SchemaType = "boolean"
	SchemaTypeEnum       SchemaType = "enum"
	SchemaTypeGeopoint   SchemaType = "geopoint"
	SchemaTypeStringArr  SchemaType = "string[]"
	SchemaTypeNumberArr  SchemaType = "number[]"
	SchemaTypeBooleanArr SchemaType = "boolean[]"
	SchemaTypeEnumArr    SchemaType = "enum[]"
	SchemaTypeVector     SchemaType = "vector"
	SchemaTypeObject     SchemaType = "object"
)

// Schema represents a document schema definition
type Schema map[string]interface{}

// SchemaField represents a field definition in the schema
type SchemaField struct {
	Type       SchemaType    `json:"type"`
	Dimensions int           `json:"dimensions,omitempty"`  // For vector fields
	Properties Schema        `json:"properties,omitempty"`  // For object fields
	EnumValues []interface{} `json:"enum_values,omitempty"` // For enum fields
	Required   bool          `json:"required,omitempty"`
}

// Document represents a document that conforms to a schema
type Document map[string]interface{}

// SchemaValidator validates documents against schemas
type SchemaValidator struct {
	schema Schema
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator(schema Schema) *SchemaValidator {
	return &SchemaValidator{schema: schema}
}

// ValidateDocument validates a document against the schema
func (sv *SchemaValidator) ValidateDocument(doc Document) error {
	return sv.validateObject(doc, sv.schema, "")
}

// validateObject validates an object against a schema
func (sv *SchemaValidator) validateObject(obj map[string]interface{}, schema Schema, path string) error {
	for key, schemaValue := range schema {
		currentPath := path
		if currentPath != "" {
			currentPath += "."
		}
		currentPath += key

		value, exists := obj[key]

		// Handle schema field definition
		if field, ok := schemaValue.(SchemaField); ok {
			if !exists && field.Required {
				return fmt.Errorf("required field '%s' is missing", currentPath)
			}
			if exists {
				if err := sv.validateValue(value, field, currentPath); err != nil {
					return err
				}
			}
			continue
		}

		// Handle simple type definitions (backward compatibility)
		if typeStr, ok := schemaValue.(string); ok {
			if !exists {
				continue // Optional field
			}
			field := SchemaField{Type: SchemaType(typeStr)}
			if strings.HasPrefix(typeStr, "vector[") && strings.HasSuffix(typeStr, "]") {
				// Parse vector dimensions
				dimStr := typeStr[7 : len(typeStr)-1]
				var dims int
				if _, err := fmt.Sscanf(dimStr, "%d", &dims); err == nil {
					field.Dimensions = dims
				}
			}
			if err := sv.validateValue(value, field, currentPath); err != nil {
				return err
			}
			continue
		}

		// Handle nested object schemas
		if nestedSchema, ok := schemaValue.(Schema); ok {
			if !exists {
				continue // Optional nested object
			}
			if nestedObj, ok := value.(map[string]interface{}); ok {
				if err := sv.validateObject(nestedObj, nestedSchema, currentPath); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("field '%s' must be an object", currentPath)
			}
		}
	}

	return nil
}

// validateValue validates a single value against a schema field
func (sv *SchemaValidator) validateValue(value interface{}, field SchemaField, path string) error {
	switch field.Type {
	case SchemaTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' must be a string", path)
		}
	case SchemaTypeNumber:
		if !isNumber(value) {
			return fmt.Errorf("field '%s' must be a number", path)
		}
	case SchemaTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field '%s' must be a boolean", path)
		}
	case SchemaTypeEnum:
		if len(field.EnumValues) > 0 {
			valid := false
			for _, enumVal := range field.EnumValues {
				if value == enumVal {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("field '%s' must be one of %v", path, field.EnumValues)
			}
		}
	case SchemaTypeVector:
		if arr, ok := value.([]interface{}); ok {
			if field.Dimensions > 0 && len(arr) != field.Dimensions {
				return fmt.Errorf("field '%s' must be a vector of %d dimensions, got %d", path, field.Dimensions, len(arr))
			}
			for i, v := range arr {
				if !isNumber(v) {
					return fmt.Errorf("field '%s[%d]' must be a number", path, i)
				}
			}
		} else if arr, ok := value.([]float32); ok {
			if field.Dimensions > 0 && len(arr) != field.Dimensions {
				return fmt.Errorf("field '%s' must be a vector of %d dimensions, got %d", path, field.Dimensions, len(arr))
			}
		} else if arr, ok := value.([]float64); ok {
			if field.Dimensions > 0 && len(arr) != field.Dimensions {
				return fmt.Errorf("field '%s' must be a vector of %d dimensions, got %d", path, field.Dimensions, len(arr))
			}
		} else {
			return fmt.Errorf("field '%s' must be a vector (array of numbers)", path)
		}
	case SchemaTypeStringArr:
		if arr, ok := value.([]interface{}); ok {
			for i, v := range arr {
				if _, ok := v.(string); !ok {
					return fmt.Errorf("field '%s[%d]' must be a string", path, i)
				}
			}
		} else {
			return fmt.Errorf("field '%s' must be an array of strings", path)
		}
	case SchemaTypeNumberArr:
		if arr, ok := value.([]interface{}); ok {
			for i, v := range arr {
				if !isNumber(v) {
					return fmt.Errorf("field '%s[%d]' must be a number", path, i)
				}
			}
		} else {
			return fmt.Errorf("field '%s' must be an array of numbers", path)
		}
	case SchemaTypeBooleanArr:
		if arr, ok := value.([]interface{}); ok {
			for i, v := range arr {
				if _, ok := v.(bool); !ok {
					return fmt.Errorf("field '%s[%d]' must be a boolean", path, i)
				}
			}
		} else {
			return fmt.Errorf("field '%s' must be an array of booleans", path)
		}
	case SchemaTypeObject:
		if obj, ok := value.(map[string]interface{}); ok {
			if field.Properties != nil {
				return sv.validateObject(obj, field.Properties, path)
			}
		} else {
			return fmt.Errorf("field '%s' must be an object", path)
		}
	}

	return nil
}

// isNumber checks if a value is a number (int, float, etc.)
func isNumber(value interface{}) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}

// GetVectorFields returns all vector fields in the schema with their dimensions
func (sv *SchemaValidator) GetVectorFields() map[string]int {
	fields := make(map[string]int)
	sv.collectVectorFields(sv.schema, "", fields)
	return fields
}

// collectVectorFields recursively collects vector fields from the schema
func (sv *SchemaValidator) collectVectorFields(schema Schema, prefix string, fields map[string]int) {
	for key, schemaValue := range schema {
		currentPath := key
		if prefix != "" {
			currentPath = prefix + "." + key
		}

		if field, ok := schemaValue.(SchemaField); ok {
			if field.Type == SchemaTypeVector {
				fields[currentPath] = field.Dimensions
			} else if field.Type == SchemaTypeObject && field.Properties != nil {
				sv.collectVectorFields(field.Properties, currentPath, fields)
			}
		} else if typeStr, ok := schemaValue.(string); ok {
			if strings.HasPrefix(typeStr, "vector[") && strings.HasSuffix(typeStr, "]") {
				dimStr := typeStr[7 : len(typeStr)-1]
				var dims int
				if _, err := fmt.Sscanf(dimStr, "%d", &dims); err == nil {
					fields[currentPath] = dims
				}
			}
		} else if nestedSchema, ok := schemaValue.(Schema); ok {
			sv.collectVectorFields(nestedSchema, currentPath, fields)
		}
	}
}

// GetSearchableFields returns all text-searchable fields in the schema
func (sv *SchemaValidator) GetSearchableFields() []string {
	var fields []string
	sv.collectSearchableFields(sv.schema, "", &fields)
	return fields
}

// collectSearchableFields recursively collects searchable fields from the schema
func (sv *SchemaValidator) collectSearchableFields(schema Schema, prefix string, fields *[]string) {
	for key, schemaValue := range schema {
		currentPath := key
		if prefix != "" {
			currentPath = prefix + "." + key
		}

		if field, ok := schemaValue.(SchemaField); ok {
			if field.Type == SchemaTypeString || field.Type == SchemaTypeStringArr {
				*fields = append(*fields, currentPath)
			} else if field.Type == SchemaTypeObject && field.Properties != nil {
				sv.collectSearchableFields(field.Properties, currentPath, fields)
			}
		} else if typeStr, ok := schemaValue.(string); ok {
			if typeStr == "string" || typeStr == "string[]" {
				*fields = append(*fields, currentPath)
			}
		} else if nestedSchema, ok := schemaValue.(Schema); ok {
			sv.collectSearchableFields(nestedSchema, currentPath, fields)
		}
	}
}

// ExtractVectors extracts all vector data from a document
func (sv *SchemaValidator) ExtractVectors(doc Document) map[string][]float32 {
	vectors := make(map[string][]float32)
	vectorFields := sv.GetVectorFields()

	for fieldPath := range vectorFields {
		if vector := sv.extractVectorFromPath(doc, fieldPath); vector != nil {
			vectors[fieldPath] = vector
		}
	}

	return vectors
}

// extractVectorFromPath extracts a vector from a specific field path in the document
func (sv *SchemaValidator) extractVectorFromPath(doc Document, path string) []float32 {
	parts := strings.Split(path, ".")
	current := map[string]interface{}(doc)

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - extract the vector
			if value, exists := current[part]; exists {
				return convertToFloat32Vector(value)
			}
		} else {
			// Navigate deeper into the object
			if next, ok := current[part].(map[string]interface{}); ok {
				current = next
			} else {
				return nil
			}
		}
	}

	return nil
}

// convertToFloat32Vector converts various vector formats to []float32
func convertToFloat32Vector(value interface{}) []float32 {
	switch v := value.(type) {
	case []float32:
		return v
	case []float64:
		result := make([]float32, len(v))
		for i, f := range v {
			result[i] = float32(f)
		}
		return result
	case []interface{}:
		result := make([]float32, len(v))
		for i, item := range v {
			if f, ok := item.(float64); ok {
				result[i] = float32(f)
			} else if f, ok := item.(float32); ok {
				result[i] = f
			} else if f, ok := item.(int); ok {
				result[i] = float32(f)
			} else {
				return nil // Invalid vector element
			}
		}
		return result
	default:
		return nil
	}
}

// ExtractSearchableText extracts all searchable text from a document
func (sv *SchemaValidator) ExtractSearchableText(doc Document) map[string]string {
	text := make(map[string]string)
	searchableFields := sv.GetSearchableFields()

	for _, fieldPath := range searchableFields {
		if textValue := sv.extractTextFromPath(doc, fieldPath); textValue != "" {
			text[fieldPath] = textValue
		}
	}

	return text
}

// extractTextFromPath extracts text from a specific field path in the document
func (sv *SchemaValidator) extractTextFromPath(doc Document, path string) string {
	parts := strings.Split(path, ".")
	current := map[string]interface{}(doc)

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - extract the text
			if value, exists := current[part]; exists {
				if str, ok := value.(string); ok {
					return str
				} else if arr, ok := value.([]interface{}); ok {
					// Join string array
					var strs []string
					for _, item := range arr {
						if s, ok := item.(string); ok {
							strs = append(strs, s)
						}
					}
					return strings.Join(strs, " ")
				}
			}
		} else {
			// Navigate deeper into the object
			if next, ok := current[part].(map[string]interface{}); ok {
				current = next
			} else {
				return ""
			}
		}
	}

	return ""
}
