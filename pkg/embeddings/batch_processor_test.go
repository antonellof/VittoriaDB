package embeddings

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockVectorizer for testing
type MockVectorizer struct {
	model      string
	dimensions int
	failCount  int
	callCount  int
}

func NewMockVectorizer(model string, dimensions int) *MockVectorizer {
	return &MockVectorizer{
		model:      model,
		dimensions: dimensions,
	}
}

func (m *MockVectorizer) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	m.callCount++
	
	// Simulate failure for testing
	if m.failCount > 0 {
		m.failCount--
		return nil, fmt.Errorf("mock failure")
	}

	// Return mock embedding
	embedding := make([]float32, m.dimensions)
	for i := range embedding {
		embedding[i] = float32(i) * 0.1
	}
	return embedding, nil
}

func (m *MockVectorizer) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	m.callCount++
	
	// Simulate failure for testing
	if m.failCount > 0 {
		m.failCount--
		return nil, fmt.Errorf("mock batch failure")
	}

	// Return mock embeddings
	embeddings := make([][]float32, len(texts))
	for i := range embeddings {
		embeddings[i] = make([]float32, m.dimensions)
		for j := range embeddings[i] {
			embeddings[i][j] = float32(i*m.dimensions+j) * 0.1
		}
	}
	return embeddings, nil
}

func (m *MockVectorizer) GetDimensions() int {
	return m.dimensions
}

func (m *MockVectorizer) GetModel() string {
	return m.model
}

func (m *MockVectorizer) Close() error {
	return nil
}

func (m *MockVectorizer) SetFailCount(count int) {
	m.failCount = count
}

func (m *MockVectorizer) GetCallCount() int {
	return m.callCount
}

func TestBatchProcessor_SuccessfulProcessing(t *testing.T) {
	mockVectorizer := NewMockVectorizer("test-model", 384)
	config := DefaultBatchProcessorConfig()
	config.BatchSize = 4
	
	processor := NewBatchProcessor(mockVectorizer, config)
	
	texts := []string{"text1", "text2", "text3", "text4", "text5"}
	
	ctx := context.Background()
	embeddings, err := processor.ProcessTexts(ctx, texts)
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(embeddings) != len(texts) {
		t.Fatalf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}
	
	stats := processor.GetStats()
	if stats.SuccessfulTexts != len(texts) {
		t.Errorf("Expected %d successful texts, got %d", len(texts), stats.SuccessfulTexts)
	}
	
	if stats.FailedTexts != 0 {
		t.Errorf("Expected 0 failed texts, got %d", stats.FailedTexts)
	}
}

func TestBatchProcessor_FallbackProcessing(t *testing.T) {
	mockVectorizer := NewMockVectorizer("test-model", 384)
	config := DefaultBatchProcessorConfig()
	config.BatchSize = 10
	config.FallbackSize = 2
	config.MaxRetries = 1
	
	processor := NewBatchProcessor(mockVectorizer, config)
	
	// Set mock to fail on first call (full batch), succeed on smaller batches
	mockVectorizer.SetFailCount(1)
	
	texts := []string{"text1", "text2", "text3", "text4"}
	
	ctx := context.Background()
	embeddings, err := processor.ProcessTexts(ctx, texts)
	
	if err != nil {
		t.Fatalf("Expected no error with fallback, got: %v", err)
	}
	
	if len(embeddings) != len(texts) {
		t.Fatalf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}
	
	stats := processor.GetStats()
	if stats.FallbacksUsed == 0 {
		t.Error("Expected fallback to be used")
	}
}

func TestBatchProcessor_IndividualFallback(t *testing.T) {
	mockVectorizer := NewMockVectorizer("test-model", 384)
	config := DefaultBatchProcessorConfig()
	config.BatchSize = 4
	config.FallbackSize = 2
	config.MaxRetries = 1
	
	processor := NewBatchProcessor(mockVectorizer, config)
	
	// Set mock to fail on batch calls (2 attempts: full batch + fallback batch)
	// but succeed on individual calls
	mockVectorizer.SetFailCount(2) // Fail batch attempts only
	
	texts := []string{"text1", "text2"}
	
	ctx := context.Background()
	embeddings, err := processor.ProcessTexts(ctx, texts)
	
	// Should eventually succeed with individual processing
	if err != nil {
		t.Fatalf("Expected no error with individual fallback, got: %v", err)
	}
	
	if len(embeddings) != len(texts) {
		t.Fatalf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}
}

func TestBatchProcessor_EmptyInput(t *testing.T) {
	mockVectorizer := NewMockVectorizer("test-model", 384)
	processor := NewBatchProcessor(mockVectorizer, DefaultBatchProcessorConfig())
	
	ctx := context.Background()
	embeddings, err := processor.ProcessTexts(ctx, []string{})
	
	if err != nil {
		t.Fatalf("Expected no error for empty input, got: %v", err)
	}
	
	if len(embeddings) != 0 {
		t.Fatalf("Expected 0 embeddings for empty input, got %d", len(embeddings))
	}
}

func TestEnhancedVectorizer_Integration(t *testing.T) {
	mockVectorizer := NewMockVectorizer("test-model", 384)
	config := &VectorizerConfig{
		Type:       VectorizerTypeSentenceTransformers,
		Model:      "test-model",
		Dimensions: 384,
		Options: map[string]interface{}{
			"batch_size":         4,
			"fallback_batch_size": 2,
			"max_workers":        2,
		},
	}
	
	enhancedVectorizer := NewEnhancedVectorizer(mockVectorizer, config)
	
	texts := []string{"text1", "text2", "text3", "text4", "text5"}
	
	ctx := context.Background()
	embeddings, stats, err := enhancedVectorizer.GenerateEmbeddingsWithStats(ctx, texts)
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(embeddings) != len(texts) {
		t.Fatalf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}
	
	if stats.SuccessfulTexts != len(texts) {
		t.Errorf("Expected %d successful texts, got %d", len(texts), stats.SuccessfulTexts)
	}
	
	if enhancedVectorizer.GetDimensions() != 384 {
		t.Errorf("Expected 384 dimensions, got %d", enhancedVectorizer.GetDimensions())
	}
	
	if enhancedVectorizer.GetModel() != "test-model" {
		t.Errorf("Expected 'test-model', got %s", enhancedVectorizer.GetModel())
	}
}

func TestBatchProcessor_PerformanceStats(t *testing.T) {
	mockVectorizer := NewMockVectorizer("test-model", 384)
	processor := NewBatchProcessor(mockVectorizer, DefaultBatchProcessorConfig())
	
	texts := make([]string, 100)
	for i := range texts {
		texts[i] = fmt.Sprintf("text_%d", i)
	}
	
	ctx := context.Background()
	start := time.Now()
	
	embeddings, err := processor.ProcessTexts(ctx, texts)
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(embeddings) != len(texts) {
		t.Fatalf("Expected %d embeddings, got %d", len(texts), len(embeddings))
	}
	
	stats := processor.GetStats()
	
	// Check that performance stats are calculated
	if stats.ProcessingTime == 0 {
		t.Error("Expected processing time to be recorded")
	}
	
	if stats.ThroughputPerSec <= 0 {
		t.Error("Expected positive throughput")
	}
	
	actualDuration := time.Since(start)
	if stats.ProcessingTime > actualDuration*2 {
		t.Errorf("Processing time seems too high: %v vs actual %v", stats.ProcessingTime, actualDuration)
	}
	
	t.Logf("Processed %d texts in %v (%.2f texts/sec)", 
		stats.SuccessfulTexts, stats.ProcessingTime, stats.ThroughputPerSec)
}
