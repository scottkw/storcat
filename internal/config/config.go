package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Config holds application settings
type Config struct {
	Theme string `json:"theme"` // "light" or "dark"
	// SidebarPosition is an intentionally-orphaned v1 leftover, distinct from
	// this milestone's rail-side concept (which lives in the frontend's
	// AppContext/localStorage, not here). Left in place, untouched
	// (26-RESEARCH.md Pitfall 2, resolved by 26-CONTEXT.md's discretion note).
	SidebarPosition          string `json:"sidebarPosition"` // "left" or "right"
	WindowWidth              int    `json:"windowWidth"`
	WindowHeight             int    `json:"windowHeight"`
	WindowX                  int    `json:"windowX"`
	WindowY                  int    `json:"windowY"`
	WindowPersistenceEnabled bool   `json:"windowPersistenceEnabled"`
	Density                  string `json:"density"`          // "Compact" or "Comfortable"
	SettingsMigrated         bool   `json:"settingsMigrated"` // set once legacy localStorage settings keys have been folded into this config
	// RailSide is the catalog rail's side ("Left" or "Right") -- a genuinely
	// new field, not a rename or repurpose of the orphaned SidebarPosition
	// above (26-CONTEXT.md's discretion resolution).
	RailSide string `json:"railSide"`
	// CatalogDirectory, DefaultFilenameRoot and SecondaryDirectory are the
	// Catalogs section's config-backed settings (SET-04) -- the rail's
	// directory chip, the create form's filename-root seed, and the
	// create-form's secondary-copy destination all read/write these fields
	// (via settingsStore.ts) instead of a private localStorage copy each.
	CatalogDirectory    string `json:"catalogDirectory"`
	DefaultFilenameRoot string `json:"defaultFilenameRoot"`
	SecondaryDirectory  string `json:"secondaryDirectory"`
	// WriteHTML and CopyToSecondary are the create flow's default option
	// values (SET-04) -- a Settings change here is what the create
	// slide-over's own toggles open with next time.
	WriteHTML       bool `json:"writeHtml"`
	CopyToSecondary bool `json:"copyToSecondary"`
	// WatchDirectory persists only, this phase (26). Phase 27's
	// WATCH-01..03 own the real fsnotify watcher; until then this field has
	// no reader beyond the Settings toggle itself, and no surface (status
	// bar, rail badge, or copy) may imply that watching is active.
	WatchDirectory bool `json:"watchDirectory"`
}

// DefaultConfig returns default application settings
func DefaultConfig() *Config {
	return &Config{
		Theme:                    "light",
		SidebarPosition:          "left",
		WindowWidth:              1200,
		WindowHeight:             800,
		WindowX:                  0,
		WindowY:                  0,
		WindowPersistenceEnabled: true,
		Density:                  "Comfortable",
		SettingsMigrated:         false,
		RailSide:                 "Left",
		CatalogDirectory:         "",
		DefaultFilenameRoot:      "",
		SecondaryDirectory:       "",
		WriteHTML:                true,
		CopyToSecondary:          false,
		WatchDirectory:           false,
	}
}

// Manager handles config persistence
type Manager struct {
	configPath string
	config     *Config
	// mu guards every read and write of config below. Get() returns a copy
	// under RLock rather than the live *Config pointer so a caller can never
	// race a concurrent Set* call mutating the same struct from a different
	// Wails-dispatched goroutine (T-26-02).
	mu sync.RWMutex
}

// storcatConfigDir resolves (and creates, if missing) the directory
// storcat's local JSON files live in -- config.json today, counts-cache.json
// alongside it. Both the config Manager and the sidecar counts cache
// resolve through this one helper so the directory can never drift between
// the two.
func storcatConfigDir() (string, error) {
	// Get config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	// Create storcat config directory
	dir := filepath.Join(configDir, "storcat")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// NewManager creates a new config manager
func NewManager() (*Manager, error) {
	storcatDir, err := storcatConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(storcatDir, "config.json")

	m := &Manager{
		configPath: configPath,
	}

	// Load existing config or use defaults
	if err := m.Load(); err != nil {
		m.config = DefaultConfig()
		// Save default config
		_ = m.Save()
	}

	return m, nil
}

// Load reads config from disk
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.mu.Lock()
			m.config = DefaultConfig()
			m.mu.Unlock()
			return nil
		}
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	m.mu.Lock()
	m.config = &config
	m.mu.Unlock()
	return nil
}

// Save writes config to disk
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveLocked()
}

// saveLocked marshals and writes m.config to disk. Callers must already
// hold m.mu (write lock) -- it takes no lock of its own so a setter can
// mutate and save atomically inside one critical section without a
// double-acquire against the exported Save().
func (m *Manager) saveLocked() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// Get returns a copy of the current config, not the live pointer. This is a
// deliberate behavior change from earlier versions: App.GetConfig, domReady
// and beforeClose only ever read fields off the result, and returning a
// copy under RLock is what makes concurrent Set* calls arriving on separate
// Wails goroutines race-free for those callers.
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfg := *m.config
	return &cfg
}

// SetTheme updates theme setting
func (m *Manager) SetTheme(theme string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.Theme = theme
	return m.saveLocked()
}

// SetSidebarPosition updates sidebar position
func (m *Manager) SetSidebarPosition(position string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.SidebarPosition = position
	return m.saveLocked()
}

// SetWindowSize updates window dimensions
func (m *Manager) SetWindowSize(width, height int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.WindowWidth = width
	m.config.WindowHeight = height
	return m.saveLocked()
}

// SetWindowPosition updates window coordinates
func (m *Manager) SetWindowPosition(x, y int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.WindowX = x
	m.config.WindowY = y
	return m.saveLocked()
}

// SetWindowPersistence updates the window state persistence toggle
func (m *Manager) SetWindowPersistence(enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.WindowPersistenceEnabled = enabled
	return m.saveLocked()
}

// GetWindowPersistence returns whether window state persistence is enabled
func (m *Manager) GetWindowPersistence() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.WindowPersistenceEnabled
}

// SetDensity updates the row-density preference
func (m *Manager) SetDensity(density string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.Density = density
	return m.saveLocked()
}

// SetSettingsMigrated marks whether the legacy localStorage settings keys
// have been folded into this config (consumed by plan 26-02's migration).
func (m *Manager) SetSettingsMigrated(migrated bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.SettingsMigrated = migrated
	return m.saveLocked()
}

// SetRailSide updates the catalog rail's side preference.
func (m *Manager) SetRailSide(side string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.RailSide = side
	return m.saveLocked()
}

// SetCatalogDirectory updates the configured catalog directory -- the same
// value the rail's directory chip and the Settings Catalogs section both
// read and write (SET-04).
func (m *Manager) SetCatalogDirectory(dir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.CatalogDirectory = dir
	return m.saveLocked()
}

// SetDefaultFilenameRoot updates the default filename root pre-filled into
// every new catalog's create form. An empty string is a valid value (SET-04
// edge case), not an error.
func (m *Manager) SetDefaultFilenameRoot(root string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.DefaultFilenameRoot = root
	return m.saveLocked()
}

// SetSecondaryDirectory updates the create form's secondary-copy
// destination.
func (m *Manager) SetSecondaryDirectory(dir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.SecondaryDirectory = dir
	return m.saveLocked()
}

// SetWriteHTML updates the create flow's write-HTML default.
func (m *Manager) SetWriteHTML(enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.WriteHTML = enabled
	return m.saveLocked()
}

// SetCopyToSecondary updates the create flow's copy-to-secondary default.
func (m *Manager) SetCopyToSecondary(enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.CopyToSecondary = enabled
	return m.saveLocked()
}

// SetWatchDirectory persists the watch-directory toggle's value. This phase
// (26) only stores the value -- Phase 27's WATCH-01..03 own the actual
// fsnotify watcher.
func (m *Manager) SetWatchDirectory(enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.WatchDirectory = enabled
	return m.saveLocked()
}
