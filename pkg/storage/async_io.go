package storage

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// AsyncIOConfig holds configuration for async I/O operations
type AsyncIOConfig struct {
	Enabled        bool          `json:"enabled"`
	WorkerPoolSize int           `json:"worker_pool_size"`
	QueueSize      int           `json:"queue_size"`
	BatchSize      int           `json:"batch_size"`
	FlushInterval  time.Duration `json:"flush_interval"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
}

// DefaultAsyncIOConfig returns default async I/O configuration
func DefaultAsyncIOConfig() *AsyncIOConfig {
	return &AsyncIOConfig{
		Enabled:        true,
		WorkerPoolSize: runtime.NumCPU() * 2,
		QueueSize:      10000,
		BatchSize:      100,
		FlushInterval:  100 * time.Millisecond,
		MaxRetries:     3,
		RetryDelay:     10 * time.Millisecond,
	}
}

// AsyncIOOperation represents an async I/O operation
type AsyncIOOperation struct {
	Type      AsyncIOOpType
	Data      []byte
	Offset    int64
	Result    chan AsyncIOResult
	Context   context.Context
	Metadata  map[string]interface{}
	CreatedAt time.Time
	Retries   int
}

// AsyncIOOpType represents the type of async I/O operation
type AsyncIOOpType int

const (
	AsyncIOOpRead AsyncIOOpType = iota
	AsyncIOOpWrite
	AsyncIOOpSync
	AsyncIOOpFlush
	AsyncIOOpBatch
)

func (op AsyncIOOpType) String() string {
	switch op {
	case AsyncIOOpRead:
		return "read"
	case AsyncIOOpWrite:
		return "write"
	case AsyncIOOpSync:
		return "sync"
	case AsyncIOOpFlush:
		return "flush"
	case AsyncIOOpBatch:
		return "batch"
	default:
		return "unknown"
	}
}

// AsyncIOResult represents the result of an async I/O operation
type AsyncIOResult struct {
	Data      []byte
	BytesRead int
	Error     error
	Duration  time.Duration
	OpType    AsyncIOOpType
}

// AsyncIOEngine provides asynchronous I/O operations
type AsyncIOEngine struct {
	config     *AsyncIOConfig
	storage    StorageEngine
	operations chan *AsyncIOOperation
	workers    []*AsyncIOWorker
	batcher    *AsyncIOBatcher
	stats      *AsyncIOStats
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	running    bool
}

// NewAsyncIOEngine creates a new async I/O engine
func NewAsyncIOEngine(storage StorageEngine, config *AsyncIOConfig) *AsyncIOEngine {
	if config == nil {
		config = DefaultAsyncIOConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	engine := &AsyncIOEngine{
		config:     config,
		storage:    storage,
		operations: make(chan *AsyncIOOperation, config.QueueSize),
		workers:    make([]*AsyncIOWorker, config.WorkerPoolSize),
		stats:      NewAsyncIOStats(),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Create batcher for batched operations
	engine.batcher = NewAsyncIOBatcher(engine, config)

	return engine
}

// Start starts the async I/O engine
func (aio *AsyncIOEngine) Start() error {
	aio.mu.Lock()
	defer aio.mu.Unlock()

	if aio.running {
		return fmt.Errorf("async I/O engine is already running")
	}

	// Start workers
	for i := 0; i < aio.config.WorkerPoolSize; i++ {
		worker := NewAsyncIOWorker(i, aio.operations, aio.storage, aio.stats)
		aio.workers[i] = worker

		aio.wg.Add(1)
		go func(w *AsyncIOWorker) {
			defer aio.wg.Done()
			w.Run(aio.ctx)
		}(worker)
	}

	// Start batcher
	aio.wg.Add(1)
	go func() {
		defer aio.wg.Done()
		aio.batcher.Run(aio.ctx)
	}()

	// Start stats collector
	aio.wg.Add(1)
	go func() {
		defer aio.wg.Done()
		aio.runStatsCollector()
	}()

	aio.running = true
	return nil
}

// Stop stops the async I/O engine
func (aio *AsyncIOEngine) Stop() error {
	aio.mu.Lock()
	defer aio.mu.Unlock()

	if !aio.running {
		return nil
	}

	// Cancel context to signal shutdown
	aio.cancel()

	// Close operations channel
	close(aio.operations)

	// Wait for all workers to finish
	aio.wg.Wait()

	aio.running = false
	return nil
}

// ReadAsync performs an asynchronous read operation
func (aio *AsyncIOEngine) ReadAsync(ctx context.Context, pageID uint32) <-chan AsyncIOResult {
	result := make(chan AsyncIOResult, 1)

	if !aio.config.Enabled {
		// Fallback to synchronous operation
		go func() {
			defer close(result)
			page, err := aio.storage.ReadPage(pageID)

			var data []byte
			var bytesRead int
			if page != nil {
				data = page.Data
				bytesRead = len(data)
			}

			result <- AsyncIOResult{
				Data:      data,
				BytesRead: bytesRead,
				Error:     err,
				OpType:    AsyncIOOpRead,
			}
		}()
		return result
	}

	op := &AsyncIOOperation{
		Type:      AsyncIOOpRead,
		Offset:    int64(pageID),
		Result:    result,
		Context:   ctx,
		CreatedAt: time.Now(),
		Metadata:  map[string]interface{}{"page_id": pageID},
	}

	select {
	case aio.operations <- op:
		aio.stats.IncrementQueued(AsyncIOOpRead)
	case <-ctx.Done():
		result <- AsyncIOResult{
			Error:  ctx.Err(),
			OpType: AsyncIOOpRead,
		}
		close(result)
	case <-aio.ctx.Done():
		result <- AsyncIOResult{
			Error:  fmt.Errorf("async I/O engine is shutting down"),
			OpType: AsyncIOOpRead,
		}
		close(result)
	}

	return result
}

// WriteAsync performs an asynchronous write operation
func (aio *AsyncIOEngine) WriteAsync(ctx context.Context, page *Page) <-chan AsyncIOResult {
	result := make(chan AsyncIOResult, 1)

	if !aio.config.Enabled {
		// Fallback to synchronous operation
		go func() {
			defer close(result)
			err := aio.storage.WritePage(page)

			result <- AsyncIOResult{
				BytesRead: len(page.Data),
				Error:     err,
				OpType:    AsyncIOOpWrite,
			}
		}()
		return result
	}

	// Serialize page data
	data := make([]byte, len(page.Data))
	copy(data, page.Data)

	op := &AsyncIOOperation{
		Type:      AsyncIOOpWrite,
		Data:      data,
		Offset:    int64(page.ID),
		Result:    result,
		Context:   ctx,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"page_id":   page.ID,
			"page_type": page.Type,
			"page_size": page.Size,
		},
	}

	select {
	case aio.operations <- op:
		aio.stats.IncrementQueued(AsyncIOOpWrite)
	case <-ctx.Done():
		result <- AsyncIOResult{
			Error:  ctx.Err(),
			OpType: AsyncIOOpWrite,
		}
		close(result)
	case <-aio.ctx.Done():
		result <- AsyncIOResult{
			Error:  fmt.Errorf("async I/O engine is shutting down"),
			OpType: AsyncIOOpWrite,
		}
		close(result)
	}

	return result
}

// SyncAsync performs an asynchronous sync operation
func (aio *AsyncIOEngine) SyncAsync(ctx context.Context) <-chan AsyncIOResult {
	result := make(chan AsyncIOResult, 1)

	op := &AsyncIOOperation{
		Type:      AsyncIOOpSync,
		Result:    result,
		Context:   ctx,
		CreatedAt: time.Now(),
	}

	select {
	case aio.operations <- op:
		aio.stats.IncrementQueued(AsyncIOOpSync)
	case <-ctx.Done():
		result <- AsyncIOResult{
			Error:  ctx.Err(),
			OpType: AsyncIOOpSync,
		}
		close(result)
	case <-aio.ctx.Done():
		result <- AsyncIOResult{
			Error:  fmt.Errorf("async I/O engine is shutting down"),
			OpType: AsyncIOOpSync,
		}
		close(result)
	}

	return result
}

// BatchWriteAsync performs asynchronous batch write operations
func (aio *AsyncIOEngine) BatchWriteAsync(ctx context.Context, pages []*Page) <-chan AsyncIOResult {
	return aio.batcher.BatchWrite(ctx, pages)
}

// GetStats returns async I/O statistics
func (aio *AsyncIOEngine) GetStats() *AsyncIOStats {
	return aio.stats.Copy()
}

func (aio *AsyncIOEngine) runStatsCollector() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			aio.stats.UpdateThroughput()
		case <-aio.ctx.Done():
			return
		}
	}
}

// AsyncIOWorker handles async I/O operations
type AsyncIOWorker struct {
	id         int
	operations <-chan *AsyncIOOperation
	storage    StorageEngine
	stats      *AsyncIOStats
}

// NewAsyncIOWorker creates a new async I/O worker
func NewAsyncIOWorker(id int, operations <-chan *AsyncIOOperation, storage StorageEngine, stats *AsyncIOStats) *AsyncIOWorker {
	return &AsyncIOWorker{
		id:         id,
		operations: operations,
		storage:    storage,
		stats:      stats,
	}
}

// Run runs the async I/O worker
func (w *AsyncIOWorker) Run(ctx context.Context) {
	for {
		select {
		case op, ok := <-w.operations:
			if !ok {
				return // Channel closed
			}
			w.processOperation(op)
		case <-ctx.Done():
			return
		}
	}
}

func (w *AsyncIOWorker) processOperation(op *AsyncIOOperation) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		w.stats.RecordOperation(op.Type, duration)

		// Send result if channel is still open
		select {
		case op.Result <- AsyncIOResult{
			Duration: duration,
			OpType:   op.Type,
		}:
		default:
		}
		close(op.Result)
	}()

	var err error
	var data []byte
	var bytesProcessed int

	switch op.Type {
	case AsyncIOOpRead:
		pageID := uint32(op.Offset)
		page, readErr := w.storage.ReadPage(pageID)
		if readErr != nil {
			err = readErr
		} else if page != nil {
			data = page.Data
			bytesProcessed = len(data)
		}

	case AsyncIOOpWrite:
		// Reconstruct page from operation data
		pageID := uint32(op.Offset)
		page := &Page{
			ID:   pageID,
			Data: op.Data,
		}

		// Set additional fields from metadata if available
		if pageType, ok := op.Metadata["page_type"].(PageType); ok {
			page.Type = pageType
		}
		if pageSize, ok := op.Metadata["page_size"].(uint16); ok {
			page.Size = pageSize
		}

		err = w.storage.WritePage(page)
		bytesProcessed = len(op.Data)

	case AsyncIOOpSync:
		err = w.storage.Sync()

	default:
		err = fmt.Errorf("unsupported async I/O operation: %s", op.Type)
	}

	// Update result
	select {
	case op.Result <- AsyncIOResult{
		Data:      data,
		BytesRead: bytesProcessed,
		Error:     err,
		Duration:  time.Since(start),
		OpType:    op.Type,
	}:
	default:
	}
}

// AsyncIOBatcher handles batched async I/O operations
type AsyncIOBatcher struct {
	engine     *AsyncIOEngine
	config     *AsyncIOConfig
	writeQueue []*Page
	mu         sync.Mutex
}

// NewAsyncIOBatcher creates a new async I/O batcher
func NewAsyncIOBatcher(engine *AsyncIOEngine, config *AsyncIOConfig) *AsyncIOBatcher {
	return &AsyncIOBatcher{
		engine:     engine,
		config:     config,
		writeQueue: make([]*Page, 0, config.BatchSize),
	}
}

// Run runs the async I/O batcher
func (b *AsyncIOBatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(b.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.flushWrites()
		case <-ctx.Done():
			b.flushWrites() // Final flush
			return
		}
	}
}

// BatchWrite queues pages for batched writing
func (b *AsyncIOBatcher) BatchWrite(ctx context.Context, pages []*Page) <-chan AsyncIOResult {
	result := make(chan AsyncIOResult, 1)

	b.mu.Lock()
	defer b.mu.Unlock()

	// Add pages to write queue
	b.writeQueue = append(b.writeQueue, pages...)

	// If queue is full, flush immediately
	if len(b.writeQueue) >= b.config.BatchSize {
		go func() {
			defer close(result)
			err := b.flushWritesLocked()
			result <- AsyncIOResult{
				BytesRead: len(pages) * PageSize,
				Error:     err,
				OpType:    AsyncIOOpBatch,
			}
		}()
	} else {
		// Return success immediately for queued writes
		go func() {
			defer close(result)
			result <- AsyncIOResult{
				BytesRead: len(pages) * PageSize,
				OpType:    AsyncIOOpBatch,
			}
		}()
	}

	return result
}

func (b *AsyncIOBatcher) flushWrites() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushWritesLocked()
}

func (b *AsyncIOBatcher) flushWritesLocked() error {
	if len(b.writeQueue) == 0 {
		return nil
	}

	// Process all queued writes
	var lastErr error
	for _, page := range b.writeQueue {
		if err := b.engine.storage.WritePage(page); err != nil {
			lastErr = err
		}
	}

	// Clear the queue
	b.writeQueue = b.writeQueue[:0]

	return lastErr
}

// AsyncIOStats tracks async I/O statistics
type AsyncIOStats struct {
	mu               sync.RWMutex
	OperationsQueued map[AsyncIOOpType]int64 `json:"operations_queued"`
	OperationsTotal  map[AsyncIOOpType]int64 `json:"operations_total"`
	OperationLatency map[AsyncIOOpType]int64 `json:"operation_latency_ns"`
	BytesRead        int64                   `json:"bytes_read"`
	BytesWritten     int64                   `json:"bytes_written"`
	ErrorCount       int64                   `json:"error_count"`
	QueueDepth       int                     `json:"queue_depth"`
	Throughput       float64                 `json:"throughput_ops_per_sec"`
	LastUpdate       time.Time               `json:"last_update"`
	StartTime        time.Time               `json:"start_time"`
}

// NewAsyncIOStats creates new async I/O statistics
func NewAsyncIOStats() *AsyncIOStats {
	return &AsyncIOStats{
		OperationsQueued: make(map[AsyncIOOpType]int64),
		OperationsTotal:  make(map[AsyncIOOpType]int64),
		OperationLatency: make(map[AsyncIOOpType]int64),
		StartTime:        time.Now(),
		LastUpdate:       time.Now(),
	}
}

// IncrementQueued increments the queued operations counter
func (s *AsyncIOStats) IncrementQueued(opType AsyncIOOpType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OperationsQueued[opType]++
}

// RecordOperation records a completed operation
func (s *AsyncIOStats) RecordOperation(opType AsyncIOOpType, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.OperationsTotal[opType]++
	s.OperationLatency[opType] += duration.Nanoseconds()

	if s.OperationsQueued[opType] > 0 {
		s.OperationsQueued[opType]--
	}
}

// UpdateThroughput updates the throughput calculation
func (s *AsyncIOStats) UpdateThroughput() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(s.LastUpdate).Seconds()

	if elapsed > 0 {
		var totalOps int64
		for _, count := range s.OperationsTotal {
			totalOps += count
		}

		s.Throughput = float64(totalOps) / now.Sub(s.StartTime).Seconds()
	}

	s.LastUpdate = now
}

// Copy returns a copy of the statistics
func (s *AsyncIOStats) Copy() *AsyncIOStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copy := &AsyncIOStats{
		OperationsQueued: make(map[AsyncIOOpType]int64),
		OperationsTotal:  make(map[AsyncIOOpType]int64),
		OperationLatency: make(map[AsyncIOOpType]int64),
		BytesRead:        s.BytesRead,
		BytesWritten:     s.BytesWritten,
		ErrorCount:       s.ErrorCount,
		QueueDepth:       s.QueueDepth,
		Throughput:       s.Throughput,
		LastUpdate:       s.LastUpdate,
		StartTime:        s.StartTime,
	}

	for k, v := range s.OperationsQueued {
		copy.OperationsQueued[k] = v
	}
	for k, v := range s.OperationsTotal {
		copy.OperationsTotal[k] = v
	}
	for k, v := range s.OperationLatency {
		copy.OperationLatency[k] = v
	}

	return copy
}
