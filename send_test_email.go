package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     📧 TEST D'ENVOI D'EMAIL HONEYPOT                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Configuration email
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	username := "honeypotprojet@gmail.com"
	password := "bivc pvlh schv ifcv"
	from := "honeypotprojet@gmail.com"
	to := []string{"honeypotprojet@gmail.com"}

	fmt.Println("📋 Configuration:")
	fmt.Printf("   SMTP: %s:%d\n", smtpHost, smtpPort)
	fmt.Printf("   De:   %s\n", from)
	fmt.Printf("   À:    %s\n", strings.Join(to, ", "))
	fmt.Println()

	// Créer l'email de test
	subject := "🧪 TEST - Honeypot Email Fonctionne !"
	body := buildTestEmail()

	fmt.Println("📧 Envoi de l'email de test...")
	fmt.Println()

	// Envoyer l'email
	err := sendEmail(smtpHost, smtpPort, username, password, from, to, subject, body)
	
	if err != nil {
		fmt.Printf("❌ ERREUR: %v\n", err)
		fmt.Println()
		fmt.Println("🔧 Solutions possibles:")
		fmt.Println("   1. Vérifie que le mot de passe est correct")
		fmt.Println("   2. Vérifie ta connexion internet")
		fmt.Println("   3. Outlook bloque parfois temporairement - réessaye dans 5 min")
		return
	}

	fmt.Println("✅ EMAIL ENVOYÉ AVEC SUCCÈS !")
	fmt.Println()
	fmt.Println("📮 Vérifie maintenant ta boîte email:")
	fmt.Println("   → https://mail.google.com/")
	fmt.Println("   → Email: honeypotprojet@gmail.com")
	fmt.Println()
	fmt.Println("⚠️  Si tu ne le vois pas:")
	fmt.Println("   - Attends 1-2 minutes")
	fmt.Println("   - Vérifie les SPAMS / Courrier indésirable")
	fmt.Println()
}

func buildTestEmail() string {
	var body strings.Builder

	body.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	body.WriteString("║  🧪  EMAIL DE TEST - CONFIGURATION RÉUSSIE                     ║\n")
	body.WriteString("╚════════════════════════════════════════════════════════════════╝\n\n")

	body.WriteString("Félicitations ! 🎉\n\n")
	body.WriteString("Si tu reçois cet email, cela signifie que la configuration\n")
	body.WriteString("email de ton honeypot fonctionne parfaitement !\n\n")

	body.WriteString("📋 INFORMATIONS DE TEST\n")
	body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	body.WriteString(fmt.Sprintf("Date et heure    : %s\n", time.Now().Format("2006-01-02 à 15:04:05 MST")))
	body.WriteString("Type de test     : Email standalone (sans Docker)\n")
	body.WriteString("Service SMTP     : Gmail (smtp.gmail.com:587)\n")
	body.WriteString("Status           : ✅ Fonctionnel\n\n")

	body.WriteString("🎯 PROCHAINES ÉTAPES\n")
	body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	body.WriteString("1. ✅ Configuration email validée\n")
	body.WriteString("2. 🐳 Lance le honeypot avec Docker:\n")
	body.WriteString("   → cd /home/anis/Honey/HoneyV3/Honeypot\n")
	body.WriteString("   → docker-compose down\n")
	body.WriteString("   → docker-compose up -d --build\n\n")
	body.WriteString("3. 🧪 Teste avec ./test_email.sh\n\n")
	body.WriteString("4. 📊 Consulte le dashboard: http://localhost:8080\n\n")

	body.WriteString("📧 EXEMPLE D'EMAIL D'ALERTE\n")
	body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	body.WriteString("Quand quelqu'un se connecte au honeypot, tu recevras des emails\n")
	body.WriteString("beaucoup plus détaillés que celui-ci, avec:\n\n")
	body.WriteString("  • 🌐 L'adresse IP de l'attaquant\n")
	body.WriteString("  • 👤 Le nom d'utilisateur utilisé\n")
	body.WriteString("  • 🔑 Le mot de passe tenté\n")
	body.WriteString("  • 💻 Les commandes exécutées\n")
	body.WriteString("  • ⚠️  Le type de menace détecté\n")
	body.WriteString("  • 💡 Les recommandations d'actions\n\n")

	body.WriteString("📚 DOCUMENTATION\n")
	body.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	body.WriteString("Consulte ces fichiers pour plus d'infos:\n")
	body.WriteString("  • EXEMPLES_EMAILS.md - Voir des exemples d'alertes\n")
	body.WriteString("  • EMAIL_IMPROVEMENTS.md - Liste des améliorations\n")
	body.WriteString("  • EMAIL_SETUP_COMPLETE.md - Guide complet\n\n")

	body.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	body.WriteString("║  Honey SSH Honeypot - Système de surveillance automatique     ║\n")
	body.WriteString("║  🌐 Dashboard: http://localhost:8080                           ║\n")
	body.WriteString("╚════════════════════════════════════════════════════════════════╝\n")

	return body.String()
}

func sendEmail(smtpHost string, smtpPort int, username, password, from string, to []string, subject, body string) error {
	// Adresse du serveur SMTP
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)

	// Authentification
	auth := smtp.PlainAuth("", username, password, smtpHost)

	// Construction du message
	msg := buildEmailMessage(from, to, subject, body)

	// Connexion TCP au serveur SMTP
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("échec de connexion au serveur SMTP: %w", err)
	}
	defer client.Close()

	// Envoyer EHLO
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("échec EHLO: %w", err)
	}

	// Démarrer STARTTLS
	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{
			ServerName: smtpHost,
			MinVersion: tls.VersionTLS12,
		}
		if err := client.StartTLS(config); err != nil {
			return fmt.Errorf("échec STARTTLS: %w", err)
		}
	}

	// Authentification
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("échec authentification: %w", err)
	}

	// Définir l'expéditeur
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("échec définition expéditeur: %w", err)
	}

	// Définir les destinataires
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("échec définition destinataire %s: %w", recipient, err)
		}
	}

	// Envoyer le message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("échec obtention writer: %w", err)
	}

	if _, err := writer.Write([]byte(msg)); err != nil {
		return fmt.Errorf("échec écriture message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("échec fermeture writer: %w", err)
	}

	return nil
}

func buildEmailMessage(from string, to []string, subject, body string) string {
	var msg strings.Builder

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")

	// Corps
	msg.WriteString(body)

	return msg.String()
}


