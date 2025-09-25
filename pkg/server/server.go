package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/config"
	"github.com/antonellof/VittoriaDB/pkg/core"
	"github.com/antonellof/VittoriaDB/pkg/processor"
	"github.com/gorilla/mux"
)

// Server represents the HTTP API server
type Server struct {
	db            core.Database
	router        *mux.Router
	server        *http.Server
	config        *ServerConfig
	unifiedConfig *config.VittoriaConfig
	processor     *processor.ProcessorFactory
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	MaxBodySize  int64
	CORS         bool
}

// NewServer creates a new HTTP server
func NewServer(db core.Database, config *ServerConfig, unifiedConfig *config.VittoriaConfig) *Server {
	s := &Server{
		db:            db,
		router:        mux.NewRouter(),
		config:        config,
		unifiedConfig: unifiedConfig,
		processor:     processor.NewProcessorFactory(),
	}

	s.setupRoutes()
	s.setupMiddleware()

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      s.router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Starting VittoriaDB server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Stopping VittoriaDB server...")
	return s.server.Shutdown(ctx)
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Health and stats
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
	s.router.HandleFunc("/stats", s.handleStats).Methods("GET")
	s.router.HandleFunc("/config", s.handleConfig).Methods("GET")

	// Collection management
	s.router.HandleFunc("/collections", s.handleCollections).Methods("GET", "POST")
	s.router.HandleFunc("/collections/{name}", s.handleCollection).Methods("GET", "DELETE")
	s.router.HandleFunc("/collections/{name}/stats", s.handleCollectionStats).Methods("GET")

	// Vector operations
	s.router.HandleFunc("/collections/{name}/vectors", s.handleVectors).Methods("POST")
	s.router.HandleFunc("/collections/{name}/vectors/batch", s.handleVectorsBatch).Methods("POST")
	s.router.HandleFunc("/collections/{name}/vectors/{id}", s.handleVector).Methods("GET", "DELETE")
	s.router.HandleFunc("/collections/{name}/search", s.handleSearch).Methods("GET", "POST")

	// Text vectorization operations (automatic embedding generation)
	s.router.HandleFunc("/collections/{name}/text", s.handleTextInsert).Methods("POST")
	s.router.HandleFunc("/collections/{name}/text/batch", s.handleTextBatch).Methods("POST")
	s.router.HandleFunc("/collections/{name}/search/text", s.handleTextSearch).Methods("GET", "POST")

	// Document processing
	s.router.HandleFunc("/collections/{name}/documents", s.handleDocumentUpload).Methods("POST")
	s.router.HandleFunc("/documents/process", s.handleDocumentProcess).Methods("POST")
	s.router.HandleFunc("/documents/supported", s.handleSupportedFormats).Methods("GET")

	// Web dashboard (simple HTML page)
	s.router.HandleFunc("/", s.handleDashboard).Methods("GET")
}

// setupMiddleware configures HTTP middleware
func (s *Server) setupMiddleware() {
	// CORS middleware
	if s.config.CORS {
		s.router.Use(s.corsMiddleware)
	}

	// Logging middleware
	s.router.Use(s.loggingMiddleware)

	// JSON content type middleware
	s.router.Use(s.jsonMiddleware)
}

// Health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := s.db.Health()
	s.writeJSON(w, http.StatusOK, health)
}

// Database stats endpoint
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.Stats(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to get stats", err)
		return
	}

	s.writeJSON(w, http.StatusOK, stats)
}

// Configuration endpoint
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if s.unifiedConfig == nil {
		s.writeError(w, http.StatusInternalServerError, "Configuration not available", nil)
		return
	}

	// Create a response with configuration and metadata
	response := map[string]interface{}{
		"config": s.unifiedConfig,
		"metadata": map[string]interface{}{
			"source":      s.unifiedConfig.Source,
			"loaded_at":   time.Now().Format(time.RFC3339),
			"version":     "v1",
			"description": "VittoriaDB unified configuration",
		},
		"features": map[string]interface{}{
			"parallel_search":    s.unifiedConfig.Search.Parallel.Enabled,
			"search_cache":       s.unifiedConfig.Search.Cache.Enabled,
			"memory_mapped_io":   s.unifiedConfig.Performance.IO.UseMemoryMap,
			"simd_optimizations": s.unifiedConfig.Performance.EnableSIMD,
			"async_io":           s.unifiedConfig.Performance.IO.AsyncIO,
		},
		"performance": map[string]interface{}{
			"max_workers":      s.unifiedConfig.Search.Parallel.MaxWorkers,
			"cache_entries":    s.unifiedConfig.Search.Cache.MaxEntries,
			"cache_ttl":        s.unifiedConfig.Search.Cache.TTL.String(),
			"max_concurrency":  s.unifiedConfig.Performance.MaxConcurrency,
			"memory_limit_mb":  s.unifiedConfig.Performance.MemoryLimit / (1024 * 1024),
		},
	}

	s.writeJSON(w, http.StatusOK, response)
}

// Collections endpoint (GET: list, POST: create)
func (s *Server) handleCollections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleListCollections(w, r)
	case "POST":
		s.handleCreateCollection(w, r)
	}
}

// List collections
func (s *Server) handleListCollections(w http.ResponseWriter, r *http.Request) {
	collections, err := s.db.ListCollections(r.Context())
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to list collections", err)
		return
	}

	response := map[string]interface{}{
		"collections": collections,
		"count":       len(collections),
	}

	s.writeJSON(w, http.StatusOK, response)
}

// Create collection
func (s *Server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	var req core.CreateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := s.db.CreateCollection(r.Context(), &req); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			s.writeError(w, http.StatusConflict, "Collection already exists", err)
		} else {
			s.writeError(w, http.StatusBadRequest, "Failed to create collection", err)
		}
		return
	}

	response := map[string]string{
		"status":     "created",
		"collection": req.Name,
	}

	s.writeJSON(w, http.StatusCreated, response)
}

// Collection endpoint (GET: info, DELETE: drop)
func (s *Server) handleCollection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	switch r.Method {
	case "GET":
		s.handleGetCollection(w, r, name)
	case "DELETE":
		s.handleDropCollection(w, r, name)
	}
}

// Get collection info
func (s *Server) handleGetCollection(w http.ResponseWriter, r *http.Request, name string) {
	collection, err := s.db.GetCollection(r.Context(), name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Collection not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to get collection", err)
		}
		return
	}

	// Get collection info (assuming we have an Info method)
	if vittoriaCollection, ok := collection.(*core.VittoriaCollection); ok {
		info, err := vittoriaCollection.Info()
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "Failed to get collection info", err)
			return
		}
		s.writeJSON(w, http.StatusOK, info)
	} else {
		s.writeError(w, http.StatusInternalServerError, "Invalid collection type", nil)
	}
}

// Drop collection
func (s *Server) handleDropCollection(w http.ResponseWriter, r *http.Request, name string) {
	if err := s.db.DropCollection(r.Context(), name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Collection not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to drop collection", err)
		}
		return
	}

	response := map[string]string{
		"status":     "deleted",
		"collection": name,
	}

	s.writeJSON(w, http.StatusOK, response)
}

// Collection stats endpoint
func (s *Server) handleCollectionStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	collection, err := s.db.GetCollection(r.Context(), name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Collection not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to get collection", err)
		}
		return
	}

	count, err := collection.Count()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to get collection count", err)
		return
	}

	stats := map[string]interface{}{
		"name":         collection.Name(),
		"dimensions":   collection.Dimensions(),
		"metric":       collection.Metric().String(),
		"vector_count": count,
	}

	s.writeJSON(w, http.StatusOK, stats)
}

// Insert vector endpoint
func (s *Server) handleVectors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	collection, err := s.db.GetCollection(r.Context(), name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Collection not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to get collection", err)
		}
		return
	}

	var vector core.Vector
	if err := json.NewDecoder(r.Body).Decode(&vector); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := collection.Insert(r.Context(), &vector); err != nil {
		s.writeError(w, http.StatusBadRequest, "Failed to insert vector", err)
		return
	}

	response := map[string]string{
		"status": "inserted",
		"id":     vector.ID,
	}

	s.writeJSON(w, http.StatusCreated, response)
}

// Batch insert vectors endpoint
func (s *Server) handleVectorsBatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	collection, err := s.db.GetCollection(r.Context(), name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Collection not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to get collection", err)
		}
		return
	}

	var req struct {
		Vectors []*core.Vector `json:"vectors"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := collection.InsertBatch(r.Context(), req.Vectors); err != nil {
		s.writeError(w, http.StatusBadRequest, "Failed to insert vectors", err)
		return
	}

	response := map[string]interface{}{
		"status":   "inserted",
		"inserted": len(req.Vectors),
		"failed":   0,
	}

	s.writeJSON(w, http.StatusCreated, response)
}

// Vector endpoint (GET: get, DELETE: delete)
func (s *Server) handleVector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["name"]
	vectorID := vars["id"]

	collection, err := s.db.GetCollection(r.Context(), collectionName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Collection not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to get collection", err)
		}
		return
	}

	switch r.Method {
	case "GET":
		s.handleGetVector(w, r, collection, vectorID)
	case "DELETE":
		s.handleDeleteVector(w, r, collection, vectorID)
	}
}

// Get vector by ID
func (s *Server) handleGetVector(w http.ResponseWriter, r *http.Request, collection core.Collection, id string) {
	vector, err := collection.Get(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Vector not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to get vector", err)
		}
		return
	}

	s.writeJSON(w, http.StatusOK, vector)
}

// Delete vector by ID
func (s *Server) handleDeleteVector(w http.ResponseWriter, r *http.Request, collection core.Collection, id string) {
	if err := collection.Delete(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Vector not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to delete vector", err)
		}
		return
	}

	response := map[string]string{
		"status": "deleted",
		"id":     id,
	}

	s.writeJSON(w, http.StatusOK, response)
}

// Search endpoint
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	collection, err := s.db.GetCollection(r.Context(), name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Collection not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to get collection", err)
		}
		return
	}

	var searchReq core.SearchRequest

	if r.Method == "GET" {
		// Parse query parameters
		if err := s.parseSearchParams(r, &searchReq); err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid search parameters", err)
			return
		}
	} else {
		// Parse JSON body
		if err := json.NewDecoder(r.Body).Decode(&searchReq); err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
			return
		}
	}

	// Set defaults
	if searchReq.Limit <= 0 {
		searchReq.Limit = 10
	}
	if searchReq.Limit > 1000 {
		searchReq.Limit = 1000
	}

	results, err := collection.Search(r.Context(), &searchReq)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Search failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, results)
}

// Parse search parameters from query string
func (s *Server) parseSearchParams(r *http.Request, req *core.SearchRequest) error {
	query := r.URL.Query()

	// Parse vector
	vectorStr := query.Get("vector")
	if vectorStr == "" {
		return fmt.Errorf("vector parameter is required")
	}

	vector, err := s.parseVectorString(vectorStr)
	if err != nil {
		return fmt.Errorf("invalid vector format: %w", err)
	}
	req.Vector = vector

	// Parse limit
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		req.Limit = limit
	}

	// Parse offset
	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return fmt.Errorf("invalid offset: %w", err)
		}
		req.Offset = offset
	}

	// Parse include flags
	req.IncludeVector = query.Get("include_vector") == "true"
	req.IncludeMetadata = query.Get("include_metadata") != "false" // default true

	// Parse filter (JSON string)
	if filterStr := query.Get("filter"); filterStr != "" {
		var filter core.Filter
		if err := json.Unmarshal([]byte(filterStr), &filter); err != nil {
			return fmt.Errorf("invalid filter format: %w", err)
		}
		req.Filter = &filter
	}

	return nil
}

// Parse vector string "[0.1,0.2,0.3]" to []float32
func (s *Server) parseVectorString(vectorStr string) ([]float32, error) {
	// Remove brackets and spaces
	vectorStr = strings.Trim(vectorStr, "[]")
	vectorStr = strings.ReplaceAll(vectorStr, " ", "")

	if vectorStr == "" {
		return nil, fmt.Errorf("empty vector")
	}

	parts := strings.Split(vectorStr, ",")
	vector := make([]float32, len(parts))

	for i, part := range parts {
		val, err := strconv.ParseFloat(part, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid float value: %s", part)
		}
		vector[i] = float32(val)
	}

	return vector, nil
}

// Simple web dashboard
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>VittoriaDB Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .endpoint { background: #f9f9f9; padding: 10px; margin: 5px 0; border-left: 4px solid #007cba; }
        code { background: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 VittoriaDB</h1>
        <p>Simple Embedded Vector Database</p>
    </div>
    
    <div class="section">
        <h2>Quick Start</h2>
        <div class="endpoint">
            <strong>Health Check:</strong> <code>GET /health</code>
        </div>
        <div class="endpoint">
            <strong>List Collections:</strong> <code>GET /collections</code>
        </div>
        <div class="endpoint">
            <strong>Create Collection:</strong> <code>POST /collections</code>
        </div>
    </div>
    
    <div class="section">
        <h2>Example Usage</h2>
        <pre><code># Create collection
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "docs", "dimensions": 4, "metric": "cosine"}'

# Insert vector
curl -X POST http://localhost:8080/collections/docs/vectors \
  -H "Content-Type: application/json" \
  -d '{"id": "doc1", "vector": [0.1, 0.2, 0.3, 0.4], "metadata": {"title": "Test"}}'

# Search
curl "http://localhost:8080/collections/docs/search?vector=0.1,0.2,0.3,0.4&limit=5"</code></pre>
    </div>
    
    <div class="section">
        <h2>API Endpoints</h2>
        <div class="endpoint"><code>GET /health</code> - Health check</div>
        <div class="endpoint"><code>GET /stats</code> - Database statistics</div>
        <div class="endpoint"><code>GET /config</code> - Current configuration</div>
        <div class="endpoint"><code>GET /collections</code> - List collections</div>
        <div class="endpoint"><code>POST /collections</code> - Create collection</div>
        <div class="endpoint"><code>GET /collections/{name}</code> - Get collection info</div>
        <div class="endpoint"><code>DELETE /collections/{name}</code> - Delete collection</div>
        <div class="endpoint"><code>POST /collections/{name}/vectors</code> - Insert vector</div>
        <div class="endpoint"><code>GET /collections/{name}/search</code> - Search vectors</div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Text insertion endpoint (automatic vectorization)
func (s *Server) handleTextInsert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	collection, err := s.db.GetCollection(r.Context(), name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Collection not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to get collection", err)
		}
		return
	}

	var textVector core.TextVector
	if err := json.NewDecoder(r.Body).Decode(&textVector); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Check if collection has vectorizer
	if !collection.HasVectorizer() {
		s.writeError(w, http.StatusBadRequest, "Collection does not have vectorizer configured", nil)
		return
	}

	if err := collection.InsertText(r.Context(), &textVector); err != nil {
		s.writeError(w, http.StatusBadRequest, "Failed to insert text", err)
		return
	}

	response := map[string]string{
		"status": "inserted",
		"id":     textVector.ID,
	}

	s.writeJSON(w, http.StatusCreated, response)
}

// Batch text insertion endpoint (automatic vectorization)
func (s *Server) handleTextBatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	collection, err := s.db.GetCollection(r.Context(), name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Collection not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to get collection", err)
		}
		return
	}

	var req struct {
		Texts []*core.TextVector `json:"texts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Check if collection has vectorizer
	if !collection.HasVectorizer() {
		s.writeError(w, http.StatusBadRequest, "Collection does not have vectorizer configured", nil)
		return
	}

	if err := collection.InsertTextBatch(r.Context(), req.Texts); err != nil {
		s.writeError(w, http.StatusBadRequest, "Failed to insert texts", err)
		return
	}

	response := map[string]interface{}{
		"status":   "inserted",
		"inserted": len(req.Texts),
		"failed":   0,
	}

	s.writeJSON(w, http.StatusCreated, response)
}

// Text search endpoint (automatic query vectorization)
func (s *Server) handleTextSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	collection, err := s.db.GetCollection(r.Context(), name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.writeError(w, http.StatusNotFound, "Collection not found", err)
		} else {
			s.writeError(w, http.StatusInternalServerError, "Failed to get collection", err)
		}
		return
	}

	// Check if collection has vectorizer
	if !collection.HasVectorizer() {
		s.writeError(w, http.StatusBadRequest, "Collection does not have vectorizer configured", nil)
		return
	}

	// Parse query parameters and request body
	var query string
	var limit int = 10
	var includeMetadata bool = true
	var includeContent bool = false
	
	if r.Method == "POST" {
		// Parse JSON body for POST requests
		var req struct {
			Query           string `json:"query"`
			Limit           int    `json:"limit"`
			IncludeMetadata bool   `json:"include_metadata"`
			IncludeContent  bool   `json:"include_content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
			return
		}
		query = req.Query
		if req.Limit > 0 {
			limit = req.Limit
		}
		includeMetadata = req.IncludeMetadata
		includeContent = req.IncludeContent
	} else {
		// Parse URL parameters for GET requests
		query = r.URL.Query().Get("query")
		if query == "" {
			s.writeError(w, http.StatusBadRequest, "Missing query parameter", nil)
			return
		}
		
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}
		
		if metadataStr := r.URL.Query().Get("include_metadata"); metadataStr != "" {
			includeMetadata = metadataStr == "true"
		}
		
		if contentStr := r.URL.Query().Get("include_content"); contentStr != "" {
			includeContent = contentStr == "true"
		}
	}

	if query == "" {
		s.writeError(w, http.StatusBadRequest, "Missing query parameter", nil)
		return
	}

	// Create search request with content inclusion
	searchReq := &core.SearchRequest{
		Limit:           limit,
		IncludeMetadata: includeMetadata,
		IncludeContent:  includeContent,
	}

	// Perform text search with automatic vectorization using the enhanced search
	// First get the vectorizer to convert text to vector
	vectorizer := collection.GetVectorizer()
	if vectorizer == nil {
		s.writeError(w, http.StatusInternalServerError, "No vectorizer available", nil)
		return
	}
	
	// Generate embedding from query text
	queryEmbedding, err := vectorizer.GenerateEmbedding(r.Context(), query)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to generate query embedding", err)
		return
	}
	
	searchReq.Vector = queryEmbedding
	results, err := collection.Search(r.Context(), searchReq)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Search failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, results)
}

// Middleware functions

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Helper functions

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string, err error) {
	errorResponse := map[string]interface{}{
		"error":  message,
		"status": status,
		"time":   time.Now().Unix(),
	}

	if err != nil {
		errorResponse["details"] = err.Error()
		log.Printf("API Error: %s - %v", message, err)
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse)
}

// Document processing handlers

// handleDocumentUpload handles document upload and processing for a collection
func (s *Server) handleDocumentUpload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["name"]

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Failed to parse multipart form", err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "No file provided", err)
		return
	}
	defer file.Close()

	// Get processing configuration from form
	config := processor.DefaultProcessingConfig()
	if chunkSize := r.FormValue("chunk_size"); chunkSize != "" {
		if size, err := strconv.Atoi(chunkSize); err == nil {
			config.ChunkSize = size
		}
	}
	if overlap := r.FormValue("chunk_overlap"); overlap != "" {
		if size, err := strconv.Atoi(overlap); err == nil {
			config.ChunkOverlap = size
		}
	}
	if lang := r.FormValue("language"); lang != "" {
		config.Language = lang
	}

	// Add metadata from form
	if metadata := r.FormValue("metadata"); metadata != "" {
		var meta map[string]string
		if err := json.Unmarshal([]byte(metadata), &meta); err == nil {
			config.Metadata = meta
		}
	}

	// Process document
	proc, err := s.processor.GetProcessorByFilename(header.Filename)
	if err != nil {
		s.writeError(w, http.StatusUnsupportedMediaType, "Unsupported document type", err)
		return
	}

	doc, err := proc.ProcessDocument(file, header.Filename, config)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to process document", err)
		return
	}

	// Get collection
	collection, err := s.db.GetCollection(r.Context(), collectionName)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "Collection not found", err)
		return
	}

	// Insert document chunks as vectors (placeholder - would need embedding generation)
	var insertedChunks []string
	for _, chunk := range doc.Chunks {
		// Use automatic text vectorization if collection has vectorizer
		if collection.HasVectorizer() {
			// Create TextVector for automatic embedding generation
			textVector := &core.TextVector{
				ID:   chunk.ID,
				Text: chunk.Content,
				Metadata: map[string]interface{}{
					"document_id":    doc.ID,
					"document_title": doc.Title,
					"chunk_content":  chunk.Content,
					"chunk_position": chunk.Position,
					"chunk_size":     chunk.Size,
				},
			}

			// Add chunk metadata
			for k, v := range chunk.Metadata {
				textVector.Metadata["chunk_"+k] = v
			}

			if err := collection.InsertText(r.Context(), textVector); err != nil {
				log.Printf("Failed to insert text chunk %s: %v", chunk.ID, err)
				continue
			}
		} else {
			// Fallback to placeholder vector for collections without vectorizer
			vector := &core.Vector{
				ID:     chunk.ID,
				Vector: make([]float32, 384), // Placeholder vector
				Metadata: map[string]interface{}{
					"document_id":    doc.ID,
					"document_title": doc.Title,
					"chunk_content":  chunk.Content,
					"chunk_position": chunk.Position,
					"chunk_size":     chunk.Size,
				},
			}

			// Add chunk metadata
			for k, v := range chunk.Metadata {
				vector.Metadata["chunk_"+k] = v
			}

			if err := collection.Insert(r.Context(), vector); err != nil {
				log.Printf("Failed to insert chunk %s: %v", chunk.ID, err)
				continue
			}
		}

		insertedChunks = append(insertedChunks, chunk.ID)
	}

	response := map[string]interface{}{
		"status":          "processed",
		"document_id":     doc.ID,
		"document_title":  doc.Title,
		"document_type":   doc.Type,
		"chunks_created":  len(doc.Chunks),
		"chunks_inserted": len(insertedChunks),
		"processing_time": time.Since(doc.ProcessedAt).Milliseconds(),
		"collection":      collectionName,
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleDocumentProcess processes a document without adding to collection
func (s *Server) handleDocumentProcess(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Failed to parse multipart form", err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "No file provided", err)
		return
	}
	defer file.Close()

	// Get processing configuration
	config := processor.DefaultProcessingConfig()
	if chunkSize := r.FormValue("chunk_size"); chunkSize != "" {
		if size, err := strconv.Atoi(chunkSize); err == nil {
			config.ChunkSize = size
		}
	}
	if overlap := r.FormValue("chunk_overlap"); overlap != "" {
		if size, err := strconv.Atoi(overlap); err == nil {
			config.ChunkOverlap = size
		}
	}

	// Process document
	proc, err := s.processor.GetProcessorByFilename(header.Filename)
	if err != nil {
		s.writeError(w, http.StatusUnsupportedMediaType, "Unsupported document type", err)
		return
	}

	doc, err := proc.ProcessDocument(file, header.Filename, config)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to process document", err)
		return
	}

	s.writeJSON(w, http.StatusOK, doc)
}

// handleSupportedFormats returns supported document formats
func (s *Server) handleSupportedFormats(w http.ResponseWriter, r *http.Request) {
	info := s.processor.GetProcessorInfo()

	response := map[string]interface{}{
		"supported_formats": info,
		"extensions":        s.processor.GetSupportedExtensions(),
		"total_processors":  len(info),
	}

	s.writeJSON(w, http.StatusOK, response)
}
