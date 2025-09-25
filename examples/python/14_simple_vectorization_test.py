#!/usr/bin/env python3
"""
Simple VittoriaDB Vectorization Test (No External Dependencies)

This script tests vectorization using only Python standard library and curl.
It compares manual vector generation with the backend approach.

Requirements:
    - Python 3 (standard library only)
    - curl (for HTTP API calls)
    - VittoriaDB server running on localhost:8080
"""

import json
import subprocess
import time
import math
from typing import List, Dict, Any, Optional

class SimpleVittoriaDBTester:
    """Simple tester using curl for HTTP API calls"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.collections_created = []
        self._check_server_health()
    
    def _check_server_health(self):
        """Check if VittoriaDB server is running"""
        try:
            result = subprocess.run([
                'curl', '-s', f'{self.base_url}/health'
            ], capture_output=True, text=True, timeout=5)
            
            if result.returncode == 0:
                health = json.loads(result.stdout)
                print(f"✅ VittoriaDB server healthy - {health['collections']} collections, {health['total_vectors']} vectors")
            else:
                raise Exception(f"curl failed with code {result.returncode}")
        except Exception as e:
            print(f"❌ VittoriaDB server not available: {e}")
            print("   Please start server with: ./vittoriadb run --port 8080")
            exit(1)
    
    def create_collection(self, name: str, dimensions: int, metric: int = 0) -> bool:
        """Create a collection via HTTP API using curl"""
        try:
            # API uses integer values: metric: 0=cosine, 1=euclidean, 2=dot_product, 3=manhattan
            # index_type: 0=flat, 1=hnsw, 2=ivf
            payload = {
                "name": name,
                "dimensions": dimensions,
                "metric": metric,
                "index_type": 0
            }
            
            result = subprocess.run([
                'curl', '-s', '-X', 'POST',
                '-H', 'Content-Type: application/json',
                '-d', json.dumps(payload),
                f'{self.base_url}/collections'
            ], capture_output=True, text=True, timeout=10)
            
            if result.returncode == 0:
                # Check if response indicates success (no error message)
                if 'error' not in result.stdout.lower():
                    self.collections_created.append(name)
                    print(f"✅ Created collection '{name}' ({dimensions}D)")
                    return True
                else:
                    print(f"❌ Failed to create collection: {result.stdout}")
                    return False
            else:
                print(f"❌ curl failed: {result.stderr}")
                return False
        except Exception as e:
            print(f"❌ Error creating collection: {e}")
            return False
    
    def insert_vectors(self, collection_name: str, vectors: List[Dict[str, Any]]) -> bool:
        """Insert vectors into collection using curl"""
        try:
            payload = {"vectors": vectors}
            
            result = subprocess.run([
                'curl', '-s', '-X', 'POST',
                '-H', 'Content-Type: application/json',
                '-d', json.dumps(payload),
                f'{self.base_url}/collections/{collection_name}/vectors/batch'
            ], capture_output=True, text=True, timeout=30)
            
            if result.returncode == 0:
                if 'error' not in result.stdout.lower():
                    print(f"✅ Inserted {len(vectors)} vectors into '{collection_name}'")
                    return True
                else:
                    print(f"❌ Failed to insert vectors: {result.stdout}")
                    return False
            else:
                print(f"❌ curl failed: {result.stderr}")
                return False
        except Exception as e:
            print(f"❌ Error inserting vectors: {e}")
            return False
    
    def search_vectors(self, collection_name: str, query_vector: List[float], limit: int = 5) -> Optional[List[Dict]]:
        """Search for similar vectors using curl"""
        try:
            payload = {
                "vector": query_vector,
                "limit": limit,
                "include_metadata": True
            }
            
            result = subprocess.run([
                'curl', '-s', '-X', 'POST',
                '-H', 'Content-Type: application/json',
                '-d', json.dumps(payload),
                f'{self.base_url}/collections/{collection_name}/search'
            ], capture_output=True, text=True, timeout=30)
            
            if result.returncode == 0:
                try:
                    response = json.loads(result.stdout)
                    return response.get("results", [])
                except json.JSONDecodeError:
                    print(f"❌ Invalid JSON response: {result.stdout}")
                    return None
            else:
                print(f"❌ curl failed: {result.stderr}")
                return None
        except Exception as e:
            print(f"❌ Error searching: {e}")
            return None
    
    def cleanup(self):
        """Clean up created collections"""
        print("\n🧹 Cleaning up collections...")
        for collection_name in self.collections_created:
            try:
                result = subprocess.run([
                    'curl', '-s', '-X', 'DELETE',
                    f'{self.base_url}/collections/{collection_name}'
                ], capture_output=True, text=True, timeout=10)
                
                if result.returncode == 0:
                    print(f"   ✅ Deleted '{collection_name}'")
                else:
                    print(f"   ⚠️  Failed to delete '{collection_name}': {result.stderr}")
            except Exception as e:
                print(f"   ❌ Error deleting '{collection_name}': {e}")

def generate_simple_vector(text: str, dimensions: int) -> List[float]:
    """
    Generate a simple semantic vector using basic hashing
    This is a simplified version for testing without external dependencies
    """
    vector = [0.0] * dimensions
    words = text.lower().split()
    
    if not words:
        return vector
    
    for i in range(dimensions):
        value = 0.0
        
        for j, word in enumerate(words):
            # Simple character-based features
            char_sum = sum(ord(c) for c in word)
            
            # Position and dimension-specific transformations
            hash_val = (char_sum * (i + 1) * (j + 1)) % 1000
            length_factor = len(word) * (i % 7 + 1)
            pos_factor = (j + 1) * (dimensions - i)
            
            value += hash_val + length_factor + pos_factor
        
        # Add dimension-specific bias
        bias = (i * i) % 17 - 8
        vector[i] = value + bias
    
    # Simple normalization
    norm = math.sqrt(sum(v * v for v in vector))
    if norm > 0:
        vector = [v / norm for v in vector]
    
    return vector

def cosine_similarity(a: List[float], b: List[float]) -> float:
    """Calculate cosine similarity between two vectors"""
    dot_product = sum(x * y for x, y in zip(a, b))
    norm_a = math.sqrt(sum(x * x for x in a))
    norm_b = math.sqrt(sum(x * x for x in b))
    
    if norm_a == 0 or norm_b == 0:
        return 0.0
    
    return dot_product / (norm_a * norm_b)

def test_basic_vectorization(tester: SimpleVittoriaDBTester):
    """Test basic vectorization and similarity"""
    print("\n" + "="*60)
    print("📊 TEST: Basic Vectorization and Similarity")
    print("="*60)
    
    collection_name = "simple_vector_test"
    dimensions = 128  # Smaller for faster processing
    
    if not tester.create_collection(collection_name, dimensions):
        return
    
    # Test documents with clear semantic relationships
    test_docs = [
        {"id": "db1", "text": "vector database storage and retrieval", "category": "database"},
        {"id": "db2", "text": "database indexing and search algorithms", "category": "database"},
        {"id": "ai1", "text": "artificial intelligence and machine learning", "category": "ai"},
        {"id": "ai2", "text": "neural networks and deep learning models", "category": "ai"},
        {"id": "cook1", "text": "cooking recipes and food preparation", "category": "cooking"},
        {"id": "cook2", "text": "kitchen techniques and culinary arts", "category": "cooking"},
    ]
    
    print(f"📝 Processing {len(test_docs)} test documents...")
    
    # Generate vectors
    vectors = []
    for doc in test_docs:
        vector = generate_simple_vector(doc["text"], dimensions)
        vectors.append({
            "id": doc["id"],
            "vector": vector,
            "metadata": {
                "text": doc["text"],
                "category": doc["category"]
            }
        })
    
    if not tester.insert_vectors(collection_name, vectors):
        return
    
    # Test queries
    test_queries = [
        {"text": "database search and indexing", "expected": "database"},
        {"text": "machine learning algorithms", "expected": "ai"},
        {"text": "food and cooking methods", "expected": "cooking"},
    ]
    
    print(f"\n🔍 Testing {len(test_queries)} queries...")
    
    total_correct = 0
    total_queries = 0
    
    for i, query in enumerate(test_queries, 1):
        print(f"\nQuery {i}: '{query['text']}'")
        print(f"Expected category: {query['expected']}")
        
        query_vector = generate_simple_vector(query["text"], dimensions)
        results = tester.search_vectors(collection_name, query_vector, limit=3)
        
        if results:
            print("📊 Results:")
            top_category = None
            
            for j, result in enumerate(results, 1):
                score = result.get("score", 0)
                category = result.get("metadata", {}).get("category", "unknown")
                text = result.get("metadata", {}).get("text", "")
                
                if j == 1:  # Top result
                    top_category = category
                
                match_indicator = "✅" if category == query["expected"] else "❌"
                print(f"   {j}. {match_indicator} Score: {score:.4f}, Category: {category}")
                print(f"      Text: {text}")
            
            if top_category == query["expected"]:
                total_correct += 1
                print(f"   🎯 TOP RESULT CORRECT!")
            else:
                print(f"   ❌ Top result incorrect (got {top_category}, expected {query['expected']})")
            
            total_queries += 1
        else:
            print("❌ No results returned")
            total_queries += 1
    
    # Calculate accuracy
    if total_queries > 0:
        accuracy = (total_correct / total_queries) * 100
        print(f"\n📊 Overall Accuracy: {total_correct}/{total_queries} ({accuracy:.1f}%)")
        
        if accuracy >= 80:
            print("   ✅ Good accuracy for simple vectorization")
        elif accuracy >= 50:
            print("   🔶 Moderate accuracy - could be improved")
        else:
            print("   ❌ Poor accuracy - vectorization needs improvement")
    
    # Test similarity analysis
    print(f"\n📈 Similarity Analysis:")
    
    # Compare within categories vs across categories
    db_vec1 = generate_simple_vector("vector database storage and retrieval", dimensions)
    db_vec2 = generate_simple_vector("database indexing and search algorithms", dimensions)
    ai_vec1 = generate_simple_vector("artificial intelligence and machine learning", dimensions)
    cook_vec1 = generate_simple_vector("cooking recipes and food preparation", dimensions)
    
    within_category = cosine_similarity(db_vec1, db_vec2)
    across_categories_1 = cosine_similarity(db_vec1, ai_vec1)
    across_categories_2 = cosine_similarity(db_vec1, cook_vec1)
    
    print(f"   Database docs similarity: {within_category:.4f}")
    print(f"   Database vs AI similarity: {across_categories_1:.4f}")
    print(f"   Database vs Cooking similarity: {across_categories_2:.4f}")
    
    avg_across = (across_categories_1 + across_categories_2) / 2
    discrimination = within_category - avg_across
    
    print(f"   Discrimination score: {discrimination:.4f}")
    
    if discrimination > 0.1:
        print("   ✅ Good semantic discrimination")
    elif discrimination > 0.05:
        print("   🔶 Moderate semantic discrimination")
    else:
        print("   ❌ Poor semantic discrimination")

def test_backend_comparison(tester: SimpleVittoriaDBTester):
    """Compare with backend-style approach"""
    print("\n" + "="*60)
    print("🔄 TEST: Backend Comparison Analysis")
    print("="*60)
    
    print("📋 Backend Implementation Analysis:")
    print("   • Uses OpenAI embeddings (1536D) for production")
    print("   • Falls back to sentence-transformers (384D)")
    print("   • Smart chunking with overlap (6000 tokens max)")
    print("   • Relevance filtering based on scores")
    print("   • Proper semantic search with context")
    
    print("\n🔧 Simple Test Implementation:")
    print("   • Uses basic hash-based vectors (128D)")
    print("   • No external dependencies")
    print("   • Simple similarity calculation")
    print("   • Good for development/testing")
    
    print("\n📊 Key Differences:")
    print("   Backend Approach:")
    print("     ✅ High-quality semantic embeddings")
    print("     ✅ Proper relevance scoring")
    print("     ✅ Production-ready accuracy")
    print("     ❌ Requires external APIs or models")
    
    print("   Simple Test Approach:")
    print("     ✅ No external dependencies")
    print("     ✅ Fast generation")
    print("     ✅ Good for basic testing")
    print("     ❌ Limited semantic understanding")
    
    print("\n🎯 Recommendations:")
    print("   • Use simple vectors for development/testing")
    print("   • Use backend approach (OpenAI/sentence-transformers) for production")
    print("   • Consider server-side vectorization for consistency")
    print("   • Implement relevance thresholds for proper matching")

def main():
    """Main test function"""
    print("🚀 Simple VittoriaDB Vectorization Test")
    print("=" * 60)
    print("Testing vectorization without external dependencies")
    print()
    
    # Initialize tester
    tester = SimpleVittoriaDBTester()
    
    try:
        # Run tests
        test_basic_vectorization(tester)
        test_backend_comparison(tester)
        
        print("\n" + "="*60)
        print("✅ SIMPLE VECTORIZATION TEST COMPLETED")
        print("="*60)
        
        print("\n🔗 Next Steps:")
        print("   1. Install sentence-transformers: pip install sentence-transformers")
        print("   2. Run advanced test: python 13_vectorization_comparison_test.py")
        print("   3. Compare with Go examples: go run examples/go/10_large_text_processing_demo.go")
        print("   4. Test backend integration with web-ui-rag")
        
    except KeyboardInterrupt:
        print("\n⏹️  Test interrupted by user")
    except Exception as e:
        print(f"\n❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
    finally:
        # Cleanup
        tester.cleanup()
        print("\n✅ Simple vectorization test completed!")

if __name__ == "__main__":
    main()
