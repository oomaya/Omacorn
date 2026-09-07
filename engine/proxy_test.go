package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetEnvDPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	p, err := getEnvDPath()
	if err != nil {
		t.Fatalf("getEnvDPath() failed: %v", err)
	}

	expected := filepath.Join(tmpDir, ".config", "environment.d", "20-omacorn-proxy.conf")
	if p != expected {
		t.Errorf("expected path %q, got %q", expected, p)
	}
}

func TestProxyLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Initially not configured
	if IsProxyConfigured() {
		t.Errorf("expected IsProxyConfigured to be false in fresh temp dir")
	}

	// Enable proxy
	if err := EnableSystemProxy(); err != nil {
		t.Fatalf("EnableSystemProxy() failed: %v", err)
	}

	if !IsProxyConfigured() {
		t.Errorf("expected IsProxyConfigured to be true after EnableSystemProxy")
	}

	p, err := getEnvDPath()
	if err != nil {
		t.Fatalf("getEnvDPath failed: %v", err)
	}

	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("failed to read proxy file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "127.0.0.1:8080") {
		t.Errorf("proxy file missing 127.0.0.1:8080: %s", content)
	}
	if !strings.Contains(content, "accounts.google.com") {
		t.Errorf("proxy file missing Google Auth exemption: %s", content)
	}

	// Disable proxy
	if err := DisableSystemProxy(); err != nil {
		t.Fatalf("DisableSystemProxy() failed: %v", err)
	}

	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("expected proxy file to be removed after DisableSystemProxy")
	}
}

func TestSyncBrowserFlags(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfgDir := filepath.Join(tmpDir, ".config")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatalf("failed to create mock config dir: %v", err)
	}

	// Create a mock brave-origin-nightly-flags.conf
	braveFlagsPath := filepath.Join(cfgDir, "brave-origin-nightly-flags.conf")
	initialContent := "--ozone-platform=wayland\n--top-chrome-touch-ui=disabled\n"
	if err := os.WriteFile(braveFlagsPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write mock brave flags: %v", err)
	}

	// Enable browser flags
	syncBrowserFlags(true)

	updatedData, err := os.ReadFile(braveFlagsPath)
	if err != nil {
		t.Fatalf("failed to read updated brave flags: %v", err)
	}
	updatedContent := string(updatedData)
	if !strings.Contains(updatedContent, "--proxy-server=http://127.0.0.1:8080") {
		t.Errorf("expected proxy-server flag to be injected: %s", updatedContent)
	}
	if !strings.Contains(updatedContent, "--top-chrome-touch-ui=disabled") {
		t.Errorf("expected original flags to be preserved: %s", updatedContent)
	}

	// Disable browser flags
	syncBrowserFlags(false)

	revertedData, err := os.ReadFile(braveFlagsPath)
	if err != nil {
		t.Fatalf("failed to read reverted brave flags: %v", err)
	}
	revertedContent := string(revertedData)
	if strings.Contains(revertedContent, "--proxy-server=http://127.0.0.1:8080") {
		t.Errorf("expected proxy-server flag to be removed: %s", revertedContent)
	}
	if !strings.Contains(revertedContent, "--ozone-platform=wayland") {
		t.Errorf("expected original flags to remain preserved: %s", revertedContent)
	}
}

func TestSyncBrowserFlags_Flatpak(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Mock Flatpak Brave app dir: ~/.var/app/com.brave.Browser/config/
	flatpakCfgDir := filepath.Join(tmpDir, ".var", "app", "com.brave.Browser", "config")
	if err := os.MkdirAll(flatpakCfgDir, 0755); err != nil {
		t.Fatalf("failed to create mock flatpak dir: %v", err)
	}

	braveFlatpakFlags := filepath.Join(flatpakCfgDir, "brave-flags.conf")
	initialContent := "--enable-features=UseOzonePlatform\n"
	if err := os.WriteFile(braveFlatpakFlags, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write mock flatpak flags: %v", err)
	}

	// 1. Enable proxy flags for Flatpak
	syncBrowserFlags(true)

	updatedData, err := os.ReadFile(braveFlatpakFlags)
	if err != nil {
		t.Fatalf("failed to read updated flatpak flags: %v", err)
	}
	updatedContent := string(updatedData)
	if !strings.Contains(updatedContent, "--proxy-server=http://127.0.0.1:8080") {
		t.Errorf("expected proxy-server flag to be injected in flatpak flags: %s", updatedContent)
	}
	if !strings.Contains(updatedContent, "--enable-features=UseOzonePlatform") {
		t.Errorf("expected original flatpak flags to be preserved: %s", updatedContent)
	}

	// 2. Disable proxy flags for Flatpak
	syncBrowserFlags(false)

	revertedData, err := os.ReadFile(braveFlatpakFlags)
	if err != nil {
		t.Fatalf("failed to read reverted flatpak flags: %v", err)
	}
	revertedContent := string(revertedData)
	if strings.Contains(revertedContent, "--proxy-server=http://127.0.0.1:8080") {
		t.Errorf("expected proxy-server flag to be removed from flatpak flags: %s", revertedContent)
	}
	if !strings.Contains(revertedContent, "--enable-features=UseOzonePlatform") {
		t.Errorf("expected original flatpak flags to remain preserved: %s", revertedContent)
	}
}

func TestToggleSystemProxy(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// 1st toggle: enables
	enabled, err := ToggleSystemProxy()
	if err != nil {
		t.Fatalf("1st ToggleSystemProxy failed: %v", err)
	}
	if !enabled {
		t.Errorf("expected 1st toggle to enable proxy")
	}

	// 2nd toggle: disables
	enabled, err = ToggleSystemProxy()
	if err != nil {
		t.Fatalf("2nd ToggleSystemProxy failed: %v", err)
	}
	if enabled {
		t.Errorf("expected 2nd toggle to disable proxy")
	}
}
