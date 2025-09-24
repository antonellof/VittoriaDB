package embeddings

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// BatchProcessorConfig holds configuration for batch processing
type BatchProcessorConfig struct {
	BatchSize      int           // Primary batch size
	FallbackSize   int           // Smaller batch size for fallback
	MaxRetries     int           // Maximum retry attempts
	RetryDelay     time.Duration // Delay between retries
	MaxWorkers     int           // Maximum concurrent workers
	EnableFallback bool          // Enable fallback to smaller batches
}

// DefaultBatchProcessorConfig returns sensible defaults
func DefaultBatchProcessorConfig() *BatchProcessorConfig {
	return &BatchProcessorConfig{
		BatchSize:      32,
		FallbackSize:   8,
		MaxRetries:     3,
		RetryDelay:     time.Second,
		MaxWorkers:     4,
		EnableFallback: true,
	}
}

// BatchProcessor provides enhanced batch processing with error recovery
type BatchProcessor struct {
	vectorizer Vectorizer
	config     *BatchProcessorConfig
	stats      *BatchProcessorStats
	mu         sync.RWMutex
}

// BatchProcessorStats tracks processing statistics
type BatchProcessorStats struct {
	TotalTexts        int           `json:"total_texts"`
	SuccessfulTexts   int           `json:"successful_texts"`
	FailedTexts       int           `json:"failed_texts"`
	BatchesProcessed  int           `json:"batches_processed"`
	FallbacksUsed     int           `json:"fallbacks_used"`
	RetriesUsed       int           `json:"retries_used"`
	ProcessingTime    time.Duration `json:"processing_time"`
	AverageLatency    time.Duration `json:"average_latency"`
	ThroughputPerSec  float64       `json:"throughput_per_sec"`
}

// NewBatchProcessor creates a new batch processor with the given vectorizer
func NewBatchProcessor(vectorizer Vectorizer, config *BatchProcessorConfig) *BatchProcessor {
	if config == nil {
		config = DefaultBatchProcessorConfig()
	}

	return &BatchProcessor{
		vectorizer: vectorizer,
		config:     config,
		stats:      &BatchProcessorStats{},
	}
}

// ProcessTexts processes multiple texts with enhanced error recovery
func (bp *BatchProcessor) ProcessTexts(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	startTime := time.Now()
	bp.resetStats(len(texts))

	// Try full batch processing first
	embeddings, err := bp.tryFullBatch(ctx, texts)
	if err == nil {
		bp.updateStats(len(texts), 0, time.Since(startTime))
		return embeddings, nil
	}

	log.Printf("Full batch processing failed: %v. Falling back to smaller batches...", err)

	// Fall back to smaller batch processing
	if bp.config.EnableFallback {
		embeddings, err = bp.processBatched(ctx, texts)
		if err == nil {
			bp.updateStats(len(embeddings), bp.stats.FailedTexts, time.Since(startTime))
			return embeddings, nil
		}
	}

	// Final fallback: process individually
	log.Printf("Batch processing failed: %v. Processing individually...", err)
	embeddings, err = bp.processIndividually(ctx, texts)
	bp.updateStats(len(embeddings), bp.stats.FailedTexts, time.Since(startTime))

	return embeddings, err
}

// tryFullBatch attempts to process all texts in a single batch
func (bp *BatchProcessor) tryFullBatch(ctx context.Context, texts []string) ([][]float32, error) {
	var lastErr error

	for attempt := 0; attempt < bp.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(bp.config.RetryDelay):
				// Continue with retry
			}
			bp.incrementRetries()
		}

		embeddings, err := bp.vectorizer.GenerateEmbeddings(ctx, texts)
		if err == nil {
			bp.incrementBatches()
			return embeddings, nil
		}

		lastErr = err
		log.Printf("Full batch attempt %d failed: %v", attempt+1, err)
	}

	return nil, fmt.Errorf("full batch processing failed after %d attempts: %w", bp.config.MaxRetries, lastErr)
}

// processBatched processes texts in smaller batches with parallel processing
func (bp *BatchProcessor) processBatched(ctx context.Context, texts []string) ([][]float32, error) {
	batchSize := bp.config.FallbackSize
	numBatches := (len(texts) + batchSize - 1) / batchSize

	// Use semaphore to limit concurrent workers
	semaphore := make(chan struct{}, bp.config.MaxWorkers)
	results := make([][]float32, len(texts))
	errors := make([]error, numBatches)

	var wg sync.WaitGroup

	for i := 0; i < numBatches; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		wg.Add(1)
		go func(batchIdx, startIdx, endIdx int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			batch := texts[startIdx:endIdx]
			embeddings, err := bp.processBatchWithRetry(ctx, batch)

			if err != nil {
				errors[batchIdx] = err
				bp.incrementFailures(len(batch))
				log.Printf("Batch %d failed: %v", batchIdx, err)
				return
			}

			// Copy results to the correct positions
			copy(results[startIdx:endIdx], embeddings)
			bp.incrementBatches()

		}(i, start, end)
	}

	wg.Wait()

	// Check for errors and filter out failed batches
	var finalResults [][]float32
	var hasErrors bool

	for i, batch := range results {
		if errors[i/batchSize] == nil && len(batch) > 0 {
			finalResults = append(finalResults, batch)
		} else {
			hasErrors = true
		}
	}

	if hasErrors && len(finalResults) == 0 {
		return nil, fmt.Errorf("all batches failed")
	}

	bp.incrementFallbacks()
	return finalResults, nil
}

// processBatchWithRetry processes a single batch with retry logic
func (bp *BatchProcessor) processBatchWithRetry(ctx context.Context, batch []string) ([][]float32, error) {
	var lastErr error

	for attempt := 0; attempt < bp.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(bp.config.RetryDelay):
				// Continue with retry
			}
			bp.incrementRetries()
		}

		embeddings, err := bp.vectorizer.GenerateEmbeddings(ctx, batch)
		if err == nil {
			return embeddings, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("batch processing failed after %d attempts: %w", bp.config.MaxRetries, lastErr)
}

// processIndividually processes each text individually as final fallback
func (bp *BatchProcessor) processIndividually(ctx context.Context, texts []string) ([][]float32, error) {
	var results [][]float32
	var errors []error

	semaphore := make(chan struct{}, bp.config.MaxWorkers)
	resultsChan := make(chan struct {
		idx       int
		embedding []float32
		err       error
	}, len(texts))

	var wg sync.WaitGroup

	for i, text := range texts {
		wg.Add(1)
		go func(idx int, txt string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			embedding, err := bp.vectorizer.GenerateEmbedding(ctx, txt)
			resultsChan <- struct {
				idx       int
				embedding []float32
				err       error
			}{idx, embedding, err}

		}(i, text)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	resultMap := make(map[int][]float32)
	for result := range resultsChan {
		if result.err != nil {
			errors = append(errors, result.err)
			bp.incrementFailures(1)
		} else {
			resultMap[result.idx] = result.embedding
		}
	}

	// Build final results in order
	for i := 0; i < len(texts); i++ {
		if embedding, exists := resultMap[i]; exists {
			results = append(results, embedding)
		}
	}

	if len(results) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all individual processing failed: %v", errors[0])
	}

	return results, nil
}

// Statistics methods
func (bp *BatchProcessor) resetStats(totalTexts int) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.stats = &BatchProcessorStats{
		TotalTexts: totalTexts,
	}
}

func (bp *BatchProcessor) updateStats(successful, failed int, duration time.Duration) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.stats.SuccessfulTexts = successful
	bp.stats.FailedTexts = failed
	bp.stats.ProcessingTime = duration

	if successful > 0 {
		bp.stats.AverageLatency = duration / time.Duration(successful)
		bp.stats.ThroughputPerSec = float64(successful) / duration.Seconds()
	}
}

func (bp *BatchProcessor) incrementBatches() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.stats.BatchesProcessed++
}

func (bp *BatchProcessor) incrementFallbacks() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.stats.FallbacksUsed++
}

func (bp *BatchProcessor) incrementRetries() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.stats.RetriesUsed++
}

func (bp *BatchProcessor) incrementFailures(count int) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.stats.FailedTexts += count
}

// GetStats returns a copy of the current statistics
func (bp *BatchProcessor) GetStats() BatchProcessorStats {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return *bp.stats
}

// GetConfig returns the current configuration
func (bp *BatchProcessor) GetConfig() *BatchProcessorConfig {
	return bp.config
}

// UpdateConfig updates the processor configuration
func (bp *BatchProcessor) UpdateConfig(config *BatchProcessorConfig) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.config = config
}
