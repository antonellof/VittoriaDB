package processor

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

// PDFProcessor handles PDF documents using github.com/ledongthuc/pdf
type PDFProcessor struct {
	chunker ChunkingStrategy
}

// NewPDFProcessor creates a new PDF processor
func NewPDFProcessor() *PDFProcessor {
	return &PDFProcessor{
		chunker: NewSentenceChunker(),
	}
}

// ProcessDocument processes a PDF document
func (p *PDFProcessor) ProcessDocument(reader io.Reader, filename string, config *ProcessingConfig) (*Document, error) {
	// Read PDF content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF document: %w", err)
	}

	// Extract text from PDF (placeholder implementation)
	text, err := p.ExtractText(strings.NewReader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to extract text from PDF: %w", err)
	}

	if text == "" {
		return nil, fmt.Errorf("PDF document contains no readable text")
	}

	// Extract title
	title := p.extractPDFTitle(string(content), filename)

	// Create base document
	doc := &Document{
		ID:          generateDocumentID(filename),
		Title:       title,
		Content:     text,
		Type:        DocumentTypePDF,
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
	doc.Metadata["pdf_size"] = fmt.Sprintf("%d", len(content))

	// Extract PDF-specific metadata (placeholder)
	p.extractPDFMetadata(string(content), doc)

	// Chunk the document
	chunks, err := p.chunker.ChunkText(text, config)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk PDF document: %w", err)
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
func (p *PDFProcessor) SupportedTypes() []DocumentType {
	return []DocumentType{DocumentTypePDF}
}

// ExtractText extracts text from PDF using github.com/ledongthuc/pdf
func (p *PDFProcessor) ExtractText(reader io.Reader) (string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF: %w", err)
	}

	// Create PDF reader
	pdfReader, err := pdf.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to parse PDF: %w", err)
	}

	var textBuilder strings.Builder
	numPages := pdfReader.NumPage()

	// Extract text from each page
	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}

		// Get plain text from page (pass empty font map for basic extraction)
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			// Log error but continue with other pages
			continue
		}

		// Clean and add page text
		pageText = strings.TrimSpace(pageText)
		if pageText != "" {
			if textBuilder.Len() > 0 {
				textBuilder.WriteString("\n\n")
			}
			textBuilder.WriteString(pageText)
		}
	}

	text := textBuilder.String()
	if text == "" {
		return "", fmt.Errorf("no readable text found in PDF")
	}

	return cleanText(text), nil
}

// ExtractMetadata extracts metadata from PDF
func (p *PDFProcessor) ExtractMetadata(reader io.Reader) (map[string]string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	metadata := make(map[string]string)
	metadata["pdf_size"] = fmt.Sprintf("%d", len(content))

	// Create PDF reader to extract metadata
	pdfReader, err := pdf.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		// If we can't parse PDF, return basic info
		metadata["pdf_version"] = "unknown"
		metadata["page_count"] = "unknown"
		metadata["error"] = "failed to parse PDF for metadata"
		return metadata, nil
	}

	// Extract basic PDF info
	numPages := pdfReader.NumPage()
	metadata["page_count"] = strconv.Itoa(numPages)

	// Try to extract PDF version from header
	pdfContent := string(content)
	if strings.Contains(pdfContent, "%PDF-1.") {
		if idx := strings.Index(pdfContent, "%PDF-1."); idx != -1 && idx+8 < len(pdfContent) {
			version := pdfContent[idx : idx+8]
			metadata["pdf_version"] = version
		}
	}

	// Extract text statistics
	text, err := p.ExtractText(bytes.NewReader(content))
	if err == nil {
		metadata["word_count"] = fmt.Sprintf("%d", countWords(text))
		metadata["char_count"] = fmt.Sprintf("%d", len(text))
	}

	return metadata, nil
}

// extractPDFTitle extracts title from PDF metadata or filename
func (p *PDFProcessor) extractPDFTitle(content, filename string) string {
	// Try to extract title from PDF metadata
	contentBytes := []byte(content)
	_, err := pdf.NewReader(bytes.NewReader(contentBytes), int64(len(contentBytes)))
	if err == nil {
		// Try to get document info (this is basic - PDF metadata can be complex)
		// For now, fall back to filename since the library doesn't expose metadata easily
	}

	// Use filename as fallback
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// extractPDFMetadata extracts PDF-specific metadata
func (p *PDFProcessor) extractPDFMetadata(content string, doc *Document) {
	doc.Metadata["pdf_processor"] = "ledongthuc/pdf"
	doc.Metadata["extraction_method"] = "library_based"

	// Try to extract additional PDF info
	contentBytes := []byte(content)
	pdfReader, err := pdf.NewReader(bytes.NewReader(contentBytes), int64(len(contentBytes)))
	if err != nil {
		doc.Metadata["pdf_error"] = "failed to parse PDF for detailed metadata"
		return
	}

	// Extract page count
	numPages := pdfReader.NumPage()
	doc.Pages = numPages
	doc.Metadata["page_count"] = strconv.Itoa(numPages)

	// Extract PDF version from content
	if strings.Contains(content, "%PDF-1.") {
		if idx := strings.Index(content, "%PDF-1."); idx != -1 && idx+8 < len(content) {
			version := content[idx : idx+8]
			doc.Metadata["pdf_version"] = version
		}
	}
}

// PDF processing is now fully implemented using github.com/ledongthuc/pdf
//
// Features:
// - Text extraction from all pages
// - Page count and PDF version detection
// - Error handling for corrupted PDFs
// - Integration with VittoriaDB chunking system
//
// For more advanced PDF features, consider:
// - github.com/pdfcpu/pdfcpu for PDF manipulation
// - External tools like pdftotext for complex layouts
// - Commercial SDKs for enterprise features
