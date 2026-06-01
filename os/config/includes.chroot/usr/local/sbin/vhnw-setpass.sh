#!/bin/sh
# Sets a known password for the live user so SSH login works for debugging.
# Runs once at boot (after live-config has created the user).
echo "vhnw:minenmaar" | chpasswd 2>/dev/null || true
