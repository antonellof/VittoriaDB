// Filtered vector search (design note)
//
// The current VittoriaCollection search path (see parallel_search.go and
// legacySearch) enumerates stored vectors, applies metadata filters, then
// scores remaining candidates. This is a correct post-filter strategy for
// small/medium in-memory collections and matches the "brute force with
// filter" baseline.
//
// When the HNSW graph in pkg/index is integrated as the ANN front-end, a
// production system typically combines:
//   - pre-filter: build a candidate set from a filter index, then HNSW; or
//   - post-filter: HNSW with over-fetch, then apply filter to the top M; or
//   - Qdrant-style index interleaving (not yet present here).
//
// Use search_filter_bench_test.go to compare filter selectivity vs no filter.

package core
