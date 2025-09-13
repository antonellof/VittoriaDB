#!/usr/bin/env python3
"""
Basic VittoriaDB usage example.

This example demonstrates:
1. Connecting to VittoriaDB (auto-starts server)
2. Creating a collection
3. Inserting vectors
4. Searching for similar vectors
5. Managing collections

Run this after installing VittoriaDB:
    pip install vittoriadb
    python examples/basic_usage.py
"""

import numpy as np
import time
import vittoriadb


def main():
    print("🚀 VittoriaDB Basic Usage Example")
    print("=" * 40)
    
    # Connect to VittoriaDB (connect to existing server)
    print("1. Connecting to VittoriaDB...")
    db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)
    
    # Check health
    health = db.health()
    print(f"   Status: {health.status}")
    print(f"   Uptime: {health.uptime}s")
    
    # Create a collection
    print("\n2. Creating collection 'documents'...")
    try:
        collection = db.create_collection(
            name="documents",
            dimensions=384,  # Common embedding dimension
            metric="cosine"
        )
        print(f"   Collection created: {collection.name}")
    except vittoriadb.CollectionError as e:
        if "already exists" in str(e):
            print("   Collection already exists, using existing one...")
            collection = db.get_collection("documents")
        else:
            raise
    
    print(f"   Dimensions: {collection.info.dimensions}")
    print(f"   Metric: {collection.info.metric.to_string()}")
    
    # Insert some sample vectors
    print("\n3. Inserting sample vectors...")
    
    # Sample document embeddings (normally from a model like sentence-transformers)
    documents = [
        {
            "id": "doc1",
            "vector": np.random.random(384).tolist(),
            "metadata": {
                "title": "Introduction to Vector Databases",
                "category": "technology",
                "author": "Alice",
                "year": 2023
            }
        },
        {
            "id": "doc2", 
            "vector": np.random.random(384).tolist(),
            "metadata": {
                "title": "Machine Learning Fundamentals",
                "category": "technology",
                "author": "Bob",
                "year": 2023
            }
        },
        {
            "id": "doc3",
            "vector": np.random.random(384).tolist(),
            "metadata": {
                "title": "Cooking with Python",
                "category": "cooking",
                "author": "Charlie",
                "year": 2022
            }
        }
    ]
    
    # Insert vectors one by one
    for doc in documents:
        collection.insert(doc["id"], doc["vector"], doc["metadata"])
        print(f"   Inserted: {doc['metadata']['title']}")
    
    # Batch insert example
    print("\n4. Batch inserting more vectors...")
    batch_docs = []
    for i in range(5):
        batch_docs.append({
            "id": f"batch_doc_{i}",
            "vector": np.random.random(384).tolist(),
            "metadata": {
                "title": f"Batch Document {i}",
                "category": "batch",
                "index": i
            }
        })
    
    result = collection.insert_batch(batch_docs)
    print(f"   Batch inserted: {result['inserted']} vectors")
    
    # Check collection stats
    print(f"\n5. Collection stats:")
    print(f"   Total vectors: {collection.count()}")
    
    # Search for similar vectors
    print("\n6. Searching for similar vectors...")
    
    # Create a query vector (normally this would be an embedding of a query)
    query_vector = np.random.random(384).tolist()
    
    # Search with different parameters
    results = collection.search(
        vector=query_vector,
        limit=3,
        include_metadata=True
    )
    
    print(f"   Found {len(results)} similar vectors:")
    for i, result in enumerate(results, 1):
        print(f"   {i}. ID: {result.id}")
        print(f"      Score: {result.score:.4f}")
        print(f"      Title: {result.metadata.get('title', 'N/A')}")
        print(f"      Category: {result.metadata.get('category', 'N/A')}")
        print()
    
    # Search with metadata filter
    print("7. Searching with metadata filter...")
    filtered_results = collection.search(
        vector=query_vector,
        limit=5,
        filter={"category": "technology"},
        include_metadata=True
    )
    
    print(f"   Found {len(filtered_results)} technology documents:")
    for result in filtered_results:
        print(f"   - {result.metadata.get('title', 'N/A')} (score: {result.score:.4f})")
    
    # Get specific vector
    print("\n8. Retrieving specific vector...")
    doc1 = collection.get("doc1")
    if doc1:
        print(f"   Retrieved: {doc1.metadata.get('title', 'N/A')}")
        print(f"   Vector dimensions: {len(doc1.vector)}")
    
    # List all collections
    print("\n9. Listing all collections...")
    collections = db.list_collections()
    for coll in collections:
        print(f"   - {coll.name}: {coll.vector_count} vectors, {coll.dimensions}D")
    
    # Database statistics
    print("\n10. Database statistics...")
    stats = db.stats()
    print(f"    Total vectors: {stats.total_vectors}")
    print(f"    Total size: {stats.total_size} bytes")
    print(f"    Collections: {len(stats.collections)}")
    
    # Cleanup
    print("\n11. Cleaning up...")
    
    # Delete a vector
    collection.delete("doc1")
    print(f"    Deleted doc1, remaining: {collection.count()} vectors")
    
    # Delete collection
    db.delete_collection("documents")
    print("    Deleted collection 'documents'")
    
    # Close connection
    db.close()
    print("    Closed connection")
    
    print("\n✅ Example completed successfully!")
    print("\nNext steps:")
    print("- Try the RAG example: python examples/rag_example.py")
    print("- Check the REST API: curl http://localhost:8080/health")
    print("- Visit the dashboard: http://localhost:8080/")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"\n❌ Error: {e}")
        print("\nTroubleshooting:")
        print("1. Make sure VittoriaDB is installed: pip install vittoriadb")
        print("2. Check if port 8080 is available")
        print("3. Try running with: python -m vittoriadb.examples.basic_usage")
