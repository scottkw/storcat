//go:build darwin

package volumes

import (
	"os"
	"path/filepath"
	"syscall"
)

// volumesRoot is macOS's general-purpose mount namespace. It is not an
// "external drives only" list -- Apple documents no formal API
// distinction between a real external volume and internal artifacts like
// Time Machine snapshot directories, which is exactly why List's shared
// skipMountEntry filter and this file's own boot-volume exclusion exist
// (25-RESEARCH.md Pitfall 4, live-verified against this machine's own
// `ls -la /Volumes`).
const volumesRoot = "/Volumes"

// mountPoints lists candidate mount points under /Volumes, excluding any
// entry whose resolved symlink target is the filesystem root ("/") --
// macOS represents the boot volume as exactly such a symlink (observed on
// this machine as "Macintosh HD" -> "/"), and a card carrying the whole
// boot disk's size would be wrong and enormous.
func mountPoints() ([]string, error) {
	entries, err := os.ReadDir(volumesRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	points := make([]string, 0, len(entries))
	for _, entry := range entries {
		full := filepath.Join(volumesRoot, entry.Name())
		resolved, err := filepath.EvalSymlinks(full)
		if err != nil {
			// A dangling symlink, or an entry that vanished between
			// ReadDir and here -- skip this one candidate rather than
			// failing the whole enumeration.
			continue
		}
		if resolved == string(filepath.Separator) {
			continue // the boot volume
		}
		points = append(points, full)
	}
	return points, nil
}

// diskUsage reports total and available-free bytes for path via the
// standard library's syscall.Statfs -- no golang.org/x/sys import needed.
// Bsize is a uint32 field on darwin (differs from linux amd64's int64,
// which is exactly why this file needs its own go:build tag rather than
// sharing a portable implementation with volumes_linux.go); both products
// are computed in uint64 before the final int64 conversion to avoid an
// intermediate overflow on the multiply. Any Statfs failure (e.g. a
// volume that vanished between mountPoints() and here) reports zero/zero
// rather than propagating an error -- List's contract is that one bad
// candidate never fails the whole enumeration.
func diskUsage(path string) (total, free int64) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0
	}
	bsize := uint64(st.Bsize)
	return int64(bsize * st.Blocks), int64(bsize * st.Bavail)
}
