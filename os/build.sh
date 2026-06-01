#!/usr/bin/env bash
# Build the VHNW Mining-OS live ISO.
#
# Produces a bootable, USB-flashable ISO that auto-starts the vhnw-agent so an
# old laptop becomes a dedicated miner on boot.
#
# Requirements (Debian/Ubuntu): live-build debootstrap xorriso squashfs-tools
#                               dosfstools  (and Go to build the agent)
# Usage:   sudo ./os/build.sh
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/.." && pwd)"
XMRIG_VER="6.21.3"
XMRIG_URL="https://github.com/xmrig/xmrig/releases/download/v${XMRIG_VER}/xmrig-${XMRIG_VER}-linux-static-x64.tar.gz"

BIN_DST="$HERE/config/includes.chroot/usr/local/bin"
mkdir -p "$BIN_DST" "$HERE/downloads"

echo "==> [1/4] Building vhnw-agent (linux/amd64)"
if command -v go >/dev/null 2>&1; then
	( cd "$REPO" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$BIN_DST/vhnw-agent" ./cmd/vhnw-agent )
elif [ -f "$REPO/dist/vhnw-agent-linux-amd64" ]; then
	cp "$REPO/dist/vhnw-agent-linux-amd64" "$BIN_DST/vhnw-agent"
else
	echo "ERROR: Go not found and dist/vhnw-agent-linux-amd64 missing." >&2
	exit 1
fi
chmod +x "$BIN_DST/vhnw-agent"

echo "==> [2/4] Fetching XMRig ${XMRIG_VER} (static x64)"
if [ ! -f "$BIN_DST/xmrig" ]; then
	curl -fsSL -o "$HERE/downloads/xmrig.tar.gz" "$XMRIG_URL"
	tar xzf "$HERE/downloads/xmrig.tar.gz" -C "$HERE/downloads"
	cp "$HERE/downloads/xmrig-${XMRIG_VER}/xmrig" "$BIN_DST/xmrig"
fi
chmod +x "$BIN_DST/xmrig"

echo "==> [3/4] Configuring live-build"
cd "$HERE"
lb clean >/dev/null 2>&1 || true
./auto/config

echo "==> [4/4] Building ISO (this takes a while; needs root + network)"
lb build

ISO="$(ls -1 "$HERE"/*.iso 2>/dev/null | head -1 || true)"
if [ -n "$ISO" ]; then
	OUT="$HERE/vhnw-mining-os-amd64.iso"
	mv "$ISO" "$OUT"
	echo "==> Done: $OUT"
	ls -lh "$OUT"
else
	echo "ERROR: no ISO produced." >&2
	exit 1
fi
