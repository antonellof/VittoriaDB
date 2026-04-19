#!/usr/bin/env python3
"""
Observability (:func:`prometheus_metrics`, ``/stats`` query fields) and IVF index.

Demonstrates:
- Prometheus text from ``GET /metrics`` (search counters, latency, approximate QPS)
- Database statistics including ``queries_total``, ``avg_query_latency``, ``queries_per_sec``
- Creating an IVF collection with ``Configure.Index.ivf_flat()`` (prototype index)
- Metadata filter: flat map vs structured filter via POST search

Requirements:
    pip install vittoriadb numpy

Usage:
    # Terminal 1: ./vittoriadb run --data-dir ./data
    python examples/python/15_observability_metrics_ivf.py

Backup and restore are **CLI-only** (server must be stopped before restore):

    ./vittoriadb backup --data-dir ./data --output backup.tar.gz
    ./vittoriadb restore --data-dir ./data-restored --input backup.tar.gz
"""

from __future__ import annotations

import numpy as np

import vittoriadb
from vittoriadb.configure import Configure


def main() -> None:
    print("Observability + IVF example")
    print("=" * 44)

    db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)

    coll_name = "metrics_ivf_demo"
    try:
        db.delete_collection(coll_name)
    except vittoriadb.CollectionError:
        pass

    dim = 64
    cfg = Configure.Index.ivf_flat(nlist=8, nprobe=2)
    collection = db.create_collection(
        name=coll_name,
        dimensions=dim,
        metric="cosine",
        index_type=vittoriadb.IndexType.IVF,
        config=cfg,
    )

    rng = np.random.default_rng(42)
    for i in range(48):
        vec = rng.random(dim, dtype=np.float32).tolist()
        collection.insert(
            f"id_{i}",
            vec,
            metadata={
                "partition": "a" if i % 2 == 0 else "b",
                "idx": i,
            },
        )

    q = rng.random(dim, dtype=np.float32).tolist()

    flat_filtered = collection.search(
        vector=q,
        limit=5,
        filter={"partition": "a"},
        include_metadata=True,
    )
    print(f"Flat-map filter (partition=a): {len(flat_filtered)} hits")

    structured = collection.search(
        vector=q,
        limit=5,
        use_post=True,
        filter={
            "and": [
                {"field": "partition", "operator": "eq", "value": "b"},
                {"field": "idx", "operator": "gte", "value": 10},
            ]
        },
        include_metadata=True,
    )
    print(f"Structured filter (POST): {len(structured)} hits")

    stats = db.stats()
    print(
        "Stats — queries_total:",
        stats.queries_total,
        "avg_query_latency:",
        round(stats.avg_query_latency, 6),
        "queries_per_sec:",
        round(stats.queries_per_sec, 6),
    )

    prom = db.prometheus_metrics()
    print("\nPrometheus /metrics (first lines):")
    for line in prom.strip().split("\n")[:12]:
        print(" ", line)

    db.delete_collection(coll_name)
    db.close()
    print("\nDone.")


if __name__ == "__main__":
    main()
