# VittoriaDB Python SDK

[![PyPI version](https://badge.fury.io/py/vittoriadb.svg)](https://pypi.org/project/vittoriadb/)
[![Python versions](https://img.shields.io/pypi/pyversions/vittoriadb.svg)](https://pypi.org/project/vittoriadb/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**Python client for [VittoriaDB](https://github.com/antonellof/VittoriaDB)** — an embedded vector database shipped as a single Go binary. The SDK talks to the server over HTTP: collections, vectors, semantic search, server-side embeddings, document upload, and RAG-friendly content storage. Optional **auto-start** downloads a matching release binary (when available) and manages the process lifecycle.

---

## Documentation

| Topic | Where |
|--------|--------|
| REST API (paths, JSON bodies, filters) | [**docs/api.md**](https://github.com/antonellof/VittoriaDB/blob/main/docs/api.md) |
| Embeddings & vectorizers | [**docs/embeddings.md**](https://github.com/antonellof/VittoriaDB/blob/main/docs/embeddings.md) |
| Installation & troubleshooting | [**docs/installation.md**](https://github.com/antonellof/VittoriaDB/blob/main/docs/installation.md) |
| Server & YAML configuration | [**docs/configuration.md**](https://github.com/antonellof/VittoriaDB/blob/main/docs/configuration.md) |
| Content storage for RAG | [**docs/content-storage.md**](https://github.com/antonellof/VittoriaDB/blob/main/docs/content-storage.md) |
| CLI (`run`, `backup`, `restore`, …) | [**docs/cli.md**](https://github.com/antonellof/VittoriaDB/blob/main/docs/cli.md) |
| Examples (Python / Go / curl) | [**examples/README.md**](https://github.com/antonellof/VittoriaDB/blob/main/examples/README.md) |

---

## Features

| Area | What the SDK exposes |
|------|----------------------|
| **Collections** | Create / list / delete; `DistanceMetric`, `IndexType` (`FLAT`, `HNSW`, `IVF` prototype) |
| **Vectors** | `insert`, `insert_batch`, `get`, `delete`, `search` with optional metadata |
| **Embeddings** | `insert_text`, `insert_text_batch`, `search_text` when the collection has a `VectorizerConfig` |
| **Documents** | `upload_file` / `upload_files` — PDF, DOCX, TXT, MD, HTML (server-side processors) |
| **Metadata filters** | Flat dict (`{"category": "tech"}`) or structured filter JSON; use `search(..., use_post=True)` for `POST` search bodies |
| **Indexes** | HNSW tuning via `config`; IVF tuning via `Configure.Index.ivf_flat(nlist=, nprobe=)` |
| **Observability** | `health()`, `stats()` (`DatabaseStats`: totals, latency, QPS), `prometheus_metrics()` → `GET /metrics` text |
| **Configuration** | `config()` → unified server config + feature flags (requires compatible server) |
| **Errors** | Typed exceptions: `ConnectionError`, `CollectionError`, `VectorError`, `SearchError`, `BinaryError` |

Backup and restore are **not** HTTP APIs: use the **`vittoriadb` CLI** (`backup` / `restore`). Stop the server before restore.

---

## Installation

```bash
pip install vittoriadb
```

On install, the package tries to download a **platform binary** from GitHub Releases at tag **`v<sdk-version>`** (e.g. `v0.7.0`). If that asset is missing, you’ll see a warning — install the binary manually or put `vittoriadb` on your `PATH`.

---

## Quick start

### Connect

```python
import vittoriadb

# Managed local server (downloads binary when release exists)
db = vittoriadb.connect()

# Existing server
db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)

db.close()
```

### Vectors and similarity search

```python
collection = db.create_collection(
    name="docs",
    dimensions=384,
    metric="cosine",
)

collection.insert(
    "doc1",
    vector=[0.1] * 384,
    metadata={"title": "Hello", "category": "demo"},
)

results = collection.search(
    vector=[0.1] * 384,
    limit=5,
    filter={"category": "demo"},
    include_metadata=True,
)

# Structured / nested filters: POST search
results = collection.search(
    vector=[0.1] * 384,
    limit=5,
    use_post=True,
    filter={
        "and": [
            {"field": "category", "operator": "eq", "value": "demo"},
        ]
    },
)
```

### Server-side embeddings

```python
from vittoriadb.configure import Configure

collection = db.create_collection(
    name="smart",
    dimensions=384,
    vectorizer_config=Configure.Vectors.auto_embeddings(),
)

collection.insert_text(
    "a1",
    "Your text here.",
    metadata={"source": "wiki"},
)

hits = collection.search_text("your query", limit=5)
```

### Observability

```python
h = db.health()           # HealthStatus
s = db.stats()            # DatabaseStats: queries_total, avg_query_latency, queries_per_sec, …
text = db.prometheus_metrics()  # Prometheus exposition from GET /metrics
```

### IVF index (prototype)

```python
from vittoriadb import IndexType
from vittoriadb.configure import Configure

collection = db.create_collection(
    name="ivf_demo",
    dimensions=64,
    index_type=IndexType.IVF,
    config=Configure.Index.ivf_flat(nlist=16, nprobe=2),
)
```

---

## RAG and content storage

Enable **`ContentStorageConfig`** on `create_collection` to persist chunk text for retrieval (`include_content=True` on `search_text`). See [**docs/content-storage.md**](https://github.com/antonellof/VittoriaDB/blob/main/docs/content-storage.md).

```python
from vittoriadb import ContentStorageConfig

collection = db.create_collection(
    name="rag",
    dimensions=384,
    vectorizer_config=Configure.Vectors.auto_embeddings(),
    content_storage=ContentStorageConfig(enabled=True, field_name="_content"),
)

collection.insert_text("id1", "Long passage …", metadata={"title": "Doc"})

for r in collection.search_text("query", limit=3, include_content=True):
    if r.has_content():
        print(r.content)
```

---

## Vectorizer presets (`Configure.Vectors`)

| Method | Use case |
|--------|----------|
| `auto_embeddings(...)` | Default server-side pipeline (often Sentence Transformers–backed on server) |
| `sentence_transformers(model=, dimensions=)` | Explicit ST model |
| `openai_embeddings(api_key=, …)` | OpenAI API |
| `huggingface_embeddings(api_key=, …)` | Hugging Face Inference API |
| `ollama_embeddings(model=, dimensions=, base_url=)` | Local Ollama |

Details: [**docs/embeddings.md**](https://github.com/antonellof/VittoriaDB/blob/main/docs/embeddings.md).

---

## Index tuning

**HNSW** — pass `index_type="hnsw"` or `IndexType.HNSW` and `config` keys such as `m`, `ef_construction`, `ef_search` (see API / server defaults).

**IVF** — `index_type=IndexType.IVF` and `config=Configure.Index.ivf_flat(nlist=16, nprobe=2)` (prototype index on server).

---

## Configuration inspection

```python
info = db.config()   # dict: unified config + metadata from GET /config
```

---

## API overview

### `VittoriaDB`

| Method | Description |
|--------|-------------|
| `connect(url=, auto_start=, port=, host=, data_dir=, extra_args=)` | Client instance |
| `create_collection(...)` | New collection |
| `get_collection(name)` | Handle to existing collection |
| `list_collections()` | List `CollectionInfo` |
| `delete_collection(name)` | Drop collection |
| `health()` | `HealthStatus` |
| `stats()` | `DatabaseStats` |
| `prometheus_metrics()` | Raw Prometheus text |
| `config()` | Server configuration JSON |
| `close()` | Cleanup / stop auto-started server |

### `Collection`

| Method | Description |
|--------|-------------|
| `insert(id, vector, metadata=)` | Single vector |
| `insert_batch(vectors)` | Batch |
| `insert_text` / `insert_text_batch` | Server-side embedding |
| `search(vector, limit=, filter=, use_post=, …)` | Similarity search |
| `search_text(query, limit=, filter=, …)` | Query embedding + search |
| `upload_file` / `upload_files` | Document → chunks + vectors |
| `get` / `delete` / `count` | CRUD helpers |

Exported types include **`Vector`**, **`SearchResult`**, **`CollectionInfo`**, **`HealthStatus`**, **`DatabaseStats`**, **`DistanceMetric`**, **`IndexType`**, **`VectorizerConfig`**, **`ContentStorageConfig`**, and configuration/error types listed in **`vittoriadb.__all__`**.

---

## Contributing & license

- Issues: [GitHub Issues](https://github.com/antonellof/VittoriaDB/issues)
- Dev setup: [DEVELOPMENT.md](DEVELOPMENT.md)
- License: [MIT](https://github.com/antonellof/VittoriaDB/blob/main/LICENSE)

---

## Changelog (high level)

- **0.7.0** — Aligned with **VittoriaDB v0.7.0** GitHub release assets; same feature set as **0.6.x** (filters, `use_post`, `prometheus_metrics`, `Configure.Index.ivf_flat`, `HealthStatus` / `DatabaseStats`); documentation table and API overview updated.
- **0.6.x** — Metadata filters, observability helpers, IVF config in client; README refresh.
- **0.5.x** — Unified server configuration; `config()`; performance-related server features.

Full release notes: [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases).
