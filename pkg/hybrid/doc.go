// Package hybrid defines scaffolding for hybrid retrieval (dense vectors +
// lexical signals). It is **not** wired into the HTTP server yet; future work:
//
//   - Maintain an inverted index (BM25 or FTS) alongside vector storage for
//     chunks that expose text fields.
//   - Support Reciprocal Rank Fusion (RRF) or weighted linear combination of
//     normalized BM25 and cosine scores (see HybridWeights).
//   - Optional sparse vector channel (compatible with Pinecone/Qdrant-style
//     workflows) once the storage layer exposes fixed-size sparse payloads.
//
// Competitive baseline (2025–2026): Weaviate (BM25 + vector), Qdrant (sparse),
// Pinecone (sparse-dense). VittoriaDB remains single-node first; hybrid is a
// major initiative layered on existing content_storage metadata.
package hybrid
