#!/bin/bash

# Script de test pour simuler des attaques sur le honeypot
# Usage: ./test_attacks.sh [basic|brute_force|commands]

set -e

# Configuration
HONEYPOT_HOST="192.168.1.144"
HONEYPOT_PORT="2222"
SSH_OPTS="-o ConnectTimeout=5 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test de connexion basique
test_basic_connection() {
    log_info "Test de connexion basique..."
    
    # Test avec des identifiants valides
    log_info "Test avec identifiants valides (admin/admin123)..."
    if ssh $SSH_OPTS -p $HONEYPOT_PORT admin@$HONEYPOT_HOST "whoami" 2>/dev/null; then
        log_success "Connexion réussie avec admin/admin123"
    else
        log_warning "Connexion échouée avec admin/admin123"
    fi
    
    # Test avec des identifiants invalides
    log_info "Test avec identifiants invalides (test/wrong)..."
    if ssh $SSH_OPTS -p $HONEYPOT_PORT test@$HONEYPOT_HOST "whoami" 2>/dev/null; then
        log_warning "Connexion réussie avec test/wrong (inattendu)"
    else
        log_success "Connexion échouée avec test/wrong (attendu)"
    fi
}

# Test d'attaque par force brute
test_brute_force() {
    log_info "Test d'attaque par force brute..."
    
    # Liste de mots de passe communs
    passwords=("password" "123456" "admin" "root" "test" "user" "guest" "qwerty" "letmein" "welcome")
    
    for password in "${passwords[@]}"; do
        log_info "Tentative avec admin/$password..."
        ssh $SSH_OPTS -p $HONEYPOT_PORT admin@$HONEYPOT_HOST "whoami" 2>/dev/null || true
        sleep 1
    done
    
    log_success "Test de force brute terminé"
}

# Test d'exécution de commandes
test_commands() {
    log_info "Test d'exécution de commandes..."
    
    # Se connecter et exécuter des commandes
    log_info "Connexion et exécution de commandes..."
    
    ssh $SSH_OPTS -p $HONEYPOT_PORT admin@$HONEYPOT_HOST << 'EOF'
echo "=== Test des commandes de base ==="
whoami
pwd
ls -la
ps aux
netstat -tuln
ifconfig
uname -a
id
groups
uptime
df -h
free -m
history
echo "=== Test des commandes de recherche ==="
find /home -name "*.txt" 2>/dev/null
grep -r "password" /etc 2>/dev/null | head -5
cat /etc/passwd
cat /etc/hosts
echo "=== Test des commandes réseau ==="
wget --help 2>/dev/null || echo "wget non disponible"
curl --help 2>/dev/null || echo "curl non disponible"
echo "=== Fin des tests ==="
exit
EOF

    log_success "Test de commandes terminé"
}

# Test de géolocalisation (simulation)
test_geolocation() {
    log_info "Test de géolocalisation..."
    
    # Simuler des connexions depuis différentes IPs
    log_info "Simulation de connexions depuis différentes IPs..."
    
    # Note: Dans un vrai test, on utiliserait des IPs différentes
    # Ici on simule juste avec des connexions multiples
    for i in {1..5}; do
        log_info "Connexion simulée $i..."
        ssh $SSH_OPTS -p $HONEYPOT_PORT admin@$HONEYPOT_HOST "whoami" 2>/dev/null || true
        sleep 2
    done
    
    log_success "Test de géolocalisation terminé"
}

# Test de l'interface web
test_web_interface() {
    log_info "Test de l'interface web..."
    
    # Tester l'accès à l'interface web
    if curl -s http://localhost:8080 > /dev/null; then
        log_success "Interface web accessible"
        
        # Tester les endpoints API
        log_info "Test des endpoints API..."
        
        if curl -s http://localhost:8080/api/statistics | grep -q "total_connections"; then
            log_success "API statistics fonctionnelle"
        else
            log_warning "API statistics non fonctionnelle"
        fi
        
        if curl -s http://localhost:8080/api/connections | grep -q "remote_addr"; then
            log_success "API connections fonctionnelle"
        else
            log_warning "API connections non fonctionnelle"
        fi
        
    else
        log_error "Interface web non accessible"
    fi
}

# Test complet
test_complete() {
    log_info "Test complet du honeypot..."
    
    test_basic_connection
    echo ""
    test_brute_force
    echo ""
    test_commands
    echo ""
    test_geolocation
    echo ""
    test_web_interface
    
    log_success "Tests complets terminés"
}

# Afficher l'aide
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  basic        - Test de connexion basique"
    echo "  brute_force  - Test d'attaque par force brute"
    echo "  commands     - Test d'exécution de commandes"
    echo "  geolocation  - Test de géolocalisation"
    echo "  web          - Test de l'interface web"
    echo "  complete     - Test complet (défaut)"
    echo "  help         - Afficher cette aide"
    echo ""
    echo "Exemples:"
    echo "  $0 basic      # Test de connexion basique"
    echo "  $0 complete   # Test complet"
}

# Fonction principale
main() {
    case "${1:-complete}" in
        basic)
            test_basic_connection
            ;;
        brute_force)
            test_brute_force
            ;;
        commands)
            test_commands
            ;;
        geolocation)
            test_geolocation
            ;;
        web)
            test_web_interface
            ;;
        complete)
            test_complete
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Commande inconnue: $1"
            show_help
            exit 1
            ;;
    esac
}

# Vérifier que le honeypot est en cours d'exécution
check_honeypot() {
    if ! nc -z $HONEYPOT_HOST $HONEYPOT_PORT 2>/dev/null; then
        log_error "Le honeypot n'est pas accessible sur $HONEYPOT_HOST:$HONEYPOT_PORT"
        log_info "Assurez-vous que le honeypot est démarré avec: ./deploy.sh start"
        exit 1
    fi
}

# Vérifier les dépendances
check_dependencies() {
    local missing_deps=()
    
    if ! command -v ssh &> /dev/null; then
        missing_deps+=("openssh-client")
    fi
    
    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    if ! command -v nc &> /dev/null; then
        missing_deps+=("netcat")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Dépendances manquantes: ${missing_deps[*]}"
        log_info "Installez-les avec: sudo apt install ${missing_deps[*]}"
        exit 1
    fi
}

# Exécuter les vérifications et le test
check_dependencies
check_honeypot
main "$@"
