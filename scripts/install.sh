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

# 1. Distro Detection
DISTRO_ID="unknown"
DISTRO_NAME="Linux"
if [[ -f /etc/os-release ]]; then
    # shellcheck disable=SC1091
    . /etc/os-release
    DISTRO_ID="${ID:-unknown}"
    DISTRO_NAME="${PRETTY_NAME:-$DISTRO_ID}"
fi
log_step "Detected platform: ${DISTRO_NAME} (${DISTRO_ID})"

# 2. Verify or Install Omacorn Binary
BIN_DIR="${HOME}/.local/bin"
mkdir -p "${BIN_DIR}"

ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64) ARCH_SUFFIX="amd64" ;;
    aarch64|arm64) ARCH_SUFFIX="arm64" ;;
    *) ARCH_SUFFIX="amd64" ;;
esac

if [[ -f "./omacorn" && -x "./omacorn" ]]; then
    log_ok "Using pre-compiled Omacorn binary from repository."
    cp ./omacorn "${BIN_DIR}/omacorn"
elif [[ -f "./main.go" ]] && command -v go &>/dev/null; then
    log_step "Go compiler detected ($(go version | awk '{print $3}')). Compiling static Omacorn binary..."
    CGO_ENABLED=0 go build -ldflags="-s -w" -o omacorn .
    cp omacorn "${BIN_DIR}/omacorn"
else
    log_step "Fetching latest static Omacorn release for linux-${ARCH_SUFFIX}..."
    OMACORN_DL_URL="https://github.com/oomaya/Omacorn/releases/latest/download/omacorn-linux-${ARCH_SUFFIX}"
    if curl -sSLf "${OMACORN_DL_URL}" -o "${BIN_DIR}/omacorn" 2>/dev/null; then
        log_ok "Downloaded static release binary from GitHub"
    elif command -v go &>/dev/null; then
        log_step "Compiling from source via go install..."
        go install github.com/oomaya/Omacorn@latest 2>/dev/null || true
        if [[ -x "${HOME}/go/bin/Omacorn" ]]; then
            cp "${HOME}/go/bin/Omacorn" "${BIN_DIR}/omacorn"
        fi
    else
        log_err "Neither pre-compiled binary, GitHub release, nor Go compiler found."
        echo "       Please install Go according to your system:"
        case "${DISTRO_ID}" in
            arch|cachyos|omarchy)
                echo "       sudo pacman -S go"
                ;;
            debian|ubuntu|pika)
                echo "       sudo apt update && sudo apt install -y golang-go libcap2-bin"
                ;;
            fedora)
                echo "       sudo dnf install golang (or compile inside distrobox/toolbox on Atomic spins)"
                ;;
            *)
                echo "       Install Go from https://go.dev/dl/ or copy a pre-built static omacorn binary."
                ;;
        esac
        exit 1
    fi
fi
chmod +x "${BIN_DIR}/omacorn"
ln -sf omacorn "${BIN_DIR}/dpipe"
if [[ -f "./main.go" ]]; then
    ln -sf omacorn "dpipe"
fi
log_ok "Installed 'omacorn' and alias 'dpipe' to ${BIN_DIR}"

# 3. Verify or Install SpoofDPI Engine
log_step "Checking Engine 1 (SpoofDPI)..."
SPOOFDPI_BIN=""
if command -v spoofdpi &>/dev/null; then
    SPOOFDPI_BIN="$(command -v spoofdpi)"
elif [[ -x "${BIN_DIR}/spoofdpi" ]]; then
    SPOOFDPI_BIN="${BIN_DIR}/spoofdpi"
elif [[ -x "${HOME}/go/bin/spoofdpi" ]]; then
    SPOOFDPI_BIN="${HOME}/go/bin/spoofdpi"
elif command -v go &>/dev/null; then
    log_step "Installing SpoofDPI via 'go install'..."
    go install github.com/xvzc/spoofdpi/cmd/spoofdpi@latest
    SPOOFDPI_BIN="${HOME}/go/bin/spoofdpi"
else
    log_step "Go not detected; downloading official static SpoofDPI release..."
    SPOOF_TMP="$(mktemp -d)"
    if curl -sSL "https://github.com/xvzc/spoofdpi/releases/download/v1.5.3/spoofdpi_1.5.3_linux_x86_64.tar.gz" -o "${SPOOF_TMP}/spoofdpi.tar.gz" 2>/dev/null; then
        tar -xzf "${SPOOF_TMP}/spoofdpi.tar.gz" -C "${BIN_DIR}"
        SPOOFDPI_BIN="${BIN_DIR}/spoofdpi"
        rm -rf "${SPOOF_TMP}"
    fi
fi

if [[ -z "${SPOOFDPI_BIN}" || ! -x "${SPOOFDPI_BIN}" ]]; then
    log_err "Failed to locate or install SpoofDPI. Install Go or download spoofdpi to ~/.local/bin/spoofdpi."
    exit 1
fi
log_ok "SpoofDPI binary located at: ${SPOOFDPI_BIN}"

# 4. Configure CAP_NET_RAW Capability
log_step "Ensuring CAP_NET_RAW capability on SpoofDPI for user-space decoy packets..."
CURRENT_CAPS=$(getcap "${SPOOFDPI_BIN}" 2>/dev/null || true)
if [[ "${CURRENT_CAPS}" != *"cap_net_raw"* ]]; then
    echo "       Requesting sudo to set: setcap cap_net_raw+ep ${SPOOFDPI_BIN}"
    sudo setcap cap_net_raw+ep "${SPOOFDPI_BIN}"
    log_ok "Granted CAP_NET_RAW to ${SPOOFDPI_BIN}"
else
    log_ok "CAP_NET_RAW is already configured."
fi

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
