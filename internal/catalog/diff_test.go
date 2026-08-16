package catalog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// TestDiff_UnreadableIsNotRemoved verifies the phase's primary data-
// integrity control (T-28-05): a subtree the re-scan cannot read (root
// still reachable, MarkUnreadableOnSkip's marker) reports `unreadable`, and
// its previously-known descendants -- files the old tree recorded inside
// it -- are never reported `removed`, since the new walk had no visibility
// into them this pass and "removed" would be a false claim.
func TestDiff_UnreadableIsNotRemoved(t *testing.T) {
	old := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "directory", Name: "./locked", Size: 20, Contents: []*models.CatalogItem{
				{Type: "file", Name: "./locked/secret.txt", Size: 20, ModTime: 100},
			}},
		},
	}
	newTree := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{
				Type: "directory", Name: "./locked", Size: 0, Contents: []*models.CatalogItem{},
				Unreadable: true, ReadError: "permission denied",
			},
		},
	}

	result := ComputeDiff(old, newTree)

	if result.Removed != 0 {
		t.Errorf("Removed = %d, want 0 -- an unreadable subtree (including its previously known contents) must never be counted removed", result.Removed)
	}
	if result.Unreadable != 1 {
		t.Errorf("Unreadable = %d, want 1", result.Unreadable)
	}
	if result.Added != 0 || result.Changed != 0 {
		t.Errorf("Added=%d Changed=%d, want 0,0", result.Added, result.Changed)
	}
}

// TestDiff_UnreadableCarriesReadError verifies an unreadable entry's row
// carries the node's own ReadError as its reason, with NewSize left at
// zero -- no size is knowable for an entry that failed to read.
func TestDiff_UnreadableCarriesReadError(t *testing.T) {
	old := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "directory", Name: "./locked", Size: 0, Contents: []*models.CatalogItem{}},
		},
	}
	newTree := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{
				Type: "directory", Name: "./locked", Size: 0, Contents: []*models.CatalogItem{},
				Unreadable: true, ReadError: "permission denied",
			},
		},
	}

	result := ComputeDiff(old, newTree)

	var entry *models.DiffEntry
	for i := range result.Entries {
		if result.Entries[i].Path == "./locked" {
			entry = &result.Entries[i]
		}
	}
	if entry == nil {
		t.Fatal("expected a diff entry for ./locked")
	}
	if entry.State != models.DiffUnreadable {
		t.Errorf("State = %v, want DiffUnreadable", entry.State)
	}
	if entry.ReadError != "permission denied" {
		t.Errorf("ReadError = %q, want %q", entry.ReadError, "permission denied")
	}
	if entry.NewSize != 0 {
		t.Errorf("NewSize = %d, want 0 -- no size is knowable for an unreadable entry", entry.NewSize)
	}
}

// TestDiff_TypeChangeYieldsRemovedAndAdded verifies a path that is a file
// in the old tree and a directory in the new tree (or vice versa) yields
// TWO rows -- one removed, one added -- never a single changed row
// (28-RESEARCH.md Assumption A3).
func TestDiff_TypeChangeYieldsRemovedAndAdded(t *testing.T) {
	old := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./thing", Size: 10, ModTime: 100},
		},
	}
	newTree := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "directory", Name: "./thing", Size: 0, Contents: []*models.CatalogItem{}},
		},
	}

	result := ComputeDiff(old, newTree)

	if result.Removed != 1 || result.Added != 1 {
		t.Errorf("Removed=%d Added=%d, want 1,1", result.Removed, result.Added)
	}
	if result.Changed != 0 {
		t.Errorf("Changed = %d, want 0 (a type change is never reported as changed)", result.Changed)
	}

	var removedEntry, addedEntry *models.DiffEntry
	for i := range result.Entries {
		e := &result.Entries[i]
		if e.Path != "./thing" {
			continue
		}
		switch e.State {
		case models.DiffRemoved:
			removedEntry = e
		case models.DiffAdded:
			addedEntry = e
		}
	}
	if removedEntry == nil || removedEntry.Type != "file" {
		t.Errorf("expected a removed entry with Type=file for ./thing, got %+v", removedEntry)
	}
	if addedEntry == nil || addedEntry.Type != "directory" {
		t.Errorf("expected an added entry with Type=directory for ./thing, got %+v", addedEntry)
	}
}

// TestDiff_CountsSumToDistinctPaths asserts the sum invariant directly
// (not just each category's individual count) over a fixture covering all
// five states plus a pruned unreadable-descendant.
func TestDiff_CountsSumToDistinctPaths(t *testing.T) {
	old := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./same.txt", Size: 10, ModTime: 100},
			{Type: "file", Name: "./resized.txt", Size: 20, ModTime: 100},
			{Type: "file", Name: "./gone.txt", Size: 30, ModTime: 100},
			{Type: "directory", Name: "./locked", Size: 5, Contents: []*models.CatalogItem{
				{Type: "file", Name: "./locked/hidden.txt", Size: 5, ModTime: 100},
			}},
		},
	}
	newTree := &models.CatalogItem{
		Type: "directory", Name: "./",
		Contents: []*models.CatalogItem{
			{Type: "file", Name: "./same.txt", Size: 10, ModTime: 100},
			{Type: "file", Name: "./resized.txt", Size: 25, ModTime: 100},
			{Type: "file", Name: "./fresh.txt", Size: 8, ModTime: 200},
			{
				Type: "directory", Name: "./locked", Size: 0, Contents: []*models.CatalogItem{},
				Unreadable: true, ReadError: "permission denied",
			},
		},
	}

	result := ComputeDiff(old, newTree)

	sum := result.Added + result.Removed + result.Changed + result.Unreadable + result.Unchanged
	// Distinct diffable paths: same.txt, resized.txt, gone.txt, locked,
	// fresh.txt = 5. locked/hidden.txt is pruned -- a descendant of an
	// unreadable node, never a diffable path this scan. No type-change
	// pair in this fixture, so no +1 extra.
	if sum != 5 {
		t.Errorf("category sum = %d, want 5", sum)
	}
	if result.Added != 1 || result.Removed != 1 || result.Changed != 1 || result.Unreadable != 1 || result.Unchanged != 1 {
		t.Errorf("got Added=%d Removed=%d Changed=%d Unreadable=%d Unchanged=%d, want 1 each",
			result.Added, result.Removed, result.Changed, result.Unreadable, result.Unchanged)
	}
}

// TestDiff_LowSimilarityBelowFloor verifies a small old tree (below the
// minimum-entries floor) replaced entirely never sets LowSimilarity, even
// though its own add/remove ratio is 100% -- a handful of entries changing
// is not evidence of a wrong-disc pick.
func TestDiff_LowSimilarityBelowFloor(t *testing.T) {
	old := &models.CatalogItem{Type: "directory", Name: "./"}
	for i := 0; i < 5; i++ {
		old.Contents = append(old.Contents, &models.CatalogItem{
			Type: "file", Name: fmt.Sprintf("./old%d.txt", i), Size: 10,
		})
	}
	newTree := &models.CatalogItem{Type: "directory", Name: "./"}
	for i := 0; i < 5; i++ {
		newTree.Contents = append(newTree.Contents, &models.CatalogItem{
			Type: "file", Name: fmt.Sprintf("./new%d.txt", i), Size: 10,
		})
	}

	result := ComputeDiff(old, newTree)

	if result.OldEntryCount != 5 {
		t.Errorf("OldEntryCount = %d, want 5", result.OldEntryCount)
	}
	if result.LowSimilarity {
		t.Error("LowSimilarity = true, want false -- 5 old entries is below the floor")
	}
}

// TestDiff_LowSimilarityAboveThreshold verifies an old tree at or above the
// floor, replaced entirely, sets LowSimilarity true.
func TestDiff_LowSimilarityAboveThreshold(t *testing.T) {
	old := &models.CatalogItem{Type: "directory", Name: "./"}
	for i := 0; i < 25; i++ {
		old.Contents = append(old.Contents, &models.CatalogItem{
			Type: "file", Name: fmt.Sprintf("./old%d.txt", i), Size: 10,
		})
	}
	newTree := &models.CatalogItem{Type: "directory", Name: "./"}
	for i := 0; i < 25; i++ {
		newTree.Contents = append(newTree.Contents, &models.CatalogItem{
			Type: "file", Name: fmt.Sprintf("./new%d.txt", i), Size: 10,
		})
	}

	result := ComputeDiff(old, newTree)

	if result.OldEntryCount != 25 {
		t.Errorf("OldEntryCount = %d, want 25", result.OldEntryCount)
	}
	// All 25 old entries removed, all 25 new entries added -- total 50,
	// ratio 50/50 = 1.0, well above the 0.6 threshold.
	if !result.LowSimilarity {
		t.Error("LowSimilarity = false, want true -- entirely replaced catalog above the floor")
	}
}

// TestComputeDiff_EndToEndWithRealUnreadableSubdirectory is the phase's own
// end-to-end verification (28-02-PLAN.md's <verification>): a real
// chmod 000 subdirectory, walked through the actual Service.Walk with
// MarkUnreadableOnSkip set (re-scan's exact call shape), diffed against a
// prior tree that had visibility into that subdirectory's contents --
// confirming the whole pipeline (not just each half in isolation) reports
// the locked path `unreadable` and reports zero `removed`.
func TestComputeDiff_EndToEndWithRealUnreadableSubdirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix-style permission bits don't apply on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses permission bits")
	}

	dir := t.TempDir()
	locked := filepath.Join(dir, "locked")
	if err := os.Mkdir(locked, 0755); err != nil {
		t.Fatalf("mkdir locked: %v", err)
	}
	if err := os.WriteFile(filepath.Join(locked, "secret.txt"), []byte("shh"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	s := NewService()
	// oldTree: a prior scan that could still see inside locked.
	oldTree, err := s.Walk(context.Background(), dir, Options{}, nil)
	if err != nil {
		t.Fatalf("Walk (old): %v", err)
	}

	if err := os.Chmod(locked, 0000); err != nil {
		t.Fatalf("chmod locked: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0755) })

	// newTree: re-scan's exact call shape -- MarkUnreadableOnSkip true.
	newOpts := Options{}
	newOpts.MarkUnreadableOnSkip = true
	newTree, err := s.Walk(context.Background(), dir, newOpts, nil)
	if err != nil {
		t.Fatalf("Walk (new): %v", err)
	}

	result := ComputeDiff(oldTree, newTree)

	if result.Removed != 0 {
		t.Errorf("Removed = %d, want 0", result.Removed)
	}
	if result.Unreadable != 1 {
		t.Errorf("Unreadable = %d, want 1", result.Unreadable)
	}
}
