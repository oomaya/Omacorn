package engine

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

const GecitUnitName = "dpipe-gecit"

// Hardened gecit template:
// 1. --doh=false to protect /etc/resolv.conf
// 2. --restore-after-bytes 600 so MSS clamping is strictly limited to the handshake
// 3. --mss 88 (sane default, avoiding the 40-byte packet shredder)
// 4. ExecStopPost to guarantee eBPF map & cgroup cleanup even on abnormal termination
// 5. CAP_PERFMON & CAP_SYS_ADMIN so BPF perf ring buffers are allowed
const gecitUnitTemplate = `[Unit]
Description=dpipe - gecit eBPF DPI bypass (hardened)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/gecit run --doh=false --fake-ttl 4 --mss 88 --restore-after-bytes 600 --ports 443
ExecStop=/usr/local/bin/gecit cleanup
ExecStopPost=/usr/local/bin/gecit cleanup
Restart=on-failure
RestartSec=3
CapabilityBoundingSet=CAP_BPF CAP_PERFMON CAP_NET_RAW CAP_NET_ADMIN CAP_SYS_ADMIN
AmbientCapabilities=CAP_BPF CAP_PERFMON CAP_NET_RAW CAP_NET_ADMIN CAP_SYS_ADMIN
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
`

func InstallGecit() error {
	if _, err := os.Stat("/usr/local/bin/gecit"); os.IsNotExist(err) {
		return fmt.Errorf("/usr/local/bin/gecit not found. Install gecit first")
	}

	unitPath := "/etc/systemd/system/" + GecitUnitName + ".service"
	currUser, _ := user.Current()
	if currUser != nil && currUser.Uid != "0" {
		cmd := exec.Command("sudo", "tee", unitPath)
		cmd.Stdin = strings.NewReader(gecitUnitTemplate)
		if err := cmd.Run(); err != nil {
			return err
		}
		exec.Command("sudo", "systemctl", "daemon-reload").Run()
		exec.Command("sudo", "systemctl", "enable", GecitUnitName).Run()
	} else {
		if err := os.WriteFile(unitPath, []byte(gecitUnitTemplate), 0644); err != nil {
			return err
		}
		RunCmd("systemctl", "daemon-reload")
		RunCmd("systemctl", "enable", GecitUnitName)
	}
	return nil
}

func StartGecit() (string, error) {
	currUser, _ := user.Current()
	if currUser != nil && currUser.Uid != "0" {
		c := exec.Command("sudo", "systemctl", "start", GecitUnitName)
		out, err := c.CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}
	return RunCmd("systemctl", "start", GecitUnitName)
}

func StopGecit() (string, error) {
	currUser, _ := user.Current()
	if currUser != nil && currUser.Uid != "0" {
		c := exec.Command("sudo", "systemctl", "stop", GecitUnitName)
		out, err := c.CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}
	return RunCmd("systemctl", "stop", GecitUnitName)
}

func EmergencyCleanup() (string, error) {
	RunCmd("systemctl", "--user", "stop", SpoofUnitName)
	currUser, _ := user.Current()
	if currUser != nil && currUser.Uid != "0" {
		exec.Command("sudo", "systemctl", "stop", GecitUnitName).Run()
		c := exec.Command("sudo", "gecit", "cleanup")
		out, err := c.CombinedOutput()
		exec.Command("sudo", "systemctl", "disable", GecitUnitName).Run()
		return strings.TrimSpace(string(out)), err
	}
	RunCmd("systemctl", "stop", GecitUnitName)
	out, err := RunCmd("gecit", "cleanup")
	RunCmd("systemctl", "disable", GecitUnitName)
	return out, err
}
