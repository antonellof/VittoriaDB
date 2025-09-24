package embeddings

import (
	"context"
	"fmt"
	"log"
)

// EnhancedVectorizer wraps any Vectorizer with batch processing capabilities
type EnhancedVectorizer struct {
	baseVectorizer Vectorizer
	batchProcessor *BatchProcessor
	config         *VectorizerConfig
}

// NewEnhancedVectorizer creates a new enhanced vectorizer with batch processing
func NewEnhancedVectorizer(baseVectorizer Vectorizer, config *VectorizerConfig) *EnhancedVectorizer {
	batchConfig := DefaultBatchProcessorConfig()
	
	// Override batch config from vectorizer options if provided
	if config != nil && config.Options != nil {
		if batchSize, ok := config.Options["batch_size"].(int); ok {
			batchConfig.BatchSize = batchSize
		}
		if fallbackSize, ok := config.Options["fallback_batch_size"].(int); ok {
			batchConfig.FallbackSize = fallbackSize
		}
		if maxWorkers, ok := config.Options["max_workers"].(int); ok {
			batchConfig.MaxWorkers = maxWorkers
		}
		if enableFallback, ok := config.Options["enable_fallback"].(bool); ok {
			batchConfig.EnableFallback = enableFallback
		}
	}

	batchProcessor := NewBatchProcessor(baseVectorizer, batchConfig)

	return &EnhancedVectorizer{
		baseVectorizer: baseVectorizer,
		batchProcessor: batchProcessor,
		config:         config,
	}
}

// GenerateEmbedding generates a single embedding from text
func (ev *EnhancedVectorizer) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// For single embeddings, use the base vectorizer directly
	return ev.baseVectorizer.GenerateEmbedding(ctx, text)
}

// GenerateEmbeddings generates multiple embeddings with enhanced batch processing
func (ev *EnhancedVectorizer) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Use batch processor for multiple texts
	embeddings, err := ev.batchProcessor.ProcessTexts(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("enhanced batch processing failed: %w", err)
	}

	// Log statistics for monitoring
	stats := ev.batchProcessor.GetStats()
	if stats.FallbacksUsed > 0 || stats.RetriesUsed > 0 {
		log.Printf("Batch processing stats: successful=%d, failed=%d, fallbacks=%d, retries=%d, throughput=%.2f/sec",
			stats.SuccessfulTexts, stats.FailedTexts, stats.FallbacksUsed, stats.RetriesUsed, stats.ThroughputPerSec)
	}

	return embeddings, nil
}

// GenerateEmbeddingsWithStats generates embeddings and returns processing statistics
func (ev *EnhancedVectorizer) GenerateEmbeddingsWithStats(ctx context.Context, texts []string) ([][]float32, *BatchProcessorStats, error) {
	embeddings, err := ev.GenerateEmbeddings(ctx, texts)
	stats := ev.batchProcessor.GetStats()
	return embeddings, &stats, err
}

// GetDimensions returns the embedding dimensions
func (ev *EnhancedVectorizer) GetDimensions() int {
	return ev.baseVectorizer.GetDimensions()
}

// GetModel returns the model name
func (ev *EnhancedVectorizer) GetModel() string {
	return ev.baseVectorizer.GetModel()
}

// Close cleans up resources
func (ev *EnhancedVectorizer) Close() error {
	return ev.baseVectorizer.Close()
}

// GetBatchProcessor returns the underlying batch processor for advanced configuration
func (ev *EnhancedVectorizer) GetBatchProcessor() *BatchProcessor {
	return ev.batchProcessor
}

// GetStats returns the current batch processing statistics
func (ev *EnhancedVectorizer) GetStats() BatchProcessorStats {
	return ev.batchProcessor.GetStats()
}

// UpdateBatchConfig updates the batch processing configuration
func (ev *EnhancedVectorizer) UpdateBatchConfig(config *BatchProcessorConfig) {
	ev.batchProcessor.UpdateConfig(config)
}
