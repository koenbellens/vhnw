# Start de grafische kiosk automatisch bij inloggen op tty1 (het laptopscherm).
# Op andere tty's / via SSH gebeurt dit niet, zodat debuggen gewoon kan.
if [ -z "$DISPLAY" ] && [ "$(tty)" = "/dev/tty1" ]; then
	exec startx -- -nocursor
fi
