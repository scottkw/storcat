package catalog

import (
	"context"
	"errors"
	"fmt"

	"storcat-wails/pkg/models"
)

// Walk builds an in-memory catalog tree by traversing sourcePath -- zero
// writes. This is the walk half of CreateCatalogWithContext's original
// walk-then-write body (service.go:140-200 before this extraction), moved
// here verbatim so a caller that never intends to write (re-scan's diff
// path, app.go's RescanCatalog) can call it directly without pulling in
// WriteCatalogFrom. CreateCatalogWithContext itself now reduces to Walk
// then WriteCatalogFrom -- Walk is the primary, promoted operation; Create
// is one variant of walking (the one that also writes).
//
// Behavior is byte-for-byte identical to the pre-extraction inline block:
// TestCreateCatalog_JSONShapeUnchanged and the source-loss tests in
// service_test.go assert this and must pass UNMODIFIED.
func (s *Service) Walk(ctx context.Context, sourcePath string, opts Options, onProgress ProgressCallback) (*models.CatalogItem, error) {
	st := &walkState{
		scanRoot:   sourcePath,
		opts:       opts,
		onProgress: onProgress,
	}

	tree, err := s.traverseDirectory(ctx, sourcePath, sourcePath, st)
	if err != nil {
		// Three outcomes, distinguished before touching the write path: a
		// cancelled context writes nothing; a source-loss error is returned
		// with its populated partial scan attached and writes nothing;
		// anything else is a genuine traversal failure.
		var srcErr *SourceUnavailableError
		if errors.As(err, &srcErr) {
			srcErr.Partial = &PartialScan{
				Tree:       tree,
				FilesSeen:  st.filesSeen,
				BytesSeen:  st.bytesSeen,
				ReadErrors: st.readErrorEntries,
			}
			return nil, srcErr
		}
		// traverseDirectory's top-of-function os.Stat/os.ReadDir failure has
		// no parent loop to run it through recordReadError+classify() the
		// way every deeper (child) failure already does when THIS call is
		// itself the recursive callee -- for the outermost call (the scan
		// root), this is that missing parent. Without this check, the scan
		// root vanishing before a single child was ever read (the common
		// case for a genuinely ejected volume) would silently fall through
		// as a generic traversal error instead of the source-loss
		// classification CRT-10 requires.
		if ctx.Err() == nil && st.classify() {
			return nil, &SourceUnavailableError{
				SourcePath: sourcePath,
				Partial: &PartialScan{
					// A nil Tree would marshal to a bare JSON "null", not a
					// valid catalog -- constructed as the same marked-
					// unreadable empty-directory shape the child-level path
					// already uses, so a partial write from this state
					// still produces a valid, honest catalog whose root
					// records that nothing could be read.
					Tree: &models.CatalogItem{
						Type:       "directory",
						Name:       "./",
						Size:       0,
						Contents:   []*models.CatalogItem{},
						Unreadable: true,
						ReadError:  err.Error(),
					},
					FilesSeen:  st.filesSeen,
					BytesSeen:  st.bytesSeen,
					ReadErrors: st.readErrorEntries,
				},
			}
		}
		return nil, fmt.Errorf("failed to traverse directory: %w", err)
	}

	// Re-check cancellation after the walk completes and before returning --
	// traverseDirectory can return a non-error partial tree in some
	// skip-and-continue paths, so this is the authoritative gate.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return tree, nil
}
