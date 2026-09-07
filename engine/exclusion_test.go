package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultExclusions(t *testing.T) {
	if len(DefaultExclusions) == 0 {
		t.Fatalf("expected non-empty DefaultExclusions")
	}

	compiled := CompileNoProxy()
	if !strings.Contains(compiled, "localhost") {
		t.Errorf("expected localhost in compiled no_proxy")
	}
	if !strings.Contains(compiled, "accounts.google.com") {
		t.Errorf("expected accounts.google.com in compiled no_proxy")
	}
	if !strings.Contains(compiled, "*.telegram.org") {
		t.Errorf("expected *.telegram.org in compiled no_proxy")
	}
	if !strings.Contains(compiled, "149.154.160.0/20") {
		t.Errorf("expected Telegram CIDR 149.154.160.0/20 in compiled no_proxy")
	}
}

func TestCompileBrowserBypass(t *testing.T) {
	bypass := CompileBrowserBypass()
	if !strings.HasPrefix(bypass, "<-loopback>;") {
		t.Errorf("expected <-loopback>; prefix in browser bypass string, got: %s", bypass)
	}
	if !strings.Contains(bypass, "accounts.google.com") {
		t.Errorf("expected accounts.google.com in browser bypass string")
	}
	if strings.Contains(bypass, "::") {
		t.Errorf("raw IPv6 CIDR should be filtered out from browser bypass flags: %s", bypass)
	}
}

func TestUserExclusionsLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Initially, user exclusions should be empty
	custom, err := LoadUserExclusions()
	if err != nil {
		t.Fatalf("LoadUserExclusions failed: %v", err)
	}
	if len(custom) != 0 {
		t.Errorf("expected 0 custom exclusions, got %d", len(custom))
	}

	// Add custom exclusion
	testDomain := "special-banking-portal.example.com"
	if err := AddExclusion(testDomain); err != nil {
		t.Fatalf("AddExclusion failed: %v", err)
	}

	// Should be listed in custom and in all
	custom, _ = LoadUserExclusions()
	if len(custom) != 1 || custom[0] != testDomain {
		t.Errorf("expected [%s], got %v", testDomain, custom)
	}

	all := GetAllExclusions()
	found := false
	for _, item := range all {
		if item == testDomain {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected %s in GetAllExclusions()", testDomain)
	}

	// Duplicate add should error
	if err := AddExclusion(testDomain); err == nil {
		t.Errorf("expected error on duplicate AddExclusion")
	}

	// Adding item in defaults should error
	if err := AddExclusion("accounts.google.com"); err == nil {
		t.Errorf("expected error adding default exclusion")
	}

	// Remove exclusion
	if err := RemoveExclusion(testDomain); err != nil {
		t.Fatalf("RemoveExclusion failed: %v", err)
	}
	custom, _ = LoadUserExclusions()
	if len(custom) != 0 {
		t.Errorf("expected 0 custom exclusions after removal, got %d", len(custom))
	}

	// Removing non-existent rule should error
	if err := RemoveExclusion("non-existent-rule.com"); err == nil {
		t.Errorf("expected error on non-existent RemoveExclusion")
	}

	// Add again and test reset
	_ = AddExclusion("site-a.com")
	_ = AddExclusion("site-b.com")
	if err := ResetExclusions(); err != nil {
		t.Fatalf("ResetExclusions failed: %v", err)
	}

	custom, _ = LoadUserExclusions()
	if len(custom) != 0 {
		t.Errorf("expected 0 exclusions after ResetExclusions, got %d", len(custom))
	}

	// Config file should be gone
	cfgPath := filepath.Join(tmpDir, ".config", "omacorn", "exclusions.conf")
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		t.Errorf("expected config file to be removed after reset")
	}
}
