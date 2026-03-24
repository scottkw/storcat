package main

import (
	"context"
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
	// Remove the .json extension and add .html
	if filepath.Ext(catalogPath) == ".json" {
		htmlPath := catalogPath[:len(catalogPath)-5] + ".html"
		return htmlPath, nil
	}
	return catalogPath + ".html", nil
}

// OpenExternal opens a URL or file in the system's default application
func (a *App) OpenExternal(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}
