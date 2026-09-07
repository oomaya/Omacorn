package engine

import (
	"os"
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

const (
	browserProxyFlag     = "--proxy-server=http://127.0.0.1:8080"
	browserBypassFlag    = "--proxy-bypass-list=<-loopback>;localhost;*.google.com;accounts.google.com;googleapis.com"
	browserCommentHeader = "# Omacorn native proxy routing (bypasses Google Auth & localhost)"
)

func syncBrowserFlags(enable bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

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

			data, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			lines := strings.Split(string(data), "\n")
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

			_ = os.WriteFile(p, []byte(strings.TrimSpace(strings.Join(newLines, "\n"))+"\n"), 0644)
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
