package index

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"sync"
	"time"
)

// IVFConfig controls coarse quantization for IVFFlatIndex (prototype).
type IVFConfig struct {
	NList  int // Number of centroid clusters (≤ vector count)
	NProbe int // Centroids to search per query
}

// DefaultIVFConfig returns sensible defaults for medium datasets.
func DefaultIVFConfig() *IVFConfig {
	return &IVFConfig{
		NList:  16,
		NProbe: 2,
	}
}

// IVFFlatIndex is a minimal inverted-file index: vectors are assigned to the
// nearest centroid; search probes the closest nprobe centroids then scans
// candidates (exact within probed lists).
type IVFFlatIndex struct {
	mu         sync.RWMutex
	dimensions int
	metric     DistanceMetric
	calc       DistanceCalculator
	cfg        *IVFConfig
	centroids  [][]float32
	lists      [][]string
	byID       map[string][]float32
	stats      *IndexStats
}

// NewIVFFlatIndex creates an IVF index (IVF + flat scan within lists).
func NewIVFFlatIndex(dimensions int, metric DistanceMetric, cfg *IVFConfig) *IVFFlatIndex {
	if cfg == nil {
		cfg = DefaultIVFConfig()
	}
	return &IVFFlatIndex{
		dimensions: dimensions,
		metric:     metric,
		calc:       NewDistanceCalculator(metric),
		cfg:        cfg,
		byID:       make(map[string][]float32),
		stats: &IndexStats{
			IndexType:  IndexTypeIVF,
			Dimensions: dimensions,
		},
	}
}

// Build clusters vectors by picking evenly spaced seeds as centroids (spike).
func (idx *IVFFlatIndex) Build(vectors []*IndexVector) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	start := time.Now()
	n := len(vectors)
	if n == 0 {
		idx.centroids, idx.lists = nil, nil
		idx.byID = make(map[string][]float32)
		idx.stats.VectorCount = 0
		idx.stats.BuildTime = time.Since(start).Milliseconds()
		return nil
	}

	nlist := idx.cfg.NList
	if nlist < 1 {
		nlist = 1
	}
	if nlist > n {
		nlist = n
	}

	centroids := make([][]float32, nlist)
	step := n / nlist
	if step < 1 {
		step = 1
	}
	for i := 0; i < nlist; i++ {
		src := i * step
		if src >= n {
			src = n - 1
		}
		v := vectors[src].Vector
		c := make([]float32, len(v))
		copy(c, v)
		centroids[i] = c
	}

	lists := make([][]string, nlist)
	byID := make(map[string][]float32, n)

	for _, iv := range vectors {
		if len(iv.Vector) != idx.dimensions {
			return fmt.Errorf("dimension mismatch for id %s", iv.ID)
		}
		bestJ := 0
		bestD := float32(math.MaxFloat32)
		for j := range centroids {
			d := idx.calc.Calculate(iv.Vector, centroids[j])
			if d < bestD {
				bestD, bestJ = d, j
			}
		}
		lists[bestJ] = append(lists[bestJ], iv.ID)
		cp := make([]float32, len(iv.Vector))
		copy(cp, iv.Vector)
		byID[iv.ID] = cp
	}

	idx.centroids = centroids
	idx.lists = lists
	idx.byID = byID
	idx.stats.VectorCount = n
	idx.stats.BuildTime = time.Since(start).Milliseconds()
	return nil
}

func (idx *IVFFlatIndex) Load(r io.Reader) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	var data struct {
		Dimensions int             `json:"dimensions"`
		Metric     DistanceMetric  `json:"metric"`
		Cfg        *IVFConfig      `json:"cfg"`
		Centroids  [][]float32     `json:"centroids"`
		Lists      [][]string      `json:"lists"`
		ByID       map[string][]float32 `json:"by_id"`
		Stats      *IndexStats     `json:"stats"`
	}
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return err
	}
	if data.Dimensions != idx.dimensions || data.Metric != idx.metric {
		return fmt.Errorf("ivf load: dimension/metric mismatch")
	}
	idx.cfg = data.Cfg
	idx.centroids = data.Centroids
	idx.lists = data.Lists
	idx.byID = data.ByID
	idx.stats = data.Stats
	return nil
}

func (idx *IVFFlatIndex) Save(w io.Writer) error {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return json.NewEncoder(w).Encode(struct {
		Dimensions int                `json:"dimensions"`
		Metric     DistanceMetric     `json:"metric"`
		Cfg        *IVFConfig         `json:"cfg"`
		Centroids  [][]float32        `json:"centroids"`
		Lists      [][]string         `json:"lists"`
		ByID       map[string][]float32 `json:"by_id"`
		Stats      *IndexStats        `json:"stats"`
	}{
		Dimensions: idx.dimensions,
		Metric:     idx.metric,
		Cfg:        idx.cfg,
		Centroids:  idx.centroids,
		Lists:      idx.lists,
		ByID:       idx.byID,
		Stats:      idx.stats,
	})
}

func (idx *IVFFlatIndex) Add(ctx context.Context, vector *IndexVector) error {
	return fmt.Errorf("IVF index: use Build or recreate for Add in prototype")
}

func (idx *IVFFlatIndex) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("IVF index: delete not supported in prototype")
}

func (idx *IVFFlatIndex) Search(ctx context.Context, query []float32, k int, params *SearchParams) ([]*Candidate, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(query) != idx.dimensions {
		return nil, fmt.Errorf("query dimensions mismatch")
	}
	if k <= 0 || len(idx.byID) == 0 {
		return []*Candidate{}, nil
	}

	nprobe := idx.cfg.NProbe
	if params != nil && params.NProbes > 0 {
		nprobe = params.NProbes
	}
	if nprobe < 1 {
		nprobe = 1
	}
	if nprobe > len(idx.centroids) {
		nprobe = len(idx.centroids)
	}

	type cpair struct {
		j int
		d float32
	}
	cdist := make([]cpair, len(idx.centroids))
	for j, c := range idx.centroids {
		cdist[j] = cpair{j, idx.calc.Calculate(query, c)}
	}
	sort.Slice(cdist, func(i, j int) bool { return cdist[i].d < cdist[j].d })

	seen := make(map[string]struct{})
	cands := make([]*Candidate, 0, k*8)

	for p := 0; p < nprobe && p < len(cdist); p++ {
		j := cdist[p].j
		for _, id := range idx.lists[j] {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			vec := idx.byID[id]
			cands = append(cands, &Candidate{
				ID:    id,
				Score: idx.calc.Calculate(query, vec),
			})
		}
	}

	sort.Slice(cands, func(i, j int) bool { return cands[i].Score < cands[j].Score })
	if k > len(cands) {
		k = len(cands)
	}
	return cands[:k], nil
}

func (idx *IVFFlatIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.byID)
}

func (idx *IVFFlatIndex) Dimensions() int { return idx.dimensions }

func (idx *IVFFlatIndex) Type() IndexType { return IndexTypeIVF }

func (idx *IVFFlatIndex) Optimize() error { return nil }

func (idx *IVFFlatIndex) Stats() *IndexStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	s := *idx.stats
	var mem int64
	for _, v := range idx.byID {
		mem += int64(len(v) * 4)
	}
	s.MemoryUsage = mem
	return &s
}
