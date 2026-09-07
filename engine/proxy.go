package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const ProxyConfigContent = `# Omacorn system-wide desktop proxy configuration
# Routes standard web traffic through local SpoofDPI daemon (port 8080)
# Exempts localhost, local addresses, and Google Auth/APIs to prevent handshake collisions
http_proxy=http://127.0.0.1:8080
https_proxy=http://127.0.0.1:8080
all_proxy=http://127.0.0.1:8080
no_proxy=localhost,127.0.0.1,*.local,*.google.com,accounts.google.com,googleapis.com
`

func getEnvDPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "environment.d", "20-omacorn-proxy.conf"), nil
}

func IsProxyConfigured() bool {
	p, err := getEnvDPath()
	if err != nil {
		return false
	}
	if _, err := os.Stat(p); err == nil {
		return true
	}
	return false
}

var BrowserFlagFiles = []string{
	"brave-origin-nightly-flags.conf",
	"brave-origin-flags.conf",
	"brave-flags.conf",
	"brave-browser-flags.conf",
	"brave-beta-flags.conf",
	"brave-nightly-flags.conf",
	"brave-dev-flags.conf",
	"chromium-flags.conf",
	"chrome-flags.conf",
	"google-chrome-flags.conf",
	"vivaldi-stable.conf",
}

var FlatpakBrowserMap = map[string]string{
	"com.brave.Browser":         "brave-flags.conf",
	"com.brave.Browser.beta":    "brave-flags.conf",
	"com.brave.Browser.nightly": "brave-flags.conf",
	"com.brave.Browser.dev":     "brave-flags.conf",
	"org.chromium.Chromium":     "chromium-flags.conf",
	"com.google.Chrome":         "chrome-flags.conf",
	"com.google.ChromeDev":      "chrome-flags.conf",
	"com.vivaldi.Vivaldi":       "vivaldi-flags.conf",
	"com.microsoft.Edge":        "edge-flags.conf",
}

const (
	browserProxyFlag     = "--proxy-server=http://127.0.0.1:8080"
	browserBypassFlag    = "--proxy-bypass-list=<-loopback>;localhost;*.google.com;accounts.google.com;googleapis.com"
	browserCommentHeader = "# Omacorn native proxy routing (bypasses Google Auth & localhost)"
)

func updateFlagFile(p string, enable bool) {
	var lines []string
	if data, err := os.ReadFile(p); err == nil {
		lines = strings.Split(string(data), "\n")
	} else if !enable {
		return
	}

	var newLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--proxy-server=") ||
			strings.HasPrefix(trimmed, "--proxy-bypass-list=") ||
			trimmed == browserCommentHeader {
			continue
		}
		newLines = append(newLines, line)
	}

	if enable {
		newLines = append(newLines, "", browserCommentHeader, browserProxyFlag, browserBypassFlag)
	}

	content := strings.TrimSpace(strings.Join(newLines, "\n"))
	if content != "" {
		content += "\n"
	}
	_ = os.WriteFile(p, []byte(content), 0644)
}

func syncBrowserFlags(enable bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// 1. Native host config directories (~/.config and ~/dotfiles/omarchy/.config)
	configDirs := []string{
		filepath.Join(home, ".config"),
		filepath.Join(home, "dotfiles", "omarchy", ".config"),
	}

	for _, dir := range configDirs {
		for _, fname := range BrowserFlagFiles {
			p := filepath.Join(dir, fname)
			info, err := os.Stat(p)
			if err != nil || info.IsDir() {
				continue
			}

			// If it is a symlink within ~/.config pointing to a sibling flags file, skip to avoid double processing
			if dir == filepath.Join(home, ".config") {
				if lTarget, err := os.Readlink(p); err == nil && !strings.HasPrefix(lTarget, "/") && !strings.HasPrefix(lTarget, "..") {
					continue
				}
			}

			updateFlagFile(p, enable)
		}
	}

	// 2. Flatpak application config directories (~/.var/app/<app-id>/config/)
	varAppDir := filepath.Join(home, ".var", "app")
	if entries, err := os.ReadDir(varAppDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			appID := entry.Name()
			flagName, ok := FlatpakBrowserMap[appID]
			if !ok {
				continue
			}

			flatpakCfgDir := filepath.Join(varAppDir, appID, "config")
			if err := os.MkdirAll(flatpakCfgDir, 0755); err != nil {
				continue
			}

			p := filepath.Join(flatpakCfgDir, flagName)
			updateFlagFile(p, enable)
		}
	}

	// 3. User-level Flatpak environment overrides (for sandboxed browsers/apps like Firefox Flatpak)
	if _, err := exec.LookPath("flatpak"); err == nil {
		if enable {
			RunCmd("flatpak", "override", "--user",
				"--env=http_proxy=http://127.0.0.1:8080",
				"--env=https_proxy=http://127.0.0.1:8080",
				"--env=all_proxy=http://127.0.0.1:8080",
				"--env=no_proxy=localhost,127.0.0.1,*.local,*.google.com,accounts.google.com,googleapis.com",
			)
		} else {
			RunCmd("flatpak", "override", "--user",
				"--unset-env=http_proxy",
				"--unset-env=https_proxy",
				"--unset-env=all_proxy",
				"--unset-env=no_proxy",
			)
		}
	}
}

func EnableSystemProxy() error {
	p, err := getEnvDPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(p, []byte(ProxyConfigContent), 0644); err != nil {
		return err
	}

	// Update active systemd user session
	os.Setenv("http_proxy", "http://127.0.0.1:8080")
	os.Setenv("https_proxy", "http://127.0.0.1:8080")
	os.Setenv("all_proxy", "http://127.0.0.1:8080")
	os.Setenv("no_proxy", "localhost,127.0.0.1,*.local,*.google.com,accounts.google.com,googleapis.com")

	RunCmd("systemctl", "--user", "import-environment", "http_proxy", "https_proxy", "all_proxy", "no_proxy")

	// Synchronize native flags for all detected Chromium/Brave/Vivaldi browsers
	syncBrowserFlags(true)

	return nil
}

func DisableSystemProxy() error {
	p, err := getEnvDPath()
	if err != nil {
		return err
	}

	_ = os.Remove(p)

	// Also remove symlink/file in dotfiles if present
	home, _ := os.UserHomeDir()
	if home != "" {
		_ = os.Remove(filepath.Join(home, "dotfiles", "omarchy", ".config", "environment.d", "20-omacorn-proxy.conf"))
	}

	RunCmd("systemctl", "--user", "unset-environment", "http_proxy", "https_proxy", "all_proxy", "no_proxy")

	// Remove proxy flags from all detected browsers
	syncBrowserFlags(false)

	return nil
}

func ToggleSystemProxy() (bool, error) {
	if IsProxyConfigured() {
		err := DisableSystemProxy()
		return false, err
	}
	err := EnableSystemProxy()
	return true, err
}
