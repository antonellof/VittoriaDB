package embeddings

import (
	"context"
	"fmt"
)

// HuggingFaceVectorizer implements the Vectorizer interface using HuggingFace models
type HuggingFaceVectorizer struct {
	model      string
	dimensions int
	apiKey     string
	config     *VectorizerConfig
}

// NewHuggingFaceVectorizer creates a new HuggingFace vectorizer
func NewHuggingFaceVectorizer(config *VectorizerConfig) (*HuggingFaceVectorizer, error) {
	if config.Model == "" {
		config.Model = "sentence-transformers/all-MiniLM-L6-v2"
	}

	apiKey, _ := config.Options["api_key"].(string) // Optional for HuggingFace

	dimensions := config.Dimensions
	if dimensions == 0 {
		// Set default dimensions based on model
		if config.Model == "sentence-transformers/all-MiniLM-L6-v2" {
			dimensions = 384
		} else {
			dimensions = 384 // Default fallback
		}
	}

	return &HuggingFaceVectorizer{
		model:      config.Model,
		dimensions: dimensions,
		apiKey:     apiKey,
		config:     config,
	}, nil
}

// GenerateEmbedding generates a single embedding from text
func (v *HuggingFaceVectorizer) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := v.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}
	return embeddings[0], nil
}

// GenerateEmbeddings generates multiple embeddings from texts
func (v *HuggingFaceVectorizer) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	// TODO: Implement HuggingFace API integration
	// For now, return an error indicating this is not implemented
	return nil, fmt.Errorf("HuggingFace vectorizer not yet implemented - please use sentence_transformers for now")
}

// GetDimensions returns the embedding dimensions
func (v *HuggingFaceVectorizer) GetDimensions() int {
	return v.dimensions
}

// GetModel returns the model name
func (v *HuggingFaceVectorizer) GetModel() string {
	return v.model
}

// Close cleans up resources
func (v *HuggingFaceVectorizer) Close() error {
	return nil
}
