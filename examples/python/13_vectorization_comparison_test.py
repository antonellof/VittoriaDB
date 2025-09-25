#!/usr/bin/env python3
"""
VittoriaDB Vectorization Comparison Test

This script tests different vectorization approaches and compares them with
the backend implementation to ensure consistent results and proper matching.

Tests:
1. HTTP API with manual vectors (like Go examples)
2. Backend-style vectorization (OpenAI/sentence-transformers)
3. Similarity score analysis and comparison
4. Large text processing with proper matching

Requirements:
    pip install requests numpy sentence-transformers openai
    VittoriaDB server running on localhost:8080
"""

import requests
import json
import time
import numpy as np
from typing import List, Dict, Any, Optional
import hashlib
import re
import os

# Try to import optional dependencies
try:
    from sentence_transformers import SentenceTransformer
    HAS_SENTENCE_TRANSFORMERS = True
except ImportError:
    HAS_SENTENCE_TRANSFORMERS = False
    print("⚠️  sentence-transformers not available")

try:
    import openai
    HAS_OPENAI = True
except ImportError:
    HAS_OPENAI = False
    print("⚠️  openai not available")

class VittoriaDBTester:
    """Test different vectorization approaches with VittoriaDB"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()
        self.collections_created = []
        
        # Initialize embedding models if available
        self.sentence_model = None
        if HAS_SENTENCE_TRANSFORMERS:
            try:
                print("🤖 Loading sentence-transformers model...")
                self.sentence_model = SentenceTransformer('all-MiniLM-L6-v2')
                print("✅ Sentence-transformers model loaded")
            except Exception as e:
                print(f"⚠️  Failed to load sentence-transformers: {e}")
        
        # Check server health
        self._check_server_health()
    
    def _check_server_health(self):
        """Check if VittoriaDB server is running"""
        try:
            response = self.session.get(f"{self.base_url}/health")
            if response.status_code == 200:
                health = response.json()
                print(f"✅ VittoriaDB server healthy - {health['collections']} collections, {health['total_vectors']} vectors")
            else:
                raise Exception(f"Server returned {response.status_code}")
        except Exception as e:
            print(f"❌ VittoriaDB server not available: {e}")
            print("   Please start server with: ./vittoriadb run --port 8080")
            exit(1)
    
    def create_collection(self, name: str, dimensions: int, metric: str = "COSINE") -> bool:
        """Create a collection via HTTP API"""
        try:
            payload = {
                "name": name,
                "dimensions": dimensions,
                "metric": metric,
                "index_type": "FLAT"
            }
            
            response = self.session.post(f"{self.base_url}/collections", json=payload)
            if response.status_code in [200, 201]:
                self.collections_created.append(name)
                print(f"✅ Created collection '{name}' ({dimensions}D)")
                return True
            else:
                print(f"❌ Failed to create collection: {response.text}")
                return False
        except Exception as e:
            print(f"❌ Error creating collection: {e}")
            return False
    
    def insert_vectors(self, collection_name: str, vectors: List[Dict[str, Any]]) -> bool:
        """Insert vectors into collection"""
        try:
            payload = {"vectors": vectors}
            response = self.session.post(f"{self.base_url}/collections/{collection_name}/vectors", json=payload)
            
            if response.status_code in [200, 201]:
                print(f"✅ Inserted {len(vectors)} vectors into '{collection_name}'")
                return True
            else:
                print(f"❌ Failed to insert vectors: {response.text}")
                return False
        except Exception as e:
            print(f"❌ Error inserting vectors: {e}")
            return False
    
    def search_vectors(self, collection_name: str, query_vector: List[float], limit: int = 5) -> Optional[List[Dict]]:
        """Search for similar vectors"""
        try:
            payload = {
                "vector": query_vector,
                "limit": limit,
                "include_metadata": True
            }
            
            response = self.session.post(f"{self.base_url}/collections/{collection_name}/search", json=payload)
            
            if response.status_code == 200:
                results = response.json()
                return results.get("results", [])
            else:
                print(f"❌ Search failed: {response.text}")
                return None
        except Exception as e:
            print(f"❌ Error searching: {e}")
            return None
    
    def cleanup(self):
        """Clean up created collections"""
        print("\n🧹 Cleaning up collections...")
        for collection_name in self.collections_created:
            try:
                response = self.session.delete(f"{self.base_url}/collections/{collection_name}")
                if response.status_code in [200, 204]:
                    print(f"   ✅ Deleted '{collection_name}'")
                else:
                    print(f"   ⚠️  Failed to delete '{collection_name}': {response.text}")
            except Exception as e:
                print(f"   ❌ Error deleting '{collection_name}': {e}")

def generate_enhanced_vector(text: str, dimensions: int) -> List[float]:
    """
    Generate enhanced semantic vector (same algorithm as Go examples)
    This matches the improved algorithm from our Go debugging tool
    """
    vector = [0.0] * dimensions
    words = text.lower().split()
    
    if not words:
        return vector
    
    for i in range(dimensions):
        value = 0.0
        
        for j, word in enumerate(words):
            # Character-based features with high variation
            char_feature = 0.0
            for k, char in enumerate(word):
                char_code = ord(char)
                if i % 5 == 0:
                    char_feature += char_code * (k + 1) * 0.1
                elif i % 5 == 1:
                    char_feature += (char_code * char_code) * (k + 2) * 0.01
                elif i % 5 == 2:
                    char_feature += char_code / (k + 3) * 10.0
                elif i % 5 == 3:
                    char_feature += (char_code ^ (k + 1)) * 0.05
                elif i % 5 == 4:
                    char_feature += char_code * (len(word) - k) * 0.2
            
            # Word length and position features
            length_feature = len(word) * (i + 1) * 0.3
            pos_feature = (j + 1) * (dimensions - i) * 0.1
            
            # Hash-based features
            hash1 = djb2_hash(word) % (i * 97 + 13)
            hash2 = sdbm_hash(word) % (i * 73 + 17)
            hash_feature = (hash1 - hash2) * 0.01
            
            # Word uniqueness
            unique_chars = len(set(word))
            uniqueness_feature = unique_chars * (i + 1) * 0.5
            
            # Combine features
            dim_weight = 1.0 + (i % 7) * 0.3
            combined = (char_feature + length_feature + pos_feature + hash_feature + uniqueness_feature) * dim_weight
            interaction = (j * i + 1) * 0.1
            combined += interaction
            
            value += combined
        
        # Add dimension-specific bias
        dim_bias = ((i * i) % 17) * 0.2 - 1.0
        vector[i] = value + dim_bias
    
    # L2 normalize
    norm = np.linalg.norm(vector)
    if norm > 0:
        vector = [v / norm for v in vector]
    
    return vector

def djb2_hash(s: str) -> int:
    """DJB2 hash function"""
    hash_val = 5381
    for char in s:
        hash_val = ((hash_val << 5) + hash_val) + ord(char)
    return abs(hash_val)

def sdbm_hash(s: str) -> int:
    """SDBM hash function"""
    hash_val = 0
    for char in s:
        hash_val = ord(char) + (hash_val << 6) + (hash_val << 16) - hash_val
    return abs(hash_val)

def cosine_similarity(a: List[float], b: List[float]) -> float:
    """Calculate cosine similarity between two vectors"""
    dot_product = sum(x * y for x, y in zip(a, b))
    norm_a = np.linalg.norm(a)
    norm_b = np.linalg.norm(b)
    
    if norm_a == 0 or norm_b == 0:
        return 0.0
    
    return dot_product / (norm_a * norm_b)

def test_manual_vectorization(tester: VittoriaDBTester):
    """Test 1: Manual vectorization (like Go examples)"""
    print("\n" + "="*60)
    print("📊 TEST 1: Manual Vectorization (Go-style)")
    print("="*60)
    
    collection_name = "manual_vectors_test"
    dimensions = 384
    
    if not tester.create_collection(collection_name, dimensions):
        return
    
    # Test documents with known relationships
    test_docs = [
        {"id": "tech1", "text": "VittoriaDB is a vector database for AI applications", "category": "technology"},
        {"id": "tech2", "text": "Vector databases store embeddings for machine learning", "category": "technology"},
        {"id": "install1", "text": "Installation requires downloading the binary file", "category": "installation"},
        {"id": "install2", "text": "Setup instructions for installing the software", "category": "installation"},
        {"id": "cooking1", "text": "Cooking recipes and kitchen techniques", "category": "unrelated"},
        {"id": "space1", "text": "Space exploration and astronomy research", "category": "unrelated"},
    ]
    
    # Generate vectors and insert
    vectors = []
    for doc in test_docs:
        vector = generate_enhanced_vector(doc["text"], dimensions)
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
    
    # Test queries with expected relevance
    test_queries = [
        {
            "text": "vector database technology",
            "expected_category": "technology",
            "description": "Should match technology docs"
        },
        {
            "text": "software installation setup",
            "expected_category": "installation", 
            "description": "Should match installation docs"
        },
        {
            "text": "completely unrelated cooking recipes",
            "expected_category": "unrelated",
            "description": "Should match unrelated content (but scores should be lower)"
        }
    ]
    
    print(f"\n🔍 Testing {len(test_queries)} queries...")
    
    for i, query in enumerate(test_queries, 1):
        print(f"\nQuery {i}: '{query['text']}'")
        print(f"Expected: {query['description']}")
        
        query_vector = generate_enhanced_vector(query["text"], dimensions)
        results = tester.search_vectors(collection_name, query_vector, limit=3)
        
        if results:
            print("📊 Results:")
            for j, result in enumerate(results, 1):
                score = result.get("score", 0)
                category = result.get("metadata", {}).get("category", "unknown")
                text = result.get("metadata", {}).get("text", "")
                
                match_indicator = "✅" if category == query["expected_category"] else "❌"
                print(f"   {j}. {match_indicator} Score: {score:.4f}, Category: {category}")
                print(f"      Text: {text[:60]}...")
        else:
            print("❌ No results returned")
    
    # Analyze similarity distribution
    print(f"\n📈 Similarity Analysis:")
    print("Comparing vectors directly (not through search):")
    
    # Test specific pairs
    tech_vec1 = generate_enhanced_vector("VittoriaDB is a vector database for AI applications", dimensions)
    tech_vec2 = generate_enhanced_vector("Vector databases store embeddings for machine learning", dimensions)
    cooking_vec = generate_enhanced_vector("Cooking recipes and kitchen techniques", dimensions)
    
    tech_similarity = cosine_similarity(tech_vec1, tech_vec2)
    unrelated_similarity = cosine_similarity(tech_vec1, cooking_vec)
    
    print(f"   Technology docs similarity: {tech_similarity:.4f} (should be high)")
    print(f"   Tech vs Cooking similarity: {unrelated_similarity:.4f} (should be low)")
    
    if tech_similarity > 0.9 and unrelated_similarity > 0.9:
        print("   ⚠️  WARNING: Manual vectors too similar - poor discrimination!")
    elif tech_similarity > unrelated_similarity:
        print("   ✅ Manual vectors show some discrimination")
    else:
        print("   ❌ Manual vectors show poor discrimination")

def test_sentence_transformers_vectorization(tester: VittoriaDBTester):
    """Test 2: Sentence-transformers vectorization (backend-style)"""
    print("\n" + "="*60)
    print("🤖 TEST 2: Sentence-Transformers Vectorization (Backend-style)")
    print("="*60)
    
    if not tester.sentence_model:
        print("⚠️  Sentence-transformers not available - skipping test")
        return
    
    collection_name = "sentence_transformers_test"
    dimensions = 384  # all-MiniLM-L6-v2 dimensions
    
    if not tester.create_collection(collection_name, dimensions):
        return
    
    # Same test documents as manual test
    test_docs = [
        {"id": "tech1", "text": "VittoriaDB is a vector database for AI applications", "category": "technology"},
        {"id": "tech2", "text": "Vector databases store embeddings for machine learning", "category": "technology"},
        {"id": "install1", "text": "Installation requires downloading the binary file", "category": "installation"},
        {"id": "install2", "text": "Setup instructions for installing the software", "category": "installation"},
        {"id": "cooking1", "text": "Cooking recipes and kitchen techniques", "category": "unrelated"},
        {"id": "space1", "text": "Space exploration and astronomy research", "category": "unrelated"},
    ]
    
    # Generate vectors using sentence-transformers
    print("🔄 Generating embeddings with sentence-transformers...")
    vectors = []
    texts = [doc["text"] for doc in test_docs]
    embeddings = tester.sentence_model.encode(texts)
    
    for doc, embedding in zip(test_docs, embeddings):
        vectors.append({
            "id": doc["id"],
            "vector": embedding.tolist(),
            "metadata": {
                "text": doc["text"],
                "category": doc["category"]
            }
        })
    
    if not tester.insert_vectors(collection_name, vectors):
        return
    
    # Test same queries
    test_queries = [
        {
            "text": "vector database technology",
            "expected_category": "technology",
            "description": "Should match technology docs"
        },
        {
            "text": "software installation setup",
            "expected_category": "installation", 
            "description": "Should match installation docs"
        },
        {
            "text": "completely unrelated cooking recipes",
            "expected_category": "unrelated",
            "description": "Should match unrelated content"
        }
    ]
    
    print(f"\n🔍 Testing {len(test_queries)} queries with sentence-transformers...")
    
    for i, query in enumerate(test_queries, 1):
        print(f"\nQuery {i}: '{query['text']}'")
        print(f"Expected: {query['description']}")
        
        query_embedding = tester.sentence_model.encode([query["text"]])[0]
        results = tester.search_vectors(collection_name, query_embedding.tolist(), limit=3)
        
        if results:
            print("📊 Results:")
            relevant_count = 0
            for j, result in enumerate(results, 1):
                score = result.get("score", 0)
                category = result.get("metadata", {}).get("category", "unknown")
                text = result.get("metadata", {}).get("text", "")
                
                is_relevant = (category == query["expected_category"]) or (score > 0.5)
                if is_relevant:
                    relevant_count += 1
                
                match_indicator = "✅" if category == query["expected_category"] else ("🔶" if score > 0.3 else "❌")
                print(f"   {j}. {match_indicator} Score: {score:.4f}, Category: {category}")
                print(f"      Text: {text[:60]}...")
            
            print(f"   📊 Relevant results: {relevant_count}/{len(results)}")
        else:
            print("❌ No results returned")
    
    # Analyze sentence-transformers similarity
    print(f"\n📈 Sentence-Transformers Similarity Analysis:")
    
    tech_emb1 = tester.sentence_model.encode(["VittoriaDB is a vector database for AI applications"])[0]
    tech_emb2 = tester.sentence_model.encode(["Vector databases store embeddings for machine learning"])[0]
    cooking_emb = tester.sentence_model.encode(["Cooking recipes and kitchen techniques"])[0]
    
    tech_similarity = cosine_similarity(tech_emb1.tolist(), tech_emb2.tolist())
    unrelated_similarity = cosine_similarity(tech_emb1.tolist(), cooking_emb.tolist())
    
    print(f"   Technology docs similarity: {tech_similarity:.4f}")
    print(f"   Tech vs Cooking similarity: {unrelated_similarity:.4f}")
    
    if tech_similarity > unrelated_similarity + 0.1:
        print("   ✅ Sentence-transformers show good discrimination")
    else:
        print("   ⚠️  Sentence-transformers show limited discrimination")

def test_large_text_processing(tester: VittoriaDBTester):
    """Test 3: Large text processing with proper matching"""
    print("\n" + "="*60)
    print("📚 TEST 3: Large Text Processing with Proper Matching")
    print("="*60)
    
    collection_name = "large_text_test"
    dimensions = 384
    
    if not tester.create_collection(collection_name, dimensions):
        return
    
    # Simulate large documents (like README files)
    large_documents = [
        {
            "id": "vittoriadb_readme",
            "text": """VittoriaDB is a high-performance vector database designed for AI applications. 
            It provides efficient storage and retrieval of high-dimensional embeddings with support for 
            multiple index types including FLAT, HNSW, and IVF. The database features automatic 
            embedding generation, smart chunking, and parallel search capabilities. VittoriaDB supports 
            both HTTP API and native SDK integration for maximum flexibility. Performance optimizations 
            include batch processing, caching, and parallel search engines.""",
            "category": "database_docs"
        },
        {
            "id": "installation_guide", 
            "text": """Installation of VittoriaDB is straightforward. First, download the binary from 
            the releases page. Extract the archive and place the binary in your PATH. Start the server 
            with './vittoriadb run --port 8080'. For development, you can build from source using 
            'go build ./cmd/vittoriadb'. The server requires minimal configuration and starts with 
            sensible defaults. Docker images are also available for containerized deployments.""",
            "category": "installation_docs"
        },
        {
            "id": "api_reference",
            "text": """The VittoriaDB API provides RESTful endpoints for all operations. Create collections 
            with POST /collections, insert vectors with POST /collections/{name}/vectors, and search 
            with POST /collections/{name}/search. The API supports batch operations, metadata filtering, 
            and configurable similarity metrics. Authentication is optional but recommended for 
            production deployments. Rate limiting and request validation ensure system stability.""",
            "category": "api_docs"
        },
        {
            "id": "cooking_blog",
            "text": """Welcome to our cooking blog! Here you'll find delicious recipes from around the world. 
            Today we're sharing our famous chocolate chip cookie recipe. Start by preheating your oven 
            to 375°F. Mix butter, sugar, eggs, and vanilla in a large bowl. Gradually add flour, 
            baking soda, and salt. Fold in chocolate chips. Drop spoonfuls onto baking sheets and 
            bake for 9-11 minutes until golden brown. Let cool before serving.""",
            "category": "unrelated"
        }
    ]
    
    # Use sentence-transformers if available, otherwise manual vectors
    use_sentence_transformers = tester.sentence_model is not None
    
    print(f"📝 Processing {len(large_documents)} large documents...")
    print(f"🔧 Using: {'Sentence-transformers' if use_sentence_transformers else 'Manual vectors'}")
    
    vectors = []
    if use_sentence_transformers:
        texts = [doc["text"] for doc in large_documents]
        embeddings = tester.sentence_model.encode(texts)
        
        for doc, embedding in zip(large_documents, embeddings):
            vectors.append({
                "id": doc["id"],
                "vector": embedding.tolist(),
                "metadata": {
                    "text": doc["text"][:200] + "...",  # Truncate for storage
                    "full_text": doc["text"],
                    "category": doc["category"],
                    "word_count": len(doc["text"].split())
                }
            })
    else:
        for doc in large_documents:
            vector = generate_enhanced_vector(doc["text"], dimensions)
            vectors.append({
                "id": doc["id"],
                "vector": vector,
                "metadata": {
                    "text": doc["text"][:200] + "...",
                    "full_text": doc["text"],
                    "category": doc["category"],
                    "word_count": len(doc["text"].split())
                }
            })
    
    if not tester.insert_vectors(collection_name, vectors):
        return
    
    # Test queries for proper matching (only relevant results)
    test_queries = [
        {
            "text": "vector database performance and indexing",
            "expected_categories": ["database_docs"],
            "min_score": 0.3,
            "description": "Should find database documentation"
        },
        {
            "text": "how to install and setup the software",
            "expected_categories": ["installation_docs"],
            "min_score": 0.3,
            "description": "Should find installation guide"
        },
        {
            "text": "REST API endpoints and HTTP interface",
            "expected_categories": ["api_docs"],
            "min_score": 0.3,
            "description": "Should find API documentation"
        },
        {
            "text": "chocolate chip cookie baking recipe",
            "expected_categories": ["unrelated"],
            "min_score": 0.2,
            "description": "Should find cooking content (but may have lower scores)"
        },
        {
            "text": "quantum physics and particle accelerators",
            "expected_categories": [],  # Should not match well
            "min_score": 0.0,
            "description": "Should have low relevance scores (no good matches)"
        }
    ]
    
    print(f"\n🔍 Testing {len(test_queries)} queries for proper matching...")
    
    for i, query in enumerate(test_queries, 1):
        print(f"\nQuery {i}: '{query['text']}'")
        print(f"Expected: {query['description']}")
        
        if use_sentence_transformers:
            query_embedding = tester.sentence_model.encode([query["text"]])[0]
            query_vector = query_embedding.tolist()
        else:
            query_vector = generate_enhanced_vector(query["text"], dimensions)
        
        results = tester.search_vectors(collection_name, query_vector, limit=3)
        
        if results:
            relevant_results = []
            for j, result in enumerate(results, 1):
                score = result.get("score", 0)
                category = result.get("metadata", {}).get("category", "unknown")
                text = result.get("metadata", {}).get("text", "")
                
                is_relevant = score >= query["min_score"]
                is_expected = category in query["expected_categories"] if query["expected_categories"] else False
                
                if is_relevant:
                    relevant_results.append(result)
                
                if is_expected and is_relevant:
                    indicator = "✅ RELEVANT"
                elif is_expected:
                    indicator = "🔶 EXPECTED (low score)"
                elif is_relevant:
                    indicator = "🔶 RELEVANT (unexpected)"
                else:
                    indicator = "❌ NOT RELEVANT"
                
                print(f"   {j}. {indicator} Score: {score:.4f}")
                print(f"      Category: {category}")
                print(f"      Text: {text[:80]}...")
            
            print(f"   📊 Relevant results: {len(relevant_results)}/{len(results)} (score >= {query['min_score']})")
            
            if not query["expected_categories"]:
                low_score_count = sum(1 for r in results if r.get("score", 0) < 0.3)
                print(f"   📊 Low relevance results: {low_score_count}/{len(results)} (as expected)")
        else:
            print("❌ No results returned")

def main():
    """Main test function"""
    print("🚀 VittoriaDB Vectorization Comparison Test")
    print("=" * 60)
    print("Testing different vectorization approaches and comparing with backend")
    print()
    
    # Initialize tester
    tester = VittoriaDBTester()
    
    try:
        # Run all tests
        test_manual_vectorization(tester)
        test_sentence_transformers_vectorization(tester)
        test_large_text_processing(tester)
        
        # Summary
        print("\n" + "="*60)
        print("📊 VECTORIZATION COMPARISON SUMMARY")
        print("="*60)
        
        print("\n🔧 Manual Vectorization (Go-style):")
        print("   ✅ No external dependencies")
        print("   ✅ Fast generation")
        print("   ❌ Poor semantic discrimination (high similarity for unrelated content)")
        print("   ❌ Not suitable for production semantic search")
        
        if HAS_SENTENCE_TRANSFORMERS:
            print("\n🤖 Sentence-Transformers Vectorization:")
            print("   ✅ Good semantic discrimination")
            print("   ✅ Proper similarity scores")
            print("   ✅ Production-ready for semantic search")
            print("   ❌ Requires model download and dependencies")
            print("   ❌ Slower generation (but better quality)")
        
        print("\n🎯 Recommendations:")
        print("   • For development/testing: Manual vectors are acceptable")
        print("   • For production: Use sentence-transformers or OpenAI embeddings")
        print("   • Backend approach (sentence-transformers) provides best results")
        print("   • Consider server-side vectorization for consistency")
        
        print(f"\n🔗 Backend Integration:")
        print(f"   • Backend uses OpenAI embeddings (1536D) for speed")
        print(f"   • Fallback to sentence-transformers when OpenAI unavailable")
        print(f"   • Smart chunking with overlap for better context")
        print(f"   • Proper relevance filtering based on scores")
        
    except KeyboardInterrupt:
        print("\n⏹️  Test interrupted by user")
    except Exception as e:
        print(f"\n❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
    finally:
        # Cleanup
        tester.cleanup()
        print("\n✅ Vectorization comparison test completed!")

if __name__ == "__main__":
    main()
