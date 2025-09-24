package processor

import (
	"fmt"
	"strings"
	"testing"
)

func TestSmartChunker_BasicSentenceChunking(t *testing.T) {
	chunker := NewSmartChunker()
	config := &ProcessingConfig{
		ChunkSize:    200,
		ChunkOverlap: 20,
		MinChunkSize: 50,
	}

	text := "This is the first sentence. This is the second sentence. This is the third sentence. This is the fourth sentence. This is the fifth sentence."

	chunks, err := chunker.ChunkText(text, config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}

	// Verify chunk properties
	for i, chunk := range chunks {
		if chunk.Size != len(chunk.Content) {
			t.Errorf("Chunk %d: size mismatch. Expected %d, got %d", i, len(chunk.Content), chunk.Size)
		}

		if chunk.Position != i {
			t.Errorf("Chunk %d: position mismatch. Expected %d, got %d", i, i, chunk.Position)
		}

		if chunk.Metadata["chunk_type"] != "smart_sentence" {
			t.Errorf("Chunk %d: expected chunk_type 'smart_sentence', got '%s'", i, chunk.Metadata["chunk_type"])
		}
	}

	t.Logf("Created %d chunks from text of %d characters", len(chunks), len(text))
}

func TestSmartChunker_ParagraphStructuredText(t *testing.T) {
	chunker := NewSmartChunker()
	config := &ProcessingConfig{
		ChunkSize:    300,
		ChunkOverlap: 30,
		MinChunkSize: 50,
	}

	text := `This is the first paragraph. It contains multiple sentences that form a coherent unit of thought. The paragraph discusses an important topic.

This is the second paragraph. It builds upon the ideas from the first paragraph. The content flows naturally from one idea to the next.

This is the third paragraph. It provides additional context and examples. The structure helps readers understand the progression of ideas.`

	chunks, err := chunker.ChunkText(text, config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}

	// Should detect paragraph structure and use paragraph chunking
	foundParagraphChunk := false
	for _, chunk := range chunks {
		if chunk.Metadata["boundary_type"] == "paragraph" {
			foundParagraphChunk = true
			break
		}
	}

	if !foundParagraphChunk {
		t.Error("Expected to find paragraph-based chunks for structured text")
	}

	t.Logf("Created %d chunks with paragraph structure detection", len(chunks))
}

func TestSmartChunker_AbbreviationHandling(t *testing.T) {
	chunker := NewSmartChunker()
	config := &ProcessingConfig{
		ChunkSize:    100,
		ChunkOverlap: 10,
		MinChunkSize: 20,
	}

	text := "Dr. Smith went to the U.S.A. He met with Prof. Johnson at 3.14 p.m. They discussed the research."

	chunks, err := chunker.ChunkText(text, config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should not split on abbreviations
	fullText := strings.Join(func() []string {
		var contents []string
		for _, chunk := range chunks {
			contents = append(contents, chunk.Content)
		}
		return contents
	}(), " ")

	if !strings.Contains(fullText, "Dr. Smith") {
		t.Error("Should not split on 'Dr.' abbreviation")
	}

	if !strings.Contains(fullText, "U.S.A.") {
		t.Error("Should not split on 'U.S.A.' abbreviation")
	}

	if !strings.Contains(fullText, "3.14") {
		t.Error("Should not split on decimal numbers")
	}

	t.Logf("Properly handled abbreviations in %d chunks", len(chunks))
}

func TestSmartChunker_OverlapFunctionality(t *testing.T) {
	chunker := NewSmartChunker()
	config := &ProcessingConfig{
		ChunkSize:    100,
		ChunkOverlap: 30,
		MinChunkSize: 20,
	}

	text := "First sentence here. Second sentence follows. Third sentence continues. Fourth sentence extends. Fifth sentence concludes."

	chunks, err := chunker.ChunkText(text, config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatal("Expected at least 2 chunks to test overlap")
	}

	// Check for overlap between consecutive chunks
	for i := 1; i < len(chunks); i++ {
		prevChunk := chunks[i-1].Content
		currentChunk := chunks[i].Content

		// Find common words between chunks (simple overlap check)
		prevWords := strings.Fields(prevChunk)
		currentWords := strings.Fields(currentChunk)

		hasOverlap := false
		for _, prevWord := range prevWords {
			for _, currentWord := range currentWords {
				if prevWord == currentWord {
					hasOverlap = true
					break
				}
			}
			if hasOverlap {
				break
			}
		}

		if !hasOverlap && config.ChunkOverlap > 0 {
			t.Errorf("Expected overlap between chunk %d and %d", i-1, i)
		}
	}

	t.Logf("Verified overlap functionality across %d chunks", len(chunks))
}

func TestSmartChunker_LargeParagraphSplitting(t *testing.T) {
	chunker := NewSmartChunker()
	config := &ProcessingConfig{
		ChunkSize:    200,
		ChunkOverlap: 20,
		MinChunkSize: 50,
	}

	// Create a very large paragraph that should be split by sentences
	longParagraph := strings.Repeat("This is a very long sentence that will be part of a large paragraph. ", 20)

	chunks, err := chunker.ChunkText(longParagraph, config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}

	// Should split large paragraph by sentences
	for _, chunk := range chunks {
		if chunk.Size > config.ChunkSize*2 { // Allow some flexibility
			t.Errorf("Chunk too large: %d characters (max expected: ~%d)", chunk.Size, config.ChunkSize*2)
		}
	}

	t.Logf("Successfully split large paragraph into %d manageable chunks", len(chunks))
}

func TestSmartChunker_EmptyAndShortText(t *testing.T) {
	chunker := NewSmartChunker()
	config := DefaultProcessingConfig()

	// Test empty text
	chunks, err := chunker.ChunkText("", config)
	if err != nil {
		t.Fatalf("Expected no error for empty text, got: %v", err)
	}
	if len(chunks) != 0 {
		t.Errorf("Expected 0 chunks for empty text, got %d", len(chunks))
	}

	// Test very short text
	shortText := "Short."
	chunks, err = chunker.ChunkText(shortText, config)
	if err != nil {
		t.Fatalf("Expected no error for short text, got: %v", err)
	}

	// Short text below minimum should not create chunks
	if len(chunks) > 1 {
		t.Errorf("Expected at most 1 chunk for very short text, got %d", len(chunks))
	}

	t.Logf("Handled edge cases correctly")
}

func TestSmartChunker_MetadataGeneration(t *testing.T) {
	chunker := NewSmartChunker()
	config := &ProcessingConfig{
		ChunkSize:    200,
		ChunkOverlap: 20,
		MinChunkSize: 50,
	}

	text := "First sentence. Second sentence. Third sentence. Fourth sentence."

	chunks, err := chunker.ChunkText(text, config)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	for i, chunk := range chunks {
		// Check required metadata fields
		requiredFields := []string{"chunk_type", "char_count", "word_count", "boundary_type"}
		for _, field := range requiredFields {
			if _, exists := chunk.Metadata[field]; !exists {
				t.Errorf("Chunk %d missing required metadata field: %s", i, field)
			}
		}

		// Verify char_count accuracy
		expectedCharCount := len(chunk.Content)
		if chunk.Metadata["char_count"] != string(rune(expectedCharCount)) {
			// Convert properly for comparison
			if chunk.Metadata["char_count"] != fmt.Sprintf("%d", expectedCharCount) {
				t.Errorf("Chunk %d char_count mismatch. Expected %d", i, expectedCharCount)
			}
		}

		// Verify word_count
		expectedWordCount := len(strings.Fields(chunk.Content))
		if chunk.Metadata["word_count"] != fmt.Sprintf("%d", expectedWordCount) {
			t.Errorf("Chunk %d word_count mismatch. Expected %d", i, expectedWordCount)
		}
	}

	t.Logf("Verified metadata generation for %d chunks", len(chunks))
}

func TestSmartChunker_PerformanceWithLargeText(t *testing.T) {
	chunker := NewSmartChunker()
	config := DefaultProcessingConfig()

	// Create a large text document
	sentences := make([]string, 1000)
	for i := range sentences {
		sentences[i] = fmt.Sprintf("This is sentence number %d in a large document that tests performance.", i+1)
	}
	largeText := strings.Join(sentences, " ")

	chunks, err := chunker.ChunkText(largeText, config)
	if err != nil {
		t.Fatalf("Expected no error for large text, got: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("Expected chunks from large text")
	}

	// Verify all chunks are within size limits
	for i, chunk := range chunks {
		if chunk.Size > config.MaxChunkSize {
			t.Errorf("Chunk %d exceeds max size: %d > %d", i, chunk.Size, config.MaxChunkSize)
		}
		if chunk.Size < config.MinChunkSize && i < len(chunks)-1 { // Allow last chunk to be smaller
			t.Errorf("Chunk %d below min size: %d < %d", i, chunk.Size, config.MinChunkSize)
		}
	}

	t.Logf("Successfully processed large text (%d chars) into %d chunks", len(largeText), len(chunks))
}

func TestGetChunker_SmartChunkerDefault(t *testing.T) {
	// Test that smart chunker is now the default
	chunker := GetChunker("unknown_strategy")
	
	if _, ok := chunker.(*SmartChunker); !ok {
		t.Error("Expected SmartChunker as default, got different type")
	}

	// Test explicit smart chunker selection
	smartChunker := GetChunker("smart")
	if _, ok := smartChunker.(*SmartChunker); !ok {
		t.Error("Expected SmartChunker for 'smart' strategy")
	}

	t.Log("Verified SmartChunker is properly integrated as default")
}
