package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/embeddings"
)

// dirByteSize returns the sum of regular file sizes under root (recursive).
func dirByteSize(root string) int64 {
	var n int64
	_ = filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info != nil && !info.IsDir() {
			n += info.Size()
		}
		return nil
	})
	return n
}

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

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	disk := dirByteSize(db.dataDir)

	return &HealthStatus{
		Status:       "healthy",
		Uptime:       int64(time.Since(db.startTime).Seconds()),
		Collections:  len(db.collections),
		TotalVectors: totalVectors,
		MemoryUsage:  int64(ms.Alloc),
		DiskUsage:    disk,
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

		info, err := collection.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get collection info: %w", err)
		}

		colDir := filepath.Join(db.dataDir, collection.Name())
		onDisk := dirByteSize(colDir)

		stats := &CollectionStats{
			Name:         collection.Name(),
			VectorCount:  count,
			Dimensions:   collection.Dimensions(),
			IndexType:    collection.indexType,
			IndexSize:    onDisk,
			LastModified: info.Modified,
		}

		collectionStats = append(collectionStats, stats)
		totalVectors += count
		totalSize += onDisk
		indexSize += onDisk
	}

	qTotal, avgLat := SearchMetricsSnapshot()
	uptimeSec := time.Since(db.startTime).Seconds()
	qps := float64(0)
	if uptimeSec > 0 && qTotal > 0 {
		qps = float64(qTotal) / uptimeSec
	}

	return &DatabaseStats{
		Collections:     collectionStats,
		TotalVectors:      totalVectors,
		TotalSize:         totalSize,
		IndexSize:         indexSize,
		QueriesTotal:      int64(qTotal),
		QueriesPerSec:     qps,
		AvgQueryLatency:   avgLat.Seconds(),
	}, nil
}

// Backup / Restore live in backup.go

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
