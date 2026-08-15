package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"storcat-wails/internal/catalog"
	"storcat-wails/internal/config"
	"storcat-wails/internal/osutil"
	"storcat-wails/internal/search"
	"storcat-wails/internal/volumes"
	"storcat-wails/pkg/models"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	catalogService *catalog.Service
	searchService  *search.Service
	configManager  *config.Manager

	// scanMu guards activeScanCancel, scanDone, and the retained-partial
	// fields below -- the product is one-scan-at-a-time by design, so a
	// single mutex-guarded field is the simplest correct implementation (no
	// scan-id-keyed map needed).
	scanMu           sync.Mutex
	activeScanCancel context.CancelFunc
	scanDone         chan struct{}

	// lastPartial, lastPartialResult and lastScanReq retain a source-loss
	// scan's walked tree and the parameters needed to write it, so
	// WritePartialCatalog can produce the catalog the user chooses to
	// salvage without re-walking anything. All three are guarded by scanMu,
	// same as the cancel handle above -- read and write both only under it.
	// A fresh StartScan clears all three before running, so a stale tree
	// from a previous failure can never be written under a new scan's name
	// (T-25-12).
	lastPartial       *catalog.PartialScan
	lastPartialResult *models.CreateCatalogResult
	lastScanReq       *lastScanRequest

	// retainedGen counts every time StartScan clears the retained-partial
	// fields above. WritePartialCatalog reads it before releasing scanMu for
	// the (slow, unguarded) filesystem write, then compares it again after
	// the write completes: a mismatch means a newer StartScan superseded the
	// tree this call was writing while the write was in flight, so the
	// result must not be cached/recorded over the newer retained state
	// (CR-02).
	retainedGen int

	// writeMu serializes WritePartialCatalog's whole check-decide-write-record
	// sequence across concurrent invocations of that one binding, so two
	// overlapping calls can never both perform the actual filesystem write --
	// the second blocks until the first releases, then re-checks
	// lastPartialResult (now set) and returns the cached result instead of
	// writing again. It is a separate lock from scanMu so StartScan/CancelScan
	// are never blocked behind an in-flight partial write.
	writeMu sync.Mutex
}

// lastScanRequest captures the parameters a source-loss scan was started
// with, so WritePartialCatalog can replay the shared write path
// (WriteCatalogFrom) without re-walking anything. sourcePath is retained
// alongside the write-path parameters for completeness even though only
// outputDir/outputRoot/copyToDirectory/opts are read by the write itself.
type lastScanRequest struct {
	title           string
	sourcePath      string
	outputDir       string
	outputRoot      string
	copyToDirectory string
	opts            catalog.Options
}

// ScanOptions is the StartScan binding's option parameter: the create-flow
// toggles (write HTML, include hidden files) plus the secondary-copy
// destination the frontend configures per scan.
type ScanOptions struct {
	WriteHTML       bool   `json:"writeHTML"`
	IncludeHidden   bool   `json:"includeHidden"`
	CopyToDirectory string `json:"copyToDirectory"`
	// TotalBytesHint is the total the selected volume card already
	// computed (from ListVolumes' TotalBytes), used as the scan's
	// denominator with no pre-pass. Zero means no such total exists (the
	// plain-folder "choose any folder" path) -- resolveScanTotal then
	// runs catalog.MeasureTree's count-only pre-pass instead. Zero is
	// therefore the wire signal for "denominator unknown," never a real
	// hint of an empty volume.
	TotalBytesHint int64 `json:"totalBytesHint"`
}

// ScanProgress is the scan-progress event payload shape. TotalBytes is
// echoed back from the caller (StartScan's own denominator) so the frontend
// never has to correlate two sources to compute a percentage.
type ScanProgress struct {
	Path       string `json:"path"`
	FilesSeen  int    `json:"filesSeen"`
	BytesSeen  int64  `json:"bytesSeen"`
	ReadErrors int    `json:"readErrors"`
	TotalBytes int64  `json:"totalBytes"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Initialize config manager
	configManager, err := config.NewManager()
	if err != nil {
		// If config fails, just create one with defaults
		configManager = &config.Manager{}
	}

	// Initialize services
	catalogService := catalog.NewService()
	searchService := search.NewService()

	// Wire the sidecar counts cache into the search service. Same
	// tolerance pattern as configManager above: a construction failure
	// must leave the app fully usable, with the search service simply
	// holding no cache (SetCountsCache is nil-safe) rather than aborting
	// startup.
	countsCache, err := config.NewCountsCache()
	if err != nil {
		countsCache = nil
	}
	searchService.SetCountsCache(countsCache)

	return &App{
		catalogService: catalogService,
		searchService:  searchService,
		configManager:  configManager,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// CreateCatalog creates a new catalog from a directory
func (a *App) CreateCatalog(title string, directoryPath string, outputName string, copyToDirectory string) (*models.CreateCatalogResult, error) {
	absPath, err := filepath.Abs(directoryPath)
	if err != nil {
		return nil, err
	}

	// Progress callback (could be used to send progress to frontend in future)
	progressCallback := func(update catalog.ProgressUpdate) {
		// For now, we don't send progress updates
		// In the future, we could use Wails events to send updates to frontend
	}

	result, err := a.catalogService.CreateCatalog(title, absPath, outputName, copyToDirectory, progressCallback)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// throttledProgress returns a catalog.ProgressCallback that forwards at
// most one event per 200ms as the scan-progress Wails event below, always
// carrying the latest counters rather than every intermediate update. This
// is the only place in the repository allowed to call runtime.EventsEmit
// for scan progress -- internal/catalog must stay usable from the CLI with
// no Wails runtime attached (COMPAT-04), so all throttling and emission
// live here. The a.ctx == nil guard makes the returned closure safe to call
// from a plain Go test with no Wails runtime attached.
func (a *App) throttledProgress(totalBytes int64) catalog.ProgressCallback {
	var lastEmit time.Time
	return func(u catalog.ProgressUpdate) {
		if a.ctx == nil {
			return
		}
		if time.Since(lastEmit) < 200*time.Millisecond {
			return
		}
		lastEmit = time.Now()
		runtime.EventsEmit(a.ctx, "scan:progress", ScanProgress{
			Path:       u.Path,
			FilesSeen:  u.FilesSeen,
			BytesSeen:  u.BytesSeen,
			ReadErrors: u.ReadErrors,
			TotalBytes: totalBytes,
		})
	}
}

// resolveScanTotal returns the scan's real denominator: hint unchanged
// when non-zero (no pre-pass runs at all), or -- when hint is zero, the
// "denominator unknown" wire signal -- the result of a count-only
// catalog.MeasureTree pre-pass over sourcePath, which is what fills that
// seam for the plain-folder case (CRT-03) where no volume total exists.
// A two-pass count for every scan was rejected (25-CONTEXT.md); running
// MeasureTree only on this branch is what keeps that cost paid at most
// once per scan, and never at all for a volume source.
//
// While the pre-pass runs, its own progress is forwarded through a
// zero-total throttled emitter -- the frontend renders that as a live
// file count, never a fabricated percentage, until this function returns
// the real total and the caller starts the walk proper with it.
func (a *App) resolveScanTotal(ctx context.Context, sourcePath string, opts catalog.Options, hint int64, testHook catalog.ProgressCallback) (int64, error) {
	if hint != 0 {
		return hint, nil
	}
	prePassProgress := a.throttledProgress(0)
	_, measuredBytes, err := catalog.MeasureTree(ctx, sourcePath, opts, func(u catalog.ProgressUpdate) {
		if testHook != nil {
			testHook(u)
		}
		prePassProgress(u)
	})
	if err != nil {
		return 0, err
	}
	return measuredBytes, nil
}

// StartScan runs a cancellable, option-driven scan: sourcePath is walked and
// the resulting catalog is written into outputDir under outputRoot (and,
// when opts.CopyToDirectory is set, copied there too), with progress
// throttled onto the scan-progress event below. Both the primary write
// destination and the secondary copy destination must resolve inside their
// respective directories per osutil.ContainsPath before anything is walked
// -- outputRoot and CopyToDirectory both arrive as free text from the
// renderer (T-25-01, T-25-02). Only one scan may run at a time (T-25-03);
// a concurrent call is rejected without disturbing the running scan's
// cancel handle.
func (a *App) StartScan(title, sourcePath, outputDir, outputRoot string, opts ScanOptions) (*models.CreateCatalogResult, error) {
	return a.startScan(title, sourcePath, outputDir, outputRoot, opts, nil)
}

// startScan is StartScan's real implementation, parameterized on an
// optional testHook invoked (before the throttled progress report) on
// every walk node. It exists so app_test.go can deterministically
// reproduce a mid-walk source loss -- removing the scan root when a known
// node is reported, exactly the technique internal/catalog/service_test.go
// already uses at the service layer -- without needing a live Wails
// runtime context, which a.throttledProgress's EventsEmit call requires
// and would otherwise log.Fatal on an invalid one. testHook is always nil
// from the real Wails-bound StartScan above.
func (a *App) startScan(title, sourcePath, outputDir, outputRoot string, opts ScanOptions, testHook catalog.ProgressCallback) (*models.CreateCatalogResult, error) {
	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, err
	}
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, err
	}
	if outputRoot == "" {
		return nil, fmt.Errorf("outputRoot must not be empty")
	}

	// Resolve the output directory's symlinks (e.g. macOS's /var ->
	// /private/var) before building the two destination paths, so both
	// sides of the containment check below are in the same normalized
	// form -- comparing a resolved base against an unresolved child would
	// otherwise report a false escape for every legitimately-nested path.
	resolvedOutput, err := filepath.EvalSymlinks(absOutput)
	if err != nil {
		return nil, err
	}

	jsonDest := filepath.Join(resolvedOutput, outputRoot+".json")
	if ok, err := osutil.ContainsPath(resolvedOutput, jsonDest); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("outputRoot %q escapes the output directory", outputRoot)
	}
	htmlDest := filepath.Join(resolvedOutput, outputRoot+".html")
	if ok, err := osutil.ContainsPath(resolvedOutput, htmlDest); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("outputRoot %q escapes the output directory", outputRoot)
	}
	absOutput = resolvedOutput

	copyToDirectory := opts.CopyToDirectory
	if copyToDirectory != "" {
		absCopy, err := filepath.Abs(copyToDirectory)
		if err != nil {
			return nil, err
		}
		resolvedCopy, err := filepath.EvalSymlinks(absCopy)
		if err != nil {
			return nil, err
		}
		copyJSONDest := filepath.Join(resolvedCopy, outputRoot+".json")
		if ok, err := osutil.ContainsPath(resolvedCopy, copyJSONDest); err != nil {
			return nil, err
		} else if !ok {
			return nil, fmt.Errorf("outputRoot %q escapes the copy-to directory", outputRoot)
		}
		copyToDirectory = resolvedCopy
	}

	a.scanMu.Lock()
	if a.activeScanCancel != nil {
		a.scanMu.Unlock()
		return nil, fmt.Errorf("a scan is already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.activeScanCancel = cancel
	a.scanDone = make(chan struct{})
	// Clear any previously retained partial scan -- a stale tree from an
	// earlier failure must never be written under this new scan's name
	// (T-25-12), asserted by TestStartScan_ClearsRetainedPartialOnNewScan.
	a.lastPartial = nil
	a.lastPartialResult = nil
	a.lastScanReq = nil
	// Bump the generation counter so an in-flight WritePartialCatalog call
	// (started against the previous retained tree, if any) can detect on
	// completion that this StartScan has since superseded it (CR-02).
	a.retainedGen++
	a.scanMu.Unlock()

	defer func() {
		a.scanMu.Lock()
		close(a.scanDone)
		a.activeScanCancel = nil
		a.scanMu.Unlock()
		cancel()
	}()

	// HaltOnSourceLoss is always true here: the GUI wants the
	// volume-vanished distinction (CRT-10/CRT-11) surfaced as a distinct,
	// salvageable error state. The CLI wrapper (CreateCatalog, in
	// internal/catalog/service.go) deliberately never sets this, preserving
	// v2.3.0's skip-and-continue behavior exactly.
	catOpts := catalog.Options{WriteHTML: opts.WriteHTML, IncludeHidden: opts.IncludeHidden, HaltOnSourceLoss: true}

	totalBytes, err := a.resolveScanTotal(ctx, absSource, catOpts, opts.TotalBytesHint, testHook)
	if err != nil {
		return nil, err
	}

	progress := a.throttledProgress(totalBytes)
	onProgress := func(u catalog.ProgressUpdate) {
		if testHook != nil {
			testHook(u)
		}
		progress(u)
	}

	result, err := a.catalogService.CreateCatalogWithContext(
		ctx, title, absSource, absOutput, outputRoot, copyToDirectory, catOpts, onProgress,
	)
	if err != nil {
		// A source loss retains its partial tree plus the parameters
		// needed to write it later; a cancellation (or any other failure)
		// retains nothing and returns as-is.
		var srcErr *catalog.SourceUnavailableError
		if errors.As(err, &srcErr) && srcErr.Partial != nil {
			a.scanMu.Lock()
			a.lastPartial = srcErr.Partial
			a.lastScanReq = &lastScanRequest{
				title:           title,
				sourcePath:      absSource,
				outputDir:       absOutput,
				outputRoot:      outputRoot,
				copyToDirectory: copyToDirectory,
				opts:            catOpts,
			}
			a.scanMu.Unlock()
		}
		return nil, err
	}
	return result, nil
}

// CancelScan cancels the in-flight scan, if any -- a cancel with no scan
// running is a no-op, not an error. The handle is a single field rather
// than a map keyed by scan id, because the product is one-scan-at-a-time by
// explicit requirement; a map would be over-built for it.
//
// Accepted limitation: cancellation cannot interrupt a syscall the walk is
// already blocked inside. The standard library gives os.ReadDir/os.Stat no
// context-awareness, so a device that is disconnecting (rather than already
// gone) can leave one syscall uninterruptible for as long as the OS takes
// to notice. Every syscall that has not yet started is cancelled promptly
// -- traverseDirectory checks ctx.Err() at the top of every call -- and
// wrapping each syscall in its own goroutine-plus-timeout would add real
// complexity for a rare, cosmetic worst case.
func (a *App) CancelScan() {
	a.cancelActiveScan()
}

// cancelActiveScan cancels the in-flight scan, if any, and reports whether
// one was running. Shared by CancelScan (the renderer-facing binding) and
// beforeClose's scan branch. The handle itself is left for the owning
// goroutine's own deferred cleanup to clear -- never nilled here -- so a
// second call after the scan has actually stopped reports false; this is
// what stops beforeClose from re-entering forever on the re-requested quit.
func (a *App) cancelActiveScan() bool {
	a.scanMu.Lock()
	defer a.scanMu.Unlock()
	if a.activeScanCancel == nil {
		return false
	}
	a.activeScanCancel()
	return true
}

// waitForScanStop blocks up to d for the current scan's scanDone channel to
// close, reporting whether the scan actually finished within that deadline.
// The channel reference is read under the mutex, then waited on outside
// it, so a concurrent StartScan/cleanup is never blocked behind this call.
func (a *App) waitForScanStop(d time.Duration) bool {
	a.scanMu.Lock()
	done := a.scanDone
	a.scanMu.Unlock()
	if done == nil {
		return true
	}
	select {
	case <-done:
		return true
	case <-time.After(d):
		return false
	}
}

// WritePartialCatalog writes the most recently retained partial scan tree
// (from a source-loss error) using the request parameters captured when
// that scan started. Idempotent: once a write succeeds, the cached result
// is returned on every later call without touching the filesystem again --
// this is what makes a second click a true no-op rather than a duplicate
// write or a second rename (T-25-13). Never re-walks anything -- the tree
// is exactly what the interrupted scan already produced.
//
// writeMu serializes the whole check-decide-write-record sequence across
// concurrent calls to this method (CR-02): the actual filesystem write
// still happens with scanMu released (it can be slow, and must not block
// StartScan/CancelScan), but a second overlapping call blocks on writeMu
// until the first finishes, then sees the now-cached lastPartialResult and
// returns it instead of writing a second time. retainedGen additionally
// guards against a StartScan racing in during the write: if a new scan has
// superseded the retained tree by the time this write completes, the result
// is still returned to this caller (their write genuinely happened) but is
// not cached over -- and does not clobber -- the newer retained state.
func (a *App) WritePartialCatalog() (*models.CreateCatalogResult, error) {
	a.writeMu.Lock()
	defer a.writeMu.Unlock()

	a.scanMu.Lock()
	if a.lastPartialResult != nil {
		result := a.lastPartialResult
		a.scanMu.Unlock()
		return result, nil
	}
	if a.lastPartial == nil || a.lastScanReq == nil {
		a.scanMu.Unlock()
		return nil, fmt.Errorf("no partial scan retained to write")
	}
	partial := a.lastPartial
	req := a.lastScanReq
	gen := a.retainedGen
	a.scanMu.Unlock()

	result, err := a.catalogService.WriteCatalogFrom(
		partial.Tree, req.title, req.outputDir, req.outputRoot, req.copyToDirectory, req.opts,
	)
	if err != nil {
		return nil, err
	}

	a.scanMu.Lock()
	defer a.scanMu.Unlock()
	if gen != a.retainedGen {
		// A newer StartScan cleared/replaced the retained state while this
		// write was in flight -- the write to disk still succeeded and is
		// returned to this caller, but must not overwrite the newer
		// retained tree's bookkeeping.
		return result, nil
	}
	a.lastPartialResult = result
	a.lastPartial = nil
	a.lastScanReq = nil

	return result, nil
}

// SearchCatalogs searches across catalog files for a term
func (a *App) SearchCatalogs(searchTerm string, catalogDir string) ([]*models.SearchResult, error) {
	absPath, err := filepath.Abs(catalogDir)
	if err != nil {
		return nil, err
	}

	return a.searchService.SearchCatalogs(searchTerm, absPath)
}

// SearchIndexed is the GUI-only capped sibling of SearchCatalogs, used by
// the ⌘K command palette. It caps the response at search.SearchIndexedCap
// while carrying the true match count in the returned Total.
func (a *App) SearchIndexed(searchTerm string, catalogDir string) (*models.SearchIndexResult, error) {
	absPath, err := filepath.Abs(catalogDir)
	if err != nil {
		return nil, err
	}

	return a.searchService.SearchIndexed(searchTerm, absPath)
}

// BrowseCatalogs returns metadata for all catalogs in a directory
func (a *App) BrowseCatalogs(catalogDir string) ([]*models.CatalogMetadata, error) {
	absPath, err := filepath.Abs(catalogDir)
	if err != nil {
		return nil, err
	}

	return a.searchService.BrowseCatalogs(absPath)
}

// LoadCatalog reads and parses a catalog JSON file
func (a *App) LoadCatalog(filePath string) (*models.CatalogItem, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	return a.searchService.LoadCatalog(absPath)
}

// LoadCatalogFlat reads and parses a catalog JSON file, returning it as a
// single flattened node array ready for the virtualized tree pane.
func (a *App) LoadCatalogFlat(filePath string) (*models.FlatCatalog, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	return a.searchService.LoadCatalogFlat(absPath)
}

// GetConfig returns the current configuration
func (a *App) GetConfig() *config.Config {
	if a.configManager == nil {
		return config.DefaultConfig()
	}
	return a.configManager.Get()
}

// SetTheme saves the theme preference
func (a *App) SetTheme(theme string) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetTheme(theme)
}

// SetSidebarPosition saves the sidebar position preference
func (a *App) SetSidebarPosition(position string) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetSidebarPosition(position)
}

// SetDensity saves the row-density preference
func (a *App) SetDensity(density string) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetDensity(density)
}

// SetSettingsMigrated marks whether the legacy localStorage settings keys
// have been folded into the Go config (plan 26-02's migration).
func (a *App) SetSettingsMigrated(migrated bool) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetSettingsMigrated(migrated)
}

// SetRailSide saves the catalog rail side preference
func (a *App) SetRailSide(side string) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetRailSide(side)
}

// SetCatalogDirectory saves the configured catalog directory -- the same
// value the rail's directory chip and Settings' Catalogs section share.
func (a *App) SetCatalogDirectory(dir string) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetCatalogDirectory(dir)
}

// SetDefaultFilenameRoot saves the default filename root pre-filled into
// every new catalog's create form. An empty string is a valid value.
func (a *App) SetDefaultFilenameRoot(root string) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetDefaultFilenameRoot(root)
}

// SetSecondaryDirectory saves the create form's secondary-copy destination.
func (a *App) SetSecondaryDirectory(dir string) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetSecondaryDirectory(dir)
}

// SetWindowSize saves the window size preference
func (a *App) SetWindowSize(width, height int) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetWindowSize(width, height)
}

// GetWindowPersistence returns whether window state persistence is enabled
func (a *App) GetWindowPersistence() bool {
	if a.configManager == nil {
		return true
	}
	return a.configManager.GetWindowPersistence()
}

// SetWindowPersistence saves the window persistence preference
func (a *App) SetWindowPersistence(enabled bool) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetWindowPersistence(enabled)
}

// SetWindowPosition saves the window position
func (a *App) SetWindowPosition(x, y int) error {
	if a.configManager == nil {
		return nil
	}
	return a.configManager.SetWindowPosition(x, y)
}

// domReady is called after the frontend DOM is ready
func (a *App) domReady(ctx context.Context) {
	cfg := a.configManager.Get()
	if cfg == nil || !cfg.WindowPersistenceEnabled {
		return
	}
	runtime.WindowSetSize(ctx, cfg.WindowWidth, cfg.WindowHeight)
	// Restore position only if non-zero (skip OS default placement for 0,0)
	if cfg.WindowX != 0 || cfg.WindowY != 0 {
		runtime.WindowSetPosition(ctx, cfg.WindowX, cfg.WindowY)
	}
}

// beforeClose is called before the application closes
func (a *App) beforeClose(ctx context.Context) bool {
	// Intercepted at the very first opportunity -- before the
	// configuration-manager nil check and the window-persistence save below
	// -- so a close request during an active scan cancels the walk and lets
	// it stop before the process actually exits (CRT-13). Non-recursion on
	// the re-requested quit is guaranteed by the handle already being
	// cleared by the scan goroutine's own deferred cleanup on the second
	// pass, not by a separate flag: cancelActiveScan then reports false and
	// this branch is skipped entirely.
	//
	// This exact cancel-then-wait-then-requery sequence is synthesised from
	// individually verified Wails primitives (OnBeforeClose's prevent-bool
	// return, runtime.Quit) rather than found documented end-to-end
	// anywhere -- it carries an explicit live-verification obligation
	// (force-quit mid-scan, this task's manual step and 25-VALIDATION.md's
	// own manual-only entry), not a claim this test suite can make on its
	// own (25-RESEARCH.md Assumption A2).
	if a.cancelActiveScan() {
		go func() {
			a.waitForScanStop(3 * time.Second)
			runtime.Quit(ctx)
		}()
		return true // prevent this close; the re-requested quit re-enters this hook
	}

	if a.configManager == nil {
		return false
	}
	cfg := a.configManager.Get()
	if cfg != nil && cfg.WindowPersistenceEnabled {
		w, h := runtime.WindowGetSize(ctx)
		x, y := runtime.WindowGetPosition(ctx)
		_ = a.configManager.SetWindowSize(w, h)
		_ = a.configManager.SetWindowPosition(x, y)
	}
	return false // false = allow close
}

// SelectDirectory opens a directory selection dialog
func (a *App) SelectDirectory() (string, error) {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Directory",
	})
	return path, err
}

// ListVolumes returns the machine's currently mounted volumes -- name,
// mount path, size, free space and a readable flag -- for the create
// flow's volume-card picker (CRT-02). Delegates to internal/volumes with
// no logic of its own, matching this file's other thin bindings.
func (a *App) ListVolumes() ([]volumes.Volume, error) {
	return volumes.List()
}

// ReadHtmlFile reads the contents of an HTML file
func (a *App) ReadHtmlFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// GetCatalogHtmlPath returns the HTML file path for a catalog
func (a *App) GetCatalogHtmlPath(catalogPath string) (string, error) {
	var htmlPath string
	if filepath.Ext(catalogPath) == ".json" {
		htmlPath = catalogPath[:len(catalogPath)-5] + ".html"
	} else {
		htmlPath = catalogPath + ".html"
	}
	if _, err := os.Stat(htmlPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("HTML file not found: %s", htmlPath)
		}
		return "", fmt.Errorf("cannot access HTML file: %w", err)
	}
	return htmlPath, nil
}

// OpenExternal opens a URL or file in the system's default application
func (a *App) OpenExternal(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}

// RevealInFileManager asks the operating system to reveal path -- a
// catalog's JSON file or its HTML companion -- in the platform's file
// manager, selected within its containing folder. catalogDir is the
// frontend's currently configured catalog directory; internal/osutil
// rejects any path that does not resolve inside it, in addition to
// validating the path itself and building an argv-only command. This
// binding adds no logic of its own.
func (a *App) RevealInFileManager(path string, catalogDir string) error {
	return osutil.RevealInFileManager(path, catalogDir)
}

// GetVersion returns the application version injected at build time
func (a *App) GetVersion() string {
	return Version
}
