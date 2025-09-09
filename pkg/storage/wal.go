package storage

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"sync"
	"time"
)

// FileWAL implements the WAL interface using a file
type FileWAL struct {
	filepath   string
	file       *os.File
	writer     *bufio.Writer
	mu         sync.Mutex
	sequence   uint64
	size       int64
	syncWrites bool
}

// NewWAL creates a new file-based WAL
func NewWAL() *FileWAL {
	return &FileWAL{
		syncWrites: true,
	}
}

// Open opens or creates a WAL file
func (w *FileWAL) Open(filepath string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.filepath = filepath

	// Open file in append mode
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open WAL file: %w", err)
	}

	w.file = file
	w.writer = bufio.NewWriter(file)

	// Get file size and last sequence number
	if err := w.initialize(); err != nil {
		return fmt.Errorf("failed to initialize WAL: %w", err)
	}

	return nil
}

// Close closes the WAL file
func (w *FileWAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer != nil {
		if err := w.writer.Flush(); err != nil {
			return err
		}
	}

	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Append adds a new entry to the WAL
func (w *FileWAL) Append(entry *WALEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Set sequence number
	w.sequence++
	entry.Sequence = w.sequence

	// Set timestamp if not set
	if entry.Timestamp == 0 {
		entry.Timestamp = time.Now().Unix()
	}

	// Calculate checksum
	entry.Checksum = w.calculateChecksum(entry)

	// Serialize entry
	data, err := w.serializeEntry(entry)
	if err != nil {
		return fmt.Errorf("failed to serialize WAL entry: %w", err)
	}

	// Write to buffer
	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write WAL entry: %w", err)
	}

	// Sync if required
	if w.syncWrites {
		if err := w.writer.Flush(); err != nil {
			return err
		}
		if err := w.file.Sync(); err != nil {
			return err
		}
	}

	w.size += int64(len(data))
	return nil
}

// Replay replays all WAL entries through the provided handler
func (w *FileWAL) Replay(handler func(*WALEntry) error) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Open file for reading
	file, err := os.Open(w.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No WAL file to replay
		}
		return fmt.Errorf("failed to open WAL for replay: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		entry, err := w.deserializeEntry(reader)
		if err != nil {
			if err.Error() == "EOF" {
				break // End of file
			}
			return fmt.Errorf("failed to deserialize WAL entry: %w", err)
		}

		// Verify checksum
		expectedChecksum := w.calculateChecksum(entry)
		if entry.Checksum != expectedChecksum {
			return fmt.Errorf("WAL entry checksum mismatch")
		}

		// Call handler
		if err := handler(entry); err != nil {
			return fmt.Errorf("WAL replay handler failed: %w", err)
		}
	}

	return nil
}

// Checkpoint marks a checkpoint in the WAL
func (w *FileWAL) Checkpoint(pageID uint32) error {
	checkpointEntry := &WALEntry{
		Type:      WALOpCommit,
		PageID:    pageID,
		Timestamp: time.Now().Unix(),
	}

	return w.Append(checkpointEntry)
}

// Truncate removes WAL entries before the specified sequence number
func (w *FileWAL) Truncate(beforeSeq uint64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Create temporary file
	tempPath := w.filepath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp WAL file: %w", err)
	}
	defer tempFile.Close()

	// Open current file for reading
	currentFile, err := os.Open(w.filepath)
	if err != nil {
		return fmt.Errorf("failed to open current WAL file: %w", err)
	}
	defer currentFile.Close()

	reader := bufio.NewReader(currentFile)
	writer := bufio.NewWriter(tempFile)

	// Copy entries with sequence >= beforeSeq
	for {
		entry, err := w.deserializeEntry(reader)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to read WAL entry during truncate: %w", err)
		}

		if entry.Sequence >= beforeSeq {
			data, err := w.serializeEntry(entry)
			if err != nil {
				return fmt.Errorf("failed to serialize WAL entry during truncate: %w", err)
			}

			if _, err := writer.Write(data); err != nil {
				return fmt.Errorf("failed to write WAL entry during truncate: %w", err)
			}
		}
	}

	// Flush and close temp file
	if err := writer.Flush(); err != nil {
		return err
	}

	// Close current file and writer
	if err := w.writer.Flush(); err != nil {
		return err
	}
	if err := w.file.Close(); err != nil {
		return err
	}

	// Replace current file with temp file
	if err := os.Rename(tempPath, w.filepath); err != nil {
		return fmt.Errorf("failed to replace WAL file: %w", err)
	}

	// Reopen file
	return w.Open(w.filepath)
}

// Private methods

func (w *FileWAL) initialize() error {
	// Get file info
	info, err := w.file.Stat()
	if err != nil {
		return err
	}

	w.size = info.Size()

	// If file is empty, we're done
	if w.size == 0 {
		w.sequence = 0
		return nil
	}

	// Find the last sequence number by reading the file
	file, err := os.Open(w.filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var lastSequence uint64

	for {
		entry, err := w.deserializeEntry(reader)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to read WAL entry during initialization: %w", err)
		}

		if entry.Sequence > lastSequence {
			lastSequence = entry.Sequence
		}
	}

	w.sequence = lastSequence
	return nil
}

func (w *FileWAL) serializeEntry(entry *WALEntry) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write fixed-size fields
	if err := binary.Write(buf, binary.LittleEndian, entry.Sequence); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, entry.Type); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, entry.PageID); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, entry.Timestamp); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, entry.Checksum); err != nil {
		return nil, err
	}

	// Write data length and data
	dataLen := uint32(len(entry.Data))
	if err := binary.Write(buf, binary.LittleEndian, dataLen); err != nil {
		return nil, err
	}
	if dataLen > 0 {
		if _, err := buf.Write(entry.Data); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (w *FileWAL) deserializeEntry(reader *bufio.Reader) (*WALEntry, error) {
	entry := &WALEntry{}

	// Read fixed-size fields
	if err := binary.Read(reader, binary.LittleEndian, &entry.Sequence); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.Type); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.PageID); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.Timestamp); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entry.Checksum); err != nil {
		return nil, err
	}

	// Read data length
	var dataLen uint32
	if err := binary.Read(reader, binary.LittleEndian, &dataLen); err != nil {
		return nil, err
	}

	// Read data if present
	if dataLen > 0 {
		entry.Data = make([]byte, dataLen)
		if _, err := reader.Read(entry.Data); err != nil {
			return nil, err
		}
	}

	return entry, nil
}

func (w *FileWAL) calculateChecksum(entry *WALEntry) uint32 {
	buf := new(bytes.Buffer)

	// Include all fields except checksum
	binary.Write(buf, binary.LittleEndian, entry.Sequence)
	binary.Write(buf, binary.LittleEndian, entry.Type)
	binary.Write(buf, binary.LittleEndian, entry.PageID)
	binary.Write(buf, binary.LittleEndian, entry.Timestamp)

	if len(entry.Data) > 0 {
		buf.Write(entry.Data)
	}

	return crc32.ChecksumIEEE(buf.Bytes())
}
