#!/bin/sh
# Zet een bekend wachtwoord voor de live-gebruiker (en root) zodat SSH-login
# werkt om te debuggen. Draait bij het opstarten; wacht tot live-config de
# gebruiker 'vhnw' heeft aangemaakt (anders mislukt chpasswd stil).
i=0
while ! id vhnw >/dev/null 2>&1; do
	i=$((i + 1))
	[ "$i" -ge 60 ] && break
	sleep 1
done
echo "vhnw:minenmaar" | chpasswd || true
echo "root:minenmaar" | chpasswd || true
