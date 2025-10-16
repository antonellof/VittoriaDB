"""
Datapizza AI + VittoriaDB RAG Pipeline
Complete RAG implementation using Datapizza's IngestionPipeline and DagPipeline
"""

import os
import logging
from typing import List, Dict, Any, Optional
from dataclasses import dataclass

from datapizza.clients.openai import OpenAIClient
from datapizza.core.vectorstore import VectorConfig
from datapizza.embedders import ChunkEmbedder
from datapizza.embedders.openai import OpenAIEmbedder
from datapizza.modules.splitters import NodeSplitter
from datapizza.modules.prompt import ChatPromptTemplate
from datapizza.modules.rewriters import ToolRewriter
from datapizza.pipeline import IngestionPipeline, DagPipeline

from vittoriadb_vectorstore import VittoriaDBVectorstore

logger = logging.getLogger(__name__)


@dataclass
class DatapizzaRAGConfig:
    """Configuration for Datapizza RAG pipeline"""
    openai_api_key: str
    vittoriadb_url: str = "http://localhost:8080"  # Default for local development
    embedding_model: str = "text-embedding-ada-002"
    embedding_dimensions: int = 1536
    llm_model: str = "gpt-4o-mini"
    chunk_size: int = 1000
    chunk_overlap: int = 200
    retrieval_k: int = 5


class DatapizzaRAGPipeline:
    """
    Complete RAG system using Datapizza AI pipelines with VittoriaDB.
    
    Features:
    - Document ingestion with automatic chunking and embedding
    - Query rewriting for improved retrieval
    - Semantic search with VittoriaDB
    - Prompt engineering with templates
    - LLM response generation
    """
    
    def __init__(self, config: DatapizzaRAGConfig):
        """Initialize the RAG pipeline"""
        self.config = config
        
        # Initialize VittoriaDB vectorstore
        self.vectorstore = VittoriaDBVectorstore(
            url=config.vittoriadb_url,
            auto_start=False
        )
        
        # Initialize OpenAI clients
        self.embedder = OpenAIEmbedder(
            api_key=config.openai_api_key,
            model_name=config.embedding_model,
        )
        
        self.llm_client = OpenAIClient(
            api_key=config.openai_api_key,
            model=config.llm_model
        )
        
        # Initialize query rewriter
        self.query_rewriter = ToolRewriter(
            client=self.llm_client,
            system_prompt="""You are a query optimization expert. 
Rewrite user queries to improve retrieval accuracy.
Make queries more specific and add relevant keywords.
Keep the original intent but make it more searchable."""
        )
        
        # Initialize prompt template
        self.prompt_template = ChatPromptTemplate(
            user_prompt_template="User question: {{user_prompt}}",
            retrieval_prompt_template="""Retrieved context:
{% for chunk in chunks %}
---
Score: {{ chunk.score }}
Content: {{ chunk.text }}
{% endfor %}

Based on the above context, answer the user's question accurately and comprehensively."""
        )
        
        logger.info("✅ Datapizza RAG Pipeline initialized")
    
    def create_collection(
        self,
        collection_name: str,
        replace_existing: bool = False
    ):
        """
        Create a collection in VittoriaDB.
        
        Args:
            collection_name: Name of the collection
            replace_existing: Whether to delete existing collection
        """
        try:
            if replace_existing:
                try:
                    self.vectorstore.delete_collection(collection_name)
                    logger.info(f"🗑️ Deleted existing collection '{collection_name}'")
                except:
                    pass
            
            self.vectorstore.create_collection(
                collection_name=collection_name,
                vector_config=[VectorConfig(name="embedding", dimensions=self.config.embedding_dimensions)]
            )
            
            logger.info(f"✅ Collection '{collection_name}' ready")
            
        except Exception as e:
            logger.error(f"❌ Failed to create collection: {e}")
            raise
    
    def create_ingestion_pipeline(self, collection_name: str) -> IngestionPipeline:
        """
        Create a Datapizza IngestionPipeline for document processing.
        
        Supports: PDF, DOCX, TXT, MD, HTML, and more via DoclingParser.
        
        Args:
            collection_name: Target collection name
            
        Returns:
            Configured IngestionPipeline
        """
        pipeline = IngestionPipeline(
            modules=[
                # Split text into chunks
                NodeSplitter(
                    max_char=self.config.chunk_size
                ),
                
                # Generate embeddings for each chunk
                ChunkEmbedder(client=self.embedder),
            ],
            vector_store=self.vectorstore,
            collection_name=collection_name
        )
        
        logger.info(f"✅ Ingestion pipeline created for '{collection_name}'")
        return pipeline
    
    def ingest_text(
        self,
        text: str,
        collection_name: str,
        metadata: Optional[Dict[str, Any]] = None
    ):
        """
        Ingest plain text into the collection.
        
        Args:
            text: Text content to ingest
            collection_name: Target collection
            metadata: Optional metadata to attach
        """
        try:
            pipeline = self.create_ingestion_pipeline(collection_name)
            
            # Create a temporary text file for DoclingParser
            import tempfile
            with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as f:
                f.write(text)
                temp_path = f.name
            
            try:
                pipeline.run(temp_path, metadata=metadata or {})
                logger.info(f"✅ Ingested text ({len(text)} chars) into '{collection_name}'")
            finally:
                # Clean up temp file
                os.unlink(temp_path)
                
        except Exception as e:
            logger.error(f"❌ Failed to ingest text: {e}")
            raise
    
    def ingest_file(
        self,
        file_path: str,
        collection_name: str,
        metadata: Optional[Dict[str, Any]] = None
    ):
        """
        Ingest a document file into the collection.
        
        Supports: PDF, DOCX, TXT, MD, HTML, and more via DoclingParser.
        
        Args:
            file_path: Path to the document (PDF, DOCX, TXT, etc.)
            collection_name: Target collection
            metadata: Optional metadata to attach
        """
        try:
            # Get file extension for logging
            file_ext = os.path.splitext(file_path)[1].upper()
            logger.info(f"📄 Processing {file_ext} file: {file_path}")
            
            pipeline = self.create_ingestion_pipeline(collection_name)
            pipeline.run(file_path, metadata=metadata or {})
            
            logger.info(f"✅ Ingested {file_ext} '{os.path.basename(file_path)}' into '{collection_name}'")
            
        except Exception as e:
            logger.error(f"❌ Failed to ingest file '{file_path}': {e}")
            raise
    
    def create_retrieval_pipeline(self, collection_name: str) -> DagPipeline:
        """
        Create a Datapizza DagPipeline for RAG retrieval.
        
        Args:
            collection_name: Collection to search
            
        Returns:
            Configured DagPipeline
        """
        dag_pipeline = DagPipeline()
        
        # Add modules
        dag_pipeline.add_module("rewriter", self.query_rewriter)
        dag_pipeline.add_module("embedder", self.embedder)
        dag_pipeline.add_module("retriever", self.vectorstore)
        dag_pipeline.add_module("prompt", self.prompt_template)
        dag_pipeline.add_module("generator", self.llm_client)
        
        # Connect modules (create the DAG)
        dag_pipeline.connect("rewriter", "embedder", target_key="text")
        dag_pipeline.connect("embedder", "retriever", target_key="query_vector")
        dag_pipeline.connect("retriever", "prompt", target_key="chunks")
        dag_pipeline.connect("prompt", "generator", target_key="memory")
        
        logger.info(f"✅ Retrieval pipeline created for '{collection_name}'")
        return dag_pipeline
    
    def query(
        self,
        question: str,
        collection_name: str,
        k: int = None,
        rewrite_query: bool = True
    ) -> Dict[str, Any]:
        """
        Query the RAG system.
        
        Args:
            question: User's question
            collection_name: Collection to search
            k: Number of results to retrieve (default: from config)
            rewrite_query: Whether to use query rewriting
            
        Returns:
            Dict with 'answer', 'chunks', 'rewritten_query'
        """
        try:
            k = k or self.config.retrieval_k
            
            # Create retrieval pipeline
            pipeline = self.create_retrieval_pipeline(collection_name)
            
            # Run the pipeline
            result = pipeline.run({
                "rewriter": {"user_prompt": question} if rewrite_query else {},
                "embedder": {"text": question} if not rewrite_query else {},
                "retriever": {
                    "collection_name": collection_name,
                    "k": k
                },
                "prompt": {"user_prompt": question},
                "generator": {"input": question}
            })
            
            # Extract results
            answer = result.get('generator', {})
            chunks = result.get('retriever', [])
            rewritten = result.get('rewriter', {}).get('rewritten_query') if rewrite_query else question
            
            logger.info(f"✅ Query completed: {len(chunks)} chunks retrieved")
            
            return {
                'answer': answer,
                'chunks': chunks,
                'rewritten_query': rewritten,
                'original_query': question
            }
            
        except Exception as e:
            logger.error(f"❌ Query failed: {e}")
            raise
    
    async def query_stream(
        self,
        question: str,
        collection_name: str,
        k: int = None,
        rewrite_query: bool = True
    ):
        """
        Query the RAG system with streaming response.
        
        Args:
            question: User's question
            collection_name: Collection to search
            k: Number of results to retrieve
            rewrite_query: Whether to use query rewriting
            
        Yields:
            Chunks of the streaming response
        """
        try:
            k = k or self.config.retrieval_k
            
            # Step 1: Rewrite query (if enabled)
            if rewrite_query:
                rewritten = await self.query_rewriter.a_invoke(user_prompt=question)
                query_text = rewritten.get('rewritten_query', question)
                yield {'type': 'query_rewritten', 'query': query_text}
            else:
                query_text = question
            
            # Step 2: Embed query
            query_embedding = self.embedder.embed(text=query_text)
            yield {'type': 'query_embedded'}
            
            # Step 3: Retrieve chunks
            chunks = self.vectorstore.search(
                collection_name=collection_name,
                query_vector=query_embedding,
                k=k
            )
            yield {'type': 'chunks_retrieved', 'count': len(chunks), 'chunks': chunks}
            
            # Step 4: Build prompt
            prompt_result = self.prompt_template.invoke(
                user_prompt=question,
                chunks=chunks
            )
            
            # Step 5: Stream LLM response
            yield {'type': 'llm_start'}
            
            async for chunk in self.llm_client.a_stream_invoke(
                input=prompt_result['memory'],
                temperature=0.7,
                max_tokens=1500
            ):
                if chunk.delta:
                    yield {'type': 'content', 'content': chunk.delta}
            
            yield {'type': 'done'}
            
        except Exception as e:
            logger.error(f"❌ Streaming query failed: {e}")
            yield {'type': 'error', 'message': str(e)}


# Factory function for easy initialization
def create_datapizza_rag_pipeline(
    openai_api_key: Optional[str] = None,
    vittoriadb_url: Optional[str] = None
) -> DatapizzaRAGPipeline:
    """
    Create a DatapizzaRAGPipeline with default configuration.
    
    Args:
        openai_api_key: OpenAI API key (defaults to OPENAI_API_KEY env var)
        vittoriadb_url: VittoriaDB server URL (defaults to VITTORIADB_URL env var or localhost:8080)
        
    Returns:
        Configured DatapizzaRAGPipeline
    """
    config = DatapizzaRAGConfig(
        openai_api_key=openai_api_key or os.getenv('OPENAI_API_KEY'),
        vittoriadb_url=vittoriadb_url or os.getenv('VITTORIADB_URL', 'http://localhost:8080'),
        embedding_model=os.getenv('OPENAI_EMBED_MODEL', 'text-embedding-ada-002'),
        embedding_dimensions=int(os.getenv('OPENAI_EMBED_DIMENSIONS', '1536')),
        llm_model=os.getenv('LLM_MODEL', 'gpt-4o-mini'),
        chunk_size=int(os.getenv('CHUNK_SIZE', '1000')),
        chunk_overlap=int(os.getenv('CHUNK_OVERLAP', '200')),
        retrieval_k=int(os.getenv('RETRIEVAL_K', '5'))
    )
    
    return DatapizzaRAGPipeline(config)

