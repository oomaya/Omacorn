package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetProbeTarget_EnvironmentOverride(t *testing.T) {
	t.Setenv("OMACORN_TEST_TARGET", "https://example.com/test")

	url, label := GetProbeTarget()
	if url != "https://example.com/test" {
		t.Errorf("expected url https://example.com/test, got %s", url)
	}
	if label != "<custom-target>" {
		t.Errorf("expected label <custom-target>, got %s", label)
	}
}

func TestGetProbeTarget_ConfigFileOverride(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("OMACORN_TEST_TARGET", "")

	cfgDir := filepath.Join(tmpDir, ".config", "omacorn")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatalf("failed to create mock config dir: %v", err)
	}

	targetFile := filepath.Join(cfgDir, "target.txt")
	if err := os.WriteFile(targetFile, []byte("https://my-private-site.org\n"), 0644); err != nil {
		t.Fatalf("failed to write target.txt: %v", err)
	}

	url, label := GetProbeTarget()
	if url != "https://my-private-site.org" {
		t.Errorf("expected url https://my-private-site.org, got %s", url)
	}
	if label != "<blocked-site>" {
		t.Errorf("expected sanitized label <blocked-site>, got %s", label)
	}
}

func TestGetProbeTarget_DefaultFallback(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("OMACORN_TEST_TARGET", "")

	url, label := GetProbeTarget()
	if !strings.HasPrefix(url, "https://") {
		t.Errorf("expected https url, got %s", url)
	}
	if label != "<test-target>" {
		t.Errorf("expected default label <test-target>, got %s", label)
	}
}

func TestRunCmd(t *testing.T) {
	out, err := RunCmd("echo", "hello", "omacorn")
	if err != nil {
		t.Fatalf("RunCmd failed: %v", err)
	}
	if strings.TrimSpace(out) != "hello omacorn" {
		t.Errorf("expected 'hello omacorn', got %q", out)
	}
}

func TestDetectHypervisor(t *testing.T) {
	isVirt, name := DetectHypervisor()
	// DetectHypervisor should not panic and return coherent strings
	if isVirt {
		if name == "" {
			t.Errorf("expected non-empty hypervisor name when isVirt is true")
		}
	}
}

func TestSystemStatusVirtualization(t *testing.T) {
	st := GetSystemStatus()
	isVirt, name := DetectHypervisor()
	if st.IsHypervisor != isVirt {
		t.Errorf("expected st.IsHypervisor == %v, got %v", isVirt, st.IsHypervisor)
	}
	if st.HypervisorName != name {
		t.Errorf("expected st.HypervisorName == %q, got %q", name, st.HypervisorName)
	}
}
