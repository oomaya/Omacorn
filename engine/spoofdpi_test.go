package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeExec(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

func stubInstallPATH(t *testing.T, home string, getcapBody string) string {
	t.Helper()
	bin := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)
	writeExec(t, filepath.Join(bin, "systemctl"), "#!/bin/sh\necho ok\nexit 0\n")
	writeExec(t, filepath.Join(bin, "getcap"), getcapBody)
	return bin
}

func TestInstallSpoofDPINotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	err := InstallSpoofDPI()
	if err == nil || !strings.Contains(err.Error(), "spoofdpi binary not found") {
		t.Fatalf("expected missing binary, got %v", err)
	}
}

func TestInstallSpoofDPIWritesUserUnit(t *testing.T) {
	home := t.TempDir()
	bin := stubInstallPATH(t, home, "#!/bin/sh\necho 'cap_net_raw=ep'\nexit 0\n")
	writeExec(t, filepath.Join(bin, "spoofdpi"), "#!/bin/sh\nexit 0\n")

	if err := InstallSpoofDPI(); err != nil {
		t.Fatal(err)
	}
	unitPath := filepath.Join(home, ".config", "systemd", "user", SpoofUnitName+".service")
	data, err := os.ReadFile(unitPath)
	if err != nil {
		t.Fatal(err)
	}
	unit := string(data)
	if !strings.Contains(unit, filepath.Join(bin, "spoofdpi")) {
		t.Fatalf("unit missing binary path:\n%s", unit)
	}
	// This host reports a hypervisor, so the raw-cap arm stays inactive.
	if !strings.Contains(unit, "--https-fake-count 0") {
		t.Fatalf("expected user-space fragmentation unit:\n%s", unit)
	}
	if strings.Contains(unit, "--https-fake-count 1") {
		t.Fatalf("virt host must not enable decoy injection:\n%s", unit)
	}
}

func TestInstallSpoofDPIGetcapFails(t *testing.T) {
	home := t.TempDir()
	bin := stubInstallPATH(t, home, "#!/bin/sh\necho no-cap >&2\nexit 1\n")
	writeExec(t, filepath.Join(bin, "spoofdpi"), "#!/bin/sh\nexit 0\n")
	if err := InstallSpoofDPI(); err != nil {
		t.Fatal(err)
	}
	unitPath := filepath.Join(home, ".config", "systemd", "user", SpoofUnitName+".service")
	data, err := os.ReadFile(unitPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "--https-fake-count 0") {
		t.Fatalf("unit:\n%s", data)
	}
}

func TestInstallSpoofDPIFromHomeGoBin(t *testing.T) {
	home := t.TempDir()
	stubInstallPATH(t, home, "#!/bin/sh\nexit 1\n")
	goBin := filepath.Join(home, "go", "bin", "spoofdpi")
	writeExec(t, goBin, "#!/bin/sh\nexit 0\n")
	if err := InstallSpoofDPI(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "systemd", "user", SpoofUnitName+".service"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), goBin) {
		t.Fatalf("unit missing home go bin:\n%s", data)
	}
}

func TestInstallSpoofDPIMkdirFails(t *testing.T) {
	home := t.TempDir()
	bin := stubInstallPATH(t, home, "#!/bin/sh\nexit 0\n")
	writeExec(t, filepath.Join(bin, "spoofdpi"), "#!/bin/sh\nexit 0\n")
	if err := os.WriteFile(filepath.Join(home, ".config"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := InstallSpoofDPI()
	if err == nil {
		t.Fatal("expected MkdirAll failure")
	}
}

func TestInstallSpoofDPIWriteFails(t *testing.T) {
	home := t.TempDir()
	bin := stubInstallPATH(t, home, "#!/bin/sh\nexit 0\n")
	writeExec(t, filepath.Join(bin, "spoofdpi"), "#!/bin/sh\nexit 0\n")
	dir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	err := InstallSpoofDPI()
	if err == nil {
		t.Fatal("expected unit write failure")
	}
}
