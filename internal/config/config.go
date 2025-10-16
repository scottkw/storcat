package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds application settings
type Config struct {
	Theme           string `json:"theme"`            // "light" or "dark"
	SidebarPosition string `json:"sidebarPosition"`  // "left" or "right"
	WindowWidth     int    `json:"windowWidth"`
	WindowHeight    int    `json:"windowHeight"`
}

// DefaultConfig returns default application settings
func DefaultConfig() *Config {
	return &Config{
		Theme:           "light",
		SidebarPosition: "left",
		WindowWidth:     1200,
		WindowHeight:    800,
	}
}

// Manager handles config persistence
type Manager struct {
	configPath string
	config     *Config
}

// NewManager creates a new config manager
func NewManager() (*Manager, error) {
	// Get config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	// Create storcat config directory
	storcatConfigDir := filepath.Join(configDir, "storcat")
	if err := os.MkdirAll(storcatConfigDir, 0755); err != nil {
		return nil, err
	}

	configPath := filepath.Join(storcatConfigDir, "config.json")

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
			m.config = DefaultConfig()
			return nil
		}
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	m.config = &config
	return nil
}

// Save writes config to disk
func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// Get returns current config
func (m *Manager) Get() *Config {
	return m.config
}

// SetTheme updates theme setting
func (m *Manager) SetTheme(theme string) error {
	m.config.Theme = theme
	return m.Save()
}

// SetSidebarPosition updates sidebar position
func (m *Manager) SetSidebarPosition(position string) error {
	m.config.SidebarPosition = position
	return m.Save()
}

// SetWindowSize updates window dimensions
func (m *Manager) SetWindowSize(width, height int) error {
	m.config.WindowWidth = width
	m.config.WindowHeight = height
	return m.Save()
}
