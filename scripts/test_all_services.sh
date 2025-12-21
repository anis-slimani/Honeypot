#!/bin/bash

# Script de test rapide pour tous les services du honeypot
# Teste SSH, HTTP, FTP et les alertes email

echo "╔════════════════════════════════════════════════════════╗"
echo "║     🧪 TEST COMPLET DU HONEYPOT                      ║"
echo "║     SSH + HTTP + FTP + Email Alerts                   ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SSH_HOST="${1:-localhost}"
SSH_PORT="${2:-2222}"
FTP_HOST="${1:-localhost}"
FTP_PORT="${3:-2121}"
HTTP_URL="http://${1:-localhost}"
DASHBOARD_URL="http://${1:-localhost}:8080"

echo -e "${BLUE}Configuration:${NC}"
echo "  SSH:       $SSH_HOST:$SSH_PORT"
echo "  FTP:       $FTP_HOST:$FTP_PORT"
echo "  HTTP:      $HTTP_URL"
echo "  Dashboard: $DASHBOARD_URL"
echo ""

# Vérifier que le honeypot est lancé
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Vérification des services${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Test SSH
echo -n "  SSH ($SSH_PORT)...  "
if nc -z -w2 $SSH_HOST $SSH_PORT 2>/dev/null; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ OFFLINE${NC}"
fi

# Test FTP
echo -n "  FTP ($FTP_PORT)...  "
if nc -z -w2 $FTP_HOST $FTP_PORT 2>/dev/null; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ OFFLINE${NC}"
fi

# Test HTTP
echo -n "  HTTP (80)...    "
if curl -s -o /dev/null -w "%{http_code}" $HTTP_URL | grep -q "200\|404"; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ OFFLINE${NC}"
fi

# Test Dashboard
echo -n "  Dashboard...    "
if curl -s -o /dev/null -w "%{http_code}" $DASHBOARD_URL | grep -q "200"; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ OFFLINE${NC}"
fi

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Test des API du Dashboard${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Test API SSH
echo -n "  /api/connections...     "
if curl -s "$DASHBOARD_URL/api/connections?limit=1" | jq . >/dev/null 2>&1; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAIL${NC}"
fi

# Test API FTP
echo -n "  /api/ftp/connections... "
if curl -s "$DASHBOARD_URL/api/ftp/connections?limit=1" | jq . >/dev/null 2>&1; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAIL${NC}"
fi

# Test API HTTP
echo -n "  /api/http/requests...   "
if curl -s "$DASHBOARD_URL/api/http/requests?limit=1" | jq . >/dev/null 2>&1; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAIL${NC}"
fi

# Test API Alerts
echo -n "  /api/alerts...          "
if curl -s "$DASHBOARD_URL/api/alerts?limit=1" | jq . >/dev/null 2>&1; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAIL${NC}"
fi

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Test Rapide FTP${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Test FTP basique
echo "  Connexion FTP et test de commandes..."
(
    sleep 1
    echo "USER admin"
    sleep 0.5
    echo "PASS admin123"
    sleep 0.5
    echo "PWD"
    sleep 0.5
    echo "SYST"
    sleep 0.5
    echo "QUIT"
) | telnet $FTP_HOST $FTP_PORT 2>/dev/null | grep -q "230\|257\|215" && echo -e "  ${GREEN}✓ FTP fonctionne correctement${NC}" || echo -e "  ${RED}✗ Problème FTP${NC}"

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Statistiques${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Statistiques SSH
echo -e "${BLUE}SSH:${NC}"
SSH_STATS=$(curl -s "$DASHBOARD_URL/api/statistics" | jq -r '"\(.total_connections) connexions, \(.successful_logins) réussies"' 2>/dev/null)
echo "  $SSH_STATS"

# Statistiques FTP
echo -e "${BLUE}FTP:${NC}"
FTP_STATS=$(curl -s "$DASHBOARD_URL/api/ftp/statistics" | jq -r '"\(.total_connections) connexions, \(.authenticated_logins) authentifiées"' 2>/dev/null)
echo "  $FTP_STATS"

# Statistiques HTTP
echo -e "${BLUE}HTTP:${NC}"
HTTP_STATS=$(curl -s "$DASHBOARD_URL/api/http/statistics" | jq -r '"\(.total_requests) requêtes, \(.total_attacks) attaques détectées"' 2>/dev/null)
echo "  $HTTP_STATS"

# Alertes
echo -e "${BLUE}Alertes:${NC}"
ALERT_COUNT=$(curl -s "$DASHBOARD_URL/api/alerts?limit=100" | jq '. | length' 2>/dev/null)
echo "  $ALERT_COUNT alertes enregistrées"

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Tests terminés !${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}Pour lancer les tests complets :${NC}"
echo "  ./bin/full-tester"
echo ""
echo -e "${BLUE}Pour tester uniquement FTP :${NC}"
echo "  ./bin/full-tester --ftp-only"
echo ""
echo -e "${BLUE}Pour voir le dashboard :${NC}"
echo "  $DASHBOARD_URL"
echo ""

