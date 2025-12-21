#!/bin/bash

# Copier le tester depuis le conteneur Docker
docker cp honey-ssh-honeypot:/root/tester ./tester 2>/dev/null
chmod +x ./tester 2>/dev/null

echo "✅ Tester prêt : ./tester"

