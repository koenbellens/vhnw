#!/usr/bin/env bash
# VHNW miner-agent installer voor Raspberry Pi OS (64-bit) / Debian aarch64.
#
# Wat dit script doet:
#   1. installeert de build-dependencies voor XMRig
#   2. bouwt XMRig (RandomX) uit de officiele bron
#   3. zet de meegeleverde vhnw-agent binary in /usr/local/bin
#   4. maakt /etc/vhnw/config.json aan (als die nog niet bestaat)
#   5. installeert + start de systemd-service (Restart=always)
#
# Gebruik (op de Pi):
#   chmod +x scripts/install-pi.sh
#   sudo ./scripts/install-pi.sh
#
# Pas daarna je wallet aan:  sudo nano /etc/vhnw/config.json
# en herstart:               sudo systemctl restart vhnw-agent

set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AGENT_BIN="${REPO_DIR}/bin/vhnw-agent"
XMRIG_DIR="/opt/xmrig"

if [[ $EUID -ne 0 ]]; then
  echo "Run dit script met sudo: sudo ./scripts/install-pi.sh" >&2
  exit 1
fi

echo "==> 1/5 build-dependencies installeren"
apt-get update
apt-get install -y git build-essential cmake automake libtool autoconf \
  libuv1-dev libssl-dev libhwloc-dev

echo "==> 2/5 XMRig bouwen (dit duurt op een Pi een paar minuten)"
if [[ ! -x "${XMRIG_DIR}/build/xmrig" ]]; then
  rm -rf "${XMRIG_DIR}"
  git clone https://github.com/xmrig/xmrig.git "${XMRIG_DIR}"
  mkdir -p "${XMRIG_DIR}/build"
  cd "${XMRIG_DIR}/build"
  cmake ..
  make -j"$(nproc)"
else
  echo "    XMRig al gebouwd in ${XMRIG_DIR}/build/xmrig — overslaan"
fi
ln -sf "${XMRIG_DIR}/build/xmrig" /usr/local/bin/xmrig

echo "==> 3/5 vhnw-agent installeren"
if [[ ! -x "${AGENT_BIN}" ]]; then
  echo "    ${AGENT_BIN} niet gevonden." >&2
  echo "    Bouw 'm eerst (go build -o bin/vhnw-agent ./cmd/vhnw-agent) of" >&2
  echo "    kopieer de meegeleverde ARM64-binary naar bin/vhnw-agent." >&2
  exit 1
fi
install -m 0755 "${AGENT_BIN}" /usr/local/bin/vhnw-agent

echo "==> 4/5 config aanmaken"
mkdir -p /etc/vhnw
if [[ ! -f /etc/vhnw/config.json ]]; then
  cp "${REPO_DIR}/config.example.json" /etc/vhnw/config.json
  # XMRig staat nu in PATH:
  sed -i 's#"miner_path": "xmrig"#"miner_path": "/usr/local/bin/xmrig"#' /etc/vhnw/config.json
  echo "    /etc/vhnw/config.json aangemaakt — VUL JE WALLET IN!"
else
  echo "    /etc/vhnw/config.json bestaat al — niet overschreven"
fi

echo "==> 5/5 systemd-service installeren"
cp "${REPO_DIR}/deploy/vhnw-agent.service" /etc/systemd/system/vhnw-agent.service
systemctl daemon-reload
systemctl enable --now vhnw-agent

IP="$(hostname -I | awk '{print $1}')"
echo
echo "Klaar! De agent draait nu."
echo "  - Web-UI:   http://${IP}:8420  (open op je telefoon/laptop in hetzelfde netwerk)"
echo "  - Wallet:   sudo nano /etc/vhnw/config.json   (daarna: sudo systemctl restart vhnw-agent)"
echo "  - Status:   systemctl status vhnw-agent"
echo "  - Logs:     journalctl -u vhnw-agent -f"
