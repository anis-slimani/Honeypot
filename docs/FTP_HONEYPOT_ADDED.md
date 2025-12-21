# FTP Honeypot - Nouvelle Fonctionnalité Ajoutée

**Date:** 2025-12-20
**Version:** 4.1

---

## 🎯 Vue d'ensemble

Un **serveur FTP honeypot complet** a été ajouté au projet, portant le nombre de protocoles supportés à **4** :
- SSH (port 2222)
- HTTP (port 80)
- HTTPS (port 443)
- **FTP (port 2121)** ✨ NOUVEAU

---

## ✨ Fonctionnalités FTP

### 1. Serveur FTP Complet

Le honeypot FTP simule un serveur FTP réaliste avec support de :

#### Commandes Standards
- **Authentification** : `USER`, `PASS`
- **Navigation** : `PWD`, `CWD`, `CDUP`, `LIST`, `NLST`
- **Transferts** : `RETR` (download), `STOR`/`STOU` (upload)
- **Gestion fichiers** : `DELE` (delete), `MKD`/`XMKD` (create dir), `RMD`/`XRMD` (remove dir)
- **Configuration** : `TYPE`, `MODE`, `PASV`, `PORT`
- **Utilitaires** : `NOOP`, `FEAT`, `OPTS`, `HELP`, `SYST`, `QUIT`

**Total : 20+ commandes FTP simulées**

### 2. Détection d'Attaques

#### Types d'attaques détectés :

| Type d'attaque | Description | Niveau |
|----------------|-------------|--------|
| **Anonymous Login** | Tentative de connexion anonymous/ftp | WARNING |
| **Brute Force** | Multiples tentatives de connexion | WARNING |
| **Malicious Upload** | Upload de fichiers suspects (.php, .exe, .sh, etc.) | CRITICAL |
| **Command Injection** | Commandes dangereuses (SITE EXEC, SITE CHMOD) | CRITICAL |
| **File Download** | Tentatives de téléchargement de fichiers sensibles | INFO |
| **Directory Manipulation** | Création/suppression de répertoires | WARNING |

#### Fichiers Suspects Détectés

Le honeypot identifie les extensions malveillantes :
- `.php`, `.jsp`, `.asp` - Web shells
- `.exe`, `.bat`, `.cmd` - Exécutables Windows
- `.sh`, `.py`, `.pl` - Scripts shell/interpréteurs

### 3. Journalisation Complète

#### Base de données - 2 nouvelles tables

**`ftp_connections`**
```sql
CREATE TABLE ftp_connections (
    id INTEGER PRIMARY KEY,
    remote_addr TEXT NOT NULL,
    username TEXT,
    password TEXT,
    authenticated BOOLEAN,
    connected_at DATETIME,
    disconnected_at DATETIME,
    duration INTEGER,
    current_dir TEXT DEFAULT '/',
    login_attempts INTEGER DEFAULT 0,
    country TEXT,
    city TEXT
)
```

**`ftp_commands`**
```sql
CREATE TABLE ftp_commands (
    id INTEGER PRIMARY KEY,
    connection_id INTEGER NOT NULL,
    command TEXT NOT NULL,
    executed_at DATETIME,
    response TEXT,
    FOREIGN KEY (connection_id) REFERENCES ftp_connections (id)
)
```

---

## 🔌 API REST - Nouveaux Endpoints

### `/api/ftp/connections`
Retourne les connexions FTP récentes

**Paramètres :**
- `limit` (optionnel) - Nombre de connexions (défaut: 100)

**Exemple :**
```bash
curl http://localhost:8080/api/ftp/connections?limit=50 | jq
```

**Réponse :**
```json
[
  {
    "id": 1,
    "remote_addr": "192.168.1.100:54321",
    "username": "admin",
    "password": "admin123",
    "authenticated": true,
    "connected_at": "2025-12-20T15:30:00Z",
    "disconnected_at": "2025-12-20T15:35:00Z",
    "duration": 300,
    "current_dir": "/",
    "login_attempts": 1,
    "country": "France",
    "city": "Paris"
  }
]
```

### `/api/ftp/commands`
Retourne les commandes FTP exécutées

**Paramètres :**
- `limit` (optionnel) - Nombre de commandes (défaut: 100)

**Exemple :**
```bash
curl http://localhost:8080/api/ftp/commands?limit=100 | jq
```

**Réponse :**
```json
[
  {
    "id": 1,
    "connection_id": 1,
    "command": "STOR malware.php",
    "executed_at": "2025-12-20T15:31:00Z",
    "remote_addr": "192.168.1.100:54321",
    "username": "admin",
    "response": "226 Transfer complete"
  }
]
```

### `/api/ftp/statistics`
Retourne les statistiques FTP

**Exemple :**
```bash
curl http://localhost:8080/api/ftp/statistics | jq
```

**Réponse :**
```json
{
  "total_connections": 152,
  "authenticated_logins": 134,
  "failed_logins": 18,
  "total_commands": 1847,
  "unique_attackers": 42,
  "connections_last_24h": 28,
  "top_usernames": [
    {"username": "admin", "count": 45},
    {"username": "root", "count": 32},
    {"username": "anonymous", "count": 28}
  ],
  "top_passwords": [
    {"password": "admin123", "count": 35},
    {"password": "password", "count": 25}
  ],
  "top_commands": [
    {"command": "USER admin", "count": 45},
    {"command": "LIST", "count": 38},
    {"command": "PWD", "count": 35}
  ],
  "top_countries": [
    {"country": "China", "count": 45},
    {"country": "Russia", "count": 32}
  ]
}
```

---

## 🔧 Configuration

### Fichier `config.yaml`

Nouvelle section FTP ajoutée :

```yaml
# Configuration du honeypot FTP
ftp:
  enabled: true
  host: "0.0.0.0"
  port: 2121  # Port FTP (21 nécessite root, on utilise 2121)
  passive_port_min: 50000
  passive_port_max: 50100
  allow_anonymous: true
  fake_users:
    - username: "admin"
      password: "admin123"
    - username: "ftpuser"
      password: "password"
    - username: "anonymous"
      password: ""
```

### Docker

**Port FTP exposé** : `2121:2121`

```yaml
ports:
  - "2222:2222"  # SSH
  - "2121:2121"  # FTP ✨ NOUVEAU
  - "80:80"      # HTTP
  - "8080:8080"  # Dashboard
  - "443:443"    # HTTPS
```

---

## 🧪 Script de Test

Un nouveau script de test complet a été créé : **`scripts/test_ftp.sh`**

### Utilisation

```bash
# Test sur localhost
./scripts/test_ftp.sh

# Test sur hôte distant
./scripts/test_ftp.sh 192.168.1.100 2121
```

### Tests Effectués (8 scénarios)

1. **Connexion Anonymous** - Test de connexion sans credentials
2. **Admin Credentials** - Test avec admin/admin123
3. **Brute Force Simulation** - 5 tentatives avec passwords différents
4. **Upload Malveillants** - Upload de shell.php, backdoor.exe, webshell.jsp
5. **Reconnaissance** - Navigation dans /etc, /var, /home
6. **Téléchargements** - Tentatives RETR /etc/passwd, config.php
7. **Manipulation Répertoires** - MKD, RMD, CWD
8. **Commandes Avancées** - FEAT, TYPE, MODE, PASV

### Exemple de sortie

```
========================================
  Test du Honeypot FTP
  Host: localhost:2121
========================================

Test: Tentative de connexion anonymous
  Username: anonymous
  Password:
220 FTP Server ready
331 Password required for anonymous
230 User anonymous logged in
✓ Test terminé

...
```

---

## 📊 Dashboard Web

### Modifications

Les 3 nouveaux endpoints FTP sont disponibles via le dashboard :

**Accès aux données FTP :**
- Connexions FTP : http://localhost:8080/api/ftp/connections
- Commandes FTP : http://localhost:8080/api/ftp/commands
- Statistiques FTP : http://localhost:8080/api/ftp/statistics

**Exemple d'intégration dashboard (à venir) :**
- Section "FTP Activity" avec graphiques
- Top attaquants FTP
- Commandes FTP en temps réel
- Statistiques d'upload/download

---

## 🔐 Sécurité

### Garanties de sécurité

✅ **Aucune exécution réelle** - Toutes les commandes sont simulées
✅ **Pas de transferts réels** - Pas de vraie data connection
✅ **Isolation complète** - Conteneur Docker isolé
✅ **Logs complets** - Toutes les activités enregistrées
✅ **Détection malware** - Identification des fichiers suspects
✅ **Timeouts** - Limite de 5 minutes par session
✅ **Alertes automatiques** - Notifications email pour activités critiques

### Ce que le honeypot NE fait PAS

❌ Ne stocke PAS réellement les fichiers uploadés
❌ N'exécute PAS les commandes système
❌ Ne donne PAS accès au vrai système de fichiers
❌ Ne crée PAS de vraies connexions data (PASV/PORT)

---

## 📈 Statistiques Collectées

### Métriques FTP

| Métrique | Description |
|----------|-------------|
| **Total connexions** | Nombre de connexions FTP reçues |
| **Connexions authentifiées** | Logins réussis |
| **Connexions échouées** | Tentatives de login ratées |
| **Total commandes** | Nombre de commandes FTP exécutées |
| **Attaquants uniques** | IPs uniques ayant tenté de se connecter |
| **Connexions 24h** | Connexions des dernières 24 heures |
| **Top usernames** | Noms d'utilisateur les plus essayés |
| **Top passwords** | Mots de passe les plus utilisés |
| **Top commandes** | Commandes FTP les plus exécutées |
| **Top pays** | Pays d'origine des attaquants |

---

## 🚀 Démarrage Rapide

### 1. Build et démarrage

```bash
# Build l'image Docker
docker-compose build --no-cache

# Démarrer tous les services (SSH + HTTP + FTP + Dashboard)
docker-compose up -d
```

### 2. Vérification

```bash
# Vérifier que les 4 services sont actifs
docker-compose logs | grep "démarré"

# Sortie attendue:
# Serveur honeypot SSH démarré sur 0.0.0.0:2222
# Honeypot HTTP démarré sur 0.0.0.0:80
# Honeypot FTP démarré sur 0.0.0.0:2121
# Interface web démarrée sur http://0.0.0.0:8080
```

### 3. Test FTP

```bash
# Lancer le script de test
./scripts/test_ftp.sh localhost 2121
```

### 4. Voir les résultats

```bash
# Dashboard web
http://localhost:8080

# API FTP stats
curl http://localhost:8080/api/ftp/statistics | jq

# Logs en temps réel
docker-compose logs -f honeypot | grep FTP
```

---

## 📝 Fichiers Modifiés/Ajoutés

### Nouveaux fichiers

| Fichier | Description | Lignes |
|---------|-------------|--------|
| `internal/services/ftp/ftp_server.go` | Serveur FTP honeypot complet | ~650 |
| `scripts/test_ftp.sh` | Script de test automatisé | ~200 |
| `docs/FTP_HONEYPOT_ADDED.md` | Cette documentation | ~500 |

### Fichiers modifiés

| Fichier | Modifications |
|---------|---------------|
| `internal/config/config.go` | Ajout de FTPConfig + validation |
| `internal/database/database.go` | 2 nouvelles tables FTP |
| `internal/web/web_server.go` | 3 nouveaux handlers API FTP (+260 lignes) |
| `main.go` | Démarrage serveur FTP + compteur services |
| `config.yaml` | Section FTP complète |
| `docker-compose.yml` | Port 2121 exposé |

**Total :** ~1500 lignes de code ajoutées

---

## 🎓 Exemples d'Utilisation

### Scénario 1: Détection Brute Force

Un attaquant tente un brute force :

```bash
# Commandes de l'attaquant
ftp localhost 2121
> user admin
> password1
> user admin
> password2
> user admin
> password3
```

**Résultat :**
- ✅ Toutes les tentatives enregistrées dans `ftp_connections`
- ✅ Alerte WARNING générée après 2+ tentatives
- ✅ Email envoyé si configuré
- ✅ Visible dans le dashboard

### Scénario 2: Upload Malveillant

Tentative d'upload d'un web shell :

```bash
ftp localhost 2121
> user admin
> admin123
> stor shell.php
```

**Résultat :**
- ✅ Connexion loggée avec credentials
- ✅ Commande `STOR shell.php` enregistrée
- ✅ Alerte **CRITICAL** générée (extension .php détectée)
- ✅ Email immédiat envoyé
- ✅ Détails disponibles dans API

### Scénario 3: Reconnaissance

Exploration du système de fichiers :

```bash
ftp localhost 2121
> user anonymous
> (vide)
> pwd
> cwd /etc
> list
> retr passwd
```

**Résultat :**
- ✅ Connexion anonymous détectée (WARNING)
- ✅ Navigation dans /etc loggée
- ✅ Tentative RETR passwd enregistrée
- ✅ Toutes les commandes tracées
- ✅ Pattern d'attaque identifié

---

## 🆚 Comparaison avec autres honeypots FTP

| Fonctionnalité | Ce Honeypot | Honeypots basiques | Cowrie FTP |
|----------------|-------------|-------------------|------------|
| **Commandes FTP** | 20+ | 5-10 | 15+ |
| **Détection malware** | ✅ | ❌ | ✅ |
| **API REST** | ✅ | ❌ | ❌ |
| **Dashboard web** | ✅ | ❌ | ❌ |
| **Alertes email** | ✅ | ⚠️ | ⚠️ |
| **Base de données** | SQLite | Logs texte | MySQL/Logs |
| **Multi-protocoles** | SSH+HTTP+FTP | FTP seul | SSH+FTP |
| **Docker** | ✅ | ⚠️ | ✅ |
| **Analyse comportementale** | ✅ | ❌ | ✅ |
| **Statistiques temps réel** | ✅ | ❌ | ⚠️ |

---

## 🔮 Évolutions Futures Possibles

### Améliorations FTP

1. **Data Connection Simulation**
   - Simuler vraie data connection PASV/PORT
   - Envoyer fausses listes de fichiers réalistes
   - Simuler transferts avec progression

2. **Honeytokens FTP**
   - Créer faux fichiers attractifs (passwords.txt, backup.sql)
   - Tracer qui télécharge quoi
   - Générer alertes spécifiques

3. **Détection Avancée**
   - Pattern matching sur séquences de commandes
   - Machine learning pour identifier comportements
   - Corrélation SSH + FTP (même IP)

4. **Dashboard FTP**
   - Page dédiée FTP dans le dashboard
   - Graphiques temps réel
   - Carte géographique des connexions FTP

5. **Support FTP over TLS (FTPS)**
   - Ajouter support AUTH TLS
   - Commandes PROT, PBSZ
   - Certificats auto-signés

---

## 📚 Ressources

### Documentation FTP

- [RFC 959 - File Transfer Protocol](https://tools.ietf.org/html/rfc959)
- [FTP Commands Reference](https://www.smartftp.com/en-us/support/ftp-commands)

### Tests et Validation

- Script de test : `scripts/test_ftp.sh`
- Commandes manuelles : `ftp localhost 2121`
- Client GUI : FileZilla vers localhost:2121

### API Documentation

- Connexions : `GET /api/ftp/connections?limit=N`
- Commandes : `GET /api/ftp/commands?limit=N`
- Stats : `GET /api/ftp/statistics`

---

## ✅ Checklist de Validation

- [x] Serveur FTP démarre correctement
- [x] 20+ commandes FTP fonctionnelles
- [x] Authentification simulée (toujours réussit)
- [x] Détection anonymous login
- [x] Détection brute force
- [x] Détection uploads malveillants (.php, .exe, .sh)
- [x] Tables base de données créées
- [x] API REST fonctionnelle (3 endpoints)
- [x] Logs complets dans SQLite
- [x] Alertes générées pour activités suspectes
- [x] Docker build réussi
- [x] Port 2121 exposé
- [x] Script de test fonctionnel
- [x] Documentation complète

---

**Statut:** ✅ **COMPLÉTÉ ET TESTÉ**
**Version:** 4.1
**Date:** 2025-12-20

Le honeypot FTP est maintenant **pleinement opérationnel** et intégré au système existant !
