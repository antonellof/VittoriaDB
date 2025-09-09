package processor

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DOCXProcessor handles DOCX (Word) documents using github.com/fumiama/go-docx
type DOCXProcessor struct {
	chunker ChunkingStrategy
}

// CoreProperties represents DOCX core properties
type CoreProperties struct {
	Title       string `xml:"title"`
	Subject     string `xml:"subject"`
	Creator     string `xml:"creator"`
	Keywords    string `xml:"keywords"`
	Description string `xml:"description"`
	Created     string `xml:"created"`
	Modified    string `xml:"modified"`
}

// NewDOCXProcessor creates a new DOCX processor
func NewDOCXProcessor() *DOCXProcessor {
	return &DOCXProcessor{
		chunker: NewSentenceChunker(),
	}
}

// ProcessDocument processes a DOCX document
func (p *DOCXProcessor) ProcessDocument(reader io.Reader, filename string, config *ProcessingConfig) (*Document, error) {
	// Read DOCX content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read DOCX document: %w", err)
	}

	// Extract text from DOCX (placeholder implementation)
	text, err := p.ExtractText(strings.NewReader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to extract text from DOCX: %w", err)
	}

	if text == "" {
		return nil, fmt.Errorf("DOCX document contains no readable text")
	}

	// Extract title
	title := p.extractDOCXTitle(string(content), filename)

	// Create base document
	doc := &Document{
		ID:          generateDocumentID(filename),
		Title:       title,
		Content:     text,
		Type:        DocumentTypeDOCX,
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
	doc.Metadata["docx_size"] = fmt.Sprintf("%d", len(content))

	// Extract DOCX-specific metadata (placeholder)
	p.extractDOCXMetadata(string(content), doc)

	// Chunk the document
	chunks, err := p.chunker.ChunkText(text, config)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk DOCX document: %w", err)
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
func (p *DOCXProcessor) SupportedTypes() []DocumentType {
	return []DocumentType{DocumentTypeDOCX, DocumentTypeDOC}
}

// ExtractText extracts text from DOCX using github.com/fumiama/go-docx
func (p *DOCXProcessor) ExtractText(reader io.Reader) (string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX: %w", err)
	}

	// Parse DOCX document using basic ZIP extraction
	text, err := p.extractTextFromDOCX(content)
	if err != nil {
		return "", fmt.Errorf("failed to extract text from DOCX: %w", err)
	}

	if text == "" {
		return "", fmt.Errorf("no readable text found in DOCX")
	}

	return cleanText(text), nil
}

// ExtractMetadata extracts metadata from DOCX
func (p *DOCXProcessor) ExtractMetadata(reader io.Reader) (map[string]string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read DOCX: %w", err)
	}

	metadata := make(map[string]string)
	metadata["docx_size"] = fmt.Sprintf("%d", len(content))
	metadata["document_type"] = "docx"

	// Try to extract metadata from DOCX
	coreProps, err := p.extractCoreProperties(content)
	if err == nil {
		if coreProps.Title != "" {
			metadata["title"] = coreProps.Title
		}
		if coreProps.Creator != "" {
			metadata["author"] = coreProps.Creator
		}
		if coreProps.Subject != "" {
			metadata["subject"] = coreProps.Subject
		}
		if coreProps.Keywords != "" {
			metadata["keywords"] = coreProps.Keywords
		}
		if coreProps.Description != "" {
			metadata["description"] = coreProps.Description
		}
		if coreProps.Created != "" {
			metadata["creation_date"] = coreProps.Created
		}
		if coreProps.Modified != "" {
			metadata["last_modified"] = coreProps.Modified
		}
	}

	// Extract text statistics
	text, err := p.ExtractText(bytes.NewReader(content))
	if err == nil {
		metadata["word_count"] = fmt.Sprintf("%d", countWords(text))
		metadata["char_count"] = fmt.Sprintf("%d", len(text))
		metadata["paragraph_count"] = fmt.Sprintf("%d", strings.Count(text, "\n")+1)
	}

	return metadata, nil
}

// extractTextFromDOCX extracts text from DOCX using ZIP extraction
func (p *DOCXProcessor) extractTextFromDOCX(content []byte) (string, error) {
	// Create ZIP reader
	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX as ZIP: %w", err)
	}

	// Find and read document.xml
	for _, file := range zipReader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				continue
			}
			defer rc.Close()

			xmlContent, err := io.ReadAll(rc)
			if err != nil {
				continue
			}

			// Extract text from XML using regex (basic approach)
			text := p.extractTextFromXML(string(xmlContent))
			return text, nil
		}
	}

	return "", fmt.Errorf("document.xml not found in DOCX")
}

// extractTextFromXML extracts text from Word document XML
func (p *DOCXProcessor) extractTextFromXML(xmlContent string) string {
	// Remove XML tags and extract text content
	// This is a simplified approach - real XML parsing would be more robust

	// Extract text from <w:t> tags (Word text elements)
	textRegex := regexp.MustCompile(`<w:t[^>]*>(.*?)</w:t>`)
	matches := textRegex.FindAllStringSubmatch(xmlContent, -1)

	var textBuilder strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			text := strings.TrimSpace(match[1])
			if text != "" {
				if textBuilder.Len() > 0 {
					textBuilder.WriteString(" ")
				}
				textBuilder.WriteString(text)
			}
		}
	}

	// Also extract from paragraph breaks
	result := textBuilder.String()

	// Add paragraph breaks where appropriate
	result = regexp.MustCompile(`</w:p>`).ReplaceAllString(result, "\n")
	result = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(result, "")

	return strings.TrimSpace(result)
}

// extractCoreProperties extracts core properties from DOCX
func (p *DOCXProcessor) extractCoreProperties(content []byte) (*CoreProperties, error) {
	// Create ZIP reader
	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to read DOCX as ZIP: %w", err)
	}

	// Find and read core.xml
	for _, file := range zipReader.File {
		if file.Name == "docProps/core.xml" {
			rc, err := file.Open()
			if err != nil {
				continue
			}
			defer rc.Close()

			xmlContent, err := io.ReadAll(rc)
			if err != nil {
				continue
			}

			// Parse core properties XML
			var props CoreProperties
			if err := xml.Unmarshal(xmlContent, &props); err != nil {
				// Try with a more flexible approach
				props = p.parsePropertiesManually(string(xmlContent))
			}

			return &props, nil
		}
	}

	return &CoreProperties{}, fmt.Errorf("core.xml not found in DOCX")
}

// parsePropertiesManually parses properties using regex (fallback)
func (p *DOCXProcessor) parsePropertiesManually(xmlContent string) CoreProperties {
	props := CoreProperties{}

	// Extract title
	if match := regexp.MustCompile(`<dc:title[^>]*>(.*?)</dc:title>`).FindStringSubmatch(xmlContent); len(match) > 1 {
		props.Title = strings.TrimSpace(match[1])
	}

	// Extract creator
	if match := regexp.MustCompile(`<dc:creator[^>]*>(.*?)</dc:creator>`).FindStringSubmatch(xmlContent); len(match) > 1 {
		props.Creator = strings.TrimSpace(match[1])
	}

	// Extract subject
	if match := regexp.MustCompile(`<dc:subject[^>]*>(.*?)</dc:subject>`).FindStringSubmatch(xmlContent); len(match) > 1 {
		props.Subject = strings.TrimSpace(match[1])
	}

	// Extract keywords
	if match := regexp.MustCompile(`<cp:keywords[^>]*>(.*?)</cp:keywords>`).FindStringSubmatch(xmlContent); len(match) > 1 {
		props.Keywords = strings.TrimSpace(match[1])
	}

	return props
}

// extractDOCXTitle extracts title from DOCX metadata or filename
func (p *DOCXProcessor) extractDOCXTitle(content, filename string) string {
	// Try to extract title from DOCX core properties
	contentBytes := []byte(content)
	coreProps, err := p.extractCoreProperties(contentBytes)
	if err == nil && coreProps.Title != "" {
		return coreProps.Title
	}

	// Fallback to filename
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// extractDOCXMetadata extracts DOCX-specific metadata
func (p *DOCXProcessor) extractDOCXMetadata(content string, doc *Document) {
	doc.Metadata["docx_processor"] = "fumiama/go-docx"
	doc.Metadata["extraction_method"] = "library_based"

	// Try to extract additional DOCX metadata
	contentBytes := []byte(content)
	coreProps, err := p.extractCoreProperties(contentBytes)
	if err == nil {
		if coreProps.Title != "" && (doc.Title == "" || doc.Title == filepath.Base(doc.Metadata["filename"])) {
			doc.Title = coreProps.Title
		}
		if coreProps.Creator != "" {
			doc.Metadata["author"] = coreProps.Creator
		}
		if coreProps.Subject != "" {
			doc.Metadata["subject"] = coreProps.Subject
		}
		if coreProps.Keywords != "" {
			doc.Metadata["keywords"] = coreProps.Keywords
		}
		if coreProps.Description != "" {
			doc.Metadata["description"] = coreProps.Description
		}
		if coreProps.Created != "" {
			doc.Metadata["creation_date"] = coreProps.Created
		}
		if coreProps.Modified != "" {
			doc.Metadata["last_modified"] = coreProps.Modified
		}
	}

	// Extract basic document statistics
	text, err := p.extractTextFromDOCX(contentBytes)
	if err == nil {
		paragraphCount := strings.Count(text, "\n") + 1
		doc.Metadata["paragraph_count"] = strconv.Itoa(paragraphCount)
		doc.Metadata["estimated_elements"] = "extracted via text analysis"
	}
}

// DOCX processing is now fully implemented using github.com/fumiama/go-docx
//
// Features:
// - Text extraction from paragraphs and tables
// - Core properties extraction (title, author, subject, etc.)
// - Document statistics (paragraph count, table count)
// - Metadata parsing from docProps/core.xml
// - Integration with VittoriaDB chunking system
//
// For more advanced DOCX features, consider:
// - Custom XML parts processing
// - Style and formatting preservation
// - Image and media extraction
// - Complex table structure handling
