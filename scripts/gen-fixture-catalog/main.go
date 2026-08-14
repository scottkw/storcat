// Command gen-fixture-catalog writes a synthetic catalog JSON file to a
// chosen directory, for exercising the tree pane and LoadCatalogFlat against
// realistic node counts without committing a generated blob to the repo.
package main

import (
	"flag"
	"fmt"
	"os"

	"storcat-wails/internal/fixture"
)

func main() {
	out := flag.String("out", os.TempDir(), "target directory for the generated catalog")
	shape := flag.String("shape", "dcim", "fixture shape: dcim, flat, or deep")
	dirs := flag.Int("dirs", 50, "top-level directories (dcim shape)")
	subdirs := flag.Int("subdirs", 50, "subdirectories per top-level directory (dcim shape)")
	files := flag.Int("files", 16, "files per subdirectory (dcim shape) or total files (flat shape)")
	depth := flag.Int("depth", 20, "nesting depth (deep shape)")
	flag.Parse()

	if err := os.MkdirAll(*out, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "gen-fixture-catalog: create output directory: %v\n", err)
		os.Exit(1)
	}

	var (
		path      string
		nodeCount int
		sizeBytes int64
		err       error
	)

	switch *shape {
	case "dcim":
		path, nodeCount, sizeBytes, err = fixture.WriteDCIMCatalog(*out, *dirs, *subdirs, *files)
	case "flat":
		path, nodeCount, sizeBytes, err = fixture.WriteFlatCatalog(*out, *files)
	case "deep":
		path, nodeCount, sizeBytes, err = fixture.WriteDeepCatalog(*out, *depth)
	default:
		fmt.Fprintf(os.Stderr, "gen-fixture-catalog: unknown shape %q (want dcim, flat, or deep)\n", *shape)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "gen-fixture-catalog: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("path=%s nodes=%d bytes=%d\n", path, nodeCount, sizeBytes)
}
