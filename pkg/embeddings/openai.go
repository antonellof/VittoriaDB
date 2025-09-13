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

// OpenAIVectorizer implements the Vectorizer interface using OpenAI embeddings
type OpenAIVectorizer struct {
	model      string
	dimensions int
	apiKey     string
	config     *VectorizerConfig
}

// NewOpenAIVectorizer creates a new OpenAI vectorizer
func NewOpenAIVectorizer(config *VectorizerConfig) (*OpenAIVectorizer, error) {
	if config.Model == "" {
		config.Model = "text-embedding-ada-002"
	}

	apiKey, ok := config.Options["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	dimensions := config.Dimensions
	if dimensions == 0 {
		// Set default dimensions based on model
		switch config.Model {
		case "text-embedding-ada-002":
			dimensions = 1536
		case "text-embedding-3-small":
			dimensions = 1536
		case "text-embedding-3-large":
			dimensions = 3072
		default:
			dimensions = 1536
		}
	}

	return &OpenAIVectorizer{
		model:      config.Model,
		dimensions: dimensions,
		apiKey:     apiKey,
		config:     config,
	}, nil
}

// GenerateEmbedding generates a single embedding from text
func (v *OpenAIVectorizer) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
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
func (v *OpenAIVectorizer) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Create HTTP request to OpenAI API
	requestBody := map[string]interface{}{
		"input": texts,
		"model": v.model,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+v.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	embeddings := make([][]float32, len(response.Data))
	for i, data := range response.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}

// GetDimensions returns the embedding dimensions
func (v *OpenAIVectorizer) GetDimensions() int {
	return v.dimensions
}

// GetModel returns the model name
func (v *OpenAIVectorizer) GetModel() string {
	return v.model
}

// Close cleans up resources
func (v *OpenAIVectorizer) Close() error {
	return nil
}
