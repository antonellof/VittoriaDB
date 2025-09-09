package processor

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// SentenceChunker implements sentence-aware chunking
type SentenceChunker struct{}

// NewSentenceChunker creates a new sentence-based chunker
func NewSentenceChunker() *SentenceChunker {
	return &SentenceChunker{}
}

// ChunkText splits text into chunks at sentence boundaries
func (c *SentenceChunker) ChunkText(text string, config *ProcessingConfig) ([]DocumentChunk, error) {
	if text == "" {
		return []DocumentChunk{}, nil
	}

	// Split text into sentences
	sentences := c.splitIntoSentences(text)
	if len(sentences) == 0 {
		return []DocumentChunk{}, nil
	}

	var chunks []DocumentChunk
	var currentChunk strings.Builder
	var currentSize int
	chunkIndex := 0

	for i, sentence := range sentences {
		sentenceLen := len(sentence)

		// If adding this sentence would exceed chunk size, finalize current chunk
		if currentSize > 0 && currentSize+sentenceLen > config.ChunkSize {
			if currentChunk.Len() >= config.MinChunkSize {
				chunk := DocumentChunk{
					ID:       fmt.Sprintf("chunk_%d", chunkIndex),
					Content:  strings.TrimSpace(currentChunk.String()),
					Position: chunkIndex,
					Size:     currentChunk.Len(),
					Metadata: map[string]string{
						"chunk_type": "sentence",
						"sentences":  fmt.Sprintf("%d", countSentences(currentChunk.String())),
					},
				}
				chunks = append(chunks, chunk)
				chunkIndex++
			}

			// Start new chunk with overlap
			currentChunk.Reset()
			currentSize = 0

			// Add overlap from previous sentences if configured
			if config.ChunkOverlap > 0 && len(chunks) > 0 {
				overlapText := c.getOverlapText(sentences, i, config.ChunkOverlap)
				currentChunk.WriteString(overlapText)
				currentSize = len(overlapText)
			}
		}

		// Add sentence to current chunk
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
			currentSize++
		}
		currentChunk.WriteString(sentence)
		currentSize += sentenceLen
	}

	// Add final chunk if it has content
	if currentChunk.Len() >= config.MinChunkSize {
		chunk := DocumentChunk{
			ID:       fmt.Sprintf("chunk_%d", chunkIndex),
			Content:  strings.TrimSpace(currentChunk.String()),
			Position: chunkIndex,
			Size:     currentChunk.Len(),
			Metadata: map[string]string{
				"chunk_type": "sentence",
				"sentences":  fmt.Sprintf("%d", countSentences(currentChunk.String())),
			},
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// splitIntoSentences splits text into sentences using regex
func (c *SentenceChunker) splitIntoSentences(text string) []string {
	// Regex pattern for sentence boundaries
	sentencePattern := regexp.MustCompile(`[.!?]+\s+`)

	// Split by sentence boundaries
	parts := sentencePattern.Split(text, -1)

	var sentences []string
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Add back the punctuation (except for the last part)
		if i < len(parts)-1 {
			// Find the punctuation that was removed
			remaining := text[len(strings.Join(parts[:i+1], "")):]
			if len(remaining) > 0 {
				match := sentencePattern.FindString(remaining)
				if match != "" {
					part += strings.TrimSpace(match)
				}
			}
		}

		sentences = append(sentences, part)
	}

	return sentences
}

// getOverlapText gets overlap text from previous sentences
func (c *SentenceChunker) getOverlapText(sentences []string, currentIndex, overlapSize int) string {
	if currentIndex == 0 || overlapSize <= 0 {
		return ""
	}

	var overlap strings.Builder
	overlapChars := 0

	// Go backwards from current sentence to build overlap
	for i := currentIndex - 1; i >= 0 && overlapChars < overlapSize; i-- {
		sentence := sentences[i]
		if overlapChars+len(sentence) <= overlapSize {
			if overlap.Len() > 0 {
				overlap.WriteString(" ")
			}
			overlap.WriteString(sentence)
			overlapChars += len(sentence) + 1
		} else {
			break
		}
	}

	return overlap.String()
}

// ParagraphChunker implements paragraph-aware chunking
type ParagraphChunker struct{}

// NewParagraphChunker creates a new paragraph-based chunker
func NewParagraphChunker() *ParagraphChunker {
	return &ParagraphChunker{}
}

// ChunkText splits text into chunks at paragraph boundaries
func (c *ParagraphChunker) ChunkText(text string, config *ProcessingConfig) ([]DocumentChunk, error) {
	if text == "" {
		return []DocumentChunk{}, nil
	}

	// Split text into paragraphs
	paragraphs := strings.Split(text, "\n\n")

	var chunks []DocumentChunk
	var currentChunk strings.Builder
	chunkIndex := 0

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// If adding this paragraph would exceed chunk size, finalize current chunk
		if currentChunk.Len() > 0 && currentChunk.Len()+len(paragraph) > config.ChunkSize {
			if currentChunk.Len() >= config.MinChunkSize {
				chunk := DocumentChunk{
					ID:       fmt.Sprintf("chunk_%d", chunkIndex),
					Content:  strings.TrimSpace(currentChunk.String()),
					Position: chunkIndex,
					Size:     currentChunk.Len(),
					Metadata: map[string]string{
						"chunk_type": "paragraph",
						"paragraphs": fmt.Sprintf("%d", strings.Count(currentChunk.String(), "\n\n")+1),
					},
				}
				chunks = append(chunks, chunk)
				chunkIndex++
			}

			// Start new chunk
			currentChunk.Reset()
		}

		// Add paragraph to current chunk
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(paragraph)
	}

	// Add final chunk if it has content
	if currentChunk.Len() >= config.MinChunkSize {
		chunk := DocumentChunk{
			ID:       fmt.Sprintf("chunk_%d", chunkIndex),
			Content:  strings.TrimSpace(currentChunk.String()),
			Position: chunkIndex,
			Size:     currentChunk.Len(),
			Metadata: map[string]string{
				"chunk_type": "paragraph",
				"paragraphs": fmt.Sprintf("%d", strings.Count(currentChunk.String(), "\n\n")+1),
			},
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// FixedSizeChunker implements fixed-size chunking
type FixedSizeChunker struct{}

// NewFixedSizeChunker creates a new fixed-size chunker
func NewFixedSizeChunker() *FixedSizeChunker {
	return &FixedSizeChunker{}
}

// ChunkText splits text into fixed-size chunks
func (c *FixedSizeChunker) ChunkText(text string, config *ProcessingConfig) ([]DocumentChunk, error) {
	if text == "" {
		return []DocumentChunk{}, nil
	}

	var chunks []DocumentChunk
	chunkIndex := 0

	for i := 0; i < len(text); i += config.ChunkSize - config.ChunkOverlap {
		end := i + config.ChunkSize
		if end > len(text) {
			end = len(text)
		}

		chunkText := text[i:end]

		// Skip chunks that are too small (unless it's the last chunk)
		if len(chunkText) < config.MinChunkSize && end < len(text) {
			continue
		}

		chunk := DocumentChunk{
			ID:       fmt.Sprintf("chunk_%d", chunkIndex),
			Content:  chunkText,
			Position: chunkIndex,
			Size:     len(chunkText),
			Metadata: map[string]string{
				"chunk_type": "fixed_size",
				"start_pos":  fmt.Sprintf("%d", i),
				"end_pos":    fmt.Sprintf("%d", end),
			},
		}
		chunks = append(chunks, chunk)
		chunkIndex++

		// Break if we've reached the end
		if end >= len(text) {
			break
		}
	}

	return chunks, nil
}

// Helper functions

// countSentences counts the number of sentences in text
func countSentences(text string) int {
	sentencePattern := regexp.MustCompile(`[.!?]+`)
	return len(sentencePattern.FindAllString(text, -1))
}

// cleanText removes extra whitespace and normalizes text
func cleanText(text string) string {
	// Remove extra whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Remove non-printable characters except newlines and tabs
	text = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' {
			return r
		}
		return -1
	}, text)

	return strings.TrimSpace(text)
}

// GetChunker returns the appropriate chunker based on strategy
func GetChunker(strategy string) ChunkingStrategy {
	switch strategy {
	case "sentence":
		return NewSentenceChunker()
	case "paragraph":
		return NewParagraphChunker()
	case "fixed_size":
		return NewFixedSizeChunker()
	default:
		return NewSentenceChunker() // Default to sentence chunking
	}
}
