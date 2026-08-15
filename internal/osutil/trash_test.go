package osutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// swapTrashSeam replaces the package-level trash seam for the duration of
// the calling test, restoring the original on cleanup -- so no test in this
// file ever reaches a real OS Trash.
func swapTrashSeam(t *testing.T, fn func(paths ...string) error) {
	t.Helper()
	original := trashSeam
	trashSeam = fn
	t.Cleanup(func() { trashSeam = original })
}

func TestTrashPaths_PassesResolvedPathsToSeam(t *testing.T) {
	dir := t.TempDir()
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks(dir): %v", err)
	}
	a := filepath.Join(dir, "a.json")
	b := filepath.Join(dir, "b.json")
	if err := os.WriteFile(a, []byte("{}"), 0644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(b, []byte("{}"), 0644); err != nil {
		t.Fatalf("write b: %v", err)
	}

	var captured []string
	swapTrashSeam(t, func(paths ...string) error {
		captured = append(captured, paths...)
		return nil
	})

	if err := TrashPaths(dir, a, b); err != nil {
		t.Fatalf("TrashPaths: %v", err)
	}
	want := []string{
		filepath.Join(resolvedDir, "a.json"),
		filepath.Join(resolvedDir, "b.json"),
	}
	if !equalSlices(captured, want) {
		t.Errorf("captured = %#v, want %#v", captured, want)
	}
}

func TestTrashPaths_RejectsOutsideCatalogDir(t *testing.T) {
	base := t.TempDir()
	catalogDir := filepath.Join(base, "catalogs")
	outsideDir := filepath.Join(base, "outside")
	if err := os.Mkdir(catalogDir, 0755); err != nil {
		t.Fatalf("mkdir catalogDir: %v", err)
	}
	if err := os.Mkdir(outsideDir, 0755); err != nil {
		t.Fatalf("mkdir outsideDir: %v", err)
	}
	path := filepath.Join(outsideDir, "catalog.json")
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	seamCalled := false
	swapTrashSeam(t, func(paths ...string) error {
		seamCalled = true
		return nil
	})

	err := TrashPaths(catalogDir, path)
	if err == nil {
		t.Fatal("expected an error for a path outside the configured catalog directory, got nil")
	}
	if seamCalled {
		t.Error("seam must not be called when a path is rejected")
	}
}

// TestTrashPaths_RejectsSiblingPrefixDirectory proves containment is
// filepath.Rel-based, not a naive string prefix check: "<tmp>/catalogs-evil"
// has "<tmp>/catalogs" as a string prefix but is not inside it.
func TestTrashPaths_RejectsSiblingPrefixDirectory(t *testing.T) {
	base := t.TempDir()
	catalogDir := filepath.Join(base, "catalogs")
	evilDir := filepath.Join(base, "catalogs-evil")
	if err := os.Mkdir(catalogDir, 0755); err != nil {
		t.Fatalf("mkdir catalogDir: %v", err)
	}
	if err := os.Mkdir(evilDir, 0755); err != nil {
		t.Fatalf("mkdir evilDir: %v", err)
	}
	path := filepath.Join(evilDir, "catalog.json")
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	seamCalled := false
	swapTrashSeam(t, func(paths ...string) error {
		seamCalled = true
		return nil
	})

	err := TrashPaths(catalogDir, path)
	if err == nil {
		t.Fatal("expected an error for a sibling prefix-sharing directory, got nil")
	}
	if seamCalled {
		t.Error("seam must not be called when a path is rejected")
	}
}

func TestTrashPaths_RejectsDisallowedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-a-catalog.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	seamCalled := false
	swapTrashSeam(t, func(paths ...string) error {
		seamCalled = true
		return nil
	})

	err := TrashPaths(dir, path)
	if err == nil {
		t.Fatal("expected an error for a disallowed extension, got nil")
	}
	if seamCalled {
		t.Error("seam must not be called when a path is rejected")
	}
}

func TestTrashPaths_RejectsDirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir.json")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	seamCalled := false
	swapTrashSeam(t, func(paths ...string) error {
		seamCalled = true
		return nil
	})

	err := TrashPaths(dir, sub)
	if err == nil {
		t.Fatal("expected an error for a directory path, got nil")
	}
	if seamCalled {
		t.Error("seam must not be called when a path is rejected")
	}
}

// TestTrashPaths_SkipsMissingPath is the retry case: after a partial
// failure, re-invoking with the same path set must only re-attempt what is
// still on disk.
func TestTrashPaths_SkipsMissingPath(t *testing.T) {
	dir := t.TempDir()
	present := filepath.Join(dir, "present.json")
	missing := filepath.Join(dir, "missing.json")
	if err := os.WriteFile(present, []byte("{}"), 0644); err != nil {
		t.Fatalf("write present: %v", err)
	}

	var captured []string
	swapTrashSeam(t, func(paths ...string) error {
		captured = append(captured, paths...)
		return nil
	})

	if err := TrashPaths(dir, present, missing); err != nil {
		t.Fatalf("TrashPaths: %v", err)
	}
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks(dir): %v", err)
	}
	want := []string{filepath.Join(resolvedDir, "present.json")}
	if !equalSlices(captured, want) {
		t.Errorf("captured = %#v, want %#v", captured, want)
	}
}

func TestTrashPaths_AllMissingReturnsNil(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "gone-a.json")
	b := filepath.Join(dir, "gone-b.json")

	seamCalled := false
	swapTrashSeam(t, func(paths ...string) error {
		seamCalled = true
		return nil
	})

	err := TrashPaths(dir, a, b)
	if err != nil {
		t.Fatalf("TrashPaths: expected nil for all-missing paths, got %v", err)
	}
	if seamCalled {
		t.Error("seam must not be called when every supplied path is already gone")
	}
}

func TestTrashPaths_EmptyCatalogDirIsAnError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.json")
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	seamCalled := false
	swapTrashSeam(t, func(paths ...string) error {
		seamCalled = true
		return nil
	})

	err := TrashPaths("", path)
	if err == nil {
		t.Fatal("expected an error for an empty catalog directory, got nil")
	}
	if seamCalled {
		t.Error("seam must not be called when catalogDir is empty")
	}
}

var errSeamSentinel = errors.New("seam failed: permission denied")

func TestTrashPaths_PropagatesSeamErrorVerbatim(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.json")
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	swapTrashSeam(t, func(paths ...string) error {
		return errSeamSentinel
	})

	err := TrashPaths(dir, path)
	if err == nil {
		t.Fatal("expected an error propagated from the seam, got nil")
	}
	if !errors.Is(err, errSeamSentinel) {
		t.Errorf("errors.Is(err, errSeamSentinel) = false, want true; err = %v", err)
	}
}

// TestTrashPaths_NonASCIIPathReachesSeamByteIdentical proves no escaping,
// quoting, or re-encoding is applied by this code before the seam is
// called -- the underlying trash library's macOS backend does its own
// (weaker) escaping downstream, but this package must hand it the exact
// resolved path.
func TestTrashPaths_NonASCIIPathReachesSeamByteIdentical(t *testing.T) {
	dir := t.TempDir()
	name := "카탈로그 report café.json"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks(dir): %v", err)
	}
	want := filepath.Join(resolvedDir, name)

	var captured []string
	swapTrashSeam(t, func(paths ...string) error {
		captured = append(captured, paths...)
		return nil
	})

	if err := TrashPaths(dir, path); err != nil {
		t.Fatalf("TrashPaths: %v", err)
	}
	if len(captured) != 1 || captured[0] != want {
		t.Errorf("captured = %#v, want [%q]", captured, want)
	}
}

func TestTrashPaths_NoPathsIsNil(t *testing.T) {
	dir := t.TempDir()

	seamCalled := false
	swapTrashSeam(t, func(paths ...string) error {
		seamCalled = true
		return nil
	})

	if err := TrashPaths(dir); err != nil {
		t.Fatalf("TrashPaths with no paths: expected nil, got %v", err)
	}
	if seamCalled {
		t.Error("seam must not be called with zero paths")
	}
}
