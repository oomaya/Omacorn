package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"dpipe/engine"
	"dpipe/tui"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.4.0"

// osExecutable returns the running binary. Tests replace it so the executable
// walk can be checked without relocating the test binary.
var osExecutable = os.Executable

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
	case "exclude", "exclusions", "bypass":
		handleExclude(os.Args[2:])
	case "run-clean", "clean-run", "spawn-direct":
		handleRunClean(os.Args[2:])
	case "update", "upgrade":
		handleUpdate()
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
	fmt.Printf(`🦄 Omacorn v%s — Pilot-Copilot Sovereign DPI Bypass Controller

Flight Architecture:
  Pilot 1 (Main Pilot):    spoofdpi (User-space proxy on 127.0.0.1:8080 - default, safe for VMs & all distros)
  Pilot 2 (Co-Pilot):      gecit    (Kernel eBPF sock_ops - high speed for bare metal)

Usage:
  %s                              Launch interactive Cyberpunk TUI dashboard
  %s start   [spoofdpi|gecit]     Handoff flight controls to selected engine (default: spoofdpi)
  %s stop    [spoofdpi|gecit|all] Disengage bypass engine(s)
  %s status  [--json]             Show flight status board (or JSON for Quickshell/Waybar)
  %s test    [spoofdpi|gecit]     Run live latency & handshake probe
  %s proxy   [on|off|toggle]      Manage desktop environment proxy (~/.config/environment.d)
  %s exclude [list|add|remove|reset] Manage split-tunneling & proxy bypass exclusions
  %s run-clean <command...>       Execute an app completely isolated from proxy env vars
  %s update                       Pull upstream, rebuild, hot-reload daemons, and re-sync
  %s logs    [spoofdpi|gecit]     Tail journal logs
  %s install [spoofdpi|gecit|all] Install systemd service(s)
  %s cleanup                      Emergency cleanup for eBPF and services
`, version, bin, bin, bin, bin, bin, bin, bin, bin, bin, bin, bin, bin)
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
	isVirt, hypervisor := engine.DetectHypervisor()

	if target == "gecit" {
		if isVirt {
			fmt.Printf("[gecit] ⚠️  WARNING: Running inside %s virtual machine.\n", strings.ToUpper(hypervisor))
			fmt.Println("       Kernel eBPF sock_ops with raw packet injection can trigger hypervisor DMA buffer faults (SVGA: Invalid PA range).")
			fmt.Println("       Recommendation: Keep Main Pilot (spoofdpi) in command for virtualized environments.")
		}

		// Mutual exclusion: disengage Main Pilot (spoofdpi) before engaging Co-Pilot (gecit)
		if engine.IsSpoofDPIRunning() {
			fmt.Println("[handoff] Disengaging Main Pilot (spoofdpi) before engaging Co-Pilot (gecit)...")
			engine.StopSpoofDPI()
		}

		out, err := engine.StartGecit()
		if err != nil {
			fmt.Printf("[gecit] Start failed: %s (%v)\n", out, err)
			return
		}
		fmt.Println("[gecit] Co-Pilot in command (eBPF sock_ops attached).")
		return
	}

	// Default: spoofdpi (Main Pilot)
	if engine.IsGecitRunning() {
		fmt.Println("[handoff] Disengaging Co-Pilot (gecit) before engaging Main Pilot (spoofdpi)...")
		engine.StopGecit()
	}

	out, err := engine.StartSpoofDPI()
	if err != nil {
		fmt.Printf("[spoofdpi] Start failed: %s (%v)\n", out, err)
		return
	}
	fmt.Printf("[spoofdpi] Main Pilot in command on %s\n", engine.SpoofDefaultAddr)
	if isVirt {
		fmt.Printf("[pilot] Platform: %s virtual machine (Hypervisor-safe user-space proxy mode active)\n", hypervisor)
	}
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
	fmt.Printf("           🦄 OMACORN v%s - Flight Status Board\n", version)
	fmt.Println("================================================================")
	spoofState := "STANDBY"
	if st.SpoofActive {
		spoofState = "ACTIVE (127.0.0.1:8080) [IN COMMAND]"
	}
	gecitState := "STANDBY"
	if st.GecitActive {
		gecitState = "ACTIVE (eBPF sock_ops) [IN COMMAND]"
	}
	if st.SpoofActive && st.GecitActive {
		gecitState = "⚠️  CONFLICT (Dual pilots active! Run 'omacorn start spoofdpi')"
	}

	proxyState := "DISABLED"
	if st.ProxyEnabled {
		proxyState = "ENABLED (~/.config/environment.d)"
	}

	vpnState := "inactive (Omacorn direct protection)"
	if st.VPNActive {
		vpnState = fmt.Sprintf("ACTIVE (%s - all traffic tunneled)", st.VPNInterface)
	}

	virtDesc := "Bare Metal (Native Linux Kernel)"
	if st.IsHypervisor {
		virtDesc = fmt.Sprintf("Virtualized (%s guest - Main Pilot recommended)", st.HypervisorName)
	}

	fmt.Printf("Pilot 1 (Main Pilot):   [spoofdpi] %s\n", spoofState)
	fmt.Printf("Pilot 2 (Co-Pilot):     [gecit]    %s\n", gecitState)
	fmt.Printf("Platform / Topology:    %s\n", virtDesc)
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
		cleanErr := strings.ReplaceAll(tErr.Error(), probeURL, displayLabel)
		fmt.Printf("• Blocked Target (%s): FAIL (%s)\n", displayLabel, cleanErr)
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

func handleExclude(args []string) {
	subcmd := "list"
	if len(args) > 0 {
		subcmd = strings.ToLower(args[0])
	}

	switch subcmd {
	case "list", "":
		custom, _ := engine.LoadUserExclusions()
		all := engine.GetAllExclusions()
		fmt.Println("================================================================")
		fmt.Printf("      🦄 OMACORN v%s — Split-Tunnel Exclusions & Bypass List\n", version)
		fmt.Println("================================================================")
		fmt.Printf("Default Exclusions: %d rules (Google Auth, Telegram MTProto, RFC1918)\n", len(engine.DefaultExclusions))
		fmt.Printf("Custom Exclusions:  %d rules (~/.config/omacorn/exclusions.conf)\n", len(custom))
		fmt.Println("----------------------------------------------------------------")
		fmt.Println("Active Rules:")
		for i, rule := range all {
			ruleType := "default"
			for _, c := range custom {
				if c == rule {
					ruleType = "custom"
					break
				}
			}
			fmt.Printf("  %2d. [%-7s] %s\n", i+1, ruleType, rule)
		}
		fmt.Println("================================================================")
		fmt.Println("Commands:")
		fmt.Println("  omacorn exclude add <domain|ip|cidr>")
		fmt.Println("  omacorn exclude remove <domain|ip|cidr>")
		fmt.Println("  omacorn exclude reset")
		fmt.Println("================================================================")

	case "add":
		if len(args) < 2 {
			fmt.Println("Usage: omacorn exclude add <domain|ip|cidr>")
			return
		}
		pattern := args[1]
		if err := engine.AddExclusion(pattern); err != nil {
			fmt.Printf("Error adding exclusion: %v\n", err)
			return
		}
		fmt.Printf("Added %q to exclusions.\n", pattern)
		if engine.IsProxyConfigured() {
			_ = engine.SyncProxyIfEnabled()
			fmt.Println("Synchronized updated exclusion to ~/.config/environment.d and browser flags.")
		}

	case "remove", "rm", "del":
		if len(args) < 2 {
			fmt.Println("Usage: omacorn exclude remove <domain|ip|cidr>")
			return
		}
		pattern := args[1]
		if err := engine.RemoveExclusion(pattern); err != nil {
			fmt.Printf("Error removing exclusion: %v\n", err)
			return
		}
		fmt.Printf("Removed %q from custom exclusions.\n", pattern)
		if engine.IsProxyConfigured() {
			_ = engine.SyncProxyIfEnabled()
			fmt.Println("Synchronized updated exclusion to ~/.config/environment.d and browser flags.")
		}

	case "reset":
		if err := engine.ResetExclusions(); err != nil {
			fmt.Printf("Error resetting exclusions: %v\n", err)
			return
		}
		fmt.Println("Reset custom exclusions to factory defaults.")
		if engine.IsProxyConfigured() {
			_ = engine.SyncProxyIfEnabled()
			fmt.Println("Synchronized updated exclusion to ~/.config/environment.d and browser flags.")
		}

	default:
		fmt.Printf("Unknown exclusion command: %s. Use list, add, remove, or reset.\n", subcmd)
	}
}

func handleRunClean(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: omacorn run-clean <command> [arguments...]")
		fmt.Println("Spawns an application with all proxy environment variables completely scrubbed.")
		return
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Filter environment variables to strip all proxy references
	var cleanEnv []string
	proxyVars := map[string]bool{
		"http_proxy":  true,
		"HTTP_PROXY":  true,
		"https_proxy": true,
		"HTTPS_PROXY": true,
		"all_proxy":   true,
		"ALL_PROXY":   true,
		"no_proxy":    true,
		"NO_PROXY":    true,
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) > 0 && proxyVars[parts[0]] {
			continue
		}
		cleanEnv = append(cleanEnv, env)
	}
	cmd.Env = cleanEnv

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Printf("Error running clean command %q: %v\n", cmdName, err)
		os.Exit(1)
	}
}

func handleUpdate() {
	fmt.Println("================================================================")
	fmt.Printf("      🦄 OMACORN v%s — Fleet Maintenance & Update Engine\n", version)
	fmt.Println("================================================================")

	home, _ := os.UserHomeDir()
	exe, _ := osExecutable()
	for _, p := range updateScriptCandidates(home, exe) {
		scriptPath := filepath.Join(p, "scripts", "update.sh")
		if info, err := os.Stat(scriptPath); err == nil && !info.IsDir() {
			cmd := exec.Command(scriptPath)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					os.Exit(exitErr.ExitCode())
				}
				fmt.Printf("Error running update script: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// Fallback for standalone binary installations without a git clone
	fmt.Println("[fallback] Git repository clone not found in standard paths.")
	fmt.Println("[fallback] Fetching latest pre-compiled static release from GitHub...")

	arch := "amd64"
	if out, err := engine.RunCmd("uname", "-m"); err == nil {
		if strings.Contains(out, "aarch64") || strings.Contains(out, "arm64") {
			arch = "arm64"
		}
	}

	downloadURL := fmt.Sprintf("https://github.com/oomaya/Omacorn/releases/latest/download/omacorn-linux-%s", arch)
	binDir := filepath.Join(home, ".local", "bin")
	_ = os.MkdirAll(binDir, 0755)
	targetBin := filepath.Join(binDir, "omacorn")
	tmpBin := targetBin + ".tmp"

	fetchCmd := exec.Command("curl", "-sSLf", downloadURL, "-o", tmpBin)
	if err := fetchCmd.Run(); err != nil {
		fmt.Printf("Failed to download release from %s: %v\n", downloadURL, err)
		return
	}
	_ = os.Chmod(tmpBin, 0755)
	if err := os.Rename(tmpBin, targetBin); err != nil {
		fmt.Printf("Failed to replace binary: %v\n", err)
		return
	}
	_ = os.Symlink("omacorn", filepath.Join(binDir, "dpipe"))

	if engine.IsSpoofDPIRunning() {
		_, _ = engine.RunCmd("systemctl", "--user", "restart", engine.SpoofUnitName)
		fmt.Println("  ✓ Hot-reloaded dpipe-spoofdpi user daemon")
	}

	if engine.IsProxyConfigured() {
		_ = engine.SyncProxyIfEnabled()
		fmt.Println("  ✓ Re-synchronized desktop proxy and split-tunnel exemptions")
	}

	fmt.Println("✓ Successfully updated Omacorn static binary!")
}

// updateScriptCandidates lists checkout directories that may contain
// scripts/update.sh. The first existing file wins. "." stays first so a
// checkout you launched from is preferred over every other guess.
func updateScriptCandidates(home, exe string) []string {
	var paths []string
	seen := map[string]bool{}
	add := func(p string) {
		if p == "" {
			return
		}
		clean := filepath.Clean(p)
		if seen[clean] {
			return
		}
		seen[clean] = true
		paths = append(paths, clean)
	}

	add(".")
	add(gitTopLevel("."))
	add(updateRootFromExecutable(exe))
	for _, p := range []string{
		filepath.Join(home, "Omacorn"),
		filepath.Join(home, "Projects", "Omacorn"),
		filepath.Join(home, "Projects", "omacorn"),
		filepath.Join(home, "src", "Omacorn"),
		filepath.Join(home, "src", "omacorn"),
		filepath.Join(home, ".local", "src", "Omacorn"),
	} {
		add(p)
	}
	return paths
}

func gitTopLevel(dir string) string {
	out, err := engine.RunCmd("git", "-C", dir, "rev-parse", "--show-toplevel")
	if err != nil || out == "" {
		return ""
	}
	return out
}

func updateRootFromExecutable(exe string) string {
	if exe == "" {
		return ""
	}
	dir := filepath.Dir(exe)
	for {
		script := filepath.Join(dir, "scripts", "update.sh")
		info, err := os.Stat(script)
		if err == nil && !info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
