#!/usr/bin/env python3
"""
VittoriaDB Document API Showcase

This example demonstrates the new document-oriented API that provides:
- Document-oriented database with flexible schemas
- Full-text search with BM25 scoring and advanced features
- Vector similarity search with multiple distance metrics
- Hybrid search combining text and vector search
- Advanced filtering, facets, sorting, and grouping
- Modern API similar to leading vector databases

Features showcased:
✅ Schema-based document structure
✅ Multiple search modes (fulltext, vector, hybrid)
✅ Advanced query capabilities (filters, facets, boosting)
✅ Flexible document insertion and management
✅ Production-ready error handling

Requirements:
    pip install vittoriadb numpy

Usage:
    python examples/python/15_document_api_showcase.py
"""

import numpy as np
import time
import vittoriadb
from vittoriadb import create


def main():
    print("🚀 VittoriaDB Document API Showcase")
    print("=" * 50)
    
    # Define a rich document schema
    print("1. Creating document database with schema...")
    schema = {
        "name": "string",
        "description": "string", 
        "price": "number",
        "category": "string",
        "tags": "string[]",
        "embedding": "vector[384]",  # Vector field for semantic search
        "meta": {
            "rating": "number",
            "reviews": "number",
            "brand": "string"
        },
        "available": "boolean"
    }
    
    # Create unified database (similar to modern vector databases)
    db = create(schema, language="english")
    print(f"   ✅ Database created with schema")
    print(f"   📋 Schema fields: {list(schema.keys())}")
    
    # Insert sample documents
    print("\n2. Inserting sample documents...")
    
    sample_docs = [
        {
            "name": "Noise Cancelling Headphones",
            "description": "Premium wireless headphones with active noise cancellation and superior sound quality",
            "price": 299.99,
            "category": "electronics",
            "tags": ["audio", "wireless", "premium", "noise-cancelling"],
            "embedding": np.random.random(384).tolist(),
            "meta": {
                "rating": 4.8,
                "reviews": 1250,
                "brand": "AudioTech"
            },
            "available": True
        },
        {
            "name": "Wireless Gaming Mouse",
            "description": "High-precision gaming mouse with customizable RGB lighting and programmable buttons",
            "price": 89.99,
            "category": "electronics", 
            "tags": ["gaming", "wireless", "rgb", "precision"],
            "embedding": np.random.random(384).tolist(),
            "meta": {
                "rating": 4.6,
                "reviews": 890,
                "brand": "GameGear"
            },
            "available": True
        },
        {
            "name": "Organic Coffee Beans",
            "description": "Single-origin organic coffee beans with rich flavor and smooth finish",
            "price": 24.99,
            "category": "food",
            "tags": ["organic", "coffee", "single-origin", "premium"],
            "embedding": np.random.random(384).tolist(),
            "meta": {
                "rating": 4.9,
                "reviews": 456,
                "brand": "BrewMaster"
            },
            "available": True
        },
        {
            "name": "Bluetooth Speaker",
            "description": "Portable waterproof Bluetooth speaker with 360-degree sound and long battery life",
            "price": 149.99,
            "category": "electronics",
            "tags": ["audio", "bluetooth", "portable", "waterproof"],
            "embedding": np.random.random(384).tolist(),
            "meta": {
                "rating": 4.4,
                "reviews": 723,
                "brand": "SoundWave"
            },
            "available": False
        },
        {
            "name": "Yoga Mat",
            "description": "Eco-friendly non-slip yoga mat with excellent grip and cushioning",
            "price": 39.99,
            "category": "fitness",
            "tags": ["yoga", "eco-friendly", "non-slip", "exercise"],
            "embedding": np.random.random(384).tolist(),
            "meta": {
                "rating": 4.7,
                "reviews": 334,
                "brand": "ZenFit"
            },
            "available": True
        }
    ]
    
    inserted_ids = []
    for doc in sample_docs:
        doc_id = db.insert(doc)
        inserted_ids.append(doc_id)
        print(f"   📄 Inserted: {doc['name']} (ID: {doc_id})")
    
    print(f"\n   ✅ Inserted {len(inserted_ids)} documents")
    
    # Demonstrate full-text search
    print("\n3. Full-text search examples...")
    
    # Basic text search
    print("   🔍 Basic text search for 'wireless':")
    results = db.search_text("wireless", limit=3)
    print(f"      Found {results['count']} results in {results['elapsed']}:")
    for i, hit in enumerate(results['hits'], 1):
        doc = hit['document']
        print(f"      {i}. {doc['name']} (score: {hit['score']:.3f})")
        print(f"         Price: ${doc['price']}, Category: {doc['category']}")
    
    # Advanced text search with filters and boosting
    print("\n   🎯 Advanced text search with filters and boosting:")
    results = db.search(
        term="premium audio",
        mode="fulltext",
        where={"category": "electronics", "available": True},
        boost={"name": 2.0, "description": 1.0},  # Boost name field
        limit=3
    )
    print(f"      Found {results['count']} electronics results:")
    for hit in results['hits']:
        doc = hit['document']
        print(f"      • {doc['name']} - ${doc['price']} (score: {hit['score']:.3f})")
    
    # Demonstrate vector search
    print("\n4. Vector similarity search...")
    
    # Create a query vector (in real use, this would be from an embedding model)
    query_vector = np.random.random(384).tolist()
    
    print("   🎯 Vector similarity search:")
    results = db.search_vector(
        vector_value=query_vector,
        vector_property="embedding",
        similarity=0.0,  # Lower threshold for demo
        limit=3
    )
    print(f"      Found {results['count']} similar items:")
    for hit in results['hits']:
        doc = hit['document']
        print(f"      • {doc['name']} (similarity: {hit['score']:.3f})")
        print(f"        Category: {doc['category']}, Rating: {doc['meta']['rating']}")
    
    # Demonstrate hybrid search
    print("\n5. Hybrid search (text + vector)...")
    
    print("   🔀 Hybrid search combining text and vector similarity:")
    results = db.search_hybrid(
        term="high quality audio",
        vector_value=query_vector,
        vector_property="embedding", 
        text_weight=0.7,    # Favor text search
        vector_weight=0.3,  # Some vector influence
        limit=3
    )
    print(f"      Found {results['count']} hybrid results:")
    for hit in results['hits']:
        doc = hit['document']
        print(f"      • {doc['name']} (combined score: {hit['score']:.3f})")
        print(f"        {doc['description'][:60]}...")
    
    # Demonstrate faceted search
    print("\n6. Faceted search and analytics...")
    
    print("   📊 Search with facets for analytics:")
    results = db.search(
        term="*",  # Match all
        mode="fulltext",
        facets={
            "category": {"type": "string", "limit": 10},
            "meta.brand": {"type": "string", "limit": 10},
            "price": {"type": "number", "ranges": [
                {"from": 0, "to": 50},
                {"from": 50, "to": 150}, 
                {"from": 150, "to": 500}
            ]}
        },
        limit=10
    )
    
    print(f"      Total documents: {results['count']}")
    if 'facets' in results:
        print("      📈 Category breakdown:")
        for category, count in results['facets']['category']['values'].items():
            print(f"         {category}: {count} items")
        
        print("      💰 Price ranges:")
        for price_range, count in results['facets']['price']['values'].items():
            print(f"         ${price_range}: {count} items")
    
    # Demonstrate advanced filtering
    print("\n7. Advanced filtering and sorting...")
    
    print("   🎛️ Complex filter: electronics under $200 with rating > 4.5:")
    results = db.search(
        mode="fulltext",
        where={
            "and": [
                {"category": "electronics"},
                {"price": {"lt": 200}},
                {"meta.rating": {"gt": 4.5}},
                {"available": True}
            ]
        },
        sort_by={"property": "meta.rating", "order": "desc"},
        limit=5
    )
    
    print(f"      Found {results['count']} matching products:")
    for hit in results['hits']:
        doc = hit['document']
        print(f"      • {doc['name']} - ${doc['price']}")
        print(f"        Rating: {doc['meta']['rating']}/5.0 ({doc['meta']['reviews']} reviews)")
    
    # Demonstrate document management
    print("\n8. Document management operations...")
    
    # Get specific document
    first_id = inserted_ids[0]
    print(f"   📖 Retrieving document {first_id}:")
    doc = db.get(first_id, include_vectors=False)
    if doc:
        print(f"      Found: {doc['name']}")
        print(f"      Price: ${doc['price']}, Rating: {doc['meta']['rating']}")
    
    # Update document
    print(f"\n   ✏️ Updating document {first_id}:")
    updated = db.update(first_id, {
        "name": "Premium Noise Cancelling Headphones", 
        "price": 279.99,  # Price drop!
        "meta": {
            "rating": 4.9,
            "reviews": 1350,
            "brand": "AudioTech"
        }
    })
    if updated:
        print("      ✅ Document updated successfully")
        
        # Verify update
        updated_doc = db.get(first_id)
        if updated_doc:
            print(f"      New name: {updated_doc['name']}")
            print(f"      New price: ${updated_doc['price']}")
    
    # Count documents
    print("\n9. Database statistics...")
    
    total_count = db.count()
    electronics_count = db.count(where={"category": "electronics"})
    available_count = db.count(where={"available": True})
    
    print(f"   📊 Total documents: {total_count}")
    print(f"   📱 Electronics: {electronics_count}")
    print(f"   ✅ Available items: {available_count}")
    
    # Performance demonstration
    print("\n10. Performance showcase...")
    
    print("   ⚡ Rapid search performance test:")
    start_time = time.time()
    
    # Perform multiple searches
    search_types = [
        ("Text search", lambda: db.search_text("premium", limit=5)),
        ("Vector search", lambda: db.search_vector(query_vector, "embedding", limit=5)),
        ("Hybrid search", lambda: db.search_hybrid("quality", query_vector, "embedding", limit=5)),
        ("Filtered search", lambda: db.search(where={"category": "electronics"}, limit=5))
    ]
    
    for search_name, search_func in search_types:
        search_start = time.time()
        results = search_func()
        search_time = (time.time() - search_start) * 1000
        print(f"      {search_name}: {len(results['hits'])} results in {search_time:.1f}ms")
    
    total_time = (time.time() - start_time) * 1000
    print(f"   ⏱️ Total time for 4 searches: {total_time:.1f}ms")
    
    # Cleanup demonstration
    print("\n11. Cleanup...")
    
    deleted_count = 0
    for doc_id in inserted_ids[1:]:  # Keep first document
        if db.delete(doc_id):
            deleted_count += 1
    
    print(f"   🗑️ Deleted {deleted_count} documents")
    print(f"   📊 Remaining documents: {db.count()}")
    
    print("\n✅ Unified API showcase completed successfully!")
    print("\n🎯 Key Features Demonstrated:")
    print("   • Schema-based document structure")
    print("   • Full-text search with BM25 scoring")
    print("   • Vector similarity search")
    print("   • Hybrid search combining text and vectors")
    print("   • Advanced filtering and faceted search")
    print("   • Document CRUD operations")
    print("   • Performance optimization")
    print("   • Production-ready error handling")
    
    print("\n📚 Next Steps:")
    print("   • Try with real embedding models (sentence-transformers, OpenAI)")
    print("   • Experiment with different schema designs")
    print("   • Build production RAG applications")
    print("   • Explore advanced query combinations")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"\n❌ Error: {e}")
        print("\nTroubleshooting:")
        print("1. Make sure VittoriaDB server is running: ./vittoriadb run")
        print("2. Install dependencies: pip install vittoriadb numpy")
        print("3. Check server logs for detailed error information")
        print("4. Verify the unified API is available at http://localhost:8080/unified")
