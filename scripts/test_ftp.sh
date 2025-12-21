#!/bin/bash
# Script de test du honeypot FTP
# Teste diverses attaques et comportements FTP

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

FTP_HOST="${1:-localhost}"
FTP_PORT="${2:-2121}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Test du Honeypot FTP${NC}"
echo -e "${BLUE}  Host: $FTP_HOST:$FTP_PORT${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Fonction pour tester une connexion FTP
test_ftp() {
    local username="$1"
    local password="$2"
    local description="$3"

    echo -e "${YELLOW}Test: $description${NC}"
    echo "  Username: $username"
    echo "  Password: $password"

    # Créer un fichier de commandes FTP
    cat > /tmp/ftp_commands_$$.txt << EOF
user $username
$password
pwd
syst
list
retr test.txt
stor malware.php
dele config.txt
mkd hacker
rmd hacker
quit
EOF

    # Exécuter les commandes FTP
    ftp -n $FTP_HOST $FTP_PORT < /tmp/ftp_commands_$$.txt 2>&1 | head -20

    # Nettoyer
    rm -f /tmp/ftp_commands_$$.txt

    echo -e "${GREEN}✓ Test terminé${NC}"
    echo ""
    sleep 1
}

# Test 1: Connexion anonymous
echo -e "${BLUE}=== Test 1: Connexion Anonymous ===${NC}"
test_ftp "anonymous" "" "Tentative de connexion anonymous"

# Test 2: Credentials admin/admin
echo -e "${BLUE}=== Test 2: Admin Credentials ===${NC}"
test_ftp "admin" "admin123" "Tentative avec admin/admin123"

# Test 3: Brute force simulation
echo -e "${BLUE}=== Test 3: Brute Force Simulation ===${NC}"
echo -e "${YELLOW}Tentatives multiples avec différents mots de passe...${NC}"
for pass in "password" "123456" "admin" "root" "test"; do
    echo "  Essai password: $pass"
    echo -e "user admin\n$pass\nquit" | ftp -n $FTP_HOST $FTP_PORT >/dev/null 2>&1
    sleep 0.5
done
echo -e "${GREEN}✓ Brute force terminé${NC}"
echo ""

# Test 4: Tentative d'upload de fichier malveillant
echo -e "${BLUE}=== Test 4: Upload de Fichier Malveillant ===${NC}"
cat > /tmp/ftp_malicious_$$.txt << EOF
user ftpuser
password
stor shell.php
stor backdoor.exe
stor webshell.jsp
stor malware.sh
quit
EOF

ftp -n $FTP_HOST $FTP_PORT < /tmp/ftp_malicious_$$.txt 2>&1 | head -15
rm -f /tmp/ftp_malicious_$$.txt
echo -e "${GREEN}✓ Test d'upload terminé${NC}"
echo ""

# Test 5: Commandes de reconnaissance
echo -e "${BLUE}=== Test 5: Reconnaissance ===${NC}"
cat > /tmp/ftp_recon_$$.txt << EOF
user admin
admin123
syst
pwd
cwd /etc
list
cwd /var
list
cwd /home
list
quit
EOF

ftp -n $FTP_HOST $FTP_PORT < /tmp/ftp_recon_$$.txt 2>&1 | head -20
rm -f /tmp/ftp_recon_$$.txt
echo -e "${GREEN}✓ Reconnaissance terminée${NC}"
echo ""

# Test 6: Tentative de téléchargement
echo -e "${BLUE}=== Test 6: Tentatives de Téléchargement ===${NC}"
cat > /tmp/ftp_download_$$.txt << EOF
user admin
admin123
retr /etc/passwd
retr /etc/shadow
retr config.php
retr .htpasswd
retr database.sql
quit
EOF

ftp -n $FTP_HOST $FTP_PORT < /tmp/ftp_download_$$.txt 2>&1 | head -20
rm -f /tmp/ftp_download_$$.txt
echo -e "${GREEN}✓ Tests de téléchargement terminés${NC}"
echo ""

# Test 7: Manipulation de répertoires
echo -e "${BLUE}=== Test 7: Manipulation de Répertoires ===${NC}"
cat > /tmp/ftp_dirs_$$.txt << EOF
user admin
admin123
mkd backdoor
mkd uploads
mkd tmp
cwd backdoor
pwd
cdup
rmd backdoor
quit
EOF

ftp -n $FTP_HOST $FTP_PORT < /tmp/ftp_dirs_$$.txt 2>&1 | head -20
rm -f /tmp/ftp_dirs_$$.txt
echo -e "${GREEN}✓ Tests de répertoires terminés${NC}"
echo ""

# Test 8: Commandes avancées
echo -e "${BLUE}=== Test 8: Commandes Avancées ===${NC}"
cat > /tmp/ftp_advanced_$$.txt << EOF
user root
password
feat
noop
type I
mode S
pasv
help
quit
EOF

ftp -n $FTP_HOST $FTP_PORT < /tmp/ftp_advanced_$$.txt 2>&1 | head -20
rm -f /tmp/ftp_advanced_$$.txt
echo -e "${GREEN}✓ Tests avancés terminés${NC}"
echo ""

# Résumé
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ Tous les tests FTP terminés !${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}Vérifiez le dashboard web pour voir les résultats :${NC}"
echo -e "  ${BLUE}http://$FTP_HOST:8080${NC}"
echo ""
echo -e "${YELLOW}Vérifiez les logs pour les alertes :${NC}"
echo -e "  ${BLUE}docker-compose logs honeypot | grep FTP${NC}"
echo ""
echo -e "${YELLOW}API FTP Stats:${NC}"
echo -e "  ${BLUE}curl http://$FTP_HOST:8080/api/ftp/statistics | jq${NC}"
echo -e "  ${BLUE}curl http://$FTP_HOST:8080/api/ftp/connections | jq${NC}"
echo -e "  ${BLUE}curl http://$FTP_HOST:8080/api/ftp/commands | jq${NC}"
echo ""
