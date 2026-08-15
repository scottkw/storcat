package catalog

import (
	"fmt"

	"storcat-wails/pkg/models"
)

// ReadErrorEntry is one recorded read failure encountered during a walk.
type ReadErrorEntry struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// PartialScan carries everything walked before a scan-root loss was
// detected, so the caller can render where the scan stopped and, if the
// user chooses, write it through the normal write path without recounting.
type PartialScan struct {
	Tree       *models.CatalogItem
	FilesSeen  int
	BytesSeen  int64
	ReadErrors []ReadErrorEntry
}

// SourceUnavailableError signals that the scan root itself became
// unreachable mid-walk (CRT-10) -- distinct from a single bad file, which
// keeps today's skip-and-continue behavior with no error at all, and
// distinct from user cancellation, which the standard library's own
// context.Canceled already identifies. No separate cancellation sentinel is
// minted here: callers distinguish the two cases with errors.Is against
// context.Canceled and errors.As against this type, which is enough
// surface without a second synonym for "the walk stopped early."
type SourceUnavailableError struct {
	SourcePath string
	Partial    *PartialScan
}

func (e *SourceUnavailableError) Error() string {
	return fmt.Sprintf("source unavailable: %s", e.SourcePath)
}
