# VittoriaDB v0.6.0 — Cleanup, real HuggingFace, reproducible benchmarks

This release is mostly about **trust**: removing dead/misleading code, making
the public API match what the engine actually does, and shipping a real,
checked-in benchmark suite so the performance numbers in the docs stop being
"trust me".

## Highlights

- **Real HuggingFace Inference API client.** `Configure.Vectors.huggingface_embeddings(...)`
  used to return `not yet implemented`. It now hits
  `https://api-inference.huggingface.co/pipeline/feature-extraction/<model>`,
  handles batched requests, optional API token, configurable endpoint and
  timeout, and both sentence-level and token-level (mean-pooled) response
  shapes.
- **Reproducible benchmarks.** New `pkg/index/bench_test.go` runs HNSW insert,
  HNSW search and flat search benchmarks via `go test -bench`. Numbers below
  are now from this checked-in suite, not from screenshots.
- **Critical I/O fix.** `pkg/core/io_optimizer.go` was truncating `float32`
  values in the async vector path because of a bad `uint32(v)` cast. It now
  uses `math.Float32bits` + `binary.LittleEndian` for proper IEEE-754
  round-tripping.
- **Python SDK upload fixed.** `Collection.upload_file()` was POSTing to
  `/collections/{name}/upload`; the actual server route is
  `/collections/{name}/documents`. Uploads now work out of the box.

## Reproducible benchmark numbers (Apple M2 Pro, dim=384, cosine)

```bash
go test -bench=. -benchmem -benchtime=3s -run=^$ ./pkg/index/
```

| Operation | Throughput | Latency / op |
|---|---|---|
| HNSW Add (M=16, efC=200) | 1,296 vectors/sec | 772 µs |
| HNSW Search (n=10k, k=10) | 7,183 queries/sec | 139 µs |
| Flat Search (n=10k, k=10) | 187 queries/sec | 5.34 ms |

HNSW is ~38× faster than flat at this scale. See `docs/performance.md` for a
side-by-side comparison with Qdrant / Milvus / Weaviate / Chroma.

## Cleanups (no behavior change for end users)

- Removed dead AVX2 placeholders (`OptimizedDotProduct`, `OptimizedCosineSimilarity`,
  `useSIMD`) from `pkg/index/distance.go` — there was no real SIMD path
  behind them. Real chunked Go implementations live in `pkg/core/simd.go`.
- `pkg/core/collection.go`: replaced an O(n²) bubble sort over candidates
  with `sort.Slice`, dropped a duplicate `sqrt` and unified distance functions
  on `math.Sqrt`, removed dead "compression" branches in `InsertText` /
  `InsertTextBatch`.
- `pkg/processor/factory.go`: stopped advertising `.doc` and `.rtf` (no
  processor handled them); the DOCX processor now only claims `.docx`. The
  `ProcessorInfo` table no longer lists "placeholder" / "not_implemented"
  rows.
- `pkg/processor/{pdf,docx}.go`: comments now describe what's actually
  implemented (PDF via `github.com/ledongthuc/pdf`, DOCX via stdlib
  `archive/zip` + `encoding/xml`) instead of saying "placeholder".
- `pkg/server/server.go`: clarified that document upload to a collection
  without a vectorizer falls back to zero vectors that the client is expected
  to overwrite later.
- Removed dev artifacts at repo root: `test-config.yaml`,
  `large_test_document.txt`, `releases/RELEASE_NOTES_v0.3.0.md`,
  `test_data/`.

## Python SDK

- Fix: `Collection.upload_file()` and `Collection.process_text()` now hit
  `/collections/{name}/documents`.
- Removed `vittoriadb.embed`, `vittoriadb.extract_text`,
  `vittoriadb.available_models` (all of which were `NotImplementedError` /
  hardcoded stubs).
- `supported_formats()` now matches the server's actual processor list
  (no more `.rtf`).
- `process_text(...)` no longer accepts the unsupported `embedding_model` /
  `overlap` kwargs; use `chunk_overlap` and configure embeddings on the
  collection instead.
- Bumped to `0.6.0`.

## Docs & examples

- `README.md`: replaced the broken `./start.sh` reference with the actual
  `./run-dev.sh` / `./docker-start.sh` entry points; softened the SIMD
  performance claims to reflect the chunked Go implementations (true CPU
  intrinsics are still planned).
- `docs/performance.md`: rewritten around the new reproducible benchmarks
  and a competitive comparison with Qdrant / Milvus / Weaviate / Chroma.
- `examples/README.md`: removed references to non-existent files
  (`10_local_vectorizer_validation_test.py`), updated the supported-format
  table to match the engine.
- `examples/python/11_all_vectorizers_comparison.py`: now reads
  `OPENAI_API_KEY` / `HUGGINGFACE_API_KEY` from the environment instead of
  passing `"dummy_key"`.

## Migration notes

- If you were using `Collection.upload_file()` against a custom server that
  exposed `/collections/{name}/upload`, switch to `/collections/{name}/documents`.
- `Configure.Vectors.huggingface_embeddings(...)` will now actually try to
  call the HF Inference API. If you were relying on the previous
  `not yet implemented` error to gate behavior, switch to
  `Configure.Vectors.sentence_transformers()` for an offline equivalent.
- `process_text(...)` no longer silently accepts `embedding_model=` /
  `overlap=`. Pass `chunk_overlap=` and configure the vectorizer on the
  collection.
