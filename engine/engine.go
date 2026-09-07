package engine

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type SystemStatus struct {
	SpoofActive    bool
	SpoofAddr      string
	SpoofUptime    string
	GecitActive    bool
	GecitUptime    string
	IsHypervisor   bool
	HypervisorName string
	ProxyEnabled   bool
	DNSProtected   bool
	DNSDetails     string
	VPNActive      bool
	VPNInterface   string
	GoogleLatency  time.Duration
	GoogleStatus   int
	GoogleErr      string
	TargetLatency  time.Duration
	TargetStatus   int
	TargetErr      string
}

func RunCmd(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	out, err := c.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// DetectHypervisor identifies whether Omacorn is running in a virtual machine (VMware, KVM, VirtualBox, etc.)
func DetectHypervisor() (bool, string) {
	out, err := RunCmd("systemd-detect-virt")
	if err == nil && out != "" && out != "none" {
		return true, out
	}
	// Fallback check
	if data, err := os.ReadFile("/sys/class/dmi/id/sys_vendor"); err == nil {
		vendor := strings.ToLower(string(data))
		if strings.Contains(vendor, "vmware") {
			return true, "vmware"
		}
		if strings.Contains(vendor, "qemu") || strings.Contains(vendor, "kvm") {
			return true, "kvm"
		}
		if strings.Contains(vendor, "virtualbox") {
			return true, "oracle"
		}
	}
	return false, ""
}

// IsSpoofDPIRunning checks if Main Pilot (SpoofDPI user service) is active
func IsSpoofDPIRunning() bool {
	out, _ := RunCmd("systemctl", "--user", "is-active", SpoofUnitName)
	return out == "active"
}

// IsGecitRunning checks if Co-Pilot (Gecit system service) is active
func IsGecitRunning() bool {
	out, _ := RunCmd("systemctl", "is-active", GecitUnitName)
	return out == "active"
}

func GetSystemStatus() SystemStatus {
	st := SystemStatus{
		SpoofAddr: SpoofDefaultAddr,
	}

	// 0. Detect Hypervisor
	isVirt, virtName := DetectHypervisor()
	st.IsHypervisor = isVirt
	st.HypervisorName = virtName

	// 1. Check SpoofDPI (Main Pilot)
	if IsSpoofDPIRunning() {
		st.SpoofActive = true
		ts, _ := RunCmd("systemctl", "--user", "show", SpoofUnitName, "-p", "ActiveEnterTimestamp")
		st.SpoofUptime = strings.TrimPrefix(ts, "ActiveEnterTimestamp=")
	}

	// 2. Check Gecit (Co-Pilot / Standby)
	if IsGecitRunning() {
		st.GecitActive = true
		ts, _ := RunCmd("systemctl", "show", GecitUnitName, "-p", "ActiveEnterTimestamp")
		st.GecitUptime = strings.TrimPrefix(ts, "ActiveEnterTimestamp=")
	}

	// 3. Check System Proxy
	st.ProxyEnabled = IsProxyConfigured()

	// 4. Check DNS
	resolvBytes, err := os.ReadFile("/etc/resolv.conf")
	if err == nil {
		content := string(resolvBytes)
		if strings.Contains(content, "systemd-resolved") || strings.Contains(content, "127.0.0.53") {
			st.DNSProtected = true
			st.DNSDetails = "systemd-resolved (protected)"
		} else {
			st.DNSProtected = false
			st.DNSDetails = "non-standard (check /etc/resolv.conf)"
		}
	}

	// 5. Check VPN status (Layer 3 tunnel detection)
	linkOut, linkErr := RunCmd("ip", "-o", "link", "show")
	if linkErr == nil {
		for _, line := range strings.Split(linkOut, "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				iface := strings.TrimSuffix(fields[1], ":")
				if strings.HasPrefix(iface, "tun") ||
					strings.HasPrefix(iface, "wg") ||
					strings.HasPrefix(iface, "proton") ||
					strings.HasPrefix(iface, "mullvad") ||
					strings.HasPrefix(iface, "tailscale") {
					st.VPNActive = true
					st.VPNInterface = iface
					break
				}
			}
		}
	}

	return st
}

func ProbeURL(targetURL string, proxyAddr string, timeout time.Duration) (int, time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	transport := &http.Transport{}
	if proxyAddr != "" {
		pURL, err := url.Parse("http://" + proxyAddr)
		if err == nil {
			transport.Proxy = http.ProxyURL(pURL)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return 0, 0, err
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return 0, elapsed, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, elapsed, nil
}

// GetProbeTarget returns the target URL to test and a sanitized display label.
// Never exposes sensitive or blocked domain names in source code or output.
func GetProbeTarget() (string, string) {
	if env := os.Getenv("OMACORN_TEST_TARGET"); env != "" {
		return env, "<custom-target>"
	}

	home, err := os.UserHomeDir()
	if err == nil {
		targetFile := filepath.Join(home, ".config", "omacorn", "target.txt")
		if data, err := os.ReadFile(targetFile); err == nil {
			target := strings.TrimSpace(string(data))
			if target != "" {
				return target, "<blocked-site>"
			}
		}
	}

	return "https://news.ycombinator.com", "<test-target>"
}
