package config

import (
	"os"
	"path/filepath"
	"sync"
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

// TestSetDensity verifies SetDensity updates the in-memory config.
func TestSetDensity(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetDensity("Compact"); err != nil {
		t.Fatalf("SetDensity(\"Compact\") error: %v", err)
	}

	if got := m.Get().Density; got != "Compact" {
		t.Errorf("Get().Density = %q, want %q", got, "Compact")
	}
}

// TestSetDensity_Persists verifies SetDensity writes synchronously to disk --
// a second Manager built on the same configPath and Load()ed reports the
// same value, proving Save() ran inside the setter (SET-05, no batching).
func TestSetDensity_Persists(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetDensity("Compact"); err != nil {
		t.Fatalf("SetDensity(\"Compact\") error: %v", err)
	}

	m2 := &Manager{configPath: m.configPath}
	if err := m2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := m2.Get().Density; got != "Compact" {
		t.Errorf("after reload: Density = %q, want %q", got, "Compact")
	}
}

// TestDefaultConfig_SettingsFields verifies DefaultConfig's new Density and
// SettingsMigrated fields.
func TestDefaultConfig_SettingsFields(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Density != "Comfortable" {
		t.Errorf("DefaultConfig().Density = %q, want %q", cfg.Density, "Comfortable")
	}
	if cfg.SettingsMigrated != false {
		t.Errorf("DefaultConfig().SettingsMigrated = true, want false")
	}
}

// TestSetSettingsMigrated verifies SetSettingsMigrated round-trips through
// Get() and through a reloaded Manager.
func TestSetSettingsMigrated(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetSettingsMigrated(true); err != nil {
		t.Fatalf("SetSettingsMigrated(true) error: %v", err)
	}
	if got := m.Get().SettingsMigrated; got != true {
		t.Errorf("Get().SettingsMigrated = %v, want true", got)
	}

	m2 := &Manager{configPath: m.configPath}
	if err := m2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got := m2.Get().SettingsMigrated; got != true {
		t.Errorf("after reload: SettingsMigrated = %v, want true", got)
	}
}

// TestSetDensity_Idempotent verifies calling SetDensity twice with the same
// value leaves the on-disk file byte-identical (SET-05 idempotency edge) --
// no duplicate keys, no appended history.
func TestSetDensity_Idempotent(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetDensity("Compact"); err != nil {
		t.Fatalf("first SetDensity(\"Compact\") error: %v", err)
	}
	first, err := os.ReadFile(m.configPath)
	if err != nil {
		t.Fatalf("ReadFile after first SetDensity: %v", err)
	}

	if err := m.SetDensity("Compact"); err != nil {
		t.Fatalf("second SetDensity(\"Compact\") error: %v", err)
	}
	second, err := os.ReadFile(m.configPath)
	if err != nil {
		t.Fatalf("ReadFile after second SetDensity: %v", err)
	}

	if string(first) != string(second) {
		t.Errorf("on-disk config not byte-identical after idempotent SetDensity:\nfirst:  %s\nsecond: %s", first, second)
	}
}

// TestSetRailSide verifies SetRailSide updates the in-memory config.
func TestSetRailSide(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetRailSide("Right"); err != nil {
		t.Fatalf("SetRailSide(\"Right\") error: %v", err)
	}

	if got := m.Get().RailSide; got != "Right" {
		t.Errorf("Get().RailSide = %q, want %q", got, "Right")
	}
}

// TestSetRailSide_Persists verifies SetRailSide writes synchronously to
// disk -- a second Manager loaded from the same path reports the same
// value.
func TestSetRailSide_Persists(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetRailSide("Right"); err != nil {
		t.Fatalf("SetRailSide(\"Right\") error: %v", err)
	}

	m2 := &Manager{configPath: m.configPath}
	if err := m2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := m2.Get().RailSide; got != "Right" {
		t.Errorf("after reload: RailSide = %q, want %q", got, "Right")
	}
}

// TestDefaultConfig_RailSide verifies DefaultConfig's RailSide default.
func TestDefaultConfig_RailSide(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.RailSide != "Left" {
		t.Errorf("DefaultConfig().RailSide = %q, want %q", cfg.RailSide, "Left")
	}
}

// TestSetRailSide_DoesNotTouchSidebarPosition pins that RailSide and
// SidebarPosition never alias -- they are unrelated concepts
// (26-RESEARCH.md Pitfall 2).
func TestSetRailSide_DoesNotTouchSidebarPosition(t *testing.T) {
	m := newTestManager(t)
	want := DefaultConfig().SidebarPosition

	if err := m.SetRailSide("Right"); err != nil {
		t.Fatalf("SetRailSide(\"Right\") error: %v", err)
	}

	if got := m.Get().SidebarPosition; got != want {
		t.Errorf("Get().SidebarPosition = %q, want unchanged %q", got, want)
	}
}

// TestSetCatalogDirectory verifies SetCatalogDirectory updates the
// in-memory config.
func TestSetCatalogDirectory(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetCatalogDirectory("/Volumes/Data"); err != nil {
		t.Fatalf("SetCatalogDirectory(\"/Volumes/Data\") error: %v", err)
	}

	if got := m.Get().CatalogDirectory; got != "/Volumes/Data" {
		t.Errorf("Get().CatalogDirectory = %q, want %q", got, "/Volumes/Data")
	}
}

// TestSetCatalogDirectory_Persists verifies SetCatalogDirectory writes
// synchronously to disk -- a second Manager loaded from the same path
// reports the same value.
func TestSetCatalogDirectory_Persists(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetCatalogDirectory("/Volumes/Data"); err != nil {
		t.Fatalf("SetCatalogDirectory(\"/Volumes/Data\") error: %v", err)
	}

	m2 := &Manager{configPath: m.configPath}
	if err := m2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := m2.Get().CatalogDirectory; got != "/Volumes/Data" {
		t.Errorf("after reload: CatalogDirectory = %q, want %q", got, "/Volumes/Data")
	}
}

// TestDefaultConfig_CatalogDirectoryEmpty verifies a fresh install has no
// catalog directory configured -- the state the empty-state chip copy
// covers.
func TestDefaultConfig_CatalogDirectoryEmpty(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.CatalogDirectory != "" {
		t.Errorf("DefaultConfig().CatalogDirectory = %q, want empty", cfg.CatalogDirectory)
	}
}

// TestSetDefaultFilenameRoot verifies SetDefaultFilenameRoot updates the
// in-memory config.
func TestSetDefaultFilenameRoot(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetDefaultFilenameRoot("my-catalog"); err != nil {
		t.Fatalf("SetDefaultFilenameRoot(\"my-catalog\") error: %v", err)
	}

	if got := m.Get().DefaultFilenameRoot; got != "my-catalog" {
		t.Errorf("Get().DefaultFilenameRoot = %q, want %q", got, "my-catalog")
	}
}

// TestSetDefaultFilenameRoot_Persists verifies SetDefaultFilenameRoot
// writes synchronously to disk -- a second Manager loaded from the same
// path reports the same value.
func TestSetDefaultFilenameRoot_Persists(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetDefaultFilenameRoot("my-catalog"); err != nil {
		t.Fatalf("SetDefaultFilenameRoot(\"my-catalog\") error: %v", err)
	}

	m2 := &Manager{configPath: m.configPath}
	if err := m2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := m2.Get().DefaultFilenameRoot; got != "my-catalog" {
		t.Errorf("after reload: DefaultFilenameRoot = %q, want %q", got, "my-catalog")
	}
}

// TestSetDefaultFilenameRoot_EmptyIsValid verifies setting "" returns nil
// and reads back as "" -- an empty root is a legitimate stored value, not
// an error and not a sentinel for "unset by mistake".
func TestSetDefaultFilenameRoot_EmptyIsValid(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetDefaultFilenameRoot("my-catalog"); err != nil {
		t.Fatalf("SetDefaultFilenameRoot(\"my-catalog\") error: %v", err)
	}
	if err := m.SetDefaultFilenameRoot(""); err != nil {
		t.Fatalf("SetDefaultFilenameRoot(\"\") error: %v", err)
	}

	if got := m.Get().DefaultFilenameRoot; got != "" {
		t.Errorf("Get().DefaultFilenameRoot = %q, want empty", got)
	}
}

// TestSetSecondaryDirectory verifies SetSecondaryDirectory updates the
// in-memory config.
func TestSetSecondaryDirectory(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetSecondaryDirectory("/Volumes/Backup"); err != nil {
		t.Fatalf("SetSecondaryDirectory(\"/Volumes/Backup\") error: %v", err)
	}

	if got := m.Get().SecondaryDirectory; got != "/Volumes/Backup" {
		t.Errorf("Get().SecondaryDirectory = %q, want %q", got, "/Volumes/Backup")
	}
}

// TestSetSecondaryDirectory_Persists verifies SetSecondaryDirectory writes
// synchronously to disk -- a second Manager loaded from the same path
// reports the same value.
func TestSetSecondaryDirectory_Persists(t *testing.T) {
	m := newTestManager(t)

	if err := m.SetSecondaryDirectory("/Volumes/Backup"); err != nil {
		t.Fatalf("SetSecondaryDirectory(\"/Volumes/Backup\") error: %v", err)
	}

	m2 := &Manager{configPath: m.configPath}
	if err := m2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := m2.Get().SecondaryDirectory; got != "/Volumes/Backup" {
		t.Errorf("after reload: SecondaryDirectory = %q, want %q", got, "/Volumes/Backup")
	}
}

// TestManager_ConcurrentSetters interleaves SetDensity, SetSettingsMigrated
// and Get() across goroutines and asserts no error is returned. Run under
// -race (SET-05 concurrency edge) -- this is the reason Manager gained a
// lock.
func TestManager_ConcurrentSetters(t *testing.T) {
	m := newTestManager(t)

	const n = 20
	var wg sync.WaitGroup
	errCh := make(chan error, n*2)

	for i := 0; i < n; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			if err := m.SetDensity("Compact"); err != nil {
				errCh <- err
			}
		}()
		go func() {
			defer wg.Done()
			if err := m.SetSettingsMigrated(true); err != nil {
				errCh <- err
			}
		}()
		go func() {
			defer wg.Done()
			_ = m.Get()
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent setter returned error: %v", err)
	}
}
