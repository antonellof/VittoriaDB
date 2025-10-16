"""
GitHub Repository Indexer
Indexes GitHub repositories for code search and RAG
"""

import os
import logging
import time
import tempfile
import shutil
from typing import List, Dict, Any, Optional, Tuple
from dataclasses import dataclass
import hashlib
import re

from github import Github
import git

# Import chunking function from rag_system
from rag_system import chunk_text

logger = logging.getLogger(__name__)

@dataclass
class CodeFile:
    """Represents a code file from GitHub"""
    path: str
    content: str
    language: str
    size: int
    metadata: Dict[str, Any]

@dataclass
class GitHubRepository:
    """Represents a GitHub repository"""
    owner: str
    name: str
    url: str
    description: str
    language: str
    stars: int
    files: List[CodeFile]
    metadata: Dict[str, Any]

class GitHubIndexer:
    """GitHub repository indexer for code search"""
    
    def __init__(self, github_token: Optional[str] = None):
        """Initialize GitHub indexer"""
        self.github_token = github_token
        self.github_client = None
        
        if github_token:
            self.github_client = Github(github_token)
        else:
            # Use unauthenticated client (rate limited)
            self.github_client = Github()
        
        # Supported code file extensions
        self.code_extensions = {
            '.py': 'python',
            '.js': 'javascript',
            '.ts': 'typescript',
            '.jsx': 'javascript',
            '.tsx': 'typescript',
            '.java': 'java',
            '.cpp': 'cpp',
            '.c': 'c',
            '.h': 'c',
            '.hpp': 'cpp',
            '.cs': 'csharp',
            '.php': 'php',
            '.rb': 'ruby',
            '.go': 'go',
            '.rs': 'rust',
            '.swift': 'swift',
            '.kt': 'kotlin',
            '.scala': 'scala',
            '.r': 'r',
            '.sql': 'sql',
            '.sh': 'bash',
            '.yaml': 'yaml',
            '.yml': 'yaml',
            '.json': 'json',
            '.xml': 'xml',
            '.html': 'html',
            '.css': 'css',
            '.scss': 'scss',
            '.less': 'less',
            '.md': 'markdown',
            '.rst': 'rst',
            '.txt': 'text'
        }
        
        # Files to ignore
        self.ignore_patterns = [
            r'\.git/',
            r'node_modules/',
            r'__pycache__/',
            r'\.pyc$',
            r'\.pyo$',
            r'\.class$',
            r'\.jar$',
            r'\.war$',
            r'\.ear$',
            r'\.zip$',
            r'\.tar\.gz$',
            r'\.rar$',
            r'\.7z$',
            r'\.exe$',
            r'\.dll$',
            r'\.so$',
            r'\.dylib$',
            r'\.log$',
            r'\.tmp$',
            r'\.cache$',
            r'build/',
            r'dist/',
            r'target/',
            r'\.DS_Store$',
            r'Thumbs\.db$'
        ]
        
        self.max_file_size = 100000  # 100KB max per file
        self.max_files_per_repo = 500  # Limit files per repository
    
    async def index_repository(self, repo_url: str) -> GitHubRepository:
        """Index a GitHub repository"""
        
        logger.info(f"🔍 Indexing repository: {repo_url}")
        
        # Parse repository URL
        owner, repo_name = self._parse_repo_url(repo_url)
        
        try:
            # Get repository info
            repo = self.github_client.get_repo(f"{owner}/{repo_name}")
            
            # Create temporary directory for cloning
            with tempfile.TemporaryDirectory() as temp_dir:
                # Clone repository
                clone_path = os.path.join(temp_dir, repo_name)
                git.Repo.clone_from(repo_url, clone_path, depth=1)  # Shallow clone
                
                # Index files
                code_files = await self._index_files(clone_path, owner, repo_name)
                
                # Create repository object
                github_repo = GitHubRepository(
                    owner=owner,
                    name=repo_name,
                    url=repo_url,
                    description=repo.description or "No description",
                    language=repo.language or "Unknown",
                    stars=repo.stargazers_count,
                    files=code_files,
                    metadata={
                        'indexed_at': time.time(),
                        'total_files': len(code_files),
                        'repository_size': sum(f.size for f in code_files),
                        'languages': list(set(f.language for f in code_files)),
                        'default_branch': repo.default_branch,
                        'created_at': repo.created_at.isoformat() if repo.created_at else None,
                        'updated_at': repo.updated_at.isoformat() if repo.updated_at else None,
                    }
                )
                
                logger.info(f"✅ Indexed {len(code_files)} files from {owner}/{repo_name}")
                return github_repo
                
        except Exception as e:
            logger.error(f"❌ Failed to index repository {repo_url}: {e}")
            raise ValueError(f"Failed to index repository: {str(e)}")
    
    async def _index_files(self, repo_path: str, owner: str, repo_name: str) -> List[CodeFile]:
        """Index files in the cloned repository"""
        
        code_files = []
        file_count = 0
        
        for root, dirs, files in os.walk(repo_path):
            # Skip ignored directories
            dirs[:] = [d for d in dirs if not self._should_ignore(os.path.join(root, d))]
            
            for file in files:
                if file_count >= self.max_files_per_repo:
                    logger.warning(f"Reached max files limit ({self.max_files_per_repo}) for {owner}/{repo_name}")
                    break
                
                file_path = os.path.join(root, file)
                relative_path = os.path.relpath(file_path, repo_path)
                
                # Skip ignored files
                if self._should_ignore(relative_path):
                    continue
                
                # Check file extension
                file_ext = os.path.splitext(file)[1].lower()
                if file_ext not in self.code_extensions:
                    continue
                
                # Check file size
                try:
                    file_size = os.path.getsize(file_path)
                    if file_size > self.max_file_size:
                        logger.debug(f"Skipping large file: {relative_path} ({file_size} bytes)")
                        continue
                except:
                    continue
                
                # Read file content
                try:
                    with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                        content = f.read()
                    
                    # Create code file object
                    code_file = CodeFile(
                        path=relative_path,
                        content=content,
                        language=self.code_extensions[file_ext],
                        size=file_size,
                        metadata={
                            'repository': f"{owner}/{repo_name}",
                            'file_extension': file_ext,
                            'lines_of_code': len(content.split('\n')),
                            'indexed_at': time.time()
                        }
                    )
                    
                    code_files.append(code_file)
                    file_count += 1
                    
                except Exception as e:
                    logger.debug(f"Failed to read file {relative_path}: {e}")
                    continue
        
        return code_files
    
    def _parse_repo_url(self, repo_url: str) -> Tuple[str, str]:
        """Parse GitHub repository URL to extract owner and repo name"""
        
        # Handle different URL formats
        patterns = [
            r'github\.com/([^/]+)/([^/]+?)(?:\.git)?/?$',
            r'github\.com/([^/]+)/([^/]+)',
        ]
        
        for pattern in patterns:
            match = re.search(pattern, repo_url)
            if match:
                owner, repo_name = match.groups()
                # Remove .git suffix if present
                repo_name = repo_name.replace('.git', '')
                return owner, repo_name
        
        raise ValueError(f"Invalid GitHub repository URL: {repo_url}")
    
    def _should_ignore(self, file_path: str) -> bool:
        """Check if file should be ignored"""
        
        for pattern in self.ignore_patterns:
            if re.search(pattern, file_path):
                return True
        
        return False
    
    async def store_repository(self, 
                             github_repo: GitHubRepository,
                             rag_system,
                             progress_callback=None) -> List[str]:
        """Store repository files in VittoriaDB with batch operations for better performance"""
        
        total_files = len(github_repo.files)
        batch_size = 10  # Process files in batches of 10
        stored_ids = []
        
        if progress_callback:
            await progress_callback(55, f"Preparing {total_files} files for batch processing...")
        
        # Prepare all documents for batch insertion
        documents = []
        for code_file in github_repo.files:
            try:
                # Create content for storage
                full_content = f"""
Repository: {github_repo.owner}/{github_repo.name}
File: {code_file.path}
Language: {code_file.language}
Lines: {code_file.metadata.get('lines_of_code', 0)}

Code:
{code_file.content}
                """.strip()
                
                # Create base metadata
                base_metadata = {
                    'type': 'github_code',
                    'repository': f"{github_repo.owner}/{github_repo.name}",
                    'repository_url': github_repo.url,
                    'file_path': code_file.path,
                    'language': code_file.language,
                    'file_size': code_file.size,
                    'lines_of_code': code_file.metadata.get('lines_of_code', 0),
                    'repository_description': github_repo.description,
                    'repository_stars': github_repo.stars,
                    'indexed_at': time.time(),
                    'content': code_file.content,  # For search results display
                    'title': f"{github_repo.name}/{code_file.path}"
                }
                
                # Chunk large files to fit within OpenAI token limits
                chunks = chunk_text(full_content, max_tokens=6000, overlap=200)
                
                if len(chunks) == 1:
                    # Single chunk - use original approach
                    documents.append({
                        'content': full_content,
                        'metadata': base_metadata
                    })
                else:
                    # Multiple chunks - create separate documents for each chunk
                    for i, chunk in enumerate(chunks):
                        chunk_metadata = {
                            **base_metadata,
                            'chunk_index': i,
                            'total_chunks': len(chunks),
                            'is_chunk': True,
                            'original_file_id': f"{github_repo.name}_{code_file.path}".replace('/', '_'),
                            'title': f"{github_repo.name}/{code_file.path} (chunk {i+1}/{len(chunks)})"
                        }
                        
                        documents.append({
                            'content': chunk,
                            'metadata': chunk_metadata
                        })
                
            except Exception as e:
                logger.error(f"Failed to prepare code file {code_file.path}: {e}")
        
        # Process documents in batches
        total_batches = (len(documents) + batch_size - 1) // batch_size
        
        for batch_idx in range(0, len(documents), batch_size):
            batch_docs = documents[batch_idx:batch_idx + batch_size]
            batch_num = (batch_idx // batch_size) + 1
            
            try:
                if progress_callback:
                    progress = 60 + int((batch_num - 1) / total_batches * 30)  # 60-90% range
                    await progress_callback(progress, f"Processing batch {batch_num}/{total_batches} ({len(batch_docs)} files)...")
                
                # Use batch insertion for better performance
                batch_ids = await rag_system.add_documents_batch(
                    documents=batch_docs,
                    collection_name='github_code'
                )
                
                stored_ids.extend(batch_ids)
                
            except Exception as e:
                logger.error(f"Failed to store batch {batch_num}: {e}")
                # Fallback to individual insertion for this batch
                for doc in batch_docs:
                    try:
                        doc_id = await rag_system.add_document(
                            content=doc['content'],
                            metadata=doc['metadata'],
                            collection_name='github_code'
                        )
                        stored_ids.append(doc_id)
                    except Exception as fallback_error:
                        logger.error(f"Failed to store individual document: {fallback_error}")
        
        logger.info(f"✅ Stored {len(stored_ids)} code files from {github_repo.owner}/{github_repo.name} using batch processing")
        return stored_ids
    
    async def index_and_store(self, 
                            repo_url: str,
                            rag_system,
                            progress_callback=None) -> Dict[str, Any]:
        """Complete GitHub indexing workflow with progress callbacks"""
        
        start_time = time.time()
        
        try:
            if progress_callback:
                await progress_callback(20, "Fetching repository information...")
            
            # Index repository
            github_repo = await self.index_repository(repo_url)
            
            if progress_callback:
                await progress_callback(50, f"Processing {len(github_repo.files)} files...")
            
            # Store in VittoriaDB with progress updates
            stored_ids = await self.store_repository(github_repo, rag_system, progress_callback)
            
            if progress_callback:
                await progress_callback(95, "Finalizing indexing...")
            
            processing_time = time.time() - start_time
            
            result = {
                'success': True,
                'message': f'Successfully indexed and stored {len(stored_ids)} files',
                'repository': f"{github_repo.owner}/{github_repo.name}",
                'repository_url': repo_url,
                'files_indexed': len(github_repo.files),
                'files_stored': len(stored_ids),
                'languages': github_repo.metadata['languages'],
                'repository_stars': github_repo.stars,
                'processing_time': processing_time
            }
            
            if progress_callback:
                await progress_callback(100, "Indexing completed successfully!")
                
            return result
            
        except Exception as e:
            logger.error(f"GitHub indexing failed: {e}")
            return {
                'success': False,
                'message': f'Failed to index repository: {str(e)}',
                'repository_url': repo_url,
                'processing_time': time.time() - start_time
            }

# Global GitHub indexer instance
_github_indexer = None

def get_github_indexer() -> GitHubIndexer:
    """Get or create global GitHub indexer instance"""
    global _github_indexer
    if _github_indexer is None:
        github_token = os.getenv('GITHUB_TOKEN')
        _github_indexer = GitHubIndexer(github_token=github_token)
    return _github_indexer
