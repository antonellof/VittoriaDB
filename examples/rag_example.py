#!/usr/bin/env python3
"""
RAG (Retrieval-Augmented Generation) example with VittoriaDB.

This example demonstrates how to build a simple RAG system using VittoriaDB
for document storage and retrieval, combined with embedding models.

Requirements:
    pip install vittoriadb sentence-transformers

Run:
    python examples/rag_example.py
"""

import vittoriadb
import time
from typing import List, Dict, Any

# Optional: Use sentence-transformers for real embeddings
try:
    from sentence_transformers import SentenceTransformer
    HAS_SENTENCE_TRANSFORMERS = True
except ImportError:
    HAS_SENTENCE_TRANSFORMERS = False
    print("Note: sentence-transformers not installed. Using random embeddings.")
    print("Install with: pip install sentence-transformers")


class SimpleRAG:
    """Simple RAG system using VittoriaDB."""
    
    def __init__(self, collection_name: str = "knowledge_base"):
        """Initialize RAG system."""
        self.collection_name = collection_name
        self.db = None
        self.collection = None
        self.model = None
        
        # Initialize embedding model
        if HAS_SENTENCE_TRANSFORMERS:
            print("Loading sentence transformer model...")
            self.model = SentenceTransformer('all-MiniLM-L6-v2')
            self.dimensions = 384
        else:
            print("Using random embeddings (install sentence-transformers for real embeddings)")
            self.dimensions = 384
    
    def connect(self):
        """Connect to VittoriaDB and create collection."""
        print("Connecting to VittoriaDB...")
        self.db = vittoriadb.connect()
        
        # Create or get collection
        try:
            self.collection = self.db.get_collection(self.collection_name)
            print(f"Using existing collection: {self.collection_name}")
        except:
            self.collection = self.db.create_collection(
                name=self.collection_name,
                dimensions=self.dimensions,
                metric="cosine"
            )
            print(f"Created new collection: {self.collection_name}")
    
    def embed_text(self, text: str) -> List[float]:
        """Generate embedding for text."""
        if self.model:
            return self.model.encode(text).tolist()
        else:
            # Fallback to random embeddings
            import numpy as np
            return np.random.random(self.dimensions).tolist()
    
    def add_documents(self, documents: List[Dict[str, Any]]):
        """Add documents to the knowledge base."""
        print(f"Adding {len(documents)} documents...")
        
        vectors = []
        for i, doc in enumerate(documents):
            # Generate embedding
            embedding = self.embed_text(doc["content"])
            
            # Create vector with metadata
            vectors.append({
                "id": doc.get("id", f"doc_{i}"),
                "vector": embedding,
                "metadata": {
                    "content": doc["content"],
                    "title": doc.get("title", f"Document {i}"),
                    "source": doc.get("source", "unknown"),
                    "category": doc.get("category", "general"),
                    "length": len(doc["content"])
                }
            })
        
        # Batch insert
        result = self.collection.insert_batch(vectors)
        print(f"Inserted {result['inserted']} documents")
        
        return result
    
    def search(self, query: str, limit: int = 5, min_score: float = 0.0) -> List[Dict[str, Any]]:
        """Search for relevant documents."""
        # Generate query embedding
        query_embedding = self.embed_text(query)
        
        # Search in VittoriaDB
        results = self.collection.search(
            vector=query_embedding,
            limit=limit,
            include_metadata=True
        )
        
        # Filter by minimum score and format results
        relevant_docs = []
        for result in results:
            if result.score >= min_score:
                relevant_docs.append({
                    "id": result.id,
                    "score": result.score,
                    "title": result.metadata.get("title", "Untitled"),
                    "content": result.metadata.get("content", ""),
                    "source": result.metadata.get("source", "unknown"),
                    "category": result.metadata.get("category", "general")
                })
        
        return relevant_docs
    
    def ask(self, question: str, context_limit: int = 3) -> Dict[str, Any]:
        """Ask a question and get context from the knowledge base."""
        print(f"\n🤔 Question: {question}")
        
        # Search for relevant context
        start_time = time.time()
        relevant_docs = self.search(question, limit=context_limit)
        search_time = time.time() - start_time
        
        if not relevant_docs:
            return {
                "question": question,
                "answer": "I couldn't find any relevant information in the knowledge base.",
                "context": [],
                "search_time": search_time
            }
        
        # Build context
        context = []
        for doc in relevant_docs:
            context.append({
                "title": doc["title"],
                "content": doc["content"][:200] + "..." if len(doc["content"]) > 200 else doc["content"],
                "score": doc["score"],
                "source": doc["source"]
            })
        
        # Simple answer generation (in a real system, you'd use an LLM here)
        answer = self._generate_simple_answer(question, relevant_docs)
        
        return {
            "question": question,
            "answer": answer,
            "context": context,
            "search_time": search_time,
            "num_results": len(relevant_docs)
        }
    
    def _generate_simple_answer(self, question: str, docs: List[Dict[str, Any]]) -> str:
        """Generate a simple answer based on retrieved documents."""
        # This is a very simple answer generation
        # In a real RAG system, you'd use an LLM like GPT-3.5/4, Claude, etc.
        
        if not docs:
            return "No relevant information found."
        
        best_doc = docs[0]  # Highest scoring document
        
        return f"""Based on the most relevant document "{best_doc['title']}" (similarity: {best_doc['score']:.3f}):

{best_doc['content'][:300]}{'...' if len(best_doc['content']) > 300 else ''}

Source: {best_doc['source']}

Note: This is a simple retrieval-based answer. In a production RAG system, 
this context would be sent to an LLM for better answer generation."""
    
    def stats(self):
        """Show knowledge base statistics."""
        count = self.collection.count()
        info = self.collection.info
        
        print(f"\n📊 Knowledge Base Stats:")
        print(f"   Documents: {count}")
        print(f"   Dimensions: {info.dimensions}")
        print(f"   Metric: {info.metric.value}")
        print(f"   Created: {info.created}")
    
    def close(self):
        """Close the connection."""
        if self.db:
            self.db.close()


def main():
    """Main RAG example."""
    print("🧠 VittoriaDB RAG Example")
    print("=" * 30)
    
    # Initialize RAG system
    rag = SimpleRAG("rag_knowledge_base")
    rag.connect()
    
    # Sample knowledge base documents
    documents = [
        {
            "id": "python_intro",
            "title": "Introduction to Python",
            "content": "Python is a high-level, interpreted programming language known for its simplicity and readability. It was created by Guido van Rossum and first released in 1991. Python supports multiple programming paradigms including procedural, object-oriented, and functional programming.",
            "source": "Python Documentation",
            "category": "programming"
        },
        {
            "id": "vector_db_basics",
            "title": "Vector Database Fundamentals",
            "content": "Vector databases are specialized databases designed to store and query high-dimensional vectors efficiently. They use similarity search algorithms like HNSW, IVF, or LSH to find similar vectors quickly. Vector databases are essential for AI applications like semantic search, recommendation systems, and RAG.",
            "source": "AI Database Guide",
            "category": "database"
        },
        {
            "id": "machine_learning",
            "title": "Machine Learning Overview",
            "content": "Machine learning is a subset of artificial intelligence that enables computers to learn and improve from experience without being explicitly programmed. It includes supervised learning, unsupervised learning, and reinforcement learning. Common algorithms include linear regression, decision trees, neural networks, and support vector machines.",
            "source": "ML Textbook",
            "category": "ai"
        },
        {
            "id": "embeddings_explained",
            "title": "Understanding Embeddings",
            "content": "Embeddings are dense vector representations of data that capture semantic meaning. Text embeddings convert words, sentences, or documents into numerical vectors where similar items have similar vectors. Popular embedding models include Word2Vec, GloVe, BERT, and sentence transformers.",
            "source": "NLP Research Paper",
            "category": "ai"
        },
        {
            "id": "rag_systems",
            "title": "Retrieval-Augmented Generation",
            "content": "RAG combines retrieval systems with generative models to provide more accurate and contextual responses. It works by first retrieving relevant documents from a knowledge base, then using those documents as context for generating answers. RAG helps reduce hallucinations in large language models.",
            "source": "RAG Research Paper",
            "category": "ai"
        }
    ]
    
    # Add documents to knowledge base
    rag.add_documents(documents)
    rag.stats()
    
    # Example questions
    questions = [
        "What is Python?",
        "How do vector databases work?",
        "What are embeddings?",
        "Explain RAG systems",
        "What is machine learning?",
        "How to build a search system?"
    ]
    
    print("\n" + "=" * 50)
    print("🔍 Question & Answer Session")
    print("=" * 50)
    
    # Ask questions
    for question in questions:
        result = rag.ask(question)
        
        print(f"\n📝 Answer (found {result['num_results']} relevant docs in {result['search_time']:.3f}s):")
        print(result["answer"])
        
        print(f"\n📚 Context used:")
        for i, ctx in enumerate(result["context"], 1):
            print(f"   {i}. {ctx['title']} (score: {ctx['score']:.3f})")
            print(f"      {ctx['content']}")
        
        print("\n" + "-" * 50)
    
    # Interactive mode
    print("\n🎯 Interactive Mode (type 'quit' to exit)")
    print("-" * 40)
    
    while True:
        try:
            question = input("\nYour question: ").strip()
            if question.lower() in ['quit', 'exit', 'q']:
                break
            
            if question:
                result = rag.ask(question)
                print(f"\n💡 Answer:")
                print(result["answer"])
                
                if result["context"]:
                    print(f"\n📖 Sources:")
                    for ctx in result["context"]:
                        print(f"   • {ctx['title']} (relevance: {ctx['score']:.3f})")
        
        except KeyboardInterrupt:
            break
    
    # Cleanup
    print("\n🧹 Cleaning up...")
    rag.close()
    print("✅ RAG example completed!")
    
    print("\n🚀 Next Steps:")
    print("- Try with real documents: upload PDFs, text files")
    print("- Integrate with OpenAI GPT for better answer generation")
    print("- Add more sophisticated retrieval strategies")
    print("- Implement conversation memory")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"\n❌ Error: {e}")
        print("\nTroubleshooting:")
        print("1. Make sure VittoriaDB is installed: pip install vittoriadb")
        print("2. For better embeddings: pip install sentence-transformers")
        print("3. Check if port 8080 is available")
