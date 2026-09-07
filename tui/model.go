package tui

import (
	"fmt"
	"time"

	"dpipe/engine"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

type probeResultMsg struct {
	googleStatus  int
	googleLatency time.Duration
	googleErr     string
	targetStatus  int
	targetLatency time.Duration
	targetErr     string
}

type Model struct {
	activeTab    int
	status       engine.SystemStatus
	logView      viewport.Model
	logsLoaded   bool
	notification string
	probing      bool
	width        int
	height       int
}

func NewModel() Model {
	m := Model{
		activeTab: 0,
		status:    engine.GetSystemStatus(),
		logView:   viewport.New(80, 18),
	}
	m.logView.Style = CardStyle
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		loadLogsCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadLogsCmd() tea.Cmd {
	return func() tea.Msg {
		out, _ := engine.RunCmd("journalctl", "--user", "-u", engine.SpoofUnitName, "-n", "30", "--no-pager")
		if out == "" {
			out, _ = engine.RunCmd("journalctl", "-u", engine.GecitUnitName, "-n", "30", "--no-pager")
		}
		return logsMsg(out)
	}
}

type logsMsg string

func runProbesCmd(status engine.SystemStatus) tea.Cmd {
	return func() tea.Msg {
		var proxyAddr string
		if status.SpoofActive {
			proxyAddr = engine.SpoofDefaultAddr
		}

		gStatus, gElapsed, gErr := engine.ProbeURL("https://accounts.google.com", proxyAddr, 4*time.Second)
		var gErrStr string
		if gErr != nil {
			gErrStr = gErr.Error()
		}

		probeURL, _ := engine.GetProbeTarget()
		tStatus, tElapsed, tErr := engine.ProbeURL(probeURL, proxyAddr, 4*time.Second)
		var tErrStr string
		if tErr != nil {
			tErrStr = tErr.Error()
		}

		return probeResultMsg{
			googleStatus:  gStatus,
			googleLatency: gElapsed,
			googleErr:     gErrStr,
			targetStatus:  tStatus,
			targetLatency: tElapsed,
			targetErr:     tErrStr,
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logView.Width = msg.Width - 4
		m.logView.Height = msg.Height - 10

	case tickMsg:
		m.status = engine.GetSystemStatus()
		cmds = append(cmds, tickCmd())

	case logsMsg:
		m.logView.SetContent(string(msg))
		m.logsLoaded = true

	case probeResultMsg:
		m.probing = false
		m.status.GoogleStatus = msg.googleStatus
		m.status.GoogleLatency = msg.googleLatency
		m.status.GoogleErr = msg.googleErr
		m.status.TargetStatus = msg.targetStatus
		m.status.TargetLatency = msg.targetLatency
		m.status.TargetErr = msg.targetErr
		m.notification = fmt.Sprintf("Probes complete: Google (%v), Blocked Target (%v)", msg.googleLatency.Round(time.Millisecond), msg.targetLatency.Round(time.Millisecond))

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			m.activeTab = (m.activeTab + 1) % 3
			if m.activeTab == 1 {
				cmds = append(cmds, loadLogsCmd())
			}

		case "1":
			m.activeTab = 0
		case "2":
			m.activeTab = 1
			cmds = append(cmds, loadLogsCmd())
		case "3":
			m.activeTab = 2

		case "s":
			m.notification = "Switching to SpoofDPI..."
			engine.StopGecit()
			_, err := engine.StartSpoofDPI()
			if err != nil {
				m.notification = fmt.Sprintf("SpoofDPI start error: %v", err)
			} else {
				m.notification = "Engine 1 (SpoofDPI) is ACTIVE."
			}
			m.status = engine.GetSystemStatus()

		case "g":
			m.notification = "Switching to Gecit (eBPF)..."
			engine.StopSpoofDPI()
			_, err := engine.StartGecit()
			if err != nil {
				m.notification = fmt.Sprintf("Gecit start error: %v", err)
			} else {
				m.notification = "Engine 2 (Gecit) is ACTIVE."
			}
			m.status = engine.GetSystemStatus()

		case "x":
			engine.StopSpoofDPI()
			engine.StopGecit()
			m.notification = "All bypass engines STOPPED."
			m.status = engine.GetSystemStatus()

		case "p":
			enabled, err := engine.ToggleSystemProxy()
			if err != nil {
				m.notification = fmt.Sprintf("Proxy toggle error: %v", err)
			} else if enabled {
				m.notification = "System proxy ENABLED in ~/.config/environment.d/"
			} else {
				m.notification = "System proxy DISABLED."
			}
			m.status = engine.GetSystemStatus()

		case "t":
			if !m.probing {
				m.probing = true
				m.notification = "Running live connectivity probes..."
				cmds = append(cmds, runProbesCmd(m.status))
			}

		case "c":
			out, err := engine.EmergencyCleanup()
			if err != nil {
				m.notification = fmt.Sprintf("Cleanup error: %v (%s)", err, out)
			} else {
				m.notification = "Emergency cleanup executed. System restored."
			}
			m.status = engine.GetSystemStatus()
		}
	}

	if m.activeTab == 1 {
		var viewCmd tea.Cmd
		m.logView, viewCmd = m.logView.Update(msg)
		cmds = append(cmds, viewCmd)
	}

	return m, tea.Batch(cmds...)
}
