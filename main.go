package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dpipe/engine"
	"dpipe/tui"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.3.0"

func main() {
	// If no arguments provided, launch the interactive Bubble Tea TUI
	if len(os.Args) < 2 {
		launchTUI()
		return
	}

	command := os.Args[1]
	arg := ""
	if len(os.Args) >= 3 {
		arg = strings.ToLower(os.Args[2])
	}

	switch command {
	case "tui", "ui":
		launchTUI()
	case "install":
		handleInstall(arg)
	case "start":
		handleStart(arg)
	case "stop":
		handleStop(arg)
	case "status":
		handleStatus(arg)
	case "proxy":
		handleProxy(arg)
	case "test":
		handleTest(arg)
	case "logs":
		handleLogs(arg)
	case "cleanup":
		handleCleanup()
	case "version", "-v", "--version":
		fmt.Printf("Omacorn v%s - Sovereign Twin-Engine DPI Bypass (command: %s)\n", version, getBinName())
	default:
		printUsage()
	}
}

func getBinName() string {
	bin := filepath.Base(os.Args[0])
	if bin == "" || bin == "." {
		return "omacorn"
	}
	return bin
}

func launchTUI() {
	m := tui.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running OD TUI: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	bin := getBinName()
	fmt.Printf(`🦄 Omacorn v%s — Sovereign Twin-Engine DPI Bypass Controller

Usage:
  %s                              Launch interactive Cyberpunk TUI dashboard
  %s start   [spoofdpi|gecit]     Start bypass engine (default: spoofdpi)
  %s stop    [spoofdpi|gecit|all] Stop bypass engine(s)
  %s status  [--json]             Show status board (or JSON for Quickshell/Waybar)
  %s test    [spoofdpi|gecit]     Run live latency & handshake probe
  %s proxy   [on|off|toggle]      Manage desktop environment proxy (~/.config/environment.d)
  %s logs    [spoofdpi|gecit]     Tail journal logs
  %s install [spoofdpi|gecit|all] Install systemd service(s)
  %s cleanup                      Emergency cleanup for eBPF and services
`, version, bin, bin, bin, bin, bin, bin, bin, bin, bin)
}

func handleInstall(target string) {
	switch target {
	case "spoofdpi":
		if err := engine.InstallSpoofDPI(); err != nil {
			fmt.Printf("[spoofdpi] ERROR: %v\n", err)
			return
		}
		fmt.Println("[spoofdpi] Installed user unit to ~/.config/systemd/user/dpipe-spoofdpi.service")
	case "gecit":
		if err := engine.InstallGecit(); err != nil {
			fmt.Printf("[gecit] ERROR: %v\n", err)
			return
		}
		fmt.Println("[gecit] Installed hardened unit to /etc/systemd/system/dpipe-gecit.service")
	default:
		_ = engine.InstallSpoofDPI()
		fmt.Println("[spoofdpi] Installed user unit.")
		_ = engine.InstallGecit()
		fmt.Println("[gecit] Installed system unit.")
	}
}

func handleStart(target string) {
	if target == "gecit" {
		out, err := engine.StartGecit()
		if err != nil {
			fmt.Printf("[gecit] Start failed: %s (%v)\n", out, err)
			return
		}
		fmt.Println("[gecit] Started (eBPF sock_ops attached).")
		return
	}

	// Default to spoofdpi
	out, err := engine.StartSpoofDPI()
	if err != nil {
		fmt.Printf("[spoofdpi] Start failed: %s (%v)\n", out, err)
		return
	}
	fmt.Printf("[spoofdpi] Started on %s\n", engine.SpoofDefaultAddr)
}

func handleStop(target string) {
	if target == "spoofdpi" || target == "all" || target == "" {
		engine.StopSpoofDPI()
		fmt.Println("[spoofdpi] Stopped.")
	}
	if target == "gecit" || target == "all" || target == "" {
		engine.StopGecit()
		fmt.Println("[gecit] Stopped.")
	}
}

func handleStatus(arg string) {
	st := engine.GetSystemStatus()

	if arg == "--json" {
		b, _ := json.MarshalIndent(st, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Println("================================================================")
	fmt.Printf("           🦄 OMACORN v%s - Status\n", version)
	fmt.Println("================================================================")
	spoofState := "inactive"
	if st.SpoofActive {
		spoofState = "ACTIVE (127.0.0.1:8080)"
	}
	gecitState := "inactive"
	if st.GecitActive {
		gecitState = "ACTIVE (eBPF sock_ops)"
	}
	proxyState := "DISABLED"
	if st.ProxyEnabled {
		proxyState = "ENABLED (~/.config/environment.d)"
	}

	vpnState := "inactive (Omacorn direct protection)"
	if st.VPNActive {
		vpnState = fmt.Sprintf("ACTIVE (%s - all traffic tunneled)", st.VPNInterface)
	}

	fmt.Printf("Engine 1: [spoofdpi]    %s\n", spoofState)
	fmt.Printf("Engine 2: [gecit]       %s\n", gecitState)
	fmt.Printf("Desktop System Proxy:   %s\n", proxyState)
	fmt.Printf("VPN Tunnel:             %s\n", vpnState)
	fmt.Printf("System DNS:             %s\n", st.DNSDetails)
	fmt.Println("================================================================")
}

func handleProxy(action string) {
	switch action {
	case "on", "enable":
		if err := engine.EnableSystemProxy(); err != nil {
			fmt.Printf("Error enabling system proxy: %v\n", err)
			return
		}
		fmt.Println("System proxy ENABLED in ~/.config/environment.d/20-omacorn-proxy.conf")
	case "off", "disable":
		if err := engine.DisableSystemProxy(); err != nil {
			fmt.Printf("Error disabling system proxy: %v\n", err)
			return
		}
		fmt.Println("System proxy DISABLED.")
	case "toggle", "":
		enabled, err := engine.ToggleSystemProxy()
		if err != nil {
			fmt.Printf("Toggle error: %v\n", err)
			return
		}
		if enabled {
			fmt.Println("System proxy ENABLED.")
		} else {
			fmt.Println("System proxy DISABLED.")
		}
	}
}

func handleTest(_ string) {
	fmt.Println("Running live probe test...")
	st := engine.GetSystemStatus()
	var proxyAddr string
	if st.SpoofActive {
		proxyAddr = engine.SpoofDefaultAddr
	}

	gCode, gTime, gErr := engine.ProbeURL("https://accounts.google.com", proxyAddr, 4*time.Second)
	if gErr != nil {
		fmt.Printf("• Google Auth: FAIL (%v)\n", gErr)
	} else {
		fmt.Printf("• Google Auth: HTTP %d in %v (TLS 1.3 Intact)\n", gCode, gTime)
	}

	probeURL, displayLabel := engine.GetProbeTarget()
	tCode, tTime, tErr := engine.ProbeURL(probeURL, proxyAddr, 4*time.Second)
	if tErr != nil {
		fmt.Printf("• Blocked Target (%s): FAIL (%v)\n", displayLabel, tErr)
	} else {
		fmt.Printf("• Blocked Target (%s): HTTP %d in %v (Unblocked)\n", displayLabel, tCode, tTime)
	}
}

func handleLogs(engineName string) {
	if engineName == "gecit" {
		c, _ := engine.RunCmd("journalctl", "-u", engine.GecitUnitName, "-f", "--no-pager")
		fmt.Println(c)
		return
	}
	c, _ := engine.RunCmd("journalctl", "--user", "-u", engine.SpoofUnitName, "-f", "--no-pager")
	fmt.Println(c)
}

func handleCleanup() {
	out, err := engine.EmergencyCleanup()
	if err != nil {
		fmt.Printf("Cleanup note: %s (%v)\n", out, err)
	} else {
		fmt.Println("Emergency cleanup executed successfully.")
	}
}
