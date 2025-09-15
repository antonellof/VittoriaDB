"""
Web Research Module
Handles web search, scraping, and knowledge storage
"""

import os
import logging
import asyncio
import time
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
import hashlib
import re

import requests
from bs4 import BeautifulSoup
from ddgs import DDGS
import httpx

logger = logging.getLogger(__name__)

@dataclass
class WebSearchResult:
    """Web search result"""
    title: str
    url: str
    snippet: str
    content: str
    timestamp: float
    source: str = "web_search"

@dataclass
class ScrapedContent:
    """Scraped web content"""
    url: str
    title: str
    content: str
    metadata: Dict[str, Any]
    success: bool
    error: Optional[str] = None

class WebResearcher:
    """Advanced web research with automatic knowledge storage"""
    
    def __init__(self, max_results: int = 5, max_content_length: int = 5000):
        """Initialize web researcher"""
        self.max_results = max_results
        self.max_content_length = max_content_length
        
        # User agent for web scraping
        self.headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
        }
        
        # Initialize search engines
        self.search_engines = {
            'duckduckgo': self._search_duckduckgo,
        }
    
    async def research_query(self, 
                           query: str,
                           search_engine: str = 'duckduckgo',
                           scrape_content: bool = True) -> List[WebSearchResult]:
        """Research a query and return structured results"""
        
        logger.info(f"🔍 Researching: {query}")
        
        # Search for results
        search_func = self.search_engines.get(search_engine, self._search_duckduckgo)
        search_results = await search_func(query)
        
        if not search_results:
            logger.warning(f"No search results found for: {query}")
            return []
        
        # Scrape content if requested
        if scrape_content:
            scraped_results = []
            for result in search_results[:self.max_results]:
                scraped = await self._scrape_url(result['url'])
                if scraped.success:
                    web_result = WebSearchResult(
                        title=result['title'],
                        url=result['url'],
                        snippet=result['snippet'],
                        content=scraped.content,
                        timestamp=time.time(),
                        source="web_search"
                    )
                    scraped_results.append(web_result)
                    logger.info(f"✅ Scraped: {result['title']}")
                else:
                    # Use snippet if scraping fails
                    web_result = WebSearchResult(
                        title=result['title'],
                        url=result['url'],
                        snippet=result['snippet'],
                        content=result['snippet'],
                        timestamp=time.time(),
                        source="web_search"
                    )
                    scraped_results.append(web_result)
                    logger.warning(f"⚠️ Scraping failed for {result['url']}: {scraped.error}")
            
            return scraped_results
        else:
            # Return search results without scraping
            return [
                WebSearchResult(
                    title=result['title'],
                    url=result['url'],
                    snippet=result['snippet'],
                    content=result['snippet'],
                    timestamp=time.time(),
                    source="web_search"
                )
                for result in search_results[:self.max_results]
            ]
    
    async def _search_duckduckgo(self, query: str) -> List[Dict[str, str]]:
        """Search using DuckDuckGo"""
        try:
            ddgs = DDGS()
            results = []
            for result in ddgs.text(query, max_results=self.max_results):
                results.append({
                    'title': result.get('title', 'No Title'),
                    'url': result.get('href', ''),
                    'snippet': result.get('body', 'No snippet available')
                })
            return results
        except Exception as e:
            logger.error(f"DuckDuckGo search failed: {e}")
            return []
    
    async def _scrape_url(self, url: str) -> ScrapedContent:
        """Scrape content from a URL"""
        try:
            # Use httpx for async requests
            async with httpx.AsyncClient(headers=self.headers, timeout=10.0) as client:
                response = await client.get(url)
                response.raise_for_status()
                
                # Parse HTML
                soup = BeautifulSoup(response.text, 'html.parser')
                
                # Remove script and style elements
                for script in soup(["script", "style", "nav", "footer", "header"]):
                    script.decompose()
                
                # Extract title
                title_tag = soup.find('title')
                title = title_tag.get_text().strip() if title_tag else 'No Title'
                
                # Extract main content
                content = self._extract_main_content(soup)
                
                # Limit content length
                if len(content) > self.max_content_length:
                    content = content[:self.max_content_length] + "..."
                
                metadata = {
                    'url': url,
                    'title': title,
                    'scraped_at': time.time(),
                    'content_length': len(content),
                    'domain': self._extract_domain(url)
                }
                
                return ScrapedContent(
                    url=url,
                    title=title,
                    content=content,
                    metadata=metadata,
                    success=True
                )
                
        except Exception as e:
            logger.error(f"Failed to scrape {url}: {e}")
            return ScrapedContent(
                url=url,
                title="Scraping Failed",
                content="",
                metadata={'url': url, 'error': str(e)},
                success=False,
                error=str(e)
            )
    
    def _extract_main_content(self, soup: BeautifulSoup) -> str:
        """Extract main content from HTML soup"""
        
        # Try to find main content areas
        main_selectors = [
            'main',
            'article',
            '.content',
            '.main-content',
            '#content',
            '.post-content',
            '.entry-content'
        ]
        
        for selector in main_selectors:
            main_content = soup.select_one(selector)
            if main_content:
                text = main_content.get_text(separator=' ', strip=True)
                if len(text) > 100:  # Ensure we have substantial content
                    return self._clean_text(text)
        
        # Fallback: extract from body
        body = soup.find('body')
        if body:
            text = body.get_text(separator=' ', strip=True)
            return self._clean_text(text)
        
        # Final fallback: all text
        return self._clean_text(soup.get_text(separator=' ', strip=True))
    
    def _clean_text(self, text: str) -> str:
        """Clean extracted text"""
        # Remove extra whitespace
        text = re.sub(r'\s+', ' ', text)
        
        # Remove common noise
        noise_patterns = [
            r'Cookie Policy.*?Accept',
            r'Privacy Policy.*?Accept',
            r'Subscribe.*?Newsletter',
            r'Follow us on.*?Twitter',
        ]
        
        for pattern in noise_patterns:
            text = re.sub(pattern, '', text, flags=re.IGNORECASE)
        
        return text.strip()
    
    def _extract_domain(self, url: str) -> str:
        """Extract domain from URL"""
        try:
            from urllib.parse import urlparse
            return urlparse(url).netloc
        except:
            return "unknown"
    
    async def store_research_results(self, 
                                   results: List[WebSearchResult],
                                   rag_system,
                                   query: str) -> List[str]:
        """Store research results in VittoriaDB"""
        stored_ids = []
        
        for result in results:
            try:
                # Create content for storage
                content = f"""
Title: {result.title}
URL: {result.url}
Query: {query}

Content:
{result.content}

Snippet:
{result.snippet}
                """.strip()
                
                # Create metadata
                metadata = {
                    'type': 'web_research',
                    'title': result.title,
                    'document_title': result.title,  # For RAG engine compatibility
                    'url': result.url,
                    'query': query,
                    'snippet': result.snippet,
                    'timestamp': result.timestamp,
                    'source': result.source,
                    'source_collection': 'web_search',  # For RAG engine identification
                    'content': result.content,  # For search results display
                    'domain': self._extract_domain(result.url)
                }
                
                # Store in VittoriaDB
                doc_id = await rag_system.add_document(
                    content=content,
                    metadata=metadata,
                    collection_name='web_research'
                )
                
                stored_ids.append(doc_id)
                logger.info(f"✅ Stored research result: {result.title}")
                
            except Exception as e:
                logger.error(f"Failed to store research result {result.title}: {e}")
        
        return stored_ids
    
    async def research_and_store(self, 
                               query: str,
                               rag_system,
                               search_engine: str = 'duckduckgo') -> Dict[str, Any]:
        """Complete research workflow: search, scrape, and store"""
        
        start_time = time.time()
        
        # Research the query
        results = await self.research_query(
            query=query,
            search_engine=search_engine,
            scrape_content=True
        )
        
        if not results:
            return {
                'success': False,
                'message': 'No research results found',
                'query': query,
                'results_count': 0,
                'stored_count': 0,
                'processing_time': time.time() - start_time,
                'results': []
            }
        
        # Store results in VittoriaDB
        stored_ids = await self.store_research_results(results, rag_system, query)
        
        processing_time = time.time() - start_time
        
        return {
            'success': True,
            'message': f'Successfully researched and stored {len(stored_ids)} results',
            'query': query,
            'results_count': len(results),
            'stored_count': len(stored_ids),
            'processing_time': processing_time,
            'results': [
                {
                    'title': r.title,
                    'url': r.url,
                    'snippet': r.snippet[:200] + "..." if len(r.snippet) > 200 else r.snippet
                }
                for r in results
            ]
        }
    
    async def stream_research_and_store(self, 
                                      query: str,
                                      rag_system,
                                      search_engine: str = 'duckduckgo'):
        """Streaming research workflow: search, scrape, and store with real-time updates"""
        
        try:
            # Step 1: Search for results
            search_func = self.search_engines.get(search_engine, self._search_duckduckgo)
            search_results = await search_func(query)
            
            if not search_results:
                yield {
                    'type': 'error',
                    'message': 'No search results found'
                }
                return
            
            # Yield initial search results immediately
            yield {
                'type': 'search_results',
                'results': [
                    {
                        'title': result['title'],
                        'url': result['url'],
                        'snippet': result['snippet']
                    }
                    for result in search_results
                ]
            }
            
            # Step 2: Scrape content and store (in background)
            scraped_results = []
            stored_count = 0
            
            for result in search_results[:self.max_results]:
                try:
                    # Scrape content
                    scraped = await self._scrape_url(result['url'])
                    
                    if scraped.success:
                        web_result = WebSearchResult(
                            title=result['title'],
                            url=result['url'],
                            snippet=result['snippet'],
                            content=scraped.content,
                            timestamp=time.time(),
                            source="web_search"
                        )
                        scraped_results.append(web_result)
                        
                        # Store in VittoriaDB
                        stored_id = await self._store_single_result(web_result, rag_system, query)
                        if stored_id:
                            stored_count += 1
                            
                except Exception as e:
                    logger.warning(f"Failed to process {result['url']}: {e}")
                    continue
            
            # Final completion update
            yield {
                'type': 'complete',
                'total_results': len(scraped_results),
                'stored_count': stored_count,
                'message': f'Completed: {stored_count} results stored'
            }
            
        except Exception as e:
            yield {
                'type': 'error',
                'message': str(e)
            }
    
    async def _store_single_result(self, result: WebSearchResult, rag_system, query: str) -> Optional[str]:
        """Store a single web search result"""
        try:
            content = f"""
Title: {result.title}
URL: {result.url}
Query: {query}

Content:
{result.content}

Snippet:
{result.snippet}
            """.strip()
            
            # Generate unique ID
            content_hash = hashlib.md5(f"{result.url}_{query}".encode()).hexdigest()
            doc_id = f"web_{content_hash}"
            
            # Store in web_research collection (will be searched by RAG engine)
            stored_doc_id = await rag_system.add_document(
                content=content,
                metadata={
                    "title": result.title,
                    "document_title": result.title,  # For RAG engine compatibility
                    "url": result.url,
                    "query": query,
                    "snippet": result.snippet,
                    "timestamp": result.timestamp,
                    "source": "web_search",  # Mark as web search result
                    "source_collection": "web_search",  # For RAG engine identification
                    "content_length": len(result.content)
                },
                collection_name="web_research"
            )
            
            return stored_doc_id
            
        except Exception as e:
            logger.error(f"Failed to store result {result.url}: {e}")
            return None

# Global web researcher instance
_web_researcher = None

def get_web_researcher() -> WebResearcher:
    """Get or create global web researcher instance"""
    global _web_researcher
    if _web_researcher is None:
        _web_researcher = WebResearcher()
    return _web_researcher
