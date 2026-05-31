# VHNW op een Raspberry Pi 5 (64-bit)

Deze gids zet de VHNW miner-agent + XMRig op een verse **Raspberry Pi OS Lite (64-bit)**.
De agent houdt XMRig draaiend (watchdog → valt nooit meer stil) en toont een web-UI met
een grote start/stop-knop + live stats.

## Snelste weg: install-script

```bash
# op de Pi, na een verse install:
sudo apt-get update && sudo apt-get install -y git
git clone https://github.com/koenbellens/vhnw.git
cd vhnw

# bouw de agent (Go nodig) OF zet de meegeleverde ARM64-binary in bin/vhnw-agent
# -- zie "Agent-binary" hieronder --

sudo ./scripts/install-pi.sh
```

Het script installeert de build-tools, bouwt XMRig, plaatst de agent en start de
systemd-service. Daarna:

```bash
sudo nano /etc/vhnw/config.json      # vul je wallet in
sudo systemctl restart vhnw-agent
```

Web-UI: `http://<pi-ip>:8420` (open op je telefoon of laptop in hetzelfde netwerk).
Het IP zie je met `hostname -I`.

## Agent-binary

Je hebt twee opties:

**A. Meegeleverde ARM64-binary gebruiken (geen Go nodig)**
Kopieer `vhnw-agent-linux-arm64` naar de Pi en zet 'm op de juiste plek:
```bash
mkdir -p ~/vhnw/bin
mv vhnw-agent-linux-arm64 ~/vhnw/bin/vhnw-agent
chmod +x ~/vhnw/bin/vhnw-agent
```

**B. Zelf bouwen op de Pi (Go vereist)**
```bash
sudo apt-get install -y golang   # of installeer Go 1.26 handmatig
cd vhnw
go build -o bin/vhnw-agent ./cmd/vhnw-agent
```

## Config

`/etc/vhnw/config.json` (door het script aangemaakt). Belangrijkste velden:

| Veld | Betekenis | Voorbeeld |
|------|-----------|-----------|
| `wallet` | unMineable: `COIN:ADRES` | `BTC:bc1q...` |
| `worker` | naam van dit apparaat | `pi5-woonkamer` |
| `pool` | unMineable RandomX-pool | `rx.unmineable.com:3333` |
| `algo` | algoritme | `rx` |
| `miner_path` | pad naar xmrig | `/usr/local/bin/xmrig` |
| `auto_start` | meteen minen bij boot | `true` |
| `threads` | 0 = automatisch | `0` |

> De wallet komt **nooit** in de repo — alleen in `/etc/vhnw/config.json` op de Pi zelf.

## Handige commando's

```bash
systemctl status vhnw-agent      # draait hij?
journalctl -u vhnw-agent -f      # live logs
sudo systemctl restart vhnw-agent
sudo systemctl stop vhnw-agent
```

## Temperatuur op de Pi

De Pi 5 wordt warm bij volledige load. Een koelblok/ventilator (of de officiele
Active Cooler) wordt sterk aangeraden. De web-UI toont de CPU-temperatuur; bij
~80°C gaat de Pi vanzelf throttlen (hashrate zakt). Houd 'm onder ~70°C voor de
beste, stabiele hashrate.

## Verwachte opbrengst

Een Pi 5 haalt ruwweg een paar honderd tot ~1000 H/s op RandomX. Dat is — net als
op de oude telefoons — vooral leuk/educatief; de winst is klein. De waarde zit in
het systeem dat **stabiel blijft draaien** en het overzicht.
