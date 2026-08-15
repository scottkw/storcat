//go:build windows

package volumes

import (
	"fmt"
	"syscall"
	"unsafe"
)

// kernel32 and its two procedures are loaded via the standard library's
// own syscall.NewLazyDLL/LazyProc -- not golang.org/x/sys/windows, per
// this phase's stdlib-only constraint. LazyDLL's doc comment flags a
// DLL-preloading concern for arbitrary DLL names; it does not apply to
// "kernel32.dll" specifically, since it is one of Windows' "known DLLs"
// and is always resolved from the protected System32 directory
// regardless of process search-path order.
var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetLogicalDrives    = kernel32.NewProc("GetLogicalDrives")
	procGetDiskFreeSpaceExW = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// mountPoints enumerates drive letters from the bitmask GetLogicalDrives
// returns -- bit 0 is A:, bit 1 is B:, and so on through bit 25 for Z:.
func mountPoints() ([]string, error) {
	r1, _, callErr := procGetLogicalDrives.Call()
	if r1 == 0 {
		return nil, fmt.Errorf("GetLogicalDrives: %w", callErr)
	}
	mask := uint32(r1)

	points := make([]string, 0, 26)
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) != 0 {
			points = append(points, fmt.Sprintf("%c:\\", 'A'+i))
		}
	}
	return points, nil
}

// diskUsage reports total and caller-available-free bytes for path via
// the Win32 GetDiskFreeSpaceExW call. Any failure (e.g. a drive that was
// removed between mountPoints() and here) reports zero/zero rather than
// propagating an error, matching the darwin/linux implementations'
// tolerance -- one bad candidate never fails the whole enumeration.
//
// COMPILE-VERIFIED ONLY: no Windows machine was available this session to
// run this at all -- see .planning/WINDOWS.md for the tracked runtime gap.
// (Compiled successfully under GOOS=windows; never executed.)
func diskUsage(path string) (total, free int64) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0
	}

	var freeAvailToCaller, totalBytes, totalFreeBytes uint64
	r1, _, _ := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeAvailToCaller)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if r1 == 0 {
		return 0, 0
	}
	return int64(totalBytes), int64(freeAvailToCaller)
}
