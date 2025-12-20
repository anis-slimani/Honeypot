# 📧 CONFIGURATION DES EMAILS POUR LE HONEYPOT

## 🎯 OBJECTIF
Recevoir de vrais emails d'alerte quand :
- ✅ Quelqu'un se connecte avec succès au honeypot
- ⚠️ Une attaque brute force est détectée (5+ tentatives échouées)
- 🚨 Une commande dangereuse est exécutée (`rm -rf`, `wget`, `curl`, etc.)

---

## 📝 ÉTAPE 1 : CRÉER UN MOT DE PASSE D'APPLICATION GMAIL

### **Pourquoi un "App Password" ?**
Gmail bloque les connexions depuis des applications tierces pour des raisons de sécurité.
Tu dois créer un mot de passe spécial pour le honeypot.

### **Comment créer un App Password :**

1. **Va sur ton compte Google** : https://myaccount.google.com/

2. **Active la validation en 2 étapes** (si pas déjà fait) :
   - Va dans "Sécurité" → "Validation en 2 étapes"
   - Suis les instructions pour l'activer

3. **Crée un mot de passe d'application** :
   - Va dans "Sécurité" → "Mots de passe des applications"
   - Ou va directement sur : https://myaccount.google.com/apppasswords
   - Sélectionne "Autre (nom personnalisé)" et écris "Honeypot"
   - Clique sur "Générer"
   - **⚠️ COPIE LE MOT DE PASSE** (16 caractères sans espaces)
   - Exemple : `abcd efgh ijkl mnop` → `abcdefghijklmnop`

---

## 🔧 ÉTAPE 2 : MODIFIER config.yaml

Ouvre le fichier `config.yaml` et modifie la section `alerts.email` :

```yaml
alerts:
  enabled: true
  email:
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "TON_EMAIL@gmail.com"          # ← Ton adresse Gmail
    password: "ton_app_password_ici"         # ← Le mot de passe d'application (16 caractères)
    from: "TON_EMAIL@gmail.com"              # ← Ton adresse Gmail (même que username)
    to: 
      - "TON_EMAIL@gmail.com"                # ← L'email où tu veux recevoir les alertes
  thresholds:
    failed_attempts: 5
    time_window: 300  # 5 minutes
```

### **Exemple concret :**
```yaml
alerts:
  enabled: true
  email:
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    username: "anis.securite@gmail.com"
    password: "abcdefghijklmnop"
    from: "anis.securite@gmail.com"
    to: 
      - "anis.securite@gmail.com"
      - "anis.perso@gmail.com"              # Tu peux ajouter plusieurs emails
  thresholds:
    failed_attempts: 5
    time_window: 300
```

---

## 🔧 ÉTAPE 3 : ACTIVER L'ENVOI RÉEL D'EMAILS

Il faut modifier le code pour utiliser le vrai système d'email au lieu de la simulation.

### **Fichier à modifier : `internal/honeypot/alerts.go`**

Remplacer la fonction `saveAndLogAlert()` pour appeler le vrai système d'email.

---

## 🧪 ÉTAPE 4 : TESTER

### **Test 1 : Email de test**
```bash
# Se connecter au honeypot et exécuter une commande test
ssh -p 2222 admin@localhost
# Password: admin123

# Taper une commande dangereuse
wget http://example.com/malware.sh
```

### **Test 2 : Attaque brute force**
```bash
# Faire 6 tentatives échouées
for i in {1..6}; do 
  sshpass -p "wrongpassword" ssh -p 2222 test@localhost
done
```

### **Test 3 : Connexion réussie**
```bash
ssh -p 2222 admin@localhost
# Password: admin123
```

---

## 📮 EXEMPLE D'EMAIL REÇU

```
De: anis.securite@gmail.com
À: anis.securite@gmail.com
Sujet: [HONEYPOT ALERT] high - brute_force

🚨 ALERTE HONEYPOT 🚨

Type: brute_force
Sévérité: high
Message: ⚠️ ATTAQUE BRUTE FORCE DÉTECTÉE ! 6 tentatives depuis 192.168.1.50
Adresse IP: 192.168.1.50
Date: 2025-12-20 15:30:45

Détails: Username: test, Password: wrongpassword, Total attempts: 6

---
Honey SSH Honeypot
Système de surveillance automatique
```

---

## 🔐 ALTERNATIVES À GMAIL

### **Option 2 : Outlook/Hotmail**
```yaml
smtp_host: "smtp-mail.outlook.com"
smtp_port: 587
username: "ton_email@outlook.com"
password: "ton_mot_de_passe"
from: "ton_email@outlook.com"
```

### **Option 3 : Yahoo Mail**
```yaml
smtp_host: "smtp.mail.yahoo.com"
smtp_port: 587
username: "ton_email@yahoo.com"
password: "ton_app_password"  # Aussi besoin d'un App Password
from: "ton_email@yahoo.com"
```

### **Option 4 : Service SMTP personnalisé**
```yaml
smtp_host: "smtp.ton-serveur.com"
smtp_port: 587
username: "ton_username"
password: "ton_password"
from: "honeypot@ton-serveur.com"
```

---

## 🚨 SÉCURITÉ

### **⚠️ NE JAMAIS :**
- ❌ Utiliser ton vrai mot de passe Gmail dans la config
- ❌ Commit le fichier config.yaml avec tes identifiants sur Git
- ❌ Partager ton App Password

### **✅ BONNES PRATIQUES :**
- ✅ Utiliser un mot de passe d'application Gmail
- ✅ Ajouter `config.yaml` dans `.gitignore` (déjà fait)
- ✅ Créer un compte Gmail dédié pour le honeypot (optionnel)
- ✅ Révoquer le mot de passe d'application si compromis

---

## 📊 TYPES D'ALERTES

Le honeypot envoie des emails pour :

| Type | Sévérité | Déclencheur |
|------|----------|-------------|
| `successful_login` | medium | Connexion réussie |
| `failed_login` | low | Tentative échouée |
| `brute_force` | high | 5+ tentatives échouées en 5 min |
| `dangerous_command` | high/critical | Commandes suspectes (`wget`, `curl`, `rm -rf`, etc.) |

---

## 🔧 DÉPANNAGE

### **Problème : "Authentication failed"**
- ✅ Vérifie que tu as bien un App Password Gmail (pas ton mot de passe normal)
- ✅ Vérifie que la validation en 2 étapes est activée
- ✅ Copie le mot de passe sans espaces

### **Problème : "Connection refused"**
- ✅ Vérifie le port SMTP (587 pour Gmail)
- ✅ Vérifie ta connexion internet

### **Problème : Pas d'email reçu**
- ✅ Vérifie tes spams/courrier indésirable
- ✅ Vérifie que `alerts.enabled: true` dans config.yaml
- ✅ Vérifie les logs du honeypot : `docker-compose logs -f`

---

## 📞 BESOIN D'AIDE ?

Si tu veux que je configure les emails avec ton vrai email :
1. Donne-moi ton adresse Gmail
2. Crée ton App Password
3. Je modifierai le code pour activer les vrais emails

