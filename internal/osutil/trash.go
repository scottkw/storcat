package osutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bios-Marcel/wastebasket/v2"
)

// trashSeam is the only deletion mechanism TrashPaths may ever reach.
// Package-level (not exported) so trash_test.go, which lives in this same
// package, can swap it for a recording fake -- no test in this package
// touches a real OS Trash. Initialized to wastebasket.Trash, which
// 27-RESEARCH.md verified by reading all three platform backends in full:
// it never falls back to permanent deletion on any of them, and every
// failure path returns an error rather than silently succeeding.
var trashSeam = wastebasket.Trash

// TrashPaths moves every path in paths to the OS Trash, after resolving and
// validating each one -- stat'd, symlink-resolved, checked for regular-file
// status, checked against the .json/.html extension allowlist, and checked
// for containment inside catalogDir -- exactly mirroring
// RevealInFileManager's validate-everything-before-acting ordering.
//
// A path that is already missing (os.IsNotExist) is silently skipped, not
// an error: 27-RESEARCH.md verified that every wastebasket backend already
// continues past a missing path internally, so re-invoking TrashPaths with
// the same path set after a partial failure re-attempts only what is still
// there and succeeds once nothing remains. If every supplied path is
// already gone, the seam is never called and this returns nil.
//
// This function performs no filesystem removal of its own -- trashSeam is
// the only deletion mechanism reachable from here. A failure returned by
// the seam is wrapped with %w (so errors.Is/errors.As still reach the
// underlying OS error) and returned verbatim; there is no local fallback of
// any kind. ACT-05 (never permanently delete) is satisfied by the absence
// of a local removal call here, and independently by wastebasket.Trash's
// own behavior on every platform it supports.
//
// Why the containment gate matters even more here than for
// RevealInFileManager: wastebasket's macOS backend shells out to osascript
// with a hand-built AppleScript string that escapes only literal double
// quotes in the path before interpolating it -- weaker than this codebase's
// own exec.Command-argv-per-element convention, and third-party code this
// phase cannot change. The gate bounds the worst case to "a .json/.html
// file this app's own os.ReadDir listing already resolved inside the
// configured catalog directory." The paths this helper receives always
// originate from BrowseCatalogs's own os.ReadDir listing, never renderer
// free-text; a future feature that lets a user type an arbitrary filename
// would need this constraint revisited (27-RESEARCH.md Assumption A4).
func TrashPaths(catalogDir string, paths ...string) error {
	if catalogDir == "" {
		return fmt.Errorf("trash: no catalog directory configured")
	}

	var resolvedPaths []string
	for _, p := range paths {
		if _, err := os.Lstat(p); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("trash %s: %w", p, err)
		}

		absPath, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("trash %s: resolve path: %w", p, err)
		}
		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return fmt.Errorf("trash %s: %w", p, err)
		}

		info, err := os.Stat(resolved)
		if err != nil {
			return fmt.Errorf("trash %s: %w", p, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("trash %s: not a regular file", p)
		}

		ext := strings.ToLower(filepath.Ext(resolved))
		if !allowedRevealExtensions[ext] {
			return fmt.Errorf("trash %s: unsupported extension %q", p, ext)
		}

		ok, err := ContainsPath(catalogDir, resolved)
		if err != nil {
			return fmt.Errorf("trash %s: resolve catalog directory: %w", p, err)
		}
		if !ok {
			return fmt.Errorf("trash %s: outside configured catalog directory", p)
		}

		resolvedPaths = append(resolvedPaths, resolved)
	}

	if len(resolvedPaths) == 0 {
		return nil
	}

	if err := trashSeam(resolvedPaths...); err != nil {
		return fmt.Errorf("trash: %w", err)
	}
	return nil
}
