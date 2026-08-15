package catalog

import (
	"log"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to path via a temp file in path's own
// directory followed by os.Rename, so a crash mid-write can never leave a
// truncated file at path. The full durability contract, in order:
//
//  1. The temp file is created in path's OWN directory -- os.Rename only
//     guarantees atomicity within a single filesystem, and the shared
//     system temp directory is commonly on a different filesystem than a
//     removable-media destination, which would make the rename fail
//     outright rather than atomically replace the file.
//  2. The temp file's own bytes are flushed to stable storage with
//     File.Sync() before it is closed -- os.Rename's atomicity says
//     nothing about whether the data written to the temp file ever
//     reached disk.
//  3. os.Rename atomically replaces the destination with the temp file.
//  4. The destination's parent directory is best-effort fsync'd after the
//     rename, so the directory entry that now points at the new contents
//     is durable too -- on POSIX this is a separate durability domain from
//     the file's own data (see syncDir). This step is deliberately
//     best-effort: it is not supported uniformly across platforms (see
//     27-CONTEXT.md's Claude's Discretion item on parent-directory fsync,
//     and this plan's <recorded_decision id="parent-directory-fsync">).
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
	if err := tmp.Sync(); err != nil {
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

	// Best-effort parent-directory fsync. The rename has already
	// succeeded and the file's own bytes are already durable (step 2
	// above), so a directory-sync failure here is a weaker-durability
	// outcome, never a failed write -- the destination is NOT removed
	// and this error is NOT propagated. Windows in particular does not
	// expose a directory handle that can be synced this way, and
	// propagating that platform's error would make every write on
	// Windows fail.
	//
	// The error is still deliberately logged (not silently discarded
	// with `_ = syncDir(...)`): a directory sync that fails
	// persistently on some filesystem/platform would otherwise leave
	// exactly the durability hole this fsync exists to close, while
	// every write still reports success -- permanently unobservable.
	// This is deliberate log-and-continue, not a swallowed check.
	if dirErr := syncDir(filepath.Dir(path)); dirErr != nil {
		log.Printf("WriteFileAtomic: parent-directory sync failed for %s: %v", filepath.Dir(path), dirErr)
	}

	return nil
}

// syncDir opens dir and calls Sync() on the resulting handle, so the
// directory entry created by a preceding os.Rename is flushed to stable
// storage. This is best-effort by design (see WriteFileAtomic's caller):
// Windows does not support syncing a directory handle the same way POSIX
// does, so a caller must not treat this function's error as fatal.
func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	syncErr := d.Sync()
	closeErr := d.Close()
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}
