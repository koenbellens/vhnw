#!/bin/sh
# VHNW kiosk: toon het dashboard schermvullend op het laptopscherm.

# Log de kiosk-start naar het home van de gebruiker (voor debuggen via SSH).
exec >"$HOME/kiosk.log" 2>&1
echo "=== kiosk gestart $(date) ==="

# Geen schermbeveiliging / energiebesparing: het scherm moet altijd aanblijven.
xset s off
xset -dpms
xset s noblank

# Verberg de muiscursor zodra die stilstaat.
unclutter -idle 0 &

# Minimale window-manager (zorgt voor nette fullscreen + focus).
openbox &

# Wacht tot de agent draait en de kiosk-pagina serveert.
KIOSK_URL="http://localhost:8420/kiosk"
i=0
while ! curl -fsS -o /dev/null "$KIOSK_URL" 2>/dev/null; do
	i=$((i + 1))
	[ "$i" -ge 120 ] && break
	sleep 1
done
echo "agent bereikbaar na ${i}s"

# Chromium in kiosk-modus. In een VM/oude laptop zonder GPU-versnelling moet de
# software-renderer gebruikt worden, anders blijft het scherm zwart.
# --no-sandbox: single-purpose miner die als de live-gebruiker draait.
while true; do
	echo "--- chromium start $(date) ---"
	chromium \
		--kiosk \
		--noerrdialogs \
		--disable-infobars \
		--no-first-run \
		--disable-translate \
		--disable-features=Translate,TranslateUI \
		--lang=nl \
		--disable-session-crashed-bubble \
		--disable-component-update \
		--check-for-update-interval=31536000 \
		--incognito \
		--no-sandbox \
		--disable-gpu \
		--disable-dev-shm-usage \
		"$KIOSK_URL"
	echo "--- chromium afgesloten $(date), opnieuw over 2s ---"
	sleep 2
done
