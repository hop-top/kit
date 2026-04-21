package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"hop.top/kit/go/storage/blob"

	_ "modernc.org/sqlite"
)

// BackupBeforeMigrate copies the database file at dbPath to a timestamped
// backup in the same directory. The backup name is:
//
//	<basename>.pre-v<version>.<20060102-150405>.bak
//
// Returns the backup path on success. If the source file does not exist
// (first run), returns ("", nil) — there is nothing to back up.
func BackupBeforeMigrate(dbPath string, version int) (string, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", nil
	}

	if err := walCheckpoint(dbPath); err != nil {
		return "", fmt.Errorf("backup: checkpoint: %w", err)
	}

	ts := time.Now().UTC().Format("20060102-150405")
	base := filepath.Base(dbPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	backupName := fmt.Sprintf("%s.pre-v%d.%s.bak", name, version, ts)
	backupPath := filepath.Join(filepath.Dir(dbPath), backupName)

	if err := copyFile(dbPath, backupPath); err != nil {
		return "", fmt.Errorf("backup %s: %w", dbPath, err)
	}
	return backupPath, nil
}

// Backup dumps the SQLite database file at dbPath into the given blob.Store
// under key. If dest is nil, falls back to local file copy alongside the DB.
func Backup(ctx context.Context, dbPath string, dest blob.Store, key string) error {
	if dest == nil {
		_, err := BackupBeforeMigrate(dbPath, 0)
		return err
	}

	if err := walCheckpoint(dbPath); err != nil {
		return fmt.Errorf("backup: checkpoint: %w", err)
	}

	f, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("backup: open db: %w", err)
	}
	defer f.Close()

	if err := dest.Put(ctx, key, f, "application/x-sqlite3"); err != nil {
		return fmt.Errorf("backup: blob put: %w", err)
	}
	return nil
}

// Restore retrieves a SQLite backup from src blob.Store at key and writes
// it to dbPath, overwriting any existing file. If src is nil, this is a
// no-op (no remote store configured).
func Restore(ctx context.Context, dbPath string, src blob.Store, key string) error {
	if src == nil {
		return nil
	}

	rc, err := src.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("restore: blob get: %w", err)
	}
	defer rc.Close()

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		return fmt.Errorf("restore: mkdir: %w", err)
	}

	tmp := dbPath + ".restore.tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("restore: create tmp: %w", err)
	}

	if _, err := io.Copy(out, rc); err != nil {
		out.Close()
		os.Remove(tmp)
		return fmt.Errorf("restore: write: %w", err)
	}
	if err := out.Sync(); err != nil {
		out.Close()
		os.Remove(tmp)
		return fmt.Errorf("restore: sync: %w", err)
	}
	out.Close()

	if err := os.Rename(tmp, dbPath); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("restore: rename: %w", err)
	}
	return nil
}

func walCheckpoint(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil // best-effort; file may not be a valid DB yet
	}
	defer db.Close()
	// Ignore errors: file may not be a WAL-mode DB or may be empty.
	_, _ = db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
