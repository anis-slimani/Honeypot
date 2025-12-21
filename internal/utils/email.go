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

	ip := s.formatIP(alert.RemoteAddr)

	// Construire le sujet selon le type d'alerte
	switch alert.Type {
	case "successful_login":
		return fmt.Sprintf("%s HONEYPOT ALERT - Successful Login from %s", icon, ip)

	case "brute_force":
		return fmt.Sprintf("%s HONEYPOT ALERT - Brute Force Attack from %s", icon, ip)

	case "dangerous_command":
		// Extraire la commande pour un aperçu court
		command := ""
		if strings.Contains(alert.Details, "Command:") {
			parts := strings.Split(alert.Details, "Command: ")
			if len(parts) > 1 {
				command = strings.TrimSpace(parts[1])
				// Limiter à 25 caractères pour le sujet
				if len(command) > 25 {
					command = command[:22] + "..."
				}
			}
		}
		if command != "" {
			return fmt.Sprintf("%s HONEYPOT ALERT - Suspicious Command '%s' from %s", icon, command, ip)
		}
		return fmt.Sprintf("%s HONEYPOT ALERT - Suspicious Command from %s", icon, ip)

	case "failed_login":
		return fmt.Sprintf("%s HONEYPOT ALERT - Failed Login from %s", icon, ip)

	default:
		return fmt.Sprintf("%s HONEYPOT ALERT - %s from %s", icon, strings.ToUpper(alert.Severity), ip)
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

	// En-tête simple et professionnel
	severityEmoji := s.getSeverityEmoji(alert.Severity)
	body.WriteString("HONEYPOT SECURITY ALERT\n")
	body.WriteString("========================================\n\n")

	// Niveau de criticité
	body.WriteString(fmt.Sprintf("CRITICALITY: %s %s\n", severityEmoji, strings.ToUpper(alert.Severity)))
	body.WriteString(fmt.Sprintf("ALERT TYPE: %s\n", s.formatAlertType(alert.Type)))
	body.WriteString(fmt.Sprintf("TIMESTAMP: %s\n\n", alert.CreatedAt.Format("2006-01-02 15:04:05 MST")))

	// Informations essentielles
	body.WriteString("THREAT DETAILS\n")
	body.WriteString("----------------------------------------\n")

	// Extraire les détails
	details := s.parseAlertDetails(alert)

	// IP Address (toujours affichée)
	body.WriteString(fmt.Sprintf("Source IP: %s\n", alert.RemoteAddr))

	// Username (mais pas le password)
	if username, ok := details["username"]; ok {
		body.WriteString(fmt.Sprintf("Username: %s\n", username))
	}

	// Commande exécutée (pour dangerous_command)
	if command, ok := details["command"]; ok {
		body.WriteString(fmt.Sprintf("Command Executed: %s\n", command))
	}

	// Nombre de tentatives (pour brute force)
	if attempts, ok := details["total attempts"]; ok {
		body.WriteString(fmt.Sprintf("Failed Attempts: %s\n", attempts))
	}

	body.WriteString("\n")

	// Analyse de menace spécifique
	body.WriteString(s.buildThreatAnalysis(alert, details))

	// Recommandations d'action
	body.WriteString(s.buildActionRecommendations(alert))

	// Footer
	body.WriteString("\n========================================\n")
	body.WriteString("Honey SSH Honeypot - Security Monitoring System\n")
	body.WriteString("Dashboard: http://localhost:8080\n")

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

// buildThreatAnalysis construit une analyse concise de la menace
func (s *SMTPEmailSender) buildThreatAnalysis(alert *models.Alert, details map[string]string) string {
	var body strings.Builder

	body.WriteString("THREAT ANALYSIS\n")
	body.WriteString("----------------------------------------\n")

	switch alert.Type {
	case "successful_login":
		body.WriteString("Status: ACTIVE INTRUSION\n")
		body.WriteString("An attacker successfully authenticated to the honeypot.\n")
		body.WriteString("The attacker now has access to the fake shell environment.\n\n")

	case "brute_force":
		body.WriteString("Status: BRUTE FORCE ATTACK DETECTED\n")
		body.WriteString("Multiple failed authentication attempts detected.\n")
		if attempts, ok := details["total attempts"]; ok {
			body.WriteString(fmt.Sprintf("Total attempts: %s in 5 minutes\n", attempts))
		}
		body.WriteString("Detection threshold: 5 attempts\n\n")

	case "dangerous_command":
		body.WriteString("Status: MALICIOUS COMMAND EXECUTED\n")

		if command, ok := details["command"]; ok {
			cmdLower := strings.ToLower(command)

			// Analyse concise du type de menace
			if strings.Contains(cmdLower, "wget") || strings.Contains(cmdLower, "curl") {
				body.WriteString("Threat Type: Malware Download Attempt\n")
			} else if strings.Contains(cmdLower, "rm -rf") {
				body.WriteString("Threat Type: Data Destruction Attempt\n")
			} else if strings.Contains(cmdLower, "sudo") || strings.Contains(cmdLower, "su") {
				body.WriteString("Threat Type: Privilege Escalation Attempt\n")
			} else if strings.Contains(cmdLower, "chmod 777") {
				body.WriteString("Threat Type: Permission Modification\n")
			} else if strings.Contains(cmdLower, "nc") || strings.Contains(cmdLower, "netcat") {
				body.WriteString("Threat Type: Reverse Shell / Backdoor Attempt\n")
			} else if strings.Contains(cmdLower, "python -c") || strings.Contains(cmdLower, "perl -e") || strings.Contains(cmdLower, "bash -i") {
				body.WriteString("Threat Type: Arbitrary Code Execution\n")
			} else if strings.Contains(cmdLower, "cat /etc/passwd") || strings.Contains(cmdLower, "cat /etc/shadow") {
				body.WriteString("Threat Type: System Enumeration\n")
			} else if strings.Contains(cmdLower, "history -c") {
				body.WriteString("Threat Type: Anti-Forensics / Cover Tracks\n")
			} else {
				body.WriteString("Threat Type: Suspicious Activity\n")
			}
		}
		body.WriteString("\n")

	case "failed_login":
		body.WriteString("Status: FAILED AUTHENTICATION\n")
		body.WriteString("Invalid credentials used - possible reconnaissance.\n\n")
	}

	return body.String()
}

// buildActionRecommendations construit les recommandations d'action
func (s *SMTPEmailSender) buildActionRecommendations(alert *models.Alert) string {
	var body strings.Builder

	body.WriteString("RECOMMENDED ACTIONS\n")
	body.WriteString("----------------------------------------\n")

	switch alert.Severity {
	case "critical":
		body.WriteString("IMMEDIATE ACTION REQUIRED:\n")
		body.WriteString("1. Review all activity from this IP immediately\n")
		body.WriteString("2. Consider blocking this IP in your firewall\n")
		body.WriteString("3. Check production systems for similar activity\n")
		body.WriteString("4. Consider reporting to authorities if necessary\n\n")

	case "high":
		body.WriteString("ACTION RECOMMENDED:\n")
		body.WriteString("1. Monitor this IP for continued activity\n")
		body.WriteString("2. Review the dashboard for full details\n")
		body.WriteString("3. Consider temporary IP blocking\n\n")

	case "medium":
		body.WriteString("MONITORING RECOMMENDED:\n")
		body.WriteString("1. Log this activity for future reference\n")
		body.WriteString("2. Monitor for pattern changes\n\n")

	case "low":
		body.WriteString("NO IMMEDIATE ACTION REQUIRED:\n")
		body.WriteString("Activity logged for analysis.\n\n")
	}

	body.WriteString(fmt.Sprintf("View full details: http://localhost:8080\nSearch for IP: %s\n", alert.RemoteAddr))

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
		return s.sendEmailSTARTTLS(addr, auth, s.config.From, s.config.To, []byte(msg))
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
