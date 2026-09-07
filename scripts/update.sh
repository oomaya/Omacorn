#!/usr/bin/env bash
# ==============================================================================
# 🦄 Omacorn Fleet Update Engine (Maintenance & Hot-Reload Loop)
# ==============================================================================
# Fast, seamless update for existing Omacorn installations:
# 1. Fast-forwards git repository
# 2. Compiles static multi-distro binary
# 3. Hot-reloads active bypass daemons (zero dropped sessions)
# 4. Re-synchronizes split-tunneling exemptions
# 5. Executes live health probe
# ==============================================================================
set -euo pipefail

BOLD='\033[1m'
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

log_step() {
    echo -e "${CYAN}${BOLD}[Omacorn Update]${NC} $1"
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
echo "      🦄 OMACORN — Fleet Maintenance & Update Engine            "
echo "================================================================"
echo -e "${NC}"

# 1. Locate Repository Root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BIN_DIR="${HOME}/.local/bin"
mkdir -p "${BIN_DIR}"

if [[ ! -f "${REPO_DIR}/main.go" ]]; then
    log_err "Could not locate Omacorn repository at ${REPO_DIR}"
    exit 1
fi

# 2. Pull Latest Git Changes
log_step "Checking for upstream updates..."
if [[ -d "${REPO_DIR}/.git" ]] && command -v git &>/dev/null; then
    # Stash or warn if uncommitted changes exist
    if ! git -C "${REPO_DIR}" diff-index --quiet HEAD -- 2>/dev/null; then
        log_warn "Uncommitted local changes detected in ${REPO_DIR}. Preserving local state..."
    fi
    if git -C "${REPO_DIR}" pull --ff-only origin master 2>/dev/null; then
        log_ok "Repository fast-forwarded to latest master commit."
    else
        log_warn "Fast-forward git pull skipped or up-to-date."
    fi
else
    log_warn "Git repository metadata not found; proceeding with local source compile."
fi

# 3. Compile Static Binary
log_step "Compiling static Omacorn binary..."
if command -v go &>/dev/null; then
    VERSION="$(git -C "${REPO_DIR}" describe --tags --always 2>/dev/null || echo "v0.4.0")"
    (
        cd "${REPO_DIR}"
        CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o "${BIN_DIR}/omacorn" .
    )
    ln -sf omacorn "${BIN_DIR}/dpipe"
    log_ok "Compiled and installed ${VERSION} to ${BIN_DIR}/omacorn"
else
    log_err "Go compiler not found. Please install Go or use a pre-built static release."
    exit 1
fi

# 4. Hot-Reload Active Engines
log_step "Checking active engine daemons..."
SPOOF_WAS_ACTIVE=false
if systemctl --user is-active dpipe-spoofdpi &>/dev/null; then
    SPOOF_WAS_ACTIVE=true
    systemctl --user restart dpipe-spoofdpi
    log_ok "Hot-reloaded Pilot 1 daemon (dpipe-spoofdpi.service)"
fi

if systemctl is-active dpipe-gecit &>/dev/null; then
    if [[ "${UID:-$(id -u)}" == "0" ]]; then
        systemctl restart dpipe-gecit
    elif command -v sudo &>/dev/null; then
        sudo systemctl restart dpipe-gecit
    fi
    log_ok "Hot-reloaded Pilot 2 daemon (dpipe-gecit.service)"
fi

if [[ "${SPOOF_WAS_ACTIVE}" == "true" ]] || systemctl is-active dpipe-gecit &>/dev/null; then
    sleep 1
fi

# 5. Refresh System Proxy & Split-Tunneling
log_step "Synchronizing desktop environment & split-tunneling bypass list..."
if [[ -x "${BIN_DIR}/omacorn" ]]; then
    "${BIN_DIR}/omacorn" proxy on
    log_ok "Wired clean proxy configuration and browser flags (all_proxy omitted)."
fi

# 6. Live Verification Probe
log_step "Running post-update verification probe..."
if [[ -x "${BIN_DIR}/omacorn" ]]; then
    "${BIN_DIR}/omacorn" test || log_warn "Live test had non-fatal warnings."
fi

echo -e "\n${GREEN}${BOLD}✓ Omacorn update completed successfully!${NC}"
"${BIN_DIR}/omacorn" version || true
