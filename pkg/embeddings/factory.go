package embeddings

import (
	"fmt"
)

// DefaultVectorizerFactory implements VectorizerFactory
type DefaultVectorizerFactory struct{}

// NewVectorizerFactory creates a new vectorizer factory
func NewVectorizerFactory() VectorizerFactory {
	return &DefaultVectorizerFactory{}
}

// CreateVectorizer creates a vectorizer based on configuration
func (f *DefaultVectorizerFactory) CreateVectorizer(config *VectorizerConfig) (Vectorizer, error) {
	if config == nil {
		return nil, fmt.Errorf("vectorizer config cannot be nil")
	}

	switch config.Type {
	case VectorizerTypeNone:
		return nil, fmt.Errorf("vectorizer type 'none' does not support automatic embedding generation")

	case VectorizerTypeSentenceTransformers:
		return NewSentenceTransformersVectorizer(config)

	case VectorizerTypeOpenAI:
		return NewOpenAIVectorizer(config)

	case VectorizerTypeHuggingFace:
		return NewHuggingFaceVectorizer(config)

	case VectorizerTypeOllama:
		return NewOllamaVectorizer(config)

	default:
		return nil, fmt.Errorf("unsupported vectorizer type: %s", config.Type.String())
	}
}

// SupportedTypes returns the list of supported vectorizer types
func (f *DefaultVectorizerFactory) SupportedTypes() []VectorizerType {
	return []VectorizerType{
		VectorizerTypeSentenceTransformers,
		VectorizerTypeOpenAI,
		VectorizerTypeHuggingFace,
		VectorizerTypeOllama,
	}
}

// GetDefaultConfig returns default configuration for a vectorizer type
func GetDefaultConfig(vectorizerType VectorizerType) *VectorizerConfig {
	switch vectorizerType {
	case VectorizerTypeSentenceTransformers:
		return &VectorizerConfig{
			Type:       VectorizerTypeSentenceTransformers,
			Model:      "all-MiniLM-L6-v2",
			Dimensions: 384,
			Options:    make(map[string]interface{}),
		}

	case VectorizerTypeOpenAI:
		return &VectorizerConfig{
			Type:       VectorizerTypeOpenAI,
			Model:      "text-embedding-ada-002",
			Dimensions: 1536,
			Options: map[string]interface{}{
				"api_key": "", // Must be provided by user
			},
		}

	case VectorizerTypeHuggingFace:
		return &VectorizerConfig{
			Type:       VectorizerTypeHuggingFace,
			Model:      "sentence-transformers/all-MiniLM-L6-v2",
			Dimensions: 384,
			Options:    make(map[string]interface{}),
		}

	case VectorizerTypeOllama:
		return &VectorizerConfig{
			Type:       VectorizerTypeOllama,
			Model:      "nomic-embed-text",
			Dimensions: 768,
			Options: map[string]interface{}{
				"base_url": "http://localhost:11434",
			},
		}

	default:
		return &VectorizerConfig{
			Type:       VectorizerTypeNone,
			Model:      "",
			Dimensions: 0,
			Options:    make(map[string]interface{}),
		}
	}
}

// ValidateConfig validates a vectorizer configuration
func ValidateConfig(config *VectorizerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Dimensions <= 0 {
		return fmt.Errorf("dimensions must be positive, got %d", config.Dimensions)
	}

	if config.Dimensions > 10000 {
		return fmt.Errorf("dimensions cannot exceed 10000, got %d", config.Dimensions)
	}

	switch config.Type {
	case VectorizerTypeOpenAI:
		if apiKey, ok := config.Options["api_key"].(string); !ok || apiKey == "" {
			return fmt.Errorf("OpenAI vectorizer requires 'api_key' in options")
		}

	case VectorizerTypeHuggingFace:
		// Optional API key for HuggingFace
		if apiKey, ok := config.Options["api_key"].(string); ok && apiKey == "" {
			return fmt.Errorf("HuggingFace API key cannot be empty if provided")
		}
	}

	return nil
}
