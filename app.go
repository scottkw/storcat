package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"storcat-wails/internal/catalog"
	"storcat-wails/internal/config"
	"storcat-wails/internal/search"
	"storcat-wails/pkg/models"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	catalogService *catalog.Service
	searchService  *search.Service
	configManager  *config.Manager
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
	progressCallback := func(path string) {
		// For now, we don't send progress updates
		// In the future, we could use Wails events to send updates to frontend
	}

	result, err := a.catalogService.CreateCatalog(title, absPath, outputName, copyToDirectory, progressCallback)
	if err != nil {
		return nil, err
	}
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

// GetVersion returns the application version injected at build time
func (a *App) GetVersion() string {
	return Version
}
