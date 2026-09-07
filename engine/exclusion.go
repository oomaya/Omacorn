package engine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultExclusions contains the curated split-tunneling exemptions:
// 1. Loopback and private RFC1918 networks
// 2. Google Auth & core API infrastructure (preserves TLS 1.3 handshakes)
// 3. Telegram Messenger domains and Data Center CIDR blocks (bypasses HTTP proxy for raw MTProto)
var DefaultExclusions = []string{
	// Localhost & LAN
	"localhost",
	"127.0.0.1",
	"::1",
	"*.local",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",

	// Google Auth & Core Services
	"*.google.com",
	"accounts.google.com",
	"googleapis.com",
	"gstatic.com",

	// Telegram Domains
	"*.telegram.org",
	"*.t.me",
	"*.telegram.me",
	"*.tdesktop.com",

	// Telegram DC IPv4 Subnets (AS44907 & AS62041)
	"91.108.4.0/22",
	"91.108.8.0/22",
	"91.108.12.0/22",
	"91.108.16.0/22",
	"91.108.56.0/22",
	"149.154.160.0/20",
	"149.154.164.0/22",
	"149.154.168.0/22",
	"149.154.172.0/22",

	// Telegram DC IPv6 Subnets
	"2001:b28:f23d::/48",
	"2001:b28:f23f::/48",
	"2001:67c:4e8::/48",
}

func getExclusionsConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "omacorn", "exclusions.conf"), nil
}

// LoadUserExclusions reads custom user-defined exclusion patterns from ~/.config/omacorn/exclusions.conf
func LoadUserExclusions() ([]string, error) {
	p, err := getExclusionsConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var custom []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		custom = append(custom, line)
	}

	return custom, nil
}

// SaveUserExclusions writes user-defined exclusion patterns to ~/.config/omacorn/exclusions.conf
func SaveUserExclusions(rules []string) error {
	p, err := getExclusionsConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var lines []string
	lines = append(lines, "# Omacorn custom split-tunneling & proxy exclusion rules")
	lines = append(lines, "# One domain, hostname, IP, or CIDR per line (e.g. example.com or 192.168.1.0/24)")
	for _, r := range rules {
		trimmed := strings.TrimSpace(r)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	lines = append(lines, "")

	return os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0644)
}

// GetAllExclusions returns a merged, deduplicated slice containing both default and custom exclusions
func GetAllExclusions() []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range DefaultExclusions {
		clean := strings.TrimSpace(item)
		if clean != "" && !seen[clean] {
			seen[clean] = true
			result = append(result, clean)
		}
	}

	custom, _ := LoadUserExclusions()
	for _, item := range custom {
		clean := strings.TrimSpace(item)
		if clean != "" && !seen[clean] {
			seen[clean] = true
			result = append(result, clean)
		}
	}

	return result
}

// AddExclusion appends a new rule to the user's exclusion list and saves it
func AddExclusion(rule string) error {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return fmt.Errorf("exclusion rule cannot be empty")
	}

	// Check if already in defaults
	for _, d := range DefaultExclusions {
		if strings.EqualFold(d, rule) {
			return fmt.Errorf("%q is already part of default exclusions", rule)
		}
	}

	custom, err := LoadUserExclusions()
	if err != nil {
		return err
	}

	for _, c := range custom {
		if strings.EqualFold(c, rule) {
			return fmt.Errorf("%q is already in custom exclusions", rule)
		}
	}

	custom = append(custom, rule)
	return SaveUserExclusions(custom)
}

// RemoveExclusion removes a rule from the user's custom exclusions
func RemoveExclusion(rule string) error {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return fmt.Errorf("exclusion rule cannot be empty")
	}

	custom, err := LoadUserExclusions()
	if err != nil {
		return err
	}

	found := false
	var updated []string
	for _, c := range custom {
		if strings.EqualFold(c, rule) {
			found = true
			continue
		}
		updated = append(updated, c)
	}

	if !found {
		return fmt.Errorf("%q not found in custom exclusions", rule)
	}

	return SaveUserExclusions(updated)
}

// ResetExclusions removes the user exclusions file, restoring factory defaults
func ResetExclusions() error {
	p, err := getExclusionsConfigPath()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// CompileNoProxy synthesizes the active exclusions into a comma-delimited string for no_proxy
func CompileNoProxy() string {
	exclusions := GetAllExclusions()
	return strings.Join(exclusions, ",")
}

// CompileBrowserBypass synthesizes the active exclusions into a semicolon-separated string for Chromium flags
func CompileBrowserBypass() string {
	exclusions := GetAllExclusions()
	// Chromium flags require <-loopback> prefix for comprehensive loopback protection
	var parts []string
	parts = append(parts, "<-loopback>")

	seen := map[string]bool{"<-loopback>": true}
	for _, item := range exclusions {
		// Chromium does not support IPv6 bracketless CIDRs cleanly in bypass list, skip raw IPv6 CIDRs for browser flag
		if strings.Contains(item, "::") {
			continue
		}
		if !seen[item] {
			seen[item] = true
			parts = append(parts, item)
		}
	}

	return strings.Join(parts, ";")
}
