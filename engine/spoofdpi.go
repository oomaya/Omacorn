package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	SpoofUnitName    = "dpipe-spoofdpi"
	SpoofDefaultPort = "8080"
	SpoofDefaultAddr = "127.0.0.1:" + SpoofDefaultPort
)

const spoofUnitTemplate = `[Unit]
Description=dpipe - SpoofDPI user daemon
After=network.target

[Service]
Type=simple
ExecStart=%s --no-tui --listen-addr %s --default-fake-ttl 8 --https-fake-count 1 --dns-mode system --log-level warn
Restart=on-failure
RestartSec=3

[Install]
WantedBy=default.target
`

func FindSpoofDPIBinary() (string, error) {
	if p, err := exec.LookPath("spoofdpi"); err == nil {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err == nil {
		p := filepath.Join(home, "go", "bin", "spoofdpi")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if _, err := os.Stat("/usr/local/bin/spoofdpi"); err == nil {
		return "/usr/local/bin/spoofdpi", nil
	}
	return "", fmt.Errorf("spoofdpi binary not found. Run: go install github.com/xvzc/spoofdpi/cmd/spoofdpi@latest")
}

func InstallSpoofDPI() error {
	binPath, err := FindSpoofDPIBinary()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	userSystemdDir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(userSystemdDir, 0755); err != nil {
		return err
	}

	unitPath := filepath.Join(userSystemdDir, SpoofUnitName+".service")
	content := fmt.Sprintf(spoofUnitTemplate, binPath, SpoofDefaultAddr)
	if err := os.WriteFile(unitPath, []byte(content), 0644); err != nil {
		return err
	}

	RunCmd("systemctl", "--user", "daemon-reload")
	RunCmd("systemctl", "--user", "enable", SpoofUnitName)
	return nil
}

func StartSpoofDPI() (string, error) {
	return RunCmd("systemctl", "--user", "start", SpoofUnitName)
}

func StopSpoofDPI() (string, error) {
	return RunCmd("systemctl", "--user", "stop", SpoofUnitName)
}
