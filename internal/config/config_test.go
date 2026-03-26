package config

import (
	"os"
	"path/filepath"
	"testing"
)

// newTestManager creates a Manager with a config file in a temp directory.
// Each test gets its own temp dir for isolation.
func newTestManager(t *testing.T) *Manager {
	t.Helper()
	dir, err := os.MkdirTemp("", "storcat-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	configPath := filepath.Join(dir, "config.json")
	m := &Manager{
		configPath: configPath,
		config:     DefaultConfig(),
	}
	if err := m.Save(); err != nil {
		t.Fatalf("failed to save initial config: %v", err)
	}
	return m
}

// TestDefaultConfig_WindowFields verifies DefaultConfig returns correct zero/true values.
func TestDefaultConfig_WindowFields(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.WindowX != 0 {
		t.Errorf("DefaultConfig().WindowX = %d, want 0", cfg.WindowX)
	}
	if cfg.WindowY != 0 {
		t.Errorf("DefaultConfig().WindowY = %d, want 0", cfg.WindowY)
	}
	if !cfg.WindowPersistenceEnabled {
		t.Errorf("DefaultConfig().WindowPersistenceEnabled = false, want true")
	}
}

// TestSetWindowPosition verifies SetWindowPosition updates the in-memory config.
func TestSetWindowPosition(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetWindowPosition(100, 200); err != nil {
		t.Fatalf("SetWindowPosition(100, 200) error: %v", err)
	}

	cfg := m.Get()
	if cfg.WindowX != 100 {
		t.Errorf("WindowX = %d, want 100", cfg.WindowX)
	}
	if cfg.WindowY != 200 {
		t.Errorf("WindowY = %d, want 200", cfg.WindowY)
	}
}

// TestSetWindowPosition_Persists verifies SetWindowPosition writes values to disk.
func TestSetWindowPosition_Persists(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetWindowPosition(150, 250); err != nil {
		t.Fatalf("SetWindowPosition(150, 250) error: %v", err)
	}

	// Reload from disk
	m2 := &Manager{configPath: m.configPath}
	if err := m2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if m2.config.WindowX != 150 {
		t.Errorf("after reload: WindowX = %d, want 150", m2.config.WindowX)
	}
	if m2.config.WindowY != 250 {
		t.Errorf("after reload: WindowY = %d, want 250", m2.config.WindowY)
	}
}

// TestSetWindowPersistence verifies SetWindowPersistence updates the in-memory config
// and GetWindowPersistence returns the updated value.
func TestSetWindowPersistence(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetWindowPersistence(false); err != nil {
		t.Fatalf("SetWindowPersistence(false) error: %v", err)
	}

	if m.GetWindowPersistence() != false {
		t.Errorf("GetWindowPersistence() = true, want false after SetWindowPersistence(false)")
	}
}

// TestSetWindowPersistence_Persists verifies SetWindowPersistence writes to disk.
func TestSetWindowPersistence_Persists(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetWindowPersistence(false); err != nil {
		t.Fatalf("SetWindowPersistence(false) error: %v", err)
	}

	// Reload from disk
	m2 := &Manager{configPath: m.configPath}
	if err := m2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if m2.config.WindowPersistenceEnabled != false {
		t.Errorf("after reload: WindowPersistenceEnabled = true, want false")
	}
}

// TestGetWindowPersistence_Default verifies a fresh config has persistence enabled.
func TestGetWindowPersistence_Default(t *testing.T) {
	m := newTestManager(t)

	if !m.GetWindowPersistence() {
		t.Errorf("GetWindowPersistence() = false on fresh config, want true")
	}
}

// TestWindowPosition_RoundTrip verifies SetWindowPosition values are accessible via Get().
func TestWindowPosition_RoundTrip(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetWindowPosition(300, 400); err != nil {
		t.Fatalf("SetWindowPosition(300, 400) error: %v", err)
	}

	cfg := m.Get()
	if cfg.WindowX != 300 {
		t.Errorf("Get().WindowX = %d, want 300", cfg.WindowX)
	}
	if cfg.WindowY != 400 {
		t.Errorf("Get().WindowY = %d, want 400", cfg.WindowY)
	}
}
