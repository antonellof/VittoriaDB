"""
VittoriaDB Python client with auto-binary management.
"""

import os
import sys
import time
import atexit
import platform
import subprocess
import requests
import json
from typing import List, Dict, Any, Optional, Union
from pathlib import Path

from .types import (
    Vector,
    SearchResult,
    CollectionInfo,
    HealthStatus,
    DatabaseStats,
    DistanceMetric,
    IndexType,
    VectorizerConfig,
    VittoriaDBError,
    ConnectionError,
    CollectionError,
    VectorError,
    SearchError,
    BinaryError
)


class VittoriaDB:
    """Main VittoriaDB client class."""

    def __init__(self, 
                 url: Optional[str] = None, 
                 auto_start: bool = True,
                 port: int = 8080,
                 host: str = "localhost",
                 data_dir: Optional[str] = None):
        """Initialize VittoriaDB client.
        
        Args:
            url: VittoriaDB server URL (if None, auto-starts local server)
            auto_start: Whether to automatically start local server
            port: Port to use for auto-started server
            host: Host to bind auto-started server to
            data_dir: Data directory for auto-started server
        """
        self.url = url or f"http://{host}:{port}"
        self.auto_start = auto_start
        self.port = port
        self.host = host
        self.data_dir = data_dir or "./data"
        self.process = None
        self.auto_started = False
        
        if auto_start and url is None:
            self._start_server()
    
    def _start_server(self) -> None:
        """Start VittoriaDB binary automatically."""
        try:
            binary_path = self._get_binary_path()
            if not os.path.exists(binary_path):
                raise BinaryError(f"VittoriaDB binary not found: {binary_path}")
            
            # Start server process
            cmd = [
                binary_path, "run",
                "--host", self.host,
                "--port", str(self.port),
                "--data-dir", self.data_dir
            ]
            
            self.process = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
            
            self.auto_started = True
            
            # Wait for server to be ready
            self._wait_for_server()
            
            # Register cleanup on exit
            atexit.register(self._cleanup)
            
        except Exception as e:
            raise BinaryError(f"Failed to start VittoriaDB server: {e}")
    
    def _get_binary_path(self) -> str:
        """Get path to appropriate binary for current platform."""
        system = platform.system().lower()
        machine = platform.machine().lower()
        
        if system == "darwin":
            arch = "arm64" if machine == "arm64" else "amd64"
            binary_name = f"vittoriadb-darwin-{arch}"
        elif system == "linux":
            arch = "arm64" if machine in ("aarch64", "arm64") else "amd64"
            binary_name = f"vittoriadb-linux-{arch}"
        elif system == "windows":
            binary_name = "vittoriadb-windows-amd64.exe"
        else:
            raise BinaryError(f"Unsupported platform: {system}-{machine}")
        
        # Look for binary in package directory
        package_dir = Path(__file__).parent
        binary_path = package_dir / "binaries" / binary_name
        
        if binary_path.exists():
            return str(binary_path)
        
        # Look for binary in PATH
        import shutil
        binary_in_path = shutil.which("vittoriadb")
        if binary_in_path:
            return binary_in_path
        
        raise BinaryError(f"VittoriaDB binary not found: {binary_name}")
    
    def _wait_for_server(self, timeout: int = 30) -> None:
        """Wait for server to be ready."""
        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                response = requests.get(f"{self.url}/health", timeout=1)
                if response.status_code == 200:
                    return
            except requests.exceptions.RequestException:
                time.sleep(0.1)
        
        raise ConnectionError("VittoriaDB server failed to start within timeout")
    
    def _cleanup(self) -> None:
        """Cleanup server process."""
        if self.process and self.auto_started:
            try:
                self.process.terminate()
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.process.kill()
            except Exception:
                pass  # Ignore cleanup errors
    
    def _make_request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        """Make HTTP request to VittoriaDB server."""
        url = f"{self.url}{endpoint}"
        
        try:
            response = requests.request(method, url, **kwargs)
            return response
        except requests.exceptions.RequestException as e:
            raise ConnectionError(f"Failed to connect to VittoriaDB server: {e}")
    
    def _handle_response(self, response: requests.Response) -> Dict[str, Any]:
        """Handle HTTP response and raise appropriate errors."""
        try:
            data = response.json()
        except json.JSONDecodeError:
            data = {"error": "Invalid JSON response"}
        
        if response.status_code >= 400:
            error_msg = data.get("error", f"HTTP {response.status_code}")
            details = data.get("details", "")
            
            if response.status_code == 404:
                raise CollectionError(f"{error_msg}: {details}")
            elif response.status_code == 409:
                raise CollectionError(f"{error_msg}: {details}")
            else:
                raise VittoriaDBError(f"{error_msg}: {details}")
        
        return data
    
    def create_collection(self, 
                         name: str, 
                         dimensions: int,
                         metric: Union[DistanceMetric, str] = DistanceMetric.COSINE,
                         index_type: Union[IndexType, str] = IndexType.FLAT,
                         config: Optional[Dict[str, Any]] = None,
                         vectorizer_config: Optional[VectorizerConfig] = None) -> 'Collection':
        """Create a new vector collection."""
        # Convert to enum values and then to integers (Go server expects integers)
        if isinstance(metric, DistanceMetric):
            metric_int = metric.value
        elif isinstance(metric, str):
            metric_int = DistanceMetric.from_string(metric).value
        else:
            metric_int = DistanceMetric.COSINE.value
            
        if isinstance(index_type, IndexType):
            index_int = index_type.value
        elif isinstance(index_type, str):
            index_int = IndexType.from_string(index_type).value
        else:
            index_int = IndexType.FLAT.value
        
        payload = {
            "name": name,
            "dimensions": dimensions,
            "metric": metric_int,
            "index_type": index_int,
            "config": config or {}
        }
        
        # Add vectorizer configuration if provided
        if vectorizer_config:
            payload["vectorizer_config"] = vectorizer_config.to_dict()
        
        response = self._make_request("POST", "/collections", json=payload)
        self._handle_response(response)
        
        return Collection(self, name)
    
    def get_collection(self, name: str) -> 'Collection':
        """Get an existing collection."""
        response = self._make_request("GET", f"/collections/{name}")
        self._handle_response(response)
        
        return Collection(self, name)
    
    def list_collections(self) -> List[CollectionInfo]:
        """List all collections."""
        response = self._make_request("GET", "/collections")
        data = self._handle_response(response)
        
        collections = []
        for collection_data in data.get("collections", []):
            collections.append(CollectionInfo.from_dict(collection_data))
        
        return collections
    
    def delete_collection(self, name: str) -> None:
        """Delete a collection."""
        response = self._make_request("DELETE", f"/collections/{name}")
        self._handle_response(response)
    
    def health(self) -> HealthStatus:
        """Get health status."""
        response = self._make_request("GET", "/health")
        data = self._handle_response(response)
        
        return HealthStatus.from_dict(data)
    
    def stats(self) -> DatabaseStats:
        """Get database statistics."""
        response = self._make_request("GET", "/stats")
        data = self._handle_response(response)
        
        return DatabaseStats.from_dict(data)
    
    def close(self) -> None:
        """Close connection and cleanup."""
        self._cleanup()


class Collection:
    """Vector collection interface."""
    
    def __init__(self, client: VittoriaDB, name: str):
        """Initialize collection."""
        self.client = client
        self.name = name
        self._info = None
    
    @property
    def info(self) -> CollectionInfo:
        """Collection information (cached)."""
        if self._info is None:
            response = self.client._make_request("GET", f"/collections/{self.name}")
            data = self.client._handle_response(response)
            self._info = CollectionInfo.from_dict(data)
        
        return self._info
    
    def insert(self, 
               id: str, 
               vector: List[float],
               metadata: Optional[Dict[str, Any]] = None) -> None:
        """Insert a single vector."""
        payload = {
            "id": id,
            "vector": vector,
            "metadata": metadata or {}
        }
        
        response = self.client._make_request(
            "POST", 
            f"/collections/{self.name}/vectors", 
            json=payload
        )
        self.client._handle_response(response)
        
        # Invalidate cached info
        self._info = None
    
    def insert_batch(self, vectors: List[Union[Vector, Dict[str, Any]]]) -> Dict[str, Any]:
        """Insert multiple vectors."""
        # Convert Vector objects to dictionaries
        vector_dicts = []
        for vector in vectors:
            if isinstance(vector, Vector):
                vector_dicts.append({
                    "id": vector.id,
                    "vector": vector.vector,
                    "metadata": vector.metadata or {}
                })
            else:
                vector_dicts.append(vector)
        
        payload = {"vectors": vector_dicts}
        
        response = self.client._make_request(
            "POST", 
            f"/collections/{self.name}/vectors/batch", 
            json=payload
        )
        result = self.client._handle_response(response)
        
        # Invalidate cached info
        self._info = None
        
        return result
    
    def search(self,
               vector: List[float],
               limit: int = 10,
               offset: int = 0,
               filter: Optional[Dict[str, Any]] = None,
               include_vector: bool = False,
               include_metadata: bool = True) -> List[SearchResult]:
        """Search for similar vectors."""
        params = {
            "vector": ",".join(map(str, vector)),
            "limit": limit,
            "offset": offset,
            "include_vector": str(include_vector).lower(),
            "include_metadata": str(include_metadata).lower()
        }
        
        if filter:
            params["filter"] = json.dumps(filter)
        
        response = self.client._make_request(
            "GET", 
            f"/collections/{self.name}/search",
            params=params
        )
        data = self.client._handle_response(response)
        
        results = []
        for result_data in data.get("results", []):
            results.append(SearchResult.from_dict(result_data))
        
        return results
    
    def get(self, id: str) -> Optional[Vector]:
        """Get a vector by ID."""
        try:
            response = self.client._make_request("GET", f"/collections/{self.name}/vectors/{id}")
            data = self.client._handle_response(response)
            
            return Vector(
                id=data["id"],
                vector=data["vector"],
                metadata=data.get("metadata", {})
            )
        except CollectionError as e:
            if "not found" in str(e).lower():
                return None
            raise
    
    def delete(self, id: str) -> None:
        """Delete a vector by ID."""
        response = self.client._make_request("DELETE", f"/collections/{self.name}/vectors/{id}")
        self.client._handle_response(response)
        
        # Invalidate cached info
        self._info = None
    
    def count(self) -> int:
        """Get total number of vectors."""
        return self.info.vector_count
    
    def insert_text(self, 
                   id: str,
                   text: str,
                   metadata: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Insert text that will be automatically vectorized.
        
        Requires the collection to have a vectorizer configured.
        
        Args:
            id: Unique identifier for the text
            text: Text content to vectorize and insert
            metadata: Optional metadata dictionary
            
        Returns:
            Dictionary with insertion status
            
        Raises:
            CollectionError: If collection doesn't have vectorizer configured
        """
        payload = {
            "id": id,
            "text": text,
            "metadata": metadata or {}
        }
        
        try:
            response = self.client._make_request("POST", f"/collections/{self.name}/text", json=payload)
            return self.client._handle_response(response)
        except Exception as e:
            if "vectorizer" in str(e).lower():
                raise CollectionError(f"Collection '{self.name}' does not have vectorizer configured. Use create_collection() with vectorizer_config parameter.")
            raise VectorError(f"Failed to insert text: {e}")
    
    def insert_text_batch(self, 
                         texts: List[Dict[str, Any]]) -> Dict[str, Any]:
        """
        Insert multiple texts that will be automatically vectorized.
        
        Requires the collection to have a vectorizer configured.
        
        Args:
            texts: List of dictionaries with 'id', 'text', and optional 'metadata' keys
            
        Returns:
            Dictionary with batch insertion status
            
        Raises:
            CollectionError: If collection doesn't have vectorizer configured
        """
        payload = {"texts": texts}
        
        try:
            response = self.client._make_request("POST", f"/collections/{self.name}/text/batch", json=payload)
            return self.client._handle_response(response)
        except Exception as e:
            if "vectorizer" in str(e).lower():
                raise CollectionError(f"Collection '{self.name}' does not have vectorizer configured. Use create_collection() with vectorizer_config parameter.")
            raise VectorError(f"Failed to insert text batch: {e}")
    
    def search_text(self, 
                   query: str,
                   limit: int = 10,
                   filter: Optional[Dict[str, Any]] = None,
                   include_metadata: bool = True) -> List[SearchResult]:
        """
        Search using a text query that will be automatically vectorized.
        
        Requires the collection to have a vectorizer configured.
        
        Args:
            query: Text query to search for
            limit: Maximum number of results to return
            filter: Optional metadata filter
            include_metadata: Whether to include metadata in results
            
        Returns:
            List of SearchResult objects
            
        Raises:
            CollectionError: If collection doesn't have vectorizer configured
        """
        payload = {
            "query": query,
            "limit": limit,
            "include_metadata": include_metadata
        }
        
        if filter:
            payload["filter"] = filter
        
        try:
            response = self.client._make_request("POST", f"/collections/{self.name}/search/text", json=payload)
            data = self.client._handle_response(response)
            
            return [SearchResult.from_dict(result) for result in data.get("results", [])]
        except Exception as e:
            if "vectorizer" in str(e).lower():
                raise CollectionError(f"Collection '{self.name}' does not have vectorizer configured. Use create_collection() with vectorizer_config parameter.")
            raise SearchError(f"Failed to search text: {e}")
    
    def upload_file(self, 
                    file_path: str,
                    chunk_size: int = 500,
                    chunk_overlap: int = 50,
                    language: str = "en",
                    metadata: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Upload and process a document file.
        
        The document will be processed, chunked, and automatically vectorized
        if the collection has a vectorizer configured.
        
        Args:
            file_path: Path to the document file
            chunk_size: Size of text chunks in characters
            chunk_overlap: Overlap between chunks in characters  
            language: Document language for processing
            metadata: Optional metadata dictionary
            
        Returns:
            Dictionary with upload status and processing results
        """
        if not os.path.exists(file_path):
            raise VectorError(f"File not found: {file_path}")
        
        with open(file_path, 'rb') as f:
            files = {'file': f}
            data = {
                'chunk_size': chunk_size,
                'chunk_overlap': chunk_overlap,
                'language': language
            }
            
            if metadata:
                data['metadata'] = json.dumps(metadata)
            
            response = self.client._make_request(
                "POST",
                f"/collections/{self.name}/upload",
                files=files,
                data=data
            )
            
            result = self.client._handle_response(response)
            
            # Invalidate cached info
            self._info = None
            
            return result
    
    def upload_files(self, 
                     file_paths: List[str],
                     **kwargs) -> Dict[str, Any]:
        """Upload multiple files."""
        results = []
        for file_path in file_paths:
            try:
                result = self.upload_file(file_path, **kwargs)
                results.append({"file": file_path, "success": True, "result": result})
            except Exception as e:
                results.append({"file": file_path, "success": False, "error": str(e)})
        
        return {
            "results": results,
            "total": len(file_paths),
            "successful": sum(1 for r in results if r["success"]),
            "failed": sum(1 for r in results if not r["success"])
        }
    
    def process_text(self,
                     text: str,
                     chunk_size: int = 500,
                     overlap: int = 50,
                     metadata: Optional[Dict[str, Any]] = None,
                     embedding_model: str = "sentence-transformers/all-MiniLM-L6-v2") -> Dict[str, Any]:
        """Process raw text into vectors."""
        # For now, we'll create a temporary file and upload it
        # In the future, we can add a direct text processing endpoint
        import tempfile
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as f:
            f.write(text)
            temp_path = f.name
        
        try:
            result = self.upload_file(
                temp_path,
                chunk_size=chunk_size,
                overlap=overlap,
                metadata=metadata,
                embedding_model=embedding_model
            )
            return result
        finally:
            os.unlink(temp_path)


# Convenience functions
def connect(url: Optional[str] = None, **kwargs) -> VittoriaDB:
    """Connect to VittoriaDB (auto-starts if needed)."""
    return VittoriaDB(url, **kwargs)


def embed(data_dir: str) -> VittoriaDB:
    """Create embedded VittoriaDB instance (future feature)."""
    raise NotImplementedError("Embedded mode coming in v0.2")


# Utility functions
def supported_formats() -> List[str]:
    """Get list of supported file formats."""
    return [".pdf", ".docx", ".doc", ".txt", ".md", ".html", ".htm", ".rtf"]


def available_models() -> List[str]:
    """Get list of available embedding models."""
    return [
        "sentence-transformers/all-MiniLM-L6-v2",
        "sentence-transformers/all-mpnet-base-v2",
        "sentence-transformers/distilbert-base-nli-stsb-mean-tokens",
        "openai/text-embedding-ada-002"
    ]


def extract_text(file_path: str) -> str:
    """Extract text from file without uploading."""
    # This would require implementing text extraction locally
    # For now, raise NotImplementedError
    raise NotImplementedError("Local text extraction coming in v0.2")
