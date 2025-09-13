package embeddings

import (
	"context"
)

// VectorizerType represents different embedding model types
type VectorizerType int

const (
	VectorizerTypeNone VectorizerType = iota
	VectorizerTypeSentenceTransformers
	VectorizerTypeOpenAI
	VectorizerTypeHuggingFace
	VectorizerTypeOllama
)

func (v VectorizerType) String() string {
	switch v {
	case VectorizerTypeNone:
		return "none"
	case VectorizerTypeSentenceTransformers:
		return "sentence_transformers"
	case VectorizerTypeOpenAI:
		return "openai"
	case VectorizerTypeHuggingFace:
		return "huggingface"
	case VectorizerTypeOllama:
		return "ollama"
	default:
		return "unknown"
	}
}

// VectorizerConfig represents the configuration for text vectorization
type VectorizerConfig struct {
	Type       VectorizerType         `json:"type" yaml:"type"`
	Model      string                 `json:"model" yaml:"model"`
	Dimensions int                    `json:"dimensions" yaml:"dimensions"`
	Options    map[string]interface{} `json:"options" yaml:"options"`
}

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Texts  []string          `json:"texts"`
	Model  string            `json:"model,omitempty"`
	Config *VectorizerConfig `json:"config,omitempty"`
}

// EmbeddingResponse represents the response from embedding generation
type EmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
	Model      string      `json:"model"`
	Dimensions int         `json:"dimensions"`
}

// Vectorizer interface for generating embeddings from text
type Vectorizer interface {
	// GenerateEmbedding generates a single embedding from text
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// GenerateEmbeddings generates multiple embeddings from texts
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)

	// GetDimensions returns the embedding dimensions
	GetDimensions() int

	// GetModel returns the model name
	GetModel() string

	// Close cleans up resources
	Close() error
}

// VectorizerFactory creates vectorizers based on configuration
type VectorizerFactory interface {
	CreateVectorizer(config *VectorizerConfig) (Vectorizer, error)
	SupportedTypes() []VectorizerType
}
