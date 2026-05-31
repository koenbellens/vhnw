# VHNW — Van Hash Naar Winst

Een crypto-mining-project naast de [VBNW](https://vbnw.nl)-community ("Van Backtest
Naar Winst"). Het doel: samen leren minen, en het betrouwbaar én overzichtelijk
maken op oude hardware (oude laptops, telefoons, een Raspberry Pi).

Dit project lost niet het *minen zelf* op — daarvoor gebruiken we
[XMRig](https://xmrig.com), de beste RandomX-miner die er is. Het lost de
problemen eromheen op:

- **Miners die stilvallen** (Pi crasht na dagen, telefoon stopt na uren).
- **Geen overzicht** van hashrate, shares, temperatuur en uptime.

## Onderdelen

Dit is **Laag 1** van een groter plan:

1. **Miner-agent + 1-knop web-UI** *(deze repo, nu)* — een klein Go-programma dat
   op een Linux-apparaat draait, XMRig beheert met een **watchdog** (herstart
   automatisch als hij stilvalt) en een webpagina toont met een grote
   **start/stop-knop** plus live statistieken.
2. **Eigen mining-OS** *(later)* — deze agent inbakken in een minimale
   Linux-image, zodat een oude laptop direct opstart in het mining-systeem.
3. **Community-dashboard** *(later)* — alle apparaten sturen hun status naar een
   centrale plek voor een live overzicht van de hele community.

## Wat de agent doet

- Start/stopt XMRig met één knop in de browser.
- **Watchdog**: herstart XMRig automatisch als het proces crasht *of* als het
  vastloopt (de ingebouwde HTTP-API van XMRig reageert niet meer).
- Live overzicht: hashrate (nu / 15m), geaccepteerde shares, mining-tijd,
  CPU-temperatuur en aantal herstarts.
- Instellingen (wallet, pool, worker, threads, TLS) aanpasbaar via de UI.

## Snel starten

### 1. Bouw de agent

```bash
go build -o bin/vhnw-agent ./cmd/vhnw-agent
```

> Vereist [Go](https://go.dev/dl/) 1.26 of nieuwer. Voor een Raspberry Pi bouw je
> bijvoorbeeld met `GOOS=linux GOARCH=arm64 go build -o bin/vhnw-agent ./cmd/vhnw-agent`.

### 2. Installeer XMRig

Download XMRig voor jouw platform via <https://xmrig.com/download> en onthoud het
pad naar het binary (of zet het in je `PATH`).

### 3. Maak je config

```bash
cp config.example.json config.json
```

Vul minimaal je `wallet` in. Voor [unMineable](https://unmineable.com) is de
`wallet` van de vorm `COIN:ADRES`, bijvoorbeeld `BTC:bc1q...` of `DOGE:D...`.
De pool-username wordt automatisch `wallet.worker#referral`.

| Veld | Uitleg |
|------|--------|
| `miner_path` | pad naar het XMRig-binary (of `xmrig` als het in PATH staat) |
| `pool` | pool-adres, standaard `rx.unmineable.com:3333` |
| `algo` | algoritme, voor unMineable RandomX is dit `rx` |
| `wallet` | `COIN:ADRES` waarop je uitbetaald wilt worden |
| `worker` | naam van dit apparaat (zichtbaar in je pool-overzicht) |
| `referral` | optionele unMineable referral-code |
| `threads` | aantal CPU-threads (0 = XMRig kiest automatisch) |
| `tls` | versleutelde verbinding naar de pool |
| `auto_start` | meteen beginnen met minen bij het opstarten van de agent |
| `health_grace_seconds` | hoe lang XMRig stil mag zijn voordat de watchdog herstart |

### 4. Start de agent

```bash
./bin/vhnw-agent -config config.json
```

Open daarna <http://localhost:8420> (of het IP van het apparaat) en druk op de
knop.

## Als achtergronddienst (systemd)

Zie [`deploy/vhnw-agent.service`](deploy/vhnw-agent.service) voor een kant-en-klare
systemd-unit, zodat de agent automatisch opstart en blijft draaien.

## Raspberry Pi 5

Voor een verse Raspberry Pi OS Lite (64-bit) is er een installatiegids +
install-script dat XMRig bouwt, de agent plaatst en de systemd-service start:
zie [`docs/PI-SETUP.md`](docs/PI-SETUP.md) en
[`scripts/install-pi.sh`](scripts/install-pi.sh).

## Testen zonder echt te minen

Voor ontwikkeling zit er een **mock-miner** in die XMRig nabootst (inclusief de
`/2/summary` API) zonder je CPU te belasten of echt te minen:

```bash
go build -o bin/mockminer ./cmd/mockminer
# zet in je config "miner_path" op het pad naar bin/mockminer
```

De mock kent extra testvlaggen via `extra_args`:

- `["--crash-after","5s"]` — simuleert een crash (test de herstart-watchdog).
- `["--hang-after","3s"]` — bevriest de API (test de hang-watchdog).

## Projectstructuur

```
cmd/vhnw-agent   de agent (binary)
cmd/mockminer    test-miner die XMRig nabootst
internal/config  laden/opslaan van de configuratie
internal/miner   procesbeheer + watchdog
internal/xmrig   client voor de XMRig HTTP-API
internal/system  systeeminfo (temperatuur, uptime)
internal/web     web-UI en JSON-API
deploy           systemd-unit
```

## Veiligheid

- Je `config.json` bevat je wallet-adres en staat daarom in `.gitignore`.
- De agent mined alleen wanneer **jij** op start drukt (tenzij `auto_start` aan staat).
- Verdiensten op oude hardware zijn klein — dit is vooral leuk en leerzaam.
