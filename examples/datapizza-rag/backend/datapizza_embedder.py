"""
Datapizza AI Embedder Service
Provides unified embedding generation using datapizza-ai library
Supports OpenAI, Ollama, and other embedding models
"""

import logging
import os
from typing import List, Union, Optional
from dataclasses import dataclass

# Import datapizza-ai embedders
from datapizza.embedders.openai import OpenAIEmbedder

logger = logging.getLogger(__name__)


@dataclass
class EmbedderConfig:
    """Configuration for embedder service"""
    provider: str = "openai"  # "openai" or "ollama"
    api_key: Optional[str] = None
    base_url: Optional[str] = None  # For local Ollama
    model_name: str = "text-embedding-ada-002"
    dimensions: int = 1536


class DatapizzaEmbedder:
    """
    Unified embedder using datapizza-ai library.
    Supports both OpenAI API and local Ollama models.
    """
    
    def __init__(self, config: Optional[EmbedderConfig] = None):
        """
        Initialize embedder with configuration.
        
        Args:
            config: EmbedderConfig instance. If None, loads from environment.
        """
        self.config = config or self._load_config_from_env()
        self.embedder = None
        self._initialize_embedder()
    
    def _load_config_from_env(self) -> EmbedderConfig:
        """Load configuration from environment variables"""
        provider = os.getenv("EMBEDDER_PROVIDER", "openai").lower()
        
        if provider == "ollama":
            # Ollama configuration
            return EmbedderConfig(
                provider="ollama",
                api_key="",  # Ollama doesn't require API key
                base_url=os.getenv("OLLAMA_BASE_URL", "http://localhost:11434/v1"),
                model_name=os.getenv("OLLAMA_EMBED_MODEL", "nomic-embed-text"),
                dimensions=int(os.getenv("OLLAMA_EMBED_DIMENSIONS", "768"))
            )
        else:
            # OpenAI configuration (default)
            return EmbedderConfig(
                provider="openai",
                api_key=os.getenv("OPENAI_API_KEY", ""),
                base_url=os.getenv("OPENAI_BASE_URL") if os.getenv("OPENAI_BASE_URL") else None,
                model_name=os.getenv("OPENAI_EMBED_MODEL", "text-embedding-ada-002"),
                dimensions=int(os.getenv("OPENAI_EMBED_DIMENSIONS", "1536"))
            )
    
    def _initialize_embedder(self):
        """Initialize the embedder using datapizza-ai library"""
        try:
            logger.info(f"🚀 Initializing Datapizza AI embedder with provider: {self.config.provider}")
            
            # Use datapizza-ai OpenAIEmbedder (works for both OpenAI and Ollama)
            embedder_kwargs = {
                "api_key": self.config.api_key or "not-needed"
            }
            
            if self.config.base_url:
                embedder_kwargs["base_url"] = self.config.base_url
                logger.info(f"📍 Using custom base URL: {self.config.base_url}")
            
            self.embedder = OpenAIEmbedder(**embedder_kwargs)
            
            logger.info(f"✅ Datapizza AI embedder initialized successfully")
            logger.info(f"   Provider: {self.config.provider}")
            logger.info(f"   Model: {self.config.model_name}")
            logger.info(f"   Dimensions: {self.config.dimensions}")
                
        except Exception as e:
            logger.error(f"❌ Failed to initialize Datapizza AI embedder: {e}")
            raise
    
    async def embed_text(self, text: Union[str, List[str]]) -> Union[List[float], List[List[float]]]:
        """
        Generate embeddings for text(s) using Datapizza AI.
        
        Args:
            text: Single text string or list of text strings
            
        Returns:
            Single embedding vector or list of embedding vectors
        """
        if not self.embedder:
            raise RuntimeError("Datapizza AI embedder not initialized")
        
        try:
            if isinstance(text, str):
                # Single text
                logger.debug(f"🔄 Generating embedding for text: {text[:100]}...")
                embedding = self.embedder.embed(text, model_name=self.config.model_name)
                logger.debug(f"✅ Generated embedding: {len(embedding)} dimensions")
                return embedding
            else:
                # Multiple texts
                logger.debug(f"🔄 Generating embeddings for {len(text)} texts...")
                embeddings = self.embedder.embed(text, model_name=self.config.model_name)
                logger.debug(f"✅ Generated {len(embeddings)} embeddings")
                return embeddings
                
        except Exception as e:
            logger.error(f"❌ Failed to generate embeddings with Datapizza AI: {e}")
            raise
    
    def embed_text_sync(self, text: Union[str, List[str]]) -> Union[List[float], List[List[float]]]:
        """
        Synchronous version of embed_text for compatibility.
        
        Args:
            text: Single text string or list of text strings
            
        Returns:
            Single embedding vector or list of embedding vectors
        """
        if not self.embedder:
            raise RuntimeError("Datapizza AI embedder not initialized")
        
        try:
            if isinstance(text, str):
                logger.debug(f"🔄 Generating embedding (sync) for text: {text[:100]}...")
                embedding = self.embedder.embed(text, model_name=self.config.model_name)
                logger.debug(f"✅ Generated embedding: {len(embedding)} dimensions")
                return embedding
            else:
                logger.debug(f"🔄 Generating embeddings (sync) for {len(text)} texts...")
                embeddings = self.embedder.embed(text, model_name=self.config.model_name)
                logger.debug(f"✅ Generated {len(embeddings)} embeddings")
                return embeddings
                
        except Exception as e:
            logger.error(f"❌ Failed to generate embeddings (sync) with Datapizza AI: {e}")
            raise
    
    def get_embedding_dimension(self) -> int:
        """Get the embedding dimension"""
        return self.config.dimensions
    
    def get_model_info(self) -> dict:
        """Get model information"""
        return {
            "provider": self.config.provider,
            "model_name": self.config.model_name,
            "dimensions": self.config.dimensions,
            "base_url": self.config.base_url,
            "library": "datapizza-ai"
        }


# Singleton instance
_embedder_instance = None


def get_embedder(config: Optional[EmbedderConfig] = None) -> DatapizzaEmbedder:
    """
    Get or create singleton embedder instance.
    
    Args:
        config: Optional configuration. Only used if creating new instance.
        
    Returns:
        DatapizzaEmbedder instance
    """
    global _embedder_instance
    
    if _embedder_instance is None:
        _embedder_instance = DatapizzaEmbedder(config)
    
    return _embedder_instance
