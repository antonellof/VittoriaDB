package processor

import (
	"io"
	"time"
)

// DocumentType represents the type of document being processed
type DocumentType string

const (
	DocumentTypePDF  DocumentType = "pdf"
	DocumentTypeDOCX DocumentType = "docx"
	DocumentTypeDOC  DocumentType = "doc"
	DocumentTypeTXT  DocumentType = "txt"
	DocumentTypeMD   DocumentType = "md"
	DocumentTypeHTML DocumentType = "html"
	DocumentTypeRTF  DocumentType = "rtf"
)

// Document represents a processed document
type Document struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Type        DocumentType      `json:"type"`
	Size        int64             `json:"size"`
	Pages       int               `json:"pages,omitempty"`
	Language    string            `json:"language,omitempty"`
	Metadata    map[string]string `json:"metadata"`
	ProcessedAt time.Time         `json:"processed_at"`
	Chunks      []DocumentChunk   `json:"chunks"`
}

// DocumentChunk represents a chunk of a document
type DocumentChunk struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Position int               `json:"position"`
	Page     int               `json:"page,omitempty"`
	Size     int               `json:"size"`
	Metadata map[string]string `json:"metadata"`
}

// ProcessingConfig contains configuration for document processing
type ProcessingConfig struct {
	ChunkSize    int               `json:"chunk_size"`     // Characters per chunk
	ChunkOverlap int               `json:"chunk_overlap"`  // Overlap between chunks
	MinChunkSize int               `json:"min_chunk_size"` // Minimum chunk size
	MaxChunkSize int               `json:"max_chunk_size"` // Maximum chunk size
	Language     string            `json:"language"`       // Document language
	Metadata     map[string]string `json:"metadata"`       // Additional metadata
}

// DefaultProcessingConfig returns default processing configuration
func DefaultProcessingConfig() *ProcessingConfig {
	return &ProcessingConfig{
		ChunkSize:    1024, // Increased for better semantic coherence (Memvid-style)
		ChunkOverlap: 128,  // Increased overlap for better context preservation
		MinChunkSize: 100,
		MaxChunkSize: 2048, // Slightly increased max size
		Language:     "en",
		Metadata:     make(map[string]string),
	}
}

// DocumentProcessor interface for processing different document types
type DocumentProcessor interface {
	// ProcessDocument processes a document from a reader
	ProcessDocument(reader io.Reader, filename string, config *ProcessingConfig) (*Document, error)

	// SupportedTypes returns the document types this processor supports
	SupportedTypes() []DocumentType

	// ExtractText extracts raw text from a document
	ExtractText(reader io.Reader) (string, error)

	// ExtractMetadata extracts metadata from a document
	ExtractMetadata(reader io.Reader) (map[string]string, error)
}

// ChunkingStrategy defines how documents are chunked
type ChunkingStrategy interface {
	// ChunkText splits text into chunks according to the strategy
	ChunkText(text string, config *ProcessingConfig) ([]DocumentChunk, error)
}

// ProcessingResult contains the result of document processing
type ProcessingResult struct {
	Document *Document `json:"document"`
	Success  bool      `json:"success"`
	Error    string    `json:"error,omitempty"`
	Duration int64     `json:"duration_ms"`
}

// ProcessingStats contains statistics about document processing
type ProcessingStats struct {
	TotalDocuments     int64            `json:"total_documents"`
	ProcessedDocuments int64            `json:"processed_documents"`
	FailedDocuments    int64            `json:"failed_documents"`
	TotalChunks        int64            `json:"total_chunks"`
	AverageChunkSize   float64          `json:"average_chunk_size"`
	ProcessingTime     time.Duration    `json:"processing_time"`
	TypeStats          map[string]int64 `json:"type_stats"`
}

// DocumentFilter allows filtering documents during processing
type DocumentFilter struct {
	MinSize      int64          `json:"min_size"`      // Minimum file size in bytes
	MaxSize      int64          `json:"max_size"`      // Maximum file size in bytes
	AllowedTypes []DocumentType `json:"allowed_types"` // Allowed document types
	RequiredMeta []string       `json:"required_meta"` // Required metadata fields
}

// ValidationError represents a document validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}
