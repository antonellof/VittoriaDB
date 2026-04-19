package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStorageEngine_WALReplayAndTruncate(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	e := NewFileStorageEngine(32)
	if err := e.Open(dbPath); err != nil {
		t.Fatalf("open: %v", err)
	}

	p := &Page{
		ID:   1,
		Type: PageTypeVectorLeaf,
		Data: []byte("hello-wal-replay"),
	}
	if err := e.WritePage(p); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Re-open: replays WAL (or reads from main file) and truncates WAL
	e2 := NewFileStorageEngine(32)
	if err := e2.Open(dbPath); err != nil {
		t.Fatalf("reopen: %v", err)
	}
	rp, err := e2.ReadPage(1)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(rp.Data) < len(p.Data) {
		t.Fatalf("short data")
	}
	if string(rp.Data[:len(p.Data)]) != string(p.Data) {
		t.Fatalf("data mismatch got %q", string(rp.Data[:len(p.Data)]))
	}

	walPath := dbPath + ".wal"
	st, err := os.Stat(walPath)
	if err != nil {
		t.Fatalf("stat wal: %v", err)
	}
	if st.Size() != 0 {
		t.Fatalf("expected WAL truncated after replay, size=%d", st.Size())
	}

	if err := e2.Close(); err != nil {
		t.Fatalf("close2: %v", err)
	}

	// Third open — data still on disk
	e3 := NewFileStorageEngine(32)
	if err := e3.Open(dbPath); err != nil {
		t.Fatalf("reopen3: %v", err)
	}
	rp3, err := e3.ReadPage(1)
	if err != nil {
		t.Fatalf("read3: %v", err)
	}
	if string(rp3.Data[:len(p.Data)]) != string(p.Data) {
		t.Fatalf("data lost after wal truncate")
	}
	_ = e3.Close()
}
