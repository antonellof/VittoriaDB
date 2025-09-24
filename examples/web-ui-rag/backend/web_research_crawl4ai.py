"""
Advanced Web Research Module with Crawl4AI
LLM-friendly web scraping and content extraction
"""

import os
import logging
import asyncio
import time

# Fix tokenizers multiprocessing warning
os.environ["TOKENIZERS_PARALLELISM"] = "false"
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
import hashlib
import re

from ddgs import DDGS
from crawl4ai import AsyncWebCrawler, BrowserConfig, CrawlerRunConfig, CacheMode
from crawl4ai.extraction_strategy import LLMExtractionStrategy, CosineStrategy
import httpx
import openai
import json

logger = logging.getLogger(__name__)

@dataclass
class WebSearchResult:
    """Enhanced web search result with Crawl4AI extraction"""
    title: str
    url: str
    snippet: str
    content: str
    timestamp: float
    source: str = "web_search"
    structured_data: Optional[Dict[str, Any]] = None
    markdown_content: Optional[str] = None
    links: Optional[List[Dict[str, str]]] = None
    media: Optional[List[Dict[str, str]]] = None

@dataclass
class ScrapedContent:
    """Enhanced scraped web content with Crawl4AI features"""
    url: str
    title: str
    content: str
    metadata: Dict[str, Any]
    success: bool
    error: Optional[str] = None
    markdown_content: Optional[str] = None
    structured_data: Optional[Dict[str, Any]] = None
    links: Optional[List[Dict[str, str]]] = None
    media: Optional[List[Dict[str, str]]] = None

class AdvancedWebResearcher:
    """Advanced web research with Crawl4AI for LLM-friendly extraction"""
    
    def __init__(self, max_results: int = 5, max_content_length: int = 3000):
        """Initialize advanced web researcher with Crawl4AI"""
        self.max_results = max_results
        self.max_content_length = max_content_length
        
        # Crawl4AI configuration
        self.browser_config = BrowserConfig(
            headless=True,
            verbose=False,
            browser_type="chromium"  # or "firefox", "webkit"
        )
        
        self.crawler_config = CrawlerRunConfig(
            wait_for_images=False,
            delay_before_return_html=2.0,
            page_timeout=15000  # 15 second timeout
        )
        
        # Initialize search engines
        self.search_engines = {
            'duckduckgo': self._search_duckduckgo,
        }
    
    async def check_url_exists(self, url: str, rag_system) -> Optional[Dict[str, Any]]:
        """Check if URL already exists in web_research collection"""
        try:
            # Search for existing URL in web_research collection
            search_results = await rag_system.search_knowledge_base(
                query=url,
                collections=['web_research'],
                limit=10,
                min_score=0.1  # Low score to catch any URL matches
            )
            
            # Check if any result has the exact URL
            for result in search_results:
                stored_url = result.metadata.get('url', '')
                if stored_url == url:
                    logger.info(f"🔄 Found existing content for: {url}")
                    return {
                        'title': result.metadata.get('title', 'Existing Content'),
                        'content': result.content,
                        'metadata': result.metadata,
                        'from_cache': True
                    }
            
            return None
            
        except Exception as e:
            logger.warning(f"Error checking URL cache: {e}")
            return None
    
    async def research_query(self, 
                           query: str,
                           search_engine: str = 'duckduckgo',
                           scrape_content: bool = True,
                           extraction_strategy: str = 'simple',
                           rag_system=None) -> List[WebSearchResult]:
        """Research a query with advanced Crawl4AI extraction"""
        
        logger.info(f"🔍 Advanced researching: {query}")
        
        # Search for results
        search_func = self.search_engines.get(search_engine, self._search_duckduckgo)
        search_results = await search_func(query)
        
        if not search_results:
            logger.warning(f"No search results found for: {query}")
            return []
        
        # Scrape content with Crawl4AI if requested
        if scrape_content:
            scraped_results = []
            
            async with AsyncWebCrawler(config=self.browser_config) as crawler:
                for result in search_results[:self.max_results]:
                    try:
                        # DuckDuckGo returns 'href' field for URL
                        url = result.get('href') or result.get('url')
                        if not url:
                            logger.warning(f"No URL found in result: {result}")
                            continue
                        
                        # Check if URL already exists in collection
                        existing_content = None
                        if rag_system:
                            existing_content = await self.check_url_exists(url, rag_system)
                        
                        if existing_content:
                            # Use existing content instead of re-scraping
                            logger.info(f"♻️ Using cached content for: {url}")
                            web_result = WebSearchResult(
                                title=existing_content['title'],
                                url=url,
                                snippet=result.get('body', ''),
                                content=existing_content['content'],
                                timestamp=time.time(),
                                source="web_search_cached",
                                structured_data=existing_content['metadata'].get('has_structured_data'),
                                markdown_content=existing_content['metadata'].get('has_markdown'),
                                links=None,  # Could extract from metadata if needed
                                media=None   # Could extract from metadata if needed
                            )
                            scraped_results.append(web_result)
                        else:
                            # Scrape fresh content
                            scraped = await self._crawl_url_advanced(
                                crawler, 
                                url, 
                                query,
                                extraction_strategy
                            )
                            
                            if scraped.success:
                                web_result = WebSearchResult(
                                    title=scraped.title,
                                    url=scraped.url,
                                    snippet=result.get('body', ''),
                                    content=scraped.content,
                                    timestamp=time.time(),
                                    source="web_search_crawl4ai",
                                    structured_data=scraped.structured_data,
                                    markdown_content=scraped.markdown_content,
                                    links=scraped.links,
                                    media=scraped.media
                                )
                                scraped_results.append(web_result)
                            
                    except Exception as e:
                        logger.error(f"Failed to crawl {result.get('href', 'unknown URL')}: {e}")
                        continue
            
            logger.info(f"✅ Successfully scraped {len(scraped_results)} results")
            return scraped_results
        
        else:
            # Return search results without scraping
            return [
                WebSearchResult(
                    title=result.get('title', 'No Title'),
                    url=result.get('href', result.get('url', '')),
                    snippet=result.get('body', ''),
                    content=result.get('body', ''),
                    timestamp=time.time(),
                    source="web_search"
                )
                for result in search_results[:self.max_results]
                if result.get('href') or result.get('url')  # Only include results with URLs
            ]
    
    async def _search_duckduckgo(self, query: str) -> List[Dict[str, str]]:
        """Search DuckDuckGo for results"""
        try:
            logger.info(f"🦆 Searching DuckDuckGo: {query}")
            
            # First, optimize query using AI and detect language
            optimized_query, detected_language = await self._optimize_search_query_with_ai(query)
            logger.info(f"🤖 AI-optimized query: {optimized_query}")
            logger.info(f"🌍 Detected language/region: {detected_language}")
            
            # Also clean query as fallback
            cleaned_query = self._clean_search_query(query)
            logger.info(f"🧹 Cleaned query: {cleaned_query}")
            
            # Use DDGS for search with retry logic
            ddgs = DDGS()
            results = []
            
            # Try the AI-optimized query first with detected language
            try:
                results = list(ddgs.text(optimized_query, max_results=self.max_results, region=detected_language))
            except Exception as e:
                logger.warning(f"AI-optimized search failed: {e}")
                
                # Fallback: try cleaned query with detected language
                try:
                    results = list(ddgs.text(cleaned_query, max_results=self.max_results, region=detected_language))
                except Exception as e2:
                    logger.warning(f"Cleaned query search failed: {e2}")
                    
                    # Fallback: try original query with detected language
                    try:
                        results = list(ddgs.text(query, max_results=self.max_results, region=detected_language))
                    except Exception as e3:
                        logger.warning(f"Original query search failed: {e3}")
                        
                        # Fallback: try with US English as last resort
                        try:
                            results = list(ddgs.text(optimized_query, max_results=self.max_results, region='us-en'))
                        except Exception as e4:
                            logger.warning(f"US English fallback failed: {e4}")
                            
                            # Last resort: try with minimal query, no region
                            simple_query = query.split()[:3]  # Take first 3 words
                            if simple_query:
                                try:
                                    results = list(ddgs.text(' '.join(simple_query), max_results=self.max_results))
                                except Exception as e5:
                                    logger.error(f"All search attempts failed: {e5}")
            
            logger.info(f"Found {len(results)} DuckDuckGo results")
            
            # Debug: Log the actual results to see what we're getting
            for i, result in enumerate(results):
                logger.info(f"🔍 Result {i+1}: {result.get('title', 'No title')} - {result.get('href', 'No URL')}")
                logger.info(f"   Snippet: {result.get('body', 'No snippet')[:100]}...")
            
            return results
            
        except Exception as e:
            logger.error(f"DuckDuckGo search failed: {e}")
            return []
    
    async def _optimize_search_query_with_ai(self, query: str) -> tuple[str, str]:
        """Use OpenAI to optimize search query and detect language for better results"""
        try:
            # Get OpenAI API key
            api_key = os.getenv("OPENAI_API_KEY")
            if not api_key:
                logger.warning("No OpenAI API key found, skipping AI optimization")
                return query, 'us-en'
            
            # Create OpenAI client
            client = openai.AsyncOpenAI(api_key=api_key)
            
            # Enhanced system prompt with language detection
            system_prompt = """You are a search query optimization expert with language detection capabilities. Your job is to:
1. Detect the language of the user query
2. Transform the query into the most effective search terms for web search engines like DuckDuckGo
3. Return both the optimized query and the detected language

LANGUAGE DETECTION:
Detect the primary language and return the appropriate DuckDuckGo region code:
- English → us-en
- Italian → it-it  
- Spanish → es-es
- French → fr-fr
- German → de-de
- Portuguese → pt-br
- Japanese → jp-jp
- Chinese → cn-zh
- Russian → ru-ru
- Dutch → nl-nl
- Korean → kr-kr

OPTIMIZATION RULES:
1. Remove question words in ANY language (how/come, what/cosa/qué, when/quando/cuándo, where/dove/dónde, why/perché/por qué, who/chi/quién, is/è/es, are/sono/son, was/era, were/erano/eran)
2. Extract key entities, names, concepts, and topics
3. Add relevant synonyms and alternative terms IN THE SAME LANGUAGE
4. Add specific terms that would appear on relevant web pages
5. Remove conversational language and make it more direct
6. Consider adding context terms that help find authoritative sources
7. Keep the optimized query in the SAME LANGUAGE as the original

EXAMPLES:
- "How is Ilaria Loconte?" → "Ilaria Loconte biography profile information" | us-en
- "Come sta Ilaria Loconte?" → "Ilaria Loconte biografia profilo informazioni" | it-it
- "¿Cómo está Ilaria Loconte?" → "Ilaria Loconte biografía perfil información" | es-es
- "What is artificial intelligence?" → "artificial intelligence AI definition overview technology" | us-en
- "Cos'è l'intelligenza artificiale?" → "intelligenza artificiale IA definizione panoramica tecnologia" | it-it
- "¿Qué es la inteligencia artificial?" → "inteligencia artificial IA definición tecnología" | es-es

RESPONSE FORMAT:
Return ONLY in this exact JSON format:
{"query": "optimized search query", "language": "region-code"}"""

            user_prompt = f"Optimize this search query and detect its language: {query}"
            
            # Make API call
            response = await client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_prompt}
                ],
                max_tokens=150,
                temperature=0.2,
                timeout=10.0
            )
            
            result_text = response.choices[0].message.content.strip()
            
            # Parse JSON response
            try:
                result = json.loads(result_text)
                optimized_query = result.get('query', query).strip()
                detected_language = result.get('language', 'us-en')
                
                # Validate the optimized query
                if len(optimized_query) < 2 or len(optimized_query) > 200:
                    logger.warning(f"AI optimization produced invalid query length: {len(optimized_query)}")
                    return query, 'us-en'
                
                # Remove any quotes or special formatting that might break the search
                optimized_query = optimized_query.replace('"', '').replace("'", "").strip()
                
                # Validate language code
                valid_languages = ['us-en', 'it-it', 'es-es', 'fr-fr', 'de-de', 'pt-br', 'jp-jp', 'cn-zh', 'ru-ru', 'nl-nl', 'kr-kr']
                if detected_language not in valid_languages:
                    detected_language = 'us-en'
                
                logger.info(f"🌍 Detected language: {detected_language}")
                return optimized_query, detected_language
                
            except json.JSONDecodeError:
                logger.warning(f"Failed to parse AI response as JSON: {result_text}")
                # Fallback: assume it's just the optimized query
                return result_text.strip(), 'us-en'
            
        except Exception as e:
            logger.warning(f"AI query optimization failed: {e}")
            return query, 'us-en'  # Return original query with default language
    
    def _clean_search_query(self, query: str) -> str:
        """Clean and optimize search query for better results"""
        # Remove question words that don't help with search
        question_words = ['how', 'what', 'when', 'where', 'why', 'who', 'is', 'are', 'was', 'were']
        
        # Split query into words
        words = query.lower().split()
        
        # Remove question words from the beginning
        while words and words[0] in question_words:
            words.pop(0)
        
        # Remove question marks and other punctuation
        cleaned = ' '.join(words).replace('?', '').replace('!', '').strip()
        
        # If query becomes too short, return original
        if len(cleaned) < 3:
            return query
            
        return cleaned
    
    async def _crawl_url_advanced(self, 
                                 crawler: AsyncWebCrawler,
                                 url: str, 
                                 query: str = None,
                                 extraction_strategy: str = 'simple') -> ScrapedContent:
        """Advanced URL crawling with Crawl4AI"""
        try:
            logger.info(f"🕷️ Crawling: {url}")
            
            # Choose extraction strategy
            strategy = None
            if extraction_strategy == 'cosine' and query:
                # Use more lenient settings to avoid empty distance matrix errors
                strategy = CosineStrategy(
                    semantic_filter=query,
                    word_count_threshold=5,  # Lower threshold to capture more content
                    max_dist=0.4,            # Higher distance tolerance
                    linkage_method="ward",
                    top_k=5                  # Get more results
                )
            elif extraction_strategy == 'llm' and query:
                # LLM-based extraction (requires OpenAI API key)
                strategy = LLMExtractionStrategy(
                    provider="openai/gpt-4o-mini",
                    api_token=os.getenv("OPENAI_API_KEY"),
                    instruction=f"Extract the most relevant information related to: {query}"
                )
            elif extraction_strategy == 'simple':
                # No extraction strategy - just use cleaned HTML (most reliable)
                strategy = None
            
            # Create run config with extraction strategy
            run_config = CrawlerRunConfig(
                extraction_strategy=strategy,
                cache_mode=CacheMode.BYPASS,
                process_iframes=True,
                remove_overlay_elements=True,
                wait_for_images=False,
                delay_before_return_html=2.0,
                page_timeout=15000  # 15 second timeout
            )
            
            # Crawl the URL with error handling for scipy distance matrix issues
            try:
                result = await crawler.arun(url=url, config=run_config)
                
                if not result.success:
                    raise Exception(f"Crawl failed: {result.error_message}")
                
                # Extract content
                content = result.extracted_content if result.extracted_content else result.cleaned_html
                markdown_content = result.markdown if result.markdown else None
                
            except Exception as e:
                # Handle scipy distance matrix errors by falling back to simple extraction
                if "empty distance matrix" in str(e) or "observations cannot be determined" in str(e):
                    logger.warning(f"⚠️ CosineStrategy failed for {url}, falling back to simple extraction: {e}")
                    
                    # Retry without extraction strategy (simple HTML cleaning)
                    simple_config = CrawlerRunConfig(
                        cache_mode=CacheMode.BYPASS,
                        process_iframes=True,
                        remove_overlay_elements=True,
                        wait_for_images=False,
                        delay_before_return_html=2.0,
                        page_timeout=15000
                    )
                    
                    result = await crawler.arun(url=url, config=simple_config)
                    
                    if not result.success:
                        raise Exception(f"Crawl failed even with simple extraction: {result.error_message}")
                    
                    # Use cleaned HTML as content
                    content = result.cleaned_html
                    markdown_content = result.markdown if result.markdown else None
                    
                    logger.info(f"✅ Successfully extracted content using simple method for {url}")
                else:
                    # Re-raise other errors
                    raise
            
            # Get title from metadata
            title = result.metadata.get('title', 'No Title') if result.metadata else 'No Title'
            
            # Limit content length
            if len(content) > self.max_content_length:
                content = content[:self.max_content_length] + "..."
            
            # Extract structured data
            structured_data = None
            if result.extracted_content:
                try:
                    import json
                    if isinstance(result.extracted_content, str):
                        structured_data = {"extracted": result.extracted_content}
                    else:
                        structured_data = result.extracted_content
                except:
                    structured_data = {"raw": str(result.extracted_content)}
            
            # Extract links
            links = []
            if result.links:
                # Handle different link formats
                if isinstance(result.links, dict):
                    internal_links = result.links.get("internal", [])
                elif isinstance(result.links, list):
                    internal_links = result.links
                else:
                    internal_links = []
                
                links = [
                    {"text": link.get("text", ""), "url": link.get("href", "")}
                    for link in internal_links[:10]  # Limit links
                ]
            
            # Extract media
            media = []
            if result.media:
                # Handle different media formats
                if isinstance(result.media, dict):
                    images = result.media.get("images", [])
                elif isinstance(result.media, list):
                    images = result.media
                else:
                    images = []
                
                media = [
                    {"type": "image", "url": img.get("src", ""), "alt": img.get("alt", "")}
                    for img in images[:5]  # Limit images
                ]
            
            metadata = {
                'url': url,
                'title': title,
                'scraped_at': time.time(),
                'content_length': len(content),
                'domain': self._extract_domain(url),
                'crawl4ai_version': 'advanced',
                'extraction_strategy': extraction_strategy,
                'success_metrics': {
                    'html_length': len(result.html) if result.html else 0,
                    'cleaned_html_length': len(result.cleaned_html) if result.cleaned_html else 0,
                    'links_found': len(links),
                    'media_found': len(media)
                }
            }
            
            return ScrapedContent(
                url=url,
                title=title,
                content=content,
                metadata=metadata,
                success=True,
                markdown_content=markdown_content,
                structured_data=structured_data,
                links=links,
                media=media
            )
            
        except Exception as e:
            logger.error(f"Failed to crawl {url}: {e}")
            # Provide user-friendly error messages
            error_msg = str(e)
            if "distance matrix" in error_msg.lower():
                error_msg = "Content extraction failed (empty content)"
            elif "timeout" in error_msg.lower():
                error_msg = "Page took too long to load"
            elif "connection" in error_msg.lower():
                error_msg = "Could not connect to page"
            elif "blocked" in error_msg.lower() or "403" in error_msg:
                error_msg = "Access blocked by website"
            elif "404" in error_msg:
                error_msg = "Page not found"
            else:
                error_msg = "Failed to read page"
                
            return ScrapedContent(
                url=url,
                title="Reading Failed",
                content="",
                metadata={'url': url, 'error': error_msg},
                success=False,
                error=error_msg
            )
    
    def _extract_domain(self, url: str) -> str:
        """Extract domain from URL"""
        try:
            from urllib.parse import urlparse
            return urlparse(url).netloc
        except:
            return "unknown"
    
    async def crawl_single_url(self, 
                              url: str, 
                              query: str = None,
                              extraction_strategy: str = 'cosine') -> ScrapedContent:
        """Crawl a single URL with advanced extraction"""
        async with AsyncWebCrawler(config=self.browser_config) as crawler:
            return await self._crawl_url_advanced(crawler, url, query, extraction_strategy)
    
    async def batch_crawl_urls(self, 
                              urls: List[str], 
                              query: str = None,
                              extraction_strategy: str = 'cosine') -> List[ScrapedContent]:
        """Batch crawl multiple URLs"""
        results = []
        
        async with AsyncWebCrawler(config=self.browser_config) as crawler:
            for url in urls:
                try:
                    result = await self._crawl_url_advanced(crawler, url, query, extraction_strategy)
                    results.append(result)
                except Exception as e:
                    logger.error(f"Failed to crawl {url}: {e}")
                    results.append(ScrapedContent(
                        url=url,
                        title="Crawling Failed",
                        content="",
                        metadata={'url': url, 'error': str(e)},
                        success=False,
                        error=str(e)
                    ))
        
        return results
    
    async def stream_research_and_store(self, 
                                      query: str,
                                      rag_system,
                                      search_engine: str = 'duckduckgo'):
        """Streaming research workflow with Crawl4AI: search, scrape, and store with real-time updates"""
        
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
                        'url': result['href'],
                        'snippet': result['body']
                    }
                    for result in search_results
                ]
            }
            
            # Step 2: Crawl content with Crawl4AI and store (streaming)
            scraped_results = []
            stored_count = 0
            
            async with AsyncWebCrawler(config=self.browser_config) as crawler:
                for i, result in enumerate(search_results[:self.max_results]):
                    try:
                        url = result['href']
                        
                        # Check if URL already exists in collection
                        existing_content = await self.check_url_exists(url, rag_system)
                        
                        if existing_content:
                            # Yield cached content message
                            yield {
                                'type': 'stored',
                                'message': f'♻️ Using cached: {existing_content["title"][:50]}...',
                                'stored_count': stored_count + 1,
                                'url': url,
                                'title': existing_content['title'],
                                'content': existing_content['content'],
                                'from_cache': True,
                                'features': {
                                    'has_structured_data': existing_content['metadata'].get('has_structured_data', False),
                                    'has_markdown': existing_content['metadata'].get('has_markdown', False),
                                    'links_found': existing_content['metadata'].get('links_count', 0),
                                    'media_found': existing_content['metadata'].get('media_count', 0)
                                }
                            }
                            
                            # Create result from cached content
                            web_result = WebSearchResult(
                                title=existing_content['title'],
                                url=url,
                                snippet=result.get('body', ''),
                                content=existing_content['content'],
                                timestamp=time.time(),
                                source="web_search_cached",
                                structured_data=existing_content['metadata'].get('has_structured_data'),
                                markdown_content=existing_content['metadata'].get('has_markdown'),
                                links=None,
                                media=None
                            )
                            scraped_results.append(web_result)
                            stored_count += 1
                            
                        else:
                            # Yield progress update for fresh scraping
                            yield {
                                'type': 'progress',
                                'message': f'Reading page {i+1}/{min(len(search_results), self.max_results)}: {result["title"][:50]}...',
                                'current': i+1,
                                'total': min(len(search_results), self.max_results),
                                'url': url,
                                'title': result['title']
                            }
                            
                            # Scrape fresh content
                            scraped = await self._crawl_url_advanced(
                                crawler, 
                                url, 
                                query,
                                'cosine'
                            )
                            
                            if scraped.success:
                                web_result = WebSearchResult(
                                    title=scraped.title,
                                    url=scraped.url,
                                    snippet=result.get('body', ''),
                                    content=scraped.content,
                                    timestamp=time.time(),
                                    source="web_search_crawl4ai",
                                    structured_data=scraped.structured_data,
                                    markdown_content=scraped.markdown_content,
                                    links=scraped.links,
                                    media=scraped.media
                                )
                                scraped_results.append(web_result)
                                
                                # Store in VittoriaDB
                                stored_id = await self._store_single_result_enhanced(web_result, rag_system, query)
                                if stored_id:
                                    stored_count += 1
                                    
                                    # Yield storage success with URL details and content
                                    yield {
                                        'type': 'stored',
                                        'message': f'✅ Finished reading: {scraped.title[:50]}...',
                                        'stored_count': stored_count,
                                        'url': url,
                                        'title': scraped.title,
                                        'content': scraped.content,
                                        'features': {
                                            'has_structured_data': bool(scraped.structured_data),
                                            'has_markdown': bool(scraped.markdown_content),
                                            'links_found': len(scraped.links) if scraped.links else 0,
                                            'media_found': len(scraped.media) if scraped.media else 0
                                        }
                                    }
                            else:
                                # Handle failed scraping - yield error status
                                error_msg = scraped.error or "Failed to read page content"
                                yield {
                                    'type': 'warning',
                                    'message': f'❌ Failed to read: {result["title"][:50]}... ({error_msg})',
                                    'url': url,
                                    'title': result['title'],
                                    'error': error_msg
                                }
                                
                    except Exception as e:
                        logger.warning(f"Failed to process {result['href']}: {e}")
                        yield {
                            'type': 'warning',
                            'message': f'❌ Failed to read: {result["title"][:50]}...',
                            'url': result['href'],
                            'title': result['title']
                        }
                        continue
            
            # Final completion update
            yield {
                'type': 'complete',
                'total_results': len(scraped_results),
                'stored_count': stored_count,
                'message': f'Finished reading {stored_count} pages'
            }
            
        except Exception as e:
            yield {
                'type': 'error',
                'message': str(e)
            }
    
    async def _store_single_result_enhanced(self, result: WebSearchResult, rag_system, query: str) -> Optional[str]:
        """Store a single enhanced web search result with Crawl4AI features"""
        try:
            # Create enhanced content with Crawl4AI data
            content_parts = [
                f"Title: {result.title}",
                f"URL: {result.url}",
                f"Query: {query}",
                "",
                "Content:",
                result.content
            ]
            
            if result.snippet:
                content_parts.extend(["", "Snippet:", result.snippet])
            
            if result.structured_data:
                content_parts.extend(["", "Structured Data:", str(result.structured_data)])
            
            if result.markdown_content:
                content_parts.extend(["", "Markdown Content:", result.markdown_content[:1000] + "..."])
            
            if result.links:
                links_text = "\n".join([f"- [{link['text']}]({link['url']})" for link in result.links[:5]])
                content_parts.extend(["", "Related Links:", links_text])
            
            enhanced_content = "\n".join(content_parts)
            
            # Generate unique ID
            content_hash = hashlib.md5(f"{result.url}_{query}_crawl4ai".encode()).hexdigest()
            doc_id = f"web_crawl4ai_{content_hash}"
            
            # Enhanced metadata
            metadata = {
                "title": result.title,
                "document_title": result.title,
                "url": result.url,
                "source": result.source,
                "source_collection": "web_search",
                "query": query,
                "timestamp": result.timestamp,
                "content_length": len(result.content),
                "has_structured_data": bool(result.structured_data),
                "has_markdown": bool(result.markdown_content),
                "links_count": len(result.links) if result.links else 0,
                "media_count": len(result.media) if result.media else 0,
                "extraction_method": "crawl4ai_cosine",
                "type": "web_research_crawl4ai"
            }
            
            # Store in web_research collection
            stored_doc_id = await rag_system.add_document(
                content=enhanced_content,
                metadata=metadata,
                collection_name="web_research"
            )
            
            return stored_doc_id
            
        except Exception as e:
            logger.error(f"Failed to store enhanced result: {e}")
            return None

# Convenience function for easy integration
async def research_with_crawl4ai(query: str, 
                                max_results: int = 5,
                                extraction_strategy: str = 'cosine') -> List[WebSearchResult]:
    """Quick research function using Crawl4AI"""
    researcher = AdvancedWebResearcher(max_results=max_results)
    return await researcher.research_query(
        query=query, 
        scrape_content=True,
        extraction_strategy=extraction_strategy
    )

# Example usage
if __name__ == "__main__":
    async def test_crawl4ai():
        """Test the Crawl4AI integration"""
        researcher = AdvancedWebResearcher(max_results=3)
        
        # Test research
        results = await researcher.research_query(
            "machine learning vector databases",
            extraction_strategy='cosine'
        )
        
        for result in results:
            print(f"\n🔍 Title: {result.title}")
            print(f"🌐 URL: {result.url}")
            print(f"📄 Content length: {len(result.content)}")
            print(f"🔗 Links found: {len(result.links) if result.links else 0}")
            print(f"🖼️ Media found: {len(result.media) if result.media else 0}")
            if result.structured_data:
                print(f"📊 Structured data available: {len(str(result.structured_data))}")
    
    # Run test
    asyncio.run(test_crawl4ai())
