package honeypot

import (
	"fmt"
	"strings"
	"time"

	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
)

// AlertManager gère les alertes du honeypot
type AlertManager struct {
	config *config.Config
	logger logger.Logger
}

// NewAlertManager crée un nouveau gestionnaire d'alertes
func NewAlertManager(cfg *config.Config, log logger.Logger) *AlertManager {
	return &AlertManager{
		config: cfg,
		logger: log,
	}
}

// OnSuccessfulConnection déclenche une alerte lors d'une connexion réussie
func (am *AlertManager) OnSuccessfulConnection(remoteAddr, username, password string) {
	alert := &models.Alert{
		Type:       "successful_login",
		Severity:   "medium",
		Message:    fmt.Sprintf("Connexion réussie détectée depuis %s", remoteAddr),
		RemoteAddr: remoteAddr,
		Details:    fmt.Sprintf("Username: %s, Password: %s", username, password),
		CreatedAt:  time.Now(),
		Sent:       false,
	}

	am.saveAndLogAlert(alert)
}

// OnFailedConnection déclenche une alerte lors d'une tentative échouée
func (am *AlertManager) OnFailedConnection(remoteAddr, username, password string) {
	// Vérifier si c'est une attaque brute force
	count := am.countRecentFailedAttempts(remoteAddr, 300) // 5 minutes

	severity := "low"
	alertType := "failed_login"
	message := fmt.Sprintf("Tentative de connexion échouée depuis %s", remoteAddr)

	if count >= 5 {
		severity = "high"
		alertType = "brute_force"
		message = fmt.Sprintf("⚠️ ATTAQUE BRUTE FORCE DÉTECTÉE ! %d tentatives depuis %s", count, remoteAddr)
	}

	alert := &models.Alert{
		Type:       alertType,
		Severity:   severity,
		Message:    message,
		RemoteAddr: remoteAddr,
		Details:    fmt.Sprintf("Username: %s, Password: %s, Total attempts: %d", username, password, count),
		CreatedAt:  time.Now(),
		Sent:       false,
	}

	am.saveAndLogAlert(alert)
}

// OnDangerousCommand déclenche une alerte lors d'une commande dangereuse
func (am *AlertManager) OnDangerousCommand(remoteAddr, username, command string) {
	dangerLevel := am.analyzeDangerLevel(command)
	
	if dangerLevel == "" {
		return // Pas une commande dangereuse
	}

	alert := &models.Alert{
		Type:       "dangerous_command",
		Severity:   dangerLevel,
		Message:    fmt.Sprintf("⚠️ Commande dangereuse exécutée par %s depuis %s", username, remoteAddr),
		RemoteAddr: remoteAddr,
		Details:    fmt.Sprintf("Command: %s", command),
		CreatedAt:  time.Now(),
		Sent:       false,
	}

	am.saveAndLogAlert(alert)
}

// analyzeDangerLevel analyse le niveau de danger d'une commande
func (am *AlertManager) analyzeDangerLevel(command string) string {
	cmd := strings.ToLower(command)

	// Commandes critiques
	criticalPatterns := []string{
		"rm -rf",
		"dd if=",
		"mkfs",
		":(){ :|:& };:", // fork bomb
		"chmod 777",
		"passwd",
		"> /dev/sda",
	}

	for _, pattern := range criticalPatterns {
		if strings.Contains(cmd, pattern) {
			return "critical"
		}
	}

	// Commandes hautement suspectes
	highPatterns := []string{
		"wget",
		"curl",
		"nc -",
		"netcat",
		"/bin/bash -i",
		"python -c",
		"perl -e",
		"bash -i",
		"sh -i",
		"base64 -d",
		"echo ",
		"history -c",
	}

	for _, pattern := range highPatterns {
		if strings.Contains(cmd, pattern) {
			return "high"
		}
	}

	// Commandes moyennement suspectes
	mediumPatterns := []string{
		"ps aux",
		"netstat",
		"ifconfig",
		"who",
		"w ",
		"last",
		"cat /etc/passwd",
		"cat /etc/shadow",
		"find / -name",
		"sudo ",
	}

	for _, pattern := range mediumPatterns {
		if strings.Contains(cmd, pattern) {
			return "medium"
		}
	}

	return "" // Pas dangereuse
}

// countRecentFailedAttempts compte les tentatives échouées récentes
func (am *AlertManager) countRecentFailedAttempts(remoteAddr string, timeWindowSeconds int) int {
	// Compter les connexions échouées dans la fenêtre de temps
	count, err := database.CountFailedConnectionsInTimeWindow(remoteAddr, timeWindowSeconds)
	if err != nil {
		am.logger.Errorf("Failed to count failed attempts: %v", err)
		return 0
	}
	return count
}

// saveAndLogAlert sauvegarde l'alerte et log
func (am *AlertManager) saveAndLogAlert(alert *models.Alert) {
	// Sauvegarder dans la base de données
	if err := database.SaveAlert(alert); err != nil {
		am.logger.Errorf("Failed to save alert: %v", err)
		return
	}

	// Logger selon la sévérité
	logMessage := fmt.Sprintf("[%s] %s - %s (Details: %s)", 
		alert.Severity, alert.Type, alert.Message, alert.Details)

	switch alert.Severity {
	case "critical":
		am.logger.Errorf("🚨 CRITICAL ALERT: %s", logMessage)
	case "high":
		am.logger.Warnf("⚠️  HIGH ALERT: %s", logMessage)
	case "medium":
		am.logger.Infof("ℹ️  MEDIUM ALERT: %s", logMessage)
	default:
		am.logger.Infof("ℹ️  ALERT: %s", logMessage)
	}

	// Simuler l'envoi d'email (pour l'instant on log juste)
	am.simulateEmailSend(alert)
}

// simulateEmailSend simule l'envoi d'un email
func (am *AlertManager) simulateEmailSend(alert *models.Alert) {
	emailContent := fmt.Sprintf(`
╔══════════════════════════════════════════════════════════════╗
║                    📧 EMAIL ENVOYÉ                           ║
╚══════════════════════════════════════════════════════════════╝

De:        security@honeypot.local
À:         honeypot@gmail.com
Sujet:     [HONEYPOT ALERT] %s - %s

🚨 ALERTE HONEYPOT 🚨

Type:      %s
Sévérité:  %s
Message:   %s
Adresse IP: %s
Date:      %s

Détails:
%s

---
Honey SSH Honeypot
Système de surveillance automatique
`, alert.Severity, alert.Type, alert.Type, alert.Severity, alert.Message, 
		alert.RemoteAddr, alert.CreatedAt.Format("2006-01-02 15:04:05"), alert.Details)

	am.logger.Infof(emailContent)
	
	// Marquer comme envoyé
	now := time.Now()
	alert.Sent = true
	alert.SentAt = &now
	database.UpdateAlert(alert)
}

