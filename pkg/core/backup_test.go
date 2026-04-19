package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBackupRestoreDataDir(t *testing.T) {
	src := t.TempDir()
	_ = os.MkdirAll(filepath.Join(src, "c1"), 0755)
	if err := os.WriteFile(filepath.Join(src, "c1", "f.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := BackupDataDir(src, &buf); err != nil {
		t.Fatalf("backup: %v", err)
	}

	dst := t.TempDir()
	if err := RestoreDataDir(dst, bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("restore: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(dst, "c1", "f.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "x" {
		t.Fatalf("content: %q", b)
	}
}
