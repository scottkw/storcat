package catalog

import "storcat-wails/pkg/models"

// ComputeDiff compares old against new (both full catalog trees, or nil for
// "no old tree exists" -- STATE-03's unreadable-catalog path) and
// categorizes every path encountered across the two into one of the four
// states this task ships (added, removed, changed, unchanged); the fifth
// (unreadable) is plan 28-02's walk-time marker and always reports zero
// here. Pure: no I/O, imports only pkg/models.
//
// Directories are compared for existence only -- added or removed, never
// changed. This is 28-CONTEXT.md's resolved research question (A2):
// driving a directory's changed state from its own ModTime would double-
// count every file add/remove alongside that file's own added/removed row,
// since a directory's mtime updates on immediate-child changes on most
// filesystems.
func ComputeDiff(old, new *models.CatalogItem) *models.DiffResult {
	oldFlat := flatten(old)
	newFlat := flatten(new)

	result := &models.DiffResult{Entries: []models.DiffEntry{}}

	for path, newItem := range newFlat {
		oldItem, existed := oldFlat[path]
		if !existed {
			result.Added++
			result.Entries = append(result.Entries, models.DiffEntry{
				Path: path, State: models.DiffAdded, Type: newItem.Type, NewSize: newItem.Size,
			})
			continue
		}
		if newItem.Type == "file" && fileChanged(oldItem, newItem) {
			result.Changed++
			result.Entries = append(result.Entries, models.DiffEntry{
				Path: path, State: models.DiffChanged, Type: newItem.Type,
				OldSize: oldItem.Size, NewSize: newItem.Size,
			})
			continue
		}
		result.Unchanged++
	}

	for path, oldItem := range oldFlat {
		if _, stillThere := newFlat[path]; stillThere {
			continue
		}
		result.Removed++
		result.Entries = append(result.Entries, models.DiffEntry{
			Path: path, State: models.DiffRemoved, Type: oldItem.Type, OldSize: oldItem.Size,
		})
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
