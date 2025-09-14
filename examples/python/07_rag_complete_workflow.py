#!/usr/bin/env python3
"""
Complete RAG (Retrieval-Augmented Generation) Example with VittoriaDB

This example demonstrates a full RAG pipeline using VittoriaDB:
1. Document ingestion and processing
2. Vector embedding generation
3. Semantic search and retrieval
4. Context-aware response generation

Requirements:
    pip install vittoriadb sentence-transformers openai requests
    
Usage:
    python 07_rag_complete_workflow.py
"""

import time
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
from sentence_transformers import SentenceTransformer
import vittoriadb

# Optional OpenAI import for advanced response generation
try:
    import openai
    # Configure OpenAI (optional - for response generation)
    # openai.api_key = os.getenv('OPENAI_API_KEY')
    HAS_OPENAI = True
except ImportError:
    HAS_OPENAI = False

@dataclass
class Document:
    """Represents a document in our knowledge base."""
    id: str
    title: str
    content: str
    metadata: Dict[str, Any]
    chunks: List[str]

# Use the centralized SearchResult from vittoriadb
SearchResult = vittoriadb.SearchResult

class RAGSystem:
    """Complete RAG system using VittoriaDB."""
    
    def __init__(self, collection_name: str = "knowledge_base"):
        self.collection_name = collection_name
        self.db = None
        self.collection = None
        self.embedding_model = None
        self.documents: Dict[str, Document] = {}
        
        # Initialize embedding model
        print("🔄 Loading embedding model...")
        self.embedding_model = SentenceTransformer('all-MiniLM-L6-v2')
        self.embedding_dim = self.embedding_model.get_sentence_embedding_dimension()
        print(f"✅ Loaded model with {self.embedding_dim} dimensions")
    
    def setup(self) -> bool:
        """Setup the RAG system."""
        print("🔄 Setting up RAG system...")
        
        try:
            # Connect to existing VittoriaDB server (don't auto-start)
            self.db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)
            
            # Create or get collection
            try:
                self.collection = self.db.create_collection(
                    name=self.collection_name,
                    dimensions=self.embedding_dim,
                    metric=vittoriadb.DistanceMetric.COSINE
                )
                print(f"✅ Collection '{self.collection_name}' created")
            except vittoriadb.CollectionError as e:
                if "already exists" in str(e):
                    self.collection = self.db.get_collection(self.collection_name)
                    print(f"ℹ️  Collection '{self.collection_name}' already exists")
                else:
                    raise
            
            return True
            
        except Exception as e:
            print(f"❌ Failed to setup VittoriaDB: {e}")
            return False
    
    def add_document(self, doc_id: str, title: str, content: str, 
                    metadata: Optional[Dict[str, Any]] = None, 
                    overwrite: bool = False) -> bool:
        """Add a document to the knowledge base."""
        if metadata is None:
            metadata = {}
        
        # Chunk the document (simple sentence-based chunking)
        chunks = self._chunk_text(content, max_chunk_size=500)
        
        # Create document
        doc = Document(
            id=doc_id,
            title=title,
            content=content,
            metadata=metadata,
            chunks=chunks
        )
        
        # Check if document already exists
        if doc_id in self.documents and not overwrite:
            print(f"ℹ️  Document '{title}' already exists. Use overwrite=True to replace it.")
            return True
        
        # Generate embeddings and store chunks
        print(f"🔄 Processing document: {title}")
        successful_chunks = 0
        for i, chunk in enumerate(chunks):
            chunk_id = f"{doc_id}_chunk_{i}"
            
            # Generate embedding
            embedding = self.embedding_model.encode(chunk).tolist()
            
            # Prepare metadata
            chunk_metadata = {
                "document_id": doc_id,
                "document_title": title,
                "chunk_index": i,
                "content": chunk,
                "chunk_size": len(chunk),
                **metadata
            }
            
            # Insert into VittoriaDB
            try:
                self.collection.insert(
                    id=chunk_id,
                    vector=embedding,
                    metadata=chunk_metadata
                )
                successful_chunks += 1
            except Exception as e:
                if "already exists" in str(e).lower() or "conflict" in str(e).lower():
                    print(f"ℹ️  Chunk {chunk_id} already exists, skipping...")
                    successful_chunks += 1  # Count as successful since it exists
                else:
                    print(f"❌ Failed to insert chunk {chunk_id}: {e}")
                    # Continue with other chunks instead of failing completely
        
        self.documents[doc_id] = doc
        print(f"✅ Added document '{title}' with {successful_chunks}/{len(chunks)} chunks successfully processed")
        return successful_chunks > 0
    
    def search(self, query: str, limit: int = 5) -> List[SearchResult]:
        """Search the knowledge base."""
        print(f"🔍 Searching for: {query}")
        
        # Generate query embedding
        query_embedding = self.embedding_model.encode(query).tolist()
        
        # Search VittoriaDB
        results = self.collection.search(
            vector=query_embedding,
            limit=limit,
            include_metadata=True
        )
        
        print(f"📊 Found {len(results)} results")
        return results
    
    def generate_response(self, query: str, context_limit: int = 3) -> str:
        """Generate a response using retrieved context."""
        # Retrieve relevant context
        search_results = self.search(query, limit=context_limit)
        
        if not search_results:
            return "I don't have enough information to answer that question."
        
        # Build context from search results
        context_parts = []
        for i, result in enumerate(search_results):
            context_parts.append(f"Context {i+1} (score: {result.score:.3f}):")
            content = result.metadata.get("content", "") if result.metadata else ""
            context_parts.append(content)
            context_parts.append("")
        
        context = "\n".join(context_parts)
        
        # Simple response generation (without OpenAI)
        response = self._generate_simple_response(query, search_results)
        
        return response
    
    def _chunk_text(self, text: str, max_chunk_size: int = 500) -> List[str]:
        """Simple text chunking by sentences."""
        import re
        
        # Split into sentences
        sentences = re.split(r'[.!?]+', text)
        sentences = [s.strip() for s in sentences if s.strip()]
        
        chunks = []
        current_chunk = []
        current_size = 0
        
        for sentence in sentences:
            sentence_size = len(sentence)
            
            if current_size + sentence_size > max_chunk_size and current_chunk:
                # Finalize current chunk
                chunks.append(". ".join(current_chunk) + ".")
                current_chunk = [sentence]
                current_size = sentence_size
            else:
                current_chunk.append(sentence)
                current_size += sentence_size
        
        # Add final chunk
        if current_chunk:
            chunks.append(". ".join(current_chunk) + ".")
        
        return chunks
    
    def _generate_simple_response(self, query: str, results: List[SearchResult]) -> str:
        """Generate a simple response based on search results."""
        if not results:
            return "I don't have information about that topic."
        
        # Extract key information from top results
        top_result = results[0]
        
        top_content = top_result.metadata.get("content", "") if top_result.metadata else ""
        response_parts = [
            f"Based on the available information (confidence: {top_result.score:.1%}):",
            "",
            top_content,
        ]
        
        # Add additional context if available
        if len(results) > 1:
            response_parts.extend([
                "",
                "Additional relevant information:",
            ])
            for result in results[1:3]:  # Add up to 2 more results
                content = result.metadata.get("content", "") if result.metadata else ""
                response_parts.append(f"• {content[:200]}...")
        
        return "\n".join(response_parts)
    
    def close(self):
        """Close the database connection."""
        if self.db:
            self.db.close()

def create_sample_knowledge_base(rag: RAGSystem):
    """Create a sample knowledge base about VittoriaDB."""
    
    documents = [
        {
            "id": "vittoriadb_overview",
            "title": "VittoriaDB Overview",
            "content": """
            VittoriaDB is a high-performance, embedded vector database designed specifically for local AI development and production deployments. Built with simplicity and performance in mind, it provides a zero-configuration solution for vector similarity search, making it perfect for RAG applications, semantic search, recommendation systems, and AI prototyping.

            The database bridges the gap between complex cloud-based vector databases and simple in-memory solutions. It offers the performance and features of enterprise vector databases while maintaining the simplicity and portability of embedded databases.
            """,
            "metadata": {"category": "overview", "source": "documentation"}
        },
        {
            "id": "vittoriadb_features",
            "title": "VittoriaDB Key Features",
            "content": """
            VittoriaDB offers several key features that make it ideal for AI applications:

            Zero Configuration: Works immediately after installation with sensible defaults.
            High Performance: HNSW indexing provides sub-millisecond search times for millions of vectors.
            Persistent Storage: ACID-compliant file-based storage with Write-Ahead Log for durability.
            Dual Interface: REST API for universal access and native Python client with auto-binary management.
            Document Processing: Built-in support for PDF, DOCX, TXT, MD, and HTML files with intelligent chunking.
            AI-Ready: Seamless integration with embedding models from Hugging Face and OpenAI.
            """,
            "metadata": {"category": "features", "source": "documentation"}
        },
        {
            "id": "vittoriadb_architecture",
            "title": "VittoriaDB Architecture",
            "content": """
            VittoriaDB uses a modular architecture with separate layers:

            HTTP API Server: Provides RESTful endpoints for all database operations, including document upload and processing.
            Vector Engine: Implements both flat (exact) and HNSW (approximate) indexing with support for multiple distance metrics including cosine, euclidean, dot product, and Manhattan.
            Storage Layer: File-based persistence with page management, WAL for durability, LRU caching, and transaction support.
            Document Processor: Handles text extraction from various formats with intelligent chunking strategies.

            All data is stored locally in configurable directories, making it perfect for development, testing, and edge deployments.
            """,
            "metadata": {"category": "architecture", "source": "documentation"}
        },
        {
            "id": "vittoriadb_use_cases",
            "title": "VittoriaDB Use Cases",
            "content": """
            VittoriaDB is perfect for various AI and machine learning applications:

            RAG Applications: Retrieval-Augmented Generation systems that need to find relevant context from large document collections.
            Semantic Search: Find documents based on meaning rather than exact keyword matches.
            Recommendation Systems: Content and product recommendations based on similarity.
            AI Prototyping: Rapid development and testing of AI applications without complex infrastructure.
            Edge Computing: Local processing without cloud dependencies for privacy and performance.
            Knowledge Management: Organize and search through company documents and knowledge bases.
            """,
            "metadata": {"category": "use_cases", "source": "documentation"}
        },
        {
            "id": "vittoriadb_performance",
            "title": "VittoriaDB Performance",
            "content": """
            VittoriaDB delivers excellent performance across different metrics:

            Insert Speed: Over 10,000 vectors per second with flat indexing, over 5,000 with HNSW.
            Search Speed: Sub-millisecond search times for 1 million vectors using HNSW indexing.
            Memory Usage: Less than 100MB for 100,000 vectors with 384 dimensions.
            Startup Time: Cold start in under 100ms, warm start in under 50ms.
            Binary Size: Approximately 8MB compressed, 25MB uncompressed.
            Scalability: Tested up to 1 million vectors, supports up to 2,048 dimensions.

            The database includes SIMD optimizations for distance calculations and efficient Go garbage collection tuning.
            """,
            "metadata": {"category": "performance", "source": "benchmarks"}
        }
    ]
    
    print("📚 Creating sample knowledge base...")
    for doc in documents:
        rag.add_document(
            doc_id=doc["id"],
            title=doc["title"],
            content=doc["content"].strip(),
            metadata=doc["metadata"]
        )
    
    print("✅ Knowledge base created successfully!")

def interactive_demo(rag: RAGSystem):
    """Run an interactive demo of the RAG system."""
    print("\n" + "="*60)
    print("🤖 VittoriaDB RAG System - Interactive Demo")
    print("="*60)
    print("Ask questions about VittoriaDB. Type 'quit' to exit.")
    print("Example questions:")
    print("  - What is VittoriaDB?")
    print("  - How fast is VittoriaDB?")
    print("  - What are the main features?")
    print("  - How does the architecture work?")
    print("-"*60)
    
    while True:
        try:
            query = input("\n❓ Your question: ").strip()
            
            if query.lower() in ['quit', 'exit', 'q']:
                print("👋 Goodbye!")
                break
            
            if not query:
                continue
            
            print("\n🔄 Processing...")
            start_time = time.time()
            
            response = rag.generate_response(query)
            
            end_time = time.time()
            
            print(f"\n🤖 Response (took {end_time - start_time:.2f}s):")
            print("-" * 50)
            print(response)
            print("-" * 50)
            
        except KeyboardInterrupt:
            print("\n👋 Goodbye!")
            break
        except Exception as e:
            print(f"❌ Error: {e}")

def main():
    """Main function to run the RAG example."""
    print("🚀 VittoriaDB RAG Example")
    print("=" * 40)
    
    # Initialize RAG system
    rag = RAGSystem(collection_name="vittoriadb_docs")
    
    # Setup
    if not rag.setup():
        print("❌ Failed to setup RAG system")
        return
    
    # Create sample knowledge base
    create_sample_knowledge_base(rag)
    
    # Run some example queries
    print("\n📋 Running example queries...")
    
    example_queries = [
        "What is VittoriaDB?",
        "How fast is VittoriaDB for searching?",
        "What file formats does VittoriaDB support?",
        "How does VittoriaDB architecture work?"
    ]
    
    for query in example_queries:
        print(f"\n❓ Query: {query}")
        response = rag.generate_response(query, context_limit=2)
        print(f"🤖 Response: {response[:200]}...")
        print("-" * 40)
    
    # Interactive demo
    try:
        interactive_demo(rag)
    except KeyboardInterrupt:
        print("\n👋 Demo ended")
    finally:
        # Clean up
        rag.close()

if __name__ == "__main__":
    main()
