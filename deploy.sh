#!/bin/bash

# Script de déploiement du Honeypot SSH
# Usage: ./deploy.sh [install|start|stop|status|test]

set -e

HONEYPOT_DIR="/opt/honeypot"
SERVICE_NAME="honeypot"
CONFIG_FILE="config.yaml"
BINARY_NAME="honeypot"

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Fonction pour afficher les messages
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

# Vérifier si le script est exécuté en tant que root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "Ce script doit être exécuté en tant que root"
        exit 1
    fi
}

# Installer les dépendances
install_dependencies() {
    log_info "Installation des dépendances..."
    
    # Mettre à jour les paquets
    apt update
    
    # Installer Go si pas déjà installé
    if ! command -v go &> /dev/null; then
        log_info "Installation de Go..."
        apt install -y golang-go
    fi
    
    # Installer SQLite
    apt install -y sqlite3
    
    # Installer les outils de développement
    apt install -y build-essential git
    
    log_success "Dépendances installées"
}

# Compiler le honeypot
build_honeypot() {
    log_info "Compilation du honeypot..."
    
    # Aller dans le répertoire du projet
    cd "$(dirname "$0")"
    
    # Nettoyer les builds précédents
    go clean
    
    # Télécharger les dépendances
    go mod tidy
    
    # Compiler
    go build -o $BINARY_NAME .
    
    if [ $? -eq 0 ]; then
        log_success "Compilation réussie"
    else
        log_error "Échec de la compilation"
        exit 1
    fi
}

# Installer le honeypot
install_honeypot() {
    log_info "Installation du honeypot..."
    
    # Créer le répertoire d'installation
    mkdir -p $HONEYPOT_DIR
    mkdir -p $HONEYPOT_DIR/logs
    mkdir -p $HONEYPOT_DIR/web/static
    mkdir -p $HONEYPOT_DIR/web/templates
    
    # Copier les fichiers
    cp $BINARY_NAME $HONEYPOT_DIR/
    cp $CONFIG_FILE $HONEYPOT_DIR/
    
    # Créer le fichier de service systemd
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Honey SSH Honeypot
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$HONEYPOT_DIR
ExecStart=$HONEYPOT_DIR/$BINARY_NAME -config=$HONEYPOT_DIR/$CONFIG_FILE
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    # Recharger systemd
    systemctl daemon-reload
    
    # Activer le service
    systemctl enable $SERVICE_NAME
    
    log_success "Honeypot installé dans $HONEYPOT_DIR"
}

# Démarrer le honeypot
start_honeypot() {
    log_info "Démarrage du honeypot..."
    
    systemctl start $SERVICE_NAME
    
    if [ $? -eq 0 ]; then
        log_success "Honeypot démarré"
        log_info "Interface web disponible sur http://localhost:8080"
    else
        log_error "Échec du démarrage"
        exit 1
    fi
}

# Arrêter le honeypot
stop_honeypot() {
    log_info "Arrêt du honeypot..."
    
    systemctl stop $SERVICE_NAME
    
    if [ $? -eq 0 ]; then
        log_success "Honeypot arrêté"
    else
        log_error "Échec de l'arrêt"
        exit 1
    fi
}

# Vérifier le statut
status_honeypot() {
    log_info "Statut du honeypot:"
    
    systemctl status $SERVICE_NAME --no-pager
    
    # Afficher les logs récents
    log_info "Logs récents:"
    journalctl -u $SERVICE_NAME -n 20 --no-pager
}

# Tester le honeypot
test_honeypot() {
    log_info "Test du honeypot..."
    
    # Vérifier si le service est actif
    if systemctl is-active --quiet $SERVICE_NAME; then
        log_success "Service actif"
    else
        log_error "Service inactif"
        return 1
    fi
    
    # Tester la connexion SSH
    log_info "Test de connexion SSH..."
    timeout 5 ssh -o ConnectTimeout=3 -o StrictHostKeyChecking=no -p 2222 test@localhost || true
    
    # Tester l'interface web
    log_info "Test de l'interface web..."
    if curl -s http://localhost:8080 > /dev/null; then
        log_success "Interface web accessible"
    else
        log_warning "Interface web non accessible"
    fi
    
    # Afficher les statistiques
    log_info "Statistiques de la base de données:"
    if [ -f "$HONEYPOT_DIR/honeypot.db" ]; then
        sqlite3 "$HONEYPOT_DIR/honeypot.db" "SELECT COUNT(*) as total_connections FROM connections;"
        sqlite3 "$HONEYPOT_DIR/honeypot.db" "SELECT COUNT(*) as total_commands FROM commands;"
    else
        log_warning "Base de données non trouvée"
    fi
}

# Afficher l'aide
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  install  - Installer le honeypot et ses dépendances"
    echo "  start    - Démarrer le honeypot"
    echo "  stop     - Arrêter le honeypot"
    echo "  status   - Afficher le statut du honeypot"
    echo "  test     - Tester le honeypot"
    echo "  help     - Afficher cette aide"
    echo ""
    echo "Exemples:"
    echo "  $0 install    # Installation complète"
    echo "  $0 start      # Démarrer le service"
    echo "  $0 test       # Tester le fonctionnement"
}

# Fonction principale
main() {
    case "${1:-help}" in
        install)
            check_root
            install_dependencies
            build_honeypot
            install_honeypot
            log_success "Installation terminée!"
            log_info "Utilisez '$0 start' pour démarrer le honeypot"
            ;;
        start)
            check_root
            start_honeypot
            ;;
        stop)
            check_root
            stop_honeypot
            ;;
        status)
            status_honeypot
            ;;
        test)
            test_honeypot
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

# Exécuter la fonction principale
main "$@"
