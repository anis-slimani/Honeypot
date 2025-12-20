package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"honey/internal/config"
	"honey/internal/models"
)

// EmailSender interface pour l'envoi d'emails
type EmailSender interface {
	SendAlert(alert *models.Alert) error
	SendTestEmail() error
}

// SMTPEmailSender implémentation SMTP pour l'envoi d'emails
type SMTPEmailSender struct {
	config config.EmailConfig
}

// NewEmailSender crée un nouveau sender email
func NewEmailSender(cfg config.EmailConfig) EmailSender {
	return &SMTPEmailSender{
		config: cfg,
	}
}

// SendAlert envoie une alerte par email
func (s *SMTPEmailSender) SendAlert(alert *models.Alert) error {
	if len(s.config.To) == 0 {
		return fmt.Errorf("no recipients configured")
	}

	// Sujet détaillé avec toutes les infos importantes
	subject := s.buildAlertSubject(alert)
	body := s.buildAlertEmailBody(alert)

	return s.sendEmail(subject, body)
}

// buildAlertSubject construit un sujet d'email informatif
func (s *SMTPEmailSender) buildAlertSubject(alert *models.Alert) string {
	severityIcon := map[string]string{
		"critical": "🔴",
		"high":     "🟠",
		"medium":   "🟡",
		"low":      "🔵",
	}
	
	icon := severityIcon[alert.Severity]
	if icon == "" {
		icon = "⚪"
	}
	
	// Extraire le username si disponible
	username := ""
	if strings.Contains(alert.Details, "Username:") {
		parts := strings.Split(alert.Details, "Username: ")
		if len(parts) > 1 {
			userPart := strings.Split(parts[1], ",")[0]
			username = strings.TrimSpace(userPart)
		}
	}
	
	// Construire le sujet selon le type d'alerte
	switch alert.Type {
	case "successful_login":
		if username != "" {
			return fmt.Sprintf("%s INTRUSION - Connexion réussie (%s) depuis %s", 
				icon, username, s.formatIP(alert.RemoteAddr))
		}
		return fmt.Sprintf("%s INTRUSION - Connexion réussie depuis %s", 
			icon, s.formatIP(alert.RemoteAddr))
			
	case "brute_force":
		return fmt.Sprintf("%s ATTAQUE - Force brute détectée depuis %s", 
			icon, s.formatIP(alert.RemoteAddr))
			
	case "dangerous_command":
		// Extraire la commande
		command := ""
		if strings.Contains(alert.Details, "Command:") {
			parts := strings.Split(alert.Details, "Command: ")
			if len(parts) > 1 {
				command = strings.TrimSpace(parts[1])
				// Limiter à 30 caractères pour le sujet
				if len(command) > 30 {
					command = command[:27] + "..."
				}
			}
		}
		if command != "" {
			return fmt.Sprintf("%s COMMANDE SUSPECTE - '%s' par %s", 
				icon, command, s.formatIP(alert.RemoteAddr))
		}
		return fmt.Sprintf("%s COMMANDE SUSPECTE depuis %s", 
			icon, s.formatIP(alert.RemoteAddr))
			
	case "failed_login":
		if username != "" {
			return fmt.Sprintf("%s Tentative échouée (%s) depuis %s", 
				icon, username, s.formatIP(alert.RemoteAddr))
		}
		return fmt.Sprintf("%s Tentative échouée depuis %s", 
			icon, s.formatIP(alert.RemoteAddr))
			
	default:
		return fmt.Sprintf("%s [HONEYPOT] %s - %s", 
			icon, strings.ToUpper(alert.Severity), alert.Type)
	}
}

// formatIP formate l'adresse IP pour le sujet
func (s *SMTPEmailSender) formatIP(remoteAddr string) string {
	// Extraire l'IP du format "IP:port"
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		return remoteAddr[:idx]
	}
	return remoteAddr
}

// SendTestEmail envoie un email de test
func (s *SMTPEmailSender) SendTestEmail() error {
	subject := "[HONEYPOT TEST] Test de configuration email"
	body := `Ceci est un email de test pour vérifier la configuration du honeypot.

Si vous recevez ce message, la configuration email fonctionne correctement.

Date: ` + time.Now().Format("2006-01-02 15:04:05") + `
Serveur: Honey SSH Honeypot`

	return s.sendEmail(subject, body)
}

// buildAlertEmailBody construit le corps de l'email d'alerte
func (s *SMTPEmailSender) buildAlertEmailBody(alert *models.Alert) string {
	var body strings.Builder

	// En-tête avec emoji selon la sévérité
	severityEmoji := s.getSeverityEmoji(alert.Severity)
	body.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	body.WriteString(fmt.Sprintf("║  %s  ALERTE HONEYPOT - %s                      \n", severityEmoji, strings.ToUpper(alert.Severity)))
	body.WriteString("╚════════════════════════════════════════════════════════════════╝\n\n")

	// Informations principales
	body.WriteString("📋 INFORMATIONS GÉNÉRALES\n")
	body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	body.WriteString(fmt.Sprintf("Type d'alerte    : %s\n", s.formatAlertType(alert.Type)))
	body.WriteString(fmt.Sprintf("Niveau de risque : %s\n", s.formatSeverity(alert.Severity)))
	body.WriteString(fmt.Sprintf("Date et heure    : %s\n", alert.CreatedAt.Format("2006-01-02 à 15:04:05 MST")))
	body.WriteString(fmt.Sprintf("Message          : %s\n\n", alert.Message))

	// Informations sur l'attaquant
	body.WriteString("👤 INFORMATIONS SUR L'ATTAQUANT\n")
	body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	body.WriteString(fmt.Sprintf("Adresse IP       : %s\n", alert.RemoteAddr))
	
	// Extraire les détails selon le type d'alerte
	details := s.parseAlertDetails(alert)
	
	if username, ok := details["username"]; ok {
		body.WriteString(fmt.Sprintf("Nom d'utilisateur: %s\n", username))
	}
	
	if password, ok := details["password"]; ok {
		body.WriteString(fmt.Sprintf("Mot de passe     : %s\n", password))
	}
	
	if attempts, ok := details["attempts"]; ok {
		body.WriteString(fmt.Sprintf("Tentatives       : %s\n", attempts))
	}
	
	body.WriteString("\n")

	// Détails spécifiques selon le type d'alerte
	body.WriteString(s.buildTypeSpecificDetails(alert, details))

	// Recommandations
	body.WriteString(s.buildRecommendations(alert))

	// Footer
	body.WriteString("\n╔════════════════════════════════════════════════════════════════╗\n")
	body.WriteString("║  Honey SSH Honeypot - Système de surveillance automatique     ║\n")
	body.WriteString("║  🌐 Dashboard: http://localhost:8080                           ║\n")
	body.WriteString("╚════════════════════════════════════════════════════════════════╝\n")

	return body.String()
}

// getSeverityEmoji retourne l'emoji correspondant à la sévérité
func (s *SMTPEmailSender) getSeverityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🔵"
	default:
		return "⚪"
	}
}

// formatAlertType formate le type d'alerte de manière lisible
func (s *SMTPEmailSender) formatAlertType(alertType string) string {
	types := map[string]string{
		"successful_login":  "Connexion réussie",
		"failed_login":      "Tentative de connexion échouée",
		"brute_force":       "Attaque par force brute",
		"dangerous_command": "Commande dangereuse exécutée",
	}
	
	if formatted, ok := types[alertType]; ok {
		return formatted
	}
	return alertType
}

// formatSeverity formate la sévérité de manière lisible
func (s *SMTPEmailSender) formatSeverity(severity string) string {
	severities := map[string]string{
		"critical": "🔴 CRITIQUE - Action immédiate requise",
		"high":     "🟠 ÉLEVÉ - Attention requise",
		"medium":   "🟡 MOYEN - Surveillance recommandée",
		"low":      "🔵 FAIBLE - Information",
	}
	
	if formatted, ok := severities[severity]; ok {
		return formatted
	}
	return severity
}

// parseAlertDetails extrait les détails de l'alerte
func (s *SMTPEmailSender) parseAlertDetails(alert *models.Alert) map[string]string {
	details := make(map[string]string)
	
	if alert.Details == "" {
		return details
	}
	
	// Parser les détails (format: "Key: Value, Key2: Value2")
	parts := strings.Split(alert.Details, ", ")
	for _, part := range parts {
		keyValue := strings.SplitN(part, ": ", 2)
		if len(keyValue) == 2 {
			key := strings.ToLower(strings.TrimSpace(keyValue[0]))
			value := strings.TrimSpace(keyValue[1])
			details[key] = value
		}
	}
	
	return details
}

// buildTypeSpecificDetails construit les détails spécifiques selon le type d'alerte
func (s *SMTPEmailSender) buildTypeSpecificDetails(alert *models.Alert, details map[string]string) string {
	var body strings.Builder
	
	switch alert.Type {
	case "successful_login":
		body.WriteString("✅ DÉTAILS DE LA CONNEXION\n")
		body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		body.WriteString("Un attaquant a réussi à se connecter au honeypot avec des\n")
		body.WriteString("identifiants valides. Cela indique une tentative d'intrusion\n")
		body.WriteString("active sur votre système.\n\n")
		body.WriteString("⚠️  L'attaquant a maintenant accès au shell factice et peut\n")
		body.WriteString("    exécuter des commandes qui seront enregistrées.\n\n")
		
	case "brute_force":
		body.WriteString("⚔️  DÉTAILS DE L'ATTAQUE\n")
		body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		body.WriteString("Une attaque automatisée par force brute a été détectée.\n")
		body.WriteString("L'attaquant tente de deviner les identifiants en essayant\n")
		body.WriteString("plusieurs combinaisons username/password.\n\n")
		
		if attempts, ok := details["total attempts"]; ok {
			body.WriteString(fmt.Sprintf("Nombre de tentatives : %s\n", attempts))
		}
		body.WriteString("Fenêtre de temps     : 5 minutes\n")
		body.WriteString("Seuil de détection   : 5 tentatives\n\n")
		
	case "dangerous_command":
		body.WriteString("⚠️  COMMANDE EXÉCUTÉE\n")
		body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		
		if command, ok := details["command"]; ok {
			body.WriteString(fmt.Sprintf("Commande : %s\n\n", command))
			
			// Analyser le type de commande
			cmdLower := strings.ToLower(command)
			body.WriteString("Type de menace détectée :\n")
			
			if strings.Contains(cmdLower, "wget") || strings.Contains(cmdLower, "curl") {
				body.WriteString("  🔻 Téléchargement de fichier malveillant\n")
				body.WriteString("     → L'attaquant tente de télécharger un script ou malware\n\n")
			}
			if strings.Contains(cmdLower, "rm -rf") {
				body.WriteString("  💣 Tentative de destruction de données\n")
				body.WriteString("     → Commande de suppression récursive détectée\n\n")
			}
			if strings.Contains(cmdLower, "sudo") || strings.Contains(cmdLower, "su") {
				body.WriteString("  🔐 Tentative d'élévation de privilèges\n")
				body.WriteString("     → L'attaquant cherche à obtenir les droits root\n\n")
			}
			if strings.Contains(cmdLower, "chmod 777") || strings.Contains(cmdLower, "chmod +x") {
				body.WriteString("  🔓 Modification des permissions\n")
				body.WriteString("     → Tentative de rendre des fichiers exécutables\n\n")
			}
			if strings.Contains(cmdLower, "nc") || strings.Contains(cmdLower, "netcat") {
				body.WriteString("  🌐 Reverse shell / Backdoor\n")
				body.WriteString("     → Tentative d'établir une connexion sortante\n\n")
			}
			if strings.Contains(cmdLower, "python -c") || strings.Contains(cmdLower, "perl -e") || strings.Contains(cmdLower, "bash -i") {
				body.WriteString("  💻 Exécution de code arbitraire\n")
				body.WriteString("     → Script malveillant exécuté directement\n\n")
			}
			if strings.Contains(cmdLower, "cat /etc/passwd") || strings.Contains(cmdLower, "cat /etc/shadow") {
				body.WriteString("  👁️  Énumération du système\n")
				body.WriteString("     → Collecte d'informations sur les utilisateurs\n\n")
			}
			if strings.Contains(cmdLower, "history -c") {
				body.WriteString("  🧹 Effacement des traces\n")
				body.WriteString("     → Tentative de supprimer l'historique des commandes\n\n")
			}
		}
		
	case "failed_login":
		body.WriteString("🔒 TENTATIVE ÉCHOUÉE\n")
		body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		body.WriteString("Tentative de connexion avec des identifiants incorrects.\n")
		body.WriteString("Cela peut être une reconnaissance ou le début d'une attaque.\n\n")
	}
	
	return body.String()
}

// buildRecommendations construit les recommandations selon la sévérité
func (s *SMTPEmailSender) buildRecommendations(alert *models.Alert) string {
	var body strings.Builder
	
	body.WriteString("💡 RECOMMANDATIONS\n")
	body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	
	switch alert.Severity {
	case "critical":
		body.WriteString("⚠️  ACTION IMMÉDIATE REQUISE :\n")
		body.WriteString("  1. Vérifiez immédiatement le dashboard pour plus de détails\n")
		body.WriteString("  2. Analysez toutes les commandes exécutées par cet attaquant\n")
		body.WriteString("  3. Bloquez cette IP dans votre firewall si nécessaire\n")
		body.WriteString("  4. Vérifiez vos vrais serveurs pour des activités similaires\n")
		body.WriteString(fmt.Sprintf("  5. Considérez le signalement de %s aux autorités\n\n", alert.RemoteAddr))
		
	case "high":
		body.WriteString("⚠️  ATTENTION REQUISE :\n")
		body.WriteString("  1. Consultez le dashboard pour analyser l'activité\n")
		body.WriteString("  2. Surveillez cette IP pour d'autres tentatives\n")
		body.WriteString("  3. Vérifiez que vos vrais services ne sont pas exposés\n")
		body.WriteString(fmt.Sprintf("  4. Envisagez de bloquer %s temporairement\n\n", alert.RemoteAddr))
		
	case "medium":
		body.WriteString("ℹ️  SURVEILLANCE RECOMMANDÉE :\n")
		body.WriteString("  1. Notez cette activité dans vos logs\n")
		body.WriteString("  2. Surveillez si l'activité continue\n")
		body.WriteString("  3. Pas d'action immédiate requise\n\n")
		
	case "low":
		body.WriteString("ℹ️  INFORMATION :\n")
		body.WriteString("  • Cette alerte est à titre informatif\n")
		body.WriteString("  • Aucune action immédiate n'est nécessaire\n")
		body.WriteString("  • Les données sont enregistrées pour analyse ultérieure\n\n")
	}
	
	body.WriteString("📊 Pour plus de détails, consultez le dashboard : http://localhost:8080\n")
	body.WriteString(fmt.Sprintf("🔍 Recherchez l'IP : %s\n", alert.RemoteAddr))
	
	return body.String()
}

// sendEmail envoie un email via SMTP
func (s *SMTPEmailSender) sendEmail(subject, body string) error {
	// Configuration SMTP
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	
	// Authentification
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)

	// Construction du message
	msg := s.buildEmailMessage(s.config.From, s.config.To, subject, body)

	// Envoi de l'email
	if s.config.SMTPPort == 587 {
		// Utiliser STARTTLS pour le port 587 (Gmail, Outlook, etc.)
		return s.sendEmailSTARTTLS(addr, auth, s.config.From, s.config.To, msg)
	} else {
		// Utiliser SMTP standard pour les autres ports
		return smtp.SendMail(addr, auth, s.config.From, s.config.To, []byte(msg))
	}
}

// sendEmailSTARTTLS envoie un email avec STARTTLS (pour Gmail, Outlook, etc.)
func (s *SMTPEmailSender) sendEmailSTARTTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Connexion TCP non chiffrée au serveur SMTP
	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Envoyer EHLO
	if err := conn.Hello("localhost"); err != nil {
		return fmt.Errorf("failed to send EHLO: %w", err)
	}

	// Démarrer TLS avec STARTTLS
	if ok, _ := conn.Extension("STARTTLS"); ok {
		config := &tls.Config{
			ServerName: s.config.SMTPHost,
			MinVersion: tls.VersionTLS12,
		}
		if err := conn.StartTLS(config); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authentification
	if err := conn.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Définir l'expéditeur
	if err := conn.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Définir les destinataires
	for _, recipient := range to {
		if err := conn.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Envoyer le message
	writer, err := conn.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err := writer.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}

// buildEmailMessage construit le message email complet
func (s *SMTPEmailSender) buildEmailMessage(from string, to []string, subject, body string) string {
	var msg strings.Builder

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")

	// Corps du message
	msg.WriteString(body)

	return msg.String()
}
