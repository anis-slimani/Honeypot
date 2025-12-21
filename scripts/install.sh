#!/bin/bash

echo "════════════════════════════════════════════════════════"
echo "🍯 INSTALLATION HONEYPOT"
echo "════════════════════════════════════════════════════════"
echo ""

# Vérifier que Go est installé
if ! command -v go &> /dev/null; then
    echo "❌ Go n'est pas installé !"
    echo "Installez Go avec : sudo apt install golang"
    exit 1
fi

echo "✓ Go version : $(go version)"
echo ""

# Créer les dossiers nécessaires
echo "📁 Création des dossiers..."
mkdir -p logs static templates

# Nettoyer et télécharger les dépendances
echo "📦 Installation des dépendances..."
rm -f go.sum
go clean -modcache
go mod tidy

if [ $? -ne 0 ]; then
    echo "❌ Erreur lors de l'installation des dépendances"
    exit 1
fi

# Compiler
echo "🔨 Compilation du honeypot..."
go build -o honeypot .

if [ $? -ne 0 ]; then
    echo "❌ Erreur de compilation"
    exit 1
fi

echo ""
echo "════════════════════════════════════════════════════════"
echo "✅ INSTALLATION RÉUSSIE !"
echo "════════════════════════════════════════════════════════"
echo ""
echo "Taille du binaire : $(ls -lh honeypot | awk '{print $5}')"
echo ""
echo "🚀 Pour lancer le honeypot :"
echo "   sudo ./honeypot -config config.yaml"
echo ""
echo "🌐 Interface web :"
echo "   http://$(hostname -I | awk '{print $1}'):8080"
echo ""
echo "🔍 Vérifier la base de données :"
echo "   ./check_db.sh"
echo ""

