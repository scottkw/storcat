package catalog

import (
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to path via a temp file in path's own
// directory followed by os.Rename, so a crash mid-write can never leave a
// truncated file at path. The temp file must live in the SAME directory as
// path -- os.Rename only guarantees atomicity within a single filesystem,
// and the shared system temp directory is commonly on a different
// filesystem than a removable-media destination, which would make the
// rename fail outright rather than atomically replace the file.
//
// Exported (not package-private) because Phase 27's rename, duplicate, and
// delete-to-Trash operations reuse this exact primitive rather than each
// retrofitting crash-safety later.
//
// Source pattern: internal/config/counts_cache.go:107-135.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "storcat-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// The temp helper above creates the file with restrictive (0600)
	// permissions; the destination must end up at the caller's requested
	// perm.
	if err := os.Chmod(tmpPath, perm); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
