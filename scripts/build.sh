#!/bin/bash

echo "╔════════════════════════════════════════════════════════╗"
echo "║     🔨 BUILD DU HONEYPOT                             ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Couleurs
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Créer le dossier bin s'il n'existe pas
mkdir -p bin

echo -e "${YELLOW}1. Compilation du tester...${NC}"
if go build -o bin/full-tester cmd/tester/main.go 2>&1; then
    echo -e "${GREEN}✓ Tester compilé : bin/full-tester${NC}"
else
    echo -e "${RED}✗ Erreur de compilation du tester${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}2. Build de l'image Docker...${NC}"
if docker-compose build 2>&1 | tail -5; then
    echo -e "${GREEN}✓ Image Docker construite${NC}"
else
    echo -e "${RED}✗ Erreur de build Docker${NC}"
    exit 1
fi

echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║     ✅ BUILD TERMINÉ !                                ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""
echo -e "${GREEN}Prochaines étapes :${NC}"
echo ""
echo "  1. Lancer le honeypot :"
echo "     docker-compose up -d"
echo ""
echo "  2. Tester les services :"
echo "     ./scripts/test_all_services.sh"
echo ""
echo "  3. Lancer les tests complets :"
echo "     ./bin/full-tester"
echo ""
echo "  4. Voir le dashboard :"
echo "     http://localhost:8080"
echo ""

