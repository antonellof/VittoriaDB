package index

import (
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// HNSWIndexImpl implements the HNSW (Hierarchical Navigable Small World) algorithm
type HNSWIndexImpl struct {
	nodes      map[string]*HNSWNode
	entryPoint *HNSWNode
	dimensions int
	metric     DistanceMetric
	calculator DistanceCalculator
	config     *HNSWConfig
	mu         sync.RWMutex
	rng        *rand.Rand
	stats      *IndexStats
	maxLayer   int
}

// NewHNSWIndex creates a new HNSW index
func NewHNSWIndex(dimensions int, metric DistanceMetric, config *HNSWConfig) HNSWIndex {
	if config == nil {
		config = DefaultHNSWConfig()
	}

	return &HNSWIndexImpl{
		nodes:      make(map[string]*HNSWNode),
		dimensions: dimensions,
		metric:     metric,
		calculator: NewDistanceCalculator(metric),
		config:     config,
		rng:        rand.New(rand.NewSource(config.Seed)),
		stats: &IndexStats{
			IndexType:   IndexTypeHNSW,
			Dimensions:  dimensions,
			VectorCount: 0,
		},
	}
}

// Build builds the HNSW index from a set of vectors
func (idx *HNSWIndexImpl) Build(vectors []*IndexVector) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	startTime := time.Now()

	// Clear existing index
	idx.nodes = make(map[string]*HNSWNode)
	idx.entryPoint = nil
	idx.maxLayer = 0

	// Add vectors one by one
	for i, vector := range vectors {
		if len(vector.Vector) != idx.dimensions {
			return fmt.Errorf("vector %d has wrong dimensions: expected %d, got %d",
				i, idx.dimensions, len(vector.Vector))
		}

		if err := idx.addVector(vector); err != nil {
			return fmt.Errorf("failed to add vector %d: %w", i, err)
		}
	}

	// Update stats
	idx.stats.VectorCount = len(idx.nodes)
	idx.stats.BuildTime = time.Since(startTime).Milliseconds()
	idx.stats.MaxLayer = idx.maxLayer
	idx.stats.AvgDegree = idx.calculateAverageDegree()

	return nil
}

// Load loads the index from a reader
func (idx *HNSWIndexImpl) Load(r io.Reader) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	decoder := json.NewDecoder(r)

	var data struct {
		Nodes      map[string]*HNSWNode `json:"nodes"`
		EntryPoint string               `json:"entry_point"`
		Dimensions int                  `json:"dimensions"`
		Metric     DistanceMetric       `json:"metric"`
		Config     *HNSWConfig          `json:"config"`
		MaxLayer   int                  `json:"max_layer"`
		Stats      *IndexStats          `json:"stats"`
	}

	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to decode HNSW index: %w", err)
	}

	// Validate
	if data.Dimensions != idx.dimensions {
		return fmt.Errorf("dimension mismatch: expected %d, got %d",
			idx.dimensions, data.Dimensions)
	}
	if data.Metric != idx.metric {
		return fmt.Errorf("metric mismatch: expected %s, got %s",
			idx.metric.String(), data.Metric.String())
	}

	idx.nodes = data.Nodes
	idx.maxLayer = data.MaxLayer
	idx.stats = data.Stats

	// Set entry point
	if data.EntryPoint != "" {
		if node, exists := idx.nodes[data.EntryPoint]; exists {
			idx.entryPoint = node
		}
	}

	return nil
}

// Save saves the index to a writer
func (idx *HNSWIndexImpl) Save(w io.Writer) error {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	encoder := json.NewEncoder(w)

	entryPointID := ""
	if idx.entryPoint != nil {
		entryPointID = idx.entryPoint.ID
	}

	data := struct {
		Nodes      map[string]*HNSWNode `json:"nodes"`
		EntryPoint string               `json:"entry_point"`
		Dimensions int                  `json:"dimensions"`
		Metric     DistanceMetric       `json:"metric"`
		Config     *HNSWConfig          `json:"config"`
		MaxLayer   int                  `json:"max_layer"`
		Stats      *IndexStats          `json:"stats"`
	}{
		Nodes:      idx.nodes,
		EntryPoint: entryPointID,
		Dimensions: idx.dimensions,
		Metric:     idx.metric,
		Config:     idx.config,
		MaxLayer:   idx.maxLayer,
		Stats:      idx.stats,
	}

	return encoder.Encode(data)
}

// Add adds a vector to the index
func (idx *HNSWIndexImpl) Add(ctx context.Context, vector *IndexVector) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Validate vector
	if len(vector.Vector) != idx.dimensions {
		return fmt.Errorf("vector has wrong dimensions: expected %d, got %d",
			idx.dimensions, len(vector.Vector))
	}

	// Check for duplicate ID
	if _, exists := idx.nodes[vector.ID]; exists {
		return fmt.Errorf("vector with ID %s already exists", vector.ID)
	}

	return idx.addVector(vector)
}

// Delete removes a vector from the index
func (idx *HNSWIndexImpl) Delete(ctx context.Context, id string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	node, exists := idx.nodes[id]
	if !exists {
		return fmt.Errorf("vector with ID %s not found", id)
	}

	// Remove connections to this node from other nodes
	for layer := 0; layer <= node.Layer; layer++ {
		if connections, hasLayer := node.Connections[layer]; hasLayer {
			for _, connID := range connections {
				if connNode, exists := idx.nodes[connID]; exists {
					idx.removeConnection(connNode, id, layer)
				}
			}
		}
	}

	// Remove the node
	delete(idx.nodes, id)

	// Update entry point if necessary
	if idx.entryPoint != nil && idx.entryPoint.ID == id {
		idx.findNewEntryPoint()
	}

	idx.stats.VectorCount = len(idx.nodes)
	return nil
}

// Search performs k-nearest neighbor search using HNSW algorithm
func (idx *HNSWIndexImpl) Search(ctx context.Context, query []float32, k int, params *SearchParams) ([]*Candidate, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	startTime := time.Now()

	// Validate query
	if len(query) != idx.dimensions {
		return nil, fmt.Errorf("query vector has wrong dimensions: expected %d, got %d",
			idx.dimensions, len(query))
	}

	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}

	if len(idx.nodes) == 0 {
		return []*Candidate{}, nil
	}

	// Get search parameters
	ef := idx.config.EfSearch
	if params != nil && params.EF > 0 {
		ef = params.EF
	}

	// Start from entry point
	if idx.entryPoint == nil {
		return []*Candidate{}, nil
	}

	// Search from top layer down to layer 1
	entryPoints := []*QueueItem{{
		ID:       idx.entryPoint.ID,
		Distance: idx.calculator.Calculate(query, idx.entryPoint.Vector),
		Vector:   idx.entryPoint.Vector,
	}}

	for layer := idx.maxLayer; layer >= 1; layer-- {
		entryPoints = idx.searchLayer(query, entryPoints, 1, layer)
	}

	// Search layer 0 with ef
	candidates := idx.searchLayer(query, entryPoints, ef, 0)

	// Convert to results and limit to k
	results := make([]*Candidate, 0, k)
	for i, candidate := range candidates {
		if i >= k {
			break
		}
		results = append(results, &Candidate{
			ID:    candidate.ID,
			Score: candidate.Distance,
		})
	}

	// Update search latency stats
	latency := time.Since(startTime).Seconds() * 1000
	idx.stats.SearchLatencyP50 = latency // Simplified
	idx.stats.SearchLatencyP99 = latency

	return results, nil
}

// Size returns the number of vectors in the index
func (idx *HNSWIndexImpl) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.nodes)
}

// Dimensions returns the vector dimensions
func (idx *HNSWIndexImpl) Dimensions() int {
	return idx.dimensions
}

// Type returns the index type
func (idx *HNSWIndexImpl) Type() IndexType {
	return IndexTypeHNSW
}

// Optimize optimizes the index
func (idx *HNSWIndexImpl) Optimize() error {
	// HNSW doesn't need explicit optimization
	return nil
}

// Stats returns index statistics
func (idx *HNSWIndexImpl) Stats() *IndexStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Calculate memory usage
	vectorMemory := int64(len(idx.nodes)) * int64(idx.dimensions) * 4 // 4 bytes per float32
	connectionMemory := int64(0)

	for _, node := range idx.nodes {
		for _, connections := range node.Connections {
			connectionMemory += int64(len(connections)) * 8 // 8 bytes per string pointer (approximate)
		}
	}

	stats := *idx.stats
	stats.MemoryUsage = vectorMemory + connectionMemory
	stats.VectorCount = len(idx.nodes)
	stats.MaxLayer = idx.maxLayer
	stats.AvgDegree = idx.calculateAverageDegree()

	return &stats
}

// HNSW-specific methods

// GetNode returns a node by ID
func (idx *HNSWIndexImpl) GetNode(id string) *HNSWNode {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.nodes[id]
}

// GetConnections returns connections for a node at a specific layer
func (idx *HNSWIndexImpl) GetConnections(id string, layer int) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if node, exists := idx.nodes[id]; exists {
		if connections, hasLayer := node.Connections[layer]; hasLayer {
			return connections
		}
	}
	return []string{}
}

// SetEfSearch sets the search parameter ef
func (idx *HNSWIndexImpl) SetEfSearch(ef int) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.config.EfSearch = ef
}

// Private methods

func (idx *HNSWIndexImpl) addVector(vector *IndexVector) error {
	// Determine layer for new node
	layer := idx.randomLevel()

	// Create new node
	node := &HNSWNode{
		ID:          vector.ID,
		Vector:      make([]float32, len(vector.Vector)),
		Layer:       layer,
		Connections: make(map[int][]string),
	}
	copy(node.Vector, vector.Vector)

	// Initialize connections for each layer
	for l := 0; l <= layer; l++ {
		node.Connections[l] = make([]string, 0)
	}

	// If this is the first node, make it the entry point
	if idx.entryPoint == nil {
		idx.entryPoint = node
		idx.maxLayer = layer
		idx.nodes[vector.ID] = node
		return nil
	}

	// Search for closest nodes starting from entry point
	entryPoints := []*QueueItem{{
		ID:       idx.entryPoint.ID,
		Distance: idx.calculator.Calculate(node.Vector, idx.entryPoint.Vector),
		Vector:   idx.entryPoint.Vector,
	}}

	// Search from top layer down to layer+1
	for l := idx.maxLayer; l >= layer+1; l-- {
		entryPoints = idx.searchLayer(node.Vector, entryPoints, 1, l)
	}

	// Search and connect at each layer from layer down to 0
	for l := min(layer, idx.maxLayer); l >= 0; l-- {
		candidates := idx.searchLayer(node.Vector, entryPoints, idx.config.EfConstruction, l)

		// Select neighbors
		maxConn := idx.config.MaxM
		if l == 0 {
			maxConn = idx.config.MaxM0
		}

		neighbors := idx.selectNeighbors(candidates, maxConn)

		// Add connections
		for _, neighbor := range neighbors {
			idx.addConnection(node, neighbor.ID, l)
			idx.addConnection(idx.nodes[neighbor.ID], node.ID, l)

			// Prune connections if necessary
			if len(idx.nodes[neighbor.ID].Connections[l]) > maxConn {
				idx.pruneConnections(idx.nodes[neighbor.ID], l, maxConn)
			}
		}

		entryPoints = neighbors
	}

	// Update entry point if new node has higher layer
	if layer > idx.maxLayer {
		idx.entryPoint = node
		idx.maxLayer = layer
	}

	idx.nodes[vector.ID] = node
	return nil
}

func (idx *HNSWIndexImpl) randomLevel() int {
	level := 0
	for idx.rng.Float64() < idx.config.ML && level < 16 { // Cap at 16 layers
		level++
	}
	return level
}

func (idx *HNSWIndexImpl) searchLayer(query []float32, entryPoints []*QueueItem, ef int, layer int) []*QueueItem {
	visited := make(map[string]bool)
	candidates := &PriorityQueue{}
	w := &PriorityQueue{}

	// Initialize with entry points
	for _, ep := range entryPoints {
		heap.Push(candidates, &QueueItem{
			ID:       ep.ID,
			Distance: ep.Distance,
			Vector:   ep.Vector,
		})
		heap.Push(w, &QueueItem{
			ID:       ep.ID,
			Distance: -ep.Distance, // Max heap for w
			Vector:   ep.Vector,
		})
		visited[ep.ID] = true
	}

	for candidates.Len() > 0 {
		current := heap.Pop(candidates).(*QueueItem)

		// Check if we should continue
		if w.Len() > 0 && current.Distance > -(*w)[0].Distance {
			break
		}

		// Explore neighbors
		if node, exists := idx.nodes[current.ID]; exists {
			if connections, hasLayer := node.Connections[layer]; hasLayer {
				for _, neighborID := range connections {
					if !visited[neighborID] {
						visited[neighborID] = true

						if neighbor, exists := idx.nodes[neighborID]; exists {
							distance := idx.calculator.Calculate(query, neighbor.Vector)

							if w.Len() < ef || distance < -(*w)[0].Distance {
								heap.Push(candidates, &QueueItem{
									ID:       neighborID,
									Distance: distance,
									Vector:   neighbor.Vector,
								})
								heap.Push(w, &QueueItem{
									ID:       neighborID,
									Distance: -distance,
									Vector:   neighbor.Vector,
								})

								if w.Len() > ef {
									heap.Pop(w)
								}
							}
						}
					}
				}
			}
		}
	}

	// Convert w to sorted slice
	result := make([]*QueueItem, w.Len())
	for i := len(result) - 1; i >= 0; i-- {
		item := heap.Pop(w).(*QueueItem)
		item.Distance = -item.Distance // Convert back to min distance
		result[i] = item
	}

	return result
}

func (idx *HNSWIndexImpl) selectNeighbors(candidates []*QueueItem, m int) []*QueueItem {
	if len(candidates) <= m {
		return candidates
	}

	// Simple selection - take closest m neighbors
	// In a more sophisticated implementation, this would use heuristics
	// to maintain connectivity and avoid hubs
	return candidates[:m]
}

func (idx *HNSWIndexImpl) addConnection(node *HNSWNode, neighborID string, layer int) {
	if connections, hasLayer := node.Connections[layer]; hasLayer {
		// Check if connection already exists
		for _, existing := range connections {
			if existing == neighborID {
				return
			}
		}
		node.Connections[layer] = append(connections, neighborID)
	}
}

func (idx *HNSWIndexImpl) removeConnection(node *HNSWNode, neighborID string, layer int) {
	if connections, hasLayer := node.Connections[layer]; hasLayer {
		for i, existing := range connections {
			if existing == neighborID {
				// Remove by swapping with last element
				connections[i] = connections[len(connections)-1]
				node.Connections[layer] = connections[:len(connections)-1]
				return
			}
		}
	}
}

func (idx *HNSWIndexImpl) pruneConnections(node *HNSWNode, layer int, maxConn int) {
	if connections, hasLayer := node.Connections[layer]; hasLayer && len(connections) > maxConn {
		// Simple pruning - keep closest neighbors
		// In practice, this should use more sophisticated heuristics
		candidates := make([]*QueueItem, 0, len(connections))
		for _, connID := range connections {
			if neighbor, exists := idx.nodes[connID]; exists {
				candidates = append(candidates, &QueueItem{
					ID:       connID,
					Distance: idx.calculator.Calculate(node.Vector, neighbor.Vector),
				})
			}
		}

		if len(candidates) > 0 {
			sort.Slice(candidates, func(i, j int) bool {
				return candidates[i].Distance < candidates[j].Distance
			})

			newConnections := make([]string, 0, maxConn)
			for i := 0; i < maxConn && i < len(candidates); i++ {
				newConnections = append(newConnections, candidates[i].ID)
			}
			node.Connections[layer] = newConnections
		}
	}
}

func (idx *HNSWIndexImpl) findNewEntryPoint() {
	maxLayer := -1
	var newEntryPoint *HNSWNode

	for _, node := range idx.nodes {
		if node.Layer > maxLayer {
			maxLayer = node.Layer
			newEntryPoint = node
		}
	}

	idx.entryPoint = newEntryPoint
	idx.maxLayer = maxLayer
}

func (idx *HNSWIndexImpl) calculateAverageDegree() float64 {
	if len(idx.nodes) == 0 {
		return 0
	}

	totalDegree := 0
	for _, node := range idx.nodes {
		for _, connections := range node.Connections {
			totalDegree += len(connections)
		}
	}

	return float64(totalDegree) / float64(len(idx.nodes))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
