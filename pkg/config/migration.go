package config

import (
	"fmt"

	"github.com/antonellof/VittoriaDB/pkg/core"
	"github.com/antonellof/VittoriaDB/pkg/embeddings"
	"github.com/antonellof/VittoriaDB/pkg/processor"
)

// MigrationAdapter provides utilities to convert between old and new configuration formats
type MigrationAdapter struct{}

// NewMigrationAdapter creates a new migration adapter
func NewMigrationAdapter() *MigrationAdapter {
	return &MigrationAdapter{}
}

// ToLegacyConfig converts unified config to legacy format for backward compatibility
func (m *MigrationAdapter) ToLegacyConfig(unified *VittoriaConfig) *LegacyConfigBundle {
	return &LegacyConfigBundle{
		Core:       m.toCoreConfig(unified),
		Embeddings: m.toEmbeddingsConfig(unified),
		Processing: m.toProcessingConfig(unified),
		Parallel:   m.toParallelSearchConfig(unified),
		Cache:      m.toSearchCacheConfig(unified),
	}
}

// FromLegacyConfig converts legacy configurations to unified format
func (m *MigrationAdapter) FromLegacyConfig(legacy *LegacyConfigBundle) *VittoriaConfig {
	config := DefaultConfig()

	if legacy.Core != nil {
		m.fromCoreConfig(legacy.Core, config)
	}

	if legacy.Embeddings != nil {
		m.fromEmbeddingsConfig(legacy.Embeddings, config)
	}

	if legacy.Processing != nil {
		m.fromProcessingConfig(legacy.Processing, config)
	}

	if legacy.Parallel != nil {
		m.fromParallelSearchConfig(legacy.Parallel, config)
	}

	if legacy.Cache != nil {
		m.fromSearchCacheConfig(legacy.Cache, config)
	}

	return config
}

// LegacyConfigBundle holds all the old configuration structures
type LegacyConfigBundle struct {
	Core       *core.Config
	Embeddings *embeddings.VectorizerConfig
	Processing *processor.ProcessingConfig
	Parallel   *core.ParallelSearchConfig
	Cache      *core.SearchCacheConfig
}

// Convert unified config to legacy core config
func (m *MigrationAdapter) toCoreConfig(unified *VittoriaConfig) *core.Config {
	return &core.Config{
		DataDir: unified.DataDir,
		Server: core.ServerConfig{
			Host:         unified.Server.Host,
			Port:         unified.Server.Port,
			ReadTimeout:  unified.Server.ReadTimeout,
			WriteTimeout: unified.Server.WriteTimeout,
			MaxBodySize:  unified.Server.MaxBodySize,
			CORS:         unified.Server.CORS,
		},
		Storage: core.StorageConfig{
			PageSize:    unified.Storage.PageSize,
			CacheSize:   unified.Storage.CacheSize,
			SyncWrites:  unified.Storage.SyncWrites,
			Compression: unified.Storage.Compression,
		},
		Index: core.IndexConfig{
			DefaultType:   m.stringToIndexType(unified.Search.Index.DefaultType),
			DefaultMetric: m.stringToDistanceMetric(unified.Search.Index.DefaultMetric),
			HNSWConfig: core.HNSWConfig{
				M:              unified.Search.Index.HNSW.M,
				MaxM:           unified.Search.Index.HNSW.MaxM,
				MaxM0:          unified.Search.Index.HNSW.MaxM0,
				ML:             unified.Search.Index.HNSW.ML,
				EfConstruction: unified.Search.Index.HNSW.EfConstruction,
				EfSearch:       unified.Search.Index.HNSW.EfSearch,
				Seed:           unified.Search.Index.HNSW.Seed,
			},
			FlatConfig: core.FlatConfig{
				BatchSize: unified.Search.Index.Flat.BatchSize,
			},
		},
		Performance: core.PerfConfig{
			MaxConcurrency: unified.Performance.MaxConcurrency,
			EnableSIMD:     unified.Performance.EnableSIMD,
			MemoryLimit:    unified.Performance.MemoryLimit,
			GCTarget:       unified.Performance.GCTarget,
		},
	}
}

// Convert unified config to legacy embeddings config
func (m *MigrationAdapter) toEmbeddingsConfig(unified *VittoriaConfig) *embeddings.VectorizerConfig {
	return &embeddings.VectorizerConfig{
		Type:       m.stringToVectorizerType(unified.Embeddings.Default.Type),
		Model:      unified.Embeddings.Default.Model,
		Dimensions: unified.Embeddings.Default.Dimensions,
		Options:    unified.Embeddings.Default.Options,
	}
}

// Convert unified config to legacy processing config
func (m *MigrationAdapter) toProcessingConfig(unified *VittoriaConfig) *processor.ProcessingConfig {
	return &processor.ProcessingConfig{
		ChunkSize:    unified.Embeddings.Processing.ChunkSize,
		ChunkOverlap: unified.Embeddings.Processing.ChunkOverlap,
		MinChunkSize: unified.Embeddings.Processing.MinChunkSize,
		MaxChunkSize: unified.Embeddings.Processing.MaxChunkSize,
		Language:     "en", // Default language
	}
}

// Convert unified config to legacy parallel search config
func (m *MigrationAdapter) toParallelSearchConfig(unified *VittoriaConfig) *core.ParallelSearchConfig {
	return &core.ParallelSearchConfig{
		Enabled:        unified.Search.Parallel.Enabled,
		MaxWorkers:     unified.Search.Parallel.MaxWorkers,
		BatchSize:      unified.Search.Parallel.BatchSize,
		UseCache:       unified.Search.Parallel.UseCache,
		PreloadVectors: unified.Search.Parallel.PreloadVectors,
	}
}

// Convert unified config to legacy search cache config
func (m *MigrationAdapter) toSearchCacheConfig(unified *VittoriaConfig) *core.SearchCacheConfig {
	return &core.SearchCacheConfig{
		Enabled:         unified.Search.Cache.Enabled,
		MaxEntries:      unified.Search.Cache.MaxEntries,
		TTL:             unified.Search.Cache.TTL,
		CleanupInterval: unified.Search.Cache.CleanupInterval,
	}
}

// Convert legacy core config to unified config
func (m *MigrationAdapter) fromCoreConfig(legacy *core.Config, unified *VittoriaConfig) {
	unified.DataDir = legacy.DataDir
	unified.Server.Host = legacy.Server.Host
	unified.Server.Port = legacy.Server.Port
	unified.Server.ReadTimeout = legacy.Server.ReadTimeout
	unified.Server.WriteTimeout = legacy.Server.WriteTimeout
	unified.Server.MaxBodySize = legacy.Server.MaxBodySize
	unified.Server.CORS = legacy.Server.CORS

	unified.Storage.PageSize = legacy.Storage.PageSize
	unified.Storage.CacheSize = legacy.Storage.CacheSize
	unified.Storage.SyncWrites = legacy.Storage.SyncWrites
	unified.Storage.Compression = legacy.Storage.Compression

	unified.Search.Index.DefaultType = m.indexTypeToString(legacy.Index.DefaultType)
	unified.Search.Index.DefaultMetric = m.distanceMetricToString(legacy.Index.DefaultMetric)
	unified.Search.Index.HNSW.M = legacy.Index.HNSWConfig.M
	unified.Search.Index.HNSW.MaxM = legacy.Index.HNSWConfig.MaxM
	unified.Search.Index.HNSW.MaxM0 = legacy.Index.HNSWConfig.MaxM0
	unified.Search.Index.HNSW.ML = legacy.Index.HNSWConfig.ML
	unified.Search.Index.HNSW.EfConstruction = legacy.Index.HNSWConfig.EfConstruction
	unified.Search.Index.HNSW.EfSearch = legacy.Index.HNSWConfig.EfSearch
	unified.Search.Index.HNSW.Seed = legacy.Index.HNSWConfig.Seed
	unified.Search.Index.Flat.BatchSize = legacy.Index.FlatConfig.BatchSize

	unified.Performance.MaxConcurrency = legacy.Performance.MaxConcurrency
	unified.Performance.EnableSIMD = legacy.Performance.EnableSIMD
	unified.Performance.MemoryLimit = legacy.Performance.MemoryLimit
	unified.Performance.GCTarget = legacy.Performance.GCTarget
}

// Convert legacy embeddings config to unified config
func (m *MigrationAdapter) fromEmbeddingsConfig(legacy *embeddings.VectorizerConfig, unified *VittoriaConfig) {
	unified.Embeddings.Default.Type = m.vectorizerTypeToString(legacy.Type)
	unified.Embeddings.Default.Model = legacy.Model
	unified.Embeddings.Default.Dimensions = legacy.Dimensions
	unified.Embeddings.Default.Options = legacy.Options
}

// Convert legacy processing config to unified config
func (m *MigrationAdapter) fromProcessingConfig(legacy *processor.ProcessingConfig, unified *VittoriaConfig) {
	unified.Embeddings.Processing.ChunkSize = legacy.ChunkSize
	unified.Embeddings.Processing.ChunkOverlap = legacy.ChunkOverlap
	unified.Embeddings.Processing.MinChunkSize = legacy.MinChunkSize
	unified.Embeddings.Processing.MaxChunkSize = legacy.MaxChunkSize
	unified.Embeddings.Processing.Strategy = "smart" // Default strategy
}

// Convert legacy parallel search config to unified config
func (m *MigrationAdapter) fromParallelSearchConfig(legacy *core.ParallelSearchConfig, unified *VittoriaConfig) {
	unified.Search.Parallel.Enabled = legacy.Enabled
	unified.Search.Parallel.MaxWorkers = legacy.MaxWorkers
	unified.Search.Parallel.BatchSize = legacy.BatchSize
	unified.Search.Parallel.UseCache = legacy.UseCache
	unified.Search.Parallel.PreloadVectors = legacy.PreloadVectors
}

// Convert legacy search cache config to unified config
func (m *MigrationAdapter) fromSearchCacheConfig(legacy *core.SearchCacheConfig, unified *VittoriaConfig) {
	unified.Search.Cache.Enabled = legacy.Enabled
	unified.Search.Cache.MaxEntries = legacy.MaxEntries
	unified.Search.Cache.TTL = legacy.TTL
	unified.Search.Cache.CleanupInterval = legacy.CleanupInterval
}

// Helper functions for type conversions
func (m *MigrationAdapter) stringToIndexType(s string) core.IndexType {
	switch s {
	case "flat":
		return core.IndexTypeFlat
	case "hnsw":
		return core.IndexTypeHNSW
	case "ivf":
		return core.IndexTypeIVF
	default:
		return core.IndexTypeFlat
	}
}

func (m *MigrationAdapter) indexTypeToString(t core.IndexType) string {
	switch t {
	case core.IndexTypeFlat:
		return "flat"
	case core.IndexTypeHNSW:
		return "hnsw"
	case core.IndexTypeIVF:
		return "ivf"
	default:
		return "flat"
	}
}

func (m *MigrationAdapter) stringToDistanceMetric(s string) core.DistanceMetric {
	switch s {
	case "cosine":
		return core.DistanceMetricCosine
	case "euclidean":
		return core.DistanceMetricEuclidean
	case "dot_product":
		return core.DistanceMetricDotProduct
	case "manhattan":
		return core.DistanceMetricManhattan
	default:
		return core.DistanceMetricCosine
	}
}

func (m *MigrationAdapter) distanceMetricToString(metric core.DistanceMetric) string {
	switch metric {
	case core.DistanceMetricCosine:
		return "cosine"
	case core.DistanceMetricEuclidean:
		return "euclidean"
	case core.DistanceMetricDotProduct:
		return "dot_product"
	case core.DistanceMetricManhattan:
		return "manhattan"
	default:
		return "cosine"
	}
}

func (m *MigrationAdapter) stringToVectorizerType(s string) embeddings.VectorizerType {
	switch s {
	case "sentence_transformers":
		return embeddings.VectorizerTypeSentenceTransformers
	case "openai":
		return embeddings.VectorizerTypeOpenAI
	case "huggingface":
		return embeddings.VectorizerTypeHuggingFace
	case "ollama":
		return embeddings.VectorizerTypeOllama
	default:
		return embeddings.VectorizerTypeSentenceTransformers
	}
}

func (m *MigrationAdapter) vectorizerTypeToString(vType embeddings.VectorizerType) string {
	switch vType {
	case embeddings.VectorizerTypeSentenceTransformers:
		return "sentence_transformers"
	case embeddings.VectorizerTypeOpenAI:
		return "openai"
	case embeddings.VectorizerTypeHuggingFace:
		return "huggingface"
	case embeddings.VectorizerTypeOllama:
		return "ollama"
	default:
		return "sentence_transformers"
	}
}

// ConfigMigrator provides high-level migration utilities
type ConfigMigrator struct {
	adapter *MigrationAdapter
}

// NewConfigMigrator creates a new config migrator
func NewConfigMigrator() *ConfigMigrator {
	return &ConfigMigrator{
		adapter: NewMigrationAdapter(),
	}
}

// MigrateToUnified migrates existing scattered configurations to unified format
func (cm *ConfigMigrator) MigrateToUnified(
	coreConfig *core.Config,
	embeddingsConfig *embeddings.VectorizerConfig,
	processingConfig *processor.ProcessingConfig,
) (*VittoriaConfig, error) {

	legacy := &LegacyConfigBundle{
		Core:       coreConfig,
		Embeddings: embeddingsConfig,
		Processing: processingConfig,
	}

	unified := cm.adapter.FromLegacyConfig(legacy)

	// Validate the migrated configuration
	if err := unified.Validate(); err != nil {
		return nil, fmt.Errorf("migrated configuration validation failed: %w", err)
	}

	return unified, nil
}

// MigrateFromUnified converts unified config back to legacy formats
func (cm *ConfigMigrator) MigrateFromUnified(unified *VittoriaConfig) *LegacyConfigBundle {
	return cm.adapter.ToLegacyConfig(unified)
}

// ValidateMigration validates that a migration preserves essential settings
func (cm *ConfigMigrator) ValidateMigration(original, migrated *VittoriaConfig) error {
	var errors []string

	// Check critical settings preservation
	if original.Server.Port != migrated.Server.Port {
		errors = append(errors, fmt.Sprintf("server port changed: %d -> %d",
			original.Server.Port, migrated.Server.Port))
	}

	if original.DataDir != migrated.DataDir {
		errors = append(errors, fmt.Sprintf("data directory changed: %s -> %s",
			original.DataDir, migrated.DataDir))
	}

	if original.Embeddings.Default.Dimensions != migrated.Embeddings.Default.Dimensions {
		errors = append(errors, fmt.Sprintf("embedding dimensions changed: %d -> %d",
			original.Embeddings.Default.Dimensions, migrated.Embeddings.Default.Dimensions))
	}

	if len(errors) > 0 {
		return fmt.Errorf("migration validation failed:\n- %s",
			fmt.Sprintf("%s", errors))
	}

	return nil
}
