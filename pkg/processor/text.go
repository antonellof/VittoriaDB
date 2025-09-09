package processor

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// TextProcessor handles plain text and markdown files
type TextProcessor struct {
	chunker ChunkingStrategy
}

// NewTextProcessor creates a new text processor
func NewTextProcessor() *TextProcessor {
	return &TextProcessor{
		chunker: NewSentenceChunker(),
	}
}

// ProcessDocument processes a text document
func (p *TextProcessor) ProcessDocument(reader io.Reader, filename string, config *ProcessingConfig) (*Document, error) {
	// Read all content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read document: %w", err)
	}

	text := string(content)

	// Clean and normalize text
	text = cleanText(text)

	if text == "" {
		return nil, fmt.Errorf("document contains no readable text")
	}

	// Determine document type
	docType := p.getDocumentType(filename)

	// Extract title from content or filename
	title := p.extractTitle(text, filename)

	// Create base document
	doc := &Document{
		ID:          generateDocumentID(filename),
		Title:       title,
		Content:     text,
		Type:        docType,
		Size:        int64(len(content)),
		Language:    config.Language,
		Metadata:    make(map[string]string),
		ProcessedAt: time.Now(),
	}

	// Add metadata
	for k, v := range config.Metadata {
		doc.Metadata[k] = v
	}
	doc.Metadata["filename"] = filename
	doc.Metadata["file_extension"] = filepath.Ext(filename)
	doc.Metadata["word_count"] = fmt.Sprintf("%d", countWords(text))
	doc.Metadata["char_count"] = fmt.Sprintf("%d", len(text))
	doc.Metadata["line_count"] = fmt.Sprintf("%d", strings.Count(text, "\n")+1)

	// Process markdown-specific metadata
	if docType == DocumentTypeMD {
		p.extractMarkdownMetadata(text, doc)
	}

	// Chunk the document
	chunks, err := p.chunker.ChunkText(text, config)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk document: %w", err)
	}

	// Add document ID to each chunk
	for i := range chunks {
		chunks[i].ID = fmt.Sprintf("%s_chunk_%d", doc.ID, i)
		chunks[i].Metadata["document_id"] = doc.ID
		chunks[i].Metadata["document_title"] = doc.Title
		chunks[i].Metadata["document_type"] = string(doc.Type)
	}

	doc.Chunks = chunks

	return doc, nil
}

// SupportedTypes returns supported document types
func (p *TextProcessor) SupportedTypes() []DocumentType {
	return []DocumentType{DocumentTypeTXT, DocumentTypeMD}
}

// ExtractText extracts raw text from the document
func (p *TextProcessor) ExtractText(reader io.Reader) (string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read text: %w", err)
	}

	return cleanText(string(content)), nil
}

// ExtractMetadata extracts metadata from the document
func (p *TextProcessor) ExtractMetadata(reader io.Reader) (map[string]string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read document: %w", err)
	}

	text := string(content)
	metadata := make(map[string]string)

	metadata["word_count"] = fmt.Sprintf("%d", countWords(text))
	metadata["char_count"] = fmt.Sprintf("%d", len(text))
	metadata["line_count"] = fmt.Sprintf("%d", strings.Count(text, "\n")+1)

	return metadata, nil
}

// getDocumentType determines document type from filename
func (p *TextProcessor) getDocumentType(filename string) DocumentType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".md", ".markdown":
		return DocumentTypeMD
	case ".txt", ".text":
		return DocumentTypeTXT
	default:
		return DocumentTypeTXT
	}
}

// extractTitle extracts title from content or uses filename
func (p *TextProcessor) extractTitle(text, filename string) string {
	lines := strings.Split(text, "\n")

	// For markdown, look for # title
	if strings.HasSuffix(strings.ToLower(filename), ".md") {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") {
				return strings.TrimSpace(strings.TrimPrefix(line, "#"))
			}
		}
	}

	// For any text, use first non-empty line if it looks like a title
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && len(line) < 100 && !strings.Contains(line, "\t") {
			return line
		}
	}

	// Fallback to filename without extension
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// extractMarkdownMetadata extracts metadata from markdown frontmatter
func (p *TextProcessor) extractMarkdownMetadata(text string, doc *Document) {
	lines := strings.Split(text, "\n")

	// Check for YAML frontmatter
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "---" {
				break
			}

			// Parse simple key: value pairs
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					value = strings.Trim(value, `"'`) // Remove quotes
					doc.Metadata["frontmatter_"+key] = value

					// Special handling for common fields
					switch strings.ToLower(key) {
					case "title":
						if doc.Title == "" || doc.Title == filepath.Base(doc.Metadata["filename"]) {
							doc.Title = value
						}
					case "author":
						doc.Metadata["author"] = value
					case "date":
						doc.Metadata["date"] = value
					case "tags":
						doc.Metadata["tags"] = value
					}
				}
			}
		}
	}

	// Extract markdown structure info
	headingCount := strings.Count(text, "#")
	doc.Metadata["heading_count"] = fmt.Sprintf("%d", headingCount)

	linkCount := strings.Count(text, "](")
	doc.Metadata["link_count"] = fmt.Sprintf("%d", linkCount)

	codeBlockCount := strings.Count(text, "```")
	doc.Metadata["code_block_count"] = fmt.Sprintf("%d", codeBlockCount/2)
}

// Helper functions

// generateDocumentID generates a unique document ID
func generateDocumentID(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Clean the name for use as ID
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	// Add timestamp to ensure uniqueness
	timestamp := time.Now().Unix()

	return fmt.Sprintf("doc_%s_%d", name, timestamp)
}

// countWords counts words in text
func countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}
