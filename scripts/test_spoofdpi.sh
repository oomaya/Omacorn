#!/usr/bin/env bash
set -euo pipefail

SPOOFDPI_BIN="${HOME}/go/bin/spoofdpi"
PORT=8080
ADDR="127.0.0.1:${PORT}"

if [[ ! -x "${SPOOFDPI_BIN}" ]]; then
    echo "ERROR: ${SPOOFDPI_BIN} not found or not executable."
    exit 1
fi

echo "=== [1/4] Starting SpoofDPI (headless on ${ADDR}) ==="
"${SPOOFDPI_BIN}" \
    --no-tui \
    --listen-addr "${ADDR}" \
    --dns-mode system \
    --log-level warn &
SPOOFDPI_PID=$!

cleanup() {
    echo "=== Cleaning up SpoofDPI (PID: ${SPOOFDPI_PID}) ==="
    kill "${SPOOFDPI_PID}" 2>/dev/null || true
    wait "${SPOOFDPI_PID}" 2>/dev/null || true
}
trap cleanup EXIT

# Wait up to 5 seconds for port to open
echo "=== [2/4] Waiting for port ${PORT} to be active ==="
for i in {1..10}; do
    if ss -lnt | grep -q ":${PORT} "; then
        echo "Port ${PORT} is open and listening."
        break
    fi
    sleep 0.5
done

echo "=== [3/4] Testing Direct vs SpoofDPI Latency & Response ==="
echo -n "Direct: https://accounts.google.com -> "
curl -s -o /dev/null -w "HTTP %{http_code}, Time: %{time_total}s\n" -m 5 https://accounts.google.com

echo -n "Via SpoofDPI: https://accounts.google.com -> "
curl -s -o /dev/null -w "HTTP %{http_code}, Time: %{time_total}s\n" --proxy "http://${ADDR}" -m 5 https://accounts.google.com

echo "=== [4/4] Testing General HTTPS Connectivity ==="
echo -n "Via SpoofDPI: https://www.cloudflare.com -> "
curl -s -o /dev/null -w "HTTP %{http_code}, Time: %{time_total}s\n" --proxy "http://${ADDR}" -m 5 https://www.cloudflare.com

echo "=== SpoofDPI test completed successfully! ==="
