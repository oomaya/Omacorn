#!/usr/bin/env bash
# ==============================================================================
# 🦄 Omacorn Installer & Deployment Engine (Arch Linux / CachyOS / Omarchy)
# ==============================================================================
# Zero-fuss, single-command setup for workstations and laptops.
# ==============================================================================
set -euo pipefail

BOLD='\033[1m'
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

log_step() {
    echo -e "${CYAN}${BOLD}[Omacorn]${NC} $1"
}

log_ok() {
    echo -e "${GREEN}${BOLD}  ✓${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}${BOLD}  !${NC} $1"
}

log_err() {
    echo -e "${RED}${BOLD}  ✗${NC} $1"
}

echo -e "${CYAN}${BOLD}"
echo "================================================================"
echo "      🦄 OMACORN v0.3.0 — Automated Deployment Suite           "
echo "================================================================"
echo -e "${NC}"

# 1. Verify Go Compiler
log_step "Checking Go environment..."
if ! command -v go &>/dev/null; then
    log_err "Go compiler not found. Please install it first:"
    echo "       sudo pacman -S go"
    exit 1
fi
log_ok "Go compiler detected: $(go version | awk '{print $3}')"

# 2. Verify or Install SpoofDPI Engine
log_step "Checking Engine 1 (SpoofDPI)..."
SPOOFDPI_BIN=""
if command -v spoofdpi &>/dev/null; then
    SPOOFDPI_BIN="$(command -v spoofdpi)"
elif [[ -x "${HOME}/go/bin/spoofdpi" ]]; then
    SPOOFDPI_BIN="${HOME}/go/bin/spoofdpi"
else
    log_step "Installing SpoofDPI via 'go install'..."
    go install github.com/xvzc/spoofdpi/cmd/spoofdpi@latest
    SPOOFDPI_BIN="${HOME}/go/bin/spoofdpi"
fi
log_ok "SpoofDPI binary located at: ${SPOOFDPI_BIN}"

# 3. Configure CAP_NET_RAW Capability
log_step "Ensuring CAP_NET_RAW capability on SpoofDPI for user-space decoy packets..."
CURRENT_CAPS=$(getcap "${SPOOFDPI_BIN}" 2>/dev/null || true)
if [[ "${CURRENT_CAPS}" != *"cap_net_raw"* ]]; then
    echo "       Requesting sudo to set: setcap cap_net_raw+ep ${SPOOFDPI_BIN}"
    sudo setcap cap_net_raw+ep "${SPOOFDPI_BIN}"
    log_ok "Granted CAP_NET_RAW to ${SPOOFDPI_BIN}"
else
    log_ok "CAP_NET_RAW is already configured."
fi

# 4. Build Omacorn CLI & TUI
log_step "Compiling Omacorn binary..."
BIN_DIR="${HOME}/.local/bin"
mkdir -p "${BIN_DIR}"
go build -o omacorn .
cp omacorn "${BIN_DIR}/omacorn"
chmod +x "${BIN_DIR}/omacorn"
ln -sf omacorn "${BIN_DIR}/dpipe"
ln -sf omacorn "dpipe"
log_ok "Installed 'omacorn' and alias 'dpipe' to ${BIN_DIR}"

# 5. Install Systemd User Unit
log_step "Installing systemd user service..."
"${BIN_DIR}/omacorn" install spoofdpi
log_ok "Installed dpipe-spoofdpi.service"

# 6. Start Omacorn Daemon
log_step "Activating Omacorn SpoofDPI daemon..."
"${BIN_DIR}/omacorn" start spoofdpi
log_ok "Engine 1 active on 127.0.0.1:8080"

# 7. Configure Desktop System Proxy & Browser Flags
log_step "Synchronizing desktop environment & native browser flags..."
"${BIN_DIR}/omacorn" proxy on
log_ok "Proxy wired to ~/.config/environment.d and all browser config files"

# 8. Install Desktop Launcher
log_step "Installing desktop application launcher (omacorn.desktop)..."
APP_DIR="${HOME}/.local/share/applications"
mkdir -p "${APP_DIR}"
cat << 'EOF' > "${APP_DIR}/omacorn.desktop"
[Desktop Entry]
Version=1.0
Type=Application
Name=Omacorn
GenericName=DPI Bypass Dashboard
Comment=Sovereign Twin-Engine DPI Evasion Daemon & Cyberpunk TUI
Exec=ghostty --title=Omacorn -e omacorn
Icon=network-vpn
Terminal=false
Categories=Network;Security;System;
EOF
update-desktop-database "${APP_DIR}" 2>/dev/null || true
log_ok "Registered Omacorn in application menu"

# 9. Live Verification Probe
echo ""
log_step "Executing live verification probe..."
"${BIN_DIR}/omacorn" test

echo ""
echo -e "${GREEN}${BOLD}================================================================"
echo "  🎉 OMACORN IS FULLY DEPLOYED & LIVE!"
echo "================================================================"
echo -e "${NC}"
echo -e "  • Launch interactive Cyberpunk TUI:  ${CYAN}omacorn${NC} (or ${CYAN}dpipe${NC})"
echo -e "  • Desktop Application Menu:         Search '${CYAN}Omacorn${NC}' (Super key)"
echo -e "  • Check status board:               ${CYAN}omacorn status${NC}"
echo -e "  • Status JSON for bars:             ${CYAN}omacorn status --json${NC}"
echo ""
