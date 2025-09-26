package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/antonellof/VittoriaDB/pkg/core"
	"github.com/gorilla/mux"
)

// handleDocumentCreate handles document database creation with schema
func (s *Server) handleDocumentCreate(w http.ResponseWriter, r *http.Request) {
	var req core.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Validate schema
	if req.Schema == nil || len(req.Schema) == 0 {
		s.writeError(w, http.StatusBadRequest, "Schema is required", nil)
		return
	}

	// Create unified database instance (for now, we'll store schema in metadata)
	// In a full implementation, this would create a separate unified database instance

	// For backward compatibility, we'll create collections based on vector fields in schema
	validator := core.NewSchemaValidator(req.Schema)
	vectorFields := validator.GetVectorFields()

	createdCollections := []string{}
	for fieldPath, dimensions := range vectorFields {
		collectionName := strings.ReplaceAll(fieldPath, ".", "_") + "_vectors"

		// Check if collection already exists with different dimensions
		if existingCollection, err := s.db.GetCollection(r.Context(), collectionName); err == nil {
			// Collection exists - check if dimensions match
			existingDims := existingCollection.Dimensions()
			if existingDims != dimensions {
				// Dimensions don't match - delete and recreate
				if err := s.db.DropCollection(r.Context(), collectionName); err != nil {
					s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete incompatible collection %s", collectionName), err)
					return
				}
			} else {
				// Dimensions match - reuse existing collection
				createdCollections = append(createdCollections, collectionName)
				continue
			}
		}

		// Create new collection
		createReq := &core.CreateCollectionRequest{
			Name:       collectionName,
			Dimensions: dimensions,
			Metric:     core.DistanceMetricCosine,
			IndexType:  core.IndexTypeHNSW,
		}

		if err := s.db.CreateCollection(r.Context(), createReq); err != nil {
			s.writeError(w, http.StatusInternalServerError, "Failed to create collection", err)
			return
		}
		createdCollections = append(createdCollections, collectionName)
	}

	// Store schema metadata in a special collection
	schemaCollectionName := "_schema_metadata"
	schemaReq := &core.CreateCollectionRequest{
		Name:       schemaCollectionName,
		Dimensions: 1, // Minimal dimension for metadata storage
		Metric:     core.DistanceMetricCosine,
		IndexType:  core.IndexTypeFlat,
	}

	if err := s.db.CreateCollection(r.Context(), schemaReq); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			s.writeError(w, http.StatusInternalServerError, "Failed to create schema collection", err)
			return
		}
	}

	// Store schema as metadata
	schemaCollection, err := s.db.GetCollection(r.Context(), schemaCollectionName)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to get schema collection", err)
		return
	}

	schemaBytes, _ := json.Marshal(req.Schema)
	schemaVector := &core.Vector{
		ID:     "schema_definition",
		Vector: []float32{1.0}, // Dummy vector
		Metadata: map[string]interface{}{
			"schema":             string(schemaBytes),
			"language":           req.Language,
			"fulltext_config":    req.FullTextConfig,
			"vectorizer_configs": req.VectorizerConfigs,
			"content_storage":    req.ContentStorage,
		},
	}

	if err := schemaCollection.Insert(r.Context(), schemaVector); err != nil {
		// Try to update if it already exists
		schemaCollection.Delete(r.Context(), "schema_definition")
		if err := schemaCollection.Insert(r.Context(), schemaVector); err != nil {
			s.writeError(w, http.StatusInternalServerError, "Failed to store schema", err)
			return
		}
	}

	response := map[string]interface{}{
		"status":              "created",
		"schema":              req.Schema,
		"vector_fields":       vectorFields,
		"searchable_fields":   validator.GetSearchableFields(),
		"created_collections": createdCollections,
	}

	s.writeJSON(w, http.StatusCreated, response)
}

// handleDocumentInsert handles document insertion with schema validation
func (s *Server) handleDocumentInsert(w http.ResponseWriter, r *http.Request) {
	var req core.InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Get schema from metadata
	schema, err := s.getStoredSchema(r.Context())
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "No schema found. Create a unified database first.", err)
		return
	}

	// Create unified database instance
	ddb := core.CreateDocumentDatabase(schema)
	if err := ddb.Open(r.Context(), &core.Config{DataDir: s.unifiedConfig.DataDir}); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to open unified database", err)
		return
	}
	defer ddb.Close()

	// Insert document
	docID, err := ddb.Insert(r.Context(), req.Document)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Failed to insert document", err)
		return
	}

	response := &core.InsertResponse{
		ID:      docID,
		Created: true,
	}

	s.writeJSON(w, http.StatusCreated, response)
}

// handleDocumentSearch handles document search with multiple modes
func (s *Server) handleDocumentSearch(w http.ResponseWriter, r *http.Request) {
	var params core.SearchParams

	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
			return
		}
	} else {
		// Parse GET parameters
		if err := s.parseDocumentSearchParams(r, &params); err != nil {
			s.writeError(w, http.StatusBadRequest, "Invalid search parameters", err)
			return
		}
	}

	// Set defaults
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 1000 {
		params.Limit = 1000
	}
	if params.Mode == "" {
		params.Mode = core.SearchModeFullText
	}

	// Get schema from metadata
	schema, err := s.getStoredSchema(r.Context())
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "No schema found. Create a unified database first.", err)
		return
	}

	// Create unified database instance
	ddb := core.CreateDocumentDatabase(schema)
	if err := ddb.Open(r.Context(), &core.Config{DataDir: s.unifiedConfig.DataDir}); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to open unified database", err)
		return
	}
	defer ddb.Close()

	// Perform search
	results, err := ddb.Search(r.Context(), &params)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Search failed", err)
		return
	}

	s.writeJSON(w, http.StatusOK, results)
}

// handleDocumentGet handles document retrieval by ID
func (s *Server) handleDocumentGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docID := vars["id"]

	if docID == "" {
		s.writeError(w, http.StatusBadRequest, "Document ID is required", nil)
		return
	}

	includeVectors := r.URL.Query().Get("include_vectors") == "true"

	// Get schema from metadata
	schema, err := s.getStoredSchema(r.Context())
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "No schema found. Create a document database first.", err)
		return
	}

	// Create document database instance
	ddb := core.CreateDocumentDatabase(schema)
	if err := ddb.Open(r.Context(), &core.Config{DataDir: s.unifiedConfig.DataDir}); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to open document database", err)
		return
	}
	defer ddb.Close()

	// Use the DocumentDatabase Get method
	response, err := ddb.Get(r.Context(), docID, includeVectors)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to get document", err)
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleDocumentUpdate handles document updates
func (s *Server) handleDocumentUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docID := vars["id"]

	var req core.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	req.ID = docID

	// Get schema from metadata
	schema, err := s.getStoredSchema(r.Context())
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "No schema found. Create a unified database first.", err)
		return
	}

	// For simplicity, we'll delete and re-insert
	// In a full implementation, this would be more sophisticated
	validator := core.NewSchemaValidator(schema)
	vectorFields := validator.GetVectorFields()

	// Delete from all collections
	for fieldPath := range vectorFields {
		collectionName := strings.ReplaceAll(fieldPath, ".", "_") + "_vectors"
		collection, err := s.db.GetCollection(r.Context(), collectionName)
		if err != nil {
			continue
		}
		collection.Delete(r.Context(), docID)
	}

	// Re-insert with updated data
	req.Document["id"] = docID
	ddb := core.CreateDocumentDatabase(schema)
	if err := ddb.Open(r.Context(), &core.Config{DataDir: s.unifiedConfig.DataDir}); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to open unified database", err)
		return
	}
	defer ddb.Close()

	_, err = ddb.Insert(r.Context(), req.Document)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Failed to update document", err)
		return
	}

	response := &core.UpdateResponse{
		ID:      docID,
		Updated: true,
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleDocumentDelete handles document deletion
func (s *Server) handleDocumentDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docID := vars["id"]

	if docID == "" {
		s.writeError(w, http.StatusBadRequest, "Document ID is required", nil)
		return
	}

	// Get schema from metadata
	schema, err := s.getStoredSchema(r.Context())
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "No schema found. Create a unified database first.", err)
		return
	}

	// Delete from all collections
	validator := core.NewSchemaValidator(schema)
	vectorFields := validator.GetVectorFields()

	deleted := false
	for fieldPath := range vectorFields {
		collectionName := strings.ReplaceAll(fieldPath, ".", "_") + "_vectors"
		collection, err := s.db.GetCollection(r.Context(), collectionName)
		if err != nil {
			continue
		}

		if err := collection.Delete(r.Context(), docID); err == nil {
			deleted = true
		}
	}

	response := &core.DeleteResponse{
		ID:      docID,
		Deleted: deleted,
	}

	s.writeJSON(w, http.StatusOK, response)
}

// handleDocumentCount handles document counting
func (s *Server) handleDocumentCount(w http.ResponseWriter, r *http.Request) {
	// Get schema from metadata
	schema, err := s.getStoredSchema(r.Context())
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "No schema found. Create a document database first.", err)
		return
	}

	// Create document database instance
	ddb := core.CreateDocumentDatabase(schema)
	if err := ddb.Open(r.Context(), &core.Config{DataDir: s.unifiedConfig.DataDir}); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to open document database", err)
		return
	}
	defer ddb.Close()

	// Use the DocumentDatabase Count method
	response, err := ddb.Count(r.Context(), nil)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to count documents", err)
		return
	}

	s.writeJSON(w, http.StatusOK, response)
}

// Helper methods

func (s *Server) getStoredSchema(ctx context.Context) (core.Schema, error) {
	schemaCollection, err := s.db.GetCollection(ctx, "_schema_metadata")
	if err != nil {
		return nil, fmt.Errorf("schema collection not found")
	}

	schemaVector, err := schemaCollection.Get(ctx, "schema_definition")
	if err != nil {
		return nil, fmt.Errorf("schema definition not found")
	}

	schemaStr, ok := schemaVector.Metadata["schema"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid schema format")
	}

	var schema core.Schema
	if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return schema, nil
}

func (s *Server) parseDocumentSearchParams(r *http.Request, params *core.SearchParams) error {
	query := r.URL.Query()

	// Basic parameters
	params.Term = query.Get("term")
	params.Mode = core.SearchMode(query.Get("mode"))
	if params.Mode == "" {
		params.Mode = core.SearchModeFullText
	}

	// Limit and offset
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = offset
		}
	}

	// Properties
	if propertiesStr := query.Get("properties"); propertiesStr != "" {
		params.Properties = strings.Split(propertiesStr, ",")
	}

	// Vector search parameters
	if vectorStr := query.Get("vector"); vectorStr != "" {
		property := query.Get("vector_property")
		if property == "" {
			property = "embedding" // Default vector property
		}

		// Parse vector values
		vectorParts := strings.Split(strings.Trim(vectorStr, "[]"), ",")
		vector := make([]float32, len(vectorParts))
		for i, part := range vectorParts {
			if val, err := strconv.ParseFloat(strings.TrimSpace(part), 32); err == nil {
				vector[i] = float32(val)
			}
		}

		params.Vector = &core.VectorSearchParams{
			Value:    vector,
			Property: property,
		}
	}

	// Similarity threshold
	if similarityStr := query.Get("similarity"); similarityStr != "" {
		if similarity, err := strconv.ParseFloat(similarityStr, 64); err == nil {
			params.Similarity = similarity
		}
	}

	// Include vectors
	params.IncludeVectors = query.Get("include_vectors") == "true"

	// Threshold
	if thresholdStr := query.Get("threshold"); thresholdStr != "" {
		if threshold, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
			params.Threshold = threshold
		}
	}

	return nil
}

// setupDocumentRoutes adds the document API routes
func (s *Server) setupDocumentRoutes() {
	// Document database management
	s.router.HandleFunc("/create", s.handleDocumentCreate).Methods("POST")

	// Document operations
	s.router.HandleFunc("/documents", s.handleDocumentInsert).Methods("POST")
	s.router.HandleFunc("/documents/{id}", s.handleDocumentGet).Methods("GET")
	s.router.HandleFunc("/documents/{id}", s.handleDocumentUpdate).Methods("PUT")
	s.router.HandleFunc("/documents/{id}", s.handleDocumentDelete).Methods("DELETE")

	// Search operations
	s.router.HandleFunc("/search", s.handleDocumentSearch).Methods("GET", "POST")

	// Utility operations
	s.router.HandleFunc("/count", s.handleDocumentCount).Methods("GET")
}
