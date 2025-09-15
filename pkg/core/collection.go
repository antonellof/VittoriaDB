package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/embeddings"
)

// VittoriaCollection implements the Collection interface
type VittoriaCollection struct {
	name           string
	dimensions     int
	metric         DistanceMetric
	indexType      IndexType
	dataDir        string
	vectors        map[string]*Vector
	mu             sync.RWMutex
	created        time.Time
	modified       time.Time
	closed         bool
	vectorizer     embeddings.Vectorizer
	contentStorage *ContentStorageConfig
}

// CollectionMetadata represents collection metadata stored on disk
type CollectionMetadata struct {
	Name           string                `json:"name"`
	Dimensions     int                   `json:"dimensions"`
	Metric         DistanceMetric        `json:"metric"`
	IndexType      IndexType             `json:"index_type"`
	Created        time.Time             `json:"created"`
	Modified       time.Time             `json:"modified"`
	ContentStorage *ContentStorageConfig `json:"content_storage,omitempty"`
}

// NewCollection creates a new collection
func NewCollection(name string, dimensions int, metric DistanceMetric, indexType IndexType, dataDir string) (*VittoriaCollection, error) {
	collection := &VittoriaCollection{
		name:           name,
		dimensions:     dimensions,
		metric:         metric,
		indexType:      indexType,
		dataDir:        filepath.Join(dataDir, name),
		vectors:        make(map[string]*Vector),
		created:        time.Now(),
		modified:       time.Now(),
		contentStorage: DefaultContentStorageConfig(),
	}

	return collection, nil
}

// NewCollectionWithContentStorage creates a new collection with custom content storage config
func NewCollectionWithContentStorage(name string, dimensions int, metric DistanceMetric, indexType IndexType, dataDir string, contentStorage *ContentStorageConfig) (*VittoriaCollection, error) {
	if contentStorage == nil {
		contentStorage = DefaultContentStorageConfig()
	}

	collection := &VittoriaCollection{
		name:           name,
		dimensions:     dimensions,
		metric:         metric,
		indexType:      indexType,
		dataDir:        filepath.Join(dataDir, name),
		vectors:        make(map[string]*Vector),
		created:        time.Now(),
		modified:       time.Now(),
		contentStorage: contentStorage,
	}

	return collection, nil
}

// GetContentStorageConfig returns the current content storage configuration
func (c *VittoriaCollection) GetContentStorageConfig() *ContentStorageConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.contentStorage == nil {
		return DefaultContentStorageConfig()
	}

	// Return a copy to prevent external modifications
	return &ContentStorageConfig{
		Enabled:    c.contentStorage.Enabled,
		FieldName:  c.contentStorage.FieldName,
		MaxSize:    c.contentStorage.MaxSize,
		Compressed: c.contentStorage.Compressed,
	}
}

// SetContentStorageConfig updates the content storage configuration
func (c *VittoriaCollection) SetContentStorageConfig(config *ContentStorageConfig) error {
	if config == nil {
		return fmt.Errorf("content storage config cannot be nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate configuration
	if config.FieldName == "" {
		return fmt.Errorf("content storage field name cannot be empty")
	}

	if config.MaxSize < 0 {
		return fmt.Errorf("content storage max size cannot be negative")
	}

	// Update configuration
	c.contentStorage = &ContentStorageConfig{
		Enabled:    config.Enabled,
		FieldName:  config.FieldName,
		MaxSize:    config.MaxSize,
		Compressed: config.Compressed,
	}

	// Mark collection as modified
	c.modified = time.Now()

	return nil
}

// LoadCollection loads an existing collection from disk
func LoadCollection(name string, dataDir string) (*VittoriaCollection, error) {
	collectionDir := filepath.Join(dataDir, name)
	metadataPath := filepath.Join(collectionDir, "metadata.json")

	// Read metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata CollectionMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Use loaded content storage config or default
	contentStorage := metadata.ContentStorage
	if contentStorage == nil {
		contentStorage = DefaultContentStorageConfig()
	}

	collection := &VittoriaCollection{
		name:           metadata.Name,
		dimensions:     metadata.Dimensions,
		metric:         metadata.Metric,
		indexType:      metadata.IndexType,
		dataDir:        collectionDir,
		vectors:        make(map[string]*Vector),
		created:        metadata.Created,
		modified:       metadata.Modified,
		contentStorage: contentStorage,
	}

	// Load vectors from disk
	if err := collection.loadVectors(); err != nil {
		return nil, fmt.Errorf("failed to load vectors: %w", err)
	}

	return collection, nil
}

// Initialize initializes the collection
func (c *VittoriaCollection) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create collection directory
	if err := os.MkdirAll(c.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create collection directory: %w", err)
	}

	// Save metadata
	if err := c.saveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// Close closes the collection
func (c *VittoriaCollection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	// Save vectors to disk
	if err := c.saveVectors(); err != nil {
		return fmt.Errorf("failed to save vectors: %w", err)
	}

	// Update metadata
	c.modified = time.Now()
	if err := c.saveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	c.closed = true
	return nil
}

// Name returns the collection name
func (c *VittoriaCollection) Name() string {
	return c.name
}

// Dimensions returns the vector dimensions
func (c *VittoriaCollection) Dimensions() int {
	return c.dimensions
}

// Metric returns the distance metric
func (c *VittoriaCollection) Metric() DistanceMetric {
	return c.metric
}

// Count returns the number of vectors in the collection
func (c *VittoriaCollection) Count() (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, fmt.Errorf("collection is closed")
	}

	return int64(len(c.vectors)), nil
}

// Insert inserts a vector into the collection
func (c *VittoriaCollection) Insert(ctx context.Context, vector *Vector) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("collection is closed")
	}

	// Validate vector
	if err := c.validateVector(vector); err != nil {
		return err
	}

	// Store vector
	c.vectors[vector.ID] = &Vector{
		ID:       vector.ID,
		Vector:   make([]float32, len(vector.Vector)),
		Metadata: make(map[string]interface{}),
	}

	// Copy vector data
	copy(c.vectors[vector.ID].Vector, vector.Vector)

	// Copy metadata
	if vector.Metadata != nil {
		for k, v := range vector.Metadata {
			c.vectors[vector.ID].Metadata[k] = v
		}
	}

	c.modified = time.Now()
	return nil
}

// InsertBatch inserts multiple vectors into the collection
func (c *VittoriaCollection) InsertBatch(ctx context.Context, vectors []*Vector) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("collection is closed")
	}

	// Validate all vectors first
	for _, vector := range vectors {
		if err := c.validateVector(vector); err != nil {
			return fmt.Errorf("invalid vector %s: %w", vector.ID, err)
		}
	}

	// Insert all vectors
	for _, vector := range vectors {
		c.vectors[vector.ID] = &Vector{
			ID:       vector.ID,
			Vector:   make([]float32, len(vector.Vector)),
			Metadata: make(map[string]interface{}),
		}

		// Copy vector data
		copy(c.vectors[vector.ID].Vector, vector.Vector)

		// Copy metadata
		if vector.Metadata != nil {
			for k, v := range vector.Metadata {
				c.vectors[vector.ID].Metadata[k] = v
			}
		}
	}

	c.modified = time.Now()
	return nil
}

// Get retrieves a vector by ID
func (c *VittoriaCollection) Get(ctx context.Context, id string) (*Vector, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("collection is closed")
	}

	vector, exists := c.vectors[id]
	if !exists {
		return nil, fmt.Errorf("vector '%s' not found", id)
	}

	// Return a copy to prevent external modification
	result := &Vector{
		ID:       vector.ID,
		Vector:   make([]float32, len(vector.Vector)),
		Metadata: make(map[string]interface{}),
	}

	copy(result.Vector, vector.Vector)
	for k, v := range vector.Metadata {
		result.Metadata[k] = v
	}

	return result, nil
}

// Delete removes a vector by ID
func (c *VittoriaCollection) Delete(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("collection is closed")
	}

	if _, exists := c.vectors[id]; !exists {
		return fmt.Errorf("vector '%s' not found", id)
	}

	delete(c.vectors, id)
	c.modified = time.Now()
	return nil
}

// Search performs vector similarity search
func (c *VittoriaCollection) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("collection is closed")
	}

	startTime := time.Now()

	// Validate search request
	if err := c.validateSearchRequest(req); err != nil {
		return nil, err
	}

	// Perform brute force search for now (will be optimized with proper indexing)
	candidates := make([]*SearchResult, 0, len(c.vectors))

	for _, vector := range c.vectors {
		// Apply metadata filter if specified
		if req.Filter != nil && !c.matchesFilter(vector.Metadata, req.Filter) {
			continue
		}

		// Calculate similarity score
		score := c.calculateSimilarity(req.Vector, vector.Vector)

		result := &SearchResult{
			ID:    vector.ID,
			Score: score,
		}

		// Include vector if requested
		if req.IncludeVector {
			result.Vector = make([]float32, len(vector.Vector))
			copy(result.Vector, vector.Vector)
		}

		// Include metadata if requested
		if req.IncludeMetadata {
			result.Metadata = make(map[string]interface{})
			for k, v := range vector.Metadata {
				result.Metadata[k] = v
			}
		}

		// Include content if requested and content storage is enabled
		if req.IncludeContent && c.contentStorage != nil && c.contentStorage.Enabled {
			if content, exists := vector.Metadata[c.contentStorage.FieldName]; exists {
				if contentStr, ok := content.(string); ok {
					result.Content = contentStr
				}
			}
		}

		candidates = append(candidates, result)
	}

	// Sort by score (descending for similarity)
	c.sortCandidates(candidates)

	// Apply limit and offset
	start := req.Offset
	if start > len(candidates) {
		start = len(candidates)
	}

	end := start + req.Limit
	if end > len(candidates) {
		end = len(candidates)
	}

	results := candidates[start:end]
	tookMS := time.Since(startTime).Milliseconds()

	return &SearchResponse{
		Results:   results,
		Total:     int64(len(candidates)),
		TookMS:    tookMS,
		RequestID: fmt.Sprintf("%d", time.Now().UnixNano()),
	}, nil
}

// Compact performs collection compaction
func (c *VittoriaCollection) Compact(ctx context.Context) error {
	// TODO: Implement compaction
	return nil
}

// Flush flushes pending changes to disk
func (c *VittoriaCollection) Flush(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("collection is closed")
	}

	// Save vectors to disk
	if err := c.saveVectors(); err != nil {
		return fmt.Errorf("failed to save vectors: %w", err)
	}

	// Update metadata
	c.modified = time.Now()
	if err := c.saveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// Info returns collection information
func (c *VittoriaCollection) Info() (*CollectionInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count, _ := c.Count()

	return &CollectionInfo{
		Name:        c.name,
		Dimensions:  c.dimensions,
		Metric:      c.metric,
		IndexType:   c.indexType,
		VectorCount: count,
		Created:     c.created,
		Modified:    c.modified,
	}, nil
}

// validateVector validates a vector before insertion
func (c *VittoriaCollection) validateVector(vector *Vector) error {
	if vector.ID == "" {
		return fmt.Errorf("vector ID cannot be empty")
	}

	if len(vector.Vector) != c.dimensions {
		return fmt.Errorf("vector dimensions (%d) don't match collection dimensions (%d)", len(vector.Vector), c.dimensions)
	}

	return nil
}

// validateSearchRequest validates a search request
func (c *VittoriaCollection) validateSearchRequest(req *SearchRequest) error {
	if len(req.Vector) != c.dimensions {
		return fmt.Errorf("query vector dimensions (%d) don't match collection dimensions (%d)", len(req.Vector), c.dimensions)
	}

	if req.Limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}

	if req.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	return nil
}

// calculateSimilarity calculates similarity between two vectors
func (c *VittoriaCollection) calculateSimilarity(a, b []float32) float32 {
	switch c.metric {
	case DistanceMetricCosine:
		return cosineSimilarity(a, b)
	case DistanceMetricEuclidean:
		return 1.0 / (1.0 + euclideanDistance(a, b))
	case DistanceMetricDotProduct:
		return dotProduct(a, b)
	case DistanceMetricManhattan:
		return 1.0 / (1.0 + manhattanDistance(a, b))
	default:
		return 0.0
	}
}

// matchesFilter checks if metadata matches the filter
func (c *VittoriaCollection) matchesFilter(metadata map[string]interface{}, filter *Filter) bool {
	// TODO: Implement proper filter matching
	// For now, return true (no filtering)
	return true
}

// sortCandidates sorts search results by score (descending)
func (c *VittoriaCollection) sortCandidates(candidates []*SearchResult) {
	// Simple bubble sort for now (will be optimized)
	n := len(candidates)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if candidates[j].Score < candidates[j+1].Score {
				candidates[j], candidates[j+1] = candidates[j+1], candidates[j]
			}
		}
	}
}

// saveMetadata saves collection metadata to disk
func (c *VittoriaCollection) saveMetadata() error {
	metadata := CollectionMetadata{
		Name:           c.name,
		Dimensions:     c.dimensions,
		Metric:         c.metric,
		IndexType:      c.indexType,
		Created:        c.created,
		Modified:       c.modified,
		ContentStorage: c.contentStorage,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	metadataPath := filepath.Join(c.dataDir, "metadata.json")
	return os.WriteFile(metadataPath, data, 0644)
}

// saveVectors saves vectors to disk
func (c *VittoriaCollection) saveVectors() error {
	vectorsPath := filepath.Join(c.dataDir, "vectors.json")

	data, err := json.MarshalIndent(c.vectors, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(vectorsPath, data, 0644)
}

// loadVectors loads vectors from disk
func (c *VittoriaCollection) loadVectors() error {
	vectorsPath := filepath.Join(c.dataDir, "vectors.json")

	// Check if vectors file exists
	if _, err := os.Stat(vectorsPath); os.IsNotExist(err) {
		// No vectors file, start with empty collection
		return nil
	}

	data, err := os.ReadFile(vectorsPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &c.vectors)
}

// Distance calculation functions
func cosineSimilarity(a, b []float32) float32 {
	var dotProduct, normA, normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(sqrt(float64(normA))) * float32(sqrt(float64(normB))))
}

func euclideanDistance(a, b []float32) float32 {
	var sum float32
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return float32(sqrt(float64(sum)))
}

func dotProduct(a, b []float32) float32 {
	var sum float32
	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
	}
	return sum
}

func manhattanDistance(a, b []float32) float32 {
	var sum float32
	for i := 0; i < len(a); i++ {
		if a[i] > b[i] {
			sum += a[i] - b[i]
		} else {
			sum += b[i] - a[i]
		}
	}
	return sum
}

// sqrt is a simple square root implementation
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}

	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// InsertText inserts text that will be automatically vectorized
func (c *VittoriaCollection) InsertText(ctx context.Context, textVector *TextVector) error {
	if c.vectorizer == nil {
		return fmt.Errorf("no vectorizer configured for collection '%s'", c.name)
	}

	// Generate embedding from text
	embedding, err := c.vectorizer.GenerateEmbedding(ctx, textVector.Text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Prepare metadata - preserve original content if enabled
	metadata := make(map[string]interface{})

	// Copy existing metadata
	if textVector.Metadata != nil {
		for k, v := range textVector.Metadata {
			metadata[k] = v
		}
	}

	// Store original content if content storage is enabled
	if c.contentStorage != nil && c.contentStorage.Enabled {
		// Check content size limits
		if c.contentStorage.MaxSize > 0 && int64(len(textVector.Text)) > c.contentStorage.MaxSize {
			return fmt.Errorf("content size (%d bytes) exceeds maximum allowed size (%d bytes)", len(textVector.Text), c.contentStorage.MaxSize)
		}

		// Store content (with optional compression in future)
		contentToStore := textVector.Text
		if c.contentStorage.Compressed {
			// TODO: Implement compression if needed
			// For now, store as-is
		}

		metadata[c.contentStorage.FieldName] = contentToStore
	}

	// Create vector and insert
	vector := &Vector{
		ID:       textVector.ID,
		Vector:   embedding,
		Metadata: metadata,
	}

	return c.Insert(ctx, vector)
}

// InsertTextBatch inserts multiple text vectors that will be automatically vectorized
func (c *VittoriaCollection) InsertTextBatch(ctx context.Context, textVectors []*TextVector) error {
	if c.vectorizer == nil {
		return fmt.Errorf("no vectorizer configured for collection '%s'", c.name)
	}

	// Extract texts for batch embedding generation
	texts := make([]string, len(textVectors))
	for i, tv := range textVectors {
		texts[i] = tv.Text
	}

	// Generate embeddings in batch
	embeddings, err := c.vectorizer.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Create vectors and insert
	vectors := make([]*Vector, len(textVectors))
	for i, tv := range textVectors {
		// Prepare metadata - preserve original content if enabled
		metadata := make(map[string]interface{})

		// Copy existing metadata
		if tv.Metadata != nil {
			for k, v := range tv.Metadata {
				metadata[k] = v
			}
		}

		// Store original content if content storage is enabled
		if c.contentStorage != nil && c.contentStorage.Enabled {
			// Check content size limits
			if c.contentStorage.MaxSize > 0 && int64(len(tv.Text)) > c.contentStorage.MaxSize {
				return fmt.Errorf("content size (%d bytes) exceeds maximum allowed size (%d bytes) for vector %s", len(tv.Text), c.contentStorage.MaxSize, tv.ID)
			}

			// Store content (with optional compression in future)
			contentToStore := tv.Text
			if c.contentStorage.Compressed {
				// TODO: Implement compression if needed
				// For now, store as-is
			}

			metadata[c.contentStorage.FieldName] = contentToStore
		}

		vectors[i] = &Vector{
			ID:       tv.ID,
			Vector:   embeddings[i],
			Metadata: metadata,
		}
	}

	return c.InsertBatch(ctx, vectors)
}

// SearchText performs text-based search (automatically vectorizes query)
func (c *VittoriaCollection) SearchText(ctx context.Context, query string, limit int, filter *Filter) (*SearchResponse, error) {
	if c.vectorizer == nil {
		return nil, fmt.Errorf("no vectorizer configured for collection '%s'", c.name)
	}

	// Generate embedding from query text
	queryEmbedding, err := c.vectorizer.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Create search request
	searchReq := &SearchRequest{
		Vector:          queryEmbedding,
		Limit:           limit,
		Offset:          0,
		Filter:          filter,
		IncludeVector:   false,
		IncludeMetadata: true,
	}

	return c.Search(ctx, searchReq)
}

// HasVectorizer returns true if the collection has a vectorizer configured
func (c *VittoriaCollection) HasVectorizer() bool {
	return c.vectorizer != nil
}

// GetVectorizer returns the collection's vectorizer
func (c *VittoriaCollection) GetVectorizer() embeddings.Vectorizer {
	return c.vectorizer
}

// SetVectorizer sets the collection's vectorizer
func (c *VittoriaCollection) SetVectorizer(vectorizer embeddings.Vectorizer) {
	c.vectorizer = vectorizer
}
