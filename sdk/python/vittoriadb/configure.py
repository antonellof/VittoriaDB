"""
VittoriaDB Configuration Module

This module provides configuration classes and utilities for VittoriaDB,
particularly for vectorizer configurations that enable automatic embedding generation.

Example usage:
    from vittoriadb.configure import Configure
    
    # Automatic embeddings with default settings
    config = Configure.Vectors.auto_embeddings()
    
    # OpenAI embeddings
    config = Configure.Vectors.openai_embeddings(api_key="your-key")
    
    # Custom sentence transformers
    config = Configure.Vectors.sentence_transformers(model="all-MiniLM-L6-v2")
"""

from typing import Dict, Any, Optional
from .types import VectorizerConfig, VectorizerType


class VectorConfiguration:
    """Configuration builder for vector operations and embeddings."""
    
    @staticmethod
    def auto_embeddings(
        model: str = "all-MiniLM-L6-v2",
        dimensions: int = 384,
        **options
    ) -> VectorizerConfig:
        """
        Configure automatic embeddings using the default vectorizer.
        
        This uses sentence-transformers by default but can fall back to
        other available vectorizers based on server configuration.
        
        Args:
            model: Model name to use (default: "all-MiniLM-L6-v2")
            dimensions: Vector dimensions (default: 384)
            **options: Additional options passed to the vectorizer
            
        Returns:
            VectorizerConfig: Configuration for automatic embeddings
        """
        return VectorizerConfig(
            type=VectorizerType.SENTENCE_TRANSFORMERS,
            model=model,
            dimensions=dimensions,
            options=options
        )
    
    @staticmethod
    def sentence_transformers(
        model: str = "all-MiniLM-L6-v2",
        dimensions: int = 384,
        **options
    ) -> VectorizerConfig:
        """
        Configure sentence transformers for embeddings.
        
        Args:
            model: Sentence transformers model name
            dimensions: Vector dimensions
            **options: Additional options
            
        Returns:
            VectorizerConfig: Configuration for sentence transformers
        """
        return VectorizerConfig(
            type=VectorizerType.SENTENCE_TRANSFORMERS,
            model=model,
            dimensions=dimensions,
            options=options
        )
    
    @staticmethod
    def openai_embeddings(
        api_key: str,
        model: str = "text-embedding-ada-002",
        dimensions: int = 1536,
        **options
    ) -> VectorizerConfig:
        """
        Configure OpenAI embeddings.
        
        Args:
            api_key: OpenAI API key
            model: OpenAI model name
            dimensions: Vector dimensions
            **options: Additional options
            
        Returns:
            VectorizerConfig: Configuration for OpenAI embeddings
        """
        config_options = {"api_key": api_key}
        config_options.update(options)
        
        return VectorizerConfig(
            type=VectorizerType.OPENAI,
            model=model,
            dimensions=dimensions,
            options=config_options
        )
    
    @staticmethod
    def huggingface_embeddings(
        api_key: Optional[str] = None,
        model: str = "sentence-transformers/all-MiniLM-L6-v2",
        dimensions: int = 384,
        **options
    ) -> VectorizerConfig:
        """
        Configure HuggingFace embeddings.
        
        Args:
            api_key: HuggingFace API token (optional for public models)
            model: HuggingFace model name
            dimensions: Vector dimensions
            **options: Additional options
            
        Returns:
            VectorizerConfig: Configuration for HuggingFace embeddings
        """
        config_options = {}
        if api_key:
            config_options["api_key"] = api_key
        config_options.update(options)
        
        return VectorizerConfig(
            type=VectorizerType.HUGGINGFACE,
            model=model,
            dimensions=dimensions,
            options=config_options
        )
    
    @staticmethod
    def ollama_embeddings(
        model: str = "nomic-embed-text",
        dimensions: int = 768,
        base_url: str = "http://localhost:11434",
        **options
    ) -> VectorizerConfig:
        """
        Configure Ollama embeddings for local inference.
        
        Args:
            model: Ollama model name
            dimensions: Vector dimensions
            base_url: Ollama server URL
            **options: Additional options
            
        Returns:
            VectorizerConfig: Configuration for Ollama embeddings
        """
        config_options = {"base_url": base_url}
        config_options.update(options)
        
        return VectorizerConfig(
            type=VectorizerType.OLLAMA,
            model=model,
            dimensions=dimensions,
            options=config_options
        )


class Configure:
    """Main configuration class for VittoriaDB."""
    
    # Vector configuration
    Vectors = VectorConfiguration
    
    @staticmethod
    def default_collection_config() -> Dict[str, Any]:
        """Get default collection configuration."""
        return {
            "metric": "cosine",
            "index_type": "hnsw",
            "config": {
                "m": 16,
                "ef_construction": 200,
                "ef_search": 50
            }
        }
    
    @staticmethod
    def performance_collection_config() -> Dict[str, Any]:
        """Get performance-optimized collection configuration."""
        return {
            "metric": "cosine", 
            "index_type": "hnsw",
            "config": {
                "m": 32,
                "ef_construction": 400,
                "ef_search": 100
            }
        }
    
    @staticmethod
    def memory_optimized_config() -> Dict[str, Any]:
        """Get memory-optimized collection configuration."""
        return {
            "metric": "cosine",
            "index_type": "flat",  # Use flat index for lower memory usage
            "config": {}
        }
