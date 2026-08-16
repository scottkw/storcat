package catalog

import "storcat-wails/pkg/models"

// similarityMinEntries and similarityThreshold gate DiffResult.LowSimilarity
// (the "did I scan the wrong disc?" signal, 28-UI-SPEC.md's Similarity
// Warning Contract). Named package-level constants so both values are
// greppable and tunable in one place. LowSimilarity is a signal only --
// nothing in Go blocks or refuses on it.
const (
	// similarityMinEntries is the floor below which LowSimilarity never
	// fires -- a small catalog replaced entirely is not evidence of a
	// wrong-disc pick, just a small catalog.
	similarityMinEntries = 20
	// similarityThreshold is the (Added+Removed)/total fraction at or
	// above which a diff is flagged low-similarity.
	similarityThreshold = 0.6
)

// ComputeDiff compares old against new (both full catalog trees, or nil for
// "no old tree exists" -- STATE-03's unreadable-catalog path) and
// categorizes every path encountered across the two into one of the five
// diff states (added, removed, changed, unreadable, unchanged). Pure: no
// I/O, imports only pkg/models.
//
// Directories are compared for existence only -- added or removed, never
// changed. This is 28-CONTEXT.md's resolved research question (A2):
// driving a directory's changed state from its own ModTime would double-
// count every file add/remove alongside that file's own added/removed row,
// since a directory's mtime updates on immediate-child changes on most
// filesystems.
//
// A path the new walk marks Unreadable (28-02's MarkUnreadableOnSkip
// option) is categorized `unreadable`, checked BEFORE the type/size/mtime
// comparisons below -- an unreadable node has no meaningful size or type
// comparison to make. Its previously-known descendants (paths that only
// existed because the OLD tree had walked into that now-unreadable
// subtree) are excluded from the diff entirely, not reported `removed`:
// the new walk never had visibility into them this pass, so "removed" would
// be a false claim that could drive a data-destroying overwrite (T-28-05,
// this task's primary data-integrity control -- see 28-RESEARCH.md's
// "load-bearing gap").
//
// A file<->directory type change at the same path emits two entries (one
// removed for the old type, one added for the new) rather than a single
// changed row -- see DiffEntry's doc comment.
func ComputeDiff(old, new *models.CatalogItem) *models.DiffResult {
	oldFlat := flatten(old)
	newFlat := flatten(new)
	oldEntryCount := len(oldFlat)

	// A path whose NEW-side node is unreadable has no visibility into its
	// own descendants (the walk never enumerated them) -- prune any OLD
	// path nested under one before diffing, so those descendants are
	// simply never encountered this scan rather than falsely reported
	// removed.
	unreadable := map[string]struct{}{}
	for path, item := range newFlat {
		if item.Unreadable {
			unreadable[path] = struct{}{}
		}
	}
	if len(unreadable) > 0 {
		for path := range oldFlat {
			if hasUnreadableAncestor(path, unreadable) {
				delete(oldFlat, path)
			}
		}
	}

	entries := []models.DiffEntry{}
	unchanged := 0

	for path, newItem := range newFlat {
		oldItem, existed := oldFlat[path]
		switch {
		case !existed:
			entries = append(entries, models.DiffEntry{
				Path: path, State: models.DiffAdded, Type: newItem.Type, NewSize: newItem.Size,
			})
		case newItem.Unreadable:
			entries = append(entries, models.DiffEntry{
				Path: path, State: models.DiffUnreadable, Type: newItem.Type, ReadError: newItem.ReadError,
			})
		case oldItem.Type != newItem.Type:
			// A different entity that happens to share a path -- not a
			// comparable edit (28-RESEARCH.md Assumption A3). Two rows,
			// both counted.
			entries = append(entries,
				models.DiffEntry{Path: path, State: models.DiffRemoved, Type: oldItem.Type, OldSize: oldItem.Size},
				models.DiffEntry{Path: path, State: models.DiffAdded, Type: newItem.Type, NewSize: newItem.Size},
			)
		case newItem.Type == "file" && fileChanged(oldItem, newItem):
			entries = append(entries, models.DiffEntry{
				Path: path, State: models.DiffChanged, Type: newItem.Type,
				OldSize: oldItem.Size, NewSize: newItem.Size,
			})
		default:
			unchanged++
		}
	}

	for path, oldItem := range oldFlat {
		if _, stillThere := newFlat[path]; stillThere {
			continue // handled above (including a type-change's removed half)
		}
		entries = append(entries, models.DiffEntry{
			Path: path, State: models.DiffRemoved, Type: oldItem.Type, OldSize: oldItem.Size,
		})
	}

	result := &models.DiffResult{Entries: entries, Unchanged: unchanged, OldEntryCount: oldEntryCount}
	for _, e := range entries {
		switch e.State {
		case models.DiffAdded:
			result.Added++
		case models.DiffRemoved:
			result.Removed++
		case models.DiffChanged:
			result.Changed++
		case models.DiffUnreadable:
			result.Unreadable++
		}
	}

	total := result.Added + result.Removed + result.Changed + result.Unreadable + result.Unchanged
	if result.OldEntryCount >= similarityMinEntries && total > 0 {
		ratio := float64(result.Added+result.Removed) / float64(total)
		result.LowSimilarity = ratio >= similarityThreshold
	}

	return result
}

// fileChanged reports whether a file entry's size or mtime differ between
// old and new. old.ModTime == 0 means "unknown, predates this field" (a
// catalog written before Phase 28) -- the guard below falls back to a
// size-only comparison so every entry in a pre-existing catalog does not
// diff as changed purely because the freshly-walked new tree carries a real
// mtime and the old one carries none.
func fileChanged(old, new *models.CatalogItem) bool {
	if old.Size != new.Size {
		return true
	}
	if old.ModTime != 0 && old.ModTime != new.ModTime {
		return true
	}
	return false
}

// hasUnreadableAncestor reports whether path is nested under any path in
// unreadable (a direct or indirect descendant), checked by plain string
// slicing rather than importing "strings" -- diff.go's purity (no I/O, no
// dependency beyond pkg/models) is load-bearing, verified by an automated
// check in this plan.
func hasUnreadableAncestor(path string, unreadable map[string]struct{}) bool {
	for u := range unreadable {
		if len(path) > len(u) && path[:len(u)] == u && path[len(u)] == '/' {
			return true
		}
	}
	return false
}

// flatten walks root's Contents recursively (excluding root itself) into a
// map keyed by Name -- already the full relative display path, set once
// from filepath.Rel by traverseDirectory -- mirroring
// internal/search/flatten.go's LoadCatalogFlat root-exclusion convention.
// Directories are included (needed for their own added/removed comparison)
// but never entered into the changed check above. A nil root (STATE-03's
// no-old-tree case) yields an empty map, not a panic.
func flatten(root *models.CatalogItem) map[string]*models.CatalogItem {
	out := map[string]*models.CatalogItem{}
	if root == nil {
		return out
	}
	var walk func(item *models.CatalogItem)
	walk = func(item *models.CatalogItem) {
		for _, child := range item.Contents {
			out[child.Name] = child
			if child.Type == "directory" {
				walk(child)
			}
		}
	}
	walk(root)
	return out
}
