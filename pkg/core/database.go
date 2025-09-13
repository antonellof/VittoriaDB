package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/embeddings"
)

// VittoriaDB implements the Database interface
type VittoriaDB struct {
	config      *Config
	dataDir     string
	collections map[string]*VittoriaCollection
	mu          sync.RWMutex
	startTime   time.Time
	closed      bool
}

// NewDatabase creates a new VittoriaDB instance
func NewDatabase() *VittoriaDB {
	return &VittoriaDB{
		collections: make(map[string]*VittoriaCollection),
		startTime:   time.Now(),
	}
}

// Open initializes the database with the given configuration
func (db *VittoriaDB) Open(ctx context.Context, config *Config) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return fmt.Errorf("database is closed")
	}

	db.config = config
	db.dataDir = config.DataDir

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(db.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Load existing collections
	if err := db.loadCollections(ctx); err != nil {
		return fmt.Errorf("failed to load collections: %w", err)
	}

	return nil
}

// Close closes the database and all collections
func (db *VittoriaDB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return nil
	}

	// Close all collections
	for _, collection := range db.collections {
		if err := collection.Close(); err != nil {
			// Log error but continue closing other collections
			fmt.Printf("Error closing collection %s: %v\n", collection.Name(), err)
		}
	}

	db.closed = true
	return nil
}

// Health returns the current health status
func (db *VittoriaDB) Health() *HealthStatus {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var totalVectors int64
	for _, collection := range db.collections {
		if count, err := collection.Count(); err == nil {
			totalVectors += count
		}
	}

	return &HealthStatus{
		Status:       "healthy",
		Uptime:       int64(time.Since(db.startTime).Seconds()),
		Collections:  len(db.collections),
		TotalVectors: totalVectors,
		MemoryUsage:  0, // TODO: Implement memory usage calculation
		DiskUsage:    0, // TODO: Implement disk usage calculation
	}
}

// CreateCollection creates a new vector collection
func (db *VittoriaDB) CreateCollection(ctx context.Context, req *CreateCollectionRequest) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return fmt.Errorf("database is closed")
	}

	// Check if collection already exists
	if _, exists := db.collections[req.Name]; exists {
		return fmt.Errorf("collection '%s' already exists", req.Name)
	}

	// Validate request
	if err := db.validateCreateCollectionRequest(req); err != nil {
		return err
	}

	// Create collection
	collection, err := NewCollection(req.Name, req.Dimensions, req.Metric, req.IndexType, db.dataDir)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Initialize collection
	if err := collection.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize collection: %w", err)
	}

	// Set up vectorizer if configured
	if req.VectorizerConfig != nil {
		factory := embeddings.NewVectorizerFactory()
		vectorizer, err := factory.CreateVectorizer(req.VectorizerConfig)
		if err != nil {
			return fmt.Errorf("failed to create vectorizer: %w", err)
		}
		collection.SetVectorizer(vectorizer)
	}

	db.collections[req.Name] = collection
	return nil
}

// GetCollection retrieves a collection by name
func (db *VittoriaDB) GetCollection(ctx context.Context, name string) (Collection, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return nil, fmt.Errorf("database is closed")
	}

	collection, exists := db.collections[name]
	if !exists {
		return nil, fmt.Errorf("collection '%s' not found", name)
	}

	return collection, nil
}

// ListCollections returns information about all collections
func (db *VittoriaDB) ListCollections(ctx context.Context) ([]*CollectionInfo, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return nil, fmt.Errorf("database is closed")
	}

	collections := make([]*CollectionInfo, 0, len(db.collections))
	for _, collection := range db.collections {
		info, err := collection.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get collection info: %w", err)
		}
		collections = append(collections, info)
	}

	return collections, nil
}

// DropCollection deletes a collection
func (db *VittoriaDB) DropCollection(ctx context.Context, name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return fmt.Errorf("database is closed")
	}

	collection, exists := db.collections[name]
	if !exists {
		return fmt.Errorf("collection '%s' not found", name)
	}

	// Close and remove collection
	if err := collection.Close(); err != nil {
		return fmt.Errorf("failed to close collection: %w", err)
	}

	// Remove collection files
	collectionDir := filepath.Join(db.dataDir, name)
	if err := os.RemoveAll(collectionDir); err != nil {
		return fmt.Errorf("failed to remove collection files: %w", err)
	}

	delete(db.collections, name)
	return nil
}

// Stats returns database statistics
func (db *VittoriaDB) Stats(ctx context.Context) (*DatabaseStats, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return nil, fmt.Errorf("database is closed")
	}

	var totalVectors int64
	var totalSize int64
	var indexSize int64
	collectionStats := make([]*CollectionStats, 0, len(db.collections))

	for _, collection := range db.collections {
		count, err := collection.Count()
		if err != nil {
			return nil, fmt.Errorf("failed to get collection count: %w", err)
		}

		stats := &CollectionStats{
			Name:         collection.Name(),
			VectorCount:  count,
			Dimensions:   collection.Dimensions(),
			IndexType:    collection.indexType,
			IndexSize:    0,          // TODO: Implement index size calculation
			LastModified: time.Now(), // TODO: Implement last modified tracking
		}

		collectionStats = append(collectionStats, stats)
		totalVectors += count
	}

	return &DatabaseStats{
		Collections:     collectionStats,
		TotalVectors:    totalVectors,
		TotalSize:       totalSize,
		IndexSize:       indexSize,
		QueriesTotal:    0, // TODO: Implement query tracking
		QueriesPerSec:   0, // TODO: Implement QPS calculation
		AvgQueryLatency: 0, // TODO: Implement latency tracking
	}, nil
}

// Backup creates a backup of the database
func (db *VittoriaDB) Backup(ctx context.Context, w io.Writer) error {
	// TODO: Implement backup functionality
	return fmt.Errorf("backup not implemented yet")
}

// Restore restores the database from a backup
func (db *VittoriaDB) Restore(ctx context.Context, r io.Reader) error {
	// TODO: Implement restore functionality
	return fmt.Errorf("restore not implemented yet")
}

// loadCollections loads existing collections from disk
func (db *VittoriaDB) loadCollections(ctx context.Context) error {
	entries, err := os.ReadDir(db.dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		collectionName := entry.Name()
		metadataPath := filepath.Join(db.dataDir, collectionName, "metadata.json")

		// Check if metadata file exists
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			continue
		}

		// Load collection metadata and create collection
		collection, err := LoadCollection(collectionName, db.dataDir)
		if err != nil {
			return fmt.Errorf("failed to load collection %s: %w", collectionName, err)
		}

		db.collections[collectionName] = collection
	}

	return nil
}

// validateCreateCollectionRequest validates the collection creation request
func (db *VittoriaDB) validateCreateCollectionRequest(req *CreateCollectionRequest) error {
	if req.Name == "" {
		return fmt.Errorf("collection name cannot be empty")
	}

	if req.Dimensions <= 0 {
		return fmt.Errorf("dimensions must be positive")
	}

	if req.Dimensions > 10000 {
		return fmt.Errorf("dimensions cannot exceed 10000")
	}

	// Validate metric
	switch req.Metric {
	case DistanceMetricCosine, DistanceMetricEuclidean, DistanceMetricDotProduct, DistanceMetricManhattan:
		// Valid metrics
	default:
		return fmt.Errorf("invalid distance metric")
	}

	// Validate index type
	switch req.IndexType {
	case IndexTypeFlat, IndexTypeHNSW:
		// Valid index types
	default:
		return fmt.Errorf("invalid index type")
	}

	return nil
}
