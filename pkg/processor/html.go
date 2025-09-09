package processor

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// HTMLProcessor handles HTML documents
type HTMLProcessor struct {
	chunker ChunkingStrategy
}

// NewHTMLProcessor creates a new HTML processor
func NewHTMLProcessor() *HTMLProcessor {
	return &HTMLProcessor{
		chunker: NewSentenceChunker(),
	}
}

// ProcessDocument processes an HTML document
func (p *HTMLProcessor) ProcessDocument(reader io.Reader, filename string, config *ProcessingConfig) (*Document, error) {
	// Read all content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTML document: %w", err)
	}

	html := string(content)

	// Extract text content from HTML
	text, err := p.ExtractText(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to extract text from HTML: %w", err)
	}

	if text == "" {
		return nil, fmt.Errorf("HTML document contains no readable text")
	}

	// Extract title from HTML
	title := p.extractHTMLTitle(html, filename)

	// Create base document
	doc := &Document{
		ID:          generateDocumentID(filename),
		Title:       title,
		Content:     text,
		Type:        DocumentTypeHTML,
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
	doc.Metadata["original_size"] = fmt.Sprintf("%d", len(html))

	// Extract HTML-specific metadata
	p.extractHTMLMetadata(html, doc)

	// Chunk the document
	chunks, err := p.chunker.ChunkText(text, config)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk HTML document: %w", err)
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
func (p *HTMLProcessor) SupportedTypes() []DocumentType {
	return []DocumentType{DocumentTypeHTML}
}

// ExtractText extracts plain text from HTML
func (p *HTMLProcessor) ExtractText(reader io.Reader) (string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read HTML: %w", err)
	}

	html := string(content)

	// Remove script and style tags with their content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	html = scriptRegex.ReplaceAllString(html, "")

	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	html = styleRegex.ReplaceAllString(html, "")

	// Remove HTML comments
	commentRegex := regexp.MustCompile(`<!--.*?-->`)
	html = commentRegex.ReplaceAllString(html, "")

	// Convert common HTML entities
	html = p.decodeHTMLEntities(html)

	// Remove all HTML tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text := tagRegex.ReplaceAllString(html, " ")

	// Clean up whitespace
	text = cleanText(text)

	return text, nil
}

// ExtractMetadata extracts metadata from HTML
func (p *HTMLProcessor) ExtractMetadata(reader io.Reader) (map[string]string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTML: %w", err)
	}

	html := string(content)
	metadata := make(map[string]string)

	// Extract meta tags
	metaRegex := regexp.MustCompile(`(?i)<meta\s+([^>]+)>`)
	matches := metaRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			attrs := p.parseHTMLAttributes(match[1])

			if name, ok := attrs["name"]; ok {
				if content, ok := attrs["content"]; ok {
					metadata["meta_"+name] = content
				}
			}

			if property, ok := attrs["property"]; ok {
				if content, ok := attrs["content"]; ok {
					metadata["meta_"+property] = content
				}
			}
		}
	}

	// Extract text content for word count
	text, _ := p.ExtractText(strings.NewReader(html))
	metadata["word_count"] = fmt.Sprintf("%d", countWords(text))
	metadata["char_count"] = fmt.Sprintf("%d", len(text))
	metadata["original_size"] = fmt.Sprintf("%d", len(html))

	return metadata, nil
}

// extractHTMLTitle extracts title from HTML
func (p *HTMLProcessor) extractHTMLTitle(html, filename string) string {
	// Try to extract from <title> tag
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		if title != "" {
			return p.decodeHTMLEntities(title)
		}
	}

	// Try to extract from <h1> tag
	h1Regex := regexp.MustCompile(`(?i)<h1[^>]*>(.*?)</h1>`)
	matches = h1Regex.FindStringSubmatch(html)
	if len(matches) > 1 {
		// Remove HTML tags from h1 content
		tagRegex := regexp.MustCompile(`<[^>]*>`)
		title := tagRegex.ReplaceAllString(matches[1], "")
		title = strings.TrimSpace(title)
		if title != "" {
			return p.decodeHTMLEntities(title)
		}
	}

	// Fallback to filename
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// extractHTMLMetadata extracts HTML-specific metadata
func (p *HTMLProcessor) extractHTMLMetadata(html string, doc *Document) {
	// Count various HTML elements
	linkCount := len(regexp.MustCompile(`(?i)<a\s+[^>]*href`).FindAllString(html, -1))
	doc.Metadata["link_count"] = fmt.Sprintf("%d", linkCount)

	imgCount := len(regexp.MustCompile(`(?i)<img\s+[^>]*src`).FindAllString(html, -1))
	doc.Metadata["image_count"] = fmt.Sprintf("%d", imgCount)

	headingCount := len(regexp.MustCompile(`(?i)<h[1-6][^>]*>`).FindAllString(html, -1))
	doc.Metadata["heading_count"] = fmt.Sprintf("%d", headingCount)

	// Extract language from html tag
	langRegex := regexp.MustCompile(`(?i)<html[^>]+lang=["']([^"']+)["']`)
	matches := langRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		doc.Metadata["html_lang"] = matches[1]
		if doc.Language == "en" { // Only override if default
			doc.Language = matches[1]
		}
	}

	// Extract charset
	charsetRegex := regexp.MustCompile(`(?i)charset=["']?([^"'\s>]+)`)
	matches = charsetRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		doc.Metadata["charset"] = matches[1]
	}
}

// parseHTMLAttributes parses HTML attributes from a string
func (p *HTMLProcessor) parseHTMLAttributes(attrString string) map[string]string {
	attrs := make(map[string]string)

	// Simple attribute parsing (name="value" or name='value')
	attrRegex := regexp.MustCompile(`(\w+)=["']([^"']*)["']`)
	matches := attrRegex.FindAllStringSubmatch(attrString, -1)

	for _, match := range matches {
		if len(match) > 2 {
			attrs[strings.ToLower(match[1])] = match[2]
		}
	}

	return attrs
}

// decodeHTMLEntities decodes common HTML entities
func (p *HTMLProcessor) decodeHTMLEntities(text string) string {
	// Common HTML entities
	entities := map[string]string{
		"&amp;":    "&",
		"&lt;":     "<",
		"&gt;":     ">",
		"&quot;":   "\"",
		"&apos;":   "'",
		"&nbsp;":   " ",
		"&copy;":   "©",
		"&reg;":    "®",
		"&trade;":  "™",
		"&mdash;":  "—",
		"&ndash;":  "–",
		"&hellip;": "…",
		"&laquo;":  "«",
		"&raquo;":  "»",
	}

	for entity, replacement := range entities {
		text = strings.ReplaceAll(text, entity, replacement)
	}

	// Decode numeric entities (basic support)
	numericRegex := regexp.MustCompile(`&#(\d+);`)
	text = numericRegex.ReplaceAllStringFunc(text, func(match string) string {
		// For simplicity, just remove numeric entities
		// In a full implementation, you'd convert the number to the actual character
		return " "
	})

	return text
}
