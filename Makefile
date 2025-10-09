# Makefile pour le projet Honey SSH Honeypot

# Variables
BINARY_NAME=honeypot
CONFIG_FILE=config.yaml
BUILD_DIR=build
LOG_DIR=logs
WEB_DIR=web

# Couleurs pour les messages
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: all build clean test run install deploy help

# Cible par défaut
all: build

# Afficher l'aide
help:
	@echo "$(BLUE)Honey SSH Honeypot - Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)Commandes disponibles:$(NC)"
	@echo "  $(YELLOW)build$(NC)     - Compiler le honeypot"
	@echo "  $(YELLOW)clean$(NC)     - Nettoyer les fichiers de build"
	@echo "  $(YELLOW)test$(NC)      - Exécuter les tests"
	@echo "  $(YELLOW)run$(NC)       - Lancer le honeypot"
	@echo "  $(YELLOW)install$(NC)   - Installer les dépendances"
	@echo "  $(YELLOW)deploy$(NC)    - Déployer le honeypot"
	@echo "  $(YELLOW)test-attacks$(NC) - Tester les attaques"
	@echo "  $(YELLOW)help$(NC)      - Afficher cette aide"
	@echo ""

# Installer les dépendances
install:
	@echo "$(BLUE)Installation des dépendances...$(NC)"
	go mod tidy
	go mod download
	@echo "$(GREEN)Dépendances installées$(NC)"

# Compiler le projet
build: install
	@echo "$(BLUE)Compilation du honeypot...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@mkdir -p $(LOG_DIR)
	@mkdir -p $(WEB_DIR)/static
	@mkdir -p $(WEB_DIR)/templates
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)Compilation réussie: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Nettoyer les fichiers de build
clean:
	@echo "$(BLUE)Nettoyage...$(NC)"
	rm -rf $(BUILD_DIR)
	go clean
	@echo "$(GREEN)Nettoyage terminé$(NC)"

# Exécuter les tests
test:
	@echo "$(BLUE)Exécution des tests...$(NC)"
	go test -v ./...
	@echo "$(GREEN)Tests terminés$(NC)"

# Lancer le honeypot
run: build
	@echo "$(BLUE)Démarrage du honeypot...$(NC)"
	@echo "$(YELLOW)Interface web: http://localhost:8080$(NC)"
	@echo "$(YELLOW)SSH honeypot: localhost:2222$(NC)"
	@echo "$(YELLOW)Utilisateurs de test: admin/admin123, root/password$(NC)"
	@echo "$(YELLOW)Appuyez sur Ctrl+C pour arrêter$(NC)"
	@echo ""
	./$(BUILD_DIR)/$(BINARY_NAME) -config $(CONFIG_FILE)

# Déployer le honeypot
deploy:
	@echo "$(BLUE)Déploiement du honeypot...$(NC)"
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "$(RED)Erreur: Le déploiement nécessite les droits root$(NC)"; \
		echo "$(YELLOW)Utilisez: sudo make deploy$(NC)"; \
		exit 1; \
	fi
	./deploy.sh install
	@echo "$(GREEN)Déploiement terminé$(NC)"

# Démarrer le service
start:
	@echo "$(BLUE)Démarrage du service honeypot...$(NC)"
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "$(RED)Erreur: Le démarrage nécessite les droits root$(NC)"; \
		echo "$(YELLOW)Utilisez: sudo make start$(NC)"; \
		exit 1; \
	fi
	./deploy.sh start
	@echo "$(GREEN)Service démarré$(NC)"

# Arrêter le service
stop:
	@echo "$(BLUE)Arrêt du service honeypot...$(NC)"
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "$(RED)Erreur: L'arrêt nécessite les droits root$(NC)"; \
		echo "$(YELLOW)Utilisez: sudo make stop$(NC)"; \
		exit 1; \
	fi
	./deploy.sh stop
	@echo "$(GREEN)Service arrêté$(NC)"

# Vérifier le statut
status:
	@echo "$(BLUE)Statut du service honeypot...$(NC)"
	./deploy.sh status

# Tester les attaques
test-attacks:
	@echo "$(BLUE)Test des attaques sur le honeypot...$(NC)"
	@chmod +x test_attacks.sh
	./test_attacks.sh complete

# Test de connexion basique
test-basic:
	@echo "$(BLUE)Test de connexion basique...$(NC)"
	@chmod +x test_attacks.sh
	./test_attacks.sh basic

# Test de force brute
test-brute:
	@echo "$(BLUE)Test de force brute...$(NC)"
	@chmod +x test_attacks.sh
	./test_attacks.sh brute_force

# Test des commandes
test-commands:
	@echo "$(BLUE)Test des commandes...$(NC)"
	@chmod +x test_attacks.sh
	./test_attacks.sh commands

# Test de l'interface web
test-web:
	@echo "$(BLUE)Test de l'interface web...$(NC)"
	@chmod +x test_attacks.sh
	./test_attacks.sh web

# Vérifier la configuration
check-config:
	@echo "$(BLUE)Vérification de la configuration...$(NC)"
	@if [ ! -f $(CONFIG_FILE) ]; then \
		echo "$(RED)Erreur: Fichier de configuration $(CONFIG_FILE) non trouvé$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Configuration valide$(NC)"

# Vérifier les dépendances système
check-deps:
	@echo "$(BLUE)Vérification des dépendances système...$(NC)"
	@command -v go >/dev/null 2>&1 || { echo "$(RED)Go n'est pas installé$(NC)"; exit 1; }
	@command -v sqlite3 >/dev/null 2>&1 || { echo "$(RED)SQLite3 n'est pas installé$(NC)"; exit 1; }
	@echo "$(GREEN)Toutes les dépendances sont installées$(NC)"

# Formatage du code
fmt:
	@echo "$(BLUE)Formatage du code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)Code formaté$(NC)"

# Vérification du code
lint:
	@echo "$(BLUE)Vérification du code...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "$(YELLOW)golangci-lint non installé, installation...$(NC)"; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run
	@echo "$(GREEN)Vérification terminée$(NC)"

# Générer la documentation
docs:
	@echo "$(BLUE)Génération de la documentation...$(NC)"
	@command -v godoc >/dev/null 2>&1 || { echo "$(YELLOW)godoc non installé, installation...$(NC)"; go install golang.org/x/tools/cmd/godoc@latest; }
	@echo "$(GREEN)Documentation générée$(NC)"
	@echo "$(YELLOW)Accédez à http://localhost:6060/pkg/honey/$(NC)"
	godoc -http=:6060

# Sauvegarde de la base de données
backup:
	@echo "$(BLUE)Sauvegarde de la base de données...$(NC)"
	@if [ -f honeypot.db ]; then \
		cp honeypot.db backup-$$(date +%Y%m%d_%H%M%S).db; \
		echo "$(GREEN)Sauvegarde créée: backup-$$(date +%Y%m%d_%H%M%S).db$(NC)"; \
	else \
		echo "$(YELLOW)Aucune base de données à sauvegarder$(NC)"; \
	fi

# Restauration de la base de données
restore:
	@echo "$(BLUE)Restauration de la base de données...$(NC)"
	@ls -la backup-*.db 2>/dev/null || { echo "$(RED)Aucune sauvegarde trouvée$(NC)"; exit 1; }
	@echo "$(YELLOW)Sauvegardes disponibles:$(NC)"
	@ls -la backup-*.db
	@echo "$(YELLOW)Spécifiez le fichier de sauvegarde: make restore BACKUP=backup-YYYYMMDD_HHMMSS.db$(NC)"

# Cible pour la restauration avec paramètre
restore-file:
	@if [ -z "$(BACKUP)" ]; then \
		echo "$(RED)Erreur: Spécifiez le fichier de sauvegarde$(NC)"; \
		echo "$(YELLOW)Usage: make restore-file BACKUP=backup-YYYYMMDD_HHMMSS.db$(NC)"; \
		exit 1; \
	fi
	@if [ ! -f "$(BACKUP)" ]; then \
		echo "$(RED)Erreur: Fichier $(BACKUP) non trouvé$(NC)"; \
		exit 1; \
	fi
	@cp $(BACKUP) honeypot.db
	@echo "$(GREEN)Base de données restaurée depuis $(BACKUP)$(NC)"

# Affichage des informations du projet
info:
	@echo "$(BLUE)Informations du projet Honey SSH Honeypot$(NC)"
	@echo ""
	@echo "$(GREEN)Version Go:$(NC) $$(go version)"
	@echo "$(GREEN)Architecture:$(NC) $$(go env GOOS)/$$(go env GOARCH)"
	@echo "$(GREEN)Répertoire de travail:$(NC) $$(pwd)"
	@echo "$(GREEN)Modules Go:$(NC) $$(go list -m all | wc -l) modules"
	@echo ""
	@if [ -f $(CONFIG_FILE) ]; then \
		echo "$(GREEN)Configuration:$(NC) $(CONFIG_FILE) ✓"; \
	else \
		echo "$(RED)Configuration:$(NC) $(CONFIG_FILE) ✗"; \
	fi
	@if [ -f honeypot.db ]; then \
		echo "$(GREEN)Base de données:$(NC) honeypot.db ✓"; \
		echo "$(GREEN)Taille:$(NC) $$(du -h honeypot.db | cut -f1)"; \
	else \
		echo "$(YELLOW)Base de données:$(NC) honeypot.db (non créée)"; \
	fi
