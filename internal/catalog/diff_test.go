package catalog

import (
	"testing"

	"storcat-wails/pkg/models"
)

// TestDiff_AddedRemovedChangedUnchanged covers the four core categories in
// one tree pair: a pure add, a pure remove, a size-change, and an untouched
// file that reports unchanged.
func TestDiff_AddedRemovedChangedUnchanged(t *testing.T) {
	old := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./same.txt", Size: 10, ModTime: 100},
			{Type: "file", Name: "./resized.txt", Size: 20, ModTime: 100},
			{Type: "file", Name: "./gone.txt", Size: 30, ModTime: 100},
		},
	}
	newTree := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./same.txt", Size: 10, ModTime: 100},
			{Type: "file", Name: "./resized.txt", Size: 25, ModTime: 100},
			{Type: "file", Name: "./fresh.txt", Size: 5, ModTime: 200},
		},
	}

	result := ComputeDiff(old, newTree)

	if result.Added != 1 {
		t.Errorf("Added = %d, want 1", result.Added)
	}
	if result.Removed != 1 {
		t.Errorf("Removed = %d, want 1", result.Removed)
	}
	if result.Changed != 1 {
		t.Errorf("Changed = %d, want 1", result.Changed)
	}
	if result.Unchanged != 1 {
		t.Errorf("Unchanged = %d, want 1", result.Unchanged)
	}
	if result.Unreadable != 0 {
		t.Errorf("Unreadable = %d, want 0 (this task ships no unreadable state)", result.Unreadable)
	}

	sum := result.Added + result.Removed + result.Changed + result.Unreadable + result.Unchanged
	if sum != 4 {
		t.Errorf("category sum = %d, want 4 (the total distinct paths across old ∪ new)", sum)
	}
}

// TestDiff_SameSizeMtimeChange verifies that a file whose size is identical
// but whose mtime differs is still reported changed -- CONTEXT's decision to
// compare size AND mtime, not size alone.
func TestDiff_SameSizeMtimeChange(t *testing.T) {
	old := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./edited.txt", Size: 10, ModTime: 100},
		},
	}
	newTree := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./edited.txt", Size: 10, ModTime: 200},
		},
	}

	result := ComputeDiff(old, newTree)

	if result.Changed != 1 {
		t.Errorf("Changed = %d, want 1 (same size, different mtime)", result.Changed)
	}
	if result.Unchanged != 0 {
		t.Errorf("Unchanged = %d, want 0", result.Unchanged)
	}
}

// TestDiff_MissingOldModTimeFallsBackToSizeOnly verifies the load-bearing
// guard: an old entry with ModTime == 0 (a catalog written before Phase 28)
// and an equal size is categorized unchanged, never changed purely because
// the freshly-walked new tree carries a real mtime and the old one carries
// none.
func TestDiff_MissingOldModTimeFallsBackToSizeOnly(t *testing.T) {
	old := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./legacy.txt", Size: 10, ModTime: 0},
		},
	}
	newTree := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./legacy.txt", Size: 10, ModTime: 12345},
		},
	}

	result := ComputeDiff(old, newTree)

	if result.Unchanged != 1 {
		t.Errorf("Unchanged = %d, want 1 (old ModTime == 0 must fall back to size-only)", result.Unchanged)
	}
	if result.Changed != 0 {
		t.Errorf("Changed = %d, want 0", result.Changed)
	}
}

// TestDiff_DirectoryNeverReportsChanged verifies 28-CONTEXT.md's resolved
// research question A2: a directory whose size differs (because a file was
// added inside it) is never itself reported as a changed entry -- only
// added/removed/unchanged apply to directories.
func TestDiff_DirectoryNeverReportsChanged(t *testing.T) {
	old := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "directory", Name: "./sub", Size: 10, ModTime: 100, Contents: []*models.CatalogItem{
				{Type: "file", Name: "./sub/a.txt", Size: 10, ModTime: 100},
			}},
		},
	}
	newTree := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "directory", Name: "./sub", Size: 25, ModTime: 200, Contents: []*models.CatalogItem{
				{Type: "file", Name: "./sub/a.txt", Size: 10, ModTime: 100},
				{Type: "file", Name: "./sub/b.txt", Size: 15, ModTime: 200},
			}},
		},
	}

	result := ComputeDiff(old, newTree)

	if result.Changed != 0 {
		t.Errorf("Changed = %d, want 0 (directories are never diffed as changed)", result.Changed)
	}
	if result.Added != 1 {
		t.Errorf("Added = %d, want 1 (./sub/b.txt)", result.Added)
	}
	// ./sub itself (size/mtime differ) and ./sub/a.txt (untouched) both
	// report unchanged.
	if result.Unchanged != 2 {
		t.Errorf("Unchanged = %d, want 2 (./sub directory + ./sub/a.txt)", result.Unchanged)
	}
}

// TestDiff_NilOldTreeReportsAllAdded covers STATE-03's no-old-tree path
// (oldTreeAvailable: false): every entry in the new tree reports added, and
// flatten(nil) must not panic.
func TestDiff_NilOldTreeReportsAllAdded(t *testing.T) {
	newTree := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./a.txt", Size: 10, ModTime: 100},
		},
	}

	result := ComputeDiff(nil, newTree)

	if result.Added != 1 {
		t.Errorf("Added = %d, want 1", result.Added)
	}
	if result.Removed != 0 || result.Changed != 0 || result.Unchanged != 0 {
		t.Errorf("expected only Added to be non-zero, got %+v", result)
	}
}
