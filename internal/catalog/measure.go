package catalog

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// MeasureTree is a recursive, count-only pass over root: it reports the
// same (files, bytes, readErrors)-shaped progress traverseDirectory
// reports, via the identical ProgressCallback type, but builds no tree at
// all. It exists only for the plain-folder case where no volume total is
// available (CRT-03's "choose any folder" path, or a volume the picker
// never had a size for) -- running a full two-pass count for every scan
// was explicitly rejected (25-CONTEXT.md): it doubles I/O at this
// milestone's target scale. Callers must never invoke this when a real
// total is already known; that branch belongs to the caller
// (App.resolveScanTotal), not to this function.
//
// It applies the identical hidden-file inclusion rule traverseDirectory
// applies (opts.IncludeHidden), and the identical single-entry tolerance:
// an unreadable subdirectory or an entry whose stat fails is skipped,
// incrementing readErrors, never aborting the measurement -- so a
// pre-pass count and the walk that follows it never disagree about which
// entries were counted. ctx is checked at the top of every directory
// boundary, matching traverseDirectory's cancellation contract: a
// cancelled context aborts the measurement and returns ctx.Err().
func MeasureTree(ctx context.Context, root string, opts Options, onProgress ProgressCallback) (files int, bytes int64, err error) {
	var readErrors int

	report := func(path string) {
		if onProgress == nil {
			return
		}
		onProgress(ProgressUpdate{Path: path, FilesSeen: files, BytesSeen: bytes, ReadErrors: readErrors})
	}

	var walk func(dirPath string) error
	walk = func(dirPath string) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		info, statErr := os.Stat(dirPath)
		if statErr != nil {
			// A single entry's stat failing is tolerated, exactly like
			// traverseDirectory's skip-and-continue path -- not the
			// terminal classification that requires re-probing the scan
			// root, which is out of scope for a count-only pre-pass.
			readErrors++
			return nil
		}

		if info.Mode().IsRegular() {
			files++
			bytes += info.Size()
			report(dirPath)
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		entries, readDirErr := os.ReadDir(dirPath)
		if readDirErr != nil {
			readErrors++
			return nil
		}

		for _, entry := range entries {
			if !opts.IncludeHidden && strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			if err := walk(filepath.Join(dirPath, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	}

	if walkErr := walk(root); walkErr != nil {
		return files, bytes, walkErr
	}
	return files, bytes, nil
}
