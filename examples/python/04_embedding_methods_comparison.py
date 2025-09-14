#!/usr/bin/env python3
"""
VittoriaDB Complete Embedding Showcase

This showcase demonstrates both client-side and server-side automatic 
embedding generation approaches, showing the evolution from manual 
embeddings to fully automated server-side processing.
"""

import sys
import os
import time
import vittoriadb
from vittoriadb.configure import Configure

def main():
    print("🚀 VittoriaDB Complete Embedding Showcase")
    print("=" * 55)
    print("Demonstrating the evolution of embedding generation:")
    print("1. Manual embeddings → 2. Client-side auto → 3. Server-side auto")
    print()
    
    # Connect to VittoriaDB
    print("📡 Connecting to VittoriaDB...")
    client = vittoriadb.connect(url="http://localhost:8080")
    
    timestamp = int(time.time())
    
    try:
        # Approach 1: Manual Embeddings (Traditional)
        print("\n" + "="*60)
        print("📊 APPROACH 1: Manual Embeddings (Traditional)")
        print("="*60)
        
        manual_collection = client.create_collection(
            name=f"ManualEmbeddings_{timestamp}",
            dimensions=4,  # Simple 4D vectors for demo
            metric="cosine"
        )
        
        print("✅ Created collection for manual embeddings")
        
        # Insert manually created vectors
        manual_vectors = [
            {"id": "manual1", "vector": [0.8, 0.2, 0.1, 0.3], "metadata": {"text": "Technology and innovation", "approach": "manual"}},
            {"id": "manual2", "vector": [0.1, 0.9, 0.2, 0.1], "metadata": {"text": "Science and research", "approach": "manual"}},
            {"id": "manual3", "vector": [0.3, 0.1, 0.8, 0.4], "metadata": {"text": "Business and finance", "approach": "manual"}},
        ]
        
        for vec in manual_vectors:
            manual_collection.insert(vec["id"], vec["vector"], vec["metadata"])
        
        print(f"✅ Inserted {len(manual_vectors)} manually created vectors")
        
        # Search with manual query vector
        query_vector = [0.7, 0.3, 0.2, 0.2]  # Similar to "technology"
        results = manual_collection.search(query_vector, limit=2)
        
        print("🔍 Manual search results:")
        for i, result in enumerate(results, 1):
            print(f"   {i}. {result.metadata['text']} (Score: {result.score:.4f})")
        
        # Approach 2: Client-Side Automatic Embeddings
        print("\n" + "="*60)
        print("🤖 APPROACH 2: Client-Side Automatic Embeddings")
        print("="*60)
        
        try:
            from sentence_transformers import SentenceTransformer
            
            client_collection = client.create_collection(
                name=f"ClientSideEmbeddings_{timestamp}",
                dimensions=384,
                metric="cosine"
            )
            
            print("✅ Created collection for client-side embeddings")
            print("🤖 Loading sentence transformer model on client...")
            
            model = SentenceTransformer('all-MiniLM-L6-v2')
            print("✅ Model loaded on client")
            
            # Insert with client-side embedding generation
            client_texts = [
                {"id": "client1", "text": "Artificial intelligence and machine learning algorithms", "category": "AI"},
                {"id": "client2", "text": "Database systems and data storage solutions", "category": "Database"},
                {"id": "client3", "text": "Web development and user interface design", "category": "Web"},
            ]
            
            for item in client_texts:
                embedding = model.encode(item["text"]).tolist()
                client_collection.insert(item["id"], embedding, {
                    "text": item["text"], 
                    "category": item["category"],
                    "approach": "client-side"
                })
            
            print(f"✅ Inserted {len(client_texts)} texts with client-side embeddings")
            
            # Search with client-side query embedding
            query_text = "machine learning and AI systems"
            query_embedding = model.encode(query_text).tolist()
            results = client_collection.search(query_embedding, limit=2)
            
            print(f"🔍 Client-side search results for '{query_text}':")
            for i, result in enumerate(results, 1):
                print(f"   {i}. {result.metadata['text'][:50]}... (Score: {result.score:.4f})")
            
        except ImportError:
            print("⚠️  Sentence transformers not available - skipping client-side demo")
            print("   Install with: pip install sentence-transformers")
        
        # Approach 3: Server-Side Automatic Embeddings
        print("\n" + "="*60)
        print("🚀 APPROACH 3: Server-Side Automatic Embeddings")
        print("="*60)
        
        server_collection = client.create_collection(
            name=f"ServerSideEmbeddings_{timestamp}",
            dimensions=384,
            metric="cosine",
            vectorizer_config=Configure.Vectors.auto_embeddings()
        )
        
        print("✅ Created collection with server-side vectorizer")
        print("🚀 Server handles all embedding generation automatically")
        
        # Insert with server-side embedding generation
        server_texts = [
            {"id": "server1", "text": "Deep neural networks and transformer architectures enable advanced natural language understanding", "category": "AI"},
            {"id": "server2", "text": "Vector databases provide efficient similarity search and retrieval for high-dimensional data", "category": "Database"},
            {"id": "server3", "text": "Modern web frameworks and responsive design create engaging user experiences", "category": "Web"},
        ]
        
        start_time = time.time()
        for item in server_texts:
            server_collection.insert_text(item["id"], item["text"], {
                "category": item["category"],
                "approach": "server-side"
            })
        server_time = time.time() - start_time
        
        print(f"✅ Inserted {len(server_texts)} texts with server-side embeddings in {server_time:.2f}s")
        
        # Search with server-side query embedding
        query_text = "neural networks and AI technology"
        start_time = time.time()
        results = server_collection.search_text(query_text, limit=2)
        search_time = time.time() - start_time
        
        print(f"🔍 Server-side search results for '{query_text}' (in {search_time:.2f}s):")
        for i, result in enumerate(results, 1):
            text = result.metadata.get('content', result.metadata.get('text', 'No text'))
            print(f"   {i}. {text[:50]}... (Score: {result.score:.4f})")
        
        # Comparison Summary
        print("\n" + "="*60)
        print("📊 COMPARISON SUMMARY")
        print("="*60)
        
        print("\n🔧 Manual Embeddings:")
        print("   ✅ Full control over vector generation")
        print("   ❌ Requires domain expertise")
        print("   ❌ Time-consuming and error-prone")
        print("   ❌ Not scalable for large datasets")
        
        print("\n🤖 Client-Side Automatic:")
        print("   ✅ Automatic embedding generation")
        print("   ✅ Uses state-of-the-art models")
        print("   ❌ Requires client-side dependencies")
        print("   ❌ Model loading overhead")
        print("   ❌ Inconsistent across clients")
        
        print("\n🚀 Server-Side Automatic:")
        print("   ✅ Fully automatic embedding generation")
        print("   ✅ No client-side dependencies")
        print("   ✅ Consistent across all clients")
        print("   ✅ Centralized model management")
        print("   ✅ Optimized server-side processing")
        print("   ✅ Zero client setup required")
        
        print(f"\n🎯 Winner: Server-Side Automatic Embeddings! 🏆")
        print(f"   The future of vector databases is server-side automation.")
        
        # Show collection statistics
        print(f"\n📊 Final Statistics:")
        collections = [
            (manual_collection, "Manual"),
            (server_collection, "Server-Side")
        ]
        
        if 'client_collection' in locals():
            collections.insert(1, (client_collection, "Client-Side"))
        
        for collection, name in collections:
            info = collection.info
            print(f"   {name}: {info.vector_count} vectors, {info.dimensions}D")
        
    except Exception as e:
        print(f"❌ Error: {e}")
        import traceback
        traceback.print_exc()
        
    finally:
        # Clean up
        print(f"\n🧹 Cleaning up...")
        try:
            client.close()
        except:
            pass
        print("✅ Showcase completed!")


if __name__ == "__main__":
    main()
