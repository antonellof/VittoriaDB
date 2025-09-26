package core

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DocumentDatabase provides a schema-based document API
type DocumentDatabase struct {
	db        Database
	schema    Schema
	validator *SchemaValidator
	tokenizer *TextTokenizer
}

// CreateDocumentDatabase creates a new document database with schema support
func CreateDocumentDatabase(schema Schema) *DocumentDatabase {
	return &DocumentDatabase{
		db:        NewDatabase(),
		schema:    schema,
		validator: NewSchemaValidator(schema),
		tokenizer: NewTextTokenizer(),
	}
}

// Open initializes the unified database
func (ddb *DocumentDatabase) Open(ctx context.Context, config *Config) error {
	return ddb.db.Open(ctx, config)
}

// Close closes the unified database
func (ddb *DocumentDatabase) Close() error {
	return ddb.db.Close()
}

// Insert inserts a document into the database
func (ddb *DocumentDatabase) Insert(ctx context.Context, doc Document) (string, error) {
	// Validate document against schema
	if err := ddb.validator.ValidateDocument(doc); err != nil {
		return "", fmt.Errorf("schema validation failed: %w", err)
	}

	// Generate document ID if not provided
	docID, ok := doc["id"].(string)
	if !ok || docID == "" {
		docID = generateDocumentID()
		doc["id"] = docID
	}

	// Extract vectors from document
	vectors := ddb.validator.ExtractVectors(doc)

	// Create collections for each vector field if they don't exist
	vectorFields := ddb.validator.GetVectorFields()
	for fieldPath, vector := range vectors {
		collectionName := ddb.getCollectionName(fieldPath)

		// Get expected dimensions from schema
		expectedDims, exists := vectorFields[fieldPath]
		if !exists {
			return "", fmt.Errorf("vector field '%s' not found in schema", fieldPath)
		}

		// Validate vector dimensions match schema
		if len(vector) != expectedDims {
			return "", fmt.Errorf("vector field '%s' has %d dimensions, expected %d", fieldPath, len(vector), expectedDims)
		}

		// Ensure collection exists with schema dimensions
		if err := ddb.ensureCollection(ctx, collectionName, expectedDims); err != nil {
			return "", fmt.Errorf("failed to ensure collection %s: %w", collectionName, err)
		}

		// Insert vector into collection
		collection, err := ddb.db.GetCollection(ctx, collectionName)
		if err != nil {
			return "", fmt.Errorf("failed to get collection %s: %w", collectionName, err)
		}

		vectorDoc := &Vector{
			ID:       docID,
			Vector:   vector,
			Metadata: ddb.createMetadata(doc, fieldPath),
		}

		if err := collection.Insert(ctx, vectorDoc); err != nil {
			return "", fmt.Errorf("failed to insert vector into %s: %w", collectionName, err)
		}
	}

	// Index searchable text fields
	searchableText := ddb.validator.ExtractSearchableText(doc)
	if len(searchableText) > 0 {
		if err := ddb.indexTextFields(ctx, docID, searchableText, doc); err != nil {
			return "", fmt.Errorf("failed to index text fields: %w", err)
		}
	}

	return docID, nil
}

// Get retrieves a document by ID
func (ddb *DocumentDatabase) Get(ctx context.Context, id string, includeVectors bool) (*GetResponse, error) {
	// Try to find the document in any of the vector collections
	vectorFields := ddb.validator.GetVectorFields()

	var foundDoc Document
	found := false

	// Search in vector collections first
	for fieldPath := range vectorFields {
		collectionName := ddb.getCollectionName(fieldPath)
		collection, err := ddb.db.GetCollection(ctx, collectionName)
		if err != nil {
			continue // Collection might not exist yet
		}

		vector, err := collection.Get(ctx, id)
		if err != nil {
			continue // Document not found in this collection
		}

		if vector != nil {
			// Reconstruct document from metadata
			foundDoc = ddb.reconstructDocument(vector.Metadata)
			if includeVectors {
				foundDoc[fieldPath] = vector.Vector
			}
			found = true
			break
		}
	}

	// If not found in vector collections, try text index
	if !found {
		textCollectionName := "_text_index"
		collection, err := ddb.db.GetCollection(ctx, textCollectionName)
		if err == nil {
			vector, err := collection.Get(ctx, id)
			if err == nil && vector != nil {
				foundDoc = ddb.reconstructDocument(vector.Metadata)
				found = true
			}
		}
	}

	return &GetResponse{
		Document: foundDoc,
		Found:    found,
	}, nil
}

// Count returns the total number of documents in the database
func (ddb *DocumentDatabase) Count(ctx context.Context, where map[string]interface{}) (*CountResponse, error) {
	// Use text index collection to count documents since all documents should be indexed there
	textCollectionName := "_text_index"
	collection, err := ddb.db.GetCollection(ctx, textCollectionName)
	if err != nil {
		// If text collection doesn't exist, try any vector collection
		vectorFields := ddb.validator.GetVectorFields()
		for fieldPath := range vectorFields {
			collectionName := ddb.getCollectionName(fieldPath)
			collection, err = ddb.db.GetCollection(ctx, collectionName)
			if err == nil {
				break
			}
		}

		if err != nil {
			return &CountResponse{Count: 0}, nil
		}
	}

	count, err := collection.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	return &CountResponse{Count: int(count)}, nil
}

// Update updates a document by ID
func (ddb *DocumentDatabase) Update(ctx context.Context, id string, doc Document, options *UpdateOptions) (*UpdateResponse, error) {
	// For now, implement update as delete + insert
	// In a production system, this would be more sophisticated

	// Check if document exists
	getResp, err := ddb.Get(ctx, id, false)
	if err != nil {
		return nil, fmt.Errorf("failed to check document existence: %w", err)
	}

	if !getResp.Found {
		return &UpdateResponse{
			ID:      id,
			Updated: false,
		}, nil
	}

	// Delete existing document
	_, err = ddb.Delete(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing document: %w", err)
	}

	// Set the ID in the new document
	doc["id"] = id

	// Insert updated document
	_, err = ddb.Insert(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert updated document: %w", err)
	}

	return &UpdateResponse{
		ID:      id,
		Updated: true,
	}, nil
}

// Delete removes a document by ID
func (ddb *DocumentDatabase) Delete(ctx context.Context, id string) (*DeleteResponse, error) {
	deleted := false

	// Delete from all vector collections
	vectorFields := ddb.validator.GetVectorFields()
	for fieldPath := range vectorFields {
		collectionName := ddb.getCollectionName(fieldPath)
		collection, err := ddb.db.GetCollection(ctx, collectionName)
		if err != nil {
			continue // Collection might not exist
		}

		err = collection.Delete(ctx, id)
		if err == nil {
			deleted = true
		}
	}

	// Delete from text index
	textCollectionName := "_text_index"
	collection, err := ddb.db.GetCollection(ctx, textCollectionName)
	if err == nil {
		err = collection.Delete(ctx, id)
		if err == nil {
			deleted = true
		}
	}

	return &DeleteResponse{
		ID:      id,
		Deleted: deleted,
	}, nil
}

// Search performs a unified search across the database
func (ddb *DocumentDatabase) Search(ctx context.Context, params *SearchParams) (*DocumentSearchResponse, error) {
	startTime := time.Now()

	var results []*DocumentSearchResult
	var err error

	switch params.Mode {
	case SearchModeFullText:
		results, err = ddb.searchFullText(ctx, params)
	case SearchModeVector:
		results, err = ddb.searchVector(ctx, params)
	case SearchModeHybrid:
		results, err = ddb.searchHybrid(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported search mode: %s", params.Mode)
	}

	if err != nil {
		return nil, err
	}

	// Apply limit and offset
	total := len(results)
	start := params.Offset
	if start > total {
		start = total
	}
	end := start + params.Limit
	if end > total {
		end = total
	}

	if start < end {
		results = results[start:end]
	} else {
		results = []*DocumentSearchResult{}
	}

	// Apply facets if requested
	var facets map[string]*FacetResult
	if len(params.Facets) > 0 {
		facets = ddb.calculateFacets(results, params.Facets)
	}

	return &DocumentSearchResponse{
		Hits:    results,
		Count:   total,
		Elapsed: time.Since(startTime),
		Facets:  facets,
	}, nil
}

// searchFullText performs full-text search
func (ddb *DocumentDatabase) searchFullText(ctx context.Context, params *SearchParams) ([]*DocumentSearchResult, error) {
	if params.Term == "" {
		return []*DocumentSearchResult{}, nil
	}

	// Tokenize search term
	tokens := ddb.tokenizer.Tokenize(params.Term)

	// Get searchable fields
	searchableFields := ddb.validator.GetSearchableFields()
	if len(params.Properties) > 0 {
		// Filter to requested properties
		requestedFields := make(map[string]bool)
		for _, prop := range params.Properties {
			requestedFields[prop] = true
		}

		var filteredFields []string
		for _, field := range searchableFields {
			if requestedFields[field] {
				filteredFields = append(filteredFields, field)
			}
		}
		searchableFields = filteredFields
	}

	// Search in text index (simplified implementation)
	// In a real implementation, this would use a proper text search index
	return ddb.searchTextIndex(ctx, tokens, searchableFields, params)
}

// searchVector performs vector similarity search
func (ddb *DocumentDatabase) searchVector(ctx context.Context, params *SearchParams) ([]*DocumentSearchResult, error) {
	if params.Vector == nil || len(params.Vector.Value) == 0 {
		return []*DocumentSearchResult{}, nil
	}

	collectionName := ddb.getCollectionName(params.Vector.Property)
	collection, err := ddb.db.GetCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %s", collectionName)
	}

	searchReq := &SearchRequest{
		Vector:          params.Vector.Value,
		Limit:           params.Limit + params.Offset, // Get more to handle offset
		Offset:          0,
		IncludeMetadata: true,
		Filter:          ddb.convertFilter(params.Where),
	}

	response, err := collection.Search(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	// Convert to unified results
	var results []*DocumentSearchResult
	for _, result := range response.Results {
		if params.Similarity > 0 && float64(result.Score) < params.Similarity {
			continue
		}

		unifiedResult := &DocumentSearchResult{
			ID:       result.ID,
			Score:    result.Score,
			Document: ddb.reconstructDocument(result.Metadata),
		}

		if params.IncludeVectors {
			unifiedResult.Document[params.Vector.Property] = result.Vector
		}

		results = append(results, unifiedResult)
	}

	return results, nil
}

// searchHybrid performs hybrid search combining text and vector search
func (ddb *DocumentDatabase) searchHybrid(ctx context.Context, params *SearchParams) ([]*DocumentSearchResult, error) {
	// Get text search results
	textResults, err := ddb.searchFullText(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("text search failed: %w", err)
	}

	// Get vector search results
	vectorResults, err := ddb.searchVector(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Combine and weight results
	return ddb.combineHybridResults(textResults, vectorResults, params.HybridWeights), nil
}

// Helper methods

func (ddb *DocumentDatabase) getCollectionName(fieldPath string) string {
	// Convert field path to collection name
	return strings.ReplaceAll(fieldPath, ".", "_") + "_vectors"
}

func (ddb *DocumentDatabase) ensureCollection(ctx context.Context, name string, dimensions int) error {
	// Check if collection exists
	_, err := ddb.db.GetCollection(ctx, name)
	if err == nil {
		return nil // Collection exists
	}

	// Create collection
	req := &CreateCollectionRequest{
		Name:       name,
		Dimensions: dimensions,
		Metric:     DistanceMetricCosine,
		IndexType:  IndexTypeHNSW,
	}

	return ddb.db.CreateCollection(ctx, req)
}

func (ddb *DocumentDatabase) createMetadata(doc Document, fieldPath string) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Copy all non-vector fields as metadata
	for key, value := range doc {
		if key != fieldPath && !ddb.isVectorField(key) {
			metadata[key] = value
		}
	}

	// Add field path for reconstruction
	metadata["_field_path"] = fieldPath

	return metadata
}

func (ddb *DocumentDatabase) isVectorField(fieldPath string) bool {
	vectorFields := ddb.validator.GetVectorFields()
	_, exists := vectorFields[fieldPath]
	return exists
}

func (ddb *DocumentDatabase) reconstructDocument(metadata map[string]interface{}) Document {
	doc := make(Document)

	// Copy metadata back to document
	for key, value := range metadata {
		if !strings.HasPrefix(key, "_") { // Skip internal fields
			doc[key] = value
		}
	}

	return doc
}

func (ddb *DocumentDatabase) convertFilter(where map[string]interface{}) *Filter {
	if len(where) == 0 {
		return nil
	}

	// Convert unified filter format to VittoriaDB filter format
	// This is a simplified implementation
	var filters []Filter

	for field, condition := range where {
		if condMap, ok := condition.(map[string]interface{}); ok {
			for op, value := range condMap {
				filter := Filter{
					Field:    field,
					Operator: FilterOp(op),
					Value:    value,
				}
				filters = append(filters, filter)
			}
		} else {
			// Simple equality filter
			filter := Filter{
				Field:    field,
				Operator: FilterOpEq,
				Value:    condition,
			}
			filters = append(filters, filter)
		}
	}

	if len(filters) == 1 {
		return &filters[0]
	} else if len(filters) > 1 {
		return &Filter{And: filters}
	}

	return nil
}

func (ddb *DocumentDatabase) indexTextFields(ctx context.Context, docID string, searchableText map[string]string, doc Document) error {
	// In a real implementation, this would index text in a full-text search engine
	// For now, we'll store it as metadata in a special text collection

	textCollectionName := "_text_index"
	if err := ddb.ensureTextCollection(ctx, textCollectionName); err != nil {
		return err
	}

	// Create a simple text vector for basic search (this would be replaced with proper text indexing)
	textVector := ddb.createTextVector(searchableText)

	collection, err := ddb.db.GetCollection(ctx, textCollectionName)
	if err != nil {
		return err
	}

	metadata := make(map[string]interface{})
	metadata["document_id"] = docID
	metadata["searchable_text"] = searchableText
	for key, value := range doc {
		if !ddb.isVectorField(key) {
			metadata[key] = value
		}
	}

	vectorDoc := &Vector{
		ID:       docID,
		Vector:   textVector,
		Metadata: metadata,
	}

	return collection.Insert(ctx, vectorDoc)
}

func (ddb *DocumentDatabase) ensureTextCollection(ctx context.Context, name string) error {
	_, err := ddb.db.GetCollection(ctx, name)
	if err == nil {
		return nil
	}

	req := &CreateCollectionRequest{
		Name:       name,
		Dimensions: 384, // Standard text embedding dimension
		Metric:     DistanceMetricCosine,
		IndexType:  IndexTypeFlat,
	}

	return ddb.db.CreateCollection(ctx, req)
}

func (ddb *DocumentDatabase) createTextVector(searchableText map[string]string) []float32 {
	// This is a placeholder implementation
	// In a real system, this would use proper text embeddings
	vector := make([]float32, 384)

	// Simple hash-based vector generation (for demonstration)
	allText := ""
	for _, text := range searchableText {
		allText += text + " "
	}

	hash := simpleHash(allText)
	for i := range vector {
		vector[i] = float32((hash >> (i % 32)) & 1)
	}

	return vector
}

func (ddb *DocumentDatabase) searchTextIndex(ctx context.Context, tokens []string, fields []string, params *SearchParams) ([]*DocumentSearchResult, error) {
	// Simplified text search implementation
	// In a real system, this would use a proper text search index with BM25 scoring

	textCollectionName := "_text_index"
	collection, err := ddb.db.GetCollection(ctx, textCollectionName)
	if err != nil {
		return []*DocumentSearchResult{}, nil // No text index yet
	}

	// Create query vector from search terms
	queryVector := ddb.createTextVector(map[string]string{"query": strings.Join(tokens, " ")})

	searchReq := &SearchRequest{
		Vector:          queryVector,
		Limit:           params.Limit + params.Offset,
		Offset:          0,
		IncludeMetadata: true,
		Filter:          ddb.convertFilter(params.Where),
	}

	response, err := collection.Search(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	var results []*DocumentSearchResult
	for _, result := range response.Results {
		// Apply text relevance filtering
		if ddb.isTextRelevant(tokens, result.Metadata, params.Threshold) {
			unifiedResult := &DocumentSearchResult{
				ID:       result.ID,
				Score:    result.Score,
				Document: ddb.reconstructDocument(result.Metadata),
			}
			results = append(results, unifiedResult)
		}
	}

	return results, nil
}

func (ddb *DocumentDatabase) isTextRelevant(tokens []string, metadata map[string]interface{}, threshold float64) bool {
	// Simple relevance check
	if searchableText, ok := metadata["searchable_text"].(map[string]interface{}); ok {
		for _, text := range searchableText {
			if textStr, ok := text.(string); ok {
				lowerText := strings.ToLower(textStr)
				matchCount := 0
				for _, token := range tokens {
					if strings.Contains(lowerText, strings.ToLower(token)) {
						matchCount++
					}
				}
				relevance := float64(matchCount) / float64(len(tokens))
				if relevance >= threshold {
					return true
				}
			}
		}
	}
	return threshold == 0 // If no threshold, include all results
}

func (ddb *DocumentDatabase) combineHybridResults(textResults, vectorResults []*DocumentSearchResult, weights *HybridWeights) []*DocumentSearchResult {
	if weights == nil {
		weights = &HybridWeights{Text: 0.5, Vector: 0.5}
	}

	// Combine results by document ID
	resultMap := make(map[string]*DocumentSearchResult)

	// Add text results
	for _, result := range textResults {
		result.Score = float32(float64(result.Score) * weights.Text)
		resultMap[result.ID] = result
	}

	// Add or combine vector results
	for _, result := range vectorResults {
		vectorScore := float32(float64(result.Score) * weights.Vector)
		if existing, exists := resultMap[result.ID]; exists {
			existing.Score += vectorScore
		} else {
			result.Score = vectorScore
			resultMap[result.ID] = result
		}
	}

	// Convert back to slice and sort by score
	var combined []*DocumentSearchResult
	for _, result := range resultMap {
		combined = append(combined, result)
	}

	// Sort by score (descending)
	for i := 0; i < len(combined)-1; i++ {
		for j := i + 1; j < len(combined); j++ {
			if combined[i].Score < combined[j].Score {
				combined[i], combined[j] = combined[j], combined[i]
			}
		}
	}

	return combined
}

func (ddb *DocumentDatabase) calculateFacets(results []*DocumentSearchResult, facetConfigs map[string]*FacetConfig) map[string]*FacetResult {
	facets := make(map[string]*FacetResult)

	for fieldPath := range facetConfigs {
		facetResult := &FacetResult{
			Values: make(map[string]int),
		}

		for _, result := range results {
			if value := ddb.extractFacetValue(result.Document, fieldPath); value != nil {
				key := fmt.Sprintf("%v", value)
				facetResult.Values[key]++
			}
		}

		facetResult.Count = len(facetResult.Values)
		facets[fieldPath] = facetResult
	}

	return facets
}

func (ddb *DocumentDatabase) extractFacetValue(doc Document, fieldPath string) interface{} {
	parts := strings.Split(fieldPath, ".")
	current := map[string]interface{}(doc)

	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		} else {
			if next, ok := current[part].(map[string]interface{}); ok {
				current = next
			} else {
				return nil
			}
		}
	}

	return nil
}

// Utility functions

func generateDocumentID() string {
	return fmt.Sprintf("doc_%d", time.Now().UnixNano())
}

func simpleHash(s string) uint32 {
	hash := uint32(2166136261)
	for _, c := range s {
		hash ^= uint32(c)
		hash *= 16777619
	}
	return hash
}
