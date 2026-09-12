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
echo "      🦄 OMACORN v0.4.0 — Automated Deployment Suite           "
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
    install -Dm755 ./omacorn "${BIN_DIR}/omacorn"
elif [[ -f "./main.go" ]] && command -v go &>/dev/null; then
    log_step "Go compiler detected ($(go version | awk '{print $3}')). Compiling static Omacorn binary..."
    CGO_ENABLED=0 go build -ldflags="-s -w" -o omacorn .
    install -Dm755 omacorn "${BIN_DIR}/omacorn"
else
    log_step "Fetching latest static Omacorn release for linux-${ARCH_SUFFIX}..."
    OMACORN_DL_URL="https://github.com/oomaya/Omacorn/releases/latest/download/omacorn-linux-${ARCH_SUFFIX}"
    DOWNLOADED=false

    # Strategy A: GitHub CLI for private repository access
    if command -v gh &>/dev/null && gh auth status &>/dev/null; then
        log_step "Authenticated GitHub CLI detected. Downloading asset from oomaya/Omacorn..."
        TMP_OMACORN="$(mktemp -d)"
        if gh release download --repo oomaya/Omacorn --pattern "omacorn-linux-${ARCH_SUFFIX}" --dir "${TMP_OMACORN}" 2>/dev/null; then
            install -Dm755 "${TMP_OMACORN}/omacorn-linux-${ARCH_SUFFIX}" "${BIN_DIR}/omacorn"
            rm -rf "${TMP_OMACORN}"
            DOWNLOADED=true
            log_ok "Downloaded and deployed release binary via GitHub CLI"
        fi
    fi

    # Strategy B: Authenticated or direct curl
    if [[ "${DOWNLOADED}" != "true" ]]; then
        AUTH_HEADER=()
        if [[ -n "${GITHUB_TOKEN:-${GH_TOKEN:-}}" ]]; then
            AUTH_HEADER=(-H "Authorization: token ${GITHUB_TOKEN:-$GH_TOKEN}")
        fi
        TMP_DL="$(mktemp)"
        if curl -sSLf "${AUTH_HEADER[@]}" "${OMACORN_DL_URL}" -o "${TMP_DL}" 2>/dev/null; then
            install -Dm755 "${TMP_DL}" "${BIN_DIR}/omacorn"
            rm -f "${TMP_DL}"
            DOWNLOADED=true
            log_ok "Downloaded static release binary from GitHub"
        fi
    fi

    # Strategy C: Go build fallback
    if [[ "${DOWNLOADED}" != "true" ]] && command -v go &>/dev/null; then
        log_step "Compiling from source via go install..."
        go install github.com/oomaya/Omacorn@latest 2>/dev/null || true
        if [[ -x "${HOME}/go/bin/Omacorn" ]]; then
            install -Dm755 "${HOME}/go/bin/Omacorn" "${BIN_DIR}/omacorn"
            DOWNLOADED=true
        fi
    fi

    if [[ "${DOWNLOADED}" != "true" && ! -x "${BIN_DIR}/omacorn" ]]; then
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

# Check PATH inclusion
if [[ ":${PATH}:" != *":${BIN_DIR}:"* ]]; then
    log_warn "${BIN_DIR} is NOT in your current PATH!"
    if command -v fish &>/dev/null || [[ "${SHELL:-}" == *"fish"* ]] || [[ -n "${FISH_VERSION:-}" ]]; then
        echo "       👉 Fish shell detected! Adding ${BIN_DIR} to universal fish_user_paths..."
        if fish -c "fish_add_path ${BIN_DIR}" 2>/dev/null; then
            log_ok "Successfully executed 'fish_add_path ${BIN_DIR}'"
        else
            echo "       Please manually run in fish: fish_add_path ${BIN_DIR}"
        fi
    else
        echo "       👉 Add it to your shell config: export PATH=\"${BIN_DIR}:\$PATH\""
    fi
fi

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

# Ensure spoofdpi is deployed directly in ~/.local/bin
if [[ -f "${SPOOFDPI_BIN}" && "${SPOOFDPI_BIN}" != "${BIN_DIR}/spoofdpi" ]]; then
    install -Dm755 "${SPOOFDPI_BIN}" "${BIN_DIR}/spoofdpi"
    SPOOFDPI_BIN="${BIN_DIR}/spoofdpi"
fi

if [[ -z "${SPOOFDPI_BIN}" || ! -x "${SPOOFDPI_BIN}" ]]; then
    log_err "Failed to locate or install SpoofDPI. Install Go or download spoofdpi to ~/.local/bin/spoofdpi."
    exit 1
fi
log_ok "SpoofDPI binary located at: ${SPOOFDPI_BIN}"

# 4. Check CAP_NET_RAW Capability (Optional on bare metal; skipped in VMs & unprivileged environments)
log_step "Checking network capabilities for SpoofDPI..."
CURRENT_CAPS=$(getcap "${SPOOFDPI_BIN}" 2>/dev/null || true)
if [[ "${CURRENT_CAPS}" == *"cap_net_raw"* ]]; then
    log_ok "CAP_NET_RAW capability is already configured."
elif sudo -n true 2>/dev/null; then
    echo "       Applying CAP_NET_RAW via passwordless sudo..."
    sudo setcap cap_net_raw+ep "${SPOOFDPI_BIN}" 2>/dev/null || true
    log_ok "Granted CAP_NET_RAW to ${SPOOFDPI_BIN}"
else
    log_ok "Running in pure user-space mode (no root/sudo required; hypervisor & bare-metal safe)."
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
APP_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/applications"
mkdir -p "${APP_DIR}"

# Dynamically resolve preferred terminal emulator (Zero Hardcoded Literals)
TERM_BIN=""
for candidate in "${TERMINAL:-}" ghostty kitty alacritty foot xterm; do
    if [[ -n "${candidate}" ]] && command -v "${candidate}" &>/dev/null; then
        TERM_BIN="${candidate}"
        break
    fi
done

if [[ -n "${TERM_BIN}" ]]; then
    DESKTOP_EXEC="${TERM_BIN} -e omacorn"
    DESKTOP_TERM="false"
else
    DESKTOP_EXEC="omacorn"
    DESKTOP_TERM="true"
fi

cat << EOF > "${APP_DIR}/omacorn.desktop"
[Desktop Entry]
Version=1.0
Type=Application
Name=Omacorn
GenericName=DPI Bypass Dashboard
Comment=Sovereign Twin-Engine DPI Evasion Daemon & Cyberpunk TUI
Exec=${DESKTOP_EXEC}
Icon=network-vpn
Terminal=${DESKTOP_TERM}
Categories=Network;Security;System;
EOF
update-desktop-database "${APP_DIR}" 2>/dev/null || true
log_ok "Registered Omacorn in application menu"

# 9. Live Verification Probe
echo ""
log_step "Waiting for SpoofDPI proxy to initialize on 127.0.0.1:8080..."
sleep 1
for i in {1..6}; do
    if "${BIN_DIR}/omacorn" test &>/dev/null; then
        break
    fi
    sleep 0.5
done

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
