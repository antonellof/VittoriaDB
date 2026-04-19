package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HuggingFaceVectorizer implements the Vectorizer interface using the
// HuggingFace Inference API (https://huggingface.co/docs/api-inference).
//
// It uses the public `feature-extraction` pipeline endpoint and works with any
// model that exposes sentence/document embeddings (e.g. sentence-transformers).
// An API token is optional for low-volume free-tier usage but strongly
// recommended.
type HuggingFaceVectorizer struct {
	model      string
	dimensions int
	apiKey     string
	endpoint   string
	httpClient *http.Client
	config     *VectorizerConfig
}

// hfDefaultDimensions maps a few well-known sentence-transformers models to
// their output dimensionality. Anything not in this map falls back to a
// probing call on the first GenerateEmbedding invocation.
var hfDefaultDimensions = map[string]int{
	"sentence-transformers/all-MiniLM-L6-v2":                    384,
	"sentence-transformers/all-MiniLM-L12-v2":                   384,
	"sentence-transformers/all-mpnet-base-v2":                   768,
	"sentence-transformers/distilbert-base-nli-stsb-mean-tokens": 768,
	"sentence-transformers/paraphrase-MiniLM-L6-v2":             384,
}

// NewHuggingFaceVectorizer creates a new HuggingFace vectorizer.
func NewHuggingFaceVectorizer(config *VectorizerConfig) (*HuggingFaceVectorizer, error) {
	if config.Model == "" {
		config.Model = "sentence-transformers/all-MiniLM-L6-v2"
	}

	apiKey, _ := config.Options["api_key"].(string)

	endpoint, _ := config.Options["endpoint"].(string)
	if endpoint == "" {
		endpoint = "https://api-inference.huggingface.co/pipeline/feature-extraction/" + config.Model
	}

	timeoutSec := 30
	if v, ok := config.Options["timeout"].(int); ok && v > 0 {
		timeoutSec = v
	}

	dimensions := config.Dimensions
	if dimensions == 0 {
		if d, ok := hfDefaultDimensions[config.Model]; ok {
			dimensions = d
		}
	}

	return &HuggingFaceVectorizer{
		model:      config.Model,
		dimensions: dimensions,
		apiKey:     apiKey,
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: time.Duration(timeoutSec) * time.Second},
		config:     config,
	}, nil
}

// GenerateEmbedding generates a single embedding from text.
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

// GenerateEmbeddings generates multiple embeddings via the HF Inference API.
func (v *HuggingFaceVectorizer) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	payload := map[string]interface{}{
		"inputs": texts,
		"options": map[string]interface{}{
			"wait_for_model": true,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if v.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+v.apiKey)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call HuggingFace inference API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read HuggingFace response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HuggingFace inference API error (status %d): %s",
			resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	embeddings, err := parseHFEmbeddings(respBody, len(texts))
	if err != nil {
		return nil, err
	}

	if v.dimensions == 0 && len(embeddings) > 0 {
		v.dimensions = len(embeddings[0])
	}

	return embeddings, nil
}

// parseHFEmbeddings normalizes the variety of shapes returned by the HF
// feature-extraction pipeline:
//   - [[f, f, ...]]                : single text, sentence embedding
//   - [[[f, ...], [f, ...]], ...]  : token-level embeddings (mean-pooled here)
//   - [[f, ...], [f, ...]]         : batched sentence embeddings
func parseHFEmbeddings(body []byte, expected int) ([][]float32, error) {
	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode HuggingFace response: %w", err)
	}

	arr, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected HuggingFace response shape: %s", string(body))
	}

	out := make([][]float32, 0, expected)
	for _, item := range arr {
		switch v := item.(type) {
		case []interface{}:
			if len(v) == 0 {
				out = append(out, nil)
				continue
			}
			if _, isFloatRow := v[0].(float64); isFloatRow {
				vec := make([]float32, len(v))
				for i, f := range v {
					vec[i] = float32(toFloat64(f))
				}
				out = append(out, vec)
			} else if tokens, isTokenMatrix := v[0].([]interface{}); isTokenMatrix {
				dim := len(tokens)
				sum := make([]float64, dim)
				for _, tok := range v {
					row, _ := tok.([]interface{})
					for i := 0; i < dim && i < len(row); i++ {
						sum[i] += toFloat64(row[i])
					}
				}
				vec := make([]float32, dim)
				count := float64(len(v))
				if count == 0 {
					count = 1
				}
				for i, s := range sum {
					vec[i] = float32(s / count)
				}
				out = append(out, vec)
			} else {
				return nil, fmt.Errorf("unexpected nested HuggingFace embedding row")
			}
		case float64:
			return nil, fmt.Errorf("unexpected scalar in HuggingFace embedding response")
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("HuggingFace returned no embeddings")
	}
	return out, nil
}

func toFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return 0
}

// GetDimensions returns the embedding dimensions.
func (v *HuggingFaceVectorizer) GetDimensions() int {
	return v.dimensions
}

// GetModel returns the model name.
func (v *HuggingFaceVectorizer) GetModel() string {
	return v.model
}

// Close cleans up resources.
func (v *HuggingFaceVectorizer) Close() error {
	return nil
}
