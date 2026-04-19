package core

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// BackupDataDir writes a gzip-compressed tar archive of dataDir to w.
// Relative paths inside the archive are rooted at "." (collection folders and files).
func BackupDataDir(dataDir string, w io.Writer) error {
	absRoot, err := filepath.Abs(dataDir)
	if err != nil {
		return fmt.Errorf("data dir: %w", err)
	}
	fi, err := os.Stat(absRoot)
	if err != nil {
		return fmt.Errorf("stat data dir: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("data path is not a directory: %s", absRoot)
	}

	gw := gzip.NewWriter(w)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = rel
		hdr.Format = tar.FormatGNU

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}
		return nil
	})
}

// RestoreDataDir extracts a gzip-compressed tar from r into dataDir.
// The destination directory is created if needed. Existing files are overwritten.
// Paths are sanitized to prevent directory traversal.
func RestoreDataDir(dataDir string, r io.Reader) error {
	absRoot, err := filepath.Abs(dataDir)
	if err != nil {
		return fmt.Errorf("data dir: %w", err)
	}
	if err := os.MkdirAll(absRoot, 0755); err != nil {
		return fmt.Errorf("mkdir data dir: %w", err)
	}

	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}
		name := filepath.Clean(hdr.Name)
		if name == "." || strings.HasPrefix(name, "..") {
			continue
		}

		target := filepath.Join(absRoot, name)
		tgtAbs, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(tgtAbs+string(os.PathSeparator), absRoot+string(os.PathSeparator)) && tgtAbs != absRoot {
			return fmt.Errorf("illegal path in archive: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(tgtAbs, 0755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(tgtAbs), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(tgtAbs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode&0777))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		default:
			// Skip symlinks and special entries for safety
			continue
		}
	}
	return nil
}

// Backup streams a gzipped tar of the database data directory to w.
func (db *VittoriaDB) Backup(ctx context.Context, w io.Writer) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.closed {
		return fmt.Errorf("database is closed")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return BackupDataDir(db.dataDir, w)
}

// Restore extracts a gzipped tar into the database data directory.
// The database must be closed; use only during offline maintenance or before Open.
func (db *VittoriaDB) Restore(ctx context.Context, r io.Reader) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if !db.closed {
		return fmt.Errorf("close the database before restore")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return RestoreDataDir(db.dataDir, r)
}
