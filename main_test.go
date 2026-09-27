package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"dpipe/engine"
)

func TestGetBinName(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tests := []struct {
		arg0     string
		expected string
	}{
		{"omacorn", "omacorn"},
		{"/usr/local/bin/omacorn", "omacorn"},
		{"./dpipe", "dpipe"},
		{"dpipe", "dpipe"},
		{"", "omacorn"},
		{".", "omacorn"},
	}

	for _, tc := range tests {
		os.Args = []string{tc.arg0}
		got := getBinName()
		if got != tc.expected {
			t.Errorf("for arg0 %q, expected %q, got %q", tc.arg0, tc.expected, got)
		}
	}
}

func TestVersionConstant(t *testing.T) {
	if version == "" {
		t.Errorf("version constant cannot be empty")
	}
	if !strings.HasPrefix(version, "0.") {
		t.Errorf("unexpected version format: %s", version)
	}
}

// cmdStubs keeps systemctl, sudo, curl, and flatpak off the live user session.
type cmdStubs struct {
	home string
	bin  string
}

func setupCmdStubs(t *testing.T) *cmdStubs {
	t.Helper()
	s := &cmdStubs{home: t.TempDir(), bin: t.TempDir()}
	t.Setenv("HOME", s.home)
	t.Setenv("PATH", s.bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	s.write(t, "sudo", `#!/bin/sh
if [ "${SUDO_FAIL}" = "1" ]; then echo denied >&2; exit 1; fi
echo sudo-ok
exit 0
`)
	s.write(t, "systemctl", `#!/bin/sh
if [ "${SYSTEMCTL_FAIL}" = "1" ]; then echo failed >&2; exit 1; fi
if [ "$1" = "--user" ]; then shift; fi
cmd="$1"
unit="$2"
if [ "$cmd" = "is-active" ]; then
  case "$unit" in
    dpipe-spoofdpi) echo "${SPOOF_STATE:-inactive}" ;;
    dpipe-gecit) echo "${GECIT_STATE:-inactive}" ;;
    *) echo inactive ;;
  esac
  exit 0
fi
echo "ActiveEnterTimestamp=Sun 2026-09-27 12:00:00 KST"
exit 0
`)
	s.write(t, "flatpak", "#!/bin/sh\nexit 0\n")
	s.write(t, "curl", `#!/bin/sh
if [ "${CURL_FAIL}" = "1" ]; then echo curl-fail >&2; exit 1; fi
out=""
prev=""
for a in "$@"; do
  if [ "$prev" = "-o" ]; then out="$a"; fi
  prev="$a"
done
if [ -n "${CURL_LOG}" ]; then printf '%s\n' "$*" >> "$CURL_LOG"; fi
if [ -n "$out" ]; then printf '#!/bin/sh\n' > "$out"; fi
exit 0
`)
	s.write(t, "uname", `#!/bin/sh
if [ -n "${UNAME_M}" ]; then echo "${UNAME_M}"; else echo x86_64; fi
exit 0
`)
	s.write(t, "ip", `#!/bin/sh
if [ "${IP_TUN}" = "1" ]; then echo "2: tun0: <POINTOPOINT,UP> mtu 1500"; exit 0; fi
echo "1: lo: <LOOPBACK,UP> mtu 65536"
exit 0
`)
	restoreProxyEnv(t)
	return s
}

func (s *cmdStubs) write(t *testing.T, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(s.bin, name), []byte(body), 0o755); err != nil {
		t.Fatalf("write stub %s: %v", name, err)
	}
}

func restoreProxyEnv(t *testing.T) {
	t.Helper()
	keys := []string{"http_proxy", "https_proxy", "all_proxy", "no_proxy"}
	prev := map[string]string{}
	present := map[string]bool{}
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			prev[k] = v
			present[k] = true
		}
	}
	t.Cleanup(func() {
		for _, k := range keys {
			if present[k] {
				_ = os.Setenv(k, prev[k])
			} else {
				_ = os.Unsetenv(k)
			}
		}
	})
}

func captureStdout(t *testing.T, fn func()) (out string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = buf.ReadFrom(r)
		close(done)
	}()
	defer func() {
		_ = w.Close()
		os.Stdout = orig
		<-done
		out = buf.String()
	}()
	fn()
	return
}

func writeProxyFile(t *testing.T, home string) {
	t.Helper()
	dir := filepath.Join(home, ".config", "environment.d")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "20-omacorn-proxy.conf")
	if err := os.WriteFile(path, []byte("http_proxy=http://127.0.0.1:8080\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func chdirEmpty(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
}

func helperEnv(home string) []string {
	env := make([]string, 0)
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "HOME=") || strings.HasPrefix(e, "OMACORN_WANT_HELPER=") {
			continue
		}
		env = append(env, e)
	}
	return append(env, "HOME="+home, "OMACORN_WANT_HELPER=1")
}

func TestHandleExcludeLifecycle(t *testing.T) {
	s := setupCmdStubs(t)

	out := captureStdout(t, func() { handleExclude(nil) })
	if !strings.Contains(out, "Custom Exclusions:  0 rules") {
		t.Fatalf("empty list:\n%s", out)
	}

	out = captureStdout(t, func() { handleExclude([]string{"add"}) })
	if !strings.Contains(out, "Usage: omacorn exclude add") {
		t.Fatalf("add usage:\n%s", out)
	}

	out = captureStdout(t, func() { handleExclude([]string{"add", "localhost"}) })
	if !strings.Contains(out, "already part of default exclusions") {
		t.Fatalf("add default:\n%s", out)
	}

	out = captureStdout(t, func() { handleExclude([]string{"add", "example.com"}) })
	if !strings.Contains(out, `Added "example.com"`) {
		t.Fatalf("add custom:\n%s", out)
	}

	out = captureStdout(t, func() { handleExclude([]string{"list"}) })
	if !strings.Contains(out, "example.com") || !strings.Contains(out, "custom") {
		t.Fatalf("list custom:\n%s", out)
	}

	out = captureStdout(t, func() { handleExclude([]string{"add", "example.com"}) })
	if !strings.Contains(out, "already in custom exclusions") {
		t.Fatalf("add duplicate:\n%s", out)
	}

	out = captureStdout(t, func() { handleExclude([]string{"remove"}) })
	if !strings.Contains(out, "Usage: omacorn exclude remove") {
		t.Fatalf("remove usage:\n%s", out)
	}

	out = captureStdout(t, func() { handleExclude([]string{"remove", "missing.example"}) })
	if !strings.Contains(out, "not found in custom exclusions") {
		t.Fatalf("remove missing:\n%s", out)
	}

	out = captureStdout(t, func() { handleExclude([]string{"rm", "example.com"}) })
	if !strings.Contains(out, `Removed "example.com"`) {
		t.Fatalf("remove:\n%s", out)
	}

	out = captureStdout(t, func() { handleExclude([]string{"reset"}) })
	if !strings.Contains(out, "Reset custom exclusions") {
		t.Fatalf("reset:\n%s", out)
	}

	out = captureStdout(t, func() { handleExclude([]string{"nope"}) })
	if !strings.Contains(out, "Unknown exclusion command: nope") {
		t.Fatalf("unknown:\n%s", out)
	}

	conf := filepath.Join(s.home, ".config", "omacorn", "exclusions.conf")
	if err := os.MkdirAll(filepath.Join(conf, "child"), 0o755); err != nil {
		t.Fatal(err)
	}
	out = captureStdout(t, func() { handleExclude([]string{"reset"}) })
	if !strings.Contains(out, "Error resetting exclusions") {
		t.Fatalf("reset error:\n%s", out)
	}
}

func TestHandleExcludeSyncsProxy(t *testing.T) {
	s := setupCmdStubs(t)
	writeProxyFile(t, s.home)

	out := captureStdout(t, func() { handleExclude([]string{"add", "example.org"}) })
	if !strings.Contains(out, "Synchronized updated exclusion") {
		t.Fatalf("add sync:\n%s", out)
	}
	out = captureStdout(t, func() { handleExclude([]string{"del", "example.org"}) })
	if !strings.Contains(out, "Synchronized updated exclusion") {
		t.Fatalf("remove sync:\n%s", out)
	}
	out = captureStdout(t, func() { handleExclude([]string{"reset"}) })
	if !strings.Contains(out, "Synchronized updated exclusion") {
		t.Fatalf("reset sync:\n%s", out)
	}
	if !engine.IsProxyConfigured() {
		t.Fatal("proxy file should remain configured after sync")
	}
}

func TestHandleStatusModes(t *testing.T) {
	s := setupCmdStubs(t)

	out := captureStdout(t, func() { handleStatus("--json") })
	if !strings.Contains(out, `"SpoofActive"`) {
		t.Fatalf("json status:\n%s", out)
	}

	out = captureStdout(t, func() { handleStatus("") })
	if !strings.Contains(out, "Flight Status Board") || !strings.Contains(out, "STANDBY") {
		t.Fatalf("text standby:\n%s", out)
	}
	if !strings.Contains(out, "DISABLED") {
		t.Fatalf("proxy should be disabled:\n%s", out)
	}
	if !strings.Contains(out, "inactive (Omacorn direct protection)") {
		t.Fatalf("vpn should be inactive:\n%s", out)
	}
	if !strings.Contains(out, "Virtualized (vmware") {
		t.Fatalf("this host is a vmware guest; virt line missing:\n%s", out)
	}

	t.Setenv("SPOOF_STATE", "active")
	writeProxyFile(t, s.home)
	out = captureStdout(t, func() { handleStatus("") })
	if !strings.Contains(out, "ACTIVE (127.0.0.1:8080) [IN COMMAND]") {
		t.Fatalf("spoof active:\n%s", out)
	}
	if !strings.Contains(out, "ENABLED (~/.config/environment.d)") {
		t.Fatalf("proxy enabled:\n%s", out)
	}

	t.Setenv("SPOOF_STATE", "inactive")
	t.Setenv("GECIT_STATE", "active")
	out = captureStdout(t, func() { handleStatus("") })
	if !strings.Contains(out, "ACTIVE (eBPF sock_ops) [IN COMMAND]") {
		t.Fatalf("gecit active:\n%s", out)
	}

	t.Setenv("SPOOF_STATE", "active")
	out = captureStdout(t, func() { handleStatus("") })
	if !strings.Contains(out, "CONFLICT") {
		t.Fatalf("dual pilot:\n%s", out)
	}

	t.Setenv("IP_TUN", "1")
	out = captureStdout(t, func() { handleStatus("") })
	if !strings.Contains(out, "ACTIVE (tun0 - all traffic tunneled)") {
		t.Fatalf("vpn active:\n%s", out)
	}
}

func TestHandleStartPilots(t *testing.T) {
	setupCmdStubs(t)

	t.Setenv("SPOOF_STATE", "active")
	out := captureStdout(t, func() { handleStart("gecit") })
	if !strings.Contains(out, "WARNING: Running inside VMWARE") {
		t.Fatalf("gecit virt warning:\n%s", out)
	}
	if !strings.Contains(out, "Disengaging Main Pilot") {
		t.Fatalf("gecit handoff:\n%s", out)
	}
	if !strings.Contains(out, "Co-Pilot in command") {
		t.Fatalf("gecit success:\n%s", out)
	}

	t.Setenv("SUDO_FAIL", "1")
	out = captureStdout(t, func() { handleStart("gecit") })
	if !strings.Contains(out, "[gecit] Start failed:") {
		t.Fatalf("gecit failure:\n%s", out)
	}

	t.Setenv("SUDO_FAIL", "")
	t.Setenv("SPOOF_STATE", "inactive")
	t.Setenv("GECIT_STATE", "active")
	out = captureStdout(t, func() { handleStart("spoofdpi") })
	if !strings.Contains(out, "Disengaging Co-Pilot") {
		t.Fatalf("spoof handoff:\n%s", out)
	}
	if !strings.Contains(out, "Main Pilot in command on "+engine.SpoofDefaultAddr) {
		t.Fatalf("spoof success:\n%s", out)
	}
	if !strings.Contains(out, "Hypervisor-safe user-space proxy mode active") {
		t.Fatalf("spoof virt line:\n%s", out)
	}

	t.Setenv("SYSTEMCTL_FAIL", "1")
	out = captureStdout(t, func() { handleStart("") })
	if !strings.Contains(out, "[spoofdpi] Start failed:") {
		t.Fatalf("spoof failure:\n%s", out)
	}
}

func TestHandleUpdateScriptSuccess(t *testing.T) {
	setupCmdStubs(t)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\necho script-ran\nexit 0\n"
	if err := os.WriteFile(filepath.Join(dir, "scripts", "update.sh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	out := captureStdout(t, func() { handleUpdate() })
	if !strings.Contains(out, "Fleet Maintenance") || !strings.Contains(out, "script-ran") {
		t.Fatalf("script success:\n%s", out)
	}
}

func TestHandleUpdateScriptExit(t *testing.T) {
	if os.Getenv("OMACORN_WANT_HELPER") == "1" {
		handleUpdate()
		return
	}
	dir := t.TempDir()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "#!/bin/sh\nexit 7\n"
	if err := os.WriteFile(filepath.Join(dir, "scripts", "update.sh"), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleUpdateScriptExit$")
	cmd.Dir = dir
	cmd.Env = helperEnv(home)
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 7 {
		t.Fatalf("expected exit 7, got %v", err)
	}
}

func TestHandleUpdateScriptNotExecutable(t *testing.T) {
	if os.Getenv("OMACORN_WANT_HELPER") == "1" {
		handleUpdate()
		return
	}
	dir := t.TempDir()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(filepath.Join(dir, "scripts", "update.sh"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleUpdateScriptNotExecutable$")
	cmd.Dir = dir
	cmd.Env = helperEnv(home)
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit 1, got %v", err)
	}
}

func TestHandleUpdateCurlPaths(t *testing.T) {
	s := setupCmdStubs(t)
	chdirEmpty(t)

	t.Setenv("CURL_FAIL", "1")
	out := captureStdout(t, func() { handleUpdate() })
	if !strings.Contains(out, "Git repository clone not found") || !strings.Contains(out, "Failed to download") {
		t.Fatalf("curl fail:\n%s", out)
	}

	t.Setenv("CURL_FAIL", "")
	t.Setenv("UNAME_M", "aarch64")
	t.Setenv("SPOOF_STATE", "active")
	logPath := filepath.Join(t.TempDir(), "curl.log")
	t.Setenv("CURL_LOG", logPath)
	writeProxyFile(t, s.home)
	out = captureStdout(t, func() { handleUpdate() })
	if !strings.Contains(out, "Successfully updated Omacorn static binary!") {
		t.Fatalf("curl success:\n%s", out)
	}
	if !strings.Contains(out, "Hot-reloaded dpipe-spoofdpi") || !strings.Contains(out, "Re-synchronized desktop proxy") {
		t.Fatalf("curl reload/sync:\n%s", out)
	}
	bin := filepath.Join(s.home, ".local", "bin", "omacorn")
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("installed binary missing: %v", err)
	}
	logged, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logged), "omacorn-linux-arm64") {
		t.Fatalf("curl log: %s", logged)
	}

	if err := os.Remove(bin); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	out = captureStdout(t, func() { handleUpdate() })
	if !strings.Contains(out, "Failed to replace binary") {
		t.Fatalf("rename fail:\n%s", out)
	}
}
