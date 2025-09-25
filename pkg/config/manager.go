package config

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ConfigManager handles configuration lifecycle, validation, and hot-reloading
type ConfigManager struct {
	current    *VittoriaConfig
	sources    []ConfigSource
	validators []Validator
	listeners  []ChangeListener
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	reloadCh   chan struct{}
}

// Validator represents a configuration validator
type Validator interface {
	Validate(config *VittoriaConfig) error
	Name() string
}

// ChangeListener represents a configuration change listener
type ChangeListener interface {
	OnConfigChange(old, new *VittoriaConfig) error
	Name() string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(sources ...ConfigSource) *ConfigManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConfigManager{
		sources:    sources,
		validators: make([]Validator, 0),
		listeners:  make([]ChangeListener, 0),
		ctx:        ctx,
		cancel:     cancel,
		reloadCh:   make(chan struct{}, 1),
	}
}

// Load loads the initial configuration
func (cm *ConfigManager) Load() error {
	config, err := LoadConfig(cm.sources...)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Run additional validators
	for _, validator := range cm.validators {
		if err := validator.Validate(config); err != nil {
			return fmt.Errorf("validation failed (%s): %w", validator.Name(), err)
		}
	}

	cm.mu.Lock()
	cm.current = config
	cm.mu.Unlock()

	return nil
}

// Get returns the current configuration (thread-safe)
func (cm *ConfigManager) Get() *VittoriaConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.current == nil {
		return DefaultConfig()
	}

	return cm.current.Clone()
}

// Reload reloads the configuration from sources
func (cm *ConfigManager) Reload() error {
	newConfig, err := LoadConfig(cm.sources...)
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	// Run additional validators
	for _, validator := range cm.validators {
		if err := validator.Validate(newConfig); err != nil {
			return fmt.Errorf("validation failed (%s): %w", validator.Name(), err)
		}
	}

	cm.mu.Lock()
	oldConfig := cm.current
	cm.current = newConfig
	cm.mu.Unlock()

	// Notify listeners of configuration change
	for _, listener := range cm.listeners {
		if err := listener.OnConfigChange(oldConfig, newConfig); err != nil {
			// Log error but don't fail the reload
			fmt.Printf("Config change listener %s failed: %v\n", listener.Name(), err)
		}
	}

	return nil
}

// Update updates specific configuration values
func (cm *ConfigManager) Update(updateFn func(*VittoriaConfig) error) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Clone current config
	newConfig := cm.current.Clone()

	// Apply update
	if err := updateFn(newConfig); err != nil {
		return fmt.Errorf("failed to apply config update: %w", err)
	}

	// Validate updated config
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("updated config validation failed: %w", err)
	}

	// Run additional validators
	for _, validator := range cm.validators {
		if err := validator.Validate(newConfig); err != nil {
			return fmt.Errorf("validation failed (%s): %w", validator.Name(), err)
		}
	}

	oldConfig := cm.current
	cm.current = newConfig

	// Notify listeners of configuration change
	for _, listener := range cm.listeners {
		if err := listener.OnConfigChange(oldConfig, newConfig); err != nil {
			// Log error but don't fail the update
			fmt.Printf("Config change listener %s failed: %v\n", listener.Name(), err)
		}
	}

	return nil
}

// AddValidator adds a configuration validator
func (cm *ConfigManager) AddValidator(validator Validator) {
	cm.validators = append(cm.validators, validator)
}

// AddChangeListener adds a configuration change listener
func (cm *ConfigManager) AddChangeListener(listener ChangeListener) {
	cm.listeners = append(cm.listeners, listener)
}

// StartWatching starts watching for configuration changes (file-based sources)
func (cm *ConfigManager) StartWatching(interval time.Duration) {
	go cm.watchLoop(interval)
}

// Stop stops the configuration manager
func (cm *ConfigManager) Stop() {
	cm.cancel()
}

func (cm *ConfigManager) watchLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			if err := cm.Reload(); err != nil {
				// Log error but continue watching
				fmt.Printf("Config reload failed: %v\n", err)
			}
		case <-cm.reloadCh:
			if err := cm.Reload(); err != nil {
				fmt.Printf("Manual config reload failed: %v\n", err)
			}
		}
	}
}

// TriggerReload triggers a manual configuration reload
func (cm *ConfigManager) TriggerReload() {
	select {
	case cm.reloadCh <- struct{}{}:
	default:
		// Channel is full, reload already pending
	}
}

// Built-in validators

// PerformanceValidator validates performance-related configuration
type PerformanceValidator struct{}

func (pv *PerformanceValidator) Name() string {
	return "performance"
}

func (pv *PerformanceValidator) Validate(config *VittoriaConfig) error {
	var errors []string

	// Check for performance anti-patterns
	if config.Search.Parallel.MaxWorkers > 100 {
		errors = append(errors, "search.parallel.max_workers > 100 may cause excessive context switching")
	}

	if config.Storage.CacheSize > 100000 {
		errors = append(errors, "storage.cache_size > 100000 may cause excessive memory usage")
	}

	if config.Search.Cache.MaxEntries > 1000000 {
		errors = append(errors, "search.cache.max_entries > 1000000 may cause excessive memory usage")
	}

	if config.Embeddings.Batch.MaxBatchSize > 1000 {
		errors = append(errors, "embeddings.batch.max_batch_size > 1000 may cause memory issues")
	}

	if len(errors) > 0 {
		return fmt.Errorf("performance validation warnings:\n- %s",
			fmt.Sprintf("%s", errors))
	}

	return nil
}

// SecurityValidator validates security-related configuration
type SecurityValidator struct{}

func (sv *SecurityValidator) Name() string {
	return "security"
}

func (sv *SecurityValidator) Validate(config *VittoriaConfig) error {
	var errors []string

	// Check for security issues
	if config.Server.Host == "0.0.0.0" && !config.Server.TLS.Enabled {
		errors = append(errors, "binding to 0.0.0.0 without TLS is insecure")
	}

	if config.Embeddings.OpenAI.APIKey == "" && config.Embeddings.Default.Type == "openai" {
		errors = append(errors, "OpenAI API key is required when using OpenAI embeddings")
	}

	if config.Server.MaxBodySize > 1<<30 { // 1GB
		errors = append(errors, "server.max_body_size > 1GB may enable DoS attacks")
	}

	if len(errors) > 0 {
		return fmt.Errorf("security validation warnings:\n- %s",
			fmt.Sprintf("%s", errors))
	}

	return nil
}

// ResourceValidator validates resource-related configuration
type ResourceValidator struct{}

func (rv *ResourceValidator) Name() string {
	return "resource"
}

func (rv *ResourceValidator) Validate(config *VittoriaConfig) error {
	var errors []string

	// Estimate memory usage
	estimatedMemory := int64(0)

	// Storage cache memory
	estimatedMemory += int64(config.Storage.CacheSize) * int64(config.Storage.PageSize)

	// Search cache memory (rough estimate)
	estimatedMemory += int64(config.Search.Cache.MaxEntries) * 1024 // 1KB per entry estimate

	// Vector memory (rough estimate based on common dimensions)
	vectorMemory := int64(config.Embeddings.Default.Dimensions) * 4 * 1000 // 1000 vectors estimate
	estimatedMemory += vectorMemory

	if config.Performance.MemoryLimit > 0 && estimatedMemory > config.Performance.MemoryLimit {
		errors = append(errors, fmt.Sprintf("estimated memory usage (%d bytes) exceeds limit (%d bytes)",
			estimatedMemory, config.Performance.MemoryLimit))
	}

	// Check CPU configuration
	if config.Performance.MaxConcurrency > config.Performance.CPU.NumThreads*4 {
		errors = append(errors, "performance.max_concurrency is much higher than available CPU threads")
	}

	if len(errors) > 0 {
		return fmt.Errorf("resource validation warnings:\n- %s",
			fmt.Sprintf("%s", errors))
	}

	return nil
}

// Built-in change listeners

// LoggingChangeListener handles logging configuration changes
type LoggingChangeListener struct{}

func (lcl *LoggingChangeListener) Name() string {
	return "logging"
}

func (lcl *LoggingChangeListener) OnConfigChange(old, new *VittoriaConfig) error {
	if old == nil || old.Logging.Level != new.Logging.Level {
		fmt.Printf("Log level changed to: %s\n", new.Logging.Level)
		// Here you would typically reconfigure the logger
	}

	if old == nil || old.Logging.Format != new.Logging.Format {
		fmt.Printf("Log format changed to: %s\n", new.Logging.Format)
		// Here you would typically reconfigure the logger format
	}

	return nil
}

// CacheChangeListener handles cache configuration changes
type CacheChangeListener struct{}

func (ccl *CacheChangeListener) Name() string {
	return "cache"
}

func (ccl *CacheChangeListener) OnConfigChange(old, new *VittoriaConfig) error {
	if old == nil {
		return nil
	}

	// Check if cache settings changed
	if old.Search.Cache.MaxEntries != new.Search.Cache.MaxEntries ||
		old.Search.Cache.TTL != new.Search.Cache.TTL {
		fmt.Printf("Search cache configuration changed, clearing cache\n")
		// Here you would typically clear and reconfigure the cache
	}

	if old.Storage.CacheSize != new.Storage.CacheSize {
		fmt.Printf("Storage cache size changed from %d to %d\n",
			old.Storage.CacheSize, new.Storage.CacheSize)
		// Here you would typically resize the storage cache
	}

	return nil
}

// CreateDefaultManager creates a configuration manager with default validators and listeners
func CreateDefaultManager(sources ...ConfigSource) *ConfigManager {
	manager := NewConfigManager(sources...)

	// Add default validators
	manager.AddValidator(&PerformanceValidator{})
	manager.AddValidator(&SecurityValidator{})
	manager.AddValidator(&ResourceValidator{})

	// Add default change listeners
	manager.AddChangeListener(&LoggingChangeListener{})
	manager.AddChangeListener(&CacheChangeListener{})

	return manager
}
