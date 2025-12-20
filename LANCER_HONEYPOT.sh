#!/bin/bash

echo "╔════════════════════════════════════════════════════════╗"
echo "║     🍯 LANCEMENT DU HONEYPOT AVEC GMAIL              ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Arrêter les conteneurs existants
echo "🛑 Arrêt des conteneurs existants..."
docker-compose down

echo ""
echo "🔨 Rebuild de l'image Docker avec la nouvelle config Gmail..."
docker-compose build

echo ""
echo "🚀 Lancement du honeypot..."
docker-compose up -d

echo ""
echo "✅ HONEYPOT LANCÉ !"
echo ""
echo "📊 Dashboard:      http://localhost:8080"
echo "🔌 SSH Honeypot:   localhost:2222"
echo "🌐 HTTP Honeypot:  http://localhost:80"
echo ""
echo "📝 Voir les logs en temps réel:"
echo "   → docker-compose logs -f"
echo ""
echo "📧 Les emails d'alerte seront envoyés à: honeypotprojet@gmail.com"
echo ""

