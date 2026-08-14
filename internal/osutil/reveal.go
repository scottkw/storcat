// Package osutil holds small, per-platform operating-system integrations
// that don't belong in internal/catalog or internal/search. RevealInFileManager
// is its first (and, this phase, only) export.
package osutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// revealArgvDarwin builds the argument vector for macOS's reveal-in-Finder
// invocation: the system open utility, the reveal switch and the path as
// two separate argument-vector elements.
func revealArgvDarwin(path string) (string, []string) {
	return "open", []string{"-R", path}
}

// revealArgvWindows builds the argument vector for Windows Explorer's
// select-this-file invocation: the file explorer binary with a single
// argument formed by joining the select switch and the path with no space.
// This is community precedent (VS Code and other Go CLI tools use the same
// shape) and 23-RESEARCH.md's own explicitly flagged, unverified assumption
// (A1) -- it is unit-tested for structure here but still needs a real
// Windows pass before it can be called verified at runtime. Even if this
// exact shape is wrong, it can at worst open the wrong Explorer window: no
// shell is ever involved, so a wrong shape cannot escalate into a second
// command.
func revealArgvWindows(path string) (string, []string) {
	return "explorer", []string{"/select," + path}
}

// revealArgvLinux builds the argument vector for a Linux desktop's default
// opener, applied to the target's parent directory. No universal
// select-this-file mechanism exists across Linux file managers (nautilus,
// dolphin and caja each need their own flag; xdg-open has none), so this
// opens the containing folder rather than attempting file-manager detection.
func revealArgvLinux(path string) (string, []string) {
	return "xdg-open", []string{filepath.Dir(path)}
}

// revealArgvFor selects the argument-vector builder for platform (as
// reported by runtime.GOOS), taking the platform as a parameter rather than
// a Go build tag. That is a deliberate deviation from 23-RESEARCH.md's
// three-build-tagged-files sketch: with platform as a parameter, all three
// shapes are exercised by the same test binary on any one development
// machine, rather than only the one shape the compiler would have selected
// for a build-tagged file. An unrecognised platform yields an empty command
// name and nil arguments -- the caller must treat that as an
// unsupported-platform error and must not spawn anything.
func revealArgvFor(platform, path string) (string, []string) {
	switch platform {
	case "darwin":
		return revealArgvDarwin(path)
	case "windows":
		return revealArgvWindows(path)
	case "linux":
		return revealArgvLinux(path)
	default:
		return "", nil
	}
}

// allowedRevealExtensions are the only file types RevealInFileManager will
// ever spawn a process for: a catalog's own JSON file, or its generated
// HTML companion. Anything else -- a directory, an executable, an
// arbitrary path -- is rejected before any argument vector is built.
var allowedRevealExtensions = map[string]bool{
	".json": true,
	".html": true,
}

// containsPath reports whether resolved -- an already absolute,
// symlink-resolved path -- lives inside catalogDir. catalogDir is resolved
// to its own absolute, symlink-resolved form here so both sides of the
// comparison are in the same normalized form before filepath.Rel runs.
//
// filepath.Rel, not strings.HasPrefix, is what makes this separator-aware:
// a naive prefix check on "/catalogs-evil".HasPrefix("/catalogs") is true,
// which would wrongly admit a sibling directory that merely shares
// catalogDir's name as a string prefix. Rel instead returns a path that
// starts with ".." whenever resolved falls outside catalogDir, which is
// exactly what both a sibling directory and a "../" escape produce.
//
// Extracted as its own pure function -- like revealArgvFor above -- so
// containment is exercised directly by table-driven tests without ever
// reaching exec.Command.
func containsPath(catalogDir, resolved string) (bool, error) {
	absCatalogDir, err := filepath.Abs(catalogDir)
	if err != nil {
		return false, err
	}
	resolvedCatalogDir, err := filepath.EvalSymlinks(absCatalogDir)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(resolvedCatalogDir, resolved)
	if err != nil {
		return false, err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}

// RevealInFileManager asks the operating system's file manager to reveal
// path, selected within its containing folder. path is resolved to an
// absolute, symlink-resolved form and validated -- must exist, must be a
// regular file, must carry an allowed extension, and must resolve inside
// catalogDir (the caller's configured catalog directory) -- before anything
// else happens; only after every check passes is an argument vector built
// and run.
//
// catalogDir closes the gap the phase's plan (T-23-02) originally accepted:
// this binding is reachable from any renderer JS, not only the one caller
// that currently supplies a safe path, so extension + regular-file alone
// let any `.json`/`.html` file on the OS user's own filesystem be revealed.
// Requiring containment restricts the blast radius back to the catalog
// directory the frontend actually configured.
//
// The command is always built as a name plus a distinct argument slice via
// exec.Command, which never invokes a shell of any kind. path is always its
// own argument-vector element and is never concatenated into a command
// line or handed to an interpreter -- no character in path, however
// hostile, can therefore be interpreted as a second command. This is the
// entire mitigation for T-23-01, the phase's highest-severity threat.
func RevealInFileManager(path string, catalogDir string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("reveal %s: resolve path: %w", path, err)
	}

	// Resolve symlinks before any other check so a link cannot point the
	// reveal at something the extension/regular-file checks below would
	// otherwise be evaluating against the wrong target.
	//
	// ponytail: EvalSymlinks-then-Stat-then-exec leaves a TOCTOU window --
	// nothing pins the resolved file between this check and the
	// exec.Command call below, so a write-access attacker could swap it in
	// between. Accepted: exploiting it already requires local write access
	// to the exact resolved path, which grants far stronger primitives than
	// "reveal the wrong file." Upgrade path if this file is ever revisited:
	// open the resolved path with O_NOFOLLOW and pass the fd, not the path.
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return fmt.Errorf("reveal %s: %w", path, err)
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return fmt.Errorf("reveal %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("reveal %s: is a directory, not a catalog or HTML file", path)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("reveal %s: not a regular file", path)
	}
	ext := strings.ToLower(filepath.Ext(resolved))
	if !allowedRevealExtensions[ext] {
		return fmt.Errorf("reveal %s: unsupported extension %q", path, ext)
	}

	if catalogDir == "" {
		return fmt.Errorf("reveal %s: no catalog directory configured", path)
	}
	ok, err := containsPath(catalogDir, resolved)
	if err != nil {
		return fmt.Errorf("reveal %s: resolve catalog directory: %w", path, err)
	}
	if !ok {
		return fmt.Errorf("reveal %s: outside configured catalog directory", path)
	}

	name, args := revealArgvFor(runtime.GOOS, resolved)
	if name == "" {
		return fmt.Errorf("reveal %s: unsupported platform %q", path, runtime.GOOS)
	}

	if err := exec.Command(name, args...).Run(); err != nil {
		return fmt.Errorf("reveal %s: %w", path, err)
	}
	return nil
}
