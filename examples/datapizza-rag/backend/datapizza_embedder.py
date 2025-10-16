"""
Datapizza AI Embedder Service
Provides unified embedding generation using datapizza-ai library
Supports OpenAI, Ollama, and other embedding models
"""

import logging
import os
from typing import List, Union, Optional
from dataclasses import dataclass

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
                base_url=os.getenv("OPENAI_BASE_URL", None),
                model_name=os.getenv("OPENAI_EMBED_MODEL", "text-embedding-ada-002"),
                dimensions=int(os.getenv("OPENAI_EMBED_DIMENSIONS", "1536"))
            )
    
    def _initialize_embedder(self):
        """Initialize the embedder using OpenAI API (datapizza-compatible)"""
        try:
            # Use OpenAI client directly (compatible with datapizza patterns)
            from openai import OpenAI
            
            logger.info(f"🚀 Initializing embedder with provider: {self.config.provider}")
            
            # Create OpenAI client with custom base_url for Ollama support
            client_kwargs = {"api_key": self.config.api_key or "not-needed"}
            
            if self.config.base_url:
                client_kwargs["base_url"] = self.config.base_url
                logger.info(f"📍 Using custom base URL: {self.config.base_url}")
            
            self.embedder = OpenAI(**client_kwargs)
            
            logger.info(f"✅ Embedder initialized successfully (datapizza-compatible)")
            logger.info(f"   Provider: {self.config.provider}")
            logger.info(f"   Model: {self.config.model_name}")
            logger.info(f"   Dimensions: {self.config.dimensions}")
            
        except Exception as e:
            logger.error(f"❌ Failed to initialize embedder: {e}")
            raise
    
    async def embed_text(self, text: Union[str, List[str]]) -> Union[List[float], List[List[float]]]:
        """
        Generate embeddings for text(s) - datapizza-compatible implementation.
        
        Args:
            text: Single text string or list of text strings
            
        Returns:
            Single embedding vector or list of embedding vectors
        """
        if not self.embedder:
            raise RuntimeError("Embedder not initialized")
        
        try:
            if isinstance(text, str):
                # Single text
                logger.debug(f"🔄 Generating embedding for text: {text[:100]}...")
                response = self.embedder.embeddings.create(
                    model=self.config.model_name,
                    input=text
                )
                embedding = response.data[0].embedding
                logger.debug(f"✅ Generated embedding: {len(embedding)} dimensions")
                return embedding
            else:
                # Multiple texts
                logger.debug(f"🔄 Generating embeddings for {len(text)} texts...")
                response = self.embedder.embeddings.create(
                    model=self.config.model_name,
                    input=text
                )
                embeddings = [data.embedding for data in response.data]
                logger.debug(f"✅ Generated {len(embeddings)} embeddings")
                return embeddings
                
        except Exception as e:
            logger.error(f"❌ Failed to generate embeddings: {e}")
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
            raise RuntimeError("Embedder not initialized")
        
        try:
            if isinstance(text, str):
                logger.debug(f"🔄 Generating embedding (sync) for text: {text[:100]}...")
                response = self.embedder.embeddings.create(
                    model=self.config.model_name,
                    input=text
                )
                embedding = response.data[0].embedding
                logger.debug(f"✅ Generated embedding: {len(embedding)} dimensions")
                return embedding
            else:
                logger.debug(f"🔄 Generating embeddings (sync) for {len(text)} texts...")
                response = self.embedder.embeddings.create(
                    model=self.config.model_name,
                    input=text
                )
                embeddings = [data.embedding for data in response.data]
                logger.debug(f"✅ Generated {len(embeddings)} embeddings")
                return embeddings
                
        except Exception as e:
            logger.error(f"❌ Failed to generate embeddings (sync): {e}")
            raise
    
    def get_embedding_dimension(self) -> int:
        """Get the embedding dimension"""
        return self.config.dimensions
    
    def get_model_info(self) -> dict:
        """Get information about the current model"""
        return {
            "provider": self.config.provider,
            "model_name": self.config.model_name,
            "dimensions": self.config.dimensions,
            "base_url": self.config.base_url
        }


# Global embedder instance
_global_embedder: Optional[DatapizzaEmbedder] = None


def get_embedder(config: Optional[EmbedderConfig] = None) -> DatapizzaEmbedder:
    """
    Get or create global embedder instance.
    
    Args:
        config: Optional configuration. If None, uses environment variables.
        
    Returns:
        DatapizzaEmbedder instance
    """
    global _global_embedder
    
    if _global_embedder is None:
        _global_embedder = DatapizzaEmbedder(config)
    
    return _global_embedder


# Example usage
if __name__ == "__main__":
    import asyncio
    
    async def test_embedder():
        """Test the embedder with different configurations"""
        
        # Test 1: OpenAI embeddings
        print("\n=== Test 1: OpenAI Embeddings ===")
        openai_config = EmbedderConfig(
            provider="openai",
            api_key=os.getenv("OPENAI_API_KEY", ""),
            model_name="text-embedding-ada-002",
            dimensions=1536
        )
        embedder_openai = DatapizzaEmbedder(openai_config)
        
        embedding = await embedder_openai.embed_text("Hello world")
        print(f"OpenAI embedding dimensions: {len(embedding)}")
        print(f"First 5 values: {embedding[:5]}")
        
        # Test 2: Ollama embeddings (local)
        print("\n=== Test 2: Ollama Embeddings (Local) ===")
        ollama_config = EmbedderConfig(
            provider="ollama",
            api_key="",
            base_url="http://localhost:11434/v1",
            model_name="nomic-embed-text",
            dimensions=768
        )
        embedder_ollama = DatapizzaEmbedder(ollama_config)
        
        embedding_ollama = await embedder_ollama.embed_text("Hello world")
        print(f"Ollama embedding dimensions: {len(embedding_ollama)}")
        print(f"First 5 values: {embedding_ollama[:5]}")
        
        # Test 3: Batch embeddings
        print("\n=== Test 3: Batch Embeddings ===")
        texts = ["Hello world", "Another text", "Third example"]
        embeddings_batch = await embedder_openai.embed_text(texts)
        print(f"Generated {len(embeddings_batch)} embeddings")
        print(f"Each embedding has {len(embeddings_batch[0])} dimensions")
    
    # Run tests
    asyncio.run(test_embedder())

