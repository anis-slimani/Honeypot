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

	subject := fmt.Sprintf("[HONEYPOT ALERT] %s - %s", alert.Severity, alert.Type)
	body := s.buildAlertEmailBody(alert)

	return s.sendEmail(subject, body)
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

	body.WriteString("🚨 ALERTE HONEYPOT 🚨\n\n")
	body.WriteString(fmt.Sprintf("Type: %s\n", alert.Type))
	body.WriteString(fmt.Sprintf("Sévérité: %s\n", alert.Severity))
	body.WriteString(fmt.Sprintf("Message: %s\n", alert.Message))
	body.WriteString(fmt.Sprintf("Adresse IP: %s\n", alert.RemoteAddr))
	body.WriteString(fmt.Sprintf("Date: %s\n", alert.CreatedAt.Format("2006-01-02 15:04:05")))
	
	if alert.Details != "" {
		body.WriteString(fmt.Sprintf("Détails: %s\n", alert.Details))
	}

	body.WriteString("\n---\n")
	body.WriteString("Honey SSH Honeypot\n")
	body.WriteString("Système de surveillance automatique")

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
		// Utiliser TLS pour le port 587
		return s.sendEmailTLS(addr, auth, s.config.From, s.config.To, msg)
	} else {
		// Utiliser SMTP standard pour les autres ports
		return smtp.SendMail(addr, auth, s.config.From, s.config.To, []byte(msg))
	}
}

// sendEmailTLS envoie un email avec TLS
func (s *SMTPEmailSender) sendEmailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Connexion au serveur SMTP
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.config.SMTPHost,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Créer le client SMTP
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authentification
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Définir l'expéditeur
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Définir les destinataires
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Envoyer le message
	writer, err := client.Data()
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
