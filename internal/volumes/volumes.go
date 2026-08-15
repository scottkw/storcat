// Package volumes enumerates the machine's mounted volumes/drives, per
// operating system, using only the Go standard library. Unlike
// internal/osutil/reveal.go -- which deliberately dispatches per-OS argv
// shapes via a runtime.GOOS parameter within one file, so a single test
// binary exercises all three shapes -- this package uses per-OS compiler
// build constraints (darwin / linux / windows) for volumes_darwin.go,
// volumes_linux.go and volumes_windows.go. That is a hard compiler requirement here, not a
// style choice: the Windows free-space call and drive-letter enumeration
// this package needs are only reachable through Windows-only stdlib
// syscall members (syscall.NewLazyDLL/LazyProc, unavailable to build), and
// the stdlib syscall.Statfs_t struct this package reads has genuinely
// different field types per operating system (darwin's Bsize is uint32,
// linux amd64's Bsize is int64) -- a single non-build-tagged file simply
// cannot compile against both.
//
// No new Go module is added for this: every per-OS primitive used here
// (syscall.Statfs on darwin/linux, syscall.NewLazyDLL/LazyProc.Call to
// reach kernel32.dll's GetDiskFreeSpaceExW/GetLogicalDrives on Windows) is
// already in the standard library.
package volumes

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Volume is one enumerated mount point: a display name, its mount path,
// total/free bytes and whether its root can currently be read. Field
// names carry json tags because this struct crosses the Wails bridge to
// the frontend unchanged (App.ListVolumes).
type Volume struct {
	Name       string `json:"name"`
	MountPath  string `json:"mountPath"`
	TotalBytes int64  `json:"totalBytes"`
	FreeBytes  int64  `json:"freeBytes"`
	Readable   bool   `json:"readable"`
}

// List enumerates candidate mount points for the current operating system
// (mountPoints, implemented per-OS below) and turns them into Volumes.
// It always returns a non-nil, possibly zero-length slice and a nil error
// when the platform-specific candidate lookup itself succeeds -- zero
// detected volumes is a real, supported outcome (CRT-02 empty), not a
// failure.
func List() ([]Volume, error) {
	candidates, err := mountPoints()
	if err != nil {
		return nil, err
	}
	return buildVolumeList(candidates), nil
}

// buildVolumeList is List's portable core, factored out so it can be unit
// tested directly against a fixed candidate slice without depending on
// this machine's actual mount namespace. A candidate that fails its
// readability probe or its size probe never drops the whole enumeration --
// it contributes a not-readable, zero-size entry instead, per List's
// contract above. Pre-walking every volume to populate a picker would be
// far too slow, so per candidate this does at most one directory-entry
// read (probeReadable) and one stat-style call (diskUsage), never a walk.
func buildVolumeList(candidates []string) []Volume {
	result := make([]Volume, 0, len(candidates))
	for _, path := range candidates {
		name := volumeNameFromMountPath(path)
		if skipMountEntry(name) {
			continue
		}
		total, free := diskUsage(path)
		result = append(result, Volume{
			Name:       name,
			MountPath:  path,
			TotalBytes: total,
			FreeBytes:  free,
			Readable:   probeReadable(path),
		})
	}
	return result
}

// volumeNameFromMountPath derives a display name from path's final
// element, preserved byte-for-byte -- spaces, non-ASCII characters and
// emoji all pass through unmodified and unmangled (CRT-02 encoding), and
// the scanner later receives this same mount path unmodified too.
func volumeNameFromMountPath(path string) string {
	trimmed := strings.TrimRight(path, string(filepath.Separator))
	return filepath.Base(trimmed)
}

// skipMountEntry reports whether name should be excluded from the volume
// list: a hidden dot-prefixed entry (matches this codebase's existing
// hidden-file-skip convention, see internal/catalog/service.go) or a
// vendor-reserved entry.
//
// ponytail/ASSUMED: the vendor-reserved prefix below is grounded in one
// live observation on this machine (macOS 26.6.1 / Darwin 25.6.0:
// "com.apple.TimeMachine.localsnapshots" under /Volumes,
// 25-RESEARCH.md Pitfall 4 / Assumption A1) plus general web search, not a
// documented Apple API. Low severity if it ever over- or under-matches on
// a different macOS version -- cosmetic, an extra or missing bogus card --
// and cheap to revisit; harmless (a no-op) on Linux/Windows, which never
// produce "com.apple."-prefixed mount names.
func skipMountEntry(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	if strings.HasPrefix(name, "com.apple.") {
		return true
	}
	return false
}

// probeReadable reports whether path's root can currently be read, by
// opening it and reading at most one directory entry -- never a full
// directory listing, since pre-walking a large volume's root for a
// yes-or-no answer would be wasted work. A path that cannot even be
// opened (e.g. permission denied) is reported not-readable, not an error;
// the caller (buildVolumeList) still lists it, tagged accordingly, rather
// than dropping it from the result (CRT-02's "read errors" status tag).
func probeReadable(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	if _, err := f.ReadDir(1); err != nil && err != io.EOF {
		return false
	}
	return true
}
