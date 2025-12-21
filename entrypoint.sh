#!/bin/sh

# Copier le tester vers l'hôte si le volume est monté
if [ -d "/host" ] && [ -f "/root/tester" ]; then
    cp /root/tester /host/tester
    chmod +x /host/tester
fi

# Lancer le honeypot
exec ./honeypot -config config.yaml

