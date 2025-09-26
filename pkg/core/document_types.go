package core

import (
	"strings"
	"time"
)

// SearchMode represents different search modes
type SearchMode string

const (
	SearchModeFullText SearchMode = "fulltext"
	SearchModeVector   SearchMode = "vector"
	SearchModeHybrid   SearchMode = "hybrid"
)

// SearchParams represents parameters for document search
type SearchParams struct {
	// Search mode
	Mode SearchMode `json:"mode"`

	// Full-text search parameters
	Term       string   `json:"term,omitempty"`
	Properties []string `json:"properties,omitempty"`

	// Vector search parameters
	Vector *VectorSearchParams `json:"vector,omitempty"`

	// Common parameters
	Limit      int                     `json:"limit"`
	Offset     int                     `json:"offset"`
	Where      map[string]interface{}  `json:"where,omitempty"`
	Facets     map[string]*FacetConfig `json:"facets,omitempty"`
	SortBy     *SortConfig             `json:"sort_by,omitempty"`
	GroupBy    *GroupConfig            `json:"group_by,omitempty"`
	Threshold  float64                 `json:"threshold,omitempty"`
	Similarity float64                 `json:"similarity,omitempty"`

	// Hybrid search parameters
	HybridWeights *HybridWeights `json:"hybrid_weights,omitempty"`

	// Response options
	IncludeVectors bool   `json:"include_vectors,omitempty"`
	Distinct       string `json:"distinct,omitempty"`

	// Advanced parameters
	Boost     map[string]float64 `json:"boost,omitempty"`
	Relevance *BM25Params        `json:"relevance,omitempty"`
	Exact     bool               `json:"exact,omitempty"`
	Tolerance int                `json:"tolerance,omitempty"`
}

// VectorSearchParams represents vector search parameters
type VectorSearchParams struct {
	Value    []float32 `json:"value"`
	Property string    `json:"property"`
}

// HybridWeights represents weights for hybrid search
type HybridWeights struct {
	Text   float64 `json:"text"`
	Vector float64 `json:"vector"`
}

// BM25Params represents BM25 scoring parameters
type BM25Params struct {
	K float64 `json:"k"` // Term frequency saturation parameter
	B float64 `json:"b"` // Document length normalization factor
	D float64 `json:"d"` // Frequency normalization lower bound
}

// FacetConfig represents facet configuration
type FacetConfig struct {
	Type   FacetType `json:"type"`
	Limit  int       `json:"limit,omitempty"`
	Offset int       `json:"offset,omitempty"`
	Sort   string    `json:"sort,omitempty"`   // "asc" or "desc"
	Ranges []Range   `json:"ranges,omitempty"` // For number facets
}

// FacetType represents different facet types
type FacetType string

const (
	FacetTypeString  FacetType = "string"
	FacetTypeNumber  FacetType = "number"
	FacetTypeBoolean FacetType = "boolean"
)

// Range represents a numeric range for facets
type Range struct {
	From float64 `json:"from"`
	To   float64 `json:"to"`
}

// SortConfig represents sorting configuration
type SortConfig struct {
	Property string `json:"property"`
	Order    string `json:"order"` // "asc" or "desc"
}

// GroupConfig represents grouping configuration
type GroupConfig struct {
	Properties []string `json:"properties"`
	MaxResult  int      `json:"max_result,omitempty"`
}

// DocumentSearchResponse represents the response from document search
type DocumentSearchResponse struct {
	Hits    []*DocumentSearchResult `json:"hits"`
	Count   int                     `json:"count"`
	Elapsed time.Duration           `json:"elapsed"`
	Facets  map[string]*FacetResult `json:"facets,omitempty"`
	Groups  []*GroupResult          `json:"groups,omitempty"`
}

// DocumentSearchResult represents a single search result
type DocumentSearchResult struct {
	ID       string   `json:"id"`
	Score    float32  `json:"score"`
	Document Document `json:"document"`
}

// FacetResult represents facet results
type FacetResult struct {
	Count  int            `json:"count"`
	Values map[string]int `json:"values"`
}

// GroupResult represents group results
type GroupResult struct {
	Values []interface{}           `json:"values"`
	Hits   []*DocumentSearchResult `json:"hits"`
}

// TextTokenizer handles text tokenization and processing
type TextTokenizer struct {
	Language      string
	Stemming      bool
	StopWords     []string
	CaseSensitive bool
}

// NewTextTokenizer creates a new text tokenizer
func NewTextTokenizer() *TextTokenizer {
	return &TextTokenizer{
		Language:      "english",
		Stemming:      false,
		StopWords:     DefaultStopWords(),
		CaseSensitive: false,
	}
}

// Tokenize tokenizes text into terms
func (tt *TextTokenizer) Tokenize(text string) []string {
	if !tt.CaseSensitive {
		text = strings.ToLower(text)
	}

	// Simple tokenization (split by whitespace and punctuation)
	tokens := strings.FieldsFunc(text, func(c rune) bool {
		return !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'))
	})

	// Remove stop words
	var filtered []string
	stopWordSet := make(map[string]bool)
	for _, word := range tt.StopWords {
		stopWordSet[word] = true
	}

	for _, token := range tokens {
		if len(token) > 0 && !stopWordSet[token] {
			if tt.Stemming {
				token = tt.stem(token)
			}
			filtered = append(filtered, token)
		}
	}

	return filtered
}

// stem applies basic stemming (simplified Porter stemmer)
func (tt *TextTokenizer) stem(word string) string {
	// Very basic stemming - remove common suffixes
	suffixes := []string{"ing", "ed", "er", "est", "ly", "s"}

	for _, suffix := range suffixes {
		if strings.HasSuffix(word, suffix) && len(word) > len(suffix)+2 {
			return word[:len(word)-len(suffix)]
		}
	}

	return word
}

// DefaultStopWords returns default English stop words
func DefaultStopWords() []string {
	return []string{
		"a", "an", "and", "are", "as", "at", "be", "by", "for", "from",
		"has", "he", "in", "is", "it", "its", "of", "on", "that", "the",
		"to", "was", "will", "with", "the", "this", "but", "they", "have",
		"had", "what", "said", "each", "which", "she", "do", "how", "their",
		"if", "up", "out", "many", "then", "them", "these", "so", "some",
		"her", "would", "make", "like", "into", "him", "time", "two", "more",
		"go", "no", "way", "could", "my", "than", "first", "been", "call",
		"who", "oil", "sit", "now", "find", "down", "day", "did", "get",
		"come", "made", "may", "part",
	}
}

// FullTextSearchConfig represents full-text search configuration
type FullTextSearchConfig struct {
	Language      string             `json:"language"`
	Stemming      bool               `json:"stemming"`
	StopWords     []string           `json:"stop_words"`
	CaseSensitive bool               `json:"case_sensitive"`
	BM25          *BM25Params        `json:"bm25"`
	Boost         map[string]float64 `json:"boost"`
}

// DefaultFullTextSearchConfig returns default full-text search configuration
func DefaultFullTextSearchConfig() *FullTextSearchConfig {
	return &FullTextSearchConfig{
		Language:      "english",
		Stemming:      true,
		StopWords:     DefaultStopWords(),
		CaseSensitive: false,
		BM25: &BM25Params{
			K: 1.2,
			B: 0.75,
			D: 0.5,
		},
		Boost: make(map[string]float64),
	}
}

// CreateRequest represents a request to create a unified database
type CreateRequest struct {
	Schema            Schema                 `json:"schema"`
	Language          string                 `json:"language,omitempty"`
	FullTextConfig    *FullTextSearchConfig  `json:"fulltext_config,omitempty"`
	VectorizerConfigs map[string]interface{} `json:"vectorizer_configs,omitempty"`
	ContentStorage    *ContentStorageConfig  `json:"content_storage,omitempty"`
}

// InsertRequest represents a document insertion request
type InsertRequest struct {
	Document Document       `json:"document"`
	Options  *InsertOptions `json:"options,omitempty"`
}

// InsertOptions represents options for document insertion
type InsertOptions struct {
	SkipValidation bool `json:"skip_validation,omitempty"`
	Upsert         bool `json:"upsert,omitempty"`
}

// UpdateRequest represents a document update request
type UpdateRequest struct {
	ID       string         `json:"id"`
	Document Document       `json:"document"`
	Options  *UpdateOptions `json:"options,omitempty"`
}

// UpdateOptions represents options for document updates
type UpdateOptions struct {
	SkipValidation bool `json:"skip_validation,omitempty"`
	Partial        bool `json:"partial,omitempty"`
}

// DeleteRequest represents a document deletion request
type DeleteRequest struct {
	ID string `json:"id"`
}

// GetRequest represents a document retrieval request
type GetRequest struct {
	ID             string `json:"id"`
	IncludeVectors bool   `json:"include_vectors,omitempty"`
}

// CountRequest represents a document count request
type CountRequest struct {
	Where map[string]interface{} `json:"where,omitempty"`
}

// CountResponse represents a document count response
type CountResponse struct {
	Count int `json:"count"`
}

// GetResponse represents a document retrieval response
type GetResponse struct {
	Document Document `json:"document,omitempty"`
	Found    bool     `json:"found"`
}

// InsertResponse represents a document insertion response
type InsertResponse struct {
	ID      string `json:"id"`
	Created bool   `json:"created"`
}

// UpdateResponse represents a document update response
type UpdateResponse struct {
	ID      string `json:"id"`
	Updated bool   `json:"updated"`
}

// DeleteResponse represents a document deletion response
type DeleteResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}
