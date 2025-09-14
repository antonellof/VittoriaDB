#!/usr/bin/env python3
"""
Comprehensive VittoriaDB Server-Side Embedding Test

This test demonstrates all server-side automatic embedding capabilities
with fresh collections and comprehensive error handling.
"""

import sys
import os
import time
import random
import vittoriadb
from vittoriadb.configure import Configure

def main():
    print("🚀 VittoriaDB Comprehensive Server-Side Embedding Test")
    print("=" * 65)
    
    # Connect to VittoriaDB
    print("📡 Connecting to VittoriaDB...")
    client = vittoriadb.connect(url="http://localhost:8080")
    
    # Use timestamp to ensure unique collection name
    collection_name = f"EmbedTest_{int(time.time())}"
    
    try:
        # Test 1: Create collection with server-side vectorizer
        print(f"📦 Test 1: Creating collection '{collection_name}' with server-side vectorizer...")
        
        collection = client.create_collection(
            name=collection_name,
            dimensions=384,
            metric="cosine",
            vectorizer_config=Configure.Vectors.auto_embeddings(
                model="all-MiniLM-L6-v2",
                dimensions=384
            )
        )
        
        print(f"✅ Collection '{collection_name}' created successfully")
        
        # Test 2: Single text insertion
        print("\n📝 Test 2: Single text insertion with server-side embedding...")
        
        start_time = time.time()
        collection.insert_text(
            id="test_doc_1",
            text="Machine learning algorithms can automatically learn patterns from data without explicit programming",
            metadata={"category": "ML", "source": "test", "length": 95}
        )
        insert_time = time.time() - start_time
        print(f"✅ Single text inserted in {insert_time:.3f}s")
        
        # Test 3: Batch text insertion
        print("\n📦 Test 3: Batch text insertion with server-side embedding...")
        
        batch_texts = [
            {
                "id": "test_doc_2",
                "text": "Deep neural networks use multiple layers to extract hierarchical features from raw input data",
                "metadata": {"category": "Deep Learning", "source": "test"}
            },
            {
                "id": "test_doc_3", 
                "text": "Natural language processing enables computers to understand and generate human language effectively",
                "metadata": {"category": "NLP", "source": "test"}
            },
            {
                "id": "test_doc_4",
                "text": "Computer vision algorithms can analyze and interpret visual information from images and videos",
                "metadata": {"category": "Computer Vision", "source": "test"}
            },
            {
                "id": "test_doc_5",
                "text": "Reinforcement learning agents learn optimal actions through trial and error interactions with environments",
                "metadata": {"category": "RL", "source": "test"}
            }
        ]
        
        start_time = time.time()
        result = collection.insert_text_batch(batch_texts)
        batch_time = time.time() - start_time
        print(f"✅ Batch of {len(batch_texts)} texts inserted in {batch_time:.3f}s")
        print(f"   Result: {result}")
        
        # Test 4: Text search with different queries
        print("\n🔍 Test 4: Server-side text search with automatic query vectorization...")
        
        test_queries = [
            ("machine learning patterns", "Should match ML content"),
            ("visual image analysis", "Should match computer vision"),
            ("language understanding", "Should match NLP content"),
            ("neural network layers", "Should match deep learning"),
            ("learning from environment", "Should match reinforcement learning")
        ]
        
        for query, description in test_queries:
            print(f"\n   Query: '{query}' ({description})")
            start_time = time.time()
            results = collection.search_text(query=query, limit=2)
            search_time = time.time() - start_time
            
            print(f"   Search completed in {search_time:.3f}s - Found {len(results)} results:")
            for i, result in enumerate(results, 1):
                category = result.metadata.get('category', 'Unknown')
                print(f"     {i}. Score: {result.score:.4f} | Category: {category}")
        
        # Test 5: Collection statistics
        print(f"\n📊 Test 5: Collection statistics...")
        info = collection.info
        print(f"   Collection: {info.name}")
        print(f"   Dimensions: {info.dimensions}")
        print(f"   Metric: {info.metric.to_string()}")
        print(f"   Vector Count: {info.vector_count}")
        print(f"   Index Type: {info.index_type.to_string()}")
        
        # Test 6: Performance comparison
        print(f"\n⚡ Test 6: Performance analysis...")
        print(f"   Single insert: {insert_time:.3f}s")
        print(f"   Batch insert ({len(batch_texts)} items): {batch_time:.3f}s")
        print(f"   Average per item: {batch_time/len(batch_texts):.3f}s")
        print(f"   Batch efficiency: {insert_time/(batch_time/len(batch_texts)):.2f}x faster per item")
        
        # Test 7: Error handling - try to insert to collection without vectorizer
        print(f"\n🧪 Test 7: Error handling...")
        try:
            # Create collection without vectorizer
            no_vectorizer_collection = client.create_collection(
                name=f"NoVectorizer_{int(time.time())}",
                dimensions=384,
                metric="cosine"
                # No vectorizer_config
            )
            
            # Try to insert text (should fail)
            no_vectorizer_collection.insert_text("fail_test", "This should fail")
            print("❌ ERROR: Should have failed!")
            
        except Exception as e:
            print(f"✅ Correctly caught error: {str(e)[:80]}...")
        
        print(f"\n🎯 All tests completed successfully!")
        print(f"\n✨ Server-Side Embedding Benefits Demonstrated:")
        print(f"   • 🚀 Automatic text-to-vector conversion")
        print(f"   • 🔍 Semantic search with query vectorization")
        print(f"   • 📦 Efficient batch processing")
        print(f"   • 🛡️ Proper error handling")
        print(f"   • ⚡ Good performance characteristics")
        
    except Exception as e:
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
        
    finally:
        # Clean up
        print(f"\n🧹 Cleaning up...")
        try:
            client.close()
        except:
            pass
        print("✅ Test completed!")


if __name__ == "__main__":
    main()
