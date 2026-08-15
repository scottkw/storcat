//go:build darwin

package volumes

import "testing"

// TestList_ExcludesBootVolumeSymlink runs List() against this machine's
// real, live /Volumes mount namespace (25-RESEARCH.md Pitfall 4,
// live-verified this session via `ls -la /Volumes`, which showed
// "Macintosh HD" as a symlink resolving to "/") and asserts the boot
// volume never survives into the result -- a card carrying the whole boot
// disk's size would otherwise appear.
func TestList_ExcludesBootVolumeSymlink(t *testing.T) {
	got, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	for _, v := range got {
		if v.MountPath == "/" {
			t.Errorf("List() included an entry whose mount path is the filesystem root: %+v", v)
		}
		if v.Name == "Macintosh HD" {
			t.Errorf("List() included the boot volume's symlink entry by name: %+v", v)
		}
	}
}
