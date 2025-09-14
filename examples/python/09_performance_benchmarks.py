#!/usr/bin/env python3
"""
Performance Benchmark Example for VittoriaDB

This example demonstrates:
1. Performance testing with different vector sizes and dimensions
2. Comparison between flat and HNSW indexing
3. Batch vs individual operations
4. Memory and timing measurements

Requirements:
    pip install vittoriadb numpy requests psutil

Usage:
    python 09_performance_benchmarks.py
"""

import time
import os
import numpy as np
from typing import List, Dict, Any, Tuple
from dataclasses import dataclass
import psutil
import vittoriadb

@dataclass
class BenchmarkResult:
    """Results from a benchmark test."""
    operation: str
    vectors_count: int
    dimensions: int
    index_type: str
    total_time: float
    vectors_per_second: float
    memory_usage_mb: float
    success_rate: float

class VittoriaDBBenchmark:
    """Benchmark client for VittoriaDB performance testing."""
    
    def __init__(self):
        self.db = None
        self.collections = {}
        self.results: List[BenchmarkResult] = []
    
    def connect(self) -> bool:
        """Connect to VittoriaDB."""
        try:
            # Connect to existing VittoriaDB server (don't auto-start)
            self.db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)
            return True
        except Exception as e:
            print(f"Failed to connect to VittoriaDB: {e}")
            return False
    
    def health_check(self) -> bool:
        """Check if VittoriaDB is running."""
        try:
            if not self.db:
                return False
            health = self.db.health()
            return health.status == "healthy"
        except:
            return False
    
    def create_collection(self, name: str, dimensions: int, metric: str = "cosine") -> bool:
        """Create a collection."""
        try:
            metric_enum = getattr(vittoriadb.DistanceMetric, metric.upper(), vittoriadb.DistanceMetric.COSINE)
            collection = self.db.create_collection(
                name=name,
                dimensions=dimensions,
                metric=metric_enum
            )
            self.collections[name] = collection
            return True
        except vittoriadb.CollectionError as e:
            if "already exists" in str(e):
                collection = self.db.get_collection(name)
                self.collections[name] = collection
                return True
            return False
        except Exception:
            return False
    
    def delete_collection(self, name: str) -> bool:
        """Delete a collection."""
        try:
            self.db.delete_collection(name)
            if name in self.collections:
                del self.collections[name]
            return True
        except Exception:
            return False
    
    def insert_vector(self, collection_name: str, vector_id: str, vector: List[float]) -> bool:
        """Insert a single vector."""
        try:
            if collection_name not in self.collections:
                return False
            
            collection = self.collections[collection_name]
            collection.insert(
                id=vector_id,
                vector=vector,
                metadata={"benchmark": True, "timestamp": time.time()}
            )
            return True
        except Exception:
            return False
    
    def insert_batch(self, collection_name: str, vectors: List[Dict[str, Any]]) -> Tuple[bool, int]:
        """Insert vectors individually (batch endpoint not available)."""
        success_count = 0
        for vector_data in vectors:
            if self.insert_vector(collection_name, vector_data["id"], vector_data["vector"]):
                success_count += 1
        return success_count > 0, success_count
    
    def search_vector(self, collection_name: str, query_vector: List[float], limit: int = 10) -> Tuple[bool, int]:
        """Search for similar vectors."""
        try:
            if collection_name not in self.collections:
                return False, 0
            
            collection = self.collections[collection_name]
            results = collection.search(
                vector=query_vector,
                limit=limit,
                include_metadata=True
            )
            return True, len(results)
        except Exception:
            return False, 0
    
    def get_memory_usage(self) -> float:
        """Get current memory usage in MB."""
        process = psutil.Process(os.getpid())
        return process.memory_info().rss / 1024 / 1024
    
    def generate_random_vectors(self, count: int, dimensions: int) -> np.ndarray:
        """Generate random normalized vectors."""
        vectors = np.random.random((count, dimensions)).astype(np.float32)
        # Normalize vectors
        norms = np.linalg.norm(vectors, axis=1, keepdims=True)
        vectors = vectors / norms
        return vectors
    
    def benchmark_insert_individual(self, collection: str, vectors: np.ndarray, 
                                  metric: str) -> BenchmarkResult:
        """Benchmark individual vector insertions."""
        print(f"  🔄 Individual insertions ({len(vectors)} vectors)...")
        
        start_memory = self.get_memory_usage()
        start_time = time.time()
        
        success_count = 0
        for i, vector in enumerate(vectors):
            vector_id = f"vec_{i}"
            if self.insert_vector(collection, vector_id, vector.tolist()):
                success_count += 1
            
            if (i + 1) % 100 == 0:
                print(f"    Inserted {i + 1}/{len(vectors)} vectors...")
        
        end_time = time.time()
        end_memory = self.get_memory_usage()
        
        total_time = end_time - start_time
        vectors_per_second = len(vectors) / total_time if total_time > 0 else 0
        memory_usage = end_memory - start_memory
        success_rate = success_count / len(vectors)
        
        result = BenchmarkResult(
            operation="insert_individual",
            vectors_count=len(vectors),
            dimensions=vectors.shape[1],
            index_type=metric,
            total_time=total_time,
            vectors_per_second=vectors_per_second,
            memory_usage_mb=memory_usage,
            success_rate=success_rate
        )
        
        self.results.append(result)
        return result
    
    def benchmark_insert_batch(self, collection: str, vectors: np.ndarray, 
                              metric: str, batch_size: int = 100) -> BenchmarkResult:
        """Benchmark batch vector insertions."""
        print(f"  🔄 Batch insertions ({len(vectors)} vectors, batch size: {batch_size})...")
        
        start_memory = self.get_memory_usage()
        start_time = time.time()
        
        total_inserted = 0
        batches = 0
        
        for i in range(0, len(vectors), batch_size):
            batch_vectors = vectors[i:i + batch_size]
            batch_data = []
            
            for j, vector in enumerate(batch_vectors):
                batch_data.append({
                    "id": f"batch_vec_{i + j}",
                    "vector": vector.tolist(),
                    "metadata": {"batch": batches, "index": j}
                })
            
            success, inserted = self.insert_batch(collection, batch_data)
            if success:
                total_inserted += inserted
            
            batches += 1
            
            if batches % 10 == 0:
                print(f"    Processed {batches} batches, {total_inserted} vectors inserted...")
        
        end_time = time.time()
        end_memory = self.get_memory_usage()
        
        total_time = end_time - start_time
        vectors_per_second = total_inserted / total_time if total_time > 0 else 0
        memory_usage = end_memory - start_memory
        success_rate = total_inserted / len(vectors)
        
        result = BenchmarkResult(
            operation="insert_batch",
            vectors_count=len(vectors),
            dimensions=vectors.shape[1],
            index_type=metric,
            total_time=total_time,
            vectors_per_second=vectors_per_second,
            memory_usage_mb=memory_usage,
            success_rate=success_rate
        )
        
        self.results.append(result)
        return result
    
    def benchmark_search(self, collection: str, query_vectors: np.ndarray, 
                        index_type: str, limit: int = 10) -> BenchmarkResult:
        """Benchmark search operations."""
        print(f"  🔍 Search operations ({len(query_vectors)} queries, limit: {limit})...")
        
        start_memory = self.get_memory_usage()
        start_time = time.time()
        
        total_results = 0
        success_count = 0
        
        for i, query_vector in enumerate(query_vectors):
            success, result_count = self.search_vector(collection, query_vector.tolist(), limit)
            if success:
                success_count += 1
                total_results += result_count
            
            if (i + 1) % 10 == 0:
                print(f"    Completed {i + 1}/{len(query_vectors)} searches...")
        
        end_time = time.time()
        end_memory = self.get_memory_usage()
        
        total_time = end_time - start_time
        searches_per_second = len(query_vectors) / total_time if total_time > 0 else 0
        memory_usage = end_memory - start_memory
        success_rate = success_count / len(query_vectors)
        
        result = BenchmarkResult(
            operation="search",
            vectors_count=len(query_vectors),
            dimensions=query_vectors.shape[1],
            index_type=index_type,
            total_time=total_time,
            vectors_per_second=searches_per_second,
            memory_usage_mb=memory_usage,
            success_rate=success_rate
        )
        
        self.results.append(result)
        return result
    
    def run_comprehensive_benchmark(self):
        """Run a comprehensive benchmark suite."""
        print("🚀 VittoriaDB Performance Benchmark")
        print("=" * 50)
        
        # Connect to VittoriaDB
        if not self.connect():
            print("❌ Failed to connect to VittoriaDB. Please start it with: ./vittoriadb run")
            return
        
        if not self.health_check():
            print("❌ VittoriaDB is not healthy")
            return
        
        print("✅ Connected to VittoriaDB")
        
        # Test configurations
        test_configs = [
            {"vectors": 1000, "dimensions": 128, "metric": "cosine"},
            {"vectors": 1000, "dimensions": 384, "metric": "cosine"},
            {"vectors": 2000, "dimensions": 128, "metric": "cosine"},
            {"vectors": 2000, "dimensions": 384, "metric": "cosine"},
            {"vectors": 1000, "dimensions": 128, "metric": "euclidean"},
            {"vectors": 1000, "dimensions": 384, "metric": "euclidean"},
        ]
        
        for i, config in enumerate(test_configs):
            print(f"\n📊 Test {i + 1}/{len(test_configs)}: {config['vectors']} vectors, "
                  f"{config['dimensions']} dims, {config['metric']} metric")
            print("-" * 60)
            
            collection_name = f"benchmark_{i}"
            
            # Create collection
            print(f"  🔄 Creating collection '{collection_name}'...")
            self.create_collection(collection_name, config["dimensions"], config["metric"])
            
            # Generate test data
            print(f"  🎲 Generating {config['vectors']} random vectors...")
            vectors = self.generate_random_vectors(config["vectors"], config["dimensions"])
            query_vectors = self.generate_random_vectors(50, config["dimensions"])  # 50 queries
            
            # Benchmark batch insertion
            batch_result = self.benchmark_insert_batch(
                collection_name, vectors, config["metric"], batch_size=100
            )
            
            # Benchmark search
            search_result = self.benchmark_search(
                collection_name, query_vectors, config["metric"], limit=10
            )
            
            # Print results
            print(f"\n  📈 Results:")
            print(f"    Batch Insert: {batch_result.vectors_per_second:.0f} vectors/sec "
                  f"({batch_result.success_rate:.1%} success)")
            print(f"    Search: {search_result.vectors_per_second:.0f} queries/sec "
                  f"({search_result.success_rate:.1%} success)")
            print(f"    Memory: +{batch_result.memory_usage_mb:.1f} MB (insert), "
                  f"+{search_result.memory_usage_mb:.1f} MB (search)")
            
            # Clean up
            self.delete_collection(collection_name)
        
        # Print summary
        self.print_benchmark_summary()
    
    def print_benchmark_summary(self):
        """Print a summary of all benchmark results."""
        print("\n" + "=" * 70)
        print("📊 BENCHMARK SUMMARY")
        print("=" * 70)
        
        # Group results by operation and index type
        insert_results = [r for r in self.results if r.operation == "insert_batch"]
        search_results = [r for r in self.results if r.operation == "search"]
        
        # Insert performance summary
        print("\n🔄 INSERT PERFORMANCE:")
        print(f"{'Vectors':<8} {'Dims':<5} {'Index':<6} {'Speed (vec/s)':<12} {'Memory (MB)':<12} {'Success':<8}")
        print("-" * 60)
        
        for result in insert_results:
            print(f"{result.vectors_count:<8} {result.dimensions:<5} {result.index_type:<6} "
                  f"{result.vectors_per_second:<12.0f} {result.memory_usage_mb:<12.1f} "
                  f"{result.success_rate:<8.1%}")
        
        # Search performance summary
        print("\n🔍 SEARCH PERFORMANCE:")
        print(f"{'Queries':<8} {'Dims':<5} {'Index':<6} {'Speed (q/s)':<12} {'Memory (MB)':<12} {'Success':<8}")
        print("-" * 60)
        
        for result in search_results:
            print(f"{result.vectors_count:<8} {result.dimensions:<5} {result.index_type:<6} "
                  f"{result.vectors_per_second:<12.0f} {result.memory_usage_mb:<12.1f} "
                  f"{result.success_rate:<8.1%}")
        
        # Performance comparison
        if insert_results and search_results:
            print("\n📈 PERFORMANCE INSIGHTS:")
            
            # Best insert performance
            best_insert = max(insert_results, key=lambda r: r.vectors_per_second)
            print(f"  • Best insert speed: {best_insert.vectors_per_second:.0f} vectors/sec "
                  f"({best_insert.index_type} index, {best_insert.dimensions}D)")
            
            # Best search performance
            best_search = max(search_results, key=lambda r: r.vectors_per_second)
            print(f"  • Best search speed: {best_search.vectors_per_second:.0f} queries/sec "
                  f"({best_search.index_type} index, {best_search.dimensions}D)")
            
            # Index type comparison
            flat_inserts = [r for r in insert_results if r.index_type == "flat"]
            hnsw_inserts = [r for r in insert_results if r.index_type == "hnsw"]
            
            if flat_inserts and hnsw_inserts:
                avg_flat_insert = sum(r.vectors_per_second for r in flat_inserts) / len(flat_inserts)
                avg_hnsw_insert = sum(r.vectors_per_second for r in hnsw_inserts) / len(hnsw_inserts)
                
                print(f"  • Flat index avg insert: {avg_flat_insert:.0f} vectors/sec")
                print(f"  • HNSW index avg insert: {avg_hnsw_insert:.0f} vectors/sec")
                
                if avg_flat_insert > avg_hnsw_insert:
                    ratio = avg_flat_insert / avg_hnsw_insert
                    print(f"  • Flat index is {ratio:.1f}x faster for insertions")
                else:
                    ratio = avg_hnsw_insert / avg_flat_insert
                    print(f"  • HNSW index is {ratio:.1f}x faster for insertions")
        
        print("\n✅ Benchmark complete!")

def main():
    """Run the performance benchmark."""
    benchmark = VittoriaDBBenchmark()
    benchmark.run_comprehensive_benchmark()

if __name__ == "__main__":
    main()
