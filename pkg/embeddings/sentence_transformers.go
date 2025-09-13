package embeddings

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// SentenceTransformersVectorizer implements the Vectorizer interface using sentence-transformers
type SentenceTransformersVectorizer struct {
	model      string
	dimensions int
	config     *VectorizerConfig
}

// NewSentenceTransformersVectorizer creates a new sentence-transformers vectorizer
func NewSentenceTransformersVectorizer(config *VectorizerConfig) (*SentenceTransformersVectorizer, error) {
	if config.Model == "" {
		config.Model = "all-MiniLM-L6-v2" // Default model
	}

	// Get dimensions for the model
	dimensions := config.Dimensions
	if dimensions == 0 {
		// Set default dimensions based on common models
		switch config.Model {
		case "all-MiniLM-L6-v2":
			dimensions = 384
		case "all-mpnet-base-v2":
			dimensions = 768
		case "paraphrase-multilingual-MiniLM-L12-v2":
			dimensions = 384
		default:
			dimensions = 384 // Default fallback
		}
	}

	return &SentenceTransformersVectorizer{
		model:      config.Model,
		dimensions: dimensions,
		config:     config,
	}, nil
}

// GenerateEmbedding generates a single embedding from text
func (v *SentenceTransformersVectorizer) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
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
func (v *SentenceTransformersVectorizer) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Create Python script to generate embeddings
	pythonScript := v.createPythonScript(texts)

	// Execute Python script
	cmd := exec.CommandContext(ctx, "python3", "-c", pythonScript)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute embedding script: %w", err)
	}

	// Parse JSON output
	var result struct {
		Embeddings [][]float32 `json:"embeddings"`
		Error      string      `json:"error,omitempty"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse embedding output: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("embedding generation error: %s", result.Error)
	}

	return result.Embeddings, nil
}

// createPythonScript creates a Python script to generate embeddings
func (v *SentenceTransformersVectorizer) createPythonScript(texts []string) string {
	// Escape texts for JSON
	textsJSON, _ := json.Marshal(texts)

	script := fmt.Sprintf(`
import json
import sys
try:
    from sentence_transformers import SentenceTransformer
    
    # Load model
    model = SentenceTransformer('%s')
    
    # Input texts
    texts = %s
    
    # Generate embeddings
    embeddings = model.encode(texts)
    
    # Convert to list of lists
    embeddings_list = [embedding.tolist() for embedding in embeddings]
    
    # Output JSON
    result = {"embeddings": embeddings_list}
    print(json.dumps(result))
    
except ImportError as e:
    error_result = {"error": "sentence-transformers not installed. Please install with: pip install sentence-transformers"}
    print(json.dumps(error_result))
    sys.exit(1)
except Exception as e:
    error_result = {"error": str(e)}
    print(json.dumps(error_result))
    sys.exit(1)
`, v.model, string(textsJSON))

	return script
}

// GetDimensions returns the embedding dimensions
func (v *SentenceTransformersVectorizer) GetDimensions() int {
	return v.dimensions
}

// GetModel returns the model name
func (v *SentenceTransformersVectorizer) GetModel() string {
	return v.model
}

// Close cleans up resources
func (v *SentenceTransformersVectorizer) Close() error {
	// No resources to clean up for sentence-transformers
	return nil
}
