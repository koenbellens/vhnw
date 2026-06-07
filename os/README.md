# VHNW Mining-OS (Laag 2)

Een minimale, bootbare Debian live-image die een **oude laptop in een
dedicated miner** verandert. USB erin → laptop aan → de `vhnw-agent` start
automatisch en begint te minen. De harde schijf van de laptop wordt **niet**
aangeraakt (alles draait vanaf de USB).

## Wat zit erin
- **vhnw-agent** (de supervisor + web-UI uit Laag 1), `linux/amd64`
- **XMRig** (statische x86_64-binary)
- **systemd auto-start** (`vhnw-agent.service`, `Restart=always`)
- Ethernet-DHCP via `systemd-networkd`, SSH-server aan
- Standaard-config met de unMineable-alias al ingevuld + `auto_start: true`

## Bouwen (op een Debian/Ubuntu-machine)
```bash
sudo apt-get install -y live-build debootstrap xorriso squashfs-tools dosfstools
sudo ./os/build.sh
```
Resultaat: `os/vhnw-mining-os-amd64.iso` (USB-flashbaar hybrid-image).

`build.sh` bouwt de agent (`go build`, amd64), haalt de XMRig-binary op, zet ze
in de image en draait `lb build`.

### Aanrader: bouwen in Docker
Als de `live-build` van je host niet overeenkomt met Debian bookworm (bv. op
Ubuntu), bouw dan in een schone bookworm-container — dit is de geteste route:
```bash
sudo ./os/build-docker.sh
```
Vereist Docker + Go op de host (de agent wordt op de host gebouwd, de rest in
de container).

## Op USB zetten
Flash de ISO met **balenaEtcher** (Windows/Mac/Linux) of **Rufus** (Windows)
naar een USB-stick (≥ 2 GB). Dit wist de USB-stick, niet de laptop.

## Booten op de laptop
1. Steek de USB in de oude laptop, ethernet-kabel erin.
2. Zet aan en kies het USB-opstartmenu (vaak `F12`/`F10`/`Esc`, soms in BIOS
   "Boot Order" → USB eerst). **Secure Boot uitzetten** in de BIOS/UEFI — de
   image is niet ondertekend en boot anders niet.
3. Hij start in VHNW Mining-OS en begint **vanzelf** te minen.
4. Bekijk het dashboard vanaf je telefoon/pc: `http://<laptop-ip>:8420`.

## Inloggen (optioneel, voor debug)
- Gebruiker `vhnw`, wachtwoord `minenmaar` (SSH of lokaal).
- Pas dit aan in `config/includes.chroot/usr/local/sbin/vhnw-setpass.sh`.

## Config aanpassen
De standaard-config staat in
`config/includes.chroot/etc/vhnw/config.json` (wordt in de image gebakken).
Wijzigingen via de web-UI gelden voor de huidige sessie; voor blijvende
wijzigingen pas je dit bestand aan en bouw je de ISO opnieuw. (Persistente
opslag op de USB is een latere uitbreiding.)

## Beperkingen v1
- **Ethernet** voor netwerk (wifi headless instellen kan nog niet).
- Geen kiosk-scherm op de laptop zelf (monitoren gaat via het netwerk).
- Erg oude laptops met weinig RAM kunnen XMRig in "light mode" forceren via
  `extra_args: ["--randomx-mode=light"]` in de config.
