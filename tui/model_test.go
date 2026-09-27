package tui

import (
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
