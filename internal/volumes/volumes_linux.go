//go:build linux

package volumes

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// candidateRoots are the two conventional removable-media roots this
// package scans: /mnt at its own level, and /media both at its own level
// and (the common desktop-automount convention -- udisks2/GVfs mount
// removable media under a per-user directory) one level down under each
// of its subdirectories. There is no equivalent per-user convention under
// /mnt, so it is scanned at only the one level.
func candidateRoots() []string {
	roots := []string{"/mnt"}
	mediaEntries, err := os.ReadDir("/media")
	if err != nil {
		return roots
	}
	roots = append(roots, "/media")
	for _, e := range mediaEntries {
		if e.IsDir() {
			roots = append(roots, filepath.Join("/media", e.Name()))
		}
	}
	return roots
}

// mountedPaths reads the kernel's own mount table so a stale empty
// directory under one of candidateRoots (never actually mounted, just
// left behind) is not offered as a volume.
func mountedPaths() (map[string]bool, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mounted := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		mounted[fields[1]] = true
	}
	return mounted, scanner.Err()
}

// mountPoints lists directory entries under candidateRoots that are also
// present as a mount point in /proc/mounts. If the mount table itself
// cannot be read, this returns a non-nil empty slice rather than an
// error -- consistent with List's contract that a platform-specific
// hiccup never fails the whole enumeration outright.
func mountPoints() ([]string, error) {
	mounted, err := mountedPaths()
	if err != nil {
		return []string{}, nil
	}

	points := make([]string, 0)
	for _, root := range candidateRoots() {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			full := filepath.Join(root, e.Name())
			if mounted[full] {
				points = append(points, full)
			}
		}
	}
	return points, nil
}

// diskUsage reports total and available-free bytes for path via the
// standard library's syscall.Statfs. Bsize is an int64 field on linux
// amd64/arm64 (differs from darwin's uint32, which is why this file needs
// its own go:build tag). Any Statfs failure reports zero/zero rather than
// propagating an error, matching volumes_darwin.go's tolerance.
func diskUsage(path string) (total, free int64) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0
	}
	bsize := uint64(st.Bsize)
	return int64(bsize * st.Blocks), int64(bsize * st.Bavail)
}
