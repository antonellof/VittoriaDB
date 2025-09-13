#!/usr/bin/env python3
"""
VittoriaDB Server-Side Automatic Embedding Demo

This example demonstrates VittoriaDB's server-side automatic text vectorization
using the Configure.Vectors.auto_embeddings() function. The server handles
all embedding generation automatically - no client-side model loading required!

Requirements:
    Server must have sentence-transformers installed

Usage:
    python server_side_embedding_demo.py
"""

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..', 'python'))

import vittoriadb
from vittoriadb.configure import Configure

def main():
    print("🚀 VittoriaDB Server-Side Automatic Embedding Demo")
    print("=" * 60)
    
    # Connect to VittoriaDB
    print("📡 Connecting to VittoriaDB...")
    client = vittoriadb.connect(url="http://localhost:8080")
    
    try:
        # Create a collection with server-side automatic text vectorization
        print("📦 Creating collection with server-side automatic embedding generation...")
        
        collection = client.create_collection(
            name="ServerSideEmbeddings",
            dimensions=384,  # Standard dimensions for all-MiniLM-L6-v2 model
            metric="cosine",
            vectorizer_config=Configure.Vectors.auto_embeddings()  # 🎯 Server-side magic!
        )
        
        print(f"✅ Collection 'ServerSideEmbeddings' created with server-side vectorization")
        
        # Sample articles for demonstration
        articles = [
            {
                "id": "article1", 
                "text": "Artificial intelligence is transforming how we process and understand data through advanced machine learning algorithms."
            },
            {
                "id": "article2", 
                "text": "Vector databases provide efficient storage and retrieval of high-dimensional embeddings for similarity search applications."
            },
            {
                "id": "article3", 
                "text": "Natural language processing enables computers to understand, interpret, and generate human language in meaningful ways."
            },
            {
                "id": "article4", 
                "text": "Deep learning neural networks can automatically extract features from raw data without manual feature engineering."
            },
            {
                "id": "article5", 
                "text": "Semantic search goes beyond keyword matching to understand the intent and contextual meaning of queries."
            }
        ]
        
        print(f"\n📝 Inserting {len(articles)} articles using server-side embedding generation...")
        
        # Insert articles using server-side automatic vectorization
        # No client-side model loading or embedding generation required!
        for article in articles:
            collection.insert_text(
                id=article["id"],
                text=article["text"],
                metadata={"content": article["text"]}
            )
            print(f"   ✅ Server vectorized: {article['text'][:60]}...")
        
        print("\n🎯 Server handled all embedding generation automatically!")
        
        # Demonstrate server-side semantic search
        print("\n🔍 Performing server-side semantic searches...")
        
        # Search 1: AI and machine learning
        query1 = "machine learning and artificial intelligence"
        print(f"\nQuery 1: '{query1}'")
        
        # Server automatically generates query embedding and searches
        results1 = collection.search_text(query=query1, limit=2)
        
        print("📊 Results (server-side vectorization):")
        for i, result in enumerate(results1, 1):
            print(f"   {i}. Score: {result.score:.4f}")
            print(f"      Content: {result.metadata['content'][:80]}...")
        
        # Search 2: Database and storage
        query2 = "database storage and data retrieval"
        print(f"\nQuery 2: '{query2}'")
        
        results2 = collection.search_text(query=query2, limit=2)
        
        print("📊 Results (server-side vectorization):")
        for i, result in enumerate(results2, 1):
            print(f"   {i}. Score: {result.score:.4f}")
            print(f"      Content: {result.metadata['content'][:80]}...")
        
        # Search 3: Language understanding
        query3 = "understanding human language and communication"
        print(f"\nQuery 3: '{query3}'")
        
        results3 = collection.search_text(query=query3, limit=2)
        
        print("📊 Results (server-side vectorization):")
        for i, result in enumerate(results3, 1):
            print(f"   {i}. Score: {result.score:.4f}")
            print(f"      Content: {result.metadata['content'][:80]}...")
        
        # Demonstrate batch insertion with server-side vectorization
        print(f"\n📦 Batch inserting additional articles...")
        
        batch_articles = [
            {"id": "batch1", "text": "Computer vision algorithms can analyze and interpret visual information from images and videos."},
            {"id": "batch2", "text": "Recommendation systems use collaborative filtering to suggest relevant items based on user preferences."},
            {"id": "batch3", "text": "Information retrieval systems help users find relevant documents from large collections of text."}
        ]
        
        # Server handles batch embedding generation automatically
        collection.insert_text_batch(batch_articles)
        print(f"✅ Server processed {len(batch_articles)} articles in batch")
        
        # Final search to show all content
        query4 = "computer algorithms and data processing"
        print(f"\nFinal Query: '{query4}'")
        
        results4 = collection.search_text(query=query4, limit=3)
        
        print("📊 Final Results:")
        for i, result in enumerate(results4, 1):
            print(f"   {i}. Score: {result.score:.4f}")
            print(f"      Content: {result.metadata['content'][:80]}...")
        
        # Show collection statistics
        print("\n📋 Collection Statistics:")
        info = collection.info
        print(f"   Name: {info.name}")
        print(f"   Dimensions: {info.dimensions}")
        print(f"   Distance Metric: {info.metric.to_string()}")
        print(f"   Total Vectors: {info.vector_count}")
        print(f"   Index Type: {info.index_type.to_string()}")
        
        print("\n✨ Server-Side Benefits:")
        print("   • No client-side model loading or memory usage")
        print("   • Consistent embeddings across all clients")
        print("   • Centralized model management and updates")
        print("   • Reduced client bandwidth and processing")
        print("   • Automatic batching and optimization")
        
    except Exception as e:
        print(f"❌ Error: {e}")
        
        # Check if it's a vectorizer configuration error
        if "vectorizer" in str(e).lower():
            print("\n💡 Note: This demo requires server-side vectorizer support.")
            print("   The server needs to have sentence-transformers installed and")
            print("   the collection must be created with vectorizer_config.")
        
        import traceback
        traceback.print_exc()
        
    finally:
        # Clean up
        print("\n🧹 Cleaning up...")
        try:
            client.close()
        except:
            pass
        print("✅ Demo completed!")


def show_api_comparison():
    """Show the API comparison between client-side and server-side embedding."""
    print("\n" + "=" * 70)
    print("📊 API COMPARISON: Client-Side vs Server-Side Embeddings")
    print("=" * 70)
    
    print("\n🟨 Client-Side Embedding (Previous Approach):")
    print("""
import vittoriadb
from sentence_transformers import SentenceTransformer

client = vittoriadb.connect()
model = SentenceTransformer('all-MiniLM-L6-v2')  # Client loads model

collection = client.create_collection("docs", dimensions=384)

# Client generates embeddings
text = "Your document content"
embedding = model.encode(text).tolist()
collection.insert("doc1", embedding, {"content": text})

# Client generates query embedding
query_embedding = model.encode("search query").tolist()
results = collection.search(query_embedding)
""")
    
    print("\n🟩 Server-Side Embedding (New Approach):")
    print("""
import vittoriadb
from vittoriadb.configure import Configure

client = vittoriadb.connect()

# Server handles all embedding generation
collection = client.create_collection(
    name="docs", 
    dimensions=384,
    vectorizer_config=Configure.Vectors.auto_embeddings()  # 🎯 Server-side!
)

# Server automatically generates embeddings
collection.insert_text("doc1", "Your document content")

# Server automatically generates query embedding
results = collection.search_text("search query")
""")
    
    print("\n🚀 Key Advantages of Server-Side Approach:")
    print("   • 📦 No client-side dependencies (no sentence-transformers install)")
    print("   • 🧠 No client memory usage for models")
    print("   • ⚡ Faster client startup (no model loading)")
    print("   • 🔄 Consistent embeddings across all clients")
    print("   • 🛠️ Centralized model management")
    print("   • 📊 Server can optimize batching and caching")


if __name__ == "__main__":
    main()
    show_api_comparison()
