# 🍯 Honey SSH Honeypot

Un honeypot SSH complet développé en Go pour détecter et analyser les tentatives d'intrusion. Ce projet simule un serveur SSH factice pour capturer les activités des attaquants et fournir des insights sur leurs comportements.

## 🎯 Objectifs

- **Détection d'intrusion** : Capturer toutes les tentatives de connexion SSH
- **Analyse comportementale** : Enregistrer et analyser les commandes exécutées
- **Surveillance en temps réel** : Interface web pour visualiser les attaques
- **Alertes automatiques** : Notifications par email en cas d'activité suspecte
- **Géolocalisation** : Identifier l'origine géographique des attaquants

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Attaquant     │───▶│   Honeypot SSH  │───▶│   Base de       │
│   (Client SSH)  │    │   (Port 2222)   │    │   Données       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  Interface Web  │
                       │  (Port 8080)    │
                       └─────────────────┘
```

## 🚀 Installation et Déploiement

### Prérequis

- Go 1.21 ou supérieur
- SQLite3
- Linux/Unix (testé sur Ubuntu)

### Installation Automatique

```bash
# Cloner le projet
git clone <repository-url>
cd honey

# Installation complète (nécessite les droits root)
sudo ./deploy.sh install

# Démarrer le honeypot
sudo ./deploy.sh start
```

### Installation Manuelle

```bash
# Compiler le projet
go mod tidy
go build -o honeypot .

# Créer les dossiers nécessaires
mkdir -p logs web/static web/templates

# Lancer le honeypot
./honeypot -config config.yaml
```

## ⚙️ Configuration

Le fichier `config.yaml` permet de personnaliser le comportement du honeypot :

```yaml
server:
  host: "0.0.0.0"
  port: 2222
  max_connections: 10
  connection_timeout: 300

auth:
  fake_users:
    - username: "admin"
      password: "admin123"
    - username: "root"
      password: "password"

shell:
  prompt: "user@honeypot:~$ "
  welcome_message: "Welcome to Ubuntu 20.04.3 LTS"

web:
  enabled: true
  host: "127.0.0.1"
  port: 8080

alerts:
  enabled: true
  email:
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "your-email@gmail.com"
    password: "your-app-password"
    to: ["admin@company.com"]
```

## 🧪 Tests et Simulation

### Script de Test Automatique

```bash
# Test complet
./test_attacks.sh complete

# Tests spécifiques
./test_attacks.sh basic        # Connexions basiques
./test_attacks.sh brute_force  # Attaques par force brute
./test_attacks.sh commands     # Exécution de commandes
./test_attacks.sh web          # Interface web
```

### Test Manuel

```bash
# Connexion SSH au honeypot
ssh -p 2222 admin@localhost
# Mot de passe: admin123

# Commandes de test
whoami
ls -la
ps aux
cat /etc/passwd
```

## 📊 Interface Web

L'interface web est accessible sur `http://localhost:8080` et fournit :

- **Tableau de bord** : Statistiques en temps réel
- **Connexions** : Liste des tentatives de connexion
- **Commandes** : Historique des commandes exécutées
- **Alertes** : Notifications de sécurité
- **Géolocalisation** : Origine des attaquants

### Endpoints API

- `GET /api/statistics` - Statistiques générales
- `GET /api/connections` - Liste des connexions
- `GET /api/commands` - Historique des commandes
- `GET /api/alerts` - Alertes de sécurité

## 🔍 Fonctionnalités

### Serveur SSH Factice

- **Authentification** : Accepte tous les mots de passe configurés
- **Shell interactif** : Simule un environnement Linux complet
- **Commandes factices** : Réponses réalistes pour les commandes courantes
- **Timeout** : Limite la durée des sessions

### Détection d'Intrusion

- **Force brute** : Détection des attaques par dictionnaire
- **Patterns suspects** : Analyse des comportements anormaux
- **Géolocalisation** : Identification de l'origine des attaques
- **Alertes** : Notifications automatiques par email

### Base de Données

- **SQLite** : Stockage local des données
- **Tables** : connexions, commandes, alertes, patterns d'attaque
- **Requêtes** : API pour l'analyse des données

## 🛡️ Sécurité

### Bonnes Pratiques

- **Isolation** : Exécuter dans une VM dédiée
- **Monitoring** : Surveiller les logs et alertes
- **Mise à jour** : Maintenir le système à jour
- **Sauvegarde** : Sauvegarder régulièrement la base de données

### Limitations

- **Port SSH** : Utilise le port 2222 par défaut (non standard)
- **Authentification** : Accepte tous les mots de passe configurés
- **Commandes** : Réponses factices, pas d'exécution réelle

## 📈 Monitoring et Alertes

### Types d'Alertes

- **Brute Force** : Tentatives multiples de connexion
- **Activité Suspecte** : Commandes dangereuses
- **Géolocalisation** : Connexions depuis des pays suspects

### Configuration Email

```yaml
alerts:
  enabled: true
  email:
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "your-email@gmail.com"
    password: "your-app-password"
    to: ["admin@company.com"]
  thresholds:
    failed_attempts: 5
    time_window: 300
```

## 🔧 Maintenance

### Commandes de Gestion

```bash
# Statut du service
sudo ./deploy.sh status

# Arrêter le honeypot
sudo ./deploy.sh stop

# Redémarrer le honeypot
sudo ./deploy.sh stop && sudo ./deploy.sh start

# Voir les logs
journalctl -u honeypot -f
```

### Sauvegarde

```bash
# Sauvegarder la base de données
cp /opt/honeypot/honeypot.db backup-$(date +%Y%m%d).db

# Sauvegarder les logs
cp -r /opt/honeypot/logs backup-logs-$(date +%Y%m%d)
```

## 🐛 Dépannage

### Problèmes Courants

1. **Port déjà utilisé**
   ```bash
   # Vérifier les ports utilisés
   netstat -tulpn | grep :2222
   ```

2. **Permissions insuffisantes**
   ```bash
   # Vérifier les permissions
   ls -la /opt/honeypot/
   ```

3. **Base de données corrompue**
   ```bash
   # Vérifier la base de données
   sqlite3 /opt/honeypot/honeypot.db ".tables"
   ```

### Logs

```bash
# Logs du service
journalctl -u honeypot -f

# Logs de l'application
tail -f /opt/honeypot/logs/honeypot.log
```

## 📚 Structure du Projet

```
honey/
├── main.go                    # Point d'entrée principal
├── config.yaml               # Configuration
├── go.mod                    # Dépendances Go
├── deploy.sh                 # Script de déploiement
├── test_attacks.sh           # Script de test
├── README.md                 # Documentation
├── internal/
│   ├── config/              # Gestion de la configuration
│   ├── database/            # Gestion de la base de données
│   ├── honeypot/            # Serveur SSH et session factice
│   ├── logger/              # Système de logging
│   ├── models/              # Modèles de données
│   ├── utils/               # Utilitaires (email, géolocalisation)
│   └── web/                 # Interface web
├── web/
│   ├── static/              # Fichiers statiques
│   └── templates/           # Templates HTML
└── logs/                    # Fichiers de logs
```

## 🤝 Contribution

1. Fork le projet
2. Créer une branche feature (`git checkout -b feature/AmazingFeature`)
3. Commit les changements (`git commit -m 'Add some AmazingFeature'`)
4. Push vers la branche (`git push origin feature/AmazingFeature`)
5. Ouvrir une Pull Request

## 📄 Licence

Ce projet est sous licence MIT. Voir le fichier `LICENSE` pour plus de détails.

## ⚠️ Avertissement

Ce honeypot est destiné à des fins éducatives et de recherche en sécurité. Utilisez-le uniquement dans des environnements contrôlés et avec les autorisations appropriées. Les auteurs ne sont pas responsables de l'utilisation abusive de ce logiciel.

## 📞 Support

Pour toute question ou problème :

1. Vérifiez la documentation
2. Consultez les issues existantes
3. Créez une nouvelle issue avec les détails du problème

---

**Développé avec ❤️ en Go pour la sécurité informatique**
