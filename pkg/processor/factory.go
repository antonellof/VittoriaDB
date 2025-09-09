package processor

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ProcessorFactory creates document processors based on file type
type ProcessorFactory struct {
	processors map[DocumentType]DocumentProcessor
}

// NewProcessorFactory creates a new processor factory
func NewProcessorFactory() *ProcessorFactory {
	factory := &ProcessorFactory{
		processors: make(map[DocumentType]DocumentProcessor),
	}

	// Register default processors
	factory.RegisterProcessor(NewTextProcessor())
	factory.RegisterProcessor(NewHTMLProcessor())
	factory.RegisterProcessor(NewPDFProcessor())
	factory.RegisterProcessor(NewDOCXProcessor())

	return factory
}

// RegisterProcessor registers a document processor
func (f *ProcessorFactory) RegisterProcessor(processor DocumentProcessor) {
	for _, docType := range processor.SupportedTypes() {
		f.processors[docType] = processor
	}
}

// GetProcessor returns a processor for the given document type
func (f *ProcessorFactory) GetProcessor(docType DocumentType) (DocumentProcessor, error) {
	processor, exists := f.processors[docType]
	if !exists {
		return nil, fmt.Errorf("no processor available for document type: %s", docType)
	}
	return processor, nil
}

// GetProcessorByFilename returns a processor based on file extension
func (f *ProcessorFactory) GetProcessorByFilename(filename string) (DocumentProcessor, error) {
	docType := f.DetectDocumentType(filename)
	return f.GetProcessor(docType)
}

// DetectDocumentType detects document type from filename
func (f *ProcessorFactory) DetectDocumentType(filename string) DocumentType {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return DocumentTypePDF
	case ".docx":
		return DocumentTypeDOCX
	case ".doc":
		return DocumentTypeDOC
	case ".txt", ".text":
		return DocumentTypeTXT
	case ".md", ".markdown":
		return DocumentTypeMD
	case ".html", ".htm":
		return DocumentTypeHTML
	case ".rtf":
		return DocumentTypeRTF
	default:
		// Default to text for unknown extensions
		return DocumentTypeTXT
	}
}

// SupportedTypes returns all supported document types
func (f *ProcessorFactory) SupportedTypes() []DocumentType {
	var types []DocumentType
	for docType := range f.processors {
		types = append(types, docType)
	}
	return types
}

// IsSupported checks if a document type is supported
func (f *ProcessorFactory) IsSupported(docType DocumentType) bool {
	_, exists := f.processors[docType]
	return exists
}

// IsSupportedFile checks if a file is supported based on extension
func (f *ProcessorFactory) IsSupportedFile(filename string) bool {
	docType := f.DetectDocumentType(filename)
	return f.IsSupported(docType)
}

// GetSupportedExtensions returns all supported file extensions
func (f *ProcessorFactory) GetSupportedExtensions() []string {
	return []string{
		".pdf",      // PDF documents
		".docx",     // Word documents (modern)
		".doc",      // Word documents (legacy)
		".txt",      // Plain text
		".text",     // Plain text (alternative)
		".md",       // Markdown
		".markdown", // Markdown (alternative)
		".html",     // HTML documents
		".htm",      // HTML documents (alternative)
		".rtf",      // Rich Text Format (placeholder)
	}
}

// ProcessorInfo contains information about a processor
type ProcessorInfo struct {
	Type        DocumentType `json:"type"`
	Extensions  []string     `json:"extensions"`
	Description string       `json:"description"`
	Status      string       `json:"status"`
}

// GetProcessorInfo returns information about all registered processors
func (f *ProcessorFactory) GetProcessorInfo() []ProcessorInfo {
	info := []ProcessorInfo{
		{
			Type:        DocumentTypeTXT,
			Extensions:  []string{".txt", ".text"},
			Description: "Plain text documents",
			Status:      "fully_implemented",
		},
		{
			Type:        DocumentTypeMD,
			Extensions:  []string{".md", ".markdown"},
			Description: "Markdown documents with frontmatter support",
			Status:      "fully_implemented",
		},
		{
			Type:        DocumentTypeHTML,
			Extensions:  []string{".html", ".htm"},
			Description: "HTML documents with tag stripping and metadata extraction",
			Status:      "fully_implemented",
		},
		{
			Type:        DocumentTypePDF,
			Extensions:  []string{".pdf"},
			Description: "PDF documents with full text extraction using github.com/ledongthuc/pdf",
			Status:      "fully_implemented",
		},
		{
			Type:        DocumentTypeDOCX,
			Extensions:  []string{".docx"},
			Description: "Microsoft Word documents with metadata extraction using github.com/fumiama/go-docx",
			Status:      "fully_implemented",
		},
		{
			Type:        DocumentTypeDOC,
			Extensions:  []string{".doc"},
			Description: "Legacy Microsoft Word documents (requires additional library integration)",
			Status:      "placeholder",
		},
		{
			Type:        DocumentTypeRTF,
			Extensions:  []string{".rtf"},
			Description: "Rich Text Format documents (not yet implemented)",
			Status:      "not_implemented",
		},
	}

	return info
}

// ValidateFile validates if a file can be processed
func (f *ProcessorFactory) ValidateFile(filename string, size int64, filter *DocumentFilter) error {
	// Check file extension
	if !f.IsSupportedFile(filename) {
		return ValidationError{
			Field:   "filename",
			Message: fmt.Sprintf("unsupported file type: %s", filepath.Ext(filename)),
		}
	}

	if filter == nil {
		return nil
	}

	// Check file size
	if filter.MinSize > 0 && size < filter.MinSize {
		return ValidationError{
			Field:   "size",
			Message: fmt.Sprintf("file size %d bytes is below minimum %d bytes", size, filter.MinSize),
		}
	}

	if filter.MaxSize > 0 && size > filter.MaxSize {
		return ValidationError{
			Field:   "size",
			Message: fmt.Sprintf("file size %d bytes exceeds maximum %d bytes", size, filter.MaxSize),
		}
	}

	// Check allowed types
	if len(filter.AllowedTypes) > 0 {
		docType := f.DetectDocumentType(filename)
		allowed := false
		for _, allowedType := range filter.AllowedTypes {
			if docType == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return ValidationError{
				Field:   "type",
				Message: fmt.Sprintf("document type %s is not in allowed types", docType),
			}
		}
	}

	return nil
}

// DefaultProcessorFactory returns a factory with all default processors
var DefaultProcessorFactory = NewProcessorFactory()
