#!/bin/sh
# DEBUG: dump kiosk-/X-status naar de seriële poort (host vangt dit op).
exec > /dev/ttyS0 2>&1
n=0
while [ "$n" -lt 5 ]; do
	n=$((n + 1))
	echo "########## VHNW DEBUG DUMP #$n $(date) ##########"
	echo "--- is-active agent/ssh/getty@tty1 ---"
	systemctl is-active vhnw-agent.service ssh.service getty@tty1.service
	echo "--- who ---"; who
	echo "--- ps (login/X/chromium) ---"
	ps -ef | grep -iE 'agetty|startx|xinit|Xorg|chromium|openbox|/bin/login' | grep -v grep
	echo "--- /dev/dri ---"; ls -l /dev/dri 2>&1
	echo "--- Xorg.0.log tail ---"; tail -40 /var/log/Xorg.0.log 2>&1
	echo "--- ~vhnw/.xsession-errors ---"; tail -30 /home/vhnw/.xsession-errors 2>&1
	echo "--- ~vhnw/kiosk.log ---"; cat /home/vhnw/kiosk.log 2>&1
	echo "--- id vhnw / shell ---"; id vhnw 2>&1; getent passwd vhnw
	echo "########## END DUMP #$n ##########"
	sleep 20
done
