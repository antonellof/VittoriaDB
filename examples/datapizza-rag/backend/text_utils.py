"""
Text processing utilities
Simple helper functions for text manipulation
"""

import re
from typing import List


def chunk_text(text: str, max_tokens: int = 6000, overlap: int = 200) -> List[str]:
    """
    Split text into chunks that fit within token limits.
    
    Args:
        text: Text to chunk
        max_tokens: Maximum tokens per chunk (conservative estimate: ~4 chars = 1 token)
        overlap: Number of characters to overlap between chunks
    
    Returns:
        List of text chunks
    """
    if not text or len(text) == 0:
        return []
    
    # Conservative estimate: 4 characters ≈ 1 token
    max_chars = max_tokens * 4
    
    # If text is short enough, return as single chunk
    if len(text) <= max_chars:
        return [text]
    
    chunks = []
    start = 0
    
    while start < len(text):
        # Calculate end position
        end = start + max_chars
        
        # If this is not the last chunk, try to break at a sentence or paragraph
        if end < len(text):
            # Look for sentence endings within the last 500 characters
            search_start = max(start + max_chars - 500, start)
            
            # Try to find sentence endings
            sentence_endings = []
            for match in re.finditer(r'[.!?]\s+', text[search_start:end]):
                sentence_endings.append(search_start + match.end())
            
            if sentence_endings:
                end = sentence_endings[-1]  # Use the last sentence ending
            else:
                # Try to break at paragraph
                paragraph_breaks = []
                for match in re.finditer(r'\n\s*\n', text[search_start:end]):
                    paragraph_breaks.append(search_start + match.start())
                
                if paragraph_breaks:
                    end = paragraph_breaks[-1]
                else:
                    # Try to break at word boundary
                    word_boundary = text.rfind(' ', search_start, end)
                    if word_boundary > start:
                        end = word_boundary
        
        # Extract chunk
        chunk = text[start:end].strip()
        if chunk:
            chunks.append(chunk)
        
        # Move start position with overlap
        start = max(end - overlap, start + 1)
        
        # Prevent infinite loop
        if start >= len(text):
            break
    
    return chunks

