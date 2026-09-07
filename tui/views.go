package tui

import (
	"fmt"
	"strings"
	"time"

	"dpipe/engine"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var s strings.Builder

	// 1. Header
	headerText := TitleStyle.Render("OMACORN DESKTOP (OD) v0.4.0")
	sysInfo := lipgloss.NewStyle().Foreground(ColorMuted).Render("Arch/Omarchy • Kernel 7.1.9 • Twin-Engine")
	headerBar := lipgloss.JoinHorizontal(lipgloss.Center, headerText, "  ", sysInfo)
	s.WriteString(headerBar)
	s.WriteByte('\n')

	// 2. Tabs
	var tabs []string
	titles := []string{"[1] Dashboard", "[2] Live Logs", "[3] Diagnostics"}
	for i, t := range titles {
		if m.activeTab == i {
			tabs = append(tabs, ActiveTabStyle.Render(t))
		} else {
			tabs = append(tabs, TabStyle.Render(t))
		}
	}
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	s.WriteString(tabRow)
	s.WriteString("\n\n")

	// 3. Main Content
	switch m.activeTab {
	case 0:
		s.WriteString(m.renderDashboard())
	case 1:
		s.WriteString(m.renderLogs())
	case 2:
		s.WriteString(m.renderDiagnostics())
	}

	// 4. Notification Banner
	if m.notification != "" {
		notif := lipgloss.NewStyle().
			Foreground(ColorCyan).
			Background(lipgloss.Color("#1e272e")).
			Padding(0, 1).
			Render("✦ " + m.notification)
		s.WriteByte('\n')
		s.WriteString(notif)
		s.WriteByte('\n')
	} else {
		s.WriteString("\n\n")
	}

	// 5. Footer Shortcuts
	keys := []string{
		HelpKeyStyle.Render("[1-3]") + HelpDescStyle.Render(" Tabs"),
		HelpKeyStyle.Render("[S]") + HelpDescStyle.Render(" SpoofDPI"),
		HelpKeyStyle.Render("[G]") + HelpDescStyle.Render(" Gecit"),
		HelpKeyStyle.Render("[X]") + HelpDescStyle.Render(" Stop All"),
		HelpKeyStyle.Render("[P]") + HelpDescStyle.Render(" Toggle Proxy"),
		HelpKeyStyle.Render("[T]") + HelpDescStyle.Render(" Probe Test"),
		HelpKeyStyle.Render("[C]") + HelpDescStyle.Render(" Cleanup"),
		HelpKeyStyle.Render("[Q]") + HelpDescStyle.Render(" Quit"),
	}
	s.WriteString(strings.Join(keys, "  "))
	s.WriteByte('\n')

	return s.String()
}

func (m Model) renderDashboard() string {
	// Pilot 1: SpoofDPI (Main Pilot) Card
	spoofBadge := BadgeInactive.Render("STANDBY")
	cardStyle1 := CardStyle
	if m.status.SpoofActive {
		spoofBadge = BadgeActive.Render("ACTIVE (IN COMMAND)")
		cardStyle1 = ActiveCardStyle
	}
	content1 := fmt.Sprintf(
		"%s  %s\nAddress: %s\nUptime:  %s\nRole:    Main Pilot • User Space ($XDG_RUNTIME_DIR) • Safe for VMs & Containers",
		lipgloss.NewStyle().Bold(true).Foreground(ColorWhite).Render("Pilot 1: SpoofDPI (Main Pilot)"),
		spoofBadge,
		m.status.SpoofAddr,
		m.status.SpoofUptime,
	)
	box1 := cardStyle1.Width(74).Render(content1)

	// Pilot 2: Gecit (Co-Pilot / Standby) Card
	gecitBadge := BadgeInactive.Render("STANDBY")
	cardStyle2 := CardStyle
	if m.status.GecitActive {
		gecitBadge = BadgeActive.Render("ACTIVE (IN COMMAND)")
		cardStyle2 = ActiveCardStyle
	}
	if m.status.SpoofActive && m.status.GecitActive {
		gecitBadge = lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Render("⚠️ DUAL PILOT CONFLICT")
	}
	gecitRole := "Co-Pilot / Standby • Kernel Root Cgroup (eBPF sock_ops) • High Speed for Bare Metal"
	if m.status.IsHypervisor {
		gecitRole = fmt.Sprintf("Co-Pilot (Standby) • ⚠️ %s Guest (eBPF packet crafting can cause hypervisor DMA faults)", strings.ToUpper(m.status.HypervisorName))
	}
	content2 := fmt.Sprintf(
		"%s  %s\nMode:    Transparent System-Wide (eBPF sock_ops)\nUptime:  %s\nRole:    %s",
		lipgloss.NewStyle().Bold(true).Foreground(ColorWhite).Render("Pilot 2: Gecit (Co-Pilot)"),
		gecitBadge,
		m.status.GecitUptime,
		gecitRole,
	)
	box2 := cardStyle2.Width(74).Render(content2)

	// System Proxy Status Card
	proxyBadge := BadgeInactive.Render("DISABLED")
	if m.status.ProxyEnabled {
		proxyBadge = BadgeActive.Render("ENABLED (~/.config/environment.d)")
	}
	vpnLine := lipgloss.NewStyle().Foreground(ColorMuted).Render("Inactive (Omacorn direct protection)")
	if m.status.VPNActive {
		vpnLine = lipgloss.NewStyle().Foreground(ColorCyan).Render(fmt.Sprintf("ACTIVE (%s - all traffic tunneled)", m.status.VPNInterface))
	}

	content3 := fmt.Sprintf(
		"%s  %s\nPath:    ~/.config/environment.d/20-omacorn-proxy.conf\nTargets: Native flags synced for Brave (all variants), Chromium, Vivaldi, Firefox\nVPN:     %s\nBypass:  accounts.google.com, localhost, RFC1918 strictly exempted",
		lipgloss.NewStyle().Bold(true).Foreground(ColorWhite).Render("Desktop System Proxy:"),
		proxyBadge,
		vpnLine,
	)
	box3 := CardStyle.Width(74).Render(content3)

	// Live Connectivity Metrics Card
	gStr := "Not tested yet (Press 'T' to probe)"
	if m.status.GoogleStatus != 0 || m.status.GoogleErr != "" {
		if m.status.GoogleErr != "" {
			gStr = lipgloss.NewStyle().Foreground(ColorRed).Render("FAIL (" + m.status.GoogleErr + ")")
		} else {
			gStr = lipgloss.NewStyle().Foreground(ColorGreen).Render(fmt.Sprintf("HTTP %d in %v (TLS 1.3 Intact)", m.status.GoogleStatus, m.status.GoogleLatency.Round(time.Millisecond)))
		}
	}

	tStr := "Not tested yet (Press 'T' to probe)"
	if m.status.TargetStatus != 0 || m.status.TargetErr != "" {
		if m.status.TargetErr != "" {
			tStr = lipgloss.NewStyle().Foreground(ColorRed).Render("BLOCKED (" + m.status.TargetErr + ")")
		} else {
			tStr = lipgloss.NewStyle().Foreground(ColorGreen).Render(fmt.Sprintf("UNBLOCKED: HTTP %d in %v", m.status.TargetStatus, m.status.TargetLatency.Round(time.Millisecond)))
		}
	}

	_, displayLabel := engine.GetProbeTarget()
	content4 := fmt.Sprintf(
		"%s\n• Google Auth (accounts.google.com): %s\n• Blocked Target (%s): %s",
		lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("Live Network Probes:"),
		gStr,
		displayLabel,
		tStr,
	)
	box4 := CardStyle.Width(74).Render(content4)

	return lipgloss.JoinVertical(lipgloss.Left, box1, box2, box3, box4)
}

func (m Model) renderLogs() string {
	header := lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("Real-Time Journal Logs (Scroll with Up/Down):")
	return header + "\n" + m.logView.View()
}

func (m Model) renderDiagnostics() string {
	dnsStatus := lipgloss.NewStyle().Foreground(ColorGreen).Render("PASS: " + m.status.DNSDetails)
	if !m.status.DNSProtected {
		dnsStatus = lipgloss.NewStyle().Foreground(ColorOrange).Render("WARN: " + m.status.DNSDetails)
	}

	exclusionsCount := len(engine.GetAllExclusions())
	splitStatus := lipgloss.NewStyle().Foreground(ColorGreen).Render(fmt.Sprintf("ACTIVE (%d rules, Google Auth & Telegram MTProto bypassed)", exclusionsCount))

	content := fmt.Sprintf(
		"%s\n\n"+
			"1. DNS Protection:\n"+
			"   %s\n\n"+
			"2. Kernel eBPF & Capability Posture:\n"+
			"   • SpoofDPI: Has CAP_NET_RAW (Enables low-TTL decoy packet injection as unprivileged user)\n"+
			"   • Gecit:    Has CAP_BPF, CAP_PERFMON, CAP_SYS_ADMIN (Allows BPF ring buffers)\n\n"+
			"3. Split-Tunneling & Protocol Exclusions:\n"+
			"   • %s\n"+
			"   • Telegram MTProto: Direct to Origin (all_proxy omitted; raw TCP preserved)\n"+
			"   • CLI Management: 'omacorn exclude list' or 'omacorn run-clean <app>'\n\n"+
			"4. Hypervisor Safety Guardrails:\n"+
			"   • Handshake MSS restore enforced at 600 bytes (Protects VMware SVGA DMA range)\n"+
			"   • --doh=false enforced (Protects /etc/resolv.conf from tampering)\n"+
			"   • ExecStopPost auto-cleanup traps enabled on service teardown\n\n"+
			"5. Emergency System Recovery:\n"+
			"   Press 'C' at any time to instantly purge all eBPF hooks and stop all daemons.",
		lipgloss.NewStyle().Bold(true).Foreground(ColorCyan).Render("System Diagnostics & Safety Guardrails"),
		dnsStatus,
		splitStatus,
	)

	return CardStyle.Width(74).Render(content)
}
