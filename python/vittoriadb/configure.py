"""
Configuration utilities for VittoriaDB, similar to Weaviate's Configure class.
"""

from .types import VectorizerConfig, VectorizerType


class Configure:
    """Configuration utilities for VittoriaDB collections."""
    
    class Vectors:
        """Vector configuration utilities."""
        
        @staticmethod
        def auto_embeddings(model: str = "nomic-embed-text", 
                           dimensions: int = None) -> VectorizerConfig:
            """
            Configure automatic text vectorization using local Ollama models.
            
            This provides high-quality embeddings using local ML models without
            external API dependencies or costs.
            
            Args:
                model: The local model name (default: "local-minilm-l6-v2")
                dimensions: Vector dimensions (default: 384)
            
            Returns:
                VectorizerConfig for local vectorizer
            
            Example:
                ```python
                import vittoriadb
                from vittoriadb.configure import Configure
                
                client = vittoriadb.connect()
                
                # Create collection with automatic embedding generation
                collection = client.create_collection(
                    name="Article",
                    dimensions=768,  # Will be auto-set based on model if not provided
                    vectorizer_config=Configure.Vectors.auto_embeddings()
                )
                
                # Insert text directly - embeddings generated automatically
                collection.insert_text("doc1", "Vector databases enable semantic search")
                collection.insert_text("doc2", "Machine learning models generate embeddings")
                
                # Search with text - query embedding generated automatically
                results = collection.search_text("Search objects by meaning", limit=1)
                ```
            
            Features:
                - Pure Go implementation (no Python subprocess)
                - No external dependencies
                - Fast local inference
                - Deterministic embeddings
                - Works offline
            """
            # Set default dimensions for Ollama models
            if dimensions is None:
                if model == "nomic-embed-text":
                    dimensions = 768
                else:
                    dimensions = 384
            
            return Configure.Vectors.ollama_embeddings(model=model, dimensions=dimensions)
        
        @staticmethod
        def sentence_transformers(model: str = "all-MiniLM-L6-v2", 
                                dimensions: int = None) -> VectorizerConfig:
            """
            Configure automatic text vectorization using sentence-transformers (Python subprocess).
            
            This approach calls Python subprocess with sentence-transformers library.
            Use this if you need compatibility with existing sentence-transformers models.
            
            Args:
                model: The sentence-transformers model name (default: "all-MiniLM-L6-v2")
                dimensions: Vector dimensions (auto-detected if None)
            
            Returns:
                VectorizerConfig for sentence-transformers
            
            Note:
                Requires Python with sentence-transformers installed on the server.
                For pure Go implementation without dependencies, use text2vec_model2vec().
            """
            # Auto-detect dimensions based on common models
            if dimensions is None:
                model_dimensions = {
                    "all-MiniLM-L6-v2": 384,
                    "all-mpnet-base-v2": 768,
                    "paraphrase-multilingual-MiniLM-L12-v2": 384,
                    "all-distilroberta-v1": 768,
                    "all-MiniLM-L12-v2": 384,
                    "multi-qa-MiniLM-L6-cos-v1": 384,
                    "multi-qa-mpnet-base-dot-v1": 768,
                    "paraphrase-MiniLM-L6-v2": 384,
                }
                dimensions = model_dimensions.get(model, 384)  # Default to 384
            
            return VectorizerConfig(
                type=VectorizerType.SENTENCE_TRANSFORMERS,
                model=model,
                dimensions=dimensions,
                options={}
            )

        @staticmethod
        def ollama_embeddings(model: str = "nomic-embed-text",
                            base_url: str = None,
                            dimensions: int = None) -> VectorizerConfig:
            """
            Configure local Ollama embedding models.
            
            This approach uses local Ollama models for high-quality embeddings
            without external API dependencies. Requires Ollama to be installed
            and running locally.
            
            Args:
                model: Ollama model name (e.g., "nomic-embed-text", "all-minilm")
                base_url: Ollama API base URL (defaults to http://localhost:11434)
                dimensions: Expected embedding dimensions (768 for nomic-embed-text)
            
            Returns:
                VectorizerConfig configured for local Ollama models
                
            Example:
                # Basic usage (requires: ollama pull nomic-embed-text)
                Configure.Vectors.ollama_embeddings()
                
                # Custom model
                Configure.Vectors.ollama_embeddings(
                    model="all-minilm",
                    dimensions=384
                )
            """
            if dimensions is None:
                if model == "nomic-embed-text":
                    dimensions = 768
                elif "minilm" in model.lower():
                    dimensions = 384
                else:
                    dimensions = 768  # Default
            
            options = {}
            if base_url:
                options["base_url"] = base_url
            else:
                options["base_url"] = "http://localhost:11434"
            
            return VectorizerConfig(
                type=VectorizerType.OLLAMA,
                model=model,
                dimensions=dimensions,
                options=options
            )

        @staticmethod
        def openai_embeddings(model: str = "text-embedding-ada-002",
                             api_key: str = None,
                             dimensions: int = None) -> VectorizerConfig:
            """
            Configure automatic text vectorization using OpenAI embeddings.
            
            Args:
                model: The OpenAI model name (default: "text-embedding-ada-002")
                api_key: OpenAI API key (required)
                dimensions: Vector dimensions (auto-detected if None)
            
            Returns:
                VectorizerConfig for OpenAI embeddings
            """
            if api_key is None:
                raise ValueError("OpenAI API key is required")
            
            # Auto-detect dimensions based on model
            if dimensions is None:
                model_dimensions = {
                    "text-embedding-ada-002": 1536,
                    "text-embedding-3-small": 1536,
                    "text-embedding-3-large": 3072,
                }
                dimensions = model_dimensions.get(model, 1536)
            
            return VectorizerConfig(
                type=VectorizerType.OPENAI,
                model=model,
                dimensions=dimensions,
                options={"api_key": api_key}
            )
        
        @staticmethod
        def huggingface_embeddings(model: str = "sentence-transformers/all-MiniLM-L6-v2",
                                  api_key: str = None,
                                  dimensions: int = None) -> VectorizerConfig:
            """
            Configure automatic text vectorization using HuggingFace models.
            
            Args:
                model: The HuggingFace model name
                api_key: HuggingFace API key (optional)
                dimensions: Vector dimensions (auto-detected if None)
            
            Returns:
                VectorizerConfig for HuggingFace embeddings
            """
            if dimensions is None:
                dimensions = 384  # Default for most sentence-transformer models
            
            options = {}
            if api_key:
                options["api_key"] = api_key
            
            return VectorizerConfig(
                type=VectorizerType.HUGGINGFACE,
                model=model,
                dimensions=dimensions,
                options=options
            )
        
        @staticmethod
        def self_provided() -> VectorizerConfig:
            """
            Configure collection to use self-provided embeddings (no automatic vectorization).
            
            Returns:
                VectorizerConfig indicating no automatic vectorization
            """
            return VectorizerConfig(
                type=VectorizerType.NONE,
                model="",
                dimensions=0,
                options={}
            )
