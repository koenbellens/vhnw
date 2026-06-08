# Laag 3 — VHNW community-hub

De **hub** is de centrale server van het community-dashboard. Elk mining-apparaat
(Pi, laptop, later ASIC of GPU) stuurt periodiek een korte status ("heartbeat")
naar de hub. De hub bewaart de laatste stand per apparaat en toont alles samen op
één live pagina.

```
  Pi-agent ─┐
  laptop-agent ─┼──POST /api/report──▶  vhnw-hub  ──▶  dashboard (browser)
  ASIC-relay ─┘                         (verzamelt + toont)
```

## Toekomstvast: apparaat-types

Elke heartbeat draagt een **type**. Daardoor passen nieuwe soorten apparaten er
vanzelf in, zonder de hub te wijzigen:

- `cpu-agent` — onze gewone agent (Pi, laptop) met XMRig.
- `asic` — een Antminer e.d. Een kleine "relay" leest de Antminer-API uit en
  stuurt dezelfde heartbeat met `type: "asic"`.
- `gpu` — latere GPU-editie.

Onbekende types worden door de hub als `cpu-agent` behandeld.

## Heartbeat-contract (`POST /api/report`)

Headers: `Content-Type: application/json` en — als de hub een token heeft —
`Authorization: Bearer <token>`.

```json
{
  "device_id": "vhnw-ab12cd34...",   // stabiele id; agent vult dit zelf in
  "name": "Pi 5 (woonkamer)",         // weergavenaam (geen wallet!)
  "type": "cpu-agent",                // cpu-agent | asic | gpu
  "miner": "XMRig 6.21.3",
  "algo": "rx",
  "pool": "rx.unmineable.com:3333",
  "state": "actief",
  "hashrate": 372.0,                  // H/s
  "shares_good": 1978,
  "shares_total": 2040,
  "temp_c": 56.8,
  "uptime_secs": 504000,
  "restarts": 0,
  "agent_version": "v1"
}
```

De agent stuurt **geen** wallet, worker of andere gevoelige gegevens — privacy by
design.

## Endpoints

| Endpoint        | Methode | Doel                                              |
|-----------------|---------|---------------------------------------------------|
| `/`             | GET     | Het dashboard (HTML).                             |
| `/api/devices`  | GET     | Alle apparaten + totalen als JSON.                |
| `/api/report`   | POST    | Eén heartbeat van een apparaat.                   |
| `/healthz`      | GET     | Health-check.                                     |

## Hub draaien

```sh
go build -o vhnw-hub ./cmd/vhnw-hub
./vhnw-hub --addr :8088 --data hub-data.json --token "GEHEIM" --offline 60
```

Instellingen kunnen ook via omgevingsvariabelen: `PORT`, `VHNW_HUB_DATA`,
`VHNW_HUB_TOKEN`, `VHNW_HUB_OFFLINE`. Laat het token leeg om elke melding te
accepteren (handig om lokaal te testen, niet voor productie).

Een apparaat staat **offline** zodra het langer dan `--offline` seconden niets
meer heeft gestuurd. Offline apparaten tellen niet mee in de totale hashrate.

## Agent koppelen

Zet in de agent-config het `community`-blok aan (zie `config.example.json`):

```json
"community": {
  "enabled": true,
  "hub_url": "https://hub.vanhashnaarwinst.nl",
  "device_name": "Pi 5 (woonkamer)",
  "device_type": "cpu-agent",
  "token": "GEHEIM",
  "interval_seconds": 15
}
```

`device_id` laat je leeg: de agent genereert er bij de eerste start zelf één en
bewaart die in de config.
