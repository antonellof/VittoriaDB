package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/core"
	"github.com/antonellof/VittoriaDB/pkg/server"
	"github.com/urfave/cli/v2"
)

// Build information (set via ldflags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GitTag    = "unknown"
)

func main() {
	app := &cli.App{
		Name:    "vittoriadb",
		Usage:   "Simple embedded vector database",
		Version: Version,
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "Show version information",
				Action: func(c *cli.Context) error {
					fmt.Printf("VittoriaDB %s\n", Version)
					fmt.Printf("Build Time: %s\n", BuildTime)
					fmt.Printf("Git Commit: %s\n", GitCommit)
					fmt.Printf("Git Tag: %s\n", GitTag)
					return nil
				},
			},
			{
				Name:  "info",
				Usage: "Show database information and file locations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "data-dir",
						Value:   "./data",
						Usage:   "Data directory path",
						EnvVars: []string{"VITTORIADB_DATA_DIR"},
					},
				},
				Action: showDatabaseInfo,
			},
			{
				Name:  "run",
				Usage: "Start the VittoriaDB server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "host",
						Value:   "localhost",
						Usage:   "Host to bind to",
						EnvVars: []string{"VITTORIADB_HOST"},
					},
					&cli.IntFlag{
						Name:    "port",
						Value:   8080,
						Usage:   "Port to listen on",
						EnvVars: []string{"VITTORIADB_PORT"},
					},
					&cli.StringFlag{
						Name:    "data-dir",
						Value:   "./data",
						Usage:   "Data directory path",
						EnvVars: []string{"VITTORIADB_DATA_DIR"},
					},
					&cli.StringFlag{
						Name:    "config",
						Value:   "",
						Usage:   "Configuration file path",
						EnvVars: []string{"VITTORIADB_CONFIG"},
					},
					&cli.BoolFlag{
						Name:  "cors",
						Value: true,
						Usage: "Enable CORS headers",
					},
				},
				Action: runServer,
			},
			{
				Name:  "create",
				Usage: "Create a new collection",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Collection name",
						Required: true,
					},
					&cli.IntFlag{
						Name:     "dimensions",
						Aliases:  []string{"dim"},
						Usage:    "Vector dimensions",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "metric",
						Value: "cosine",
						Usage: "Distance metric (cosine, euclidean, dot_product, manhattan)",
					},
					&cli.StringFlag{
						Name:  "index",
						Value: "flat",
						Usage: "Index type (flat, hnsw)",
					},
					&cli.StringFlag{
						Name:  "data-dir",
						Value: "./data",
						Usage: "Data directory path",
					},
				},
				Action: createCollection,
			},
			{
				Name:  "stats",
				Usage: "Show database statistics",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "data-dir",
						Value: "./data",
						Usage: "Data directory path",
					},
				},
				Action: showStats,
			},
			{
				Name:  "backup",
				Usage: "Backup database",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "data-dir",
						Value: "./data",
						Usage: "Data directory path",
					},
					&cli.StringFlag{
						Name:     "output",
						Usage:    "Output backup file",
						Required: true,
					},
				},
				Action: backupDatabase,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runServer(c *cli.Context) error {
	// Create database configuration
	config := &core.Config{
		DataDir: c.String("data-dir"),
		Server: core.ServerConfig{
			Host:         c.String("host"),
			Port:         c.Int("port"),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			MaxBodySize:  10 << 20, // 10MB
			CORS:         c.Bool("cors"),
		},
		Storage: core.StorageConfig{
			PageSize:    4096,
			CacheSize:   100,
			SyncWrites:  true,
			Compression: false,
		},
		Index: core.IndexConfig{
			DefaultType:   core.IndexTypeFlat,
			DefaultMetric: core.DistanceMetricCosine,
		},
		Performance: core.PerfConfig{
			MaxConcurrency: 100,
			EnableSIMD:     true,
			MemoryLimit:    1 << 30, // 1GB
			GCTarget:       100,
		},
	}

	// Create and open database
	db := core.NewDatabase()
	ctx := context.Background()

	if err := db.Open(ctx, config); err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create server configuration
	serverConfig := &server.ServerConfig{
		Host:         config.Server.Host,
		Port:         config.Server.Port,
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		MaxBodySize:  config.Server.MaxBodySize,
		CORS:         config.Server.CORS,
	}

	// Create and start server
	srv := server.NewServer(db, serverConfig)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Received shutdown signal...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown server
		if err := srv.Stop(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}

		// Close database
		if err := db.Close(); err != nil {
			log.Printf("Database close error: %v", err)
		}

		os.Exit(0)
	}()

	// Get absolute path for data directory
	absDataDir, err := filepath.Abs(config.DataDir)
	if err != nil {
		absDataDir = config.DataDir
	}

	// Enhanced startup information
	log.Printf("🚀 VittoriaDB %s starting...", Version)
	log.Printf("📁 Data directory: %s", absDataDir)
	log.Printf("🌐 HTTP server: http://%s:%d", config.Server.Host, config.Server.Port)
	log.Printf("📊 Web dashboard: http://%s:%d/", config.Server.Host, config.Server.Port)
	log.Printf("⚙️  Configuration:")
	log.Printf("   • Index type: %s", config.Index.DefaultType)
	log.Printf("   • Distance metric: %s", config.Index.DefaultMetric)
	log.Printf("   • Page size: %d bytes", config.Storage.PageSize)
	log.Printf("   • Cache size: %d pages", config.Storage.CacheSize)
	log.Printf("   • CORS enabled: %t", config.Server.CORS)
	log.Printf("   • Max concurrency: %d", config.Performance.MaxConcurrency)
	log.Printf("   • Memory limit: %d MB", config.Performance.MemoryLimit/(1024*1024))

	// Start server (blocking)
	if err := srv.Start(); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func createCollection(c *cli.Context) error {
	// Parse metric
	var metric core.DistanceMetric
	switch c.String("metric") {
	case "cosine":
		metric = core.DistanceMetricCosine
	case "euclidean":
		metric = core.DistanceMetricEuclidean
	case "dot_product":
		metric = core.DistanceMetricDotProduct
	case "manhattan":
		metric = core.DistanceMetricManhattan
	default:
		return fmt.Errorf("invalid metric: %s", c.String("metric"))
	}

	// Parse index type
	var indexType core.IndexType
	switch c.String("index") {
	case "flat":
		indexType = core.IndexTypeFlat
	case "hnsw":
		indexType = core.IndexTypeHNSW
	default:
		return fmt.Errorf("invalid index type: %s", c.String("index"))
	}

	// Create database configuration
	config := &core.Config{
		DataDir: c.String("data-dir"),
	}

	// Create and open database
	db := core.NewDatabase()
	ctx := context.Background()

	if err := db.Open(ctx, config); err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create collection
	req := &core.CreateCollectionRequest{
		Name:       c.String("name"),
		Dimensions: c.Int("dimensions"),
		Metric:     metric,
		IndexType:  indexType,
	}

	if err := db.CreateCollection(ctx, req); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	fmt.Printf("Collection '%s' created successfully\n", req.Name)
	fmt.Printf("  Dimensions: %d\n", req.Dimensions)
	fmt.Printf("  Metric: %s\n", metric.String())
	fmt.Printf("  Index: %s\n", indexType.String())

	return nil
}

func showStats(c *cli.Context) error {
	// Create database configuration
	config := &core.Config{
		DataDir: c.String("data-dir"),
	}

	// Create and open database
	db := core.NewDatabase()
	ctx := context.Background()

	if err := db.Open(ctx, config); err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get stats
	stats, err := db.Stats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	// Display stats
	fmt.Println("VittoriaDB Statistics")
	fmt.Println("====================")
	fmt.Printf("Total Vectors: %d\n", stats.TotalVectors)
	fmt.Printf("Total Size: %d bytes\n", stats.TotalSize)
	fmt.Printf("Index Size: %d bytes\n", stats.IndexSize)
	fmt.Printf("Collections: %d\n", len(stats.Collections))
	fmt.Println()

	if len(stats.Collections) > 0 {
		fmt.Println("Collections:")
		for _, collection := range stats.Collections {
			fmt.Printf("  %s:\n", collection.Name)
			fmt.Printf("    Vectors: %d\n", collection.VectorCount)
			fmt.Printf("    Dimensions: %d\n", collection.Dimensions)
			fmt.Printf("    Index: %s\n", collection.IndexType.String())
			fmt.Printf("    Last Modified: %s\n", collection.LastModified.Format(time.RFC3339))
			fmt.Println()
		}
	}

	return nil
}

func backupDatabase(c *cli.Context) error {
	// TODO: Implement backup functionality
	return fmt.Errorf("backup functionality not implemented yet")
}

func showDatabaseInfo(c *cli.Context) error {
	dataDir := c.String("data-dir")

	// Get absolute path
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		absDataDir = dataDir
	}

	fmt.Printf("🚀 VittoriaDB %s - Database Information\n", Version)
	fmt.Println("=====================================")
	fmt.Printf("📁 Data Directory: %s\n", absDataDir)
	fmt.Printf("📍 Relative Path: %s\n", dataDir)
	fmt.Println()

	// Check if data directory exists
	if _, err := os.Stat(absDataDir); os.IsNotExist(err) {
		fmt.Printf("❌ Data directory does not exist\n")
		fmt.Printf("   Run 'vittoriadb run' to create it automatically\n")
		return nil
	}

	// List collections
	entries, err := os.ReadDir(absDataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %w", err)
	}

	collections := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			collections = append(collections, entry.Name())
		}
	}

	if len(collections) == 0 {
		fmt.Printf("📂 Collections: None found\n")
		fmt.Printf("   Create collections using the API or CLI\n")
	} else {
		fmt.Printf("📂 Collections (%d found):\n", len(collections))
		for _, collection := range collections {
			collectionPath := filepath.Join(absDataDir, collection)
			fmt.Printf("   • %s/\n", collection)

			// List files in collection directory
			collectionEntries, err := os.ReadDir(collectionPath)
			if err != nil {
				fmt.Printf("     ❌ Error reading collection: %v\n", err)
				continue
			}

			for _, file := range collectionEntries {
				if !file.IsDir() {
					info, err := file.Info()
					if err != nil {
						continue
					}
					size := info.Size()
					sizeStr := formatFileSize(size)
					fmt.Printf("     - %s (%s)\n", file.Name(), sizeStr)
				}
			}
		}
	}

	fmt.Println()
	fmt.Printf("💡 Configuration Options:\n")
	fmt.Printf("   --data-dir <path>              Set custom data directory\n")
	fmt.Printf("   VITTORIADB_DATA_DIR=<path>     Environment variable\n")
	fmt.Println()
	fmt.Printf("🔧 Management Commands:\n")
	fmt.Printf("   vittoriadb run --data-dir <path>    Start server with custom directory\n")
	fmt.Printf("   vittoriadb stats --data-dir <path>  Show database statistics\n")

	return nil
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
