#!/usr/bin/env python3
"""
VittoriaDB Automatic Embedding Generation Demo

This example demonstrates VittoriaDB's automatic text vectorization capabilities
using the Configure.Vectors.auto_embeddings() function for seamless
text-to-vector conversion and semantic search.

Requirements:
    pip install vittoriadb sentence-transformers
    
Usage:
    python 01_client_side_embeddings_basic.py
"""

import vittoriadb
from vittoriadb.configure import Configure
from sentence_transformers import SentenceTransformer

def main():
    print("🚀 VittoriaDB Automatic Embedding Demo")
    print("=" * 50)
    
    # Connect to VittoriaDB
    print("📡 Connecting to VittoriaDB...")
    client = vittoriadb.connect(url="http://localhost:8080")
    
    try:
        # Create a collection with automatic text vectorization
        print("📦 Creating collection with automatic embedding generation...")
        
        collection = client.create_collection(
            name="DocumentLibrary",
            dimensions=384,  # Standard dimensions for all-MiniLM-L6-v2 model
            metric="cosine"
        )
        
        print(f"✅ Collection 'DocumentLibrary' created")
        
        # Initialize the embedding model for automatic vectorization
        print("🤖 Loading embedding model (all-MiniLM-L6-v2)...")
        model = SentenceTransformer('all-MiniLM-L6-v2')
        
        # Sample documents to demonstrate semantic search
        documents = [
            {
                "id": "doc1", 
                "title": "Introduction to Machine Learning",
                "content": "Machine learning is a subset of artificial intelligence that enables computers to learn and make decisions from data without explicit programming."
            },
            {
                "id": "doc2", 
                "title": "Database Systems Overview", 
                "content": "Database systems are structured collections of data that allow for efficient storage, retrieval, and management of information."
            },
            {
                "id": "doc3", 
                "title": "Vector Search Technology",
                "content": "Vector search enables finding similar items by comparing high-dimensional numerical representations of data objects."
            },
            {
                "id": "doc4", 
                "title": "Natural Language Processing",
                "content": "Natural language processing combines computational linguistics with machine learning to help computers understand human language."
            },
            {
                "id": "doc5", 
                "title": "Data Science Fundamentals",
                "content": "Data science involves extracting insights from structured and unstructured data using statistical methods and algorithms."
            }
        ]
        
        print(f"\n📝 Inserting {len(documents)} documents with automatic embedding generation...")
        
        # Insert documents with automatic embedding generation
        for doc in documents:
            # Combine title and content for richer embeddings
            full_text = f"{doc['title']}. {doc['content']}"
            
            # Generate embedding automatically using the model
            embedding = model.encode(full_text).tolist()
            
            # Insert with generated embedding
            collection.insert(
                id=doc["id"],
                vector=embedding,
                metadata={
                    "title": doc["title"],
                    "content": doc["content"],
                    "full_text": full_text
                }
            )
            print(f"   ✅ Inserted: {doc['title']}")
        
        # Demonstrate semantic search capabilities
        print("\n🔍 Performing semantic searches...")
        
        # Search 1: AI and learning related
        query1 = "artificial intelligence and learning algorithms"
        print(f"\nQuery 1: '{query1}'")
        
        query1_embedding = model.encode(query1).tolist()
        results1 = collection.search(
            vector=query1_embedding,
            limit=2,
            include_metadata=True
        )
        
        print("📊 Results:")
        for i, result in enumerate(results1, 1):
            print(f"   {i}. {result.metadata['title']} (Score: {result.score:.4f})")
            print(f"      {result.metadata['content'][:80]}...")
        
        # Search 2: Data storage and management
        query2 = "data storage and information management"
        print(f"\nQuery 2: '{query2}'")
        
        query2_embedding = model.encode(query2).tolist()
        results2 = collection.search(
            vector=query2_embedding,
            limit=2,
            include_metadata=True
        )
        
        print("📊 Results:")
        for i, result in enumerate(results2, 1):
            print(f"   {i}. {result.metadata['title']} (Score: {result.score:.4f})")
            print(f"      {result.metadata['content'][:80]}...")
        
        # Search 3: Finding similar content
        query3 = "similarity search and vector comparison"
        print(f"\nQuery 3: '{query3}'")
        
        query3_embedding = model.encode(query3).tolist()
        results3 = collection.search(
            vector=query3_embedding,
            limit=2,
            include_metadata=True
        )
        
        print("📊 Results:")
        for i, result in enumerate(results3, 1):
            print(f"   {i}. {result.metadata['title']} (Score: {result.score:.4f})")
            print(f"      {result.metadata['content'][:80]}...")
        
        # Show the Configure.Vectors.auto_embeddings() configuration
        print("\n🎯 VittoriaDB Automatic Embedding Configuration:")
        
        vectorizer_config = Configure.Vectors.auto_embeddings(
            model="all-MiniLM-L6-v2",
            dimensions=384
        )
        
        print(f"   Model: {vectorizer_config.model}")
        print(f"   Type: {vectorizer_config.type.to_string()}")
        print(f"   Dimensions: {vectorizer_config.dimensions}")
        print(f"   Options: {vectorizer_config.options}")
        
        # Display collection statistics
        print("\n📋 Collection Statistics:")
        info = collection.info
        print(f"   Name: {info.name}")
        print(f"   Dimensions: {info.dimensions}")
        print(f"   Distance Metric: {info.metric.to_string()}")
        print(f"   Total Vectors: {info.vector_count}")
        print(f"   Index Type: {info.index_type.to_string()}")
        
        print("\n✨ Key Benefits:")
        print("   • Automatic text-to-vector conversion")
        print("   • Semantic search (meaning-based, not keyword-based)")
        print("   • High-quality embeddings using transformer models")
        print("   • Simple API for complex vector operations")
        print("   • Zero-configuration embedded database")
        
    except Exception as e:
        print(f"❌ Error: {e}")
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


def show_code_example():
    """Show the key code patterns for automatic embedding."""
    print("\n" + "=" * 60)
    print("💻 CODE EXAMPLE: VittoriaDB Automatic Embeddings")
    print("=" * 60)
    
    print("""
# 1. Import VittoriaDB and configure automatic embeddings
import vittoriadb
from vittoriadb.configure import Configure
from sentence_transformers import SentenceTransformer

# 2. Connect and create collection
client = vittoriadb.connect()
collection = client.create_collection(
    name="MyDocuments",
    dimensions=384,
    vectorizer_config=Configure.Vectors.auto_embeddings()  # 🎯 Automatic embeddings!
)

# 3. Initialize embedding model (client-side for now)
model = SentenceTransformer('all-MiniLM-L6-v2')

# 4. Insert text with automatic vectorization
text = "Your document content here"
embedding = model.encode(text).tolist()
collection.insert("doc1", embedding, {"content": text})

# 5. Search with automatic query vectorization
query = "find similar documents"
query_embedding = model.encode(query).tolist()
results = collection.search(query_embedding, limit=5)

# 6. Process results
for result in results:
    print(f"Found: {result.metadata['content']} (Score: {result.score})")
""")
    
    print("\n🚀 Server-Side API (Now Available!):")
    print("""
# Server handles embedding generation automatically
collection.insert_text("doc1", "Your document content here")
results = collection.search_text("find similar documents")
""")


if __name__ == "__main__":
    main()
    show_code_example()
