package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaVectorizer implements text vectorization using local Ollama models
// This provides real ML embeddings without external API dependencies
type OllamaVectorizer struct {
	model      string
	dimensions int
	config     *VectorizerConfig
	client     *http.Client
	baseURL    string
}

// NewOllamaVectorizer creates a new Ollama vectorizer
func NewOllamaVectorizer(config *VectorizerConfig) (*OllamaVectorizer, error) {
	if config.Model == "" {
		config.Model = "nomic-embed-text" // Default embedding model for Ollama
	}

	dimensions := config.Dimensions
	if dimensions == 0 {
		dimensions = 768 // Default for nomic-embed-text
	}

	// Get Ollama base URL from config or use default
	baseURL := "http://localhost:11434"
	if url, exists := config.Options["base_url"]; exists {
		if urlStr, ok := url.(string); ok {
			baseURL = urlStr
		}
	}

	return &OllamaVectorizer{
		model:      config.Model,
		dimensions: dimensions,
		config:     config,
		client: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for local model inference
		},
		baseURL: baseURL,
	}, nil
}

// GenerateEmbedding generates a single embedding using Ollama
func (v *OllamaVectorizer) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := v.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}
	return embeddings[0], nil
}

// GenerateEmbeddings generates multiple embeddings using Ollama
func (v *OllamaVectorizer) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := v.callOllamaAPI(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// OllamaEmbeddingRequest represents the request format for Ollama embeddings API
type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaEmbeddingResponse represents the response format from Ollama embeddings API
type OllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// callOllamaAPI makes API call to local Ollama service
func (v *OllamaVectorizer) callOllamaAPI(ctx context.Context, text string) ([]float32, error) {
	// Prepare request
	requestBody := OllamaEmbeddingRequest{
		Model:  v.model,
		Prompt: text,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/embeddings", v.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama (is it running?): %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response OllamaEmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("no embeddings returned from Ollama")
	}

	// Convert from float64 to float32
	embedding := make([]float32, len(response.Embedding))
	for i, val := range response.Embedding {
		embedding[i] = float32(val)
	}

	return embedding, nil
}

// Interface compliance methods
func (v *OllamaVectorizer) GetDimensions() int {
	return v.dimensions
}

func (v *OllamaVectorizer) GetModel() string {
	return v.model
}

func (v *OllamaVectorizer) Close() error {
	return nil
}

func (v *OllamaVectorizer) Dimensions() int {
	return v.GetDimensions()
}

func (v *OllamaVectorizer) ModelName() string {
	return v.GetModel()
}

func (v *OllamaVectorizer) Type() VectorizerType {
	return VectorizerTypeOllama
}
