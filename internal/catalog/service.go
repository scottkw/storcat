package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"storcat-wails/pkg/models"
)

// ProgressUpdate is reported to a ProgressCallback as the walk proceeds.
// FilesSeen and BytesSeen are running totals for the whole scan so far (not
// deltas since the previous update); ReadErrors counts skipped single-entry
// read failures so far. This type lives under internal/, so no consumer
// outside this module can import it -- its one existing caller (the CLI, via
// the CreateCatalog wrapper) passes nil.
type ProgressUpdate struct {
	Path       string `json:"path"`
	FilesSeen  int    `json:"filesSeen"`
	BytesSeen  int64  `json:"bytesSeen"`
	ReadErrors int    `json:"readErrors"`
}

// ProgressCallback is called during directory traversal with the current
// progress snapshot.
type ProgressCallback func(ProgressUpdate)

// Service handles catalog creation and management
type Service struct{}

// NewService creates a new catalog service
func NewService() *Service {
	return &Service{}
}

// maxReadErrorEntries caps the number of recorded read-error entries kept
// on a walkState, so a device failing on every entry cannot grow an
// unbounded slice. The readErrors counter itself stays uncapped.
const maxReadErrorEntries = 50

// walkState carries per-scan counters and options through the recursive
// traverseDirectory walk. Counter mutation happens at the call sites in
// traverseDirectory, never inside report, so a file's size is added exactly
// once regardless of how many times report is called.
type walkState struct {
	scanRoot   string
	opts       Options
	onProgress ProgressCallback

	filesSeen  int
	bytesSeen  int64
	readErrors int

	readErrorEntries []ReadErrorEntry
	terminal         bool
}

// report forwards the current counters to onProgress for displayPath, when a
// callback is configured. It never mutates a counter itself.
func (st *walkState) report(displayPath string) {
	if st.onProgress == nil {
		return
	}
	st.onProgress(ProgressUpdate{
		Path:       displayPath,
		FilesSeen:  st.filesSeen,
		BytesSeen:  st.bytesSeen,
		ReadErrors: st.readErrors,
	})
}

// recordReadError increments the uncapped read-error counter and appends a
// bounded entry (the most recent maxReadErrorEntries only).
func (st *walkState) recordReadError(path string, err error) {
	st.readErrors++
	st.readErrorEntries = append(st.readErrorEntries, ReadErrorEntry{Path: path, Reason: err.Error()})
	if len(st.readErrorEntries) > maxReadErrorEntries {
		st.readErrorEntries = st.readErrorEntries[len(st.readErrorEntries)-maxReadErrorEntries:]
	}
}

// classify re-probes the SCAN ROOT -- not the failing subdirectory -- with a
// cheap stat, and sets/returns st.terminal when the root itself no longer
// stats. When HaltOnSourceLoss is false, classification never runs and
// every read failure keeps today's skip-and-continue behavior unchanged.
// This root re-probe is the whole classification: a failure that leaves the
// root reachable is one bad entry; a failure that takes the root with it is
// a vanished source.
func (st *walkState) classify() bool {
	if !st.opts.HaltOnSourceLoss {
		return false
	}
	if _, err := os.Stat(st.scanRoot); err != nil {
		st.terminal = true
	}
	return st.terminal
}

// CreateCatalog is the CLI's compatibility-preserving thin wrapper around
// CreateCatalogWithContext: context.Background() (no cancellation),
// sourcePath and outputDir both set to directoryPath (the CLI has always
// written its output into the directory it scanned), and
// Options{WriteHTML: true} -- NOT the zero value. The CLI has always written
// both JSON and HTML unconditionally; constructing Options{} here would
// silently drop HTML output from every `storcat create` run with no
// compile-time signal (see TestCreateCatalog_WrapperWritesHTML). cli/create.go
// calls this method unedited and must never need to change.
func (s *Service) CreateCatalog(title, directoryPath, outputRoot string, copyToDirectory string, onProgress ProgressCallback) (*models.CreateCatalogResult, error) {
	return s.CreateCatalogWithContext(
		context.Background(),
		title,
		directoryPath, // sourcePath: walked
		directoryPath, // outputDir: written -- SAME as source, preserving today's exact behavior
		outputRoot,
		copyToDirectory,
		Options{WriteHTML: true}, // NOT the zero value -- see the doc comment above
		onProgress,
	)
}

// CreateCatalogWithContext walks sourcePath and writes the resulting catalog
// into outputDir under outputRoot. The entire tree is built in memory before
// any write is attempted -- that ordering, not a rollback mechanism, is what
// makes "a cancelled scan writes nothing" true. A future change must never
// replace this with an incremental-write design that writes partial results
// as the walk proceeds.
func (s *Service) CreateCatalogWithContext(ctx context.Context, title, sourcePath, outputDir, outputRoot, copyToDirectory string, opts Options, onProgress ProgressCallback) (*models.CreateCatalogResult, error) {
	st := &walkState{
		scanRoot:   sourcePath,
		opts:       opts,
		onProgress: onProgress,
	}

	tree, err := s.traverseDirectory(ctx, sourcePath, sourcePath, st)
	if err != nil {
		// Three outcomes, distinguished before touching the write path: a
		// cancelled context writes nothing; a source-loss error is returned
		// with its populated partial scan attached and writes nothing;
		// anything else is a genuine traversal failure.
		var srcErr *SourceUnavailableError
		if errors.As(err, &srcErr) {
			srcErr.Partial = &PartialScan{
				Tree:       tree,
				FilesSeen:  st.filesSeen,
				BytesSeen:  st.bytesSeen,
				ReadErrors: st.readErrorEntries,
			}
			return nil, srcErr
		}
		return nil, fmt.Errorf("failed to traverse directory: %w", err)
	}

	// Re-check cancellation after the walk completes and before any write is
	// attempted -- traverseDirectory can return a non-error partial tree in
	// some skip-and-continue paths, so this is the authoritative gate.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return s.WriteCatalogFrom(tree, title, outputDir, outputRoot, copyToDirectory, opts)
}

// WriteCatalogFrom writes an already-built tree to outputDir under
// outputRoot: the JSON file always, the HTML sibling only when
// opts.WriteHTML, then a secondary copy of whichever files were written when
// copyToDirectory is non-empty. This is the single write path
// CreateCatalogWithContext uses, and the one later plans' partial-catalog
// and retry flows reuse.
func (s *Service) WriteCatalogFrom(tree *models.CatalogItem, title, outputDir, outputRoot, copyToDirectory string, opts Options) (*models.CreateCatalogResult, error) {
	jsonPath := filepath.Join(outputDir, outputRoot+".json")
	if err := s.writeJSONFile(tree, jsonPath); err != nil {
		return nil, fmt.Errorf("failed to write JSON: %w", err)
	}

	result := &models.CreateCatalogResult{
		JsonPath:  jsonPath,
		FileCount: s.countFiles(tree),
		TotalSize: tree.Size,
	}

	if opts.WriteHTML {
		htmlPath := filepath.Join(outputDir, outputRoot+".html")
		if err := s.writeHTMLFile(tree, title, htmlPath); err != nil {
			return nil, fmt.Errorf("failed to write HTML: %w", err)
		}
		result.HtmlPath = htmlPath
	}

	if copyToDirectory != "" {
		copyJSONPath := filepath.Join(copyToDirectory, outputRoot+".json")
		if err := s.copyFile(jsonPath, copyJSONPath); err != nil {
			return nil, fmt.Errorf("failed to copy JSON: %w", err)
		}
		result.CopyJsonPath = copyJSONPath

		if opts.WriteHTML {
			copyHTMLPath := filepath.Join(copyToDirectory, outputRoot+".html")
			if err := s.copyFile(result.HtmlPath, copyHTMLPath); err != nil {
				return nil, fmt.Errorf("failed to copy HTML: %w", err)
			}
			result.CopyHtmlPath = copyHTMLPath
		}
	}

	return result, nil
}

// traverseDirectory recursively builds catalog structure. ctx.Err() is
// checked at the very top, before os.Stat, so cancellation is prompt for
// every syscall that hasn't started yet (an already-in-flight blocked
// syscall is a documented, accepted Go runtime limitation -- see
// 25-RESEARCH.md Pitfall 3).
func (s *Service) traverseDirectory(ctx context.Context, dirPath, basePath string, st *walkState) (*models.CatalogItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("scan cancelled: %w", err)
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}

	// Calculate relative path
	relPath, err := filepath.Rel(filepath.Dir(basePath), dirPath)
	if err != nil {
		return nil, err
	}

	// Convert to Unix-style path with ./ prefix
	displayPath := "./" + filepath.ToSlash(relPath)
	if relPath == filepath.Base(basePath) {
		displayPath = "./"
	}

	// Handle files
	if info.Mode().IsRegular() {
		st.filesSeen++
		st.bytesSeen += info.Size()
		st.report(displayPath)
		return &models.CatalogItem{
			Type: "file",
			Name: displayPath,
			Size: info.Size(),
		}, nil
	}

	// Handle directories
	if info.IsDir() {
		st.report(displayPath)

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			st.recordReadError(dirPath, err)
			node := &models.CatalogItem{
				Type:     "directory",
				Name:     displayPath,
				Size:     0,
				Contents: []*models.CatalogItem{},
			}
			if st.classify() {
				// The scan root itself is gone: this is the origin node --
				// mark it and propagate a source-loss error carrying what
				// was built so far (nothing, at this node).
				node.Unreadable = true
				node.ReadError = err.Error()
				return node, &SourceUnavailableError{SourcePath: st.scanRoot}
			}
			// Root still reachable: today's skip-and-continue behavior,
			// unchanged -- an empty, error-free directory node.
			return node, nil
		}

		// Sort entries: directories first, then files, alphabetically
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].IsDir() && !entries[j].IsDir() {
				return true
			}
			if !entries[i].IsDir() && entries[j].IsDir() {
				return false
			}
			return entries[i].Name() < entries[j].Name()
		})

		var contents []*models.CatalogItem
		var totalSize int64

		for _, entry := range entries {
			// Skip hidden files (starting with .) unless IncludeHidden is set
			if !st.opts.IncludeHidden && strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			childPath := filepath.Join(dirPath, entry.Name())
			childItem, err := s.traverseDirectory(ctx, childPath, basePath, st)
			if err != nil {
				if ctx.Err() != nil {
					// Cancellation must propagate, not be swallowed as a
					// single-entry read error -- otherwise the walk would
					// keep going and CancelWritesNothing would not hold.
					return nil, err
				}

				var srcErr *SourceUnavailableError
				if errors.As(err, &srcErr) {
					// A deeper level already detected and marked the
					// origin node. Keep whatever partial child came back,
					// stop iterating this directory's remaining siblings,
					// and propagate the same error upward unmarked -- only
					// the origin node carries the marker fields.
					if childItem != nil {
						contents = append(contents, childItem)
						totalSize += childItem.Size
					}
					if contents == nil {
						contents = []*models.CatalogItem{}
					}
					return &models.CatalogItem{
						Type:     "directory",
						Name:     displayPath,
						Size:     totalSize,
						Contents: contents,
					}, err
				}

				// Plain single-entry failure (e.g. os.Stat failed for this
				// child). Record it and classify against the scan root.
				st.recordReadError(childPath, err)
				if st.classify() {
					if contents == nil {
						contents = []*models.CatalogItem{}
					}
					node := &models.CatalogItem{
						Type:     "directory",
						Name:     displayPath,
						Size:     totalSize,
						Contents: contents,
					}
					node.Unreadable = true
					node.ReadError = err.Error()
					return node, &SourceUnavailableError{SourcePath: st.scanRoot}
				}
				// Skip items we can't access -- unchanged byte-for-byte
				// behavior when the root is still reachable.
				continue
			}

			contents = append(contents, childItem)
			totalSize += childItem.Size
		}

		if contents == nil {
			contents = []*models.CatalogItem{}
		}

		return &models.CatalogItem{
			Type:     "directory",
			Name:     displayPath,
			Size:     totalSize,
			Contents: contents,
		}, nil
	}

	return nil, fmt.Errorf("unsupported file type: %s", dirPath)
}

// writeJSONFile writes catalog in bare object format (no indentation),
// routed through WriteFileAtomic so a crash mid-write can never leave a
// truncated .json at path.
func (s *Service) writeJSONFile(catalog *models.CatalogItem, path string) error {
	jsonBytes, err := json.Marshal(catalog)
	if err != nil {
		return err
	}

	return WriteFileAtomic(path, jsonBytes, 0644)
}

// writeHTMLFile generates the HTML catalog with exact tree formatting
func (s *Service) writeHTMLFile(catalog *models.CatalogItem, title, path string) error {
	treeStructure := s.generateTreeStructure(catalog, true, "")

	fileCount := s.countFiles(catalog)
	dirCount := s.countDirectories(catalog)
	totalSize := s.formatBytes(catalog.Size)

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
 <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
 <meta name="Author" content="Made by 'tree'">
 <meta name="GENERATOR" content="$Version: $ tree v1.7.0 (c) 1996 - 2014 by Steve Baker, Thomas Moore, Francesc Rocher, Florian Sesser, Kyosuke Tokoro $">
 <title>%s</title>
 <style type="text/css">
  <!--
  BODY { font-family : ariel, monospace, sans-serif; }
  P { font-weight: normal; font-family : ariel, monospace, sans-serif; color: black; background-color: transparent;}
  B { font-weight: normal; color: black; background-color: transparent;}
  A:visited { font-weight : normal; text-decoration : none; background-color : transparent; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  A:link    { font-weight : normal; text-decoration : none; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  A:hover   { color : #000000; font-weight : normal; text-decoration : underline; background-color : yellow; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  A:active  { color : #000000; font-weight: normal; background-color : transparent; margin : 0px 0px 0px 0px; padding : 0px 0px 0px 0px; display: inline; }
  .VERSION { font-size: small; font-family : arial, sans-serif; }
  .NORM  { color: black;  background-color: transparent;}
  .FIFO  { color: purple; background-color: transparent;}
  .CHAR  { color: yellow; background-color: transparent;}
  .DIR   { color: blue;   background-color: transparent;}
  .BLOCK { color: yellow; background-color: transparent;}
  .LINK  { color: aqua;   background-color: transparent;}
  .SOCK  { color: fuchsia;background-color: transparent;}
  .EXEC  { color: green;  background-color: transparent;}
  -->
 </style>
</head>
<body>
	<h1>%s</h1><p>
	%s
	<br><br>
	</p>
	<p>

 %s used in %d directories, %d files
	<br><br>
	</p>
	<hr>
	<p class="VERSION">
		 tree v1.7.0 © 1996 - 2014 by Steve Baker and Thomas Moore <br>
		 HTML output hacked and copyleft © 1998 by Francesc Rocher <br>
		 JSON output hacked and copyleft © 2014 by Florian Sesser <br>
		 Charsets / OS/2 support © 2001 by Kyosuke Tokoro
	</p>
</body>
</html>`, html.EscapeString(title), html.EscapeString(title), treeStructure, totalSize, dirCount, fileCount)

	return WriteFileAtomic(path, []byte(htmlContent), 0644)
}

// generateTreeStructure creates the tree visual representation
func (s *Service) generateTreeStructure(item *models.CatalogItem, isLast bool, prefix string) string {
	var result strings.Builder

	connector := "├── "
	if isLast {
		connector = "└── "
	}

	sizeDisplay := s.formatBytesForDisplay(item.Size)
	itemName := filepath.Base(item.Name)

	result.WriteString(fmt.Sprintf("%s%s%s&nbsp;&nbsp;%s<br>\n",
		prefix, connector, sizeDisplay, html.EscapeString(itemName)))

	// Process children for directories
	if item.Type == "directory" && item.Contents != nil && len(item.Contents) > 0 {
		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		for i, child := range item.Contents {
			childIsLast := i == len(item.Contents)-1
			result.WriteString(s.generateTreeStructure(child, childIsLast, newPrefix))
		}
	}

	return result.String()
}

// formatBytesForDisplay formats bytes with bracket and padding (e.g., "[271M]", "[   0]")
func (s *Service) formatBytesForDisplay(bytes int64) string {
	if bytes == 0 {
		return "[   0]"
	}
	formatted := s.formatBytes(bytes)
	return fmt.Sprintf("[%4s]", formatted)
}

// formatBytes converts bytes to human-readable format (e.g., "271M", "3.4M")
func (s *Service) formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}

	const unit = 1024
	sizes := []string{"B", "K", "M", "G", "T"}

	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit && exp < len(sizes)-1; n /= unit {
		div *= unit
		exp++
	}

	value := float64(bytes) / float64(div)

	// Format to 1 decimal place, but strip .0
	formatted := fmt.Sprintf("%.1f", value)
	if strings.HasSuffix(formatted, ".0") {
		formatted = strings.TrimSuffix(formatted, ".0")
	}

	return formatted + sizes[exp+1]
}

// countFiles counts total files in catalog
func (s *Service) countFiles(catalog *models.CatalogItem) int {
	if catalog.Type == "file" {
		return 1
	}

	count := 0
	if catalog.Type == "directory" && catalog.Contents != nil {
		for _, child := range catalog.Contents {
			count += s.countFiles(child)
		}
	}

	return count
}

// countDirectories counts total directories in catalog
func (s *Service) countDirectories(catalog *models.CatalogItem) int {
	if catalog.Type == "file" {
		return 0
	}

	count := 1 // Count this directory
	if catalog.Type == "directory" && catalog.Contents != nil {
		for _, child := range catalog.Contents {
			count += s.countDirectories(child)
		}
	}

	return count
}

// copyFile copies a file from src to dst
func (s *Service) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
