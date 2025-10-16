# Code Cleanup Plan

## Analysis Summary

### Frontend Usage (from web-ui-rag/frontend)
The frontend **ONLY** calls these endpoints:
1. `/rag/stream` - Main chat with RAG
2. `/upload` - File upload
3. `/github/index` - Index GitHub repos
4. `/research/stream` - Web research with streaming
5. `/chat/sessions` - Create chat session
6. `/chat/save` - Save chat history
7. `/cancel` - Cancel operations

### Current Issues

#### 1. Duplicate RAG Implementations
- `rag_engine.py` (989 lines) - Legacy, barely used
- `rag_system.py` (1202 lines) - Actually used by most endpoints
- **Problem**: 90% functionality overlap, causing confusion and maintenance burden

#### 2. Unused Endpoints (to remove)
From main.py (27 endpoints total, only 7 used):
- `/chat` - Unused, frontend uses `/rag/stream`
- `/chat/stream` - Unused, frontend uses `/rag/stream`
- `/rag/stats` - Only uses rag_engine (can merge with `/stats`)
- `/research` - Unused, frontend uses `/research/stream`
- `/research/legacy` - Legacy code
- `/research/crawl4ai` - Duplicate of `/research/stream`
- `/search` - Not called by frontend
- `/search/more` - Not called by frontend
- `/debug/optimize-query` - Debug endpoint
- `/documents/{collection_name}/original` - Not used
- `/documents/{collection_name}` - Not used
- `/delete/documents/{collection_name}/{document_id}` - Not used
- `/delete/documents/{collection_name}` - Not used
- `/rag/document` - Not used
- `/chat/sessions/{session_id}/history` - Not used

#### 3. Code to Consolidate
- Both rag_engine and rag_system have:
  - Document chunking
  - Embedding generation
  - Search functionality
  - Response generation
  - Collection management

### Cleanup Actions

#### Phase 1: Remove rag_engine.py (Legacy Code)
- Remove `/Users/d695663/Desktop/Dev/CognitoraVector/examples/datapizza-rag/backend/rag_engine.py`
- Update main.py to remove rag_engine imports and initialization
- Keep only rag_system.py as the single source of truth

#### Phase 2: Remove Unused Endpoints
Remove from main.py:
1. `/chat` (line ~419)
2. `/chat/stream` (line ~829)
3. `/rag/stats` (line ~373) - merge into `/stats`
4. `/research` (line ~1786)
5. `/research/legacy` (line ~1953)
6. `/research/crawl4ai` (line ~1977)
7. `/search` (line ~2171)
8. `/search/more` (line ~2239)
9. `/debug/optimize-query` (line ~2303)
10. `/documents/{collection_name}/original` (line ~2327)
11. `/documents/{collection_name}` (line ~2346)
12. `/delete/documents` endpoints (lines ~2366, ~2394)
13. `/rag/document` (line ~1434)
14. `/chat/sessions/{session_id}/history` (line ~2692)

#### Phase 3: Simplify rag_system.py
- Remove unused helper functions
- Clean up imports
- Remove dead code paths

### Expected Results
- **Before**: 2768 lines in main.py, 989 lines in rag_engine.py, 1202 lines in rag_system.py
- **After**: ~1500 lines in main.py, 0 lines in rag_engine.py (deleted), ~1000 lines in rag_system.py
- **Total reduction**: ~2500 lines (~50% reduction)

