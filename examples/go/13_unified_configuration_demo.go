package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/antonellof/VittoriaDB/pkg/config"
)

func main() {
	fmt.Println("🔧 VittoriaDB Unified Configuration System Demo")
	fmt.Println("==============================================")

	// Demo 1: Load default configuration
	fmt.Println("\n📋 1. Loading Default Configuration:")
	defaultConfig := config.DefaultConfig()
	fmt.Printf("   • Server: %s:%d\n", defaultConfig.Server.Host, defaultConfig.Server.Port)
	fmt.Printf("   • Data Directory: %s\n", defaultConfig.DataDir)
	fmt.Printf("   • Parallel Search: %t (workers: %d)\n",
		defaultConfig.Search.Parallel.Enabled, defaultConfig.Search.Parallel.MaxWorkers)
	fmt.Printf("   • Search Cache: %t (entries: %d)\n",
		defaultConfig.Search.Cache.Enabled, defaultConfig.Search.Cache.MaxEntries)
	fmt.Printf("   • Memory-mapped I/O: %t\n", defaultConfig.Performance.IO.UseMemoryMap)

	// Demo 2: Load configuration from multiple sources
	fmt.Println("\n🔄 2. Loading Configuration from Multiple Sources:")

	// Set some environment variables for demo
	os.Setenv("VITTORIA_PORT", "9090")
	os.Setenv("VITTORIA_LOG_LEVEL", "debug")
	os.Setenv("VITTORIA_SEARCH_PARALLEL_MAX_WORKERS", "16")

	// Create flags map
	flags := map[string]string{
		"host":     "localhost", // Use localhost to avoid security warning
		"data-dir": "/tmp/vittoria-demo",
	}

	// Load with precedence: defaults < env vars < flags
	unifiedConfig, err := config.LoadConfigWithOverrides("", "VITTORIA_", flags)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("   • Server: %s:%d (host from flags, port from env)\n",
		unifiedConfig.Server.Host, unifiedConfig.Server.Port)
	fmt.Printf("   • Data Directory: %s (from flags)\n", unifiedConfig.DataDir)
	fmt.Printf("   • Log Level: %s (from env)\n", unifiedConfig.Logging.Level)
	fmt.Printf("   • Parallel Workers: %d (from env)\n", unifiedConfig.Search.Parallel.MaxWorkers)

	// Demo 3: Configuration validation
	fmt.Println("\n✅ 3. Configuration Validation:")
	if err := unifiedConfig.Validate(); err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Configuration is valid\n")
	}

	// Demo 4: Configuration manager with hot-reloading
	fmt.Println("\n🔄 4. Configuration Manager with Validation:")

	manager := config.CreateDefaultManager(
		config.FromDefaults(),
		config.FromEnv("VITTORIA_"),
		config.FromFlags(flags),
	)

	if err := manager.Load(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	currentConfig := manager.Get()
	fmt.Printf("   • Loaded configuration from: %s\n", currentConfig.Source)
	fmt.Printf("   • Performance settings:\n")
	fmt.Printf("     - Max concurrency: %d\n", currentConfig.Performance.MaxConcurrency)
	fmt.Printf("     - SIMD enabled: %t\n", currentConfig.Performance.EnableSIMD)
	fmt.Printf("     - Memory-mapped I/O: %t\n", currentConfig.Performance.IO.UseMemoryMap)
	fmt.Printf("     - Async I/O: %t\n", currentConfig.Performance.IO.AsyncIO)

	// Demo 5: Configuration updates
	fmt.Println("\n🔧 5. Dynamic Configuration Updates:")

	err = manager.Update(func(cfg *config.VittoriaConfig) error {
		cfg.Search.Parallel.MaxWorkers = 32
		cfg.Search.Cache.MaxEntries = 2000
		cfg.Performance.IO.VectorizedOps = true
		return nil
	})

	if err != nil {
		fmt.Printf("   ❌ Update failed: %v\n", err)
	} else {
		updatedConfig := manager.Get()
		fmt.Printf("   ✅ Configuration updated successfully:\n")
		fmt.Printf("     - Parallel workers: %d → %d\n",
			currentConfig.Search.Parallel.MaxWorkers, updatedConfig.Search.Parallel.MaxWorkers)
		fmt.Printf("     - Cache entries: %d → %d\n",
			currentConfig.Search.Cache.MaxEntries, updatedConfig.Search.Cache.MaxEntries)
		fmt.Printf("     - Vectorized ops: %t → %t\n",
			currentConfig.Performance.IO.VectorizedOps, updatedConfig.Performance.IO.VectorizedOps)
	}

	// Demo 6: Legacy configuration migration
	fmt.Println("\n🔄 6. Legacy Configuration Migration:")

	migrator := config.NewConfigMigrator()
	legacyBundle := migrator.MigrateFromUnified(unifiedConfig)

	fmt.Printf("   • Converted to legacy format:\n")
	fmt.Printf("     - Core config data dir: %s\n", legacyBundle.Core.DataDir)
	fmt.Printf("     - Core config server port: %d\n", legacyBundle.Core.Server.Port)
	fmt.Printf("     - Embeddings config type: %s\n", legacyBundle.Embeddings.Type.String())
	fmt.Printf("     - Embeddings config model: %s\n", legacyBundle.Embeddings.Model)
	fmt.Printf("     - Processing config chunk size: %d\n", legacyBundle.Processing.ChunkSize)

	// Convert back to unified
	reconverted := migrator.MigrateFromUnified(unifiedConfig)
	fmt.Printf("   • Round-trip conversion successful: %t\n",
		reconverted.Core.DataDir == unifiedConfig.DataDir)

	// Demo 7: Export configuration to YAML
	fmt.Println("\n📄 7. Configuration Export:")

	yamlData, err := unifiedConfig.ToYAML()
	if err != nil {
		fmt.Printf("   ❌ YAML export failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Configuration exported to YAML (%d bytes)\n", len(yamlData))

		// Show a snippet
		lines := strings.Split(string(yamlData), "\n")
		fmt.Printf("   📝 Sample (first 5 lines):\n")
		for i, line := range lines[:min(5, len(lines))] {
			fmt.Printf("      %d: %s\n", i+1, line)
		}
	}

	// Demo 8: Performance and resource validation
	fmt.Println("\n⚡ 8. Performance and Resource Validation:")

	// Create a performance validator
	perfValidator := &config.PerformanceValidator{}
	if err := perfValidator.Validate(unifiedConfig); err != nil {
		fmt.Printf("   ⚠️  Performance warnings: %v\n", err)
	} else {
		fmt.Printf("   ✅ Performance configuration is optimal\n")
	}

	// Create a resource validator
	resourceValidator := &config.ResourceValidator{}
	if err := resourceValidator.Validate(unifiedConfig); err != nil {
		fmt.Printf("   ⚠️  Resource warnings: %v\n", err)
	} else {
		fmt.Printf("   ✅ Resource configuration is within limits\n")
	}

	// Demo 9: Configuration recommendations
	fmt.Println("\n💡 9. Configuration Recommendations:")

	fmt.Printf("   📊 Current Settings Analysis:\n")
	fmt.Printf("     • Vector dimensions: %d (affects memory usage)\n",
		unifiedConfig.Embeddings.Default.Dimensions)
	fmt.Printf("     • Batch size: %d (affects throughput)\n",
		unifiedConfig.Embeddings.Batch.DefaultBatchSize)
	fmt.Printf("     • Cache TTL: %s (affects hit rate)\n",
		unifiedConfig.Search.Cache.TTL)

	fmt.Printf("   🎯 Recommendations:\n")
	if unifiedConfig.Search.Parallel.MaxWorkers > 50 {
		fmt.Printf("     • Consider reducing parallel workers (current: %d) to avoid context switching\n",
			unifiedConfig.Search.Parallel.MaxWorkers)
	}
	if unifiedConfig.Search.Cache.MaxEntries < 100 {
		fmt.Printf("     • Consider increasing cache size (current: %d) for better performance\n",
			unifiedConfig.Search.Cache.MaxEntries)
	}
	if !unifiedConfig.Performance.IO.UseMemoryMap {
		fmt.Printf("     • Enable memory-mapped I/O for better performance\n")
	}
	if !unifiedConfig.Performance.EnableSIMD {
		fmt.Printf("     • Enable SIMD optimizations for faster vector operations\n")
	}

	// Cleanup
	manager.Stop()
	os.Unsetenv("VITTORIA_PORT")
	os.Unsetenv("VITTORIA_LOG_LEVEL")
	os.Unsetenv("VITTORIA_SEARCH_PARALLEL_MAX_WORKERS")

	fmt.Println("\n🎉 Configuration demo completed successfully!")
	fmt.Println("\n📚 Next Steps:")
	fmt.Println("   • Use 'vittoriadb config generate' to create your own config file")
	fmt.Println("   • Use 'vittoriadb config validate' to check your configuration")
	fmt.Println("   • Use 'vittoriadb config env --list' to see all environment variables")
	fmt.Println("   • Use 'vittoriadb run --config your-config.yaml' to start with custom config")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to demonstrate configuration scenarios
func demonstrateConfigScenarios() {
	fmt.Println("\n🎭 Configuration Scenarios:")

	scenarios := []struct {
		name        string
		description string
		envVars     map[string]string
		flags       map[string]string
	}{
		{
			name:        "Development",
			description: "Local development with debug logging",
			envVars: map[string]string{
				"VITTORIA_LOG_LEVEL":              "debug",
				"VITTORIA_SEARCH_CACHE_ENABLED":   "false",
				"VITTORIA_PERF_IO_USE_MEMORY_MAP": "false",
			},
			flags: map[string]string{
				"host":     "localhost",
				"port":     "8080",
				"data-dir": "./dev-data",
			},
		},
		{
			name:        "Production",
			description: "Production deployment with optimizations",
			envVars: map[string]string{
				"VITTORIA_LOG_LEVEL":                   "info",
				"VITTORIA_SEARCH_PARALLEL_MAX_WORKERS": "32",
				"VITTORIA_SEARCH_CACHE_MAX_ENTRIES":    "10000",
				"VITTORIA_PERF_IO_USE_MEMORY_MAP":      "true",
				"VITTORIA_PERF_ENABLE_SIMD":            "true",
			},
			flags: map[string]string{
				"host":     "0.0.0.0",
				"port":     "8080",
				"data-dir": "/var/lib/vittoriadb",
			},
		},
		{
			name:        "High-Performance",
			description: "Maximum performance configuration",
			envVars: map[string]string{
				"VITTORIA_SEARCH_PARALLEL_MAX_WORKERS":   "64",
				"VITTORIA_SEARCH_CACHE_MAX_ENTRIES":      "50000",
				"VITTORIA_EMBEDDINGS_BATCH_DEFAULT_SIZE": "128",
				"VITTORIA_PERF_IO_VECTORIZED_OPS":        "true",
				"VITTORIA_PERF_CPU_PARALLEL_COMPUTE":     "true",
			},
			flags: map[string]string{
				"cache-size": "10000",
			},
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n   🎯 %s Scenario: %s\n", scenario.name, scenario.description)

		// Set environment variables
		for key, value := range scenario.envVars {
			os.Setenv(key, value)
		}

		// Load configuration
		cfg, err := config.LoadConfigWithOverrides("", "VITTORIA_", scenario.flags)
		if err != nil {
			fmt.Printf("      ❌ Failed to load: %v\n", err)
			continue
		}

		fmt.Printf("      • Log level: %s\n", cfg.Logging.Level)
		fmt.Printf("      • Parallel workers: %d\n", cfg.Search.Parallel.MaxWorkers)
		fmt.Printf("      • Cache entries: %d\n", cfg.Search.Cache.MaxEntries)
		fmt.Printf("      • Memory-mapped I/O: %t\n", cfg.Performance.IO.UseMemoryMap)

		// Cleanup environment variables
		for key := range scenario.envVars {
			os.Unsetenv(key)
		}
	}
}
