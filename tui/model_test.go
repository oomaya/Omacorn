package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dpipe/engine"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/golden"
)

func newTestModel() Model {
	m := NewModel()
	m.width = 80
	m.height = 24
	m.status = engine.SystemStatus{
		SpoofActive:    true,
		SpoofAddr:      engine.SpoofDefaultAddr,
		SpoofUptime:    "Sun 2026-09-27 12:00:00 KST",
		GecitActive:    false,
		GecitUptime:    "",
		IsHypervisor:   false,
		HypervisorName: "",
		ProxyEnabled:   true,
		DNSProtected:   true,
		DNSDetails:     "systemd-resolved (protected)",
		VPNActive:      false,
		VPNInterface:   "",
		GoogleStatus:   200,
		GoogleLatency:  45 * time.Millisecond,
		GoogleErr:      "",
		TargetStatus:   200,
		TargetLatency:  82 * time.Millisecond,
		TargetErr:      "",
	}
	m.logView.Width = 76
	m.logView.Height = 14
	m.logView.SetContent("line 1: pilot initialized\nline 2: connected\nline 3: ready")
	m.logsLoaded = true
	m.notification = "Test notification active"
	return m
}

func TestViewGolden(t *testing.T) {
	m := newTestModel()
	m.activeTab = 0
	golden.RequireEqual(t, []byte(m.View()))
}

func TestRenderDashboardGolden(t *testing.T) {
	m := newTestModel()
	golden.RequireEqual(t, []byte(m.renderDashboard()))
}

func TestRenderLogsGolden(t *testing.T) {
	m := newTestModel()
	golden.RequireEqual(t, []byte(m.renderLogs()))
}

func TestRenderDiagnosticsGolden(t *testing.T) {
	m := newTestModel()
	golden.RequireEqual(t, []byte(m.renderDiagnostics()))
}

func TestUpdateMessages(t *testing.T) {
	m := newTestModel()

	// 1. WindowSizeMsg
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	newM, _ := m.Update(sizeMsg)
	updated := newM.(Model)
	if updated.width != 100 || updated.height != 30 {
		t.Errorf("expected dimensions 100x30, got %dx%d", updated.width, updated.height)
	}

	// 2. LogsMsg
	logs := logsMsg("sample test log entry\nsecond line")
	newM, _ = updated.Update(logs)
	updated = newM.(Model)
	if !updated.logsLoaded {
		t.Errorf("expected logsLoaded to be true")
	}

	// 3. ProbeResultMsg
	probeMsg := probeResultMsg{
		googleStatus:  200,
		googleLatency: 50 * time.Millisecond,
		googleErr:     "",
		targetStatus:  204,
		targetLatency: 120 * time.Millisecond,
		targetErr:     "",
	}
	newM, _ = updated.Update(probeMsg)
	updated = newM.(Model)
	if updated.status.GoogleStatus != 200 || updated.status.TargetStatus != 204 {
		t.Errorf("unexpected status after probeResultMsg: %+v", updated.status)
	}
	if !strings.Contains(updated.notification, "Probes complete") {
		t.Errorf("expected probe complete notification, got: %s", updated.notification)
	}

	// 4. Tab navigation keys
	tabKeys := []struct {
		key         string
		expectedTab int
	}{
		{"1", 0},
		{"2", 1},
		{"3", 2},
		{"tab", 0}, // from 2 wraps back to 0
	}
	for _, tc := range tabKeys {
		newM, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.key)})
		updated = newM.(Model)
		if updated.activeTab != tc.expectedTab {
			t.Errorf("key %q expected tab %d, got %d", tc.key, tc.expectedTab, updated.activeTab)
		}
	}

	// 5. Probe trigger key ('t')
	newM, cmd := updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	updated = newM.(Model)
	if !updated.probing || cmd == nil {
		t.Errorf("expected probing to be true and cmd returned on 't'")
	}

	// 6. TickMsg
	newM, cmd = updated.Update(tickMsg(time.Now()))
	updated = newM.(Model)
	if cmd == nil {
		t.Errorf("expected tickCmd returned on tickMsg")
	}

	// 7. Quit keys
	newM, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Errorf("expected quit cmd on 'q'")
	}
}

func TestModelInit(t *testing.T) {
	m := newTestModel()

	cmd := m.Init()
	if cmd == nil {
		t.Errorf("expected Init() to return batch command")
	}
}

// isolatePilotCommands puts stub systemctl/sudo/flatpak binaries ahead of PATH
// and points HOME at a temp dir. Update keys must not touch the live user session.
func isolatePilotCommands(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	script := []byte(`#!/bin/sh
name=$(basename "$0")
case ",${STUB_FAIL}," in
*,"$name",*) echo "stub-fail" >&2; exit 1 ;;
esac
echo "stub-ok"
exit 0
`)
	for _, name := range []string{"systemctl", "sudo", "journalctl", "flatpak", "gecit", "getcap"} {
		if err := os.WriteFile(filepath.Join(bin, name), script, 0o755); err != nil {
			t.Fatalf("write stub %s: %v", name, err)
		}
	}
	t.Setenv("STUB_FAIL", "")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
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

func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestUpdateUncoveredBranches(t *testing.T) {
	isolatePilotCommands(t)
	m := newTestModel()

	got, cmd := m.Update(nil)
	if _, ok := got.(Model); !ok {
		t.Fatalf("nil message: expected Model, got %T", got)
	}
	if cmd != nil {
		t.Fatal("nil message on dashboard: expected no command")
	}

	type unknownMsg struct{}
	got, cmd = m.Update(unknownMsg{})
	if _, ok := got.(Model); !ok {
		t.Fatalf("unknown message: expected Model, got %T", got)
	}
	if cmd != nil {
		t.Fatal("unknown message on dashboard: expected no command")
	}

	m.activeTab = 1
	got, _ = m.Update(nil)
	if got.(Model).activeTab != 1 {
		t.Fatal("nil message on logs tab should keep the tab")
	}
	m.activeTab = 0

	got, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c: expected quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c: expected QuitMsg, got %T", cmd())
	}

	got, cmd = m.Update(keyRunes("tab"))
	updated := got.(Model)
	if updated.activeTab != 1 {
		t.Fatalf("tab from dashboard: expected logs tab, got %d", updated.activeTab)
	}
	if cmd == nil {
		t.Fatal("tab onto logs: expected loadLogsCmd")
	}

	updated.probing = true
	updated.activeTab = 0
	before := updated.notification
	got, cmd = updated.Update(keyRunes("t"))
	updated = got.(Model)
	if !updated.probing {
		t.Fatal("second t: probing must stay latched")
	}
	if updated.notification != before {
		t.Fatalf("second t: notification changed to %q", updated.notification)
	}
	if cmd != nil {
		t.Fatal("second t: expected no new probe command")
	}

	got, _ = updated.Update(keyRunes("s"))
	updated = got.(Model)
	if updated.notification != "Pilot 1 (SpoofDPI) is IN COMMAND." {
		t.Fatalf("s success: %q", updated.notification)
	}

	updated.status.IsHypervisor = false
	updated.status.HypervisorName = ""
	got, _ = updated.Update(keyRunes("g"))
	updated = got.(Model)
	if updated.notification != "Pilot 2 (Gecit) is IN COMMAND." {
		t.Fatalf("g bare-metal success: %q", updated.notification)
	}

	t.Setenv("STUB_FAIL", "sudo")
	updated.status.IsHypervisor = true
	updated.status.HypervisorName = "vmware"
	got, _ = updated.Update(keyRunes("g"))
	updated = got.(Model)
	if !strings.Contains(updated.notification, "Gecit start error") {
		t.Fatalf("g error: %q", updated.notification)
	}

	t.Setenv("STUB_FAIL", "")
	got, _ = updated.Update(keyRunes("x"))
	updated = got.(Model)
	if updated.notification != "All bypass pilots DISENGAGED." {
		t.Fatalf("x: %q", updated.notification)
	}

	got, _ = updated.Update(keyRunes("p"))
	updated = got.(Model)
	if updated.notification != "System proxy ENABLED in ~/.config/environment.d/" {
		t.Fatalf("p enable: %q", updated.notification)
	}
	got, _ = updated.Update(keyRunes("p"))
	updated = got.(Model)
	if updated.notification != "System proxy DISABLED." {
		t.Fatalf("p disable: %q", updated.notification)
	}

	homeFile := filepath.Join(t.TempDir(), "not-a-home")
	if err := os.WriteFile(homeFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", homeFile)
	got, _ = updated.Update(keyRunes("p"))
	updated = got.(Model)
	if !strings.Contains(updated.notification, "Proxy toggle error") {
		t.Fatalf("p error: %q", updated.notification)
	}

	t.Setenv("STUB_FAIL", "")
	got, _ = updated.Update(keyRunes("c"))
	updated = got.(Model)
	if updated.notification != "Emergency cleanup executed. System restored." {
		t.Fatalf("c success: %q", updated.notification)
	}

	t.Setenv("STUB_FAIL", "sudo")
	got, _ = updated.Update(keyRunes("c"))
	updated = got.(Model)
	if !strings.Contains(updated.notification, "Cleanup error") {
		t.Fatalf("c error: %q", updated.notification)
	}

	t.Setenv("STUB_FAIL", "systemctl")
	got, _ = updated.Update(keyRunes("s"))
	updated = got.(Model)
	if !strings.Contains(updated.notification, "SpoofDPI start error") {
		t.Fatalf("s error: %q", updated.notification)
	}
}
