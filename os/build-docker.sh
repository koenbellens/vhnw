#!/usr/bin/env bash
# Build the VHNW Mining-OS ISO inside a Debian bookworm container.
#
# This avoids version mismatches between the host's live-build and the target
# Debian release. Requires Docker. The agent binary is staged from
# dist/vhnw-agent-linux-amd64 (build it first with:
#   GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/vhnw-agent-linux-amd64 ./cmd/vhnw-agent )
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/.." && pwd)"

# Build the agent on the host (the container has no Go toolchain). build.sh
# inside the container picks up dist/vhnw-agent-linux-amd64 automatically.
if command -v go >/dev/null 2>&1; then
	echo "==> Pre-building vhnw-agent (linux/amd64) on host"
	( cd "$REPO" && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -o dist/vhnw-agent-linux-amd64 ./cmd/vhnw-agent )
elif [ ! -f "$REPO/dist/vhnw-agent-linux-amd64" ]; then
	echo "ERROR: Go not found and dist/vhnw-agent-linux-amd64 missing." >&2
	echo "Install Go, or pre-build the agent, then re-run." >&2
	exit 1
fi

docker run --rm --privileged \
	-v "$REPO":/work -w /work \
	debian:bookworm \
	bash -c '
		set -euo pipefail
		export DEBIAN_FRONTEND=noninteractive
		apt-get update -qq
		apt-get install -y --no-install-recommends \
			live-build debootstrap xorriso squashfs-tools dosfstools \
			curl ca-certificates >/dev/null
		bash os/build.sh
	'
