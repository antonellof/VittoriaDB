"""
Example: Using Datapizza AI + VittoriaDB RAG Pipeline
Demonstrates the complete RAG workflow with Datapizza's pipeline architecture
"""

import asyncio
import os
from datapizza_rag_pipeline import create_datapizza_rag_pipeline


async def main():
    """
    Complete RAG example using Datapizza AI pipelines with VittoriaDB
    """
    
    print("🍕 Datapizza AI + VittoriaDB RAG Pipeline Example\n")
    
    # Step 1: Initialize the pipeline
    print("Step 1: Initializing RAG pipeline...")
    rag = create_datapizza_rag_pipeline(
        openai_api_key=os.getenv('OPENAI_API_KEY'),
        vittoriadb_url=os.getenv('VITTORIADB_URL', 'http://vittoriadb:8080')
    )
    print("✅ Pipeline initialized\n")
    
    # Step 2: Create a collection
    print("Step 2: Creating collection...")
    collection_name = "datapizza_demo"
    rag.create_collection(collection_name, replace_existing=True)
    print(f"✅ Collection '{collection_name}' created\n")
    
    # Step 3: Ingest some documents
    print("Step 3: Ingesting documents...")
    
    documents = [
        {
            "text": """
            Datapizza AI is a modern Python framework for building Gen AI solutions.
            It provides tools for embeddings, LLM clients, RAG pipelines, and more.
            The framework is designed to be modular, extensible, and production-ready.
            """,
            "metadata": {"source": "datapizza_docs", "title": "What is Datapizza AI?"}
        },
        {
            "text": """
            VittoriaDB is a zero-configuration embedded vector database with HNSW indexing.
            It features ACID storage, REST API, and is perfect for local AI development.
            VittoriaDB can be deployed as a single Go binary.
            """,
            "metadata": {"source": "vittoriadb_docs", "title": "What is VittoriaDB?"}
        },
        {
            "text": """
            RAG (Retrieval-Augmented Generation) combines semantic search with LLM generation.
            The process involves: 1) Embedding documents, 2) Storing in vector DB,
            3) Retrieving relevant chunks, 4) Generating answers with LLM.
            Datapizza AI provides IngestionPipeline and DagPipeline for RAG workflows.
            """,
            "metadata": {"source": "rag_guide", "title": "RAG Explained"}
        }
    ]
    
    for i, doc in enumerate(documents):
        rag.ingest_text(
            text=doc["text"],
            collection_name=collection_name,
            metadata=doc["metadata"]
        )
        print(f"✅ Ingested document {i+1}/{len(documents)}: {doc['metadata']['title']}")
    
    print(f"\n✅ All {len(documents)} documents ingested\n")
    
    # Step 4: Query the RAG system (non-streaming)
    print("Step 4: Querying RAG system (non-streaming)...")
    question = "What is Datapizza AI and how does it relate to RAG?"
    
    print(f"❓ Question: {question}\n")
    
    result = rag.query(
        question=question,
        collection_name=collection_name,
        k=3,
        rewrite_query=True
    )
    
    print(f"🔄 Rewritten query: {result['rewritten_query']}")
    print(f"📚 Retrieved {len(result['chunks'])} chunks")
    print(f"\n💬 Answer:\n{result['answer']}\n")
    
    # Step 5: Query with streaming
    print("Step 5: Querying RAG system (streaming)...")
    question2 = "How do I use VittoriaDB for vector storage?"
    print(f"❓ Question: {question2}\n")
    
    print("💬 Streaming answer: ", end="", flush=True)
    
    async for chunk in rag.query_stream(
        question=question2,
        collection_name=collection_name,
        k=3,
        rewrite_query=True
    ):
        if chunk['type'] == 'query_rewritten':
            print(f"\n🔄 Rewritten: {chunk['query']}\n")
            print("💬 Answer: ", end="", flush=True)
        elif chunk['type'] == 'chunks_retrieved':
            print(f"\n📚 Retrieved {chunk['count']} chunks\n")
            print("💬 Answer: ", end="", flush=True)
        elif chunk['type'] == 'content':
            print(chunk['content'], end="", flush=True)
        elif chunk['type'] == 'done':
            print("\n\n✅ Streaming complete")
        elif chunk['type'] == 'error':
            print(f"\n❌ Error: {chunk['message']}")
    
    print("\n🎉 Datapizza AI + VittoriaDB RAG Pipeline Demo Complete!")


if __name__ == "__main__":
    asyncio.run(main())

