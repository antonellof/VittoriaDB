# VittoriaDB v0.7.0 — Durability, ops, filters, IVF prototype, SDK refresh

Focus: **production-oriented single-node** behaviour (WAL replay, backup/restore), clearer **REST filter** handling, **observability**, an **IVF index spike**, plus **Python SDK** parity and docs.

## Highlights

### Storage & recovery
- **WAL replay**: Crash recovery reapplies WAL entries; sequence preserved; header and page cache aligned with replay semantics (`pkg/storage`).
- **Page deserialization fix**: Stable header parsing after replay-related hardening.

### Backup & CLI
- **Backup / restore MVP**: gzip+tar streaming of the data directory from core; **`vittoriadb backup`** and **`vittoriadb restore`** (restore requires DB closed).

### REST API & search
- **Metadata filters**: **`ParseFilterJSON`** accepts both **flat maps** (`{"k":"v"}` → implicit equality) and **structured** `Filter` trees; wired for GET query `filter=` and POST search bodies (`pkg/core/filter_compat.go`, server).
- **Prometheus `/metrics`**: Search counters and latency gauges for operators (`pkg/server`, `pkg/core/metrics.go`).
- **`/stats`**: Reports **queries_total**, average query latency, approximate QPS (`pkg/core/database.go`).

### Indexes
- **IVF prototype**: **`IndexTypeIVF`** wired in **`pkg/index/factory.go`** with **`nlist` / `nprobe`**; IVF-flat index spike (`pkg/index/ivf.go`). Add/delete on IVF remain limited by design for this milestone.

### Research scaffolding
- **`pkg/hybrid`**: Types and docs for future BM25 / fusion-style hybrid retrieval (not wired to HTTP).

### Benchmarks (Apple M2 Pro, `pkg/index`, `-benchtime=2s`)

```bash
go test -bench=. -benchmem -benchtime=2s -run=^$ ./pkg/index/
```

| Benchmark | Approx. throughput |
|-----------|-------------------|
| HNSW insert | ~1.3k vectors/sec |
| HNSW search (10k, k=10) | ~7k queries/sec |
| Flat search (10k, k=10) | ~186 queries/sec |

See **`docs/performance.md`** for methodology.

## Python SDK (`vittoriadb==0.7.0`)

- **`prometheus_metrics()`**, **`stats()`** / **`health()`**, **`HealthStatus`** / **`DatabaseStats`** exports.
- **`collection.search(..., use_post=True)`** for structured filters via POST.
- **`Configure.Index.ivf_flat(...)`** for IVF collection config.
- **`README.md`** on PyPI aligned with REST docs and features.

Binary download during `pip install` resolves **`vittoriadb-v0.7.0-<platform>`** assets from this GitHub release.

## Upgrade notes

- After upgrade, run a normal workload; WAL replay runs when opening existing data dirs if applicable.
- For restore, stop the server before **`vittoriadb restore`**.
- IVF is **experimental**: prefer **HNSW** / **Flat** for production until IVF is hardened.
