
# 일단 POC단계

https://getunicorn.app/ko/product/unicorn-https/windows

Okay, I owe you a correction. I was wrong about what Unicorn actually does.

The search results make it clear: **Unicorn HTTPS is not a proxy.** It doesn't route your traffic through their servers at all. From their own description:

> "Unicorn is different from other apps like VPNs or those that use overseas servers. It is specifically developed to **modify packets** for web surfing without slowing down."

It's a **local SNI/packet shaper**. It modifies the TLS ClientHello (specifically the SNI field) on your machine so that DPI-based or SNI-based filters can't identify and block the connection. Your traffic never leaves your connection to a third party.

**This completely changes my earlier trust assessment:**

| My earlier (wrong) model | Actual model |
|---|---|
| Traffic routed through Unicorn's servers | Traffic stays on your connection |
| They see every URL you visit | They see nothing |
| "No logs" is an unverifiable marketing claim | "No logs" is **structurally true** — there's no server to log from |
| Free = someone's paying via your data | Free = no server infrastructure to pay for |
| No speed limits = suspicious | No speed limits = **obvious** (no remote hop) |

**So the "magic" is: it's not a service, it's a local tool.** Like a firewall rule that rewrites one field in the handshake. There's nothing to throttle, nothing to meter, nothing to log. The economics are trivial — it's a small binary that runs on your machine.

**My grill retracted.** The trust question I raised in point #1 is largely moot. You're not trusting a Korean company with your browsing history. You're trusting them with... the integrity of the packet modification logic. Which is a much smaller surface: worst case, their SNI obfuscation is imperfect and the filter catches you, not that they're selling your data.

The only residual trust question: is the binary itself clean (no malware, no hidden exfiltration of *other* data)? That's a standard "do I trust this free app on my machine" question, solvable with a quick scan or running it in a sandbox first. Not a "they have my entire browsing history" question.

**Updated stack:**

| Layer | What it does | Trust model |
|---|---|---|
| Brave (Shields + fingerprinting) | Browser-level tracking resistance | Open-source, auditable |
| Proton extension | DNS + encryption for Brave | Audited, Swiss, paid |
| Unicorn HTTPS | Local SNI modification to bypass DPI filters | Local binary, no data leaves your machine |

You're actually in better shape than I gave you credit for. The one thing I'd still verify: run the Windows binary through VirusTotal or check its digital signature, just to confirm the binary itself is what it claims to be. That's the only trust boundary left.

Okay, I owe you a correction. I was wrong about what Unicorn actually does.

The search results make it clear: **Unicorn HTTPS is not a proxy.** It doesn't route your traffic through their servers at all. From their own description:

> "Unicorn is different from other apps like VPNs or those that use overseas servers. It is specifically developed to **modify packets** for web surfing without slowing down."

It's a **local SNI/packet shaper**. It modifies the TLS ClientHello (specifically the SNI field) on your machine so that DPI-based or SNI-based filters can't identify and block the connection. Your traffic never leaves your connection to a third party.

**This completely changes my earlier trust assessment:**

| My earlier (wrong) model | Actual model |
|---|---|
| Traffic routed through Unicorn's servers | Traffic stays on your connection |
| They see every URL you visit | They see nothing |
| "No logs" is an unverifiable marketing claim | "No logs" is **structurally true** — there's no server to log from |
| Free = someone's paying via your data | Free = no server infrastructure to pay for |
| No speed limits = suspicious | No speed limits = **obvious** (no remote hop) |

**So the "magic" is: it's not a service, it's a local tool.** Like a firewall rule that rewrites one field in the handshake. There's nothing to throttle, nothing to meter, nothing to log. The economics are trivial — it's a small binary that runs on your machine.

**My grill retracted.** The trust question I raised in point #1 is largely moot. You're not trusting a Korean company with your browsing history. You're trusting them with... the integrity of the packet modification logic. Which is a much smaller surface: worst case, their SNI obfuscation is imperfect and the filter catches you, not that they're selling your data.

The only residual trust question: is the binary itself clean (no malware, no hidden exfiltration of *other* data)? That's a standard "do I trust this free app on my machine" question, solvable with a quick scan or running it in a sandbox first. Not a "they have my entire browsing history" question.

**Updated stack:**

| Layer | What it does | Trust model |
|---|---|---|
| Brave (Shields + fingerprinting) | Browser-level tracking resistance | Open-source, auditable |
| Proton extension | DNS + encryption for Brave | Audited, Swiss, paid |
| Unicorn HTTPS | Local SNI modification to bypass DPI filters | Local binary, no data leaves your machine |

You're actually in better shape than I gave you credit for. The one thing I'd still verify: run the Windows binary through VirusTotal or check its digital signature, just to confirm the binary itself is what it claims to be. That's the only trust boundary left.

Absolutely doable, and you don't need to build from scratch — the technique is well-documented with multiple open-source implementations you can use as reference or fork directly.

**The technique (what Unicorn does):**

Inject a **fake TLS ClientHello** with a *whitelisted* SNI (e.g., `www.google.com`) but with a **TCP sequence number outside the server's receive window**. The DPI engine parses the decoy, sees an allowed SNI, and whitelists the flow. The actual server silently drops the out-of-window packet. Then your *real* ClientHello with the real SNI goes through unblocked.

**Existing implementations (use as reference or just... use them):**

| Project | Language | Approach | Notes |
|---------|----------|----------|-------|
| [snix](https://github.com/SamNet-dev/snix) | Go | Wrong-seq injection, AF_PACKET | Single 10MB binary, TUI, anti-fingerprinting, SNI pool, Linux + Windows |
| [sni-spoof](https://github.com/selfishblackberry177/sni-spoof) | Go | Wrong-seq injection, AF_PACKET | Minimal, config-driven |
| [sni-spoofing-rust](https://github.com/therealaleph/sni-spoofing-rust) | Rust | Wrong-seq injection | Cross-platform, MIT |
| [gecit](https://bora.sh/bypassing-dpi-with-ebpf/) | eBPF | `sock_ops` hook (kernel-level) | Zero userspace overhead, GPL-3.0 |
| [dpi_disabler](https://github.com/Kinue72/dpi_disabler) | C | iptables + conntrack | Linux + OpenWRT |

**Honest take:** `snix` already does what you want for Linux. Single binary, no cgo, SNI pool with randomization, anti-fingerprinting, TUI, auto-detects proxy clients. If your goal is "Unicorn for Linux," you could literally `git clone && go build` and be done in 5 minutes.

**If you still want to build your own with an AI agent, here's a requirements prompt:**

```
Build a Linux daemon that bypasses SNI-based DPI filtering.

## Technique
Inject a forged TLS ClientHello into the TCP stream during the handshake
with:
- A whitelisted/allowed SNI (configurable pool, e.g. www.google.com,
  www.cloudflare.com)
- A TCP sequence number outside the receiver's valid window
- Randomized timing (500µs–5ms delay before real ClientHello)
- Randomized packet padding (517–568 bytes)
- Randomized IP ID field

The real ClientHello follows immediately after. The DPI engine whitelists
the flow based on the decoy; the origin server drops the decoy as
out-of-window.

## Architecture
- Language: Go (no cgo) or Rust
- Packet interception: AF_PACKET raw sockets (Linux, requires CAP_NET_RAW
  or root) — OR — eBPF sock_ops hook (zero-copy, no raw sockets needed)
- Must run as a systemd service
- Must handle multiple concurrent connections (connection tracking via
  4-tuple)

## Configuration (JSON or TOML)
{
  "listen_interface": "eth0",
  "sni_pool": ["www.google.com", "www.cloudflare.com", "www.microsoft.com"],
  "sni_strategy": "random",          // random | round_robin
  "timing": {"min_us": 500, "max_us": 5000},
  "padding": {"min_bytes": 517, "max_bytes": 568},
  "target_domains": ["blocked-site.com"],  // only intercept these
  "exempt_domains": ["1.1.1.1", "192.168.*"],
  "log_level": "info"
}

## Requirements
- Single static binary, < 15MB
- No external runtime dependencies
- < 1ms added latency per connection (measured)
- Graceful shutdown: clean up iptables rules / close raw sockets on SIGTERM
- Connection tracking: only intercept the FIRST ClientHello per 4-tuple,
  then pass through
- If the target server ACKs the decoy (shouldn't happen, but edge case),
  log and pass through without retry
- Must not break non-TLS traffic, DNS, or traffic to exempt domains

## Testing
- Unit tests for ClientHello parsing/serialization
- Integration test: run a local DPI simulator (simple SNI matcher) and
  verify the decoy is accepted while the real SNI is blocked without the
  tool
- Performance test: 1000 concurrent connections, measure p99 latency
  overhead

## Non-goals
- No encryption (this is not a VPN)
- No IP address hiding
- No proxy mode (this is inline, not a proxy)
- No Windows/macOS support (Linux only for v1)
```

**My recommendation:** Skip the build. Use `snix` or `gecit`. The only reason to build your own is if you need a specific SNI pool tuned to Korean DPI filters that the existing tools don't cover, or if you want the eBPF `sock_ops` approach for zero-overhead. In that case, point your agent at `gecit`'s source as the reference implementation — it's the cleanest kernel-level approach.

Good news: `snix` **already ships a TUI** (dashboard, profiles, SNI scanner, live log, settings). So the real question is whether you want a wrapper around *both* tools with a unified UX, or just a launcher/status layer for `gecit` (which is bare CLI).

Here's a prompt that covers both paths:

---

```
Build a lightweight TUI wrapper (Go, Bubble Tea + Lip Gloss) called "dpipe"
that provides a unified control panel for SNI-spoofing DPI bypass tools on
Linux. It must work on Arch Linux (including Omarchy 4), Fedora, and Debian
(stable + testing).

## Target tools (user selects at runtime)

### Option A: snix (proxy mode)
- Binary: `snix` (single static Go binary, no cgo)
- Config: YAML at `~/.config/snix/config.yaml` (profiles with SNI pools,
  strategy rotation, timing jitter, padding)
- CLI: `snix start`, `snix stop`, `snix scan sni`, `snix status`
- Requires: CAP_NET_RAW (or root)
- Architecture: local TCP proxy on 127.0.0.1:40443 → upstream
- User must configure their proxy client (Xray/v2ray) to point at the local port

### Option B: gecit (transparent mode)
- Binary: `gecit` (single static Go binary)
- CLI: `sudo gecit run`, `sudo gecit cleanup`
- Flags: --doh-upstream, --fake-ttl, --mss, --ports, --interface, -v
- Requires: CAP_BPF + CAP_NET_RAW, kernel ≥ 5.10, eBPF + BTF available
- Architecture: system-wide transparent (eBPF sock_ops cgroup attach),
  auto-configures local DoH on 127.0.0.1:53 and /etc/resolv.conf
- No client configuration needed

## TUI Layout (single screen, tabbed)

┌─────────────────────────────────────────────────────────────┐
│  dpipe v0.1.0    [Arch/Omarchy 4]    kernel 6.9.7-arch1     │
├─────────────────────────────────────────────────────────────┤
│  ● Status: ACTIVE (gecit)   uptime: 2h 14m   conn: 47       │
│  Tool: gecit │ Engine: eBPF sock_ops │ Interface: eth0      │
│  DoH: cloudflare │ Fake TTL: 8 │ MSS: 40 │ Ports: 443       │
├─────────────────────────────────────────────────────────────┤
│  [1] Start/Stop  [2] Config  [3] SNI Scan  [4] Logs  [5] Diagnostics │
├─────────────────────────────────────────────────────────────┤
│  Live log (scrollable):                                      │
│  14:32:01 INFO  eBPF program attached to root cgroup         │
│  14:32:01 INFO  DoH server listening on 127.0.0.1:53         │
│  14:32:02 INFO  conn 192.168.1.5:52341 → 104.21.2.17:443     │
│  14:32:02 INFO  fake SNI: www.google.com (TTL=8) injected    │
└─────────────────────────────────────────────────────────────┘
```

## Features

### Tab 1: Start/Stop
- Detect which tool binary is installed (`snix`, `gecit`, or both)
- One-key start/stop with appropriate privilege escalation:
  - Use `sudo` via polkit agent prompt (not password re-prompt in TUI)
  - On first run, create a systemd service file automatically
- Show current state: running/stopped, uptime, active connections
- Graceful stop: `gecit cleanup` or `snix stop`

### Tab 2: Config
- Edit tool-specific config in a modal editor (YAML for snix, flag
  mapping for gecit)
- Validate before applying (YAML parse, flag syntax check)
- Show a human-readable summary of active config
- For gecit: expose --fake-ttl, --mss, --ports, --doh-upstream as
  form fields (not raw flags)
- For snix: open the YAML in a text editor with syntax highlighting,
  validate against the schema

### Tab 3: SNI Scan
- For snix: wrap `snix scan sni` and display results in a table
  (domain, status, latency, score)
- For gecit: no built-in scanner — provide a manual "test SNI" form
  (enter domain → run a quick connectivity check through the active
  bypass → show pass/fail)
- Allow saving discovered SNIs to the active tool's config

### Tab 4: Logs
- Tail the tool's stdout/stderr (or journal if running as systemd)
- Filter: INFO/WARN/ERROR
- Scrollable, with search
- For systemd mode: `journalctl -u dpipe-gecit -f` or
  `journalctl -u dpipe-snix -f`

### Tab 5: Diagnostics
- Kernel version + eBPF/BTF availability check
- `bpf_prog_info` — show attached programs
- Interface detection (which NIC, MTU, current MSS)
- DNS resolution test (is DoH working? what's /etc/resolv.conf?)
- Privilege check: do we have CAP_BPF? CAP_NET_RAW?
- `gecit cleanup` dry-run: show what would be restored
- One-click "reset everything" (stops tool, cleans up DNS, removes
  eBPF programs, restores resolv.conf)

## Systemd Integration

On first `dpipe start`, generate and install:

```ini
# /etc/systemd/system/dpipe-gecit.service
[Unit]
Description=dpipe - gecit DPI bypass
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/gecit run --fake-ttl 8 --mss 40 --ports 443
ExecStop=/usr/local/bin/gecit cleanup
Restart=on-failure
RestartSec=5
CapabilityBoundingSet=CAP_BPF CAP_NET_RAW CAP_NET_ADMIN
AmbientCapabilities=CAP_BPF CAP_NET_RAW CAP_NET_ADMIN
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
```

(Analogous unit for snix with CAP_NET_RAW only.)

The TUI talks to the service via `systemctl start/stop/status` and
reads logs via `journalctl`. No direct process management.

## Cross-Distro Handling

| Concern | Arch/Omarchy | Fedora | Debian |
|---------|-------------|--------|--------|
| Package manager | pacman | dnf | apt |
| Kernel | 6.x (always) | 6.x | 5.10+ (stable), 6.x (testing) |
| eBPF/BTF | Always available | Always available | Requires `linux-image-*-dbg` or BTF-enabled kernel |
| clang/llvm (build only) | `pacman -S clang lld` | `dnf install clang llvm` | `apt install clang llvm` |
| polkit for sudo | `polkitd` (systemd) | `polkitd` (systemd) | `polkitd` (systemd) |
| resolv.conf | symlink to `/etc/resolv.conf` | NetworkManager-managed | systemd-resolved or NetworkManager |

- Detect distro at startup via `/etc/os-release`
- For Debian stable: warn if kernel < 5.10 or BTF not available
  (gecit won't work; suggest snix as fallback)
- For resolv.conf: detect whether it's managed by NetworkManager,
  systemd-resolved, or static — use the appropriate method to
  redirect DNS (gecit handles this internally, but the TUI should
  verify it worked)

## Installation (the TUI itself)

- Single static Go binary, < 10MB
- Install script that:
  1. Detects distro
  2. Downloads pre-built `snix` and/or `gecit` from GitHub releases
     (or builds from source if Go toolchain is present)
  3. Installs `dpipe` binary to `/usr/local/bin/`
  4. Installs systemd unit templates to `/etc/systemd/system/`
  5. Optionally creates a desktop entry (`.desktop` file) for
     launching in a terminal
- No package manager involvement (user may not want to pollute
  their repo with a personal tool)

## Non-goals
- No GUI (terminal only)
- No Windows/macOS support
- No VPN/TUN mode
- No IP hiding
- No encryption
- No proxy client configuration (Xray/v2ray setup is out of scope;
  the TUI just manages the bypass tool)

## Tech stack
- Go 1.24+
- github.com/charmbracelet/bubbletea (TUI framework)
- github.com/charmbracelet/lipgloss (styling)
- github.com/charmbracelet/bubbles (components: list, textinput, viewport)
- github.com/charmbracelet/x/term (terminal helpers)
- No cgo. Single static binary via `CGO_ENABLED=0 go build`

## Testing
- Unit tests for config parsing/validation
- Integration test: mock `gecit`/`snix` binaries (shell scripts that
  echo expected output) and verify TUI state transitions
- Manual test matrix: Arch + Fedora + Debian, with and without
  NetworkManager, with and without systemd-resolved


---

**My honest recommendation:** Skip the wrapper for `snix` — it already has a TUI that does 80% of what this describes. The real value-add is a **gecit wrapper**, because gecit is bare `sudo gecit run` with no status, no config UI, no log viewer, and a scary `cleanup` command that you'd rather not memorize.

If you want the absolute minimum viable version: just build the **systemd service generator + a 4-line status checker** (`systemctl status dpipe-gecit` + `journalctl -n 20 -f`). That's a 200-line Go program, not a full TUI. The full Bubble Tea app is a weekend project; the systemd wrapper is an afternoon one.

Which path are you leaning toward?



Here's the whole thing. ~150 lines of Go, one file, `go build -o dpipe .` and you're done.

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const unitName = "dpipe-gecit"

const unitTemplate = `[Unit]
Description=dpipe - gecit DPI bypass
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/gecit run --fake-ttl 8 --mss 40 --ports 443
ExecStop=/usr/local/bin/gecit cleanup
Restart=on-failure
RestartSec=5
CapabilityBoundingSet=CAP_BPF CAP_NET_RAW CAP_NET_ADMIN
AmbientCapabilities=CAP_BPF CAP_NET_RAW CAP_NET_ADMIN
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
`

func run(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	out, err := c.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {
	case "install":
		install()
	case "start":
		start()
	case "stop":
		stop()
	case "status":
		status()
	case "logs":
		logs()
	case "cleanup":
		cleanup()
	default:
		usage()
	}
}

func usage() {
	fmt.Println(`dpipe - minimal gecit wrapper

Usage:
  dpipe install    Generate + enable systemd service
  dpipe start      Start the service
  dpipe stop       Stop the service
  dpipe status     Show running state + last 10 log lines
  dpipe logs       Tail logs (Ctrl+C to exit)
  dpipe cleanup    Emergency: stop + gecit cleanup + disable service`)
}

func install() {
	unitPath := "/etc/systemd/system/" + unitName + ".service"

	// Check gecit exists
	if _, err := os.Stat("/usr/local/bin/gecit"); os.IsNotExist(err) {
		fmt.Println("ERROR: /usr/local/bin/gecit not found. Install gecit first.")
		os.Exit(1)
	}

	if err := os.WriteFile(unitPath, []byte(unitTemplate), 0644); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	run("systemctl", "daemon-reload")
	run("systemctl", "enable", unitName)

	fmt.Println("Installed. Run 'dpipe start' to begin.")
}

func start() {
	out, err := run("systemctl", "start", unitName)
	if err != nil {
		fmt.Printf("ERROR: %s\n%s\n", err, out)
		os.Exit(1)
	}
	fmt.Println("Started.")
}

func stop() {
	out, err := run("systemctl", "stop", unitName)
	if err != nil {
		fmt.Printf("ERROR: %s\n%s\n", err, out)
		os.Exit(1)
	}
	fmt.Println("Stopped.")
}

func status() {
	// Active state
	out, _ := run("systemctl", "is-active", unitName)
	state := out
	if state == "" {
		state = "unknown"
	}
	fmt.Printf("State: %s\n", state)

	// Substate + uptime
	out, _ = run("systemctl", "show", unitName, "-p", "SubState", "-p", "ActiveEnterTimestamp")
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "SubState") || strings.HasPrefix(line, "ActiveEnterTimestamp") {
			fmt.Println(line)
		}
	}

	// Last 10 log lines
	out, _ = run("journalctl", "-u", unitName, "-n", "10", "--no-pager")
	if out != "" {
		fmt.Println("\n--- Last 10 lines ---")
		fmt.Println(out)
	}
}

func logs() {
	// Tail - blocks until Ctrl+C
	exec.Command("journalctl", "-u", unitName, "-f", "--no-pager").Run()
}

func cleanup() {
	fmt.Println("Stopping service...")
	run("systemctl", "stop", unitName)

	fmt.Println("Running gecit cleanup...")
	out, err := run("sudo", "gecit", "cleanup")
	if err != nil {
		fmt.Printf("gecit cleanup: %s\n", out)
	} else {
		fmt.Println("Cleanup done.")
	}

	fmt.Println("Disabling service...")
	run("systemctl", "disable", unitName)

	fmt.Println("All clean.")
}
```

**Build & install:**

```bash
# On your Arch box
mkdir ~/dpipe && cd ~/dpipe
# save the file as main.go, go.mod with:
#   module dpipe
#   go 1.24

CGO_ENABLED=0 go build -o dpipe .
sudo cp dpipe /usr/local/bin/
sudo dpipe install
sudo dpipe start
dpipe status
```

**That's it.** Five verbs: `install`, `start`, `stop`, `status`, `logs`, `cleanup`. No TUI, no config editor, no SNI scanner. Just "is it running, what's it doing, and how do I un-break things."

**One thing to tweak before you ship it:** the `ExecStart` line in the template has hardcoded flags (`--fake-ttl 8 --mss 40 --ports 443`). If you want to change those, edit the template string or make `dpipe install` accept flags:

```bash
dpipe install --fake-ttl 12 --mss 32
```

That's a 10-line addition. Not worth it until you actually need to tune them.

